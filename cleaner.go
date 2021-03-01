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
// nothing else - i.e. the results are not deleted.
package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-cleaner

import (
	"bufio"
	"flag"
	"os"
	"strings"

	"database/sql"

	"github.com/google/uuid"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	configFileEnvVariableName = "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// IsValidUUID function checks if provided string contains a correct UUID.
func IsValidUUID(input string) bool {
	_, err := uuid.Parse(input)
	return err == nil
}

func readClusterList(filename string) (ClusterList, error) {
	log.Debug().Msg("Cluster list read")

	var clusterList = make([]ClusterName, 0)

	// disable "G304 (CWE-22): Potential file inclusion via variable"
	// #nosec G304
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
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
		if IsValidUUID(line) {
			clusterList = append(clusterList, ClusterName(line))
			log.Info().Str("input", line).Msg("Proper cluster ID")
		} else {
			log.Error().Str("input", line).Msg("Not a proper cluster ID")
		}
	}
	log.Info().Int("number of clusters to delete", len(clusterList)).Msg("Cluster list finished")

	return clusterList, nil
}

func doSelectedOperation(config ConfigStruct, connection *sql.DB, performCleanup bool) error {
	if performCleanup {
		clusterList, err := readClusterList(config.Cleaner.ClusterListFile)
		if err != nil {
			log.Err(err).Msg("Read cluster list")
			return err
		}
		err = performCleanupInDB(connection, clusterList)
		if err != nil {
			log.Err(err).Msg("Performing cleanup")
			return err
		}
	} else {
		err := displayAllOldRecords(connection, config.Cleaner.MaxAge)
		if err != nil {
			log.Err(err).Msg("Selecting records from database")
			return err
		}
	}
	// everything seems to be fine
	return nil
}

func main() {
	var performCleanup bool

	flag.BoolVar(&performCleanup, "cleanup", false, "perform database cleanup")
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

	connection, err := initDatabaseConnection(config.Storage)
	if err != nil {
		log.Err(err).Msg("Connection to database not established")
	}

	err = doSelectedOperation(config, connection, performCleanup)
	if err != nil {
		log.Err(err).Msg("Operation failed")
	}

	log.Debug().Msg("Finished")
}
