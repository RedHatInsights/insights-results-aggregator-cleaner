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

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cleaner "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

func TestReadOrgIDNoResults(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer connection.Close()

	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	// call the tested function
	org_id, err := cleaner.ReadOrgID(connection, "123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// check the org ID returned from tested function
	if org_id != -1 {
		t.Errorf("wrong org_id returned: %d", org_id)
	}

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestReadOrgIDResult(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer connection.Close()

	rows := sqlmock.NewRows([]string{"org_id"})
	rows.AddRow("42")

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

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestReadOrgIDOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer connection.Close()

	// expected query performed by tested function
	expectedQuery := "select org_id from report where cluster = \\$1"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)

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
	defer connection.Close()

	rows := sqlmock.NewRows([]string{})

	// expected query performed by tested function
	expectedQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

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
	defer connection.Close()

	// expected query performed by tested function
	expectedQuery := "select cluster_id, rule_id, count\\(\\*\\) as cnt from cluster_rule_toggle group by cluster_id, rule_id having count\\(\\*\\)>1 order by cnt desc;"
	mock.ExpectQuery(expectedQuery).WillReturnError(mockedError)

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

	// check if all expectations were met
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
