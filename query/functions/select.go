package functions

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/storage"
	"github.com/influxdata/ifql/query/plan"
)

const SelectKind = "select"

type SelectOpSpec struct {
	Database string `json:"database"`
}

func init() {
	query.RegisterOpSpec(SelectKind, newSelectOp)
	plan.RegisterProcedureSpec(SelectKind, newSelectProcedure, SelectKind)
	execute.RegisterSource(SelectKind, createSelectSource)
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
	Where    *storage.Predicate

	DescendingSet bool
	Descending    bool

	LimitSet bool
	Limit    int64
	Offset   int64

	WindowSet bool
	Window    plan.WindowSpec
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

func createSelectSource(prSpec plan.ProcedureSpec, id execute.DatasetID, sr execute.StorageReader, now time.Time) execute.Source {
	spec := prSpec.(*SelectProcedureSpec)
	var w execute.Window
	if spec.WindowSet {
		w = execute.Window{
			Every:  execute.Duration(spec.Window.Every),
			Period: execute.Duration(spec.Window.Period),
			Round:  execute.Duration(spec.Window.Round),
			Start:  execute.Time(spec.Window.Start.Time(now).UnixNano()),
		}
	} else {
		w = execute.Window{
			Every:  execute.Duration(spec.Bounds.Stop.Time(now).UnixNano() - spec.Bounds.Start.Time(now).UnixNano()),
			Period: execute.Duration(spec.Bounds.Stop.Time(now).UnixNano() - spec.Bounds.Start.Time(now).UnixNano()),
			Start:  execute.Time(spec.Bounds.Start.Time(now).UnixNano()),
		}
	}
	currentTime := w.Start + execute.Time(w.Period)
	bounds := execute.Bounds{
		Start: execute.Time(spec.Bounds.Start.Time(now).UnixNano()),
		Stop:  execute.Time(spec.Bounds.Stop.Time(now).UnixNano()),
	}
	return execute.NewStorageSource(
		id,
		sr,
		execute.ReadSpec{
			Database:   spec.Database,
			Predicate:  spec.Where,
			Limit:      spec.Limit,
			Descending: spec.Descending,
		},
		bounds,
		w,
		currentTime,
	)
}
