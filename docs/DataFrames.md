# DataFrames

DataFrames are the basic data model for the query engine.
A DataFrame represents a time bounded set of data.
DataFrames are a matrix where rows labels are series keys and columns labels are timestamps.
DataFrames are organized into blocks.
A block represents data from a single group and window.

DataFrames created from select statements by default have each row of as its own group and the entire frame exists in a single window.
(This may not be good defaults need to explore more)



While bounded the data may still be too large to maintain in RAM.
Bounds on data indicate how aggregate transformations, etc. should behave.
Batching the data so that it can be processed as a stream is an orthogonal issue and as such is not part of the DataFrame interface.

## Sparse vs. Dense

A DataFrame is an interface and there are different implementations of that interface.
Namely there will be three different implementations.

* Dense
* Sparse Row Optimized
* Sparse Column Optimized

A dense matrix implementation assumes that there is little to no missing data.
A dense matrix is typically "row-major" meaning its optimized for row based operations, at this point it doesn't seem helpful to have a column major dense implementation.
A sparse matrix implementation assumes that there is a significant amount of missing data.
Sparse implementations can be optimized for either row or column operations.

Since different processes access data in different patterns the planning step will be responsible for deciding which implementation is best at which steps in a query.
The planner will add transformations procedures for conversions between the different implementations.




## ????

DataFrame doesn't need to exist? Simply process blocks?
How does recovery work? What kinds of recovery events will there be?

Block
BlockKey?

Watermark
Trigger
AccumulatorMode

Transformation
Dataset

