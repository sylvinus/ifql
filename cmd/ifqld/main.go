package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/influxdata/ifql"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/influxdb/models"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/query", http.HandlerFunc(HandleQuery))
	log.Printf("Starting on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
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

	ctx := context.Background()
	results, err := ifql.Query(ctx, query, verbose, trace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error executing query %s", err.Error())))
		log.Println("Error:", err)
		return
	}

	if req.FormValue("format") == "line" {
		writeLineResults(results, w)
	} else {
		writeJSONResults(results, w)
	}
}

func iterateResults(results []execute.Result, f func(measurement, fieldName string, tags map[string]string, value interface{}, t time.Time)) {
	for _, r := range results {
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
			fmt.Println("Error iterating through results:", err)
		}
	}
}

func writeJSONResults(results []execute.Result, w http.ResponseWriter) {

}

func writeLineResults(results []execute.Result, w http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)

	iterateResults(results, func(m, f string, tags map[string]string, val interface{}, t) {
							p, err := models.NewPoint(m, models.NewTags(tags), map[string]interface{}{f: val}, t)
					if err != nil {
						log.Println("error creating new point", err)
						return
					}
					w.Write([]byte(p.String()))
					w.Write([]byte("\n"))

		})
}
