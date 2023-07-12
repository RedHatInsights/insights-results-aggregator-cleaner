/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

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

package main_test

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner_test.html

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"
	"os"
	"testing"

	main "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func checkCapture(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// TestShowVersion checks the function showVersion
func TestShowVersion(t *testing.T) {
	const expected = "Insights Results Aggregator Cleaner version 1.0\n"

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowVersion()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expected)
}

// TestShowAuthors checks the function showAuthors
func TestShowAuthors(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowAuthors()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Red Hat Inc.")
}

// TestShowConfiguration checks the function ShowConfiguration
func TestShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Storage = main.StorageConfiguration{
		Driver:     "postgres",
		PGUsername: "foo",
		PGPassword: "bar",
		PGHost:     "baz",
		PGDBName:   "aggregator",
		PGParams:   ""}
	configuration.Logging = main.LoggingConfiguration{
		Debug:    true,
		LogLevel: ""}
	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge:          "3 days",
		ClusterListFile: "cluster_list.txt"}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.New(os.Stderr))

		main.ShowConfiguration(&configuration)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Driver")
	assert.Contains(t, output, "Level")
	assert.Contains(t, output, "Records max age")
}

func TestIsValidUUID(t *testing.T) {
	type UUID struct {
		id    string
		valid bool
	}

	uuids := []UUID{
		UUID{
			id:    "",
			valid: false,
		},
		UUID{
			id:    "00000000-0000-0000-0000-000000000000",
			valid: true,
		},
		UUID{
			id:    "5d5892d4-1f74-4ccf-91af-548dfc9767aa",
			valid: true,
		},
		UUID{ // x at beginning
			id:    "xd5892d4-1f74-4ccf-91af-548dfc9767aa",
			valid: false,
		},
		UUID{ // wrong separator
			id:    "5d5892d4-1f74-4cc-f91af-548dfc9767aa",
			valid: false,
		},
	}

	for _, uuid := range uuids {
		v := main.IsValidUUID(uuid.id)
		assert.Equal(t, v, uuid.valid)

	}

}

// TestDoSelectedOperationShowVersion checks the function showVersion called
// via doSelectedOperation function
func TestDoSelectedOperationShowVersion(t *testing.T) {
	const expected = "Insights Results Aggregator Cleaner version 1.0\n"

	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       true,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expected)
}

// TestDoSelectedOperationShowAuthors checks the function showAuthors called
// via doSelectedOperation function
func TestDoSelectedOperationShowAuthors(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       true,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Red Hat Inc.")
}

// TestDoSelectedOperationShowConfiguration checks the function
// showConfiguration called via doSelectedOperation function
func TestDoSelectedOperationShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: true,
	}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.New(os.Stderr))

		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Driver")
	assert.Contains(t, output, "Level")
	assert.Contains(t, output, "Records max age")
}

// TestReadClusterList checks the function readClusterList from
// cleaner.go using correct cluster list file
func TestReadClusterList(t *testing.T) {
	// cluster list file with 8 clusters in total:
	// 5 correct cluster names
	// 3 incorrect cluster names
	clusterList, improperClusterCount, err := main.ReadClusterList("tests/cluster_list.txt", "")

	// file is correct - no errors should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 3)
	assert.Len(t, clusterList, 5)

	// finally check actual cluster names
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("55d892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("5d5892d3-1f74-4ccf-91af-548dfc9767bb"))
	assert.Contains(t, clusterList, main.ClusterName("00000000-0000-0000-0000-000000000000"))
	assert.Contains(t, clusterList, main.ClusterName("11111111-1111-1111-1111-111111111111"))
}

// TestReadClusterListNoFile checks the function readClusterList from
// cleaner.go in case the cluster list file does not exists
func TestReadClusterListNoFile(t *testing.T) {
	_, _, err := main.ReadClusterListFromFile("tests/this_does_not_exists.txt")

	// in this case we expect error to be thrown
	assert.Error(t, err)
}

// TestReadClusterListCLICase1 checks the function readClusterList from
// cleaner.go using provided CLI arguments
func TestReadClusterListCLICase1(t *testing.T) {
	// just one cluster name is specified on CLI
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa"
	clusterList, improperClusterCount, err := main.ReadClusterList("tests/cluster_list.txt", input)

	// input is correct - no errors should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 1)

	// finally check actual cluster names (only one name expected)
	assert.Contains(t, clusterList, main.ClusterName(input))
}

// TestReadClusterList checks the function readClusterList from
// cleaner.go using provided CLI arguments
func TestReadClusterListCLICase2(t *testing.T) {
	// two cluster names are specified on CLI
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa,ffffffff-1f74-4ccf-91af-548dfc9767aa"

	// input is correct - no errors should be thrown
	clusterList, improperClusterCount, err := main.ReadClusterList("tests/cluster_list.txt", input)

	// both cluster names are correct
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 2)

	// finally check actual cluster names
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("ffffffff-1f74-4ccf-91af-548dfc9767aa"))
}

