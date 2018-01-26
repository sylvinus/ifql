# IFQL (Influx Query Language)

`ifqld` is an HTTP server for running **IFQL** queries to one or more InfluxDB
servers.

`ifqld` runs on port `8093` by default

### Specification
Here is the rough design specification for details until we get documentation up: http://bit.ly/ifql-spec

### INSTALLATION
1. Upgrade to InfluxDB >= 1.4.1
https://portal.influxdata.com/downloads


2. Update the InfluxDB configuration file to enable **IFQL** processing; restart
the InfluxDB server. InfluxDB will open port `8082` to accept **IFQL** queries.

> **This port has no authentication.**

```
[ifql]
  enabled = true
  log-enabled = true
  bind-address = ":8082"
```

3. Download `ifqld` and install from https://github.com/influxdata/ifql/releases

4. Start `ifqld` with the InfluxDB host and port of `8082`. To run in federated
mode (see below), add the `--host` option for each InfluxDB host.

```sh
ifqld --verbose --host localhost:8082
```

5. To run a query POST an **IFQL** query string to `/query` as the `q` parameter:
```sh
curl -XPOST --data-urlencode \
'q=from(db:"telegraf")
.filter(fn: (r) => r["_measurement"] == "cpu" AND r["_field"] == "usage_user")
.range(start:-170h).sum()' \
localhost:8093/query
```

#### docker compose

To spin up a testing environment you can run:

```
docker-compose up
```

Inside the `root` directory. It will spin up an `influxdb` and `ifqld` daemon
ready to be used. `influxd` is exposed on port `8086` and port `8082`.


### Prometheus metrics
Metrics are exposed on `/metrics`.
`ifqld` records the number of queries and the number of different functions within **IFQL** queries

### Federated Mode
By passing the `--host` option multiple times `ifqld` will query multiple
InfluxDB servers.

For example:

```sh
ifqld --host influxdb1:8082 --host influxdb2:8082
```

The results from multiple InfluxDB are merged together as if there was
one server.


### Supported Functions

Example: `from(db: "telegraf")`

##### options
* `db` string
    `from(db: "telegraf")`

* `hosts` array of strings
    `from(db: "telegraf", hosts:["host1", "host2"])`

#### count
Counts the number of results

Example: `from(db:"telegraf").count()`

#### first
Returns the first result of the query

Example: `from(db: "telegraf").first()`

#### group
Groups results by a user-specified set of tags

##### options

*  `by` array of strings
Group by these specific tag names
Cannot be used with `except` option

Example: `from(db: "telegraf").range(start: -30m).group(by: ["tag_a", "tag_b"])`

*  `keep` array of strings
Keep specific tag keys that were not in `by` in the results

Example: `from(db: "telegraf").range(start: -30m).group(by: ["tag_a", "tag_b"], keep:["tag_c"])`
*  `except` array of strings
Group by all but these tag keys
Cannot be used with `by` option

Example: `from(db: "telegraf").range(start: -30m).group(except: ["tag_a"], keep:["tag_b", "tag_c"])`

#### join
Join two time series together on time and the list of `on` keys.

Example:

```
cpu = from(db: "telegraf").filter(fn: (r) => r["_measurement"] == "cpu" and r["_field"] == "usage_user").range(start: -30m)
mem = from(db: "telegraf").filter(fn: (r) => r["_measurement"] == "mem" and r["_field"] == "used_percent"}).range(start: -30m)
join(tables:{cpu:cpu, mem:mem}, on:["host"], fn: (tables) => tables.cpu["_value"] + tables.mem["_value"])
````

##### options

* `tables` map of tables
Map of tables to join. Currently only two tables are allowed.

* `on` array of strings
List of tag keys that when equal produces a result set.

* `fn`

Defines the function that merges the values of the tables.
The function must defined to accept a single parameter.
The parameter is a map, which uses the same keys found in the `tables` map.
The function is called for each joined set of records from the tables.

#### last
Returns the last result of the query

Example: `from(db: "telegraf").last()`

#### limit
Restricts the number of rows returned in the results.

Example: `from(db: "telegraf").limit(n: 10)`

#### max

Returns the max value within the results

Example:
```
from(db:"foo")
    .filter(fn: (r) => r["_measurement"]=="cpu" AND
                r["_field"] == "usage_system" AND
                r["service"] == "app-server")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

#### mean
Returns the mean of the values within the results

Example:
```
from(db:"foo")
    .filter(fn: (r) => r["_measurement"] == "mem" AND
                r["_field"] == "used_percent")
    .range(start:-12h)
    .window(every:10m)
    .mean()
```

