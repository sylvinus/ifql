package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/influxdata/ifql"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/influxdb/models"
	"github.com/jessevdk/go-flags"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type options struct {
	Hosts []string `long:"host" short:"h" description:"influx hosts to query from. Can be specified more than once for multiple hosts." default:"localhost:8082" env:"HOSTS" env-delim:","`
	Addr  string   `long:"bind-address" short:"b" description:"The address to listen on for HTTP requests" default:":8093" env:"BIND_ADDRESS"`
}

var hosts []string
var storageReader execute.StorageReader

func main() {
	option := &options{}
	parser := flags.NewParser(option, flags.Default)
	parser.ShortDescription = `IFQLD`
	parser.LongDescription = `Options for the IFQLD server`

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/query", http.HandlerFunc(HandleQuery))

	hosts = option.Hosts
	log.Printf("Starting on %s\n", option.Addr)
	log.Fatal(http.ListenAndServe(option.Addr, nil))
}

func HandleQuery(w http.ResponseWriter, req *http.Request) {
	query := req.FormValue("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("must pass query in q parameter"))
		return
	}

	verbose := req.FormValue("verbose") != ""
	trace := req.FormValue("trace") != ""

	results, err := ifql.Query(
		req.Context(),
		query,
		&ifql.Options{
			Verbose: verbose,
			Trace:   trace,
			Hosts:   hosts,
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error executing query %s", err.Error())))
		log.Println("Error:", err)
		return
	}

	if req.FormValue("format") == "json" {
		writeJSONChunks(results, w)
		return
	}

	switch req.Header.Get("Content-Type") {
	case "application/json":
		writeJSONChunks(results, w)
	default:
		writeLineResults(results, w)
	}
}

func iterateResults(r execute.Result, f func(measurement, fieldName string, tags map[string]string, value interface{}, t time.Time)) {
	blocks := r.Blocks()

	err := blocks.Do(func(b execute.Block) {

		times := b.Times()
		times.DoTime(func(ts []execute.Time, rr execute.RowReader) {
			for i, time := range ts {
				var measurement, fieldName string
				tags := map[string]string{}
				var value interface{}

				for j, c := range rr.Cols() {
					if c.IsTag {
						if c.Label == "_measurement" {
							measurement = rr.AtString(i, j)
						} else if c.Label == "_field" {
							fieldName = rr.AtString(i, j)
						} else {
							tags[c.Label] = rr.AtString(i, j)
						}
					} else {
						switch c.Type {
						case execute.TTime:
							value = rr.AtTime(i, j)
						case execute.TString:
							value = rr.AtString(i, j)
						case execute.TFloat:
							value = rr.AtFloat(i, j)
						case execute.TInt:
							value = "int not supported"
						default:
							value = "unknown"
						}
					}
				}

				f(measurement, fieldName, tags, value, time.Time())
			}
		})
	})
	if err != nil {
		log.Println("Error iterating through results:", err)
	}
}

type header struct {
	SeriesID int64             `json:"seriesID"`
	Tags     map[string]string `json:"tags"`
}

type chunk struct {
	Points []point `json:"points"`
}

type point struct {
	Value   interface{}       `json:"value"`
	Time    int64             `json:"time"`
	Context map[string]string `json:"context,omitempty"`
}

func writeJSONChunks(results []execute.Result, w http.ResponseWriter) {
	seriesID := int64(0)
	for _, r := range results {
		blocks := r.Blocks()

		err := blocks.Do(func(b execute.Block) {
			seriesID++

			// output header
			h := header{SeriesID: seriesID, Tags: b.Tags()}
			bb, err := json.Marshal(h)
			if err != nil {
				log.Println("error marshaling header: ", err.Error())
				return
			}
			_, err = w.Write(bb)
			if err != nil {
				log.Println("error writing header: ", err.Error())
				return
			}
			_, err = w.Write([]byte("\n"))
			if err != nil {
				log.Println("error writing newline: ", err.Error())
			}

			times := b.Times()
			times.DoTime(func(ts []execute.Time, rr execute.RowReader) {
				ch := chunk{Points: make([]point, len(ts))}
				for i, time := range ts {
					ch.Points[i].Time = time.Time().UnixNano()

					for j, c := range rr.Cols() {
						if !c.IsCommon && c.Type == execute.TString {
							if ch.Points[i].Context == nil {
								ch.Points[i].Context = make(map[string]string)
							}
							ch.Points[i].Context[c.Label] = rr.AtString(i, j)
						} else if !c.IsTag && c.Type != execute.TTime {
							switch c.Type {
							case execute.TFloat:
								ch.Points[i].Value = rr.AtFloat(i, j)
							case execute.TInt:
								ch.Points[i].Value = rr.AtInt(i, j)
							case execute.TString:
								ch.Points[i].Value = rr.AtString(i, j)
							case execute.TUInt:
								ch.Points[i].Value = rr.AtUInt(i, j)
							case execute.TBool:
								ch.Points[i].Value = rr.AtBool(i, j)
							default:
								ch.Points[i].Value = "unknown"
							}
						}
					}
				}

				// write it out
				b, err := json.Marshal(ch)
				if err != nil {
					log.Println("error marshaling chunk: ", err.Error())
					return
				}
				_, err = w.Write(b)
				if err != nil {
					log.Println("error writing chunk: ", err.Error())
					return
				}
				_, err = w.Write([]byte("\n"))
				if err != nil {
					log.Println("error writing newline: ", err.Error())
					return
				}
				w.(http.Flusher).Flush()
			})
		})
		if err != nil {
			log.Println("Error iterating through results:", err)
		}
	}
}

func writeJSONResults(results []execute.Result, w http.ResponseWriter) {

}

func writeLineResults(results []execute.Result, w http.ResponseWriter) {
	for _, r := range results {
		iterateResults(r, func(m, f string, tags map[string]string, val interface{}, t time.Time) {
			p, err := models.NewPoint(m, models.NewTags(tags), map[string]interface{}{f: val}, t)
			if err != nil {
				log.Println("error creating new point", err)
				return
			}
			w.Write([]byte(p.String()))
			w.Write([]byte("\n"))
		})
	}
}