// TestReadClusterList checks the function readClusterList from
// cleaner.go using provided CLI arguments
func TestReadClusterListCLICase3(t *testing.T) {
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa,this-is-not-correct"
	clusterList, improperClusterCount, err := main.ReadClusterList("tests/cluster_list.txt", input)

	// just the first cluster name is correct
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 1)
	assert.Len(t, clusterList, 1)

	// finally check actual cluster names (just one correct cluster name is expected)
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
}

// TestReadClusterList checks the function readClusterList from
// cleaner.go using provided CLI arguments
func TestReadClusterListCLICase4(t *testing.T) {
	input := "this-is-not-correct,this-also-is-not-correct"
	clusterList, improperClusterCount, err := main.ReadClusterList("tests/cluster_list.txt", input)

	// both cluster names are incorrect, but the whole algorithm does not throw an error
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 2)
	assert.Len(t, clusterList, 0)
}

// TestReadClusterListFromFile checks the function readClusterListFromFile from
// cleaner.go using correct cluster list file with 5 correct clusters and 3
// incorrect clusters.
func TestReadClusterListFromFile(t *testing.T) {
	// cluster list file with 8 clusters in total:
	// 5 correct cluster names
	// 3 incorrect cluster names
	clusterList, improperClusterCount, err := main.ReadClusterListFromFile("tests/cluster_list.txt")

	// file is correct - no errors should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 3)
	assert.Len(t, clusterList, 5)

	// finally check actual cluster names
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("55d892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("5d5892d3-1f74-4ccf-91af-548dfc9767bb"))
	assert.Contains(t, clusterList, main.ClusterName("00000000-0000-0000-0000-000000000000"))
	assert.Contains(t, clusterList, main.ClusterName("11111111-1111-1111-1111-111111111111"))
}

// TestReadClusterListFromFileNoFile checks the function
// readClusterListFromFile from cleaner.go in case the cluster list file does
// not exists
func TestReadClusterListFromFileNoFile(t *testing.T) {
	_, _, err := main.ReadClusterListFromFile("tests/this_does_not_exists.txt")

	// file does not exist -> error should be thrown
	assert.Error(t, err)
}

// TestReadClusterListFromFileEmptyFile checks the function
// readClusterListFromFile from cleaner.go in case the special /dev/null file is to be read
func TestReadClusterListFromFileEmptyFile(t *testing.T) {
	clusterList, improperClusterCount, err := main.ReadClusterListFromFile("tests/empty_cluster_list.txt")

	// it's empty so no error should be reported
	assert.NoError(t, err)

	// and the content should be empty
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 0)
}

// TestReadClusterListFromFileNullFile checks the function
// readClusterListFromFile from cleaner.go in case the special /dev/null file is to be read
func TestReadClusterListFromFileNullFile(t *testing.T) {
	clusterList, improperClusterCount, err := main.ReadClusterListFromFile("/dev/null")

	// it's empty so no error should be reported
	assert.NoError(t, err)

	// and the content should be empty
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 0)
}

// TestReadClusterListFromCLIArgumentEmptyInput check the function
// readClusterListFromCLIArgument from cleaner.go
func TestReadClusterListFromCLIArgumentEmptyInput(t *testing.T) {
	clusterList, improperClusterCount, err := main.ReadClusterListFromCLIArgument("")

	// it's empty so no error should be reported
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 1)
	assert.Len(t, clusterList, 0)
}

// TestReadClusterListFromCLIArgumentOneCluster check the function
// readClusterListFromCLIArgument from cleaner.go
func TestReadClusterListFromCLIArgumentOneCluster(t *testing.T) {
	// only one (correct) cluster
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa"
	clusterList, improperClusterCount, err := main.ReadClusterListFromCLIArgument(input)

	// input is correct -> no error should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 1)

	// finally check actual cluster names (just one cluster name is expected)
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
}

