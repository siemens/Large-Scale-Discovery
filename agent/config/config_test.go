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
	"os/exec"
	"reflect"
	"testing"
)

// TestGetConfig verifies that GetConfig returns the currently stored global agent configuration.
func TestGetConfig(t *testing.T) {

	// Skip when required executables are absent. The config UnmarshalJSON validates their paths.
	if _, errLookPath := exec.LookPath("nmap"); errLookPath != nil {
		t.Skip("Integration test skipped: Nmap not available")
		return
	}
	if _, errLookPath := exec.LookPath("python3"); errLookPath != nil {
		t.Skip("Integration test skipped: python3 not available")
		return
	}

	// Set the initial global agent config, in case all test cases get executed at once it will be already set
	Set(&Agent{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want Agent
	}{
		{
			name: "empty-agent-config",
			want: Agent{},
		},
		{
			name: "default-agent-config",
			want: defaultAgentConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		// Advance the global config state for the next test case
		conf := defaultAgentConfigFactory()
		Set(&conf)
	}
}
