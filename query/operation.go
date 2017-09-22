package query

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// Operation denotes a single operation in a query.
type Operation struct {
	ID   OperationID   `json:"id"`
	Spec OperationSpec `json:"spec"`
}

func (o *Operation) UnmarshalJSON(data []byte) error {
	type operationJSON struct {
		ID   OperationID     `json:"id"`
		Kind OperationKind   `json:"kind"`
		Spec json.RawMessage `json:"spec"`
	}
	oj := operationJSON{}
	err := json.Unmarshal(data, &oj)
	if err != nil {
		return err
	}
	o.ID = oj.ID
	spec, err := unmarshalOpSpec(oj.Kind, oj.Spec)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal operation %q", oj.ID)
	}
	o.Spec = spec
	return nil
}

func unmarshalOpSpec(k OperationKind, data []byte) (OperationSpec, error) {
	createOpSpec, ok := kindToOp[k]
	if !ok {
		return nil, fmt.Errorf("unknown operation spec kind %v", k)
	}
	spec := createOpSpec()

	if len(data) > 0 {
		err := json.Unmarshal(data, spec)
		if err != nil {
			return nil, err
		}
	}
	return spec, nil
}

func (o Operation) MarshalJSON() ([]byte, error) {
	type operationJSON struct {
		ID   OperationID   `json:"id"`
		Kind OperationKind `json:"kind"`
		Spec OperationSpec `json:"spec"`
	}
	oj := operationJSON{
		ID:   o.ID,
		Kind: o.Spec.Kind(),
		Spec: o.Spec,
	}
	return json.Marshal(oj)
}

type CreateOperationSpec func() OperationSpec

// OperationSpec specifies an operation as part of a query.
type OperationSpec interface {
	// Kind returns the kind of the operation.
	Kind() OperationKind
}

// OperationID is a unique ID within a query for the operation.
type OperationID string

// OperationKind denotes the kind of operations.
type OperationKind string

var kindToOp = make(map[OperationKind]CreateOperationSpec)

// RegisterOpSpec registers an operation spec with a given kind.
// If the kind has already been registered the call panics.
func RegisterOpSpec(k OperationKind, c CreateOperationSpec) {
	if kindToOp[k] != nil {
		panic(fmt.Errorf("duplicate registration for operation kind %v", k))
	}
	kindToOp[k] = c
}

func NumberOfOperations() int {
	return len(kindToOp)
}
