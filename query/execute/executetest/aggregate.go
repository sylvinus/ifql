package executetest

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query/execute"
)

func AggregateProcessTestHelper(t *testing.T, aggF execute.AggFunc, num int, wantValue float64) {
	t.Helper()

	start := execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano())
	stop := start + execute.Time(time.Hour)

	d := NewDataset(RandomDatasetID())
	c := execute.NewBlockBuilderCache()
	c.SetTriggerSpec(execute.DefaultTriggerSpec)

	bounds := execute.Bounds{
		Start: start,
		Stop:  stop,
	}

	agg := execute.NewAggregateTransformation(d, c, bounds, aggF)

	block0 := &Block{
		Bnds: execute.Bounds{
			Start: start,
			Stop:  stop,
		},
		ColMeta: []execute.ColMeta{
			{Label: "time", Type: execute.TTime},
			{Label: "value", Type: execute.TFloat},
		},
	}

	for i := 0; i < num; i++ {
		block0.Data = append(block0.Data, []interface{}{
			start + execute.Time(time.Duration(i)*10*time.Second),
			float64(i),
		})
	}

	parentID := RandomDatasetID()
	agg.Process(parentID, block0)

	got := BlocksFromCache(c)

	want := []*Block{{
		Bnds: execute.Bounds{
			Start: start,
			Stop:  stop,
		},
		ColMeta: []execute.ColMeta{
			{Label: "time", Type: execute.TTime},
			{Label: "value", Type: execute.TFloat},
		},
		Data: [][]interface{}{
			{stop, wantValue},
		},
	}}

	sort.Sort(SortedBlocks(got))
	sort.Sort(SortedBlocks(want))

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
	}
}
