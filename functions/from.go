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

const FromKind = "from"

type FromOpSpec struct {
	Database string   `json:"database"`
	Hosts    []string `json:"hosts"`
}

func init() {
	ifql.RegisterFunction(FromKind, createFromOpSpec)
	query.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromKind, createFromSource)
}

func createFromOpSpec(args map[string]ifql.Value, ctx ifql.Context) (query.OperationSpec, error) {
	spec := new(FromOpSpec)
	if value, ok := args["db"]; ok {
		if value.Type != ifql.TString {
			return nil, fmt.Errorf(`from function "db" argument must be a string, got %v`, value.Type)
		}
		spec.Database = value.Value.(string)
	} else {
		return nil, errors.New(`from function requires the "db" argument`)
	}

	if value, ok := args["hosts"]; ok {
		if value.Type != ifql.TArray {
			return nil, fmt.Errorf(`from function "hosts" argument must be a list of strings, got %v. Example from(hosts:["a:8082", "b:8082"]).`, value.Type)
		}
		list := value.Value.(ifql.Array)
		if list.Type != ifql.TString {
			return nil, fmt.Errorf(`from function "hosts" argument must be a list of strings, got list of %v. Example from(hosts:["a:8082", "b:8082"]).`, list.Type)
		}
		spec.Hosts = list.Elements.([]string)
	}
	return spec, nil
}

func newFromOp() query.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() query.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
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
	GroupExcept []string
	GroupKeep   []string

	AggregateSet  bool
	AggregateType string
}

func newFromProcedure(qs query.OperationSpec) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Database: spec.Database,
		Hosts:    spec.Hosts,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}
func (s *FromProcedureSpec) TimeBounds() plan.BoundsSpec {
	return s.Bounds
}
func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)

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

func createFromSource(prSpec plan.ProcedureSpec, id execute.DatasetID, sr execute.StorageReader, ctx execute.Context) execute.Source {
	spec := prSpec.(*FromProcedureSpec)
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
			GroupExcept:   spec.GroupExcept,
			GroupKeep:     spec.GroupKeep,
			AggregateType: spec.AggregateType,
		},
		bounds,
		w,
		currentTime,
	)
}
