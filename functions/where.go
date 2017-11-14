package functions

import (
	"errors"
	"fmt"
	"log"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const WhereKind = "where"

type WhereOpSpec struct {
	Expression expression.Expression `json:"expression"`
}

func init() {
	ifql.RegisterFunction(WhereKind, createWhereOpSpec)
	query.RegisterOpSpec(WhereKind, newWhereOp)
	plan.RegisterProcedureSpec(WhereKind, newWhereProcedure, WhereKind)
	execute.RegisterTransformation(WhereKind, createWhereTransformation)
}

func createWhereOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	expValue, ok := args["exp"]
	if !ok {
		return nil, errors.New(`where function requires an argument "exp"`)
	}
	if expValue.Type != ifql.TExpression {
		return nil, fmt.Errorf(`where function argument "exp" must be an expression, got %v`, expValue.Type)
	}

	return &WhereOpSpec{
		Expression: expression.Expression{
			Root: expValue.Value.(expression.Node),
		},
	}, nil
}
func newWhereOp() query.OperationSpec {
	return new(WhereOpSpec)
}

func (s *WhereOpSpec) Kind() query.OperationKind {
	return WhereKind
}

type WhereProcedureSpec struct {
	Expression expression.Expression
}

func newWhereProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*WhereOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &WhereProcedureSpec{
		Expression: spec.Expression,
	}, nil
}

func (s *WhereProcedureSpec) Kind() plan.ProcedureKind {
	return WhereKind
}
func (s *WhereProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WhereProcedureSpec)
	//TODO copy expression
	ns.Expression = s.Expression
	return ns
}

func (s *WhereProcedureSpec) PushDownRule() plan.PushDownRule {
	return plan.PushDownRule{
		Root:    SelectKind,
		Through: []plan.ProcedureKind{GroupKind, LimitKind, RangeKind},
	}
}
func (s *WhereProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*SelectProcedureSpec)
	if selectSpec.WhereSet {
		root = dup()
		selectSpec = root.Spec.(*SelectProcedureSpec)
		selectSpec.WhereSet = false
		selectSpec.Where = expression.Expression{}
		return
	}
	selectSpec.WhereSet = true
	selectSpec.Where = s.Expression
}

func createWhereTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, ctx execute.Context) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WhereProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache()
	d := execute.NewDataset(id, mode, cache)
	t := NewWhereTransformation(d, cache, s)
	return t, d, nil
}

type whereTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	names     []string
	scope     execute.Scope
	scopeCols map[string]int
	ces       map[execute.DataType]expressionOrError

	colMap []int
}

type expressionOrError struct {
	Err  error
	Expr execute.CompiledExpression
}

func NewWhereTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *WhereProcedureSpec) *whereTransformation {
	names := execute.ExpressionNames(spec.Expression.Root)
	types := make(map[string]execute.DataType, len(names))
	ces := make(map[execute.DataType]expressionOrError, len(execute.ValueDataTypes))
	for _, n := range names {
		if n != "$" {
			types[n] = execute.TString
		}
	}
	for _, typ := range execute.ValueDataTypes {
		types["$"] = typ
		ce, err := execute.CompileExpression(spec.Expression, types)
		ces[typ] = expressionOrError{
			Err:  err,
			Expr: ce,
		}
		if err == nil && ce.Type() != execute.TBool {
			ces[typ] = expressionOrError{
				Err:  errors.New("expression does not evaluate to boolean"),
				Expr: nil,
			}
		}
	}

	return &whereTransformation{
		d:         d,
		cache:     cache,
		names:     names,
		scope:     make(execute.Scope),
		scopeCols: make(map[string]int),
		ces:       ces,
	}
}

func (t *whereTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) {
	t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *whereTransformation) Process(id execute.DatasetID, b execute.Block) {
	builder, new := t.cache.BlockBuilder(b)
	if new {
		execute.AddBlockCols(b, builder)
	}

	ncols := builder.NCols()
	if cap(t.colMap) < ncols {
		t.colMap = make([]int, ncols)
		for j := range t.colMap {
			t.colMap[j] = j
		}
	} else {
		t.colMap = t.colMap[:ncols]
	}

	// Prepare scope
	cols := b.Cols()
	valueIdx := execute.ValueIdx(cols)
	for j, c := range cols {
		if c.Label == execute.ValueColLabel {
			t.scopeCols["$"] = valueIdx
		} else {
			for _, k := range t.names {
				if k == c.Label {
					t.scopeCols[c.Label] = j
					break
				}
			}
		}
	}

	valueCol := cols[valueIdx]
	exprErr := t.ces[valueCol.Type]
	if exprErr.Err != nil {
		log.Printf("expression does not support type %v: %v", valueCol.Type, exprErr.Err)
		return
	}
	ce := exprErr.Expr

	// Append only matching rows to block
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			for _, k := range t.names {
				t.scope[k] = execute.ValueForRow(i, t.scopeCols[k], rr)
			}
			if pass, err := ce.EvalBool(t.scope); !pass {
				if err != nil {
					log.Printf("failed to evaluate expression: %v", err)
				}
				continue
			}
			for j, c := range cols {
				if c.IsCommon {
					continue
				}
				switch c.Type {
				case execute.TBool:
					builder.AppendBool(j, rr.AtBool(i, t.colMap[j]))
				case execute.TInt:
					builder.AppendInt(j, rr.AtInt(i, t.colMap[j]))
				case execute.TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, t.colMap[j]))
				case execute.TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, t.colMap[j]))
				case execute.TString:
					builder.AppendString(j, rr.AtString(i, t.colMap[j]))
				case execute.TTime:
					builder.AppendTime(j, rr.AtTime(i, t.colMap[j]))
				default:
					execute.PanicUnknownType(c.Type)
				}
			}
		}
	})
}

func (t *whereTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) {
	t.d.UpdateWatermark(mark)
}
func (t *whereTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) {
	t.d.UpdateProcessingTime(pt)
}
func (t *whereTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
func (t *whereTransformation) SetParents(ids []execute.DatasetID) {
}