#### min
Returns the min value within the results

Example:
```
from(db:"foo")
    .filter(fn: (r) => r[ "_measurement"] == "cpu" AND
                r["_field" ]== "usage_system")
    .range(start:-12h)
    .window(every:10m, period: 5m)
    .min()
```


#### range
Filters the results by time boundaries

Example:
```
from(db:"foo")
    .filter(fn: (r) => r["_measurement"] == "cpu" AND
                r["_field"] == "usage_system")
    .range(start:-12h, stop: -15m)
```

##### options
* start duration
Specifies the oldest time to be included in the results

* stop duration or timestamp
Specifies exclusive upper time bound
Defaults to "now"

#### sample

Example to sample every fifth point starting from the second element:
```
from(db:"foo")
    .filter(fn: (r) => r["_measurement"] == "cpu" AND
                r["_field"] == "usage_system")
    .range(start:-1d)
    .sample(n: 5, pos: 1)
```

##### options
* `n`
Sample every Nth element
* `pos`
Position offset from start of results to begin sampling
`pos` must be less than `n`
If `pos` less than 0, a random offset is used.
Default is -1 (random offset)

#### set
Add tag of key and value to set
Example: `from(db: "telegraf").set(key: "mykey", value: "myvalue")`
##### options
* `key` string
* `value` string

#### skew
Skew of the results

Example: `from(db: "telegraf").range(start: -30m, stop: -15m).skew()`

#### sort
Sorts the results by the specified columns
Default sort is ascending

Example:
```
from(db:"telegraf")
    .filter(fn: (r) => r["_measurement"] == "system" AND
                r["_field"] == "uptime")
    .range(start:-12h)
    .sort(cols:["region", "host", "value"])
```

##### options
* `cols` array of strings
List of columns used to sort; precedence from left to right.
Default is `["value"]`

For example, this sorts by uptime descending to find the longest
running instances.

```
from(db:"telegraf")
    .filter(fn: (r) => r["_measurement"] == "system" AND
                r["_field"] == "uptime")
    .range(start:-12h)
    .sort(desc: true)
```

* `desc` bool
Sort results descending

#### spread
Difference between min and max values

Example: `from(db: "telegraf").range(start: -30m).spread()`

#### stddev
Standard Deviation of the results

Example: `from(db: "telegraf").range(start: -30m, stop: -15m).stddev()`

#### sum
Sum of the results

Example: `from(db: "telegraf").range(start: -30m, stop: -15m).sum()`

#### filter
Filters the results using an expression

Example:
```
from(db:"foo")
    .filter(fn: (r) => r["_measurement"]=="cpu" AND
                r["_field"] == "usage_system" AND
                r["service"] == "app-server")
    .range(start:-12h)
    .max()
```

##### options

* `fn` function(record) bool

Function to when filtering the records.
The function must accept a single parameter which will be the records and return a boolean value.
Records which evaluate to true, will be included in the results.

#### window
Partitions the results by a given time range

##### options
* `every` duration
Duration of time between windows

Defaults to `period`'s value
```
from(db:"foo")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

* `period` duration
Duration of the windowed parition
```
from(db:"foo")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

Default to `every`'s value
* `start` time
The time of the initial window parition.

* `round` duration
Rounds a window's bounds to the nearest duration

Example:
```
from(db:"foo")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

### Custom Functions

IFQL also allows the user to define their own functions.
The function syntax is:

```
(parameter list) => <function body>
```

The list of parameters is simply a list of identifiers with optional default values.
The function body is either a single expression which is returned or a block of statements.
Functions may be assigned to identifiers to given them a name.

Examples:

```
// Define a simple addition function
add = (a,b) => a + b

// Define a helper function to get data from a telegraf measurement.
// By default the database is expected to be named "telegraf".
telegrafM = (measurement, db="telegraf") =>
    from(db:db)
        .filter(fn: (r) => r._measurement == measurement)

// Define a helper function for a common join operation
// Use block syntax since we have more than a single expression
abJoin = (measurementA, measurementB, on) => {
    a = telegrafM(measurement:measurementA)
    b = telegrafM(measurement:measurementB)
    join(
        tables:{a:a, b:b},
        on:on,
        // Return a map from the join fn,
        // this creates a table with a column for each key in the map.
        // Note the () around the map to indicate a single map expression instead of function block.
        fn: (t) => ({
            a: t.a._value,
            b: t.b._value,
        }),
    )
}
```
