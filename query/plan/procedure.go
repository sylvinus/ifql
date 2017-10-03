package plan

import (
	"fmt"
	"time"

	"github.com/influxdata/ifql/query"
	uuid "github.com/satori/go.uuid"
)

type ProcedureID uuid.UUID

func (id ProcedureID) String() string {
	return uuid.UUID(id).String()
}

var ZeroProcedureID ProcedureID

type Procedure struct {
	ID       ProcedureID
	Parents  []ProcedureID
	Children []ProcedureID
	Spec     ProcedureSpec
}

type CreateProcedureSpec func(query.OperationSpec) (ProcedureSpec, error)

// ProcedureSpec specifies an operation as part of a query.
type ProcedureSpec interface {
	// Kind returns the kind of the procedure.
	Kind() ProcedureKind
}

type PushDownProcedureSpec interface {
	PushDownRule() PushDownRule
	PushDown(root *Procedure)
}

type BoundedProcedureSpec interface {
	TimeBounds() BoundsSpec
}

// TODO(nathanielc): make this more formal using commute/associative properties
type PushDownRule struct {
	Root    ProcedureKind
	Through []ProcedureKind
}

// ProcedureKind denotes the kind of operations.
type ProcedureKind string

type BoundsSpec struct {
	Start query.Time
	Stop  query.Time
}

func (b BoundsSpec) Union(o BoundsSpec, now time.Time) (u BoundsSpec) {
	u.Start = b.Start
	if o.Start.Time(now).Before(b.Start.Time(now)) {
		u.Start = o.Start
	}
	u.Stop = b.Stop
	if o.Stop.Time(now).After(b.Stop.Time(now)) {
		u.Stop = o.Stop
	}
	return
}

type WindowSpec struct {
	Every  query.Duration
	Period query.Duration
	Round  query.Duration
	Start  query.Time
}

var kindToProcedure = make(map[ProcedureKind]CreateProcedureSpec)
var queryOpToProcedure = make(map[query.OperationKind][]CreateProcedureSpec)

// RegisterProcedureSpec registers a new procedure with the specified kind.
// The call panics if the kind is not unique.
func RegisterProcedureSpec(k ProcedureKind, c CreateProcedureSpec, qks ...query.OperationKind) {
	if kindToProcedure[k] != nil {
		panic(fmt.Errorf("duplicate registration for procedure kind %v", k))
	}
	kindToProcedure[k] = c
	for _, qk := range qks {
		queryOpToProcedure[qk] = append(queryOpToProcedure[qk], c)
	}
}
