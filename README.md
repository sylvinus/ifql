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
'q=select(db:"telegraf")
.filter(exp:{"_measurement" == "cpu" AND "_field" == "usage_user"})
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

Example: `select(db: "telegraf")`

##### options
* `db` string
    `select(db: "telegraf")`

* `hosts` array of strings
    `select(db: "telegraf", hosts:["host1", "host2"])`

#### count
Counts the number of results

Example: `select(db:"telegraf").count()`

#### first
Returns the first result of the query

Example: `select(db: "telegraf").first()`

#### group
Groups results by a user-specified set of tags

##### options

*  `by` array of strings
Group by these specific tag names
Cannot be used with `ignore` option

Example: `select(db: "telegraf").range(start: -30m).group(by: ["tag_a", "tag_b"])`

*  `keep` array of strings
Keep specific tag keys that were not in `by` in the results 

Example: `select(db: "telegraf").range(start: -30m).group(by: ["tag_a", "tag_b"], keep:["tag_c"])`
*  `ignore` array of strings
Group by all but these tag keys
Cannot be used with `by` option

Example: `select(db: "telegraf").range(start: -30m).group(ignore: ["tag_a"], keep:["tag_b", "tag_c"])`

#### join
Join two time series together on time and the list of `on` keys.

Example:
```
var cpu = select(db: "telegraf").filter(exp:{"_measurement" == "cpu" and "_field" == "usage_user"}).range(start: -30m)
select(db: "telegraf").filter(exp:{"_measurement" == "mem" and "_field" == "used_percent"}).range(start: -30m)
.join(on:["host"], exp:{$ + cpu})
````

The special identifier `$` represents the current value of the query.  It can
only be used in expressions.

##### options

* `on` array of strings
List of tag keys that when equal produces a result set.

* `exp` 

Defines the expression that merges the joined results sets together

#### last
Returns the last result of the query

Example: `select(db: "telegraf").last()`

#### limit
Restricts the number of rows returned in the results.

Example: `select(db: "telegraf").limit(n: 10)`

#### max

Returns the max value within the results

Example:
```
select(db:"foo")
    .filter(exp:{"_measurement"=="cpu" AND 
                "_field"=="usage_system" AND 
                "service"=="app-server"})
    .range(start:-12h)
    .window(every:10m)
    .max()
```

#### mean
Returns the mean of the values within the results

Example:
```
select(db:"foo")
    .filter(exp:{"_measurement"=="mem" AND 
                "_field"=="used_percent"})
    .range(start:-12h)
    .window(every:10m)
    .mean()
```

#### min
Returns the min value within the results

Example:
```
select(db:"foo")
    .filter(exp:{"_measurement"=="cpu" AND 
                "_field"=="usage_system"})
    .range(start:-12h)
    .window(every:10m, period: 5m)
    .min()
```


#### range
Filters the results by time boundaries

Example:
```
select(db:"foo")
    .filter(exp:{"_measurement"=="cpu" AND 
                "_field"=="usage_system"})
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
select(db:"foo")
    .filter(exp:{"_measurement"=="cpu" AND 
                "_field"=="usage_system"})
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
Example: `select(db: "telegraf").set(key: "mykey", value: "myvalue")`
##### options
* `key` string
* `value` string

#### skew
Skew of the results

Example: `select(db: "telegraf").range(start: -30m, stop: -15m).skew()`

#### sort
Sorts the results by the specified columns
Default sort is ascending

Example: 
```
select(db:"telegraf")
    .filter(exp:{"_measurement"=="system" AND 
                "_field"=="uptime"})
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
select(db:"telegraf")
    .filter(exp:{"_measurement"=="system" AND 
                "_field"=="uptime"})
    .range(start:-12h)
    .sort(desc: true)
```

* `desc` bool 
Sort results descending

#### spread
Difference between min and max values

Example: `select(db: "telegraf").range(start: -30m).spread()`

#### stddev
Standard Deviation of the results

Example: `select(db: "telegraf").range(start: -30m, stop: -15m).stddev()`

#### sum
Sum of the results

Example: `select(db: "telegraf").range(start: -30m, stop: -15m).sum()`

#### filter
Filters the results using an expression

Example:
```
select(db:"foo")
    .filter(exp:{"_measurement"=="cpu" AND 
                "_field"=="usage_system" AND 
                "service"=="app-server"})
    .range(start:-12h)
    .max()
```

#### window
Partitions the results by a given time range 

##### options
* `every` duration
Duration of time between windows

Defaults to `period`'s value
```
select(db:"foo")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

* `period` duration
Duration of the windowed parition
```
select(db:"foo")
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
select(db:"foo")
    .range(start:-12h)
    .window(every:10m)
    .max()
```

