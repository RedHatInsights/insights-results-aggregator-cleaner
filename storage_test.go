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
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/database_test.html

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	cleaner "github.com/RedHatInsights/insights-results-aggregator-cleaner"
	"github.com/stretchr/testify/assert"
)

const (
	cluster1ID   = "123e4567-e89b-12d3-a456-426614173998"
	cluster2ID   = "567e4567-4321-12d3-a456-426614173777"
	rule1ID      = "rule.test|KEY"
	defaultOrgID = 42
)

// checkConnectionClose function performs mocked DB closing operation and checks
// if the connection is properly closed from unit tests.
func checkConnectionClose(t *testing.T, connection *sql.DB) {
	// connection to mocked DB needs to be closed properly
	err := connection.Close()

	assert.NoError(t, err)
}

// checkAllExpectations function checks if all database-related operations have
// been really met.
func checkAllExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	// check if all expectations were met
	err := mock.ExpectationsWereMet()

	assert.NoError(t, err)
}

// expectOrgIDQuery mocks an expect of a repetetive query to check whether cluster
// belongs to given org
func expectOrgIDQuery(mock sqlmock.Sqlmock) {
	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id"})
	rows.AddRow(defaultOrgID)

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
}

// TestReadOrgIDNoResults checks the function readOrgID.
func TestReadOrgIDNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check the org ID returned from tested function
	if org_id != -1 {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestReadOrgIDResults checks the function readOrgID.
func TestReadOrgIDResult(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	expectOrgIDQuery(mock)

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check the org ID returned from tested function
	if org_id != defaultOrgID {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestReadOrgIDOnError checks error handling in function readOrgID.
func TestReadOrgIDOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614173999")
	if err == nil {
		t.Fatalf("error was expected while updating stats")
	}

	// check the org ID returned from tested function
	if org_id != -1 {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformDisplayMultipleRuleDisableNoResults checks the basic behaviour of
// performDisplayMultipleRuleDisable function.
func TestPerformDisplayMultipleRuleDisableNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// first query to be performed
	query1 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_rule_toggle
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`
	// call the tested function
	err = cleaner.PerformDisplayMultipleRuleDisable(connection, nil, query1, "cluster_rule_toggle")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformDisplayMultipleRuleDisableOnError checks the error handling
// ability in performDisplayMultipleRuleDisable function.
func TestPerformDisplayMultipleRuleDisableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// first query to be performed
	query1 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_rule_toggle
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`
	// call the tested function
	err = cleaner.PerformDisplayMultipleRuleDisable(connection, nil, query1, "cluster_rule_toggle")
	if err == nil {
		t.Fatalf("error was expected while updating stats")
	}

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformDisplayMultipleRuleDisableOnScanError checks the error handling
// ability in performDisplayMultipleRuleDisable function regarding wrong values returned from query.
func TestPerformDisplayMultipleRuleDisableOnScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows1 := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	rows1.AddRow(nil, rule1ID, 1)

	// expected query performed by tested function
	expectedQuery1 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows1)

	// org_id query is not expected, as the first query should fail

	mock.ExpectClose()

	// first query to be performed
	query1 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_rule_toggle
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`
	// call the tested function
	err = cleaner.PerformDisplayMultipleRuleDisable(connection, nil, query1, "cluster_rule_toggle")
	// must throw error
	assert.Error(t, err)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformDisplayMultipleRuleDisableResults checks the basic behaviour of
// performDisplayMultipleRuleDisable function with results returned. Contents
// of generated file(s) is checked in displayMultipleRuleDisableResulsts test cases
func TestPerformDisplayMultipleRuleDisableResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows1 := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	rows1.AddRow(cluster1ID, rule1ID, 1)

	// expected query performed by tested function
	expectedQuery1 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows1)

	// prepare mocked result for SQL query
	expectOrgIDQuery(mock)

	mock.ExpectClose()

	// first query to be performed
	query1 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_rule_toggle
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`
	// call the tested function
	err = cleaner.PerformDisplayMultipleRuleDisable(connection, nil, query1, "cluster_rule_toggle")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayMultipleRuleDisableResultsScanError checks the basic behaviour of
// displayMultipleRuleDisable function with results returned without defining the filenames.
func TestDisplayMultipleRuleDisableResultsScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow(nil, rule1ID, 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// another org_id query
	mock.ExpectClose()

	// call the tested function without filename (only printed in logs)
	err = cleaner.DisplayMultipleRuleDisable(connection, "")
	assert.Error(t, err)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayMultipleRuleDisableOnError checks the error handling
// ability in displayMultipleRuleDisable function.
func TestDisplayMultipleRuleDisableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnError(mockedError)

	// org_id query is not expected because first query should fail

	mock.ExpectClose()

	// call the tested function without filename (only printed in logs)
	err = cleaner.DisplayMultipleRuleDisable(connection, "")

	assert.Error(t, err)

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayMultipleRuleDisableResultsNoOutput checks the basic behaviour of
// displayMultipleRuleDisable function with results returned without defining the filenames.
func TestDisplayMultipleRuleDisableResultsNoOutput(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow(cluster1ID, rule1ID, 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	// prepare mocked result for SQL query
	feedbackRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	feedbackRows.AddRow(cluster2ID, rule1ID, 1)

	// expected query performed by tested function
	feedbackQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(feedbackQuery).WillReturnRows(feedbackRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	// another org_id query
	mock.ExpectClose()

	// call the tested function without filename (only printed in logs)
	err = cleaner.DisplayMultipleRuleDisable(connection, "")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayMultipleRuleDisableResultsFileOutput checks the basic behaviour of
// displayMultipleRuleDisable function with results returned and checks whether
// the files were generated correctly.
func TestDisplayMultipleRuleDisableResultsFileOutput(t *testing.T) {
	const outFile = "testdisable.out"

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow(cluster1ID, rule1ID, 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	// prepare mocked result for SQL query
	feedbackRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	feedbackRows.AddRow(cluster2ID, rule1ID, 1)

	// expected query performed by tested function
	feedbackQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(feedbackQuery).WillReturnRows(feedbackRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	// another org_id query
	mock.ExpectClose()

	// call the tested function with filename
	err = cleaner.DisplayMultipleRuleDisable(connection, outFile)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)

	// check contents of the output file
	outputFile, err := os.Open(outFile)
	assert.NoError(t, err)

	scanner := bufio.NewScanner(outputFile)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// two lines must be in the file
	assert.Len(t, lines, 2)

	// 4 comma separated values
	ruleToggleLine := strings.Split(lines[0], ",")
	assert.Len(t, ruleToggleLine, 4)

	// check elements in csv
	assert.Equal(t, ruleToggleLine[0], fmt.Sprint(defaultOrgID))
	assert.Equal(t, ruleToggleLine[1], cluster1ID)
	assert.Equal(t, ruleToggleLine[2], rule1ID)
	assert.Equal(t, ruleToggleLine[3], "1")

	ruleFeedbackLine := strings.Split(lines[1], ",")
	assert.Equal(t, ruleFeedbackLine[0], fmt.Sprint(defaultOrgID))
	assert.Equal(t, ruleFeedbackLine[1], cluster2ID)
	assert.Equal(t, ruleFeedbackLine[2], rule1ID)
	assert.Equal(t, ruleFeedbackLine[3], "1")

	err = outputFile.Close()
	assert.NoError(t, err)
	// delete test file from filesystem
	err = os.Remove(outFile)
	assert.NoError(t, err)
}

// TestDisplayMultipleRuleDisableResultsFileError checks the basic behaviour of
// displayMultipleRuleDisable function with results returned and an invalid filename
func TestDisplayMultipleRuleDisableResultsFileError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")
	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow(cluster1ID, rule1ID, 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	// prepare mocked result for SQL query
	feedbackRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	feedbackRows.AddRow(cluster1ID, rule1ID, 1)

	// expected query performed by tested function
	feedbackQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(feedbackQuery).WillReturnRows(feedbackRows)

	// prepare mocked org_id query result for SQL query
	expectOrgIDQuery(mock)

	mock.ExpectClose()

	// call the tested function with invalid filename
	err = cleaner.DisplayMultipleRuleDisable(connection, "/")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformListOfOldReportsNoResults checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldReportsNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformListOfOldReportsResults checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldReportsResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow(cluster1ID, reportedAt, updatedAt)

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformListOfOldScanError checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow(nil, reportedAt, updatedAt)

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	assert.Error(t, err)

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformListOfOldDBError checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldDBError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	assert.Error(t, err)

	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayAllOldRecordsNoOutput checks the basic behaviour of
// displayAllOldRecords function without a filename defined.
func TestDisplayAllOldRecordsNoOutput(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

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

	// call the tested function without filename (stdout)
	err = cleaner.DisplayAllOldRecords(connection, "10", "")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDisplayAllOldRecordsFileOutput checks the basic behaviour of
// displayAllOldRecords function without a filename defined.
func TestDisplayAllOldRecordsFileOutput(t *testing.T) {
	const outFile = "testold.out"

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow(cluster1ID, reportedAt, updatedAt)
	rows.AddRow(cluster2ID, reportedAt, updatedAt)

	// expected queries performed by tested function
	expectedQuery1 := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows)

	expectedQuery2 := "SELECT org_id, rule_fqdn, error_key, rule_id, rating, last_updated_at FROM advisor_ratings WHERE last_updated_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY last_updated_at"
	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows)

	expectedQuery3 := "SELECT topic, partition, topic_offset, key, consumed_at, message FROM consumer_error WHERE consumed_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY consumed_at"
	mock.ExpectQuery(expectedQuery3).WillReturnRows(rows)

	mock.ExpectClose()

	// call the tested function without filename (stdout)
	err = cleaner.DisplayAllOldRecords(connection, "10", outFile)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)

	// check contents of the output file
	outputFile, err := os.Open(outFile)
	assert.NoError(t, err)

	scanner := bufio.NewScanner(outputFile)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// two lines must be in the file
	assert.Len(t, lines, 2)

	// 4 comma separated values
	line1 := strings.Split(lines[0], ",")
	assert.Len(t, line1, 4)

	// check elements in csv
	assert.Equal(t, line1[0], cluster1ID)
	assert.Equal(t, line1[1], reportedAt.Format(time.RFC3339))
	assert.Equal(t, line1[2], updatedAt.Format(time.RFC3339))
	assert.Equal(t, line1[3], "1")

	line2 := strings.Split(lines[1], ",")
	assert.Equal(t, line2[0], cluster2ID)
	assert.Equal(t, line2[1], reportedAt.Format(time.RFC3339))
	assert.Equal(t, line2[2], updatedAt.Format(time.RFC3339))
	assert.Equal(t, line2[3], "1")

	err = outputFile.Close()
	assert.NoError(t, err)
	// delete test file from filesystem
	err = os.Remove(outFile)
	assert.NoError(t, err)
}

// TestDisplayAllOldRecordsWithFileError checks the basic behaviour of
// displayAllOldRecords function with file error
func TestDisplayAllOldRecordsWithFileError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

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

	// call the tested function with invalid filename ("/")
	err = cleaner.DisplayAllOldRecords(connection, "10", "/")
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformListOfOldReportsOnError checks the error handling
// ability in performListOfOldReports function.
func TestPerformListOfOldReportsOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	if err == nil {
		t.Fatalf("error was expected while updating stats")
	}

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDeleteRecordFromTable checks the basic behaviour of
// deleteRecordFromTable function.
func TestDeleteRecordFromTable(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedExec := "DELETE FROM table_x WHERE key_x = \\$"
	mock.ExpectExec(expectedExec).WithArgs("key_value").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectClose()

	// call the tested function
	affected, err := cleaner.DeleteRecordFromTable(connection, "table_x", "key_x", "key_value")
	assert.NoError(t, err, "error not expected while calling tested function")

	// test number of affected rows
	if affected != 1 {
		t.Errorf("wrong number of rows affected: %d", affected)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestDeleteRecordFromTableOnError checks the error handling in
// deleteRecordFromTable function.
func TestDeleteRecordFromTableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedExec := "DELETE FROM table_x WHERE key_x = \\$"
	mock.ExpectExec(expectedExec).WithArgs("key_value").WillReturnError(mockedError)
	mock.ExpectClose()

	// call the tested function
	affected, err := cleaner.DeleteRecordFromTable(connection, "table_x", "key_x", "key_value")
	if err == nil {
		t.Fatalf("error was expected while updating stats")
	}

	// test number of affected rows
	if affected != 0 {
		t.Errorf("wrong number of rows affected: %d", affected)
	}

	// check if the error is correct
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformVacuumDB checks the basic behaviour of
// PerformVacuumDB function.
func TestPerformVacuumDB(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	// expected query performed by tested function
	expectedExec := "DELETE FROM table_x WHERE key_x = \\$"
	mock.ExpectExec(expectedExec).WithArgs("key_value").WillReturnResult(sqlmock.NewResult(1, 1))

	expectedVacuum := "VACUUM VERBOSE;"
	mock.ExpectExec(expectedVacuum).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectClose()

	// call the tested function
	affected, err := cleaner.DeleteRecordFromTable(connection, "table_x", "key_x", "key_value")
	assert.NoError(t, err, "error not expected while calling tested function")

	// test number of affected rows
	if affected != 1 {
		t.Errorf("wrong number of rows affected: %d", affected)
	}

	err = cleaner.PerformVacuumDB(connection)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestFillInDatabaseByTestData checks the basic behaviour of
// fillInDatabaseByTestData function.
func TestFillInDatabaseByTestData(t *testing.T) {
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

	err = cleaner.FillInDatabaseByTestData(connection)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformCleanupInDB checks the basic behaviour of
// performCleanupInDB function.
func TestPerformCleanupInDB(t *testing.T) {
	expectedResult := make(map[string]int)

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	assert.NoError(t, err, "error creating SQL mock")

	clusterNames := cleaner.ClusterList{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-1111-1111-111111111111",
		"5d5892d4-1f74-4ccf-91af-548dfc9767aa",
	}

	for _, clusterName := range clusterNames {
		for _, tableAndKey := range cleaner.TablesAndKeys {
			// expected query performed by tested function
			expectedExec := fmt.Sprintf("DELETE FROM %v WHERE %v = \\$", tableAndKey.TableName, tableAndKey.KeyName)
			mock.ExpectExec(expectedExec).WithArgs(clusterName).WillReturnResult(sqlmock.NewResult(1, 2))

			// two deleted rows for each cluster
			expectedResult[tableAndKey.TableName] += 2
		}
	}

	mock.ExpectClose()

	deletedRows, err := cleaner.PerformCleanupInDB(connection, clusterNames)
	assert.NoError(t, err, "error not expected while calling tested function")

	// check tables have correct number of deleted rows for each table
	for tableName, deletedRowCount := range deletedRows {
		assert.Equal(t, expectedResult[tableName], deletedRowCount)
	}

	// check if DB can be closed successfully
	checkConnectionClose(t, connection)

	// check all DB expectactions happened correctly
	checkAllExpectations(t, mock)
}

// TestPerformCleanupInDBNoConnection checks the basic behaviour of
// performCleanupInDB function when connection is not established.
func TestPerformCleanupInDBNoConnection(t *testing.T) {
	// connection that is not constructed correctly
	var connection *sql.DB = nil

	clusterNames := cleaner.ClusterList{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-1111-1111-111111111111",
		"5d5892d4-1f74-4ccf-91af-548dfc9767aa",
	}

	_, err := cleaner.PerformCleanupInDB(connection, clusterNames)

	assert.Error(t, err, "error is expected while calling tested function")
}

// TestInitDatabaseNoConfiguration checks how initDatabaseConnection function
// behave if null configuration is used
func TestInitDatabaseNoConfiguration(t *testing.T) {
	// not initialized storage configuration
	var configurationPtr *cleaner.StorageConfiguration = nil

	// call tested function
	connection, err := cleaner.InitDatabaseConnection(configurationPtr)

	// check output from tested function
	assert.Error(t, err, "error is expected while calling tested function")
	assert.Nil(t, connection, "connection should not be established")
}

// TestInitDatabaseWrongDriver checks how initDatabaseConnection function
// behave if configuration with wrong driver is used
func TestInitDatabaseWrongDriver(t *testing.T) {
	// not initialized storage configuration
	configuration := cleaner.StorageConfiguration{
		Driver: "wrong-one",
	}

	// call tested function
	connection, err := cleaner.InitDatabaseConnection(&configuration)

	// check output from tested function
	assert.Error(t, err, "error is expected while calling tested function")
	assert.Nil(t, connection, "connection should not be established")
}
