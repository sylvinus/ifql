package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

const MapKind = "map"

type MapOpSpec struct {
	Fn *ast.ArrowFunctionExpression `json:"fn"`
}

func init() {
	query.RegisterMethod(MapKind, createMapOpSpec)
	query.RegisterOpSpec(MapKind, newMapOp)
	plan.RegisterProcedureSpec(MapKind, newMapProcedure, MapKind)
	execute.RegisterTransformation(MapKind, createMapTransformation)
}

func createMapOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	resolved, err := f.Resolve()
	if err != nil {
		return nil, err
	}

	return &MapOpSpec{
		Fn: resolved,
	}, nil
}
func newMapOp() query.OperationSpec {
	return new(MapOpSpec)
}

func (s *MapOpSpec) Kind() query.OperationKind {
	return MapKind
}

type MapProcedureSpec struct {
	Fn *ast.ArrowFunctionExpression
}

func newMapProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MapOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &MapProcedureSpec{
		Fn: spec.Fn,
	}, nil
}

func (s *MapProcedureSpec) Kind() plan.ProcedureKind {
	return MapKind
}
func (s *MapProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MapProcedureSpec)
	ns.Fn = s.Fn.Copy().(*ast.ArrowFunctionExpression)
	return ns
}

func createMapTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MapProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewBlockBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewMapTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type mapTransformation struct {
	d     execute.Dataset
	cache execute.BlockBuilderCache

	references []execute.Reference
	scope      execute.Scope
	scopeCols  map[string]int
	ces        map[execute.DataType]expressionOrError
}

func NewMapTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *MapProcedureSpec) (*mapTransformation, error) {
	if len(spec.Fn.Params) != 1 {
		return nil, fmt.Errorf("map functions should only have a single parameter, got %v", spec.Fn.Params)
	}
	objectName := spec.Fn.Params[0].Name
	references, err := execute.FindReferences(spec.Fn)
	if err != nil {
		return nil, err
	}

	valueRP := execute.Reference{objectName, "_value"}.Path()
	types := make(map[execute.ReferencePath]execute.DataType, len(references))
	for _, r := range references {
		if len(r) != 2 {
			return nil, fmt.Errorf("found invalid reference in the map function %q", r)
		}
		rp := r.Path()
		if rp != valueRP {
			types[rp] = execute.TString
		}
	}

	ces := make(map[execute.DataType]expressionOrError, len(execute.ValueDataTypes))
	for _, typ := range execute.ValueDataTypes {
		types[valueRP] = typ
		ce, err := execute.CompileExpression(spec.Fn, types)
		ces[typ] = expressionOrError{
			Err:  err,
			Expr: ce,
		}
	}

	return &mapTransformation{
		d:          d,
		cache:      cache,
		references: references,
		scope:      make(execute.Scope),
		scopeCols:  make(map[string]int),
		ces:        ces,
	}, nil
}

func (t *mapTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *mapTransformation) Process(id execute.DatasetID, b execute.Block) error {
	// Check that the function supports the col type.
	cols := b.Cols()
	valueIdx := execute.ValueIdx(cols)
	valueCol := cols[valueIdx]
	exprErr := t.ces[valueCol.Type]
	if exprErr.Err != nil {
		return errors.Wrapf(exprErr.Err, "expression does not support type %v", valueCol.Type)
	}
	ce := exprErr.Expr

	builder, new := t.cache.BlockBuilder(b)
	if new {
		// Add columns to builder
		for j, c := range cols {
			if j == valueIdx {
				builder.AddCol(execute.ColMeta{
					Label: execute.ValueColLabel,
					Type:  ce.Type(),
				})
				continue
			}

			builder.AddCol(c)
			if c.IsTag && c.IsCommon {
				builder.SetCommonString(j, b.Tags()[c.Label])
			}
		}
	}

	// Prepare scope
	for j, c := range cols {
		if c.Label == execute.ValueColLabel {
			t.scopeCols["_value"] = valueIdx
		} else {
			for _, r := range t.references {
				if r[1] == c.Label {
					t.scopeCols[c.Label] = j
					break
				}
			}
		}
	}

	// Append modified rows
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			for _, r := range t.references {
				t.scope[r.Path()] = execute.ValueForRow(i, t.scopeCols[r[1]], rr)
			}
			v, err := ce.Eval(t.scope)
			if err != nil {
				log.Printf("failed to evaluate map expression: %v", err)
				continue
			}
			for j, c := range cols {
				if c.IsCommon {
					continue
				}
				if j == valueIdx {
					switch val := v.Value.(type) {
					case bool:
						builder.AppendBool(j, val)
					case int64:
						builder.AppendInt(j, val)
					case uint64:
						builder.AppendUInt(j, val)
					case float64:
						builder.AppendFloat(j, val)
					case string:
						builder.AppendString(j, val)
					case execute.Time:
						builder.AppendTime(j, val)
					default:
						execute.PanicUnknownType(v.Type)
					}
					continue
				}
				switch c.Type {
				case execute.TBool:
					builder.AppendBool(j, rr.AtBool(i, j))
				case execute.TInt:
					builder.AppendInt(j, rr.AtInt(i, j))
				case execute.TUInt:
					builder.AppendUInt(j, rr.AtUInt(i, j))
				case execute.TFloat:
					builder.AppendFloat(j, rr.AtFloat(i, j))
				case execute.TString:
					builder.AppendString(j, rr.AtString(i, j))
				case execute.TTime:
					builder.AppendTime(j, rr.AtTime(i, j))
				default:
					execute.PanicUnknownType(c.Type)
				}
			}
		}
	})
	return nil
}

func (t *mapTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *mapTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *mapTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
