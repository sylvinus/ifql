package execute_test

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
)

func TestAggregate_Process(t *testing.T) {
	// All test cases use a simple SumAgg
	testCases := []struct {
		name   string
		bounds execute.Bounds
		data   []*executetest.Block
		want   func(b execute.Bounds) []*executetest.Block
	}{
		{
			name: "single",
			bounds: execute.Bounds{
				Start: 0,
				Stop:  100,
			},
			data: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 0,
					Stop:  100,
				},
				ColMeta: []execute.ColMeta{
					{Label: "time", Type: execute.TTime},
					{Label: "value", Type: execute.TFloat},
				},
				Data: [][]interface{}{
					{execute.Time(0), 0.0},
					{execute.Time(10), 1.0},
					{execute.Time(20), 2.0},
					{execute.Time(30), 3.0},
					{execute.Time(40), 4.0},
					{execute.Time(50), 5.0},
					{execute.Time(60), 6.0},
					{execute.Time(70), 7.0},
					{execute.Time(80), 8.0},
					{execute.Time(90), 9.0},
				},
			}},
			want: func(b execute.Bounds) []*executetest.Block {
				return []*executetest.Block{{
					Bnds: b,
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), 45.0},
					},
				}}
			},
		},
		{
			name: "multiple blocks",
			bounds: execute.Bounds{
				Start: 0,
				Stop:  200,
			},
			data: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  100,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(0), 0.0},
						{execute.Time(10), 1.0},
						{execute.Time(20), 2.0},
						{execute.Time(30), 3.0},
						{execute.Time(40), 4.0},
						{execute.Time(50), 5.0},
						{execute.Time(60), 6.0},
						{execute.Time(70), 7.0},
						{execute.Time(80), 8.0},
						{execute.Time(90), 9.0},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 100,
						Stop:  200,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), 10.0},
						{execute.Time(110), 11.0},
						{execute.Time(120), 12.0},
						{execute.Time(130), 13.0},
						{execute.Time(140), 14.0},
						{execute.Time(150), 15.0},
						{execute.Time(160), 16.0},
						{execute.Time(170), 17.0},
						{execute.Time(180), 18.0},
						{execute.Time(190), 19.0},
					},
				},
			},
			want: func(b execute.Bounds) []*executetest.Block {
				return []*executetest.Block{{
					Bnds: b,
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(100), 45.0},
						{execute.Time(200), 145.0},
					},
				}}
			},
		},
		{
			name: "multiple blocks with tags",
			bounds: execute.Bounds{
				Start: 0,
				Stop:  200,
			},
			data: []*executetest.Block{
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  100,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(0), 0.0, "a"},
						{execute.Time(10), 1.0, "a"},
						{execute.Time(20), 2.0, "a"},
						{execute.Time(30), 3.0, "a"},
						{execute.Time(40), 4.0, "a"},
						{execute.Time(50), 5.0, "a"},
						{execute.Time(60), 6.0, "a"},
						{execute.Time(70), 7.0, "a"},
						{execute.Time(80), 8.0, "a"},
						{execute.Time(90), 9.0, "a"},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 0,
						Stop:  100,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(0), 0.3, "b"},
						{execute.Time(10), 1.3, "b"},
						{execute.Time(20), 2.3, "b"},
						{execute.Time(30), 3.3, "b"},
						{execute.Time(40), 4.3, "b"},
						{execute.Time(50), 5.3, "b"},
						{execute.Time(60), 6.3, "b"},
						{execute.Time(70), 7.3, "b"},
						{execute.Time(80), 8.3, "b"},
						{execute.Time(90), 9.3, "b"},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 100,
						Stop:  200,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(100), 10.0, "a"},
						{execute.Time(110), 11.0, "a"},
						{execute.Time(120), 12.0, "a"},
						{execute.Time(130), 13.0, "a"},
						{execute.Time(140), 14.0, "a"},
						{execute.Time(150), 15.0, "a"},
						{execute.Time(160), 16.0, "a"},
						{execute.Time(170), 17.0, "a"},
						{execute.Time(180), 18.0, "a"},
						{execute.Time(190), 19.0, "a"},
					},
				},
				{
					Bnds: execute.Bounds{
						Start: 100,
						Stop:  200,
					},
					ColMeta: []execute.ColMeta{
						{Label: "time", Type: execute.TTime},
						{Label: "value", Type: execute.TFloat},
						{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
					},
					Data: [][]interface{}{
						{execute.Time(100), 10.3, "b"},
						{execute.Time(110), 11.3, "b"},
						{execute.Time(120), 12.3, "b"},
						{execute.Time(130), 13.3, "b"},
						{execute.Time(140), 14.3, "b"},
						{execute.Time(150), 15.3, "b"},
						{execute.Time(160), 16.3, "b"},
						{execute.Time(170), 17.3, "b"},
						{execute.Time(180), 18.3, "b"},
						{execute.Time(190), 19.3, "b"},
					},
				},
			},
			want: func(b execute.Bounds) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: b,
						ColMeta: []execute.ColMeta{
							{Label: "time", Type: execute.TTime},
							{Label: "value", Type: execute.TFloat},
							{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
						},
						Data: [][]interface{}{
							{execute.Time(100), 45.0, "a"},
							{execute.Time(200), 145.0, "a"},
						},
					},
					{
						Bnds: b,
						ColMeta: []execute.ColMeta{
							{Label: "time", Type: execute.TTime},
							{Label: "value", Type: execute.TFloat},
							{Label: "t1", Type: execute.TString, IsTag: true, IsCommon: true},
						},
						Data: [][]interface{}{
							{execute.Time(100), 48.0, "b"},
							{execute.Time(200), 148.0, "b"},
						},
					},
				}
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewBlockBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(execute.DefaultTriggerSpec)

			agg := execute.NewAggregateTransformation(d, c, tc.bounds, new(functions.SumAgg))

			parentID := executetest.RandomDatasetID()
			for _, b := range tc.data {
				if err := agg.Process(parentID, b); err != nil {
					t.Fatal(err)
				}
			}

			want := tc.want(tc.bounds)
			got := executetest.BlocksFromCache(c)

			sort.Sort(executetest.SortedBlocks(got))
			sort.Sort(executetest.SortedBlocks(want))

			if !cmp.Equal(want, got, cmpopts.EquateNaNs()) {
				t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
			}
		})
	}
}
