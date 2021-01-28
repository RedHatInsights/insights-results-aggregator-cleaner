---
layout: default
---
# Description

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
