/*
Copyright © 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// Definition of custom data types used by this tool.

// ClusterName represents name of cluster in format
// c8590f31-e97e-4b85-b506-c45ce1911a12 (it must be proper UUID).
type ClusterName string

// ClusterList represents a list of cluster names/ids (see ClusterName data
// type declared above)
type ClusterList []ClusterName

// TableAndKey represents a key for given table used by cleanup process. Each
// row is deleted by specifying table name and a key
type TableAndKey struct {
	TableName string
	KeyName   string
}

// Summary represents summary info to be displayed in a table after cleanup
// part
type Summary struct {
	ProperClusterEntries   int
	ImproperClusterEntries int
	DeletionsForTable      map[string]int
}
