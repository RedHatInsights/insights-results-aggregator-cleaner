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

const (
	configFileEnvVariableName = "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// Messages
const (
	version                  = "Insights Results Aggregator Cleaner version 1.0"
	authors                  = "Pavel Tisnovsky, Red Hat Inc."
	properClusterID          = "Proper cluster ID"
	notProperClusterID       = "Not a proper cluster ID"
	improperClusterEntries   = "improper cluster entries"
	numberOfClustersToDelete = "number of clusters to delete"
	clusterListFinished      = "Cluster list finished"
	inputWithClusterID       = "input"
)

// IsValidUUID function checks if provided string contains a correct UUID.
func IsValidUUID(input string) bool {
	_, err := uuid.Parse(input)
	return err == nil
}

// readClusterList function reads list of clusters from provided text file or
// from CLI argument.
func readClusterList(filename string, clusters string) (ClusterList, int, error) {
	// if clusters are not specified on command line, read list of clusters
	// from file
	if clusters == "" {
		return readClusterListFromFile(filename)
	}
	// apparently list of clusters is specified on command line, so let's
	// use it properly
	return readClusterListFromCLIArgument(clusters)
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
	// #nosec G304
	file, err := os.Open(filename)
	if err != nil {
		return nil, improperClusterCounter, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Err(err).Msg("File close failed")
		}
	}()

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

// doSelectedOperation function performs selected operation: check data
// retention, cleanup selected data, or fill-id database by test data
func doSelectedOperation(config ConfigStruct, connection *sql.DB,
	showVersion bool, showAuthors bool, performCleanup bool,
	detectMultipleRuleDisable bool, fillInDatabase bool,
	printSummaryTable bool, clusters string,
	output string) error {
	switch {
	case showVersion:
		fmt.Println(version)
		return nil
	case showAuthors:
		fmt.Println(authors)
		return nil
	case performCleanup:
		// cleanup operation
		clusterList, improperClusterCounter, err := readClusterList(config.Cleaner.ClusterListFile, clusters)
		if err != nil {
			log.Err(err).Msg("Read cluster list")
			return err
		}
		deletionsForTable, err := performCleanupInDB(connection, clusterList)
		if err != nil {
			log.Err(err).Msg("Performing cleanup")
			return err
		}
		if printSummaryTable {
			var summary Summary
			summary.ProperClusterEntries = len(clusterList)
			summary.ImproperClusterEntries = improperClusterCounter
			summary.DeletionsForTable = deletionsForTable
			PrintSummaryTable(summary)
		}
		return nil
	case detectMultipleRuleDisable:
		// detect clusters that have the same rule(s) disabled by different users
		err := displayMultipleRuleDisable(connection, output)
		if err != nil {
			log.Err(err).Msg("Selecting records from database")
			return err
		}
		// everything seems to be fine
		return nil
	case fillInDatabase:
		// fill-in database by test data
		err := fillInDatabaseByTestData(connection)
		if err != nil {
			log.Err(err).Msg("Fill-in database by test data")
			return err
		}
		// everything seems to be fine
		return nil
	default:
		// display old records in database
		err := displayAllOldRecords(connection, config.Cleaner.MaxAge, output)
		if err != nil {
			log.Err(err).Msg("Selecting records from database")
			return err
		}
		// everything seems to be fine
		return nil
	}
	// we should not end there
	return nil
}

func main() {
	var performCleanup bool
	var printSummaryTable bool
	var detectMultipleRuleDisable bool
	var fillInDatabase bool
	var showVersion bool
	var showAuthors bool
	var maxAge string
	var clusters string
	var output string

	// define and parse all command line options
	flag.BoolVar(&performCleanup, "cleanup", false, "perform database cleanup")
	flag.BoolVar(&printSummaryTable, "summary", false, "print summary table after cleanup")
	flag.BoolVar(&detectMultipleRuleDisable, "multiple-rule-disable", false, "list clusters with the same rule(s) disabled by different users")
	flag.BoolVar(&fillInDatabase, "fill-in-db", false, "fill-in database by test data")
	flag.BoolVar(&showVersion, "version", false, "show cleaner version")
	flag.BoolVar(&showAuthors, "authors", false, "show authors")
	flag.StringVar(&maxAge, "max-age", "", "max age for displaying old records")
	flag.StringVar(&clusters, "clusters", "", "list of clusters to cleanup")
	flag.StringVar(&output, "output", "", "filename for old cluster listing")
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
	if maxAge != "" {
		config.Cleaner.MaxAge = maxAge
	}

	// initialize connection to database
	connection, err := initDatabaseConnection(config.Storage)
	if err != nil {
		log.Err(err).Msg("Connection to database not established")
	}

	// perform selected operation
	err = doSelectedOperation(config, connection, showVersion, showAuthors,
		performCleanup, detectMultipleRuleDisable, fillInDatabase,
		printSummaryTable, clusters,
		output)
	if err != nil {
		log.Err(err).Msg("Operation failed")
	}

	log.Debug().Msg("Finished")
}
