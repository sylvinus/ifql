package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/promql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

var queryStr = flag.String("query", `select(database:"mydb").where(exp:{"_measurement" == "m0"}).range(start:-170h).sum()`, "Query to run")

func main() {

	flag.Parse()
	results, err := doQuery(*queryStr)
	if err != nil {
		fmt.Println("E!", err)
		os.Exit(1)
	}
	for _, r := range results {
		blocks := r.Blocks()
		blocks.Do(func(b execute.Block) {
			fmt.Printf("Block Tags: %v Bounds: %v\nTime:\tValue:\tTags:\n", b.Tags(), b.Bounds())
			cells := b.Cells()
			cells.Do(func(cs []execute.Cell) {
				for _, c := range cs {
					fmt.Print(c.Time)
					fmt.Print("\t")
					fmt.Print(c.Value)
					fmt.Print("\t")
					fmt.Print(c.Tags)
					fmt.Println()
				}
			})
		})
	}

}

func ifqlSpec(query string) (*query.QuerySpec, error) {
	return ifql.NewQuery(query)
}

func promqlSpec(query string) (*query.QuerySpec, error) {
	return promql.Build(query)
}

func doQuery(queryStr string) ([]execute.Result, error) {
	fmt.Println("Running query", queryStr)
	qSpec, err := ifql.NewQuery(queryStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse query")
	}

	lplanner := plan.NewLogicalPlanner()
	lp, err := lplanner.Plan(qSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create abstract plan")
	}

	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil, time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create concrete plan")
	}

	storage, err := execute.NewStorageReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create storage reader")
	}

	executor := execute.NewExecutor(storage)
	r, err := executor.Execute(context.Background(), p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	return r, nil
}
