package functions_test

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
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

func TestFixedWindow_PassThrough(t *testing.T) {
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

func TestFixedWindow_Process(t *testing.T) {
	testCases := []struct {
		name          string
		start         execute.Time
		every, period execute.Duration
		num           int
		want          func(start execute.Time) []*executetest.Block
	}{
		{
			name: "nonoverlapping_nonaligned",
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
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
			},
		},
		{
			name: "nonoverlapping_aligned",
			// Use a time that is aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
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
			},
		},
		{
			name: "overlapping_nonaligned",
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(2 * time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
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
							Start: start,
							Stop:  start + execute.Time(2*time.Minute),
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
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
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
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
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
			},
		},
		{
			name: "overlapping_aligned",
			// Use a time that is aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(2 * time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
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
							Start: start,
							Stop:  start + execute.Time(2*time.Minute),
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
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
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
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
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
			},
		},
		{
			name: "underlapping_nonaligned",
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(2 * time.Minute),
			period: execute.Duration(time.Minute),
			num:    24,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start + 1*execute.Time(time.Minute),
							Stop:  start + 2*execute.Time(time.Minute),
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
							Start: start + execute.Time(3*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "time", Type: execute.TTime},
							{Label: "value", Type: execute.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(230*time.Second), 23.0},
						},
					},
				}
			},
		},
		{
			name: "underlapping_aligned",
			// Use a time that is  aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(2 * time.Minute),
			period: execute.Duration(time.Minute),
			num:    24,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start + 1*execute.Time(time.Minute),
							Stop:  start + 2*execute.Time(time.Minute),
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
							Start: start + execute.Time(3*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "time", Type: execute.TTime},
							{Label: "value", Type: execute.TFloat},
						},
						Data: [][]interface{}{
							{start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(230*time.Second), 23.0},
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			start := tc.start
			stop := start + execute.Time(time.Hour)

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
					Every:  tc.every,
					Period: tc.period,
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

			for i := 0; i < tc.num; i++ {
				block0.Data = append(block0.Data, []interface{}{
					start + execute.Time(time.Duration(i)*10*time.Second),
					float64(i),
				})
			}

			parentID := executetest.RandomDatasetID()
			fw.Process(parentID, block0)

			got := executetest.BlocksFromCache(c)

			sort.Sort(executetest.SortedBlocks(got))
			want := tc.want(start)
			sort.Sort(executetest.SortedBlocks(want))

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
			}
		})
	}
}
