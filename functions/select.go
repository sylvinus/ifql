package functions

import (
	"errors"
	"fmt"

	"github.com/influxdata/ifql/expression"
	"github.com/influxdata/ifql/ifql"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SelectKind = "select"

type SelectOpSpec struct {
	Database string `json:"database"`
}

func init() {
	ifql.RegisterFunction(SelectKind, createSelectOpSpec)
	query.RegisterOpSpec(SelectKind, newSelectOp)
	plan.RegisterProcedureSpec(SelectKind, newSelectProcedure, SelectKind)
	execute.RegisterSource(SelectKind, createSelectSource)
}

func createSelectOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	dbValue, ok := args["db"]
	if !ok {
		return nil, errors.New(`select function requires the "db" argument`)
	}
	if dbValue.Type != ifql.TString {
		return nil, fmt.Errorf(`select function "db" argument must be a string, got %v`, dbValue.Type)
	}

	return &SelectOpSpec{
		Database: dbValue.Value.(string),
	}, nil
}

func newSelectOp() query.OperationSpec {
	return new(SelectOpSpec)
}

func (s *SelectOpSpec) Kind() query.OperationKind {
	return SelectKind
}

type SelectProcedureSpec struct {
	Database string

	BoundsSet bool
	Bounds    plan.BoundsSpec

	WhereSet bool
	Where    expression.Expression

	DescendingSet bool
	Descending    bool

	LimitSet bool
	Limit    int64
	Offset   int64

	WindowSet bool
	Window    plan.WindowSpec

	GroupingSet bool
	OrderByTime bool
	MergeAll    bool
	GroupKeys   []string
	GroupIgnore []string
	GroupKeep   []string

	AggregateSet  bool
	AggregateType string
}

func newSelectProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	s := new(SelectProcedureSpec)
	spec, ok := qs.(*SelectOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	s.Database = spec.Database
	return s, nil
}

func (s *SelectProcedureSpec) Kind() plan.ProcedureKind {
	return SelectKind
}
func (s *SelectProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SelectProcedureSpec)

	ns.Database = s.Database

	ns.BoundsSet = s.BoundsSet
	ns.Bounds = s.Bounds

	ns.WhereSet = s.WhereSet
	// TODO copy predicate
	ns.Where = s.Where

	ns.DescendingSet = s.DescendingSet
	ns.Descending = s.Descending

	ns.LimitSet = s.LimitSet
	ns.Limit = s.Limit
	ns.Offset = s.Offset

	ns.WindowSet = s.WindowSet
	ns.Window = s.Window

	ns.AggregateSet = s.AggregateSet
	ns.AggregateType = s.AggregateType

	return ns
}

func createSelectSource(prSpec plan.ProcedureSpec, id execute.DatasetID, sr execute.StorageReader, ctx execute.Context) execute.Source {
	spec := prSpec.(*SelectProcedureSpec)
	var w execute.Window
	if spec.WindowSet {
		w = execute.Window{
			Every:  execute.Duration(spec.Window.Every),
			Period: execute.Duration(spec.Window.Period),
			Round:  execute.Duration(spec.Window.Round),
			Start:  ctx.ResolveTime(spec.Window.Start),
		}
	} else {
		duration := execute.Duration(ctx.ResolveTime(spec.Bounds.Stop)) - execute.Duration(ctx.ResolveTime(spec.Bounds.Start))
		w = execute.Window{
			Every:  duration,
			Period: duration,
			Start:  ctx.ResolveTime(spec.Bounds.Start),
		}
	}
	currentTime := w.Start + execute.Time(w.Period)
	bounds := execute.Bounds{
		Start: ctx.ResolveTime(spec.Bounds.Start),
		Stop:  ctx.ResolveTime(spec.Bounds.Stop),
	}
	return execute.NewStorageSource(
		id,
		sr,
		execute.ReadSpec{
			Database:      spec.Database,
			Predicate:     spec.Where,
			Limit:         spec.Limit,
			Descending:    spec.Descending,
			OrderByTime:   spec.OrderByTime,
			MergeAll:      spec.MergeAll,
			GroupKeys:     spec.GroupKeys,
			GroupIgnore:   spec.GroupIgnore,
			GroupKeep:     spec.GroupKeep,
			AggregateType: spec.AggregateType,
		},
		bounds,
		w,
		currentTime,
	)
}
