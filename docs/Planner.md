# Planner Design

This document lays out the design of the planner.

## Interface

The Planner inter is defined as:

```go
type Planner interface {
    Plan(Query, []Storage) Plan
}
```

The planner is responsible for taking a query DAG and a set of available storage interfaces and produce a plan DAG.

