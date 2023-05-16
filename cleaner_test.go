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
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/cleaner_test.html

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"
	"os"
	"testing"

	main "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func checkCapture(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// TestShowVersion checks the function showVersion
func TestShowVersion(t *testing.T) {
	const expected = "Insights Results Aggregator Cleaner version 1.0\n"

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowVersion()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expected)
}

// TestShowAuthors checks the function showAuthors
func TestShowAuthors(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowAuthors()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Red Hat Inc.")
}

// TestShowConfiguration checks the function ShowConfiguration
func TestShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	configuration.Storage = main.StorageConfiguration{
		Driver:     "postgres",
		PGUsername: "foo",
		PGPassword: "bar",
		PGHost:     "baz",
		PGDBName:   "aggregator",
		PGParams:   ""}
	configuration.Logging = main.LoggingConfiguration{
		Debug:    true,
		LogLevel: ""}
	configuration.Cleaner = main.CleanerConfiguration{
		MaxAge:          "3 days",
		ClusterListFile: "cluster_list.txt"}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.New(os.Stderr))

		main.ShowConfiguration(&configuration)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Driver")
	assert.Contains(t, output, "Level")
	assert.Contains(t, output, "Records max age")
}

func TestIsValidUUID(t *testing.T) {
	type UUID struct {
		id    string
		valid bool
	}

	uuids := []UUID{
		UUID{
			id:    "",
			valid: false,
		},
		UUID{
			id:    "00000000-0000-0000-0000-000000000000",
			valid: true,
		},
		UUID{
			id:    "5d5892d4-1f74-4ccf-91af-548dfc9767aa",
			valid: true,
		},
		UUID{ // x at beginning
			id:    "xd5892d4-1f74-4ccf-91af-548dfc9767aa",
			valid: false,
		},
		UUID{ // wrong separator
			id:    "5d5892d4-1f74-4cc-f91af-548dfc9767aa",
			valid: false,
		},
	}

	for _, uuid := range uuids {
		v := main.IsValidUUID(uuid.id)
		assert.Equal(t, v, uuid.valid)

	}

}

// TestDoSelectedOperationShowVersion checks the function showVersion called
// via doSelectedOperation function
func TestDoSelectedOperationShowVersion(t *testing.T) {
	const expected = "Insights Results Aggregator Cleaner version 1.0\n"

	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       true,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expected)
}

// TestDoSelectedOperationShowAuthors checks the function showAuthors called
// via doSelectedOperation function
func TestDoSelectedOperationShowAuthors(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       true,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Red Hat Inc.")
}

// TestDoSelectedOperationShowConfiguration checks the function
// showConfiguration called via doSelectedOperation function
func TestDoSelectedOperationShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}

	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: true,
	}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.New(os.Stderr))

		code, err := main.DoSelectedOperation(&configuration, nil, cliFlags)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "Driver")
	assert.Contains(t, output, "Level")
	assert.Contains(t, output, "Records max age")
}
