package executetest

import (
	"math"

	"github.com/influxdata/arrow/memory"
	"github.com/influxdata/ifql/query/execute"
)

var (
	UnlimitedResource                           = &execute.Resource{Limit: math.MaxInt64}
	UnlimitedAllocator        execute.Allocator = memory.NewGoAllocator()
	UnlimitedColListAllocator                   = execute.NewColListAllocator(UnlimitedAllocator)
)