// TestReadClusterListFromCLIArgumentOneIncorrectCluster check the function
// readClusterListFromCLIArgument from cleaner.go
func TestReadClusterListFromCLIArgumentOneIncorrectCluster(t *testing.T) {
	// only one (incorrect) cluster
	input := "foo-bar-baz"
	clusterList, improperClusterCount, err := main.ReadClusterListFromCLIArgument(input)

	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 1)
	assert.Len(t, clusterList, 0)
}

// TestReadClusterListFromCLIArgumentTwoClusters check the function
// readClusterListFromCLIArgument from cleaner.go
func TestReadClusterListFromCLIArgumentTwoClusters(t *testing.T) {
	// both clusters are correct
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa,5d5892d4-1f74-4ccf-91af-548dfc9767bb"
	clusterList, improperClusterCount, err := main.ReadClusterListFromCLIArgument(input)

	// input is correct -> no error should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 0)
	assert.Len(t, clusterList, 2)

	// finally check actual cluster names (just one correct cluster name is expected)
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767bb"))
}

// TestReadClusterListFromCLIArgumentImproperCluster check the function
// readClusterListFromCLIArgument from cleaner.go
func TestReadClusterListFromCLIArgumentImproperCluster(t *testing.T) {
	// first cluster is correct, second one incorrect
	input := "5d5892d4-1f74-4ccf-91af-548dfc9767aa,foo-bar-baz"
	clusterList, improperClusterCount, err := main.ReadClusterListFromCLIArgument(input)

	// no error should be thrown
	assert.NoError(t, err)

	// check returned content
	assert.Equal(t, improperClusterCount, 1)
	assert.Len(t, clusterList, 1)

	// finally check actual cluster names (just one correct cluster name is expected)
	assert.Contains(t, clusterList, main.ClusterName("5d5892d4-1f74-4ccf-91af-548dfc9767aa"))
}

// TestPrintSummaryTableBasicCase check the behaviour of function
// PrintSummaryTable for summary with zero changes made in database.
func TestPrintSummaryTableBasicCase(t *testing.T) {
	const expected = `+--------------------------+-------+
|         SUMMARY          | COUNT |
+--------------------------+-------+
| Proper cluster entries   |     0 |
| Improper cluster entries |     0 |
|                          |       |
+--------------------------+-------+
|     TOTAL DELETIONS      |   0   |
+--------------------------+-------+
`

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		summary := main.Summary{
			ProperClusterEntries:   0,
			ImproperClusterEntries: 0,
			DeletionsForTable:      make(map[string]int),
		}
		main.PrintSummaryTable(summary)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	assert.Contains(t, output, expected)
}

// TestPrintSummaryTableProperClusterEntries check the behaviour of function
// PrintSummaryTable for summary with non zero changes made in database.
func TestPrintSummaryTableProperClusterEntries(t *testing.T) {
	const expected = `+--------------------------+-------+
|         SUMMARY          | COUNT |
+--------------------------+-------+
| Proper cluster entries   |    42 |
| Improper cluster entries |     0 |
|                          |       |
+--------------------------+-------+
|     TOTAL DELETIONS      |   0   |
+--------------------------+-------+
`

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		summary := main.Summary{
			ProperClusterEntries:   42,
			ImproperClusterEntries: 0,
			DeletionsForTable:      make(map[string]int),
		}
		main.PrintSummaryTable(summary)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	assert.Contains(t, output, expected)
}

// TestPrintSummaryTableImproperClusterEntries check the behaviour of function
// PrintSummaryTable for summary with non zero changes made in database.
func TestPrintSummaryTableImproperClusterEntries(t *testing.T) {
	const expected = `+--------------------------+-------+
|         SUMMARY          | COUNT |
+--------------------------+-------+
| Proper cluster entries   |     0 |
| Improper cluster entries |    42 |
|                          |       |
+--------------------------+-------+
|     TOTAL DELETIONS      |   0   |
+--------------------------+-------+
`

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		summary := main.Summary{
			ProperClusterEntries:   0,
			ImproperClusterEntries: 42,
			DeletionsForTable:      make(map[string]int),
		}
		main.PrintSummaryTable(summary)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	assert.Contains(t, output, expected)
}
