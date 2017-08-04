package execute

import "github.com/influxdata/ifql/query/plan"

type Executor interface {
	Execute(plan.Plan) ([]Result, error)
}
