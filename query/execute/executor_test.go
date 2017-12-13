package execute_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/plan"
)

var epoch = time.Unix(0, 0)

func TestExecutor_Execute(t *testing.T) {
	testCases := []struct {
		name string
		src  []execute.Block
		plan *plan.PlanSpec
		exp  [][]*executetest.Block
	}{
		{
			name: "simple aggregate",
			src: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					execute.TimeCol,
					execute.ColMeta{
						Label: execute.ValueColLabel,
						Type:  execute.TFloat,
					},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0},
					{execute.Time(1), 2.0},
					{execute.Time(2), 3.0},
					{execute.Time(3), 4.0},
					{execute.Time(4), 5.0},
				},
			}},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Bounds: plan.BoundsSpec{
					Start: query.Time{Absolute: time.Unix(0, 1)},
					Stop:  query.Time{Absolute: time.Unix(0, 5)},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									Relative:   -5,
									IsRelative: true,
								},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{}, Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Children: nil,
					},
				},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("sum"),
				},
			},
			exp: [][]*executetest.Block{[]*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					execute.TimeCol,
					execute.ColMeta{
						Label: execute.ValueColLabel,
						Type:  execute.TFloat,
					},
				},
				Data: [][]interface{}{
					{execute.Time(5), 15.0},
				},
			}}},
		},
		{
			name: "simple join",
			src: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					execute.TimeCol,
					execute.ColMeta{
						Label: execute.ValueColLabel,
						Type:  execute.TInt,
					},
				},
				Data: [][]interface{}{
					{execute.Time(0), int64(1)},
					{execute.Time(1), int64(2)},
					{execute.Time(2), int64(3)},
					{execute.Time(3), int64(4)},
					{execute.Time(4), int64(5)},
				},
			}},
			plan: &plan.PlanSpec{
				Now: epoch.Add(5),
				Resources: query.ResourceManagement{
					ConcurrencyQuota: 1,
					MemoryBytesQuota: math.MaxInt64,
				},
				Bounds: plan.BoundsSpec{
					Start: query.Time{Absolute: time.Unix(0, 1)},
					Stop:  query.Time{Absolute: time.Unix(0, 5)},
				},
				Procedures: map[plan.ProcedureID]*plan.Procedure{
					plan.ProcedureIDFromOperationID("from"): {
						ID: plan.ProcedureIDFromOperationID("from"),
						Spec: &functions.FromProcedureSpec{
							Database:  "mydb",
							BoundsSet: true,
							Bounds: plan.BoundsSpec{
								Start: query.Time{
									Relative:   -5,
									IsRelative: true,
								},
							},
						},
						Parents:  nil,
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("sum")},
					},
					plan.ProcedureIDFromOperationID("sum"): {
						ID:   plan.ProcedureIDFromOperationID("sum"),
						Spec: &functions.SumProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("count"): {
						ID:   plan.ProcedureIDFromOperationID("count"),
						Spec: &functions.CountProcedureSpec{},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("from"),
						},
						Children: []plan.ProcedureID{plan.ProcedureIDFromOperationID("join")},
					},
					plan.ProcedureIDFromOperationID("join"): {
						ID: plan.ProcedureIDFromOperationID("join"),
						Spec: &functions.MergeJoinProcedureSpec{
							Eval: &ast.ArrowFunctionExpression{
								Params: []*ast.Identifier{
									{Name: "a"},
									{Name: "b"},
								},
								Body: &ast.BinaryExpression{
									Operator: ast.DivisionOperator,
									Left: &ast.MemberExpression{
										Object: &ast.Identifier{
											Name: "a",
										},
										Property: &ast.StringLiteral{Value: "_value"},
									},
									Right: &ast.MemberExpression{
										Object: &ast.Identifier{
											Name: "b",
										},
										Property: &ast.StringLiteral{Value: "_value"},
									},
								},
							},
						},
						Parents: []plan.ProcedureID{
							plan.ProcedureIDFromOperationID("sum"),
							plan.ProcedureIDFromOperationID("count"),
						},
						Children: nil}},
				Results: []plan.ProcedureID{
					plan.ProcedureIDFromOperationID("join"),
				},
			},
			exp: [][]*executetest.Block{[]*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					execute.TimeCol,
					execute.ColMeta{
						Label: execute.ValueColLabel,
						Type:  execute.TInt,
					},
				},
				Data: [][]interface{}{
					{execute.Time(5), int64(3)},
				},
			}}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := execute.Config{
				StorageReader: &storageReader{blocks: tc.src},
			}
			exe := execute.NewExecutor(c)
			results, err := exe.Execute(context.Background(), tc.plan)
			if err != nil {
				t.Fatal(err)
			}
			got := make([][]*executetest.Block, len(results))
			for i, r := range results {
				if err := r.Blocks().Do(func(b execute.Block) error {
					got[i] = append(got[i], executetest.ConvertBlock(b))
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}

			if !cmp.Equal(got, tc.exp) {
				t.Error("unexpected results -want/+got", cmp.Diff(tc.exp, got))
			}
		})
	}
}

type storageReader struct {
	blocks []execute.Block
}

func (s storageReader) Close() {}
func (s storageReader) Read(context.Context, map[string]string, execute.ReadSpec, execute.Time, execute.Time) (execute.BlockIterator, error) {
	return &storageBlockIterator{
		s: s,
	}, nil
}

type storageBlockIterator struct {
	s storageReader
}

func (bi *storageBlockIterator) Do(f func(execute.Block) error) error {
	for _, b := range bi.s.blocks {
		if err := f(b); err != nil {
			return err
		}
	}
	return nil
}
