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
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

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
		ShowVersion:               true,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
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
		ShowVersion:               false,
		ShowAuthors:               true,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
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
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         true,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
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

// TestDoSelectedOperationVacuumDatabase checks the function
// vacuumDB called via doSelectedOperation function
func TestDoSelectedOperationVacuumDatabase(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            true,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
	}

	// call tested function
	code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, code, main.ExitStatusPerformVacuumError)
}

// TestDoSelectedOperationPerformCleanup checks the function
// performCleanup called via doSelectedOperation function
func TestDoSelectedOperationPerformCleanup(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            true,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
	}

	// call tested function
	code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, code, main.ExitStatusPerformCleanupError)
}

// TestDoSelectedOperationDetectMultipleRuleDisable checks the function
// detectMultipleRuleDisable called via doSelectedOperation function
func TestDoSelectedOperationDetectMultipleRuleDisable(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: true,
		FillInDatabase:            false,
	}

	// call tested function
	code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, code, main.ExitStatusStorageError)
}

// TestDoSelectedOperationFillInDatabase checks the function
// fillInDatabase called via doSelectedOperation function
func TestDoSelectedOperationFillInDatabase(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            true,
	}

	// call tested function
	code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, code, main.ExitStatusFillInStorageError)
}

// TestDoSelectedOperationDefaultOperation checks the function
// displayOldRecords called via doSelectedOperation function
func TestDoSelectedOperationDefaultOperation(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:               false,
		ShowAuthors:               false,
		ShowConfiguration:         false,
		VacuumDatabase:            false,
		PerformCleanup:            false,
		DetectMultipleRuleDisable: false,
		FillInDatabase:            false,
	}

	// call tested function
	code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, code, main.ExitStatusStorageError)
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

// TestPrintSummaryTableOneTableDeletion check the behaviour of function
// PrintSummaryTable for summary with one deletion in one table.
func TestPrintSummaryTableOneTableDeletion(t *testing.T) {
	const expected = `+--------------------------------+-------+
|            SUMMARY             | COUNT |
+--------------------------------+-------+
| Proper cluster entries         |     0 |
| Improper cluster entries       |     0 |
|                                |       |
| Deletions from table 'TABLE_X' |     1 |
+--------------------------------+-------+
|        TOTAL DELETIONS         |   1   |
+--------------------------------+-------+
`

	deletions := map[string]int{
		"TABLE_X": 1,
	}
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		summary := main.Summary{
			ProperClusterEntries:   0,
			ImproperClusterEntries: 0,
			DeletionsForTable:      deletions,
		}
		main.PrintSummaryTable(summary)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	assert.Contains(t, output, expected)
}

// TestPrintSummaryTableTwoTablesDeletions check the behaviour of function
// PrintSummaryTable for summary with multiple deletions in two tables.
func TestPrintSummaryTableTwoTablesDeletions(t *testing.T) {
	// we work with map and there is no guarantees which order will be choosen in runtime
	const expected1 = `+--------------------------------+-------+
|            SUMMARY             | COUNT |
+--------------------------------+-------+
| Proper cluster entries         |     0 |
| Improper cluster entries       |     0 |
|                                |       |
| Deletions from table 'TABLE_X' |     1 |
| Deletions from table 'TABLE_Y' |     2 |
+--------------------------------+-------+
|        TOTAL DELETIONS         |   3   |
+--------------------------------+-------+
`
	const expected2 = `+--------------------------------+-------+
|            SUMMARY             | COUNT |
+--------------------------------+-------+
| Proper cluster entries         |     0 |
| Improper cluster entries       |     0 |
|                                |       |
| Deletions from table 'TABLE_Y' |     2 |
| Deletions from table 'TABLE_X' |     1 |
+--------------------------------+-------+
|        TOTAL DELETIONS         |   3   |
+--------------------------------+-------+
`

	deletions := map[string]int{
		"TABLE_X": 1,
		"TABLE_Y": 2,
	}
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		summary := main.Summary{
			ProperClusterEntries:   0,
			ImproperClusterEntries: 0,
			DeletionsForTable:      deletions,
		}
		main.PrintSummaryTable(summary)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	// again: we work with map and there is no guarantees which order will
	// be choosen in runtime
	if output != expected1 && output != expected2 {
		t.Error("Unexpected output", output)
	}
}

// TestVacuumDBPositiveCase check the function vacuumDB when the DB
// operation pass without any error
func TestVacuumDBPositiveCase(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	expectedVacuum := "VACUUM VERBOSE;"
	mock.ExpectExec(expectedVacuum).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectClose()

	// call the tested function
	status, err := main.VacuumDB(connection)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check the status
	assert.Equal(t, status, main.ExitStatusOK)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestVacuumDBNegativeCase check the function vacuumDB when the DB
// operation pass with an error
func TestVacuumDBNegativeCase(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	expectedVacuum := "VACUUM VERBOSE;"
	mock.ExpectExec(expectedVacuum).WillReturnError(mockedError)

	mock.ExpectClose()

	// call the tested function
	status, err := main.VacuumDB(connection)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, status, main.ExitStatusPerformVacuumError)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestVacuumDBNoConnection check the function vacuumDB when the
// connection to DB is not established
func TestVacuumDBNoConnection(t *testing.T) {
	// call the tested function
	status, err := main.VacuumDB(nil)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, status, main.ExitStatusPerformVacuumError)
}

