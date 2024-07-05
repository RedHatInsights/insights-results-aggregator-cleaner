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

package main

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/database.html

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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"
)

// Error messages
const (
	canNotConnectToDataStorageMessage = "Can not connect to data storage"
	unableToCloseDBRowsHandle         = "Unable to close the DB rows handle"
	connectionNotEstablished          = "Connection to database was not established"
	reportedMsg                       = "reported"
	lastCheckedMsg                    = "lastChecked"
	ageMsg                            = "age"
	reportsCountMsg                   = "reports count"
	maxAgeMissing                     = "max-age parameter is missing"
	invalidSchemaMsg                  = "Invalid DB schema to be cleaned up: '%s'"
	affectedMsg                       = "Affected"
)

// Other messages
const (
	tableName      = "table"
	clusterNameMsg = "cluster"
	fileOpenMsg    = "File open"
	fileCloseMsg   = "File close"
	flushWriterMsg = "Flush writer"
	writeToFileMsg = "Write to file"
)

// SQL commands
const (
	selectOldOCPReports = `
	    SELECT cluster, reported_at, last_checked_at
	      FROM report
	     WHERE reported_at < NOW() - $1::INTERVAL
	     ORDER BY reported_at`

	selectOldAdvisorRatings = `
	    SELECT org_id, rule_fqdn, error_key, rule_id, rating, last_updated_at
	      FROM advisor_ratings
	     WHERE last_updated_at < NOW() - $1::INTERVAL
	     ORDER BY last_updated_at`

	selectOldConsumerErrors = `
	    SELECT topic, partition, topic_offset, key, consumed_at, message
	      FROM consumer_error
	     WHERE consumed_at < NOW() - $1::INTERVAL
	     ORDER BY consumed_at`

	selectOldDVOReports = `
	    SELECT org_id, cluster_id, reported_at, last_checked_at
	      FROM dvo.dvo_report
	     WHERE reported_at < NOW() - $1::INTERVAL
	     ORDER BY reported_at`

	deleteOldOCPReports = `
		DELETE FROM report
		 WHERE reported_at < NOW() - $1::INTERVAL`

	deleteOldConsumerErrors = `
		DELETE FROM consumer_error
		 WHERE consumed_at < NOW() - $1::INTERVAL`

	deleteOldOCPRuleHits = `
		DELETE FROM rule_hit
		 WHERE (cluster_id, org_id) IN (
			SELECT cluster, org_id
			FROM report
			WHERE reported_at < NOW() - $1::INTERVAL)
		 OR (cluster_id, org_id) NOT IN (
		    SELECT cluster, org_id
			FROM report
		 )`

	deleteOldOCPRecommendation = `
		DELETE FROM recommendation
		 WHERE created_at < NOW() - $1::INTERVAL`

	deleteOldDVOReports = `
		DELETE FROM dvo.dvo_report
		 WHERE reported_at < NOW() - $1::INTERVAL`
)

// DB schemas
const (
	DBSchemaOCPRecommendations = "ocp_recommendations"
	DBSchemaDVORecommendations = "dvo_recommendations"
)

var emptyJSON = json.RawMessage(`{}`)

