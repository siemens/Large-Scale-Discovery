/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package nmapargs_test

import "testing"

// TestRules verifies that mutually exclusive or contradictory flag combinations are rejected.
func TestRules(t *testing.T) {

	// Multiple TCP scan types are not allowed in the same invocation.
	expectInvalid(t, "-sS -sT", "only one TCP scan type")
	expectInvalid(t, "-sS -sA", "only one TCP scan type")
	expectInvalid(t, "-sN -sF", "only one TCP scan type")
	expectInvalid(t, "-sS -sT -sA", "only one TCP scan type")

	// -sn disables port scanning, so port-dependent flags are incompatible.
	expectInvalid(t, "-sn -sV", "disables port scanning")
	expectInvalid(t, "-sn -O", "disables port scanning")
	expectInvalid(t, "-sn -A", "disables port scanning")
	expectInvalid(t, "-sn -p 80", "disables port scanning")
	expectInvalid(t, "-sn --script http-title", "disables port scanning")
	expectInvalid(t, "-sn -sS", "disables port scanning")

	// -sO (IP protocol scan) conflicts with TCP/UDP scan types.
	expectInvalid(t, "-sO -sS", "-sO")
	expectInvalid(t, "-sO -sT", "-sO")

	// FTP bounce scan (-b) conflicts with other scan types.
	expectInvalid(t, "-b ftp.example.com -sS", "FTP bounce")

	// Numeric range inversions: minimum must not exceed maximum.
	expectInvalid(t, "--min-rate 500 --max-rate 100", "must not be greater than")
	expectInvalid(t, "--min-hostgroup 50 --max-hostgroup 10", "must not be greater than")
	expectInvalid(t, "--min-parallelism 20 --max-parallelism 5", "must not be greater than")

	// -sP and -sL also disable port scanning, not just -sn.
	expectInvalid(t, "-sP -sV", "disables port scanning")
	expectInvalid(t, "-sP -p 80", "disables port scanning")
	expectInvalid(t, "-sL -p 80", "disables port scanning")
	expectInvalid(t, "-sL -sS", "disables port scanning")

	// -sO conflicts with scan types beyond -sS and -sT.
	expectInvalid(t, "-sO -sA", "-sO")
	expectInvalid(t, "-sO -sW", "-sO")
	expectInvalid(t, "-sO -sN", "-sO")

	// -b (FTP bounce) conflicts with UDP and SCTP scan types.
	expectInvalid(t, "-b ftp.example.com -sU", "FTP bounce")
	expectInvalid(t, "-b ftp.example.com -sY", "FTP bounce")

	// -4 -6  are mutually exclusive.
	expectInvalid(t, "-4 -6", "mutually exclusive")

	// --send-eth and --send-ip are mutually exclusive.
	expectInvalid(t, "--send-eth --send-ip", "mutually exclusive")

	// --privileged and --unprivileged are mutually exclusive.
	expectInvalid(t, "--privileged --unprivileged", "mutually exclusive")

	// --exclude and --excludefile are mutually exclusive.
	expectInvalid(t, "--exclude 10.0.0.1 --excludefile hosts.txt", "mutually exclusive")

	// Version modifiers without -sV or -A are rejected.
	expectInvalid(t, "--version-intensity 5", "requires -sV or -A")
	expectInvalid(t, "--version-light", "requires -sV or -A")
	expectInvalid(t, "--version-all", "requires -sV or -A")
	expectInvalid(t, "--version-trace", "requires -sV or -A")

	// Version modifiers with -A are accepted (happy path).
	expectValid(t, "-A --version-intensity 5")
	expectValid(t, "-A --version-light")

	// OS modifiers without -O or -A are rejected.
	expectInvalid(t, "--max-os-tries 3", "requires -O or -A")
	expectInvalid(t, "--osscan-limit", "requires -O or -A")
	expectInvalid(t, "--osscan-guess", "requires -O or -A")
	expectInvalid(t, "--fuzzy", "requires -O or -A")

	// OS modifiers with -A are accepted (happy path).
	expectValid(t, "-A --max-os-tries 3")
	expectValid(t, "-A --osscan-limit")

	// Time-duration range inversions are caught across different units.
	expectInvalid(t, "--min-rtt-timeout 5s --max-rtt-timeout 100ms", "must not be greater than")
	expectInvalid(t, "--scan-delay 2s --max-scan-delay 500ms", "must not be greater than")

	// Time-duration ranges with correct ordering are accepted.
	expectValid(t, "--min-rtt-timeout 100ms --max-rtt-timeout 5s")
	expectValid(t, "--scan-delay 500ms --max-scan-delay 2s")

	// -f and --mtu are contradictory.
	expectInvalid(t, "-f --mtu 24", "contradictory")

	// --initial-rtt-timeout must not exceed --max-rtt-timeout.
	expectInvalid(t, "--initial-rtt-timeout 10s --max-rtt-timeout 1s", "must not be greater than")
	expectValid(t, "--min-rtt-timeout 100ms --initial-rtt-timeout 500ms --max-rtt-timeout 5s")

	// Nmap auto-adjusts --min-rtt-timeout vs --initial-rtt-timeout, so ordering is not enforced.
	expectValid(t, "--min-rtt-timeout 500ms --initial-rtt-timeout 100ms")

	// --scanflags is useless without a port scan and should be rejected with -sn/-sP/-sL.
	expectInvalid(t, "-sn --scanflags SYN", "disables port scanning")

	// Duplicate -p is rejected (Nmap only allows one -p per invocation).
	expectInvalid(t, "-p 22 -p 80", "only be specified once")
	expectInvalid(t, "--exclude-ports 80 --exclude-ports 443", "only be specified once")

	// -p conflicts with --top-ports and --port-ratio.
	expectInvalid(t, "-p 80 --top-ports 100", "contradictory")
	expectInvalid(t, "-p 80 --port-ratio 0.5", "contradictory")

	// --top-ports and --port-ratio are mutually exclusive.
	expectInvalid(t, "--top-ports 100 --port-ratio 0.5", "mutually exclusive")

	// --traceroute and -sI (idle scan) cannot be combined.
	expectInvalid(t, "--traceroute -sI zombie.example.com", "idle scan")
}
