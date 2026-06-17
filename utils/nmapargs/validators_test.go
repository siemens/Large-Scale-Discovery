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

// TestBadArgValues verifies that invalid argument values are caught by the per-flag validators.
func TestBadArgValues(t *testing.T) {

	// Each case passes a syntactically correct flag with a semantically invalid value.
	expectInvalid(t, "-sS -p0-65536", "out of range")
	expectInvalid(t, "-T 9", "0–5")
	expectInvalid(t, "-T 6", "0–5")
	expectInvalid(t, "--version-intensity 10", "at most 9")
	expectInvalid(t, "--version-intensity -1", "at least 0")
	expectInvalid(t, "--mtu 10", "multiple of 8")
	expectInvalid(t, "--mtu abc", "positive integer")
	expectInvalid(t, "--host-timeout 30x", "time value")
	expectInvalid(t, "--min-rtt-timeout 5sec", "time value")
	expectInvalid(t, "-p abc!!", "invalid port list")
	expectInvalid(t, "--top-ports abc", "expected an integer")
	expectInvalid(t, "--max-retries -1", "at least 0")
	expectInvalid(t, "--port-ratio 1.5", "[0, 1)")
	expectInvalid(t, "--port-ratio abc", "[0, 1)")
	expectInvalid(t, "--scanflags INVALID", "URG/ACK")
	expectInvalid(t, "-D !!!bad!!!", "invalid decoy list")
	expectInvalid(t, "--spoof-mac gg:hh:ii:jj:kk:ll", "MAC address")
	expectInvalid(t, "192.168.1.0/24 -sS -p 80", "targets are specified separately")
}

// TestPortListEdgeCases exercises unusual port list inputs that the validator must catch.
func TestPortListEdgeCases(t *testing.T) {

	// Unknown protocol prefix (X: is not a valid Nmap protocol prefix).
	expectInvalid(t, "-p X:80", "unknown protocol prefix")

	// Protocol prefix with no port number following it.
	expectInvalid(t, "-p T:", "protocol prefix with no port")

	// Trailing comma produces an empty entry.
	expectInvalid(t, "-p 80,", "empty entry")

	// Double comma produces an empty entry between valid ports.
	expectInvalid(t, "-p 80,,443", "empty entry")

	// Inverted port range where start exceeds end.
	expectInvalid(t, "-p 443-80", "start must not exceed end")

	// Port ratio of exactly 1.0 is outside the valid [0, 1) range.
	expectInvalid(t, "--port-ratio 1", "[0, 1)")

	// Negative data-length is rejected.
	expectInvalid(t, "--data-length -5", "at least 0")

	// Fractional data-length is rejected — Nmap only accepts integers.
	expectInvalid(t, "--data-length 3.5", "integer")

	// MTU of zero is rejected — must be a positive integer.
	expectInvalid(t, "--mtu 0", "positive integer")

	// --max-rtt-timeout and --initial-rtt-timeout reject zero (Nmap fatals: "must be greater than 0").
	expectInvalid(t, "--max-rtt-timeout 0", "greater than 0")
	expectInvalid(t, "--max-rtt-timeout 0ms", "greater than 0")
	expectInvalid(t, "--initial-rtt-timeout 0", "greater than 0")

	// --min-rtt-timeout 0 is valid (Nmap allows it).
	expectValid(t, "--min-rtt-timeout 0")
}

// TestOptionalArgBadValue verifies that optional-argument flags reject invalid glued values.
func TestOptionalArgBadValue(t *testing.T) {

	// -PS accepts an optional port list, so an invalid one must be caught.
	expectInvalid(t, "-PSabc!!", "invalid port list")

	// -PA with an invalid glued port list.
	expectInvalid(t, "-PAabc!!", "invalid port list")
}

// TestScanflagsRange verifies that numeric --scanflags values are validated against the 0–255 range.
func TestScanflagsRange(t *testing.T) {

	// Valid boundary values must be accepted.
	expectValid(t, "--scanflags 0")
	expectValid(t, "--scanflags 255")
	expectValid(t, "--scanflags 128")

	// Values outside the 0–255 range must be rejected.
	expectInvalid(t, "--scanflags 256", "0–255")
	expectInvalid(t, "--scanflags 999", "0–255")
}

// TestVerbosityLevels verifies that -v accepts optional level suffixes.
func TestVerbosityLevels(t *testing.T) {

	// Valid verbosity specifications must be accepted.
	expectValid(t, "-v")
	expectValid(t, "-v2")
	expectValid(t, "-vv")
	expectValid(t, "-vvv")

	// Invalid verbosity suffixes must be rejected.
	expectInvalid(t, "-vx", "verbosity level")
}

// TestDebugLevels verifies that -d accepts optional level suffixes.
func TestDebugLevels(t *testing.T) {

	// Valid debug specifications must be accepted.
	expectValid(t, "-d")
	expectValid(t, "-d3")
	expectValid(t, "-dd")
	expectValid(t, "-ddd")

	// Invalid debug suffixes must be rejected.
	expectInvalid(t, "-dx", "debug level")
}

// TestFragmentFlag verifies that -f accepts repeated 'f' characters but nothing else.
func TestFragmentFlag(t *testing.T) {

	// Standard, double, and triple fragmentation must be accepted.
	expectValid(t, "-f")
	expectValid(t, "-ff")
	expectValid(t, "-fff")

	// Invalid trailing characters must be rejected.
	expectInvalid(t, "-fg", "repeated")
}

// TestSourcePort verifies that -g and --source-port accept only single port numbers.
func TestSourcePort(t *testing.T) {

	// Valid single port numbers must be accepted.
	expectValid(t, "-g 53")
	expectValid(t, "--source-port 80")

	// Port lists and ranges must be rejected.
	expectInvalid(t, "-g 53,80", "single port")
	expectInvalid(t, "--source-port 1-1024", "single port")
}

// TestLegacyAliases verifies that legacy Nmap flag aliases are accepted.
func TestLegacyAliases(t *testing.T) {

	// -PN and -P0 are legacy aliases for -Pn.
	expectValid(t, "-PN")
	expectValid(t, "-P0")
}

// TestZombieHost verifies that -sI zombie host specifications are validated correctly.
func TestZombieHost(t *testing.T) {

	// Valid zombie host specifications must be accepted.
	expectValid(t, "-sI 192.168.1.1")
	expectValid(t, "-sI zombie.example.com")
	expectValid(t, "-sI 192.168.1.1:80")
	expectValid(t, "-sI zombie.example.com:8080")
	expectValid(t, "-sI [::1]")

	// Invalid zombie host specifications must be rejected.
	expectInvalid(t, "-sI 192.168.1.1:99999", "probe port")
	expectInvalid(t, "-sI host!name", "zombie host")
}
