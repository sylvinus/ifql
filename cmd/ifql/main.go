package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/influxdata/ifql"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/tracing"
	"github.com/opentracing/opentracing-go"
)

var version string
var commit string
var date string

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var verbose = flag.Bool("v", false, "print verbose output")
var trace = flag.Bool("trace", false, "print trace output")

var hosts = make(hostList, 0)

func init() {
	flag.Var(&hosts, "host", "An InfluxDB host to connect to. Can be provided multiple times.")
}

type hostList []string

func (l *hostList) String() string {
	return "<host>..."
}

func (l *hostList) Set(s string) error {
	*l = append(*l, s)
	return nil
}

var defaultStorageHosts = []string{"localhost:8082"}

func usage() {
	fmt.Println("Usage: ifql [OPTIONS] <query>")
	fmt.Println()
	fmt.Println("Runs a query using the IFQL engine.")
	fmt.Println()
	fmt.Println("The query argument is either a string query \nor a path to a file prefixed with an '@'.")
	fmt.Println()
	fmt.Println("Options:")

	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if tr := tracing.Open("ifql"); tr != nil {
		defer tr.Close()
	}

	// Start cpuprofile
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	ctx := context.Background()
	span := opentracing.StartSpan("query")
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	var queryStr string
	if args := flag.Args(); len(args) == 1 {
		q, err := loadQuery(args[0])
		if err != nil {
			log.Fatal(err)
		}
		queryStr = q
	} else {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Running query:\n", queryStr)
	if len(hosts) == 0 {
		hosts = defaultStorageHosts
	}
	c, err := ifql.NewController(ifql.Config{
		Hosts:            hosts,
		ConcurrencyQuota: runtime.NumCPU() * 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	q, err := c.QueryWithCompile(ctx, queryStr)
	if err != nil {
		log.Fatal(err)
	}
	defer q.Done()

	if *verbose {
		octets, err := json.MarshalIndent(q.Spec, "", "    ")
		if err != nil {
			fmt.Println(string(octets))
		}
	}

	results, ok := <-q.Ready
	if !ok {
		err := q.Err()
		log.Fatal(err)
	}

	for _, r := range results {
		blocks := r.Blocks()
		err := blocks.Do(func(b execute.Block) error {
			execute.NewFormatter(b, nil).WriteTo(os.Stdout)
			return nil
		})
		if err != nil {
			fmt.Println("Error:", err)
		}
	}

	// Write out memprofile
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func loadQuery(q string) (string, error) {
	if len(q) > 0 && q[0] == '@' {
		f, err := os.Open(q[1:])
		if err != nil {
			return "", err
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		q = string(data)
	}
	return q, nil
}
