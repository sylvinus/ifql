package executetest

import (
	"math"

	"github.com/influxdata/ifql/query/execute"
)

var UnlimitedAllocator = &execute.Allocator{
	Resource: &execute.Resource{
		Limit: math.MaxInt64,
	},
}
