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

import (
	"strings"
	"testing"

	"github.com/siemens/Large-Scale-Discovery/utils/nmapargs"
)

// expectValid is a test helper that asserts the given Nmap argument string is accepted as valid.
func expectValid(t *testing.T, args string) {

	// Run the validator and check for unexpected errors.
	t.Helper()
	r := nmapargs.Validate(args)
	if !r.Valid {
		t.Errorf("expected VALID for %q, got errors:\n  %s", args, strings.Join(r.Errors, "\n  "))
	}
}

// expectInvalid is a test helper that asserts the given Nmap argument string is rejected,
// with at least one error message containing wantErrContains (case-insensitive).
func expectInvalid(t *testing.T, args, wantErrContains string) {

	// Run the validator and check that it rejects the input.
	t.Helper()
	r := nmapargs.Validate(args)
	if r.Valid {
		t.Errorf("expected INVALID for %q, but got Valid=true", args)
		return
	}

	// Search for the expected substring in the reported errors.
	for _, e := range r.Errors {
		if strings.Contains(strings.ToLower(e), strings.ToLower(wantErrContains)) {
			return
		}
	}
	t.Errorf("expected error containing %q for %q, got:\n  %s",
		wantErrContains, args, strings.Join(r.Errors, "\n  "))
}

// TestValid exercises the happy path for a wide range of valid Nmap invocations.
func TestValid(t *testing.T) {

	// Each case represents a valid Nmap argument string that must pass validation.
	cases := []struct{ name, args string }{
		{"syn scan with ports 1", "-sS -p 22,80,443"},
		{"syn scan with ports 2", "-sS -p22,80,443"},
		{"syn scan with ports range", "-sS -p0-65535"},
		{"syn scan with all ports", "-sS -p-"},
		{"udp + syn", "-sU -sS -p U:53,T:80"},
		{"version detection", "-sV --version-intensity 5"},
		{"version light", "-sV --version-light"},
		{"version all", "-sV --version-all"},
		{"os detection", "-O"},
		{"os with osscan-guess", "-O --osscan-guess"},
		{"os with osscan-limit", "-O --osscan-limit"},
		{"aggressive", "-A"},
		{"aggressive + osscan-guess", "-A --osscan-guess"},
		{"timing glued", "-T4"},
		{"timing separate", "-T 3"},
		{"host timeout", "--host-timeout 30m"},
		{"rtt range valid", "--min-rtt-timeout 100ms --max-rtt-timeout 5s"},
		{"rate range valid", "--min-rate 10 --max-rate 500"},
		{"rate fractional", "--min-rate 0.5"},
		{"parallelism range valid", "--min-parallelism 5 --max-parallelism 50"},
		{"parallelism zero auto", "--min-parallelism 0"},
		{"top ports", "--top-ports 100"},
		{"port ratio", "--port-ratio 0.5"},
		{"fast mode", "-F"},
		{"no ping", "-Pn"},
		{"list scan", "-sL"},
		{"ping scan", "-sn"},
		{"PS with ports", "-PS22,443"},
		{"PS without ports", "-PS"},
		{"PA with ports", "-PA80"},
		{"decoys RND", "-D RND:5"},
		{"decoys explicit with ME", "-D 192.168.1.1,ME"},
		{"source port long", "--source-port 53"},
		{"source port short", "-g 80"},
		{"legacy pn", "-PN"},
		{"legacy p0", "-P0"},
		{"spoof mac zero", "--spoof-mac 0"},
		{"spoof mac address", "--spoof-mac 00:11:22:33:44:55"},
		{"spoof mac vendor", "--spoof-mac Apple"},
		{"mtu valid", "--mtu 24"},
		{"mtu 8", "--mtu 8"},
		{"badsum", "--badsum"},
		{"defeat rst ratelimit", "--defeat-rst-ratelimit"},
		{"no dns", "-n"},
		{"always resolve", "-R"},
		{"ipv6", "-6"},
		{"fragment", "-f"},
		{"double fragment", "-ff"},
		{"triple fragment", "-fff"},
		{"send eth", "--send-eth"},
		{"send ip", "--send-ip"},
		{"privileged", "--privileged"},
		{"unprivileged", "--unprivileged"},
		{"output normal", "-oN /tmp/out.txt"},
		{"output xml", "-oX result.xml"},
		{"output all", "-oA scan"},
		{"verbosity", "-v"},
		{"verbosity level", "-v2"},
		{"verbosity stacked", "-vvv"},
		{"debug", "-d"},
		{"debug level", "-d3"},
		{"debug stacked", "-ddd"},
		{"script default", "-sC"},
		{"script named", "--script http-title"},
		{"script with args", `--script-args "user=admin,pass=test"`},
		{"scan delay", "--scan-delay 500ms"},
		{"max retries", "--max-retries 3"},
		{"exclude ports", "--exclude-ports 8080,9090"},
		{"scanflags symbolic", "--scanflags SYNFIN"},
		{"scanflags numeric", "--scanflags 6"},
		{"traceroute", "--traceroute"},
		{"dns servers", "--dns-servers 8.8.8.8,1.1.1.1"},
		{"system dns", "--system-dns"},
		{"open only", "--open"},
		{"reason", "--reason"},
		{"stats every", "--stats-every 10s"},
		{"sctp init", "-sY"},
		{"sctp cookie", "-sZ"},
		{"sctp + syn", "-sY -sS"},
		{"udp + sctp", "-sU -sY"},
		{"exclude hosts", "--exclude 192.168.1.1,10.0.0.0/8"},
		{"inline long flag value", "--top-ports=200"},
		{"version-intensity inline", "-sV --version-intensity=7"},
		{"deprecated sR", "-sR"},
		{"deprecated PB", "-PB"},
		{"deprecated PI", "-PI"},
		{"deprecated PD", "-PD"},
		{"deprecated PT", "-PT"},
		{"deprecated PT ports", "-PT80"},
		{"deprecated I", "-I"},
		{"deprecated M", "-M 10"},
		{"unique", "--unique"},
		{"discovery ignore rst", "--discovery-ignore-rst"},
		{"allports", "--allports"},
		{"nogcc", "--nogcc"},
		{"proxy singular", "--proxy http://proxy:8080"},
		{"route dst", "--route-dst 10.0.0.1"},
		{"deprecated xml osclass", "--deprecated-xml-osclass"},
		{"long vv", "--vv"},
		{"long ff", "--ff"},
		{"rH alias", "--rH"},
		{"empty string", ""},
		{"multiple spaces between flags", "-sS    -p   22"},
		{"tabs between flags", "-sS\t-p\t80"},
		{"duplicate flag", "-sS -sS -p 80"},
		{"data-length zero", "--data-length 0"},
		{"port-ratio zero", "--port-ratio 0"},
		{"backslash escaped space in value", `-oN file\ name`},
		{"quoted value preserving spaces", `--script-args "user=admin pass=test"`},
	}

	// Run each case as a subtest.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { expectValid(t, tc.args) })
	}
}

