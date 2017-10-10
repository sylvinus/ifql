package functions_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/functions"
	"github.com/influxdata/ifql/query/querytest"
)

func TestWindowOperation_Marshaling(t *testing.T) {
	//TODO: Test marshalling of triggerspec
	data := []byte(`{"id":"window","kind":"window","spec":{"every":"1m","period":"1h","start":"-4h","round":"1s"}}`)
	op := &query.Operation{
		ID: "window",
		Spec: &functions.WindowOpSpec{
			Every:  query.Duration(time.Minute),
			Period: query.Duration(time.Hour),
			Start: query.Time{
				Relative: -4 * time.Hour,
			},
			Round: query.Duration(time.Second),
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFixedWindowPassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
		fw := functions.NewFixedWindowTransformation(
			d,
			c,
			execute.Bounds{},
			execute.Window{
				Every:  execute.Duration(time.Minute),
				Period: execute.Duration(time.Minute),
			},
		)
		return fw
	})
}

func TestFixedWindowProcess(t *testing.T) {
	stop := execute.Now()
	start := stop - execute.Time(time.Hour)

	d := executetest.NewDataset(executetest.RandomDatasetID())
	c := execute.NewBlockBuilderCache()
	c.SetTriggerSpec(execute.DefaultTriggerSpec)

	fw := functions.NewFixedWindowTransformation(
		d,
		c,
		execute.Bounds{
			Start: start,
			Stop:  stop,
		},
		execute.Window{
			Every:  execute.Duration(time.Minute),
			Period: execute.Duration(time.Minute),
			Start:  start,
		},
	)

	block0 := &executetest.Block{
		Bnds: execute.Bounds{
			Start: start,
			Stop:  stop,
		},
		ColMeta: []execute.ColMeta{
			{Label: "time", Type: execute.TTime},
			{Label: "value", Type: execute.TFloat},
		},
	}

	for i := 0; i < 15; i++ {
		block0.Data = append(block0.Data, []interface{}{
			start + execute.Time(time.Duration(i)*10*time.Second),
			float64(i),
		})
	}

	blocks := []*executetest.Block{block0}
	want := []*executetest.Block{
		{
			Bnds: execute.Bounds{
				Start: start,
				Stop:  start + execute.Time(time.Minute),
			},
			ColMeta: []execute.ColMeta{
				{Label: "time", Type: execute.TTime},
				{Label: "value", Type: execute.TFloat},
			},
			Data: [][]interface{}{
				{start, 0.0},
				{start + execute.Time(10*time.Second), 1.0},
				{start + execute.Time(20*time.Second), 2.0},
				{start + execute.Time(30*time.Second), 3.0},
				{start + execute.Time(40*time.Second), 4.0},
				{start + execute.Time(50*time.Second), 5.0},
			},
		},
		{
			Bnds: execute.Bounds{
				Start: start + execute.Time(1*time.Minute),
				Stop:  start + execute.Time(2*time.Minute),
			},
			ColMeta: []execute.ColMeta{
				{Label: "time", Type: execute.TTime},
				{Label: "value", Type: execute.TFloat},
			},
			Data: [][]interface{}{
				{start + execute.Time(60*time.Second), 6.0},
				{start + execute.Time(70*time.Second), 7.0},
				{start + execute.Time(80*time.Second), 8.0},
				{start + execute.Time(90*time.Second), 9.0},
				{start + execute.Time(100*time.Second), 10.0},
				{start + execute.Time(110*time.Second), 11.0},
			},
		},
		{
			Bnds: execute.Bounds{
				Start: start + execute.Time(2*time.Minute),
				Stop:  start + execute.Time(3*time.Minute),
			},
			ColMeta: []execute.ColMeta{
				{Label: "time", Type: execute.TTime},
				{Label: "value", Type: execute.TFloat},
			},
			Data: [][]interface{}{
				{start + execute.Time(120*time.Second), 12.0},
				{start + execute.Time(130*time.Second), 13.0},
				{start + execute.Time(140*time.Second), 14.0},
			},
		},
	}

	parentID := executetest.RandomDatasetID()
	for _, b := range blocks {
		fw.Process(parentID, b)
	}

	got := executetest.BlocksFromCache(c)

	if !cmp.Equal(want, got) {
		t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
	}
}
