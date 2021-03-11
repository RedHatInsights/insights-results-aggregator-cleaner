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

// This source file contains an implementation of interface between Go code and
// (almost any) SQL database like PostgreSQL, SQLite, or MariaDB.
//
// It is possible to configure connection to selected database by using
// StorageConfiguration structure. Currently that structure contains two
// configurable parameter:
//
// Driver - a SQL driver, like "sqlite3", "pq" etc.
// DataSource - specification of data source. The content of this parameter depends on the database used.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-cleaner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"time"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"
)

// DBDriver type for db driver enum
type DBDriver int

const (
	// DBDriverSQLite3 shows that db driver is sqlite
	DBDriverSQLite3 DBDriver = iota
	// DBDriverPostgres shows that db driver is postgres
	DBDriverPostgres
	// DBDriverGeneral general sql(used for mock now)
	DBDriverGeneral
)

// Error messages
const canNotConnectToDataStorageMessage = "Can not connect to data storage"

// initDatabaseConnection initializes driver, checks if it's supported and
// initializes connection to the storage.
func initDatabaseConnection(configuration StorageConfiguration) (*sql.DB, error) {
	driverName := configuration.Driver
	dataSource := ""
	log.Info().Str("driverName", configuration.Driver).Msg("DB connection configuration")

	// initialize connection into selected database using the right driver
	switch driverName {
	case "sqlite3":
		//driverType := DBDriverSQLite3
		//driver = &sqlite3.SQLiteDriver{}
		dataSource = configuration.SQLiteDataSource
	case "postgres":
		//driverType := DBDriverPostgres
		//driver = &pq.Driver{}
		dataSource = fmt.Sprintf(
			"postgresql://%v:%v@%v:%v/%v?%v",
			configuration.PGUsername,
			configuration.PGPassword,
			configuration.PGHost,
			configuration.PGPort,
			configuration.PGDBName,
			configuration.PGParams,
		)
	default:
		err := fmt.Errorf("driver %v is not supported", driverName)
		log.Err(err).Msg(canNotConnectToDataStorageMessage)
		return nil, err
	}

	// try to initialize connection to the storage
	connection, err := sql.Open(driverName, dataSource)

	// check if establishing a connection was successful
	if err != nil {
		log.Err(err).Msg(canNotConnectToDataStorageMessage)
		return nil, err
	}

	return connection, nil
}

// displayAllOldRecords functions read all old records, ie. records that are
// older than the specified time duration. Those records are simply displayed.
func displayAllOldRecords(connection *sql.DB, maxAge string, output string) error {
	var fout *os.File = nil
	var writer *bufio.Writer = nil

	if output != "" {
		fout, err := os.Create(output)
		if err != nil {
			log.Error().Err(err).Msg("File open")
		}
		writer = bufio.NewWriter(fout)

	}

	defer func() {
		if writer != nil {
			err := writer.Flush()
			if err != nil {
				log.Error().Err(err).Msg("Flush writer")
			}
		}
	}()

	defer func() {
		if fout != nil {
			err := fout.Close()
			if err != nil {
				log.Error().Err(err).Msg("File close")
			}
		}
	}()

	return performListOfOldReports(connection, maxAge, writer)
}

func performListOfOldReports(connection *sql.DB, maxAge string, writer *bufio.Writer) error {
	query := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW() - $1::INTERVAL ORDER BY reported_at"
	rows, err := connection.Query(query, maxAge)
	if err != nil {
		return err
	}

	// used to compute a real record age
	now := time.Now()

	// iterate over all old records
	for rows.Next() {
		var (
			clusterName string
			reported    time.Time
			lastChecked time.Time
		)

		// read one old record from the report table
		if err := rows.Scan(&clusterName, &reported, &lastChecked); err != nil {
			// close the result set in case of any error
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Unable to close the DB rows handle")
			}
			return err
		}

		// compute the real record age
		age := int(math.Ceil(now.Sub(reported).Hours() / 24)) // in days

		// prepare for the report
		reportedF := reported.Format(time.RFC3339)
		lastCheckedF := lastChecked.Format(time.RFC3339)

		// just print the report
		log.Info().Str("cluster", clusterName).
			Str("reported", reportedF).
			Str("lastChecked", lastCheckedF).
			Int("age", age).
			Msg("Old report")

		if writer != nil {
			_, err := fmt.Fprintf(writer, "%s,%s,%s,%d\n", clusterName, reportedF, lastCheckedF, age)
			if err != nil {
				log.Error().Err(err).Msg("write to file")
			}
		}
	}
	return nil
}

