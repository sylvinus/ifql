# Dashboard: Cluster Stats
**Purpose:** For a single cluster, show important performance stats so that administrators can quickly identify performance issues.  

## Queries in the dashboard


|Query| Cell Type  | Notes | 
|---|---|---|
|Total cluster CPU| Single Stat | regex operator =~ not implemented yet | 
  
**InfluxQL**
```
SELECT sum("last") from 
  (SELECT last("n_cpus") 
   FROM "telegraf"."default"."system" 
   WHERE time > now() - 2m and cluster_id = :Cluster_Id: 
     AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/) 
   GROUP BY host)
```
       
**IFQL**
```
CID = :Cluster_Id:
from(db:"telegraf") 
  |> range(start:-2m) 
  |> filter( 
      fn: (r) => r["_measurement"] == "system 
        AND r["_field"] == "n_cpus"
        AND r["cluster_id"]==CID 
        AND (r["host"] =~ /.*data.*/  OR r["host"] =~ /tot-.*-(3|4)/))
  |> group(by:["host"]) 
  |> last()
  |> sum()
```

|Query| Cell Type  | Notes | 
|---|---|---|
|\# of Nodes| Single Stat | regex operator =~ not implemented yet | 
  
**InfluxQL**
```
SELECT count("last") from 
  (SELECT last("n_cpus") 
    FROM "telegraf"."default"."system" 
    WHERE time > now() - 2m and cluster_id = :Cluster_Id: 
      AND (host =~ /.*data.*/ OR host =~ /tot-.*-(3|4)/) 
    GROUP BY host)
```
       
**IFQL**
```
CID = :Cluster_Id:
from(db:"telegraf") 
  |> range(start:-2m) 
  |> filter( fn: (r) => r["cluster_id"]==CID AND (r["host"] =~ /.*data.*/  OR r["host"] =~ /tot-.*-(3|4)/))
  |> group(by:["host"]) 
  |> last()
  |> sum()
```
