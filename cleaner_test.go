/*
Copyright © 2021, 2022 Red Hat, Inc.

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
	"github.com/stretchr/testify/assert"
	"testing"

	main "github.com/RedHatInsights/insights-results-aggregator-cleaner"
)

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
