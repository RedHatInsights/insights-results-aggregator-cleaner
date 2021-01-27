package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	configFileEnvVariableName = "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

func main() {
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

	err = displayAllOldRecords(connection, config.Cleaner.MaxAge)
	if err != nil {
		log.Err(err).Msg("Selecting records from database")
	}

	log.Debug().Msg("Finished")
}
