# IFQL (Influx Query Language)

IFQLD is an HTTP server for processing IFQL queries to 1 or more InfluxDB servers.

This repo contains the spec and implementation of IFQL.


ifqld runs on 8093 by default

## example with docker-compose

To spin up a testing environment you can run:

```
docker-compose up
```

Inside the `root` directory. It will spin up an `influxdb` and `ifqld` daemon
ready to be used. `influxd` is exposed on port `8086`.

### INSTALLATION
1. Upgrade to InfluxDB 1.4
https://portal.influxdata.com/downloads


2. Update the InfluxDB configuration file to enable ifql queries
```
[ifql]
  enabled = true
  log-enabled = true
  bind-address = ":8082"
```

InfluxDB will open port 8082 to accept IFQL queries.  This port has no authentication.


3. Install IFQLD: https://github.com/influxdata/ifql/releases


4. ifqld --host localhost:8082


5. POST

```
select(db:"telegraf").last()
```
