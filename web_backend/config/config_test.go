/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"reflect"
	"testing"
)

// TestGetConfig verifies that GetConfig returns the currently stored global web configuration.
func TestGetConfig(t *testing.T) {

	// Set the initial global web config, in case all test cases get executed at once it will be already set
	Set(&Web{})

	// Prepare and run test cases
	tests := []struct {
		name     string
		wantConf Web
	}{
		{
			name:     "empty-web-config",
			wantConf: Web{},
		},
		{
			name:     "default-web-config",
			wantConf: defaultWebConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfig()
			got.Jwt.Secret = ""         // Remove random data from struct
			tt.wantConf.Jwt.Secret = "" // Remove random data from struct
			if !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.wantConf)
			}
		})

		// Advance the global config state for the next test case
		conf := defaultWebConfigFactory()
		Set(&conf)
	}
}
