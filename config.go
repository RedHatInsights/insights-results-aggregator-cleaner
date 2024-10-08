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

// This source file contains definition of data type named ConfigStruct that
// represents configuration of Insights Results Aggregator Cleaner. This source
// file also contains function named LoadConfiguration that can be used to load
// configuration from provided configuration file and/or from environment
// variables. Additionally several specific functions named
// GetStorageConfiguration, GetLoggingConfiguration, and
// GetCleanerConfiguration are to be used to return specific configuration
// options.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-cleaner

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/config.html

// Default name of configuration file is config.toml
// It can be changed via environment variable INSIGHTS_RESULTS_CLEANER_CONFIG_FILE

// An example of configuration file that can be used in devel environment:
//
// [storage]
// db_driver = "postgres"
// pg_username = "postgres"
// pg_password = "postgres"
// pg_host = "localhost"
// pg_port = 5432
// pg_db_name = "aggregator"
// pg_params = "sslmode=disable"
// schema = "ocp_recommendations"
//
// [logging]
// debug = true
// log_level = ""
//
// [cleaner]
// max_age = "90 days"
// cluster_list_file = "cluster_list.txt"
//
//
// Environment variables that can be used to override configuration file settings:
// INSIGHTS_RESULTS_CLEANER__STORAGE__DB_DRIVER
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_USERNAME
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PASSWORD
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_HOST
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PORT
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_DB_NAME
// INSIGHTS_RESULTS_CLEANER__STORAGE__PG_PARAMS
// INSIGHTS_RESULTS_CLEANER__STORAGE__SCHEMA
// INSIGHTS_RESULTS_CLEANER__LOGGING__DEBUG
// INSIGHTS_RESULTS_CLEANER__LOGGING__LOG_DEVEL
// INSIGHTS_RESULTS_CLEANER__CLEANER__MAX_AGE

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/RedHatInsights/insights-operator-utils/logger"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Common constants used for during logging and error reporting
const (
	filenameAttribute               = "filename"
	parsingConfigurationFileMessage = "parsing configuration file"
)

// ConfigStruct is a structure holding the whole service configuration
type ConfigStruct struct {
	Storage StorageConfiguration              `mapstructure:"storage" toml:"storage"`
	Logging logger.LoggingConfiguration       `mapstructure:"logging" toml:"logging"`
	Cleaner CleanerConfiguration              `mapstructure:"cleaner" toml:"cleaner"`
	Sentry  logger.SentryLoggingConfiguration `mapstructure:"sentry" toml:"sentry"`
}

// CleanerConfiguration represents configuration for the main cleaner
type CleanerConfiguration struct {
	// MaxAge is specification of max age for records to be cleaned
	MaxAge string `mapstructure:"max_age" toml:"max_age"`
	// ClusterListFile contains file name with list of clusters to delete
	ClusterListFile string `mapstructure:"cluster_list_file" toml:"cluster_list_file"`
}

// StorageConfiguration represents configuration of data storage
type StorageConfiguration struct {
	Driver           string `mapstructure:"db_driver" toml:"db_driver"`
	SQLiteDataSource string `mapstructure:"sqlite_datasource" toml:"sqlite_datasource"`
	PGUsername       string `mapstructure:"pg_username" toml:"pg_username"`
	PGPassword       string `mapstructure:"pg_password" toml:"pg_password"`
	PGHost           string `mapstructure:"pg_host" toml:"pg_host"`
	PGPort           int    `mapstructure:"pg_port" toml:"pg_port"`
	PGDBName         string `mapstructure:"pg_db_name" toml:"pg_db_name"`
	PGParams         string `mapstructure:"pg_params" toml:"pg_params"`
	Schema           string `mapstructure:"schema" toml:"schema"`
}