// TestCleanupNoConnection check the function cleanup when the
// connection to DB is not established
func TestCleanupNoConnection(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}

	configuration.Cleaner = main.CleanerConfiguration{
		ClusterListFile: "tests/cluster_list.txt",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// call the tested function
	status, err := main.Cleanup(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.cleanup")

	// check the status
	assert.Equal(t, status, main.ExitStatusPerformCleanupError)
}

// TestCleanupOnReadClusterListError check the function cleanup when
// cluster list can not be retrieved
func TestCleanupOnReadClusterListError(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}

	configuration.Cleaner = main.CleanerConfiguration{
		// non-existent file
		ClusterListFile: "tests/this_dos_not_exists.txt",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// call the tested function
	status, err := main.Cleanup(&configuration, nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.cleanup")

	// check the status
	assert.Equal(t, status, main.ExitStatusPerformCleanupError)
}

// TestCleanup check the function cleanup when
// summary table should not be printed
func TestCleanup(t *testing.T) {
	// prepare new mocked connection to database
	connection, _, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}

	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge:          "3 days",
		ClusterListFile: "cluster_list.txt",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		PrintSummaryTable: false,
	}

	// call the tested function
	status, err := main.Cleanup(&configuration, connection, cliFlags)

	// error is not expected
	assert.NoError(t, err, "error is not expected while calling main.cleanup")

	// check the status
	assert.Equal(t, status, main.ExitStatusOK)
}

// TestCleanupPrintSummaryTable check the function cleanup when
// summary table should be printed
func TestCleanupPrintSummaryTable(t *testing.T) {
	// prepare new mocked connection to database
	connection, _, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}

	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge:          "3 days",
		ClusterListFile: "cluster_list.txt",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		PrintSummaryTable: true,
	}

	// call the tested function
	status, err := main.Cleanup(&configuration, connection, cliFlags)

	// error is not expected
	assert.NoError(t, err, "error is not expected while calling main.cleanup")

	// check the status
	assert.Equal(t, status, main.ExitStatusOK)
}

// TestCleanupCheckSummaryTableContent check the function cleanup when
// summary table should be printed
func TestCleanupCheckSummaryTableContent(t *testing.T) {
	var expectedOutputLines []string = []string{
		"+-----------------------------------------------------------+-------+",
		"|                          SUMMARY                          | COUNT |",
		"+-----------------------------------------------------------+-------+",
		"| Proper cluster entries                                    |     5 |",
		"| Improper cluster entries                                  |     2 |",
		"|                                                           |       |",
		"| Deletions from table 'cluster_rule_user_feedback'         |     0 |",
		"| Deletions from table 'cluster_user_rule_disable_feedback' |     0 |",
		"| Deletions from table 'rule_hit'                           |     0 |",
		"| Deletions from table 'recommendation'                     |     0 |",
		"| Deletions from table 'report_info'                        |     0 |",
		"| Deletions from table 'report'                             |     0 |",
		"| Deletions from table 'cluster_rule_toggle'                |     0 |",
		"+-----------------------------------------------------------+-------+",
		"|                      TOTAL DELETIONS                      |   0   |",
		"+-----------------------------------------------------------+-------+",
	}

	// prepare new mocked connection to database
	connection, _, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}

	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge:          "3 days",
		ClusterListFile: "cluster_list.txt",
	}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		PrintSummaryTable: true,
	}

	var status int

	// call the tested function
	output, err := capture.StandardOutput(func() {
		status, _ = main.Cleanup(&configuration, connection, cliFlags)
	})

	// check the captured text
	checkCapture(t, err)

	// check if captured text contains expected summary table
	for _, expectedLine := range expectedOutputLines {
		assert.Contains(t, output, expectedLine)
	}

	// check the status
	assert.Equal(t, status, main.ExitStatusOK)
}

// TestDetectMultipleRuleDisable check the function detectMultipleRuleDisable when the
// connection to DB is not established
func TestDetectMultipleRuleDisable(t *testing.T) {
	// stub for CLI flags needed to call the tested function
	cliFlags := main.CliFlags{}

	// call the tested function with null connection
	status, err := main.DetectMultipleRuleDisable(nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.cleanup")

	// check the status
	assert.Equal(t, status, main.ExitStatusStorageError)
}

// TestFillInDatabase checks the basic behaviour of
// fillInDatabase function.
func TestFillInDatabase(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	clusterNames := [...]string{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-1111-1111-111111111111",
		"5d5892d4-1f74-4ccf-91af-548dfc9767aa",
	}

	for _, clusterName := range clusterNames {
		mock.ExpectExec("INSERT INTO report").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_rule_toggle").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_rule_user_feedback").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_user_rule_disable_feedback").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO rule_hit").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectClose()

	exitCode, err := main.FillInDatabase(connection)
	assert.NoError(t, err, "error not expected while calling tested function")
	assert.Equal(t, exitCode, main.ExitStatusOK)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestFillInDatabaseOnError checks the basic behaviour of
// fillInDatabase function.
func TestFillInDatabaseOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	clusterNames := [...]string{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-1111-1111-111111111111",
		"5d5892d4-1f74-4ccf-91af-548dfc9767aa",
	}

	for _, clusterName := range clusterNames {
		mock.ExpectExec("INSERT INTO report").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_rule_toggle").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_rule_user_feedback").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO cluster_user_rule_disable_feedback").WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO rule_hit").WithArgs(clusterName).WillReturnError(mockedError)
	}

	mock.ExpectClose()

	exitCode, err := main.FillInDatabase(connection)
	assert.Error(t, err, "error is expected while calling tested function")
	assert.Equal(t, exitCode, main.ExitStatusFillInStorageError)
	assert.Equal(t, err, mockedError)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestFillInDatabaseNoConnection checks the basic behaviour of
