# Executor Design

This document lays out the design of the executor.

## Interface

The Executor interface is defined as:

```go
type Executor interface {
    Execute(context.Context, Plan) ([]DataFrame, ErrorState)
}
```

The executor is responsible for taking a specific plan from the Planner and executing it to produce the result which is a list of DataFrames or an error state.

##  Concepts

The executor interacts with many different systems and has its own internal systems
Below is a list of concepts within the executor.

| Concept                      | Description                                                                                                                                                               |
| -------                      | -----------                                                                                                                                                               |
| Bounded Data                 | Datasets that are finite, in other words `batch` data.                                                                                                                    |
| Unbounded Data               | Datasets that have no know end, or are infinite, in other words `stream` data.                                                                                            |
| Event time                   | The time the event actually occurred.                                                                                                                                     |
| Processing time              | The time the event is processed. This time may be completely out of order with respect to its event time and the event time of other events with similar processing time. |
| Watermarks                   | Watermarks communicate the lag between event time and processing time. Watermarks define a bound on the event time of data that has been observed.                        |
| Horizon | Duration after the watermark of which data is simply dropped since it arrived to late. |
| Triggers                     | Triggers communicate when data should be materialized.                                                                                                                    |
| Accumulation                 | Accumulation defines how different results from events of the same window can be combined into a single result.                                                           |
| Data Frame                   | An abstraction over bounded datasets. The dataset is resilient because its lineage is known and it can be recreated in the event of loss or corruption.                   |
| Discretized Stream (DStream) | An unbounded data set that has been discretized into a sequence of bounded Data Frames.                                                                                   |
| Operation                    | A operation performs a transformation on a set of parent Data Frames and produces a set of child Data Frames.                                                             |
| Execution State              | Execution state tracks the state of an execution.                                                                                                                         |

## Relations


The relations between the various concepts is defined below using the Go language.

```
type Event interface {
    Time() time.Time
    ProcessingTime() time.Time
}

type DataFrame interface {
    Watermark() Watermark
    Triggers() <-chan Trigger
    Accumulation() Accumulation
    Lineage() Lineage
    // Cursor provides access to data within the data frame.
    Cursor() Cursor
}

type DStream interface {
    Stream() <-chan DataFrame
}

type Watermark interface {
    Bound() time.Time
}

type Trigger interface {}

type Lineage interface {
    Operations() Operations
}

type OperationClass
const (
    NarrowOperation OperationClass = iota
    WideOperation
)

type Operations interface {
    Class() OperationClass
    Parents() []DataFrames
}
```


## Triggers and Materialization

Data frames are materialized lazily, meaning that operations of a plan do no work until their children data frames request that they be materialized.
Data frames may be materialized partially or completely.

Triggers define when a data frame should be materialized.
For example new data may arrive out of order after a data frame has already been materialized.
In this case we much choose whether to re-materialize the data frame or wait a period of time for more data to arrive before materializing.
Triggers take as input various events like a change in the watermark of a Data Frame


## Execution State

While both queries and plans are specifications the execution state encapsulates the implementation and state of executing a query.

