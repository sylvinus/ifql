package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/promql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

func main() {
	//if len(os.Args) != 2 {
	//	fmt.Println("Usage: ifql <query>")
	//	fmt.Println(os.Args)
	//	os.Exit(1)
	//}
	//queryStr := os.Args[1]
	// 	queryStr := `{
	//   "operations": [
	//     {
	//       "id": "select",
	//       "kind": "select",
	//       "spec": {
	//         "database": "mydb"
	//       }
	//     },
	//     {
	//       "id": "range",
	//       "kind": "range",
	//       "spec": {
	//         "start": "-4h",
	//         "stop": "now"
	//       }
	//     },
	//     {
	//       "id": "sum",
	//       "kind": "sum"
	//     }
	//   ],
	//   "edges": [
	//     {
	//       "parent": "select",
	//       "child": "range"
	//     },
	//     {
	//       "parent": "range",
	//       "child": "sum"
	//     }
	//   ]
	// }`

	query := `select(database:"mydb").where(exp:{"_measurement" = "m0"}).range(start:-170h).sum()`
	qSpec, err := ifqlSpec(query)
	/*
		query := `sum(node_cpu{_measurement="m0"}[170h])`
		qSpec, err := promqlSpec(queryStr)
	*/
	if err != nil {
		fmt.Println("E!", err)
		os.Exit(1)
	}
	results, err := doQuery(qSpec)
	if err != nil {
		fmt.Println("E!", err)
		os.Exit(1)
	}
	fmt.Println(len(results))
	for _, r := range results {
		fmt.Println(r.ColsIndex())
		rows, _ := r.RowSlice(0)
		fmt.Println(rows)
	}

}

func ifqlSpec(query string) (*query.QuerySpec, error) {
	return ifql.NewQuery(query)
}

func promqlSpec(query string) (*query.QuerySpec, error) {
	return promql.Build(query)
}

func doQuery(qSpec *query.QuerySpec) ([]execute.DataFrame, error) {
	// var qSpec query.QuerySpec
	// // TODO parse query
	// //qSpec = parser.Parse(q)
	// err := json.Unmarshal([]byte(queryStr), &qSpec)
	// if err != nil {
	// 	return nil, err
	// }
	aplanner := plan.NewAbstractPlanner()
	ap, err := aplanner.Plan(qSpec)
	if err != nil {
		return nil, err
	}

	planner := plan.NewPlanner()
	p, err := planner.Plan(ap, nil, time.Now())
	if err != nil {
		return nil, err
	}

	storage, err := execute.NewStorageReader()
	if err != nil {
		return nil, err
	}

	executor := execute.NewExecutor(storage)
	return executor.Execute(context.Background(), p)
}
