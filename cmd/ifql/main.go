package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ifql <query>")
		fmt.Println(os.Args)
		os.Exit(1)
	}
	queryStr := os.Args[1]

	results, err := doQuery(queryStr)
	if err != nil {
		fmt.Println("E!", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", results)
}

func doQuery(queryStr string) ([]execute.DataFrame, error) {
	var qSpec query.QuerySpec
	// TODO parse query
	//qSpec = parser.Parse(q)
	err := json.Unmarshal([]byte(queryStr), &qSpec)
	if err != nil {
		return nil, err
	}

	aplanner := plan.NewAbstractPlanner()
	ap, err := aplanner.Plan(&qSpec)
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
