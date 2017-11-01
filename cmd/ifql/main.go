package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/influxql"
	"github.com/influxdata/ifql/promql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/pkg/errors"
)

var (
	queryStr   = flag.String("query", `select(database:"mydb").where(exp:{"_measurement" == "m0"}).range(start:-170h).sum()`, "Query to run")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
	verbose    = flag.Bool("v", false, "print verbose output")
	trace      = flag.Bool("trace", false, "print trace output")
	lang       = flag.String("lang", "ifql", "specify language for query option (ifql, promql, influxql)")
)

func main() {
	flag.Parse()

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

	var fn specFunc
	switch *lang {
	case "", "ifql":
		fn = ifqlSpec

	case "promql":
		fn = promqlSpec

	case "influxql":
		fn = influxQLSepc

	default:
		log.Fatalf("invalid query language: %s", *lang)
	}

	ctx := context.Background()
	results, err := doQuery(ctx, fn, *queryStr, *verbose, *trace)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	for _, r := range results {
		blocks := r.Blocks()
		err := blocks.Do(func(b execute.Block) {
			fmt.Printf("%v\n", execute.Formatted(b, execute.Squeeze()))
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

type specFunc func(string) (*query.QuerySpec, error)

func ifqlSpec(query string) (*query.QuerySpec, error) {
	return ifql.NewQuery(query)
}

func promqlSpec(query string) (*query.QuerySpec, error) {
	return promql.Build(query)
}

func influxQLSepc(query string) (*query.QuerySpec, error) {
	return influxql.ParseQuery(query)
}

func doQuery(ctx context.Context, fn specFunc, queryStr string, verbose, trace bool) ([]execute.Result, error) {
	fmt.Println("Running query", queryStr)
	qSpec, err := fn(queryStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
	}

	var opts []execute.Option
	if verbose {
		opts = append(opts, execute.Verbose())
	}
	if trace {
		opts = append(opts, execute.Trace())
	}
	return execute.Execute(ctx, qSpec, opts...)
}
