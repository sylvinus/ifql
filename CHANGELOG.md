## v0.0.3 [2017-12-08]

### Features

- [#166](https://github.com/influxdata/ifql/issues/166) Initial Resource Management API is in place. Now queries can be submitted with varying priorities and limits on concurrency and memory usage.
- [#164](https://github.com/influxdata/ifql/issues/164) Opentracing support for queries.
- [#139](https://github.com/influxdata/ifql/issues/139) Join is now a global function.
- [#130](https://github.com/influxdata/ifql/issues/130) Add error handling of duplicate arguments to functions.
- [#100](https://github.com/influxdata/ifql/issues/100) Add error handling of unknown arguments to functions.

### Bugfixes

- [#153](https://github.com/influxdata/ifql/issues/153) Fix issues with line protocol output if the _measurement and _field tags were missing.

## v0.0.2 [2017-11-21]

Release after some initial community feedback.

## v0.0.1 [2017-11-13]
Initial release of ifqld
