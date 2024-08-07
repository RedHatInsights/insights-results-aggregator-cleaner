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
// https://redhatinsights.github.io/insights-results-aggregator-cleaner/packages/export_test.html

// Export for testing
//
// This source file contains name aliases of all package-private functions
// that need to be called from unit tests. Aliases should start with uppercase
// letter because unit tests belong to different package.
//
// Please look into the following blogpost:
// https://medium.com/@robiplus/golang-trick-export-for-test-aa16cbd7b8cd
// to see why this trick is needed.
var (
	TablesAndKeysInOCPDatabase = tablesAndKeysInOCPDatabase
	TablesAndKeysInDVODatabase = tablesAndKeysInDVODatabase

	// functions from the storage.go source file
	ReadOrgID                         = readOrgID
	DisplayMultipleRuleDisable        = displayMultipleRuleDisable
	DisplayAllOldRecords              = displayAllOldRecords
	PerformDisplayMultipleRuleDisable = performDisplayMultipleRuleDisable
	PerformListOfOldOCPReports        = performListOfOldOCPReports
	PerformListOfOldDVOReports        = performListOfOldDVOReports
	PerformListOfOldRatings           = performListOfOldRatings
	PerformListOfOldConsumerErrors    = performListOfOldConsumerErrors
	DeleteRecordFromTable             = deleteRecordFromTable
	PerformCleanupInDB                = performCleanupInDB
	PerformCleanupAllInDB             = performCleanupAllInDB
	PerformVacuumDB                   = performVacuumDB
	FillInDatabaseByTestData          = fillInDatabaseByTestData
	InitDatabaseConnection            = initDatabaseConnection

	// functions from the cleaner.go source file
	ShowVersion                    = showVersion
	ShowAuthors                    = showAuthors
	ShowConfiguration              = showConfiguration
	DoSelectedOperation            = doSelectedOperation
	ReadClusterList                = readClusterList
	ReadClusterListFromFile        = readClusterListFromFile
	ReadClusterListFromCLIArgument = readClusterListFromCLIArgument
	VacuumDB                       = vacuumDB
	Cleanup                        = cleanup
	CleanupAll                     = cleanupAll
	FillInDatabase                 = fillInDatabase
	DisplayOldRecords              = displayOldRecords
	DetectMultipleRuleDisable      = detectMultipleRuleDisable

	// constants
	MaxAgeMissing     = maxAgeMissing
	TablesToDeleteOCP = tablesToDeleteOCP
	TablesToDeleteDVO = tablesToDeleteDVO
	AllTablesToDelete = allTablesToDelete
	EmptyJSON         = emptyJSON
)
