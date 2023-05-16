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

// Simple service that can be used to identify clusters, for which we are
// keeping very old data (>30 days) in the database. This means that the
// cluster is no longer available or that the customer has disabled the
// Insights Operator, either way it means that these data are no longer
// relevant to us and should be pruned.
//
// Such clusters can be detected very easily by checking the timestamps stored
// (along other information) in the `report` table in Insights Results
// Aggregator database.
//
// Currently this service just displays such clusters (cluster IDs) and do
// nothing else - i.e. the results are not deleted by default.
//
// Additionally it is possible to detect and displays clusters where any rule
// have been disabled by multiple users at the same time. Such records might
// have to be deleted in order to maintain database consistency.
package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-cleaner
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner.html

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"database/sql"

	"github.com/google/uuid"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/olekukonko/tablewriter"
)

// Messages
const (
	versionMessage               = "Insights Results Aggregator Cleaner version 1.0"
	authorsMessage               = "Pavel Tisnovsky, Red Hat Inc."
	properClusterID              = "Proper cluster ID"
	notProperClusterID           = "Not a proper cluster ID"
	improperClusterEntries       = "improper cluster entries"
	numberOfClustersToDelete     = "number of clusters to delete"
	clusterListFinished          = "Cluster list finished"
	inputWithClusterID           = "input"
	selectingRecordsFromDatabase = "Selecting records from database"
)

// Exit codes
const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota

	// ExitStatusStorageError is returned in case of any storage-related
	// error
	ExitStatusStorageError

	// ExitStatusFillInStorageError is returned in case the fill-in DB
	// operation failed
	ExitStatusFillInStorageError

	// ExitStatusPerformCleanupError is returned when DB cleanup operation
	// failed for any reason
	ExitStatusPerformCleanupError

	// ExitStatusPerformVacuumError is returned when DB vacuuming operation
	// have failed for any reason
	ExitStatusPerformVacuumError
)

const (
	configFileEnvVariableName = "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// showVersion function displays version information.
func showVersion() {
	fmt.Println(versionMessage)
}

// showAuthors function displays information about authors.
func showAuthors() {
	fmt.Println(authorsMessage)
}

// IsValidUUID function checks if provided string contains a correct UUID.
func IsValidUUID(input string) bool {
	_, err := uuid.Parse(input)
	return err == nil
}

// readClusterList function reads list of clusters from provided text file or
// from CLI argument.
func readClusterList(filename, clusters string) (ClusterList, int, error) {
	// if clusters are not specified on command line, read list of clusters
	// from file
	if clusters == "" {
		return readClusterListFromFile(filename)
	}
	// apparently list of clusters is specified on command line, so let's
	// use it properly
	return readClusterListFromCLIArgument(clusters)
}

// showConfiguration function displays actual configuration.
func showConfiguration(config *ConfigStruct) {
	storageConfig := GetStorageConfiguration(config)
	log.Info().
		Str("Driver", storageConfig.Driver).
		Str("DB Name", storageConfig.PGDBName).
		Str("Username", storageConfig.PGUsername). // password is omitted on purpose
		Str("Host", storageConfig.PGHost).
		Int("DB Port", storageConfig.PGPort).
		Msg("Storage configuration")

	loggingConfig := GetLoggingConfiguration(config)
	log.Info().
		Str("Level", loggingConfig.LogLevel).
		Bool("Pretty colored debug logging", loggingConfig.Debug).
		Bool("Log to cloudwatch", loggingConfig.LoggingToCloudWatchEnabled).
		Msg("Logging configuration")

	cleanerConfiguration := GetCleanerConfiguration(config)
	log.Info().
		Str("Records max age", cleanerConfiguration.MaxAge).
		Str("Cluster list file", cleanerConfiguration.ClusterListFile).
		Msg("Cleaner configuration")
}

// readClusterListFromCLIArgument reads list of clusters from CLI argument
func readClusterListFromCLIArgument(clusters string) (ClusterList, int, error) {
	log.Debug().Msg("Cluster list read from CLI argument")

	improperClusterCounter := 0

	var clusterList = make([]ClusterName, 0)

	v := strings.Split(clusters, ",")

	for _, cluster := range v {
		cluster := strings.Trim(cluster, " ")
		// check if line contains proper cluster ID (as UUID)
		if IsValidUUID(cluster) {
			clusterList = append(clusterList, ClusterName(cluster))
			log.Info().Str(inputWithClusterID, cluster).Msg(properClusterID)
		} else {
			log.Error().Str(inputWithClusterID, cluster).Msg(notProperClusterID)
			improperClusterCounter++
		}
	}
	log.Info().Int(numberOfClustersToDelete, len(clusterList)).Msg(clusterListFinished)
	log.Info().Int(improperClusterEntries, improperClusterCounter).Msg(clusterListFinished)

	return clusterList, improperClusterCounter, nil
}

// readClusterListFromFile function reads list of clusters from provided text
// file.
func readClusterListFromFile(filename string) (ClusterList, int, error) {
	log.Debug().Msg("Cluster list read from file")

	improperClusterCounter := 0

	var clusterList = make([]ClusterName, 0)

	// disable "G304 (CWE-22): Potential file inclusion via variable"
	file, err := os.Open(filename) // #nosec G304
	if err != nil {
		return nil, improperClusterCounter, err
	}

	// start reading from the file with a reader
	reader := bufio.NewReader(file)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, "\n")
		// check if line contains proper cluster ID (as UUID)
		if IsValidUUID(line) {
			clusterList = append(clusterList, ClusterName(line))
			log.Info().Str(inputWithClusterID, line).Msg(properClusterID)
		} else {
			log.Error().Str(inputWithClusterID, line).Msg(notProperClusterID)
			improperClusterCounter++
		}
	}
	log.Info().Int(numberOfClustersToDelete, len(clusterList)).Msg(clusterListFinished)
	log.Info().Int(improperClusterEntries, improperClusterCounter).Msg(clusterListFinished)

	// close file and catch any I/O error
	err = file.Close()
	if err != nil {
		// if error is detected during file close, we need to inform
		// caller about it
		log.Err(err).Msg("File close failed")
		return clusterList, improperClusterCounter, err
	}

	return clusterList, improperClusterCounter, nil
}

