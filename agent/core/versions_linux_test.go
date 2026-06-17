//go:build linux

/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"testing"
)

// TestParseNmapVersion verifies that the Nmap version string is correctly extracted from canned output.
func TestParseNmapVersion(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name: "nmap-7x-with-npcap",
			output: "Nmap version 7.94 ( https://nmap.org )\n" +
				"Platform: x86_64-pc-linux-gnu\n" +
				"Compiled with: liblua-5.4.6 libpcre2-10.42 libz-1.2.13 openssl-3.1.1\n" +
				"\n" +
				"Npcap version 1.79, based on libpcap version 1.10.4\n",
			want: "7.94",
		},
		{
			name: "nmap-7x-without-npcap",
			output: "Nmap version 7.94 ( https://nmap.org )\n" +
				"Platform: x86_64-pc-linux-gnu\n" +
				"Compiled with: liblua-5.4.6 libpcre2-10.42 libz-1.2.13 openssl-3.1.1\n",
			want: "7.94",
		},
		{
			name:   "garbage-input",
			output: "something unparseable that contains no version",
			want:   "",
		},
		{
			name:   "empty-input",
			output: "",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseNmapVersion(tt.output); got != tt.want {
				t.Errorf("parseNmapVersion() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

// TestParseSslyzeVersion verifies that the SSLyze version string is correctly extracted from canned output.
func TestParseSslyzeVersion(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "sslyze-5x",
			output: "5.2.0\n",
			want:   "5.2.0",
		},
		{
			name:   "garbage-input",
			output: "something unparseable with spaces and no semver",
			want:   "",
		},
		{
			name:   "empty-input",
			output: "",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSslyzeVersion(tt.output); got != tt.want {
				t.Errorf("parseSslyzeVersion() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
