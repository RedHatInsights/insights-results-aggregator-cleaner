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

package main_test

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/database_test.html

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	cleaner "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

// TestReadOrgIDNoResults checks the function readOrgID.
func TestReadOrgIDNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// check the org ID returned from tested function
	if org_id != -1 {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestReadOrgIDResults checks the function readOrgID.
func TestReadOrgIDResult(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"org_id"})
	rows.AddRow(42)

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// check the org ID returned from tested function
	if org_id != 42 {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestReadOrgIDOnError checks error handling in function readOrgID.
func TestReadOrgIDOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformDisplayMultipleRuleDisableNoResults checks the basic behaviour of
// performDisplayMultipleRuleDisable function.
func TestPerformDisplayMultipleRuleDisableNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformDisplayMultipleRuleDisableOnError checks the error handling
// ability in performDisplayMultipleRuleDisable function.
func TestPerformDisplayMultipleRuleDisableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformDisplayMultipleRuleDisableResults checks the basic behaviour of
// performDisplayMultipleRuleDisable function with results returned.
func TestPerformDisplayMultipleRuleDisableResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows1 := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	rows1.AddRow("123e4567-e89b-12d3-a456-426614173998", "rule.test|KEY", 1)

	// expected query performed by tested function
	expectedQuery1 := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery1).WillReturnRows(rows1)

	// prepare mocked result for SQL query
	rows2 := sqlmock.NewRows([]string{"org_id"})
	rows2.AddRow(42)

	// expected query performed by tested function
	expectedQuery2 := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows2)
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
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDisplayMultipleRuleDisableResultsStringOutput checks the basic behaviour of
// displayMultipleRuleDisable function with results returned, printing to std out.
func TestDisplayMultipleRuleDisableResultsStringOutput(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow("123e4567-e89b-12d3-a456-426614173998", "rule.test|KEY", 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// prepare mocked org_id query result for SQL query
	orgIDRows := sqlmock.NewRows([]string{"org_id"})
	orgIDRows.AddRow(1)

	// expected query performed by tested function
	orgIDQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(orgIDQuery).WillReturnRows(orgIDRows)

	// prepare mocked result for SQL query
	feedbackRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	feedbackRows.AddRow("123e4567-e89b-12d3-a456-426614173998", "rule.test|KEY", 1)

	// expected query performed by tested function
	feedbackQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(feedbackQuery).WillReturnRows(feedbackRows)

	// prepare mocked org_id query result for SQL query
	orgIDRows2 := sqlmock.NewRows([]string{"org_id"})
	orgIDRows2.AddRow(1)

	// expected query performed by tested function
	orgIDQuery2 := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(orgIDQuery2).WillReturnRows(orgIDRows2)
	// another org_id query
	mock.ExpectClose()

	// call the tested function without filename (stdout)
	err = cleaner.DisplayMultipleRuleDisable(connection, "")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDisplayMultipleRuleDisableResultsFileError checks the basic behaviour of
// displayMultipleRuleDisable function with results returned, printing to std out.
func TestDisplayMultipleRuleDisableResultsFileError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// prepare mocked result for SQL query
	toggleRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	toggleRows.AddRow("123e4567-e89b-12d3-a456-426614173998", "rule.test|KEY", 1)

	// expected query performed by tested function
	toggleQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(toggleQuery).WillReturnRows(toggleRows)

	// prepare mocked org_id query result for SQL query
	orgIDRows := sqlmock.NewRows([]string{"org_id"})
	orgIDRows.AddRow(1)

	// expected query performed by tested function
	orgIDQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(orgIDQuery).WillReturnRows(orgIDRows)

	// prepare mocked result for SQL query
	feedbackRows := sqlmock.NewRows([]string{"cluster_id", "rule_id", "cnt"})
	feedbackRows.AddRow("123e4567-e89b-12d3-a456-426614173998", "rule.test|KEY", 1)

	// expected query performed by tested function
	feedbackQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_user_rule_disable_feedback group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(feedbackQuery).WillReturnRows(feedbackRows)

	// prepare mocked org_id query result for SQL query
	orgIDRows2 := sqlmock.NewRows([]string{"org_id"})
	orgIDRows2.AddRow(1)

	// expected query performed by tested function
	orgIDQuery2 := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(orgIDQuery2).WillReturnRows(orgIDRows2)
	// another org_id query
	mock.ExpectClose()

	// call the tested function with invalid filename
	err = cleaner.DisplayMultipleRuleDisable(connection, "/")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformListOfOldReportsNoResults checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldReportsNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformListOfOldReportsResults checks the basic behaviour of
// performListOfOldReports function.
func TestPerformListOfOldReportsResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow("123e4567-e89b-12d3-a456-426614173998", reportedAt, updatedAt)

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function
	err = cleaner.PerformListOfOldReports(connection, "10", nil)
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDisplayAllOldRecordsWithWriter checks the basic behaviour of
// displayAllOldRecords function.
func TestDisplayAllOldRecordsWithWriter(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow("123e4567-e89b-12d3-a456-426614173998", reportedAt, updatedAt)

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function without filename (stdout)
	err = cleaner.DisplayAllOldRecords(connection, "10", "")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDisplayAllOldRecordsWithFileError checks the basic behaviour of
// displayAllOldRecords function with file error
func TestDisplayAllOldRecordsWithFileError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"cluster", "reported_at", "last_checked"})
	reportedAt := time.Now()
	updatedAt := time.Now()
	rows.AddRow("123e4567-e89b-12d3-a456-426614173998", reportedAt, updatedAt)

	// expected query performed by tested function
	expectedQuery := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW\\(\\) - \\$1::INTERVAL ORDER BY reported_at"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// call the tested function with invalid filename ("/")
	err = cleaner.DisplayAllOldRecords(connection, "10", "/")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformListOfOldReportsOnError checks the error handling
// ability in performListOfOldReports function.
func TestPerformListOfOldReportsOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDeleteRecordFromTable checks the basic behaviour of
// deleteRecordFromTable function.
func TestDeleteRecordFromTable(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// expected query performed by tested function
	expectedExec := "DELETE FROM table_x WHERE key_x = \\$"
	mock.ExpectExec(expectedExec).WithArgs("key_value").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectClose()

	// call the tested function
	affected, err := cleaner.DeleteRecordFromTable(connection, "table_x", "key_x", "key_value")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// test number of affected rows
	if affected != 1 {
		t.Errorf("wrong number of rows affected: %d", affected)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDeleteRecordFromTableOnError checks the error handling in
// deleteRecordFromTable function.
func TestDeleteRecordFromTableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformVacuumDB checks the basic behaviour of
// PerformVacuumDB function.
func TestPerformVacuumDB(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// expected query performed by tested function
	expectedExec := "DELETE FROM table_x WHERE key_x = \\$"
	mock.ExpectExec(expectedExec).WithArgs("key_value").WillReturnResult(sqlmock.NewResult(1, 1))

	expectedVacuum := "VACUUM VERBOSE;"
	mock.ExpectExec(expectedVacuum).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectClose()

	// call the tested function
	affected, err := cleaner.DeleteRecordFromTable(connection, "table_x", "key_x", "key_value")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// test number of affected rows
	if affected != 1 {
		t.Errorf("wrong number of rows affected: %d", affected)
	}

	err = cleaner.PerformVacuumDB(connection)
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestFillInDatabaseByTestData checks the basic behaviour of
// fillInDatabaseByTestData function.
func TestFillInDatabaseByTestData(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPerformCleanupInDB checks the basic behaviour of
// performCleanupInDB function.
func TestPerformCleanupInDB(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

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
		}
	}

	mock.ExpectClose()

	_, err = cleaner.PerformCleanupInDB(connection, clusterNames)

	err = connection.Close()
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
