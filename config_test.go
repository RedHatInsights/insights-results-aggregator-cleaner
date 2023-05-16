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

package main_test

// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/config_test.html

import (
	"os"

	"testing"

	"github.com/RedHatInsights/insights-operator-utils/tests/helpers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

func mustLoadConfiguration(envVar string) {
	_, err := main.LoadConfiguration(envVar, "tests/config1")
	if err != nil {
		panic(err)
	}
}

func mustSetEnv(t *testing.T, key, val string) {
	err := os.Setenv(key, val)
	helpers.FailOnError(t, err)
}

// TestLoadDefaultConfiguration loads a configuration file for testing
func TestLoadDefaultConfiguration(t *testing.T) {
	os.Clearenv()
	mustLoadConfiguration("nonExistingEnvVar")
}

// TestLoadConfigurationFromEnvVariable tests loading the config. file for
// testing from an environment variable
func TestLoadConfigurationFromEnvVariable(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "tests/config2")
	mustLoadConfiguration("INSIGHTS_RESULTS_CLEANER_CONFIG_FILE")
}

// TestLoadConfigurationNonEnvVarUnknownConfigFile tests loading an unexisting
// config file when no environment variable is provided
func TestLoadConfigurationNonEnvVarUnknownConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "foobar")
	assert.Nil(t, err)
}

// TestLoadConfigurationBadConfigFile tests loading an unexisting config file
// when no environment variable is provided
func TestLoadConfigurationBadConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "tests/config3")
	assert.Contains(t, err.Error(), `fatal error config file: While parsing config:`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests loading a
// non-existent configuration file set in environment
func TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests that if env
// var is provided, it must point to a valid config file
func TestLoadingConfigurationEnvVariableBadValueDefaultConfigFailure(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "tests/config1")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadCleanerConfiguration tests loading the cleaner configuration
// sub-tree
func TestLoadCleanerConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"

	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	cleanerCfg := main.GetCleanerConfiguration(&config)

	assert.Equal(t, "90 days", cleanerCfg.MaxAge)
	assert.Equal(t, "cluster_list.txt", cleanerCfg.ClusterListFile)
}

// TestLoadStorageConfiguration tests loading the storage configuration
// sub-tree
func TestLoadStorageConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	storageCfg := main.GetStorageConfiguration(&config)

	assert.Equal(t, "sqlite3", storageCfg.Driver)
	assert.Equal(t, "user", storageCfg.PGUsername)
	assert.Equal(t, "password", storageCfg.PGPassword)
	assert.Equal(t, "localhost", storageCfg.PGHost)
	assert.Equal(t, 5432, storageCfg.PGPort)
	assert.Equal(t, "notifications", storageCfg.PGDBName)
	assert.Equal(t, "", storageCfg.PGParams)
}

// TestLoadLoggingConfiguration tests loading the logging configuration
// sub-tree
func TestLoadLoggingConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	loggingCfg := main.GetLoggingConfiguration(&config)

	assert.Equal(t, true, loggingCfg.Debug)
	assert.Equal(t, "", loggingCfg.LogLevel)
}

// TestLoadConfigurationFromEnvVariableClowderEnabled tests loading the config.
// file for testing from an environment variable. Clowder config is enabled in
// this case.
func TestLoadConfigurationFromEnvVariableClowderEnabled(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_CLEANER_CONFIG_FILE", "tests/config2")
	mustSetEnv(t, "ACG_CONFIG", "tests/clowder_config.json")
	mustLoadConfiguration("INSIGHTS_RESULTS_CLEANER_CONFIG_FILE")
}