// fillInDatabase function when connection is not established.
func TestFillInDatabaseNoConnection(t *testing.T) {
	exitCode, err := main.FillInDatabase(nil)
	assert.Error(t, err, "error is expected while calling tested function")
	assert.Equal(t, exitCode, main.ExitStatusFillInStorageError)
}

// TestDisplayOldRecordsNoConnection checks the basic behaviour of
// displayOldRecords function when connection is not established.
func TestDisplayOldRecordsNoConnection(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge: "3 days",
	}

	cliFlags := main.CliFlags{}

	exitCode, err := main.DisplayOldRecords(&configuration, nil, cliFlags)
	assert.Error(t, err, "error is expected while calling tested function")
	assert.Equal(t, exitCode, main.ExitStatusStorageError)
}

// TestDisplayOldRecordsProperConnection checks the basic behaviour of
// displayOldRecords function when connection is established.
func TestDisplayOldRecordsProperConnection(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge: "3 days",
	}

	// command line flags
	cliFlags := main.CliFlags{}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow(cluster1ID, reportedAt, updatedAt)

	// expected queries performed by tested function
	expectedQuery1 := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)

	expectedQuery2 := "SELECT org_id, rule_fqdn, error_key, rule_id, rating, last_updated_at FROM advisor_ratings WHERE last_updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY last_updated_at"
	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows)

	expectedQuery3 := "SELECT topic, partition, topic_offset, key, consumed_at, message FROM consumer_error WHERE consumed_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY consumed_at"
	mock.ExpectQuery(expectedQuery3).WillReturnRows(rows)

	mock.ExpectClose()

	// call the tested function
	exitCode, err := main.DisplayOldRecords(&configuration, connection, cliFlags)

	// and check its output
	assert.NoError(t, err, "error is not expected while calling tested function")
	assert.Equal(t, main.ExitStatusOK, exitCode)
}

// TestDetectMultipleRuleDisablesNoConnection check the function
// detectMultipleRuleDisable when the connection to DB is not established
func TestDetectMultipleRuleDisablesNoConnection(t *testing.T) {
	// command line flags
	cliFlags := main.CliFlags{}

	// call the tested function
	status, err := main.DetectMultipleRuleDisable(nil, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.vacuumDB")

	// check the status
	assert.Equal(t, status, main.ExitStatusStorageError)
}

// TestDetectMultipleRuleDisablesProperConnection check the function
// detectMultipleRuleDisable when the connection to DB is established
func TestDetectMultipleRuleDisablesProperConnection(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// command line flags
	cliFlags := main.CliFlags{}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected queries performed by tested function
	expectedQuery1 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	expectedQuery2 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	status, err := main.DetectMultipleRuleDisable(connection, cliFlags)

	// error is not expected
	assert.NoError(t, err, "error is not expected while calling main.detectMultipleRuleDisable")

	// check the status
	assert.Equal(t, status, main.ExitStatusOK)
}

// TestDetectMultipleRuleDisablesOnError1 check the function
// detectMultipleRuleDisable when DB error is thrown
func TestDetectMultipleRuleDisablesOnError1(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// command line flags
	cliFlags := main.CliFlags{}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected queries performed by tested function
	expectedQuery1 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	expectedQuery2 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)
	mock.ExpectQuery(expectedQuery2).WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	status, err := main.DetectMultipleRuleDisable(connection, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.detectMultipleRuleDisable")
	assert.Equal(t, err, mockedError)

	// check the status
	assert.Equal(t, status, main.ExitStatusStorageError)
}

// TestDetectMultipleRuleDisablesOnError2 check the function
// detectMultipleRuleDisable when DB error is thrown
func TestDetectMultipleRuleDisablesOnError2(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// command line flags
	cliFlags := main.CliFlags{}

	// expected queries performed by tested function
	expectedQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	status, err := main.DetectMultipleRuleDisable(connection, cliFlags)

	// error is expected
	assert.Error(t, err, "error is expected while calling main.detectMultipleRuleDisable")
	assert.Equal(t, err, mockedError)

	// check the status
	assert.Equal(t, status, main.ExitStatusStorageError)
}