// deleteRecordFromTable function deletes selected records (identified by
// cluster name) from database
func deleteRecordFromTable(connection *sql.DB, table string, key string, clusterName ClusterName) (int, error) {
	// it is not possible to use parameter for table name or a key
	// disable "G202 (CWE-89): SQL string concatenation (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G202
	sqlStatement := "DELETE FROM " + table + " WHERE " + key + " = $1;"
	// println(sqlStatement)

	// perform the SQL statement
	result, err := connection.Exec(sqlStatement, clusterName)
	if err != nil {
		return 0, err
	}

	// read number of affected (deleted) rows
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// tablesAndKeys contains list of all tables together with keys used to select
// records to be deleted
var tablesAndKeys = [...]TableAndKey{
	{
		TableName: "cluster_rule_toggle",
		KeyName:   "cluster_id",
	},
	{
		TableName: "cluster_rule_user_feedback",
		KeyName:   "cluster_id",
	},
	{
		TableName: "cluster_user_rule_disable_feedback",
		KeyName:   "cluster_id",
	},
	{
		TableName: "rule_hit",
		KeyName:   "cluster_id",
	},
	// must be at the end due to constraints
	{
		TableName: "report",
		KeyName:   "cluster",
	},
}

// performCleanupInDB function cleans up all data for selected cluster names
func performCleanupInDB(connection *sql.DB,
	clusterList ClusterList) (map[string]int, error) {

	// initialize counters
	deletionsForTable := make(map[string]int)
	for _, tableAndKey := range tablesAndKeys {
		deletionsForTable[tableAndKey.TableName] = 0
	}

	// perform cleanup for selected cluster names
	log.Info().Msg("Cleanup started")
	for _, clusterName := range clusterList {
		for _, tableAndKey := range tablesAndKeys {
			// try to delete record from selected table
			affected, err := deleteRecordFromTable(connection,
				tableAndKey.TableName,
				tableAndKey.KeyName,
				clusterName)
			if err != nil {
				log.Error().
					Err(err).
					Str("table", tableAndKey.TableName).
					Msg("Unable to delete record")
			} else {
				log.Info().
					Int("Affected", affected).
					Str("table", tableAndKey.TableName).
					Str("cluster", string(clusterName)).
					Msg("Delete record")
				deletionsForTable[tableAndKey.TableName] += affected
			}
		}
	}
	log.Info().Msg("Cleanup finished")
	return deletionsForTable, nil
}

// fillInDatabaseByTestData function fill-in database by test data (not to be
// used against production database)
func fillInDatabaseByTestData(connection *sql.DB) error {
	log.Info().Msg("Fill-in database started")
	var lastError error = nil

	clusterNames := [...]string{
		"00000000-0000-0000-0000-000000000000",
		"11111111-1111-1111-1111-111111111111",
		"5d5892d4-1f74-4ccf-91af-548dfc9767aa"}

	sqlStatements := [...]string{
		"INSERT INTO report (org_id, cluster, report, reported_at, last_checked_at, kafka_offset) values(1, $1, '', '2021-01-01', '2021-01-01', 10)",
		"INSERT INTO cluster_rule_toggle (cluster_id, rule_id, user_id, disabled, disabled_at, enabled_at, updated_at) values($1, 1, 1, 0, '2021-01-01', '2021-01-01', '2021-01-01')",
		"INSERT INTO cluster_rule_user_feedback (cluster_id, rule_id, user_id, message, user_vote, added_at, updated_at) values($1, 1, 1, 'foobar', 1, '2021-01-01', '2021-01-01')",
		"INSERT INTO cluster_user_rule_disable_feedback (cluster_id, user_id, rule_id, message, added_at, updated_at) values($1, 1, 1, 'foobar', '2021-01-01', '2021-01-01')",
		"INSERT INTO rule_hit (org_id, cluster_id, rule_fqdn, error_key, template_data) values(1, $1, 'foo', 'bar', '')",
	}

	for _, clusterName := range clusterNames {
		log.Info().
			Str("cluster name", clusterName).
			Msg("data for new cluster")

		for _, sqlStatement := range sqlStatements {
			log.Info().
				Str("SQL statement", sqlStatement).
				Msg("inserting")
			// perform the SQL statement
			_, err := connection.Exec(sqlStatement, clusterName)
			if err != nil {
				// failure is usually ok - it might mean that
				// the record with given cluster name already
				// exists
				log.Err(err).Msg("Insert error")
				lastError = err
			}
		}

	}
	log.Info().Msg("Fill-in database finished")
	return lastError
}
