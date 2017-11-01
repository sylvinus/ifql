package querytest

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/query"
)

// MockOperationSpec is a mock of the OperationSpec interface
type MockOperationSpec struct {
	Op query.OperationKind
}

// Kind returns the mocked Op
func (m *MockOperationSpec) Kind() query.OperationKind {
	return m.Op
}

func OperationMarshalingTestHelper(t *testing.T, data []byte, expOp *query.Operation) {
	t.Helper()

	// Ensure we can properly unmarshal a spec
	gotOp := new(query.Operation)
	if err := json.Unmarshal(data, gotOp); err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(gotOp, expOp) {
		t.Errorf("unexpected operation -want/+got %s", cmp.Diff(expOp, gotOp))
	}

	// Marshal the spec and ensure we can unmarshal it again.
	data, err := json.Marshal(expOp)
	if err != nil {
		t.Fatal(err)
	}
	gotOp = new(query.Operation)
	if err := json.Unmarshal(data, gotOp); err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(gotOp, expOp) {
		t.Errorf("unexpected operation after marshalling -want/+got %s", cmp.Diff(expOp, gotOp))
	}
}
