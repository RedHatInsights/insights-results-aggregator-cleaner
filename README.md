# insights-results-aggregator-cleaner

[![GoDoc](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner?status.svg)](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-results-aggregator-cleaner/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-results-aggregator-cleaner)](https://goreportcard.com/report/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![Build Status](https://travis-ci.com/RedHatInsights/insights-results-aggregator-cleaner.svg?branch=master)](https://travis-ci.com/RedHatInsights/insights-results-aggregator-cleaner)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/RedHatInsights/insights-results-aggregator-cleaner)
[![License](https://img.shields.io/badge/license-Apache-blue)](https://github.com/RedHatInsights/insights-results-aggregator-cleaner/blob/master/LICENSE)

<!-- vim-markdown-toc GFM -->

* [Description](#description)
    * [Usage](#usage)
    * [Default operation](#default-operation)
    * [Data cleanup](#data-cleanup)
    * [Test data generation](#test-data-generation)
    * [Exit status](#exit-status)
    * [Building](#building)
    * [Makefile targets](#makefile-targets)
    * [Configuration](#configuration)
    * [Usage](#usage-1)
        * [Output example](#output-example)
* [Database structure](#database-structure)
    * [Table `report`](#table-report)
    * [Table `cluster_rule_toggle`](#table-cluster_rule_toggle)
    * [Table `cluster_rule_user_feedback`](#table-cluster_rule_user_feedback)
    * [Table `cluster_user_rule_disable_feedback`](#table-cluster_user_rule_disable_feedback)
    * [Table `consumer_error`](#table-consumer_error)
    * [Table `migration_info `](#table-migration_info-)
    * [Table `rule_hit`](#table-rule_hit)
    * [Table `recommendation`](#table-recommendation)
    * [Database tables affected by this service](#database-tables-affected-by-this-service)
* [Package manifest](#package-manifest)

<!-- vim-markdown-toc -->


## Description

Simple service that can be used to identify clusters, for which we are keeping
very old data (>30 days) in the database. This means that the cluster is no
longer available or that the customer has disabled the Insights Operator,
either way it means that these data are no longer relevant to us and should be
pruned.

Such clusters can be detected very easily by checking the timestamps stored
(along other information) in the `report` table in Insights Results Aggregator
database.

Additionally the service can delete records for provided list of clusters. This
means the clusters can be deleted on demand and in controlled environment.

### Usage

```
Usage of cleaner:
  -authors
        show authors
  -cleanup
        perform database cleanup
  -clusters string
        list of clusters to cleanup
  -fill-in-db
        fill-in database by test data
  -max-age string
        max age for displaying old records
  -multiple-rule-disable
        list clusters with the same rule(s) disabled by different users
  -output string
        filename for old cluster listing
  -show-configuration
        show configuration
  -summary
        print summary table after cleanup
  -vacuum
        vacuum database
  -version
        show cleaner version
```

### Default operation

Currently this service just displays such clusters (cluster IDs) and do nothing
else - i.e. the results are not deleted by default.

### Data cleanup

In order to delete data, the `-cleanup` command line option needs to be used.
In this case the file `cluster_list.txt` should contain list of clusters to be
deleted.

Optionally it is possible to specify list of clusters to be cleaned up by using
the `clusters ...` command line option.

### Test data generation

Command line option `-fill-in-db` can be used to insert some test data into
database. Don't use it on production, of course.

### Exit status

```
0 the tool finished with success
1 is returned in case of any storage-related error
2 is returned in case the fill-in DB operation failed
3 is returned when DB cleanup operation failed for any reason
```

### Building

Go version 1.14 or newer is required to build this tool.

```
make build
```

### Makefile targets

Available targets are:

```
clean                Run go clean
build                Keep this rule for compatibility
fmt                  Run go fmt -w for all sources
lint                 Run golint
vet                  Run go vet. Report likely mistakes in source code
cyclo                Run gocyclo
ineffassign          Run ineffassign checker
shellcheck           Run shellcheck
errcheck             Run errcheck
goconst              Run goconst checker
gosec                Run gosec checker
abcgo                Run ABC metrics checker
json-check           Check all JSONs for basic syntax
style                Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts
run                  Build the project and executes the binary
test                 Run the unit tests
before_commit        Checks done before commit
help                 Show this help screen
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

## Database structure

List of tables:

```
                       List of relations
 Schema |                Name                | Type  |  Owner
--------+------------------------------------+-------+----------
 public | cluster_rule_toggle                | table | postgres
 public | cluster_rule_user_feedback         | table | postgres
 public | cluster_user_rule_disable_feedback | table | postgres
 public | consumer_error                     | table | postgres
 public | migration_info                     | table | postgres
 public | report                             | table | postgres
 public | rule_hit                           | table | postgres
 public | recommendation                     | table | postgres
 public | advisor_ratings                    | table | postgres
 public | rule_disable                       | table | postgres

```

### Table `report`

```
     Column      |            Type             |     Modifiers
-----------------+-----------------------------+--------------------
 org_id          | integer                     | not null
 cluster         | character varying           | not null
 report          | character varying           | not null
 reported_at     | timestamp without time zone |
 last_checked_at | timestamp without time zone |
 kafka_offset    | bigint                      | not null default 0
Indexes:
    "report_pkey" PRIMARY KEY, btree (org_id, cluster)
    "report_cluster_key" UNIQUE CONSTRAINT, btree (cluster)
    "report_kafka_offset_btree_idx" btree (kafka_offset)
Referenced by:
    TABLE "cluster_rule_user_feedback" CONSTRAINT "cluster_rule_user_feedback_cluster_id_fkey" FOREIGN KEY (cluster_id) REFERENCES report(cluster) ON DELETE CASCADE
```

### Table `cluster_rule_toggle`

```
   Column    |            Type             | Modifiers
-------------+-----------------------------+-----------
 cluster_id  | character varying           | not null
 rule_id     | character varying           | not null
 user_id     | character varying           | not null
 disabled    | smallint                    | not null
 disabled_at | timestamp without time zone |
 enabled_at  | timestamp without time zone |
 updated_at  | timestamp without time zone | not null
Indexes:
    "cluster_rule_toggle_pkey" PRIMARY KEY, btree (cluster_id, rule_id, user_id)
Check constraints:
    "cluster_rule_toggle_disabled_check" CHECK (disabled >= 0 AND disabled <= 1)
```

### Table `cluster_rule_user_feedback`

```
   Column   |            Type             | Modifiers
------------+-----------------------------+-----------
 cluster_id | character varying           | not null
 rule_id    | character varying           | not null
 user_id    | character varying           | not null
 message    | character varying           | not null
 user_vote  | smallint                    | not null
 added_at   | timestamp without time zone | not null
 updated_at | timestamp without time zone | not null
Indexes:
    "cluster_rule_user_feedback_pkey1" PRIMARY KEY, btree (cluster_id, rule_id, user_id)
Foreign-key constraints:
    "cluster_rule_user_feedback_cluster_id_fkey" FOREIGN KEY (cluster_id) REFERENCES report(cluster) ON DELETE CASCADE
```

### Table `cluster_user_rule_disable_feedback`

```
   Column   |            Type             | Modifiers
------------+-----------------------------+-----------
 cluster_id | character varying           | not null
 user_id    | character varying           | not null
 rule_id    | character varying           | not null
 message    | character varying           | not null
 added_at   | timestamp without time zone | not null
 updated_at | timestamp without time zone | not null
Indexes:
    "cluster_user_rule_disable_feedback_pkey" PRIMARY KEY, btree (cluster_id, user_id, rule_id)
```

### Table `consumer_error`

```
             Table "public.consumer_error"
    Column    |            Type             | Modifiers
--------------+-----------------------------+-----------
 topic        | character varying           | not null
 partition    | integer                     | not null
 topic_offset | integer                     | not null
 key          | character varying           |
 produced_at  | timestamp without time zone | not null
 consumed_at  | timestamp without time zone | not null
 message      | character varying           |
 error        | character varying           | not null
Indexes:
    "consumer_error_pkey" PRIMARY KEY, btree (topic, partition, topic_offset)
```

### Table `migration_info `

```
 Column  |  Type   | Modifiers
---------+---------+-----------
 version | integer | not null
```

### Table `rule_hit`

```
    Column     |       Type        | Modifiers
---------------+-------------------+-----------
 org_id        | integer           | not null
 cluster_id    | character varying | not null
 rule_fqdn     | character varying | not null
 error_key     | character varying | not null
 template_data | character varying | not null
Indexes:
    "rule_hit_pkey" PRIMARY KEY, btree (cluster_id, org_id, rule_fqdn, error_key)
```

### Table `recommendation`

```
                                 Table "public.recommendation"
   Column   |            Type             | Collation | Nullable |           Default            
------------+-----------------------------+-----------+----------+------------------------------
 org_id     | integer                     |           | not null | 
 cluster_id | character varying           |           | not null | 
 rule_fqdn  | text                        |           | not null | 
 error_key  | character varying           |           | not null | 
 rule_id    | character varying           |           | not null | '.'::character varying
 created_at | timestamp without time zone |           |          | timezone('utc'::text, now())
Indexes:
    "recommendation_pk" PRIMARY KEY, btree (org_id, cluster_id, rule_fqdn, error_key)
```

### Database tables affected by this service

Figuring out which reports are older than the specified time:
* `report`

Actually cleaning the data for given cluster:

* `report` by `cluster`
* `cluster_rule_toggle` by `cluster_id`
* `cluster_rule_user_feedback` by `cluster_id`
* `cluster_user_rule_disable_feedback` by `cluster_id`
* `rule_hit` by `cluster_id`
* `recommendation` by `cluster_id`

## Package manifest

Package manifest is available at [docs/manifest.txt](docs/manifest.txt).