// PrintSummaryTable function displays a table with summary information about
// cleanup step.
func PrintSummaryTable(summary Summary) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(60)

	// table header
	table.SetHeader([]string{"Summary", "Count"})

	table.Append([]string{"Proper cluster entries",
		strconv.Itoa(summary.ProperClusterEntries)})
	table.Append([]string{"Improper cluster entries",
		strconv.Itoa(summary.ImproperClusterEntries)})
	table.Append([]string{"", ""})

	totalDeletions := 0

	// prepare rows with info about deletions
	for tableName, deletions := range summary.DeletionsForTable {
		totalDeletions += deletions
		table.Append([]string{"Deletions from table '" + tableName + "'",
			strconv.Itoa(deletions)})
	}

	// table footer
	table.SetFooter([]string{"Total deletions",
		strconv.Itoa(totalDeletions)})

	// display the whole table
	table.Render()
}

// vacuumDB function starts the database vacuuming operation
func vacuumDB(connection *sql.DB) (int, error) {
	err := performVacuumDB(connection)
	if err != nil {
		log.Err(err).Msg("Performing vacuuming database")
		return ExitStatusPerformVacuumError, err
	}
	return ExitStatusOK, nil
}

// cleanup function starts the cleanup operation
func cleanup(configuration *ConfigStruct, connection *sql.DB, cliFlags CliFlags) (int, error) {
	// cleanup operation
	clusterList, improperClusterCounter, err := readClusterList(
		configuration.Cleaner.ClusterListFile,
		cliFlags.Clusters)
	if err != nil {
		log.Err(err).Msg("Read cluster list")
		return ExitStatusPerformCleanupError, err
	}
	deletionsForTable, err := performCleanupInDB(connection, clusterList)
	if err != nil {
		log.Err(err).Msg("Performing cleanup")
		return ExitStatusPerformCleanupError, err
	}
	if cliFlags.PrintSummaryTable {
		var summary Summary
		summary.ProperClusterEntries = len(clusterList)
		summary.ImproperClusterEntries = improperClusterCounter
		summary.DeletionsForTable = deletionsForTable
		PrintSummaryTable(summary)
	}
	return ExitStatusOK, nil
}

