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
	"github.com/influxdata/ifql/promql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/pkg/errors"
)

var queryStr = flag.String("query", `select(database:"mydb").where(exp:{"_measurement" == "m0"}).range(start:-170h).sum()`, "Query to run")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var verbose = flag.Bool("v", false, "print verbose output")
var trace = flag.Bool("trace", false, "print trace output")

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

	ctx := context.Background()
	results, err := doQuery(ctx, *queryStr, *verbose, *trace)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	for _, r := range results {
		blocks := r.Blocks()
		err := blocks.Do(func(b execute.Block) {
			execute.NewFormatter(b, nil).WriteTo(os.Stdout)
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

func ifqlSpec(query string) (*query.QuerySpec, error) {
	return ifql.NewQuery(query)
}

func promqlSpec(query string) (*query.QuerySpec, error) {
	return promql.Build(query)
}

func doQuery(ctx context.Context, queryStr string, verbose, trace bool) ([]execute.Result, error) {
	fmt.Println("Running query", queryStr)
	qSpec, err := ifql.NewQuery(queryStr)
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
