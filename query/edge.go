package query

type Edge struct {
	Parent OperationID `json:"parent"`
	Child  OperationID `json:"child"`
}