// detectMultipleRuleDisable function detects clusters that have the same
// rule(s) disabled by different users
func detectMultipleRuleDisable(connection *sql.DB, cliFlags CliFlags) (int, error) {
	err := displayMultipleRuleDisable(connection, cliFlags.Output)
	if err != nil {
		log.Err(err).Msg(selectingRecordsFromDatabase)
		return ExitStatusStorageError, err
	}
	// everything seems to be fine
	return ExitStatusOK, nil
}

// fillInDatabase function fills-in database by test data
func fillInDatabase(connection *sql.DB) (int, error) {
	err := fillInDatabaseByTestData(connection)
	if err != nil {
		log.Err(err).Msg("Fill-in database by test data")
		return ExitStatusFillInStorageError, err
	}
	// everything seems to be fine
	return ExitStatusOK, nil
}

// displayOldRecords function displays old records in database
func displayOldRecords(configuration *ConfigStruct, connection *sql.DB, cliFlags CliFlags) (int, error) {
	err := displayAllOldRecords(connection,
		configuration.Cleaner.MaxAge, cliFlags.Output)
	if err != nil {
		log.Err(err).Msg(selectingRecordsFromDatabase)
		return ExitStatusStorageError, err
	}
	// everything seems to be fine
	return ExitStatusOK, nil
}

// doSelectedOperation function performs selected operation: check data
// retention, cleanup selected data, or fill-id database by test data
func doSelectedOperation(configuration *ConfigStruct, connection *sql.DB, cliFlags CliFlags) (int, error) {
	switch {
	case cliFlags.ShowVersion:
		showVersion()
		return ExitStatusOK, nil
	case cliFlags.ShowAuthors:
		showAuthors()
		return ExitStatusOK, nil
	case cliFlags.ShowConfiguration:
		showConfiguration(configuration)
		return ExitStatusOK, nil
	case cliFlags.VacuumDatabase:
		return vacuumDB(connection)
	case cliFlags.PerformCleanup:
		return cleanup(configuration, connection, cliFlags)
	case cliFlags.DetectMultipleRuleDisable:
		return detectMultipleRuleDisable(connection, cliFlags)
	case cliFlags.FillInDatabase:
		return fillInDatabase(connection)
	default:
		return displayOldRecords(configuration, connection, cliFlags)
	}
	// we should not end there
}

func main() {
	// command line flags
	var cliFlags CliFlags

	// define and parse all command line options
	flag.BoolVar(&cliFlags.PerformCleanup, "cleanup", false, "perform database cleanup")
	flag.BoolVar(&cliFlags.PrintSummaryTable, "summary", false, "print summary table after cleanup")
	flag.BoolVar(&cliFlags.DetectMultipleRuleDisable, "multiple-rule-disable", false, "list clusters with the same rule(s) disabled by different users")
	flag.BoolVar(&cliFlags.FillInDatabase, "fill-in-db", false, "fill-in database by test data")
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.ShowVersion, "version", false, "show cleaner version")
	flag.BoolVar(&cliFlags.ShowAuthors, "authors", false, "show authors")
	flag.BoolVar(&cliFlags.VacuumDatabase, "vacuum", false, "vacuum database")
	flag.StringVar(&cliFlags.MaxAge, "max-age", "", "max age for displaying old records")
	flag.StringVar(&cliFlags.Clusters, "clusters", "", "list of clusters to cleanup")
	flag.StringVar(&cliFlags.Output, "output", "", "filename for old cluster listing")

	// parse all command line flags
	flag.Parse()

	// config has exactly the same structure as *.toml file
	config, err := LoadConfiguration(configFileEnvVariableName, defaultConfigFileName)
	if err != nil {
		log.Err(err).Msg("Load configuration")
	}

	if config.Logging.Debug {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Debug().Msg("Started")

	// override default value read from configuration file
	if cliFlags.MaxAge != "" {
		config.Cleaner.MaxAge = cliFlags.MaxAge
	}

	// initialize connection to database
	connection, err := initDatabaseConnection(config.Storage)
	if err != nil {
		log.Err(err).Msg("Connection to database not established")
	}

	// perform selected operation
	exitStatus, err := doSelectedOperation(&config, connection, cliFlags)
	if err != nil {
		log.Err(err).Msg("Operation failed")
		os.Exit(exitStatus)
		return
	}

	// finito
	log.Debug().Msg("Finished")
	os.Exit(ExitStatusOK)
}
