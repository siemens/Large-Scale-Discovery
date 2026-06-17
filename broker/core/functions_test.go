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

// TestExtractMinHostgroup verifies extractMinHostgroup parses the --min-hostgroup flag across all supported formats.
func TestExtractMinHostgroup(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name string
		args string
		want int
	}{
		{
			name: "empty-args",
			args: "",
			want: 0,
		},
		{
			name: "flag-not-present",
			args: "-sV --open -T4",
			want: 0,
		},
		{
			name: "flag-present-lowercase",
			args: "-T4 --min-hostgroup 32",
			want: 32,
		},
		{
			name: "flag-present-uppercase",
			args: "-T4 --MIN-HOSTGROUP 64",
			want: 64,
		},
		{
			name: "flag-present-mixedcase",
			args: "--Min-Hostgroup 128",
			want: 128,
		},
		{
			name: "flag-at-start",
			args: "--min-hostgroup 16 -sV",
			want: 16,
		},
		{
			name: "default-64-value",
			args: "--min-hostgroup 64",
			want: 64,
		},
		{
			name: "value-1",
			args: "--min-hostgroup 1",
			want: 1,
		},
		{
			name: "flag-value-not-integer",
			args: "--min-hostgroup abc",
			want: 0,
		},
		{
			name: "other-flags-only",
			args: "--min-rate 100 --max-retries 3",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMinHostgroup(tt.args)
			if got != tt.want {
				t.Errorf("extractMinHostgroup() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

// TestExtractHostTimeoutMinutes verifies extractHostTimeoutMinutes converts Nmap timeout values to minutes correctly.
func TestExtractHostTimeoutMinutes(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name string
		args string
		want float64
	}{
		{
			name: "empty-args",
			args: "",
			want: 0,
		},
		{
			name: "flag-not-present",
			args: "-sV --open -T4",
			want: 0,
		},
		{
			name: "minutes-unit",
			args: "--host-timeout 60m",
			want: 60,
		},
		{
			name: "minutes-unit-uppercase",
			args: "--HOST-TIMEOUT 30M",
			want: 30,
		},
		{
			name: "hours-unit",
			args: "--host-timeout 2h",
			want: 120,
		},
		{
			name: "hours-unit-1h",
			args: "--host-timeout 1h",
			want: 60,
		},
		{
			name: "seconds-unit",
			args: "--host-timeout 120s",
			want: 2,
		},
		{
			name: "seconds-unit-60s",
			args: "--host-timeout 60s",
			want: 1,
		},
		{
			name: "milliseconds-unit",
			args: "--host-timeout 60000ms",
			want: 1,
		},
		{
			name: "milliseconds-unit-120000ms",
			args: "--host-timeout 120000ms",
			want: 2,
		},
		{
			name: "no-unit-means-seconds",
			args: "--host-timeout 300",
			want: 5,
		},
		{
			name: "no-unit-60",
			args: "--host-timeout 60",
			want: 1,
		},
		{
			name: "flag-with-other-flags",
			args: "-T4 --host-timeout 720m --open",
			want: 720,
		},
		{
			name: "invalid-value",
			args: "--host-timeout abc",
			want: 0,
		},
		{
			name: "invalid-value-with-unit",
			args: "--host-timeout abcm",
			want: 0,
		},
		{
			name: "other-flags-only",
			args: "--min-hostgroup 64 -T4",
			want: 0,
		},
		{
			name: "default-nmap-timeout-minutes",
			args: "--host-timeout 720m",
			want: float64(defaultNmapHosttimeoutMinutes),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHostTimeoutMinutes(tt.args)
			if got != tt.want {
				t.Errorf("extractHostTimeoutMinutes() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
