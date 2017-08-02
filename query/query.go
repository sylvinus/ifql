package query

// QuerySpec specifies a query.
type QuerySpec struct {
	Operations []Operation
	Edges      []Edge
}

func (q *QuerySpec) Walk(f func(p Operation) error) error {
	panic("not implemented")
}

// ValidateDAG ensures the query is a valid DAG.
func (q *QuerySpec) ValidateDAG() error {
	panic("not implemented")
}
