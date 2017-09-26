package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/promql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
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
			fmt.Println(execute.TableFmt{Block: b})
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

	return execute.Execute(qSpec)
}
