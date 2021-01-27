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

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ConfigStruct is a structure holding the whole service configuration
type ConfigStruct struct {
	Storage StorageConfiguration `mapstructure:"storage" toml:"storage"`
	Logging LoggingConfiguration `mapstructure:"logging" toml:"logging"`
	Cleaner CleanerConfiguration `mapstructure:"cleaner" toml:"cleaner"`
}

// LoggingConfiguration represents configuration for logging in general
type LoggingConfiguration struct {
	// Debug enables pretty colored logging
	Debug bool `mapstructure:"debug" toml:"debug"`

	// LogLevel sets logging level to show. Possible values are:
	// "debug"
	// "info"
	// "warn", "warning"
	// "error"
	// "fatal"
	//
	// logging level won't be changed if value is not one of listed above
	LogLevel string `mapstructure:"log_level" toml:"log_level"`

	// LoggingToCloudWatchEnabled enables logging to CloudWatch
	// (configuration for CloudWatch is in CloudWatchConfiguration)
	LoggingToCloudWatchEnabled bool `mapstructure:"logging_to_cloud_watch_enabled" toml:"logging_to_cloud_watch_enabled"`
}

// CleanerConfiguration represents configuration for the main cleaner
type CleanerConfiguration struct {
	// MaxAge is specification of max age for records to be cleaned
	MaxAge string `mapstructure:"max_age" toml:"max_age"`
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
}

// LoadConfiguration loads configuration from defaultConfigFile, file set in
// configFileEnvVariableName or from env
func LoadConfiguration(configFileEnvVariableName string, defaultConfigFile string) (ConfigStruct, error) {
	var config ConfigStruct

	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		log.Info().Str("filename", defaultConfigFile).Msg("Parsing configuration file")
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// viper is not smart enough to understand the structure of
		// config by itself
		fakeTomlConfigWriter := new(bytes.Buffer)

		err := toml.NewEncoder(fakeTomlConfigWriter).Encode(config)
		if err != nil {
			return config, err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))
		if err != nil {
			return config, err
		}
	} else if err != nil {
		return config, fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env

	const envPrefix = "INSIGHTS_RESULTS_CLEANER_"

	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	err = viper.Unmarshal(&config)
	return config, err
}
