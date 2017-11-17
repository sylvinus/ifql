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
	Database string   `json:"database"`
	Hosts    []string `json:"hosts"`
}

func init() {
	ifql.RegisterFunction(SelectKind, createSelectOpSpec)
	query.RegisterOpSpec(SelectKind, newSelectOp)
	plan.RegisterProcedureSpec(SelectKind, newSelectProcedure, SelectKind)
	execute.RegisterSource(SelectKind, createSelectSource)
}

func createSelectOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(SelectOpSpec)
	if value, ok := args["db"]; ok {
		if value.Type != ifql.TString {
			return nil, fmt.Errorf(`select function "db" argument must be a string, got %v`, value.Type)
		}
		spec.Database = value.Value.(string)
	} else {
		return nil, errors.New(`select function requires the "db" argument`)
	}

	if value, ok := args["hosts"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf(`select function "hosts" argument must be a list of strings, got %v. Example select(hosts:["a:8082", "b:8082"]).`, value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf(`select function "hosts" argument must be a list of strings, got list of %v. Example select(hosts:["a:8082", "b:8082"]).`, list.Type)
		}
		spec.Hosts = list.Elements.([]string)
	}
	return spec, nil
}

func newSelectOp() query.OperationSpec {
	return new(SelectOpSpec)
}

func (s *SelectOpSpec) Kind() query.OperationKind {
	return SelectKind
}

type SelectProcedureSpec struct {
	Database string
	Hosts    []string

	BoundsSet bool
	Bounds    plan.BoundsSpec

	FilterSet bool
	Filter    expression.Expression

	DescendingSet bool
	Descending    bool

	LimitSet     bool
	PointsLimit  int64
	SeriesLimit  int64
	SeriesOffset int64

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
	spec, ok := qs.(*SelectOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &SelectProcedureSpec{
		Database: spec.Database,
		Hosts:    spec.Hosts,
	}, nil
}

func (s *SelectProcedureSpec) Kind() plan.ProcedureKind {
	return SelectKind
}
func (s *SelectProcedureSpec) TimeBounds() plan.BoundsSpec {
	return s.Bounds
}
func (s *SelectProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SelectProcedureSpec)

	ns.Database = s.Database

	if len(s.Hosts) > 0 {
		ns.Hosts = make([]string, len(s.Hosts))
		copy(ns.Hosts, s.Hosts)
	}

	ns.BoundsSet = s.BoundsSet
	ns.Bounds = s.Bounds

	ns.FilterSet = s.FilterSet
	// TODO copy predicate
	ns.Filter = s.Filter

	ns.DescendingSet = s.DescendingSet
	ns.Descending = s.Descending

	ns.LimitSet = s.LimitSet
	ns.PointsLimit = s.PointsLimit
	ns.SeriesLimit = s.SeriesLimit
	ns.SeriesOffset = s.SeriesOffset

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
			Hosts:         spec.Hosts,
			Predicate:     spec.Filter,
			PointsLimit:   spec.PointsLimit,
			SeriesLimit:   spec.SeriesLimit,
			SeriesOffset:  spec.SeriesOffset,
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
