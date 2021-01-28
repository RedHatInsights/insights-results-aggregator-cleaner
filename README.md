# insights-results-aggregator-cleaner

[![GoDoc](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner?status.svg)](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-results-aggregator-cleaner/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-results-aggregator-cleaner)](https://goreportcard.com/report/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![Build Status](https://travis-ci.com/RedHatInsights/insights-results-aggregator-cleaner.svg?branch=master)](https://travis-ci.com/RedHatInsights/insights-results-aggregator-cleaner)

## Description

Simple service that can be used to identify clusters, for which we are keeping
very old data (>30 days) in the database. This means that the cluster is no
longer available or that the customer has disabled the Insights Operator,
either way it means that these data are no longer relevant to us and should be
pruned.

Such clusters can be detected very easily by checking the timestamps stored
(along other information) in the `report` table in Insights Results Aggregator
database.

Currently this service just displays such clusters (cluster IDs) and do nothing
else - i.e. the results are not deleted.

### Building

```
make build
```

### Configuration

Default name of configuration file is `config.toml`.
It can be changed via environment variable `INSIGHTS_RESULTS_CLEANER_CONFIG_FILE`.

An example of configuration file that can be used in devel environment:

```
[storage]
db_driver = "postgres"
pg_username = "postgres"
pg_password = "postgres"
pg_host = "localhost"
pg_port = 5432
pg_db_name = "aggregator"
pg_params = "sslmode=disable"

[logging]
debug = true
log_level = ""

[cleaner]
max_age = "90 days"
```

Environment variables that can be used to override configuration file settings:

```
INSIGHTS_RESULTS_CLEANER__STORAGE__DB_DRIVER
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_USERNAME
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PASSWORD
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_HOST
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PORT
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_DB_NAME
INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PARAMS
INSIGHTS_RESULTS_CLEANER__LOGGING__DEBUG
INSIGHTS_RESULTS_CLEANER__LOGGING__LOG_DEVEL
INSIGHTS_RESULTS_CLEANER__CLEANER__MAX_AGE
```
### Usage

Just the service needs to be started:

```
./insights-results-aggregator-cleaner
```

#### Output example

* Logging is set to `true`

```
{"level":"info","filename":"config","time":"2021-01-28T10:08:47+01:00","message":"Parsing configuration file"}
10:08AM DBG Started
10:08AM INF DB connection configuration driverName=postgres
10:08AM INF Old report age=394 cluster=5d5892d4-1f74-4ccf-91af-548dfc9767aa lastChecked=2020-04-09T06:16:02Z reported=2020-01-01T00:00:00Z
10:08AM INF Old report age=363 cluster=b0c2d108-0603-41c3-9a8f-0a37eba5df49 lastChecked=2020-01-23T16:15:59Z reported=2020-02-01T00:00:00Z
10:08AM INF Old report age=334 cluster=22222222-bbbb-cccc-dddd-ffffffffffff lastChecked=2020-01-23T16:15:59Z reported=2020-03-01T00:00:00Z
10:08AM INF Old report age=303 cluster=33333333-bbbb-cccc-dddd-ffffffffffff lastChecked=2020-01-23T16:15:59Z reported=2020-04-01T00:00:00Z
10:08AM INF Old report age=273 cluster=5d5892d3-1f74-4ccf-91af-548dfc9767ac lastChecked=2020-04-02T09:00:05Z reported=2020-05-01T00:00:00Z
10:08AM INF Old report age=242 cluster=5d5892d3-1f74-4ccf-91af-548dfc9767aa lastChecked=2020-04-09T06:16:02Z reported=2020-06-01T00:00:00Z
10:08AM INF Old report age=212 cluster=6d5892d3-1f74-4ccf-91af-548dfc9767aa lastChecked=2020-04-02T09:00:05Z reported=2020-07-01T00:00:00Z
10:08AM INF Old report age=181 cluster=5e5892d3-1f74-4ccf-91af-548dfc9767aa lastChecked=2020-04-02T09:00:05Z reported=2020-08-01T00:00:00Z
10:08AM INF Old report age=150 cluster=c0c2d108-0603-41c3-9a8f-0a37eba5df49 lastChecked=2020-01-23T16:15:59Z reported=2020-09-01T00:00:00Z
10:08AM INF Old report age=120 cluster=abaaaaaa-bbbb-cccc-dddd-ffffffffffff lastChecked=2020-01-23T16:15:59Z reported=2020-10-01T00:00:00Z
10:08AM DBG Finished
```

* Logging is set to `false`

```
{"level":"info","filename":"config","time":"2021-01-28T10:09:49+01:00","message":"Parsing configuration file"}
{"level":"debug","time":"2021-01-28T10:09:49+01:00","message":"Started"}
{"level":"info","driverName":"postgres","time":"2021-01-28T10:09:49+01:00","message":"DB connection configuration"}
{"level":"info","cluster":"5d5892d4-1f74-4ccf-91af-548dfc9767aa","reported":"2020-01-01T00:00:00Z","lastChecked":"2020-04-09T06:16:02Z","age":394,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"b0c2d108-0603-41c3-9a8f-0a37eba5df49","reported":"2020-02-01T00:00:00Z","lastChecked":"2020-01-23T16:15:59Z","age":363,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"22222222-bbbb-cccc-dddd-ffffffffffff","reported":"2020-03-01T00:00:00Z","lastChecked":"2020-01-23T16:15:59Z","age":334,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"33333333-bbbb-cccc-dddd-ffffffffffff","reported":"2020-04-01T00:00:00Z","lastChecked":"2020-01-23T16:15:59Z","age":303,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"5d5892d3-1f74-4ccf-91af-548dfc9767ac","reported":"2020-05-01T00:00:00Z","lastChecked":"2020-04-02T09:00:05Z","age":273,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"5d5892d3-1f74-4ccf-91af-548dfc9767aa","reported":"2020-06-01T00:00:00Z","lastChecked":"2020-04-09T06:16:02Z","age":242,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"6d5892d3-1f74-4ccf-91af-548dfc9767aa","reported":"2020-07-01T00:00:00Z","lastChecked":"2020-04-02T09:00:05Z","age":212,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"5e5892d3-1f74-4ccf-91af-548dfc9767aa","reported":"2020-08-01T00:00:00Z","lastChecked":"2020-04-02T09:00:05Z","age":181,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"c0c2d108-0603-41c3-9a8f-0a37eba5df49","reported":"2020-09-01T00:00:00Z","lastChecked":"2020-01-23T16:15:59Z","age":150,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"info","cluster":"abaaaaaa-bbbb-cccc-dddd-ffffffffffff","reported":"2020-10-01T00:00:00Z","lastChecked":"2020-01-23T16:15:59Z","age":120,"time":"2021-01-28T10:09:49+01:00","message":"Old report"}
{"level":"debug","time":"2021-01-28T10:09:49+01:00","message":"Finished"}
```
