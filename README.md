# IFQL (Influx Query Language)

`ifqld` is an HTTP server for processing **IFQL** queries to one or more InfluxDB
servers.

`ifqld` runs on `8093` by default

### INSTALLATION
1. Upgrade to InfluxDB >= 1.4.1
https://portal.influxdata.com/downloads


2. Update the InfluxDB configuration file to enable **IFQL** processing. InfluxDB
will open port `8082` to accept **IFQL** queries.

> **This port has no authentication.**

```
[ifql]
  enabled = true
  log-enabled = true
  bind-address = ":8082"
```

3. Install `ifqld`: https://github.com/influxdata/ifql/releases

4. Start `ifqld` with the InfluxDB host and port of `8082`. To run in federated
mode, add the `--host` option for each InfluxDB host.

```sh
ifqld --verbose --host localhost:8082
```

5. To run a query POST an **IFQL** query string to `/query` as the `q` parameter:
```sh
curl -XPOST --data-urlencode \
'q=select(db:"telegraf")
.where(exp:{"_measurement" == "cpu" AND "_field" == "usage_user"})
.range(start:-170h).sum()' \
localhost:8093/query
```

#### Docker compose

To spin up a testing environment you can run:

```
docker-compose up
```

Inside the `root` directory. It will spin up an `influxdb` and `ifqld` daemon
ready to be used. `influxd` is exposed on port `8086` and port `8082`.


### Prometheus metrics
Metrics are exposed on `/metrics`.
`ifqld` records the number of queries and the number of different functions within **IFQL** queries