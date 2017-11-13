# IFQL (Influx Functional Query Language)

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
