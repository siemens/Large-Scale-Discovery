/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"reflect"
	"testing"
)

// TestGetConfig verifies that GetConfig returns the currently stored global importer configuration.
func TestGetConfig(t *testing.T) {

	// Set the initial global importer config, in case all test cases get executed at once it will be already set
	Set(&Importer{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want Importer
	}{
		{
			name: "empty-importer-config",
			want: Importer{},
		},
		{
			name: "default-importer-config",
			want: defaultImporterConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		// Advance the global config state for the next test case
		conf := defaultImporterConfigFactory()
		Set(&conf)
	}
}
