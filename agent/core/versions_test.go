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
			name: "nmap-7x-with-npcap-line",
			output: "Nmap version 7.94 ( https://nmap.org )\n" +
				"Platform: x86_64-pc-windows-msvc\n" +
				"Compiled with: nmap-libpcap-1.10.4 nmap-libdnet-1.12 ipv6\n" +
				"\n" +
				"Npcap version 1.79, based on libpcap version 1.10.4\n",
			want: "7.94",
		},
		{
			name: "nmap-7x-without-npcap-line",
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

// TestParseSslyzeVersion verifies that the SSLyze version string is correctly extracted from
// canned -h output. Test snippets reconstructed from upstream source (command_line_parser.py)
// at the named tags; the "SSLyze version X.Y.Z" anchor line is invariant across all 5.x/6.x.
func TestParseSslyzeVersion(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			// Reconstructed from sslyze 5.1.3 source (tag v5.1.3).
			// --keyform uses metavar=KEY_FORMAT; --starttls uses metavar=PROTOCOL (no argparse choices=);
			// --sni uses metavar=SERVER_NAME_INDICATION. No --custom_tls_config in this version.
			name: "sslyze-5.1.3-help",
			output: "usage: sslyze [-h] [--update_trust_stores] [--cert CERTIFICATE_FILE]\n" +
				"              [--key KEY_FILE] [--keyform KEY_FORMAT] [--pass PASSPHRASE]\n" +
				"              [--json_out JSON_FILE] [--targets_in TARGET_FILE] [--quiet]\n" +
				"              [--slow_connection] [--https_tunnel PROXY_SETTINGS]\n" +
				"              [--starttls PROTOCOL] [--xmpp_to HOSTNAME]\n" +
				"              [--sni SERVER_NAME_INDICATION]\n" +
				"              [--mozilla_config {modern,intermediate,old,disable}]\n" +
				"              [target ...]\n" +
				"\n" +
				"SSLyze version 5.1.3\n" +
				"\n" +
				"positional arguments:\n" +
				"  target\n" +
				"\n" +
				"options:\n" +
				"  -h, --help            show this help message and exit\n" +
				"  --update_trust_stores\n" +
				"  --cert CERTIFICATE_FILE\n" +
				"  --key KEY_FILE\n" +
				"  --keyform KEY_FORMAT\n" +
				"  --pass PASSPHRASE\n" +
				"  --json_out JSON_FILE\n" +
				"  --targets_in TARGET_FILE\n" +
				"  --quiet\n" +
				"  --slow_connection\n" +
				"  --https_tunnel PROXY_SETTINGS\n" +
				"  --starttls PROTOCOL\n" +
				"  --xmpp_to HOSTNAME\n" +
				"  --sni SERVER_NAME_INDICATION\n" +
				"  --mozilla_config {modern,intermediate,old,disable}\n",
			want: "5.1.3",
		},
		{
			// Reconstructed from sslyze 5.2.1 source; same argparse layout as 5.1.3.
			name: "sslyze-5.2.1-help",
			output: "usage: sslyze [-h] [--update_trust_stores] [--cert CERTIFICATE_FILE]\n" +
				"              [--key KEY_FILE] [--keyform KEY_FORMAT] [--pass PASSPHRASE]\n" +
				"              [--json_out JSON_FILE] [--targets_in TARGET_FILE] [--quiet]\n" +
				"              [--slow_connection] [--https_tunnel PROXY_SETTINGS]\n" +
				"              [--starttls PROTOCOL] [--xmpp_to HOSTNAME]\n" +
				"              [--sni SERVER_NAME_INDICATION]\n" +
				"              [--mozilla_config {modern,intermediate,old,disable}]\n" +
				"              [target ...]\n" +
				"\n" +
				"SSLyze version 5.2.1\n" +
				"\n" +
				"positional arguments:\n" +
				"  target\n" +
				"\n" +
				"options:\n" +
				"  -h, --help            show this help message and exit\n" +
				"  --update_trust_stores\n" +
				"  --cert CERTIFICATE_FILE\n" +
				"  --key KEY_FILE\n" +
				"  --keyform KEY_FORMAT\n" +
				"  --pass PASSPHRASE\n" +
				"  --json_out JSON_FILE\n" +
				"  --targets_in TARGET_FILE\n" +
				"  --quiet\n" +
				"  --slow_connection\n" +
				"  --https_tunnel PROXY_SETTINGS\n" +
				"  --starttls PROTOCOL\n" +
				"  --xmpp_to HOSTNAME\n" +
				"  --sni SERVER_NAME_INDICATION\n" +
				"  --mozilla_config {modern,intermediate,old,disable}\n",
			want: "5.2.1",
		},
		{
			// Reconstructed from sslyze 6.0.0 source (tag v6.0.0); CLI structure identical to 5.1.3
			// for version-detection purposes. --version still absent in 6.x.
			name: "sslyze-6.0.0-help",
			output: "usage: sslyze [-h] [--update_trust_stores] [--cert CERTIFICATE_FILE]\n" +
				"              [--key KEY_FILE] [--keyform KEY_FORMAT] [--pass PASSPHRASE]\n" +
				"              [--json_out JSON_FILE] [--targets_in TARGET_FILE] [--quiet]\n" +
				"              [--slow_connection] [--https_tunnel PROXY_SETTINGS]\n" +
				"              [--starttls PROTOCOL] [--xmpp_to HOSTNAME]\n" +
				"              [--sni SERVER_NAME_INDICATION]\n" +
				"              [--mozilla_config {modern,intermediate,old,disable}]\n" +
				"              [target ...]\n" +
				"\n" +
				"SSLyze version 6.0.0\n" +
				"\n" +
				"positional arguments:\n" +
				"  target\n" +
				"\n" +
				"options:\n" +
				"  -h, --help            show this help message and exit\n" +
				"  --update_trust_stores\n" +
				"  --cert CERTIFICATE_FILE\n" +
				"  --key KEY_FILE\n" +
				"  --keyform KEY_FORMAT\n" +
				"  --pass PASSPHRASE\n" +
				"  --json_out JSON_FILE\n" +
				"  --targets_in TARGET_FILE\n" +
				"  --quiet\n" +
				"  --slow_connection\n" +
				"  --https_tunnel PROXY_SETTINGS\n" +
				"  --starttls PROTOCOL\n" +
				"  --xmpp_to HOSTNAME\n" +
				"  --sni SERVER_NAME_INDICATION\n" +
				"  --mozilla_config {modern,intermediate,old,disable}\n",
			want: "6.0.0",
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

// TestParsePythonVersion verifies that the Python version string is correctly extracted from canned output.
func TestParsePythonVersion(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "python-3x-standard",
			output: "Python 3.11.2\n",
			want:   "3.11.2",
		},
		{
			name:   "python-3x-no-trailing-newline",
			output: "Python 3.9.7",
			want:   "3.9.7",
		},
		{
			name:   "python-2x",
			output: "Python 2.7.18\n",
			want:   "2.7.18",
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
			if got := parsePythonVersion(tt.output); got != tt.want {
				t.Errorf("parsePythonVersion() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
