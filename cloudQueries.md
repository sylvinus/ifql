# Dashboard: Cluster Stats
**Purpose:** For a single cluster, show important performance stats so that administrators can quickly identify performance issues.

## Queries in the dashboard Cluster Stats

```
// potential builtins
projectField = (f, table=<-) => filter(table:table, fn: (r) => r._field == f)
projectMeasurement = (m, table=<-) => filter(table:table, fn: (r) => r._measurement == f)
select = (measurement, field, table=<-) => projectMeasurement(m, table) |> projectField(f)
withTag = (key, value, table=<-) => filter(table:table, fn: (r) => r[key] == value)



// I don't need these for this set of queries, but I think they would be nice to have
// also they would need to write a custom func in Go
AnyOf = (table=<-, functions) =>  // filter that returns true if any of the list of input functions is true
OneOf = (table=<-, functions) => // filter that returns true if exactly one of the input functions is true
AllOf = (table=<-, functions) => // conjunction of filter functions.

// helper functions //
hostFilter = (table=<-) => filter(table:table, fn: (r) => (r["host"] =~ /.*data.*/  OR r["host"] =~ /tot-.*-(3|4)/))
// could be a built-in but we don't want to get too carried away with compound functions
fromRange = (forDB, forRange) => from(db:forDB) |> range(forRange)


CID = :Cluster_Id:

AggregateCPUCluster = (CID, agFn) =>
    fromRange(db:"telegraf", range:start-2m)
      |> select(measurement: "system", field: "n_cpus")
      |> withTag(key: "cluster_id", value: CID)
      |> hostFilter()
      |> group(by:["host"])
      |> last()
      |> agFn()



/* InfluxQL TotalClusterCPU Query:
SELECT sum("last") from
  (SELECT last("n_cpus")
   FROM "telegraf"."default"."system"
   WHERE time > now() - 2m and cluster_id = :Cluster_Id:
     AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/)
   GROUP BY host)
*/
TotalClusterCPU = AggregateCPUCluster(:Cluster_Id:, sum)

/* InfluxQL NumberOfNodes Query:
SELECT count("last") from
  (SELECT last("n_cpus")
    FROM "telegraf"."default"."system"
    WHERE time > now() - 2m and cluster_id = :Cluster_Id:
      AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/)
    GROUP BY host)
*/
NumberOfNodes = AggregateCPUCluster(:Cluster_Id:, count)


/* InfluxQL Memory Per Data Node Query:
SELECT last("max") from
  (SELECT max("total")/1073741824 FROM
    "telegraf"."default"."mem"
    WHERE "cluster_id" = :Cluster_Id:
     AND time > :dashboardTime:
     AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/)
     GROUP BY :interval:, host)
*/
MemPerDataNode = (CID, DASHTIME, INTERVAL) =>
    fromRange(db:"telegraf", range:DASHTIME)
      |> select(measurement: "mem", field: "total")
      |> withTag(key: "cluster_id", value: CID)
      |> hostfilter()
      |> window(every: INTERVAL)
      |> group(by: ["host"]
      |> max() / 1073741824

/* InfluxQL Disk Usage Query
SELECT last("used")/1073741824 AS "used" FROM
  "telegraf"."default"."disk"
  WHERE time > :dashboardTime:
    AND cluster_id = :Cluster_Id:
    AND (host =~ /.data./ OR host =~ /tot-.*-(3|4)/)
  FILL(0)

    // fill seems to be used to return a default value. Not sure how IFQL behaves
*/
DiskUsage = (CID, DASHTIME) =>
    fromRange(db:"telegraf", range:DASHTIME)
      |> select(measurement: "disk", field: "used")
      |> withTag(key: "cluster_id", value: CID)
      |> hostfilter()
      |> last() / 1073741824


/* InfluxQL Disk Allocated Query
SELECT last("max") from
  (SELECT max("total")/1073741824 FROM
    "telegraf"."default"."disk"
    WHERE time > :dashboardTime:
      AND cluster_id = :Cluster_Id:
      AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/)
    GROUP BY :interval:)
*/
DiskAllocated = (CID, DASHTIME, INTERVAL) =>
  DiskUsage(CID, DASHTIME)
    |> window(INTERVAL)
    |> last()


/* InfluxQL Percent Availability Query
SELECT (sum("service_up") / count("service_up"))*100 AS "up_time"
  FROM "watcher"."autogen"."ping"
  WHERE cluster_id = :Cluster_Id:
    AND time > :dashboardTime:
  FILL(0)
*/

PercentAvailability = (CID, DASHTIME) =>
  serviceUp =
    fromRange(db:"watcher", range: DASHTIME)
      |> select(measurement: "ping", field: "service_up")
      |> withTag(key: "cluster_id", value: CID)

  up_time = sum(serviceUp) / count(serviceUp) * 100
  return up_time


/* InfluxQL CPU Utilization
SELECT mean("usage_user") AS "Usage" FROM
  "telegraf"."default"."cpu"
  WHERE time > :dashboardTime:
    AND cluster_id = :Cluster_Id:
  GROUP BY :interval:,host
*/

Usage = (CID, DASHTIME, INTERVAL) =>
  fromRange(db: "telegraf", range: DASHTIME)
    |> select(measurement: "cpu", field: "usage_user")
    |> withTag(key: "cluster_id", value: CID)
    |> window(INTERVAL)
    |> group(by: ["host"])
    |> mean()

```