/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package nmapargs

import "fmt"

// rule inspects the full set of parsed flags and returns any logical
// conflict messages it finds. An empty slice means no conflict was detected.
type rule func(flags map[string]bool) []string

var allRules = []rule{

	// Only one TCP scan type may be used at a time.
	func(flags map[string]bool) []string {

		// Collect every TCP scan-type flag that was seen.
		var found []string
		for flag := range flags {
			if tcpPortScanFlags[flag] {
				found = append(found, flag)
			}
		}

		// More than one TCP scan type in the same invocation is an error.
		if len(found) > 1 {
			return []string{fmt.Sprintf(
				"only one TCP scan type may be used at a time, but got: %v", found,
			)}
		}

		// Return nil as everything went fine.
		return nil
	},

	// -sn / -sP / -sL disable port scanning and conflict with port-dependent flags.
	func(flags map[string]bool) []string {

		// Check whether any no-port-scan flag was used.
		noScan := false
		var noScanFlag string
		for flag := range flags {
			if noPortScanFlags[flag] {
				noScan = true
				noScanFlag = flag
				break
			}
		}

		// If no port-scan-disabling flag was found, there is nothing to check.
		if !noScan {
			return nil
		}

		// List all flags that are incompatible with disabled port scanning.
		incompatible := []string{
			"-p", "--top-ports", "--port-ratio", "--exclude-ports",
			"-sV", "--version-intensity", "--version-light", "--version-all", "--version-trace",
			"-O", "--osscan-limit", "--osscan-guess", "--fuzzy",
			"-A",
			"-sC", "--script", "--script-args", "--script-args-file",
			"--scanflags",
			"-sS", "-sT", "-sU", "-sA", "-sW", "-sM", "-sN", "-sF", "-sX", "-sI", "-sY", "-sZ", "-sO",
		}

		// Report every incompatible flag that was also present.
		var errs []string
		for _, bad := range incompatible {
			if flags[bad] {
				errs = append(errs, fmt.Sprintf(
					"%s disables port scanning, so %s has no effect and is contradictory",
					noScanFlag, bad,
				))
			}
		}

		// Return reported errors.
		return errs
	},

	// -sO (IP protocol scan) conflicts with TCP scan types.
	func(flags map[string]bool) []string {

		// Skip this rule if -sO was not used.
		if !flags["-sO"] {
			return nil
		}

		// List all TCP scan types that conflict with IP protocol scan.
		conflicting := []string{"-sS", "-sT", "-sA", "-sW", "-sM", "-sN", "-sF", "-sX", "-sI"}

		// Report every conflicting TCP scan type that was also present.
		var errs []string
		for _, flag := range conflicting {
			if flags[flag] {
				errs = append(errs, fmt.Sprintf(
					"-sO (IP protocol scan) cannot be combined with %s", flag,
				))
			}
		}

		// Return reported errors.
		return errs
	},

	// -b (FTP bounce scan) is exclusive with all other scan types.
	func(flags map[string]bool) []string {

		// Skip this rule if -b was not used.
		if !flags["-b"] {
			return nil
		}

		// List all scan types that conflict with FTP bounce scan.
		others := []string{
			"-sS", "-sT", "-sU", "-sA", "-sW", "-sM",
			"-sN", "-sF", "-sX", "-sI", "-sY", "-sZ", "-sO",
		}

		// Report every conflicting scan type that was also present.
		var errs []string
		for _, flag := range others {
			if flags[flag] {
				errs = append(errs, fmt.Sprintf(
					"-b (FTP bounce scan) cannot be combined with %s", flag,
				))
			}
		}

		// Return reported errors.
		return errs
	},

	// -4 -6 are mutually exclusive.
	func(flags map[string]bool) []string {

		// These flags assert opposite privilege states.
		if flags["-4"] && flags["-6"] {
			return []string{"-4 and -6 are mutually exclusive"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// --send-eth and --send-ip are mutually exclusive layer-2/layer-3 send modes.
	func(flags map[string]bool) []string {

		// Only one packet sending method can be active at a time.
		if flags["--send-eth"] && flags["--send-ip"] {
			return []string{"--send-eth and --send-ip are mutually exclusive"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// --privileged and --unprivileged are mutually exclusive privilege modes.
	func(flags map[string]bool) []string {

		// Only one privilege assumption can be active at a time.
		if flags["--privileged"] && flags["--unprivileged"] {
			return []string{"--privileged and --unprivileged are mutually exclusive"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// --exclude and --excludefile are mutually exclusive target exclusion methods.
	func(flags map[string]bool) []string {

		// Nmap only accepts one exclusion source at a time.
		if flags["--exclude"] && flags["--excludefile"] {
			return []string{"--exclude and --excludefile are mutually exclusive"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// Version detection modifiers require -sV or -A to be meaningful.
	func(flags map[string]bool) []string {

		// These flags tune version detection but have no effect without it enabled.
		versionModifiers := []string{"--version-intensity", "--version-light", "--version-all", "--version-trace"}

		// Skip this rule if version detection is already enabled.
		if flags["-sV"] || flags["-A"] {
			return nil
		}

		// Report every version modifier that was used without version detection.
		var errs []string
		for _, mod := range versionModifiers {
			if flags[mod] {
				errs = append(errs, fmt.Sprintf("%s requires -sV or -A (version detection)", mod))
			}
		}

		// Return reported errors.
		return errs
	},

	// -f and --mtu are contradictory because --mtu silently overrides -f.
	func(flags map[string]bool) []string {

		// Both fragment flags set the same internal MTU value; using both is contradictory.
		if flags["-f"] && flags["--mtu"] {
			return []string{"-f and --mtu are contradictory — --mtu overrides the fragment size set by -f"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// OS detection modifiers require -O or -A to be meaningful.
	func(flags map[string]bool) []string {

		// These flags tune OS detection but have no effect without it enabled.
		osModifiers := []string{"--max-os-tries", "--osscan-limit", "--osscan-guess", "--fuzzy"}

		// Skip this rule if OS detection is already enabled (-A implies -O).
		if flags["-O"] || flags["-A"] {
			return nil
		}

		// Report every OS modifier that was used without OS detection.
		var errs []string
		for _, mod := range osModifiers {
			if flags[mod] {
				errs = append(errs, fmt.Sprintf("%s requires -O or -A (OS detection)", mod))
			}
		}

		// Return reported errors.
		return errs
	},

	// -p conflicts with --top-ports and --port-ratio, which are alternative port selection methods.
	func(flags map[string]bool) []string {

		// An explicit port list overrides top-port selection; using both is contradictory.
		var errs []string
		if flags["-p"] && flags["--top-ports"] {
			errs = append(errs, "-p and --top-ports are contradictory — use one or the other")
		}
		if flags["-p"] && flags["--port-ratio"] {
			errs = append(errs, "-p and --port-ratio are contradictory — use one or the other")
		}

		// Return reported errors.
		return errs
	},

	// --top-ports and --port-ratio write to the same internal Nmap field, so only one may be used.
	func(flags map[string]bool) []string {

		// Both flags set the same top-port-level value; using both is contradictory.
		if flags["--top-ports"] && flags["--port-ratio"] {
			return []string{"--top-ports and --port-ratio are mutually exclusive"}
		}

		// Return nil as everything went fine.
		return nil
	},

	// --traceroute and -sI (idle scan) cannot be combined.
	func(flags map[string]bool) []string {

		// Nmap does not support traceroute during idle scans.
		if flags["--traceroute"] && flags["-sI"] {
			return []string{"--traceroute cannot be combined with -sI (idle scan)"}
		}

		// Return nil as everything went fine.
		return nil
	},
}
