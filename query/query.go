package query

import (
	"errors"
	"fmt"
)

// QuerySpec specifies a query.
type QuerySpec struct {
	Operations []*Operation
	Edges      []Edge

	sorted []*Operation
}

type Edge struct {
	Parent OperationID `json:"parent"`
	Child  OperationID `json:"child"`
}

// ExpressionSpec specifies an expression.
// Expressions may return any value that the storage supports.
//TODO flesh this out.
type ExpressionSpec interface{}

// Walk calls f on each operation exactly once.
// The function f will be called on an operation only after
// all of its parents have already been passed to f.
func (q *QuerySpec) Walk(f func(o *Operation) error) error {
	if len(q.sorted) == 0 {
		if err := q.sort(); err != nil {
			return err
		}
	}
	for _, o := range q.sorted {
		err := f(o)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate ensures the query is a valid DAG.
func (q *QuerySpec) Validate() error {
	return q.sort()
}

// sort populates the sorted field and also validates that the query is a valid DAG.
func (q *QuerySpec) sort() error {
	children, roots, err := q.determineChildrenAndRoots()
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return errors.New("query has no root nodes")
	}

	tMarks := make(map[OperationID]bool)
	pMarks := make(map[OperationID]bool)

	for _, r := range roots {
		if err := q.visit(tMarks, pMarks, children, r); err != nil {
			return err
		}
	}
	//reverse q.sorted
	for i, j := 0, len(q.sorted)-1; i < j; i, j = i+1, j-1 {
		q.sorted[i], q.sorted[j] = q.sorted[j], q.sorted[i]
	}
	return nil
}

func (q *QuerySpec) computeLookup() (map[OperationID]*Operation, error) {
	lookup := make(map[OperationID]*Operation, len(q.Operations))
	for _, o := range q.Operations {
		if _, ok := lookup[o.OperationID]; ok {
			return nil, fmt.Errorf("found duplicate operation ID %q", o.OperationID)
		}
		lookup[o.OperationID] = o
	}
	return lookup, nil
}

func (q *QuerySpec) determineChildrenAndRoots() (children map[OperationID][]*Operation, roots []*Operation, _ error) {
	lookup, err := q.computeLookup()
	if err != nil {
		return nil, nil, err
	}
	children = make(map[OperationID][]*Operation, len(q.Operations))
	parentCount := make(map[OperationID]int)
	for _, e := range q.Edges {
		// Build children map
		c, ok := lookup[e.Child]
		if !ok {
			return nil, nil, fmt.Errorf("edge references unknown child operation %q", e.Child)
		}
		children[e.Parent] = append(children[e.Parent], c)

		// Count parents of each operation
		parentCount[e.Child]++
	}
	for _, o := range q.Operations {
		count := parentCount[o.OperationID]
		if count == 0 {
			roots = append(roots, o)
		}
	}
	return
}

// Depth first search topological sorting of a DAG.
// https://en.wikipedia.org/wiki/Topological_sorting#Algorithms
func (q *QuerySpec) visit(tMarks, pMarks map[OperationID]bool, children map[OperationID][]*Operation, o *Operation) error {
	id := o.OperationID
	if tMarks[id] {
		return errors.New("found cycle in query")
	}

	if !pMarks[id] {
		tMarks[id] = true
		for _, c := range children[id] {
			if err := q.visit(tMarks, pMarks, children, c); err != nil {
				return err
			}
		}
		pMarks[id] = true
		tMarks[id] = false
		q.sorted = append(q.sorted, o)
	}
	return nil
}
