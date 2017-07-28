# Executor Design

This document lays out the design of the executor.

## Interface

The Executor interface is defined as:

```go
type Executor interface {
    Execute(context.Context, Plan) ([]DataFrame, ErrorState)
}
```

The executor is responsible for taking a plan and executing it to produce the result which is a list of DataFrames or an error state.