// TestUnknownFlags verifies that unrecognized flags are rejected.
func TestUnknownFlags(t *testing.T) {

	// Each of these flags does not exist in the Nmap flag registry.
	expectInvalid(t, "--foobar", "unknown flag")
	expectInvalid(t, "--resolve-al", "unknown flag")
	expectInvalid(t, "-sQ", "unknown flag")
	expectInvalid(t, "--scan-type=X", "unknown flag")
}

// TestMissingArgs verifies that flags requiring an argument are rejected when the argument is absent.
func TestMissingArgs(t *testing.T) {

	// Each flag below requires an argument but none is provided.
	expectInvalid(t, "--top-ports", "requires an argument")
	expectInvalid(t, "-p", "requires an argument")
	expectInvalid(t, "--host-timeout", "requires an argument")
	expectInvalid(t, "--min-rate", "requires an argument")
	expectInvalid(t, "--version-intensity", "requires an argument")
	expectInvalid(t, "-T", "requires an argument")
	expectInvalid(t, "-oN", "requires an argument")
	expectInvalid(t, "--dns-servers", "requires an argument")
}

// TestNoArgFlagWithValue verifies that no-argument flags reject inline values via '='.
func TestNoArgFlagWithValue(t *testing.T) {

	// These flags accept no argument, but an inline value was provided.
	expectInvalid(t, "--badsum=yes", "takes no argument")
	expectInvalid(t, "--open=true", "takes no argument")
	expectInvalid(t, "--traceroute=1", "takes no argument")
}