// initDatabaseConnection initializes driver, checks if it's supported and
// initializes connection to the storage.
func initDatabaseConnection(configuration *StorageConfiguration) (*sql.DB, error) {
	// check if storage configuration structure has been initialized and
	// passed to this function properly
	if configuration == nil {
		const message = "StorageConfiguration structure should be provided"
		err := errors.New(message)
		log.Error().Msg(message)
		return nil, err
	}

	driverName := configuration.Driver
	dataSource := ""
	log.Info().Str("driverName", configuration.Driver).Msg("DB connection configuration")

	// initialize connection into selected database using the right driver
	switch driverName {
	case "sqlite3":
		dataSource = configuration.SQLiteDataSource
	case "postgres":
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

// displayMultipleRuleDisable function read and displays clusters where
// multiple users have disabled some rules.
func displayMultipleRuleDisable(connection *sql.DB, output string) error {
	var fout *os.File
	var writer *bufio.Writer

	if output != "" {
		// create output file
		// disable G304 (CWE-22): Potential file inclusion via variable (Confidence: HIGH, Severity: MEDIUM)
		fout, err := os.Create(output) // #nosec G304
		if err != nil {
			log.Error().Err(err).Msg(fileOpenMsg)
		}
		// an object used to write to file
		writer = bufio.NewWriter(fout)
	}

	defer func() {
		// output needs to be flushed at the end
		if writer != nil {
			err := writer.Flush()
			if err != nil {
				log.Error().Err(err).Msg(flushWriterMsg)
			}
		}
	}()

	defer func() {
		// file needs to be closed at the end
		if fout != nil {
			err := fout.Close()
			if err != nil {
				log.Error().Err(err).Msg(fileCloseMsg)
			}
		}
	}()

	// first query to be performed
	query1 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_rule_toggle
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`
	// second query to be performed
	query2 := `
                select cluster_id, rule_id, count(*) as cnt
                  from cluster_user_rule_disable_feedback
                 group by cluster_id, rule_id
                having count(*)>1
                 order by cnt desc;
`

	// perform the first query and display results
	err := performDisplayMultipleRuleDisable(connection, writer, query1,
		"cluster_rule_toggle")
	// the first query+display function might throw some error
	if err != nil {
		return err
	}

	// perform second query and display results
	err = performDisplayMultipleRuleDisable(connection, writer, query2,
		"cluster_user_rule_disable_feedback")
	// second query+display function might throw some error
	return err
}

// performDisplayMultipleRuleDisable function displays cluster names and org
// ids where multiple users disabled any rule
func performDisplayMultipleRuleDisable(connection *sql.DB,
	writer *bufio.Writer, query string, tableName string) error {
	// perform given query to database
	rows, err := connection.Query(query)
	if err != nil {
		return err
	}

	// iterate over all records that has been found
	for rows.Next() {
		var (
			clusterName string
			ruleID      string
			count       int
		)

		// read one report
		if err := rows.Scan(&clusterName, &ruleID, &count); err != nil {
			// close the result set in case of any error
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
			}
			return err
		}

		// try to read organization ID for given cluster name
		orgID, err := readOrgID(connection, clusterName)
		if err != nil {
			log.Error().Err(err).Msg("readOrgID")
			return err
		}

		// just print the report, including organization ID
		log.Info().Str("table", tableName).
			Int("org ID", orgID).
			Str(clusterNameMsg, clusterName).
			Str("rule ID", ruleID).
			Int("count", count).
			Msg("Multiple rule disable")

		// export to file (if enabled)
		if writer != nil {
			_, err := fmt.Fprintf(writer, "%d,%s,%s,%d\n", orgID, clusterName, ruleID, count)
			if err != nil {
				log.Error().Err(err).Msg(writeToFileMsg)
			}
		}
	}
	return nil
}

// readOrgID function tries to read organization ID for given cluster name
func readOrgID(connection *sql.DB, clusterName string) (int, error) {
	query := "select org_id from report where cluster = $1"

	// perform the query
	rows, err := connection.Query(query, clusterName)
	if err != nil {
		log.Debug().Msg("query")
		return -1, err
	}

	// and check the result (if any)
	if rows.Next() {
		var orgID int

		// read one organization ID returned in query result
		if err := rows.Scan(&orgID); err != nil {
			// proper error logging will be performed elsewhere
			log.Debug().Str(clusterNameMsg, clusterName).Msg("scan")

			// close the result set in case of any error
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
			}
			return -1, err
		}

		return orgID, nil
	}

	// no result?
	log.Debug().Str(clusterNameMsg, clusterName).Msg("no org_id for cluster")
	return -1, nil
}

func createOutputFile(output string) (*os.File, *bufio.Writer) {
	var fout *os.File
	var writer *bufio.Writer

	if output != "" {
		// create output file
		// disable G304 (CWE-22): Potential file inclusion via variable (Confidence: HIGH, Severity: MEDIUM)
		fout, err := os.Create(output) // #nosec G304
		if err != nil {
			log.Error().Err(err).Msg(fileOpenMsg)
		}
		// an object used to write to file
		writer = bufio.NewWriter(fout)
	}
	return fout, writer
}

// displayAllOldRecords function read all old records, ie. records that are
// older than the specified time duration. Those records are simply displayed.
func displayAllOldRecords(connection *sql.DB, maxAge, output string, schema string) error {
	// check if connection has been initialized
	if connection == nil {
		log.Error().Msg(connectionNotEstablished)
		return errors.New(connectionNotEstablished)
	}

	fout, writer := createOutputFile(output)

	defer func() {
		// output needs to be flushed at the end
		if writer != nil {
			err := writer.Flush()
			if err != nil {
				log.Error().Err(err).Msg(flushWriterMsg)
			}
		}
	}()

	defer func() {
		// file needs to be closed at the end
		if fout != nil {
			err := fout.Close()
			if err != nil {
				log.Error().Err(err).Msg(fileCloseMsg)
			}
		}
	}()

	switch schema {
	case DBSchemaOCPRecommendations:
		// main function of this tool is ability to delete old reports
		err := performListOfOldOCPReports(connection, maxAge, writer)
		// skip next operation on first error
		if err != nil {
			return err
		}

		// but we might be interested in other tables as well, especially advisor ratings
		err = performListOfOldRatings(connection, maxAge)
		// skip next operation on first error
		if err != nil {
			return err
		}

		// also but we might be interested in other consumer errors
		err = performListOfOldConsumerErrors(connection, maxAge)
		// skip next operation on first error
		if err != nil {
			return err
		}
	case DBSchemaDVORecommendations:
		// main function of this tool is ability to delete old reports
		err := performListOfOldDVOReports(connection, maxAge, writer)
		// skip next operation on first error
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid database schema to be investigated: '%s'", schema)
	}

	return nil
}

func listOldDatabaseRecords(connection *sql.DB, maxAge string,
	writer *bufio.Writer, query string,
	logEntry string, countLogEntry string,
	callback func(rows *sql.Rows, writer *bufio.Writer) (int, error)) error {
	log.Info().Msg(logEntry + " begin")
	rows, err := connection.Query(query, maxAge)
	if err != nil {
		return err
	}

	count, err := callback(rows, writer)
	if err != nil {
		log.Error().Err(err).Msg("Query error")
		return err
	}

	log.Info().Int(countLogEntry, count).Msg(logEntry + " end")
	return nil
}

// performListOfOldOCPReports read and displays old records read from reported_at
// table
func performListOfOldOCPReports(connection *sql.DB, maxAge string, writer *bufio.Writer) error {
	return listOldDatabaseRecords(connection, maxAge, writer, selectOldOCPReports, "List of old OCP reports", reportsCountMsg,
		func(rows *sql.Rows, writer *bufio.Writer) (int, error) {
			// used to compute a real record age
			now := time.Now()

			// reports count
			count := 0

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
						log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
					}
					return count, err
				}

				// compute the real record age
				age := int(math.Ceil(now.Sub(reported).Hours() / 24)) // in days

				// prepare for the report
				reportedF := reported.Format(time.RFC3339)
				lastCheckedF := lastChecked.Format(time.RFC3339)

				// just print the report
				log.Info().Str(clusterNameMsg, clusterName).
					Str(reportedMsg, reportedF).
					Str(lastCheckedMsg, lastCheckedF).
					Int(ageMsg, age).
					Msg("Old OCP report")

				if writer != nil {
					_, err := fmt.Fprintf(writer, "%s,%s,%s,%d\n", clusterName, reportedF, lastCheckedF, age)
					if err != nil {
						log.Error().Err(err).Msg(writeToFileMsg)
					}
				}
				count++
			}
			return count, nil
		})
}

// performListOfOldDVOReports read and displays old records read from dvo.dvo_report
// table
func performListOfOldDVOReports(connection *sql.DB, maxAge string, writer *bufio.Writer) error {
	return listOldDatabaseRecords(connection, maxAge, writer, selectOldDVOReports, "List of old DVO reports", reportsCountMsg,
		func(rows *sql.Rows, writer *bufio.Writer) (int, error) {
			// used to compute a real record age
			now := time.Now()

			// reports count
			count := 0

			// iterate over all old records
			for rows.Next() {
				var (
					orgID       int
					clusterName string
					reported    time.Time
					lastChecked time.Time
				)

				// read one old record from the report table
				if err := rows.Scan(&orgID, &clusterName, &reported, &lastChecked); err != nil {
					// close the result set in case of any error
					if closeErr := rows.Close(); closeErr != nil {
						log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
					}
					return count, err
				}

				// compute the real record age
				age := int(math.Ceil(now.Sub(reported).Hours() / 24)) // in days

				// prepare for the report
				reportedF := reported.Format(time.RFC3339)
				lastCheckedF := lastChecked.Format(time.RFC3339)

				// just print the report
				log.Info().Str(clusterNameMsg, clusterName).
					Str(reportedMsg, reportedF).
					Str(lastCheckedMsg, lastCheckedF).
					Int(ageMsg, age).
					Msg("Old DVO report")

				if writer != nil {
					_, err := fmt.Fprintf(writer, "%d,%s,%s,%s,%d\n", orgID, clusterName, reportedF, lastCheckedF, age)
					if err != nil {
						log.Error().Err(err).Msg(writeToFileMsg)
					}
				}
				count++
			}
			return count, nil
		})
}

// performListOfOldRatings read and displays old Advisor ratings read from
// advisor_ratings table
func performListOfOldRatings(connection *sql.DB, maxAge string) error {
	return listOldDatabaseRecords(connection, maxAge, nil, selectOldAdvisorRatings, "List of old Advisor ratings", "ratings count",
		func(rows *sql.Rows, _ *bufio.Writer) (int, error) {
			// used to compute a real record age
			now := time.Now()

			// reports count
			count := 0

			// iterate over all old records
			for rows.Next() {
				var (
					orgID         string
					ruleFQDN      string
					errorKey      string
					ruleID        string
					rating        int
					lastUpdatedAt time.Time
				)

				// read one old record from the report table
				if err := rows.Scan(&orgID, &ruleFQDN, &errorKey, &ruleID, &rating, &lastUpdatedAt); err != nil {
					// close the result set in case of any error
					if closeErr := rows.Close(); closeErr != nil {
						log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
					}
					return count, err
				}

				// compute the real error age
				age := int(math.Ceil(now.Sub(lastUpdatedAt).Hours() / 24)) // in days

				// prepare for the report
				lastUpdatedAtF := lastUpdatedAt.Format(time.RFC3339)

				// just print the report
				log.Info().
					Str("organization", orgID).
					Str("rule FQDN", ruleFQDN).
					Str("error key", errorKey).
					Int("rating", rating).
					Str("updated at", lastUpdatedAtF).
					Int("rating age", age).
					Msg("Old Advisor rating")
				count++
			}
			return count, nil
		})
}

// performListOfOldConsumerErrors read and displays consumer errors stored in
// consumer_errors table
func performListOfOldConsumerErrors(connection *sql.DB, maxAge string) error {
	return listOldDatabaseRecords(connection, maxAge, nil, selectOldConsumerErrors, "List of old consumer errors", "errors count",
		func(rows *sql.Rows, _ *bufio.Writer) (int, error) {
			// used to compute a real record age
			now := time.Now()

			// reports count
			count := 0

			// iterate over all old records
			for rows.Next() {
				var (
					topic      string
					partition  int
					offset     int
					key        string
					consumedAt time.Time
					message    string
				)

				// read one old record from the report table
				if err := rows.Scan(&topic, &partition, &offset, &key, &consumedAt, &message); err != nil {
					// close the result set in case of any error
					if closeErr := rows.Close(); closeErr != nil {
						log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
					}
					return count, err
				}

				// compute the real error age
				age := int(math.Ceil(now.Sub(consumedAt).Hours() / 24)) // in days

				// prepare for the report
				consumedF := consumedAt.Format(time.RFC3339)

				// just print the report
				log.Info().
					Str("topic", topic).
					Int("partition", partition).
					Int("offset", offset).
					Str("key", key).
					Str("message", message).
					Str("consumed", consumedF).
					Int("error age", age).
					Msg("Old consumer error")
				count++
			}
			return count, nil
		})
}

// deleteRecordFromTable function deletes selected records (identified by
// cluster name) from database
func deleteRecordFromTable(connection *sql.DB, table, key string, clusterName ClusterName) (int, error) {
	// it is not possible to use parameter for table name or a key
	// disable "G202 (CWE-89): SQL string concatenation (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G202
	sqlStatement := "DELETE FROM " + table + " WHERE " + key + " = $1;"

	// perform the SQL statement
	// #nosec G202
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

var (
	tablesToDeleteOCP = []TableAndDeleteStatement{
		{
			TableName:       "rule_hit",
			DeleteStatement: deleteOldOCPRuleHits,
		},
		{
			TableName:       "report",
			DeleteStatement: deleteOldOCPReports,
		},
		{
			TableName:       "consumer_error",
			DeleteStatement: deleteOldConsumerErrors,
		},
		{
			TableName:       "recommendation",
			DeleteStatement: deleteOldOCPRecommendation,
		},
	}

	tablesToDeleteDVO = []TableAndDeleteStatement{
		{
			TableName:       "dvo.dvo_report",
			DeleteStatement: deleteOldDVOReports,
		},
	}
	allTablesToDelete = append(tablesToDeleteOCP, tablesToDeleteDVO...)
)

// deleteOldRecordsFromTable function deletes old records from database
// each delete query must have just one parameter that will be populated with
// the maxAge value
func deleteOldRecordsFromTable(connection *sql.DB, sqlStatement, maxAge string, dryRun bool) (int, error) {
	if dryRun {
		sqlStatement = strings.Replace(sqlStatement, "DELETE", "SELECT", -1)
	}
	result, err := connection.Exec(sqlStatement, maxAge)
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

// tablesAndKeysInOCPDatabase contains list of all tables together with keys used to select
// records to be deleted
var tablesAndKeysInOCPDatabase = []TableAndKey{
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
	{
		TableName: "recommendation",
		KeyName:   "cluster_id",
	},
	{
		TableName: "report_info",
		KeyName:   "cluster_id",
	},
	// must be at the end due to constraints
	{
		TableName: "report",
		KeyName:   "cluster",
	},
}

var tablesAndKeysInDVODatabase = []TableAndKey{
	{
		TableName: "dvo_report",
		KeyName:   "cluster_id",
	},
}

// performVacuumDB vacuums the whole database
func performVacuumDB(connection *sql.DB) error {
	log.Info().Msg("Vacuuming started")
	sqlStatement := "VACUUM VERBOSE;"

	// perform the SQL statement
	_, err := connection.Exec(sqlStatement)
	if err != nil {
		return err
	}
	log.Info().Msg("Vacuuming finished")
	return nil
}

// performCleanupInDB function cleans up all data for selected cluster names
func performCleanupInDB(connection *sql.DB,
	clusterList ClusterList, schema string) (map[string]int, error) {
	// return value
	deletionsForTable := make(map[string]int)

	// check if connection has been initialized
	if connection == nil {
		log.Error().Msg(connectionNotEstablished)
		return deletionsForTable, errors.New(connectionNotEstablished)
	}

	// this is actually shorter than using map + map selector + test for key existence
	// and it allow us to do fine tuning for (any) DB schema in future
	var tablesAndKeys []TableAndKey
	switch schema {
	case DBSchemaOCPRecommendations:
		tablesAndKeys = tablesAndKeysInOCPDatabase
	case DBSchemaDVORecommendations:
		tablesAndKeys = tablesAndKeysInDVODatabase
	default:
		return deletionsForTable, fmt.Errorf(invalidSchemaMsg, schema)
	}

	// initialize counters
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
					Str(tableName, tableAndKey.TableName).
					Msg("Unable to delete record")
			} else {
				log.Info().
					Int(affectedMsg, affected).
					Str(tableName, tableAndKey.TableName).
					Str(clusterNameMsg, string(clusterName)).
					Msg("Delete record")
				deletionsForTable[tableAndKey.TableName] += affected
			}
		}
	}
	log.Info().Msg("Cleanup finished")
	return deletionsForTable, nil
}

// performCleanupAllInDB function cleans up all data for all cluster names
func performCleanupAllInDB(connection *sql.DB, maxAge string, dryRun bool) (
	map[string]int, error) {
	deletionsForTable := make(map[string]int)
	if maxAge == "" {
		return deletionsForTable, errors.New(maxAgeMissing)
	}
	log.Debug().Str("Max age", maxAge).Msg("Cleaning all old records from DB")

	if connection == nil {
		log.Error().Msg(connectionNotEstablished)
		return deletionsForTable, errors.New(connectionNotEstablished)
	}

	// perform cleanup for selected cluster names
	log.Info().Msg("Cleanup-all started")
	for _, tableAndDeleteStatement := range allTablesToDelete {
		// try to delete record from selected table
		affected, err := deleteOldRecordsFromTable(connection,
			tableAndDeleteStatement.DeleteStatement,
			maxAge, dryRun)
		if err != nil {
			log.Error().
				Err(err).
				Str(tableName, tableAndDeleteStatement.TableName).
				Msg("Unable to delete records")
			return deletionsForTable, err
		}
		log.Info().
			Int(affectedMsg, affected).
			Str(tableName, tableAndDeleteStatement.TableName).
			Bool("Dry run", dryRun).
			Msg("Delete records")
		deletionsForTable[tableAndDeleteStatement.TableName] = affected
	}
	log.Info().Msg("Cleanup-all finished")
	return deletionsForTable, nil
}

// fillInDatabaseByTestData function fill-in database by test data (not to be
// used against production database)
func fillInDatabaseByTestData(connection *sql.DB, schema string) error {
	log.Info().Msg("Fill-in database started")

	switch schema {
	case DBSchemaOCPRecommendations:
		return fillInOCPDatabaseByTestData(connection)
	case DBSchemaDVORecommendations:
		return fillInDVODatabaseByTestData(connection)
	default:
		return fmt.Errorf("Invalid DB schema '%s'", schema)
	}
}

// fillInOCPDatabaseByTestData function fills-in OCP database by test data
// (not to be used against production database)
func fillInOCPDatabaseByTestData(connection *sql.DB) error {
	var lastError error

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
				Msg("inserting into OCP database")
			// perform the SQL statement
			_, err := connection.Exec(sqlStatement, clusterName)
			if err != nil {
				// failure is usually ok - it might mean that
				// the record with given cluster name already
				// exists
				log.Err(err).Msg("Insert error (OCP)")
				lastError = err
			}
		}
	}
	log.Info().Msg("Fill-in OCP database finished")
	return lastError
}

// fillInDVODatabaseByTestData function fills-in DVO database by test data
// (not to be used against production database)
func fillInDVODatabaseByTestData(connection *sql.DB) error {
	/* Table that needs to be filled-in has the following schema:
	    CREATE TABLE dvo.dvo_report (
	    org_id          INTEGER NOT NULL,
	    cluster_id      VARCHAR NOT NULL,
	    namespace_id    VARCHAR NOT NULL,
	    namespace_name  VARCHAR,
	    report          TEXT,
	    recommendations INTEGER NOT NULL,
	    objects         INTEGER NOT NULL,
	    reported_at     TIMESTAMP,
	    last_checked_at TIMESTAMP,
		rule_hits_count JSONB
	    PRIMARY KEY(org_id, cluster_id, namespace_id)
	)
	*/

	const insertStatement = `
	    INSERT INTO dvo.dvo_report
	           (org_id, cluster_id, namespace_id, namespace_name, report, recommendations, objects, reported_at, last_checked_at, rule_hits_count)
		   values
		   ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

	type Record struct {
		OrgID           int
		ClusterID       string
		NamespaceID     string
		NamespaceName   string
		Report          string
		Recommendations int
		Objects         int
		ReportedAt      string
		LastCheckedAt   string
		RuleHitsCount   json.RawMessage
	}

	const cluster1 = "00000001-0001-0001-0001-000000000001"
	const cluster2 = "00000002-0002-0002-0002-000000000002"
	const cluster3 = "00000003-0003-0003-0003-000000000003"

	records := []Record{
		{
			OrgID:           1,
			ClusterID:       cluster1,
			NamespaceID:     "fbcbe2d3-e398-4b40-9d5e-4eb46fe8286f",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 1,
			Objects:         6,
			ReportedAt:      "2021-01-01",
			LastCheckedAt:   "2021-01-01",
			RuleHitsCount:   emptyJSON,
		},
		{
			OrgID:           1,
			ClusterID:       cluster2,
			NamespaceID:     "e6ed9bb3-efc3-46a6-b3ae-3f1a6e59546c",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 2,
			Objects:         5,
			ReportedAt:      "2021-01-01",
			LastCheckedAt:   "2021-01-01",
			RuleHitsCount:   emptyJSON,
		},
		{
			OrgID:           2,
			ClusterID:       cluster3,
			NamespaceID:     "e6ed9bb3-efc3-46a6-b3ae-3f1a6e59546c",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 3,
			Objects:         4,
			ReportedAt:      "2021-01-01",
			LastCheckedAt:   "2021-01-01",
			RuleHitsCount:   emptyJSON,
		},
		{
			OrgID:           3,
			ClusterID:       cluster1,
			NamespaceID:     "e6ed9bb3-efc3-46a6-b3ae-3f1a6e59546c",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 4,
			Objects:         3,
			ReportedAt:      "2021-01-01",
			LastCheckedAt:   "2021-01-01",
			RuleHitsCount:   emptyJSON,
		},
		{
			OrgID:           3,
			ClusterID:       cluster2,
			NamespaceID:     "e6ed9bb3-efc3-46a6-b3ae-3f1a6e59546c",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 5,
			Objects:         2,
			ReportedAt:      "2022-01-01",
			LastCheckedAt:   "2022-01-01",
			RuleHitsCount:   emptyJSON,
		},
		{
			OrgID:           3,
			ClusterID:       cluster3,
			NamespaceID:     "e6ed9bb3-efc3-46a6-b3ae-3f1a6e59546c",
			NamespaceName:   "not set",
			Report:          "",
			Recommendations: 6,
			Objects:         1,
			ReportedAt:      "2023-01-01",
			LastCheckedAt:   "2023-01-01",
			RuleHitsCount:   emptyJSON,
		},
	}

	var lastError error

	for _, record := range records {
		log.Info().
			Str("Insert statement", insertStatement).
			Msg("inserting into DVO database")
		// perform the SQL statement
		_, err := connection.Exec(insertStatement,
			record.OrgID, record.ClusterID, record.NamespaceID,
			record.NamespaceName, record.Report, record.Recommendations,
			record.Objects, record.ReportedAt, record.LastCheckedAt,
			record.RuleHitsCount)
		if err != nil {
			// failure is usually ok - it might mean that
			// the record with given org_id + cluster name already
			// exists
			log.Err(err).Msg("Insert error (DVO)")
			lastError = err
		}
	}
	log.Info().Msg("Fill-in DVO database finished")
	return lastError
}
