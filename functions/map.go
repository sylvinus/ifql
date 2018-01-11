package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
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

	fn *execute.RowMapFn
}

func NewMapTransformation(d execute.Dataset, cache execute.BlockBuilderCache, spec *MapProcedureSpec) (*mapTransformation, error) {
	fn, err := execute.NewRowMapFn(spec.Fn)
	if err != nil {
		return nil, err
	}
	return &mapTransformation{
		d:     d,
		cache: cache,
		fn:    fn,
	}, nil
}

func (t *mapTransformation) RetractBlock(id execute.DatasetID, meta execute.BlockMetadata) error {
	return t.d.RetractBlock(execute.ToBlockKey(meta))
}

func (t *mapTransformation) Process(id execute.DatasetID, b execute.Block) error {
	// Prepare the functions for the column types.
	cols := b.Cols()
	err := t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	builder, new := t.cache.BlockBuilder(b)
	if !new {
		return fmt.Errorf("received duplicate block bounds: %v tags: %v", b.Bounds(), b.Tags())
	}

	// Add tag columns to builder
	colMap := make([]int, 0, len(cols))
	for j, c := range cols {
		if !c.IsValue() {
			nj := builder.AddCol(c)
			if c.Common {
				builder.SetCommonString(nj, b.Tags()[c.Label])
			}
			colMap = append(colMap, j)
		}
	}

	meta := t.fn.MapMeta()
	// Add new value columns
	for _, p := range meta.Properties {
		builder.AddCol(execute.ColMeta{
			Label: p.Key,
			Type:  p.Type,
			Kind:  execute.ValueColKind,
		})
	}

	bCols := builder.Cols()
	// Append modified rows
	b.Times().DoTime(func(ts []execute.Time, rr execute.RowReader) {
		for i := range ts {
			m, err := t.fn.Eval(i, rr)
			if err != nil {
				log.Printf("failed to evaluate map expression: %v", err)
				continue
			}
			for j, c := range bCols {
				if c.Common {
					// We already set the common tag values
					continue
				}
				switch c.Kind {
				case execute.TimeColKind:
					builder.AppendTime(j, rr.AtTime(i, colMap[j]))
				case execute.TagColKind:
					builder.AppendString(j, rr.AtString(i, colMap[j]))
				case execute.ValueColKind:
					v := m.Values[c.Label]
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
				default:
					log.Printf("unknown column kind %v", c.Kind)
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