// expectBlockedInvalid is a test helper that asserts an Nmap argument string is rejected
// due to a blocked flag when validation runs with the given blocked list.
func expectBlockedInvalid(t *testing.T, args string, blocked []string, wantErrContains string) {

	// Run the validator with a blocked flag list and check that it rejects the input.
	t.Helper()
	r := nmapargs.Validate(args, nmapargs.WithBlockFlags(blocked))
	if r.Valid {
		t.Errorf("expected INVALID for %q with blocked %v, but got Valid=true", args, blocked)
		return
	}

	// Search for the expected substring in the reported errors.
	for _, e := range r.Errors {
		if strings.Contains(strings.ToLower(e), strings.ToLower(wantErrContains)) {
			return
		}
	}
	t.Errorf("expected error containing %q for %q, got:\n  %s",
		wantErrContains, args, strings.Join(r.Errors, "\n  "))
}

// TestBlockedFlags verifies that the blocked-flag mechanism rejects flags on the blocklist.
func TestBlockedFlags(t *testing.T) {

	// Long flags on the blocklist are rejected.
	expectBlockedInvalid(t, "--script http-title", nmapargs.BlockFlags, "blocked by policy")
	expectBlockedInvalid(t, "--resume scan.gnmap", nmapargs.BlockFlags, "blocked by policy")

	// Short flags on the blocklist are rejected.
	expectBlockedInvalid(t, "-iL targets.txt", nmapargs.BlockFlags, "blocked by policy")
	expectBlockedInvalid(t, "-oN output.txt", nmapargs.BlockFlags, "blocked by policy")
	expectBlockedInvalid(t, "-S 10.0.0.1", nmapargs.BlockFlags, "blocked by policy")
	expectBlockedInvalid(t, "-D RND:5", nmapargs.BlockFlags, "blocked by policy")

	// Deprecated aliases on the blocklist are also rejected.
	expectBlockedInvalid(t, "-i targets.txt", nmapargs.BlockFlags, "blocked by policy")
	expectBlockedInvalid(t, "-m output.gnmap", nmapargs.BlockFlags, "blocked by policy")

	// A blocked flag's argument must not cause a spurious bare-word error.
	r := nmapargs.Validate("-iL targets.txt", nmapargs.WithBlockFlags(nmapargs.BlockFlags))
	for _, e := range r.Errors {
		if strings.Contains(e, "targets are specified separately") {
			t.Errorf("blocked flag argument leaked as bare-word error: %s", e)
		}
	}

	// Flags not on the blocklist remain valid.
	r = nmapargs.Validate("-sS -p 80", nmapargs.WithBlockFlags(nmapargs.BlockFlags))
	if !r.Valid {
		t.Errorf("expected VALID for non-blocked flags, got errors:\n  %s", strings.Join(r.Errors, "\n  "))
	}
}

// TestBackwardCompat verifies that Validate still works without options (backward compatible).
func TestBackwardCompat(t *testing.T) {

	// Calling Validate without options must still work correctly.
	r := nmapargs.Validate("-sS -p 80")
	if !r.Valid {
		t.Errorf("expected VALID for backward-compat call, got errors:\n  %s", strings.Join(r.Errors, "\n  "))
	}

	// Flags that would be blocked with BlockFlags are accepted without options.
	r = nmapargs.Validate("--script http-title")
	if !r.Valid {
		t.Errorf("expected VALID without blocklist, got errors:\n  %s", strings.Join(r.Errors, "\n  "))
	}
}