// LoadConfiguration function loads configuration from defaultConfigFile, file
// set in configFileEnvVariableName or from environment variables
func LoadConfiguration(configFileEnvVariableName, defaultConfigFile string) (ConfigStruct, error) {
	var config ConfigStruct

	// env. variable holding name of configuration file
	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		log.Info().Str(filenameAttribute, configFile).Msg(parsingConfigurationFileMessage)
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		log.Info().Str(filenameAttribute, defaultConfigFile).Msg(parsingConfigurationFileMessage)
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	// try to read the whole configuration
	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// If configuration file is not present (which might be correct
		// in some environment) we need to read configuration from
		// environment variables. The problem is that Viper is not
		// smart enough to understand the structure of config by
		// itself, so we need to read fake config file
		fakeTomlConfigWriter := new(bytes.Buffer)

		err := toml.NewEncoder(fakeTomlConfigWriter).Encode(config)
		if err != nil {
			return config, err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))

		// check for error during parsing
		if err != nil {
			return config, err
		}
	} else if err != nil {
		// error is processed on caller side
		return config, fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env

	const envPrefix = "INSIGHTS_RESULTS_CLEANER_"

	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	// try to unmarshall configuration and check for (any) error
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("fatal - can not unmarshal configuration: %s", err)
	}

	// updated configuration by introducing Clowder-related things
	if err := updateConfigFromClowder(&config); err != nil {
		fmt.Println("Error loading clowder configuration")
		return config, err
	}
	return config, err
}

// GetStorageConfiguration function returns storage configuration
func GetStorageConfiguration(config *ConfigStruct) StorageConfiguration {
	return config.Storage
}

// GetLoggingConfiguration function returns logging configuration
func GetLoggingConfiguration(config *ConfigStruct) logger.LoggingConfiguration {
	return config.Logging
}

// GetSentryConfiguration function returns sentry configuration
func GetSentryConfiguration(config *ConfigStruct) logger.SentryLoggingConfiguration {
	return config.Sentry
}

// GetCleanerConfiguration returns cleaner configuration
func GetCleanerConfiguration(config *ConfigStruct) CleanerConfiguration {
	return config.Cleaner
}

// updateConfigFromClowder function updates the current config with the values
// defined in clowder
func updateConfigFromClowder(c *ConfigStruct) error {
	if !clowder.IsClowderEnabled() || clowder.LoadedConfig == nil {
		// can not use Zerolog at this moment!
		// we have to use standard output
		fmt.Println("Clowder is disabled")
		return nil
	}

	fmt.Println("Clowder is enabled")

	// get DB configuration from clowder
	c.Storage.PGDBName = clowder.LoadedConfig.Database.Name
	c.Storage.PGHost = clowder.LoadedConfig.Database.Hostname
	c.Storage.PGPort = clowder.LoadedConfig.Database.Port
	c.Storage.PGUsername = clowder.LoadedConfig.Database.Username
	c.Storage.PGPassword = clowder.LoadedConfig.Database.Password

	return nil
}

// StringSet type is a poor man's implementation of set of strings
type StringSet map[string]struct{}

// allSupportedDrivers constructs set with names of all supported database
// drivers
func allSupportedDrivers() StringSet {
	var drivers = make(StringSet)
	drivers["sqlite3"] = struct{}{}
	drivers["postgres"] = struct{}{}
	return drivers
}

// allSupportedSchemas constructs set with names of all supported database
// schemas
func allSupportedSchemas() StringSet {
	var schemas = make(StringSet)
	schemas["ocp_recommendations"] = struct{}{}
	schemas["dvo_recommendations"] = struct{}{}
	return schemas
}

// CheckConfiguration function checks if loaded configuration contains expected
// items
func CheckConfiguration(config *ConfigStruct) error {
	drivers := allSupportedDrivers()
	schemas := allSupportedSchemas()

	storageCfg := GetStorageConfiguration(config)
	driver := storageCfg.Driver
	schema := storageCfg.Schema

	if driver == "" {
		return fmt.Errorf("Database driver is not specified in configuration")
	}

	if schema == "" {
		return fmt.Errorf("Database schema is not specified in configuration")
	}

	_, found := drivers[driver]
	if !found {
		return fmt.Errorf("Incorrect database driver found in configuration: %s", driver)
	}

	_, found = schemas[schema]
	if !found {
		return fmt.Errorf("Incorrect database schema found in configuration: %s", schema)
	}

	return nil
}
