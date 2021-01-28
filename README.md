# insights-results-aggregator-cleaner

[![GoDoc](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner?status.svg)](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-results-aggregator-cleaner/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-results-aggregator-cleaner)](https://goreportcard.com/report/github.com/RedHatInsights/insights-results-aggregator-cleaner)
[![Build Status](https://travis-ci.org/RedHatInsights/insights-results-aggregator-cleaner.svg?branch=master)](https://travis-ci.org/RedHatInsights/insights-results-aggregator-cleaner)

## Description

Simple service that can be used to identify clusters, for which we are keeping
very old data (>30 days) in the database. This means that the cluster is no
longer available or that the customer has disabled the Insights Operator,
either way it means that these data are no longer relevant to us and should be
pruned.

Such clusters can be detected very easily by checking the timestamps stored
(along other information) in the `result` table in Insights Results Aggregator
database.

Currently this service just displays such clusters (cluster IDs) and do nothing
else - i.e. the results are not deleted.
