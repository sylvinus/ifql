package execute_test

import (
	"testing"

	"github.com/influxdata/arrow/memory"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
)

func TestNewArrowBlockBuilder_MultipleBlocks(t *testing.T) {
	var (
		assert executetest.Assert
		alloc  = memory.NewCheckedAllocator(memory.NewGoAllocator())
	)
	defer alloc.AssertSize(t, 0)

	cols := []execute.ColMeta{
		{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
		{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
	}

	bb := execute.NewArrowBlockBuilder(alloc)
	bb.AddCol(cols[0])
	bb.AddCol(cols[1])

	var (
		expTimes  = []execute.Time{1000, 1010, 1020, 1030, 2000, 2010, 2020, 2030}
		expFloats = []float64{1.0, 1.1, 1.2, 1.3, 2.0, 2.1, 2.2, 2.3}
		gotTimes  []execute.Time
		gotFloats []float64
	)

	bb.AppendTimes(0, expTimes[:4])
	bb.AppendFloats(1, expFloats[:4])

	blk, _ := bb.Block()
	blk.Times().DoTime(func(v []execute.Time, rr execute.RowReader) {
		for i, ts := range v {
			gotTimes = append(gotTimes, ts)
			gotFloats = append(gotFloats, rr.AtFloat(i, 1))
		}
	})
	assert.Equal(t, expTimes[:4], gotTimes)
	assert.Equal(t, expFloats[:4], gotFloats)

	gotTimes, gotFloats = nil, nil

	bb.AppendTimes(0, expTimes[4:])
	bb.AppendFloats(1, expFloats[4:])

	blk, _ = bb.Block()
	blk.Times().DoTime(func(v []execute.Time, rr execute.RowReader) {
		for i, ts := range v {
			gotTimes = append(gotTimes, ts)
			gotFloats = append(gotFloats, rr.AtFloat(i, 1))
		}
	})
	assert.Equal(t, expTimes, gotTimes)
	assert.Equal(t, expFloats, gotFloats)

	bb.Release()
}
