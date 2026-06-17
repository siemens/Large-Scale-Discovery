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

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reTime = regexp.MustCompile(`^\d+(\.\d+)?(ms|s|m|h)?$`) // Matches Nmap time expressions like 100ms, 3s, 1m, 0.5h

// rePortList is a character-set guard for port list strings.
// Includes P/p for IP-protocol qualifier, a-z/A-Z for service names,
// * and ? for wildcards, and [ ] for bracket-filter notation.
var rePortList = regexp.MustCompile(`^[TUSPtusp:0-9,\-a-zA-Z*?\[\]]+$`)

var reDecoys = regexp.MustCompile( // Matches comma-separated decoy specs (IPs, RND[:n], ME)
	`^(RND(:\d+)?|ME|[a-zA-Z0-9.\-:/\[\]]+)(,(RND(:\d+)?|ME|[a-zA-Z0-9.\-:/\[\]]+))*$`,
)
var reScanFlag = regexp.MustCompile(`^\d+$|^(URG|ACK|PSH|RST|SYN|FIN|ECE|CWR)+$`) // Matches symbolic or numeric TCP flags
var reMac = regexp.MustCompile(                                                   // Matches MAC address, 0, or vendor prefix
	`^([0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}|0|[a-zA-Z][a-zA-Z0-9_\-]*)$`,
)

// validatePortNumber checks that a single port string is a valid number in the 0–65535 range.
func validatePortNumber(s string) error {

	// Parse the port string as an integer and validate its range.
	n, errConv := strconv.Atoi(s)
	if errConv != nil || n < 0 || n > 65535 {
		return fmt.Errorf("port %q is out of range (must be 0–65535)", s)
	}

	// Return nil as everything went fine.
	return nil
}

// validatePortList validates a full Nmap port list expression.
// Supports individual ports, ranges (80-443), and protocol-prefixed entries (T:80, U:53).
func validatePortList(s string) error {

	// Reject an empty port list immediately.
	if s == "" {
		return fmt.Errorf("port list must not be empty")
	}

	// Quick character-set guard — reject obviously malformed input early.
	if !rePortList.MatchString(s) {
		return fmt.Errorf("invalid port list %q — expected e.g. 22,80,443 or T:80,U:53 or 1-1024", s)
	}

	// Split on commas and validate each entry individually.
	entries := strings.Split(s, ",")
	for _, entry := range entries {

		// A trailing or double comma produces an empty entry.
		if entry == "" {
			return fmt.Errorf("port list %q contains an empty entry (trailing or double comma?)", s)
		}

		// Strip an optional protocol prefix (T:, U:, S:, P:).
		if len(entry) >= 2 && entry[1] == ':' {
			proto := entry[0]

			// Only the four recognised protocol letters are valid prefixes:
			// T (TCP), U (UDP), S (SCTP), P (IP protocol number).
			if proto != 'T' && proto != 'U' && proto != 'S' && proto != 'P' &&
				proto != 't' && proto != 'u' && proto != 's' && proto != 'p' {
				return fmt.Errorf("unknown protocol prefix %q in port list", string(proto))
			}

			// Advance past the "X:" prefix.
			entry = entry[2:]

			// A prefix with no port following it is invalid.
			if entry == "" {
				return fmt.Errorf("protocol prefix with no port in %q", s)
			}
		}

		// A bare "-" is nmap's shorthand for the full port range (1–65535).
		if entry == "-" {
			continue
		}

		// Detect a numeric port range: both sides of the first "-" must be all digits.
		// This prevents misinterpreting service names like "http-alt" as ranges.
		if idx := strings.Index(entry, "-"); idx != -1 {
			lo, hi := entry[:idx], entry[idx+1:]
			loIsDigit := lo != "" && isAllDigits(lo)
			hiIsDigit := hi != "" && isAllDigits(hi)

			if loIsDigit || hiIsDigit {
				// Treat as a numeric range — both ends must be present and valid.
				if lo == "" || hi == "" {
					return fmt.Errorf("invalid port range %q — both ends must be specified", entry)
				}

				if !loIsDigit {
					return fmt.Errorf("invalid port range %q — start must be a number", entry)
				}

				if !hiIsDigit {
					return fmt.Errorf("invalid port range %q — end must be a number", entry)
				}

				// Validate each endpoint is within 0–65535.
				if err := validatePortNumber(lo); err != nil {
					return err
				}

				if err := validatePortNumber(hi); err != nil {
					return err
				}

				// Parse both ends for the ordering check (errors already handled above).
				loN, _ := strconv.Atoi(lo)
				hiN, _ := strconv.Atoi(hi)

				// The start of a range must not exceed its end.
				if loN > hiN {
					return fmt.Errorf("port range %q is invalid — start must not exceed end", entry)
				}

				continue
			}
		}

		// If the entry is all digits, validate it as a single port number.
		if isAllDigits(entry) {
			if err := validatePortNumber(entry); err != nil {
				return err
			}
			continue
		}

		// Otherwise treat as a service name or wildcard (e.g. "ssh", "http*", "http-alt").
		// nmap resolves these against nmap-services; syntactically any non-empty string
		// composed of the allowed characters is accepted here.
	}

	// Return nil as everything went fine.
	return nil
}

// isAllDigits reports whether s is non-empty and contains only ASCII digits.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// validateTime validates an Nmap time expression (e.g. 100ms, 3s, 1m, 0.5h).
func validateTime(s string) error {

	// Match against the time pattern.
	if !reTime.MatchString(s) {
		return fmt.Errorf("expected a time value like 100ms / 3s / 1m / 0.5h, got %q", s)
	}

	// Return nil as everything went fine.
	return nil
}

// validateTimePositive validates an Nmap time expression and rejects zero values.
// Nmap fatals on --max-rtt-timeout 0 and --initial-rtt-timeout 0 ("must be greater than 0").
func validateTimePositive(s string) error {

	// First validate the time format itself.
	if errTime := validateTime(s); errTime != nil {
		return errTime
	}

	// Parse the duration and reject zero.
	ms, errParse := parseTimeDuration(s)
	if errParse != nil {
		return errParse
	}

	// A zero duration is not allowed for this flag.
	if ms <= 0 {
		return fmt.Errorf("must be greater than 0, got %q", s)
	}

	// Return nil as everything went fine.
	return nil
}

// validateIntRange returns a validator that checks s is an integer within [min, max].
// Pass math.MinInt / math.MaxInt to leave a bound unconstrained.
func validateIntRange(min, max int) func(string) error {
	return func(s string) error {

		// Parse the value as a signed integer.
		n, errConv := strconv.Atoi(s)
		if errConv != nil {
			return fmt.Errorf("expected an integer, got %q", s)
		}

		// Check the lower bound if one was requested.
		if n < min {
			return fmt.Errorf("must be at least %d, got %d", min, n)
		}

		// Check the upper bound if one was requested.
		if n > max {
			return fmt.Errorf("must be at most %d, got %d", max, n)
		}

		// Return nil as everything went fine.
		return nil
	}
}

// validateFloatRange returns a validator that checks s is a float within [min, max].
// Pass -math.MaxFloat64 / math.MaxFloat64 to leave a bound unconstrained.
func validateFloatRange(min, max float64) func(string) error {
	return func(s string) error {

		// Parse the value as a 64-bit float.
		f, errConv := strconv.ParseFloat(s, 64)
		if errConv != nil {
			return fmt.Errorf("expected a number, got %q", s)
		}

		// Check the lower bound if one was requested.
		if f < min {
			return fmt.Errorf("must be at least %g, got %g", min, f)
		}

		// Check the upper bound if one was requested.
		if f > max {
			return fmt.Errorf("must be at most %g, got %g", max, f)
		}

		// Return nil as everything went fine.
		return nil
	}
}

// timingNames is the set of word names nmap accepts for -T in addition to 0–5.
var timingNames = map[string]bool{
	"paranoid": true, "sneaky": true, "polite": true,
	"normal": true, "aggressive": true, "insane": true,
}

// validateTimingTemplate validates that the argument is a digit 0–5 or a timing word name.
func validateTimingTemplate(s string) error {

	// Accept a single digit in the range '0'–'5'.
	if len(s) == 1 && s[0] >= '0' && s[0] <= '5' {
		return nil
	}

	// Also accept the named timing levels (case-insensitive, matching nmap behaviour).
	if timingNames[strings.ToLower(s)] {
		return nil
	}

	return fmt.Errorf("timing template must be 0–5 or a name (paranoid/sneaky/polite/normal/aggressive/insane), got %q", s)
}

// validateMtu validates that the string is a positive integer and a multiple of 8.
func validateMtu(s string) error {

	// Parse the MTU value as an integer.
	n, errConv := strconv.Atoi(s)
	if errConv != nil || n <= 0 {
		return fmt.Errorf("MTU must be a positive integer, got %q", s)
	}

	// Nmap requires the MTU to be a multiple of 8.
	if n%8 != 0 {
		return fmt.Errorf("MTU must be a multiple of 8, got %d", n)
	}

	// Return nil as everything went fine.
	return nil
}

// validateScanFlags validates that the string is a valid combination of TCP flag names or a numeric value.
func validateScanFlags(s string) error {

	// Match against symbolic flag names (URG, ACK, etc.) or a plain number.
	if !reScanFlag.MatchString(s) {
		return fmt.Errorf(
			"--scanflags expects a combination of URG/ACK/PSH/RST/SYN/FIN/ECE/CWR or a numeric value, got %q", s,
		)
	}

	// A numeric value must fit in a single TCP flag byte (0–255).
	if isAllDigits(s) {
		n, errConv := strconv.Atoi(s)
		if errConv != nil || n < 0 || n > 255 {
			return fmt.Errorf("numeric --scanflags value must be 0–255, got %s", s)
		}
	}

	// Return nil as everything went fine.
	return nil
}

// validateDecoys validates a comma-separated decoy specification (IPs, RND[:n], or ME).
func validateDecoys(s string) error {

	// Match against the decoy list pattern.
	if !reDecoys.MatchString(s) {
		return fmt.Errorf("invalid decoy list %q — expected comma-separated IPs, RND[:n], or ME", s)
	}

	// Return nil as everything went fine.
	return nil
}

// validateMac validates a MAC address, the literal "0", or a vendor prefix string.
func validateMac(s string) error {

	// Match against a full MAC address, zero, or an alphanumeric vendor prefix.
	if !reMac.MatchString(s) {
		return fmt.Errorf(
			"--spoof-mac expects a MAC address (xx:xx:xx:xx:xx:xx), 0, or a vendor prefix, got %q", s,
		)
	}

	// Return nil as everything went fine.
	return nil
}

// validatePortRatio validates that the string is a float in the [0, 1) range,
// matching Nmap's requirement that --port-ratio be >= 0 and < 1.
func validatePortRatio(s string) error {

	// Parse the port-ratio value as a float.
	f, errParse := strconv.ParseFloat(s, 64)
	if errParse != nil || f < 0 || f >= 1 {
		return fmt.Errorf("--port-ratio must be a float in [0, 1), got %q", s)
	}

	// Return nil as everything went fine.
	return nil
}

// reHostname matches valid hostname/IP characters: letters, digits, dots, colons (IPv6),
// hyphens, square brackets, and percent (zone IDs).
var reHostname = regexp.MustCompile(`^[a-zA-Z0-9.\-:\[\]%]+$`)

// validateZombieHost validates an idle-scan (-sI) zombie host specification.
// The format is host[:probeport], where host is an IP address or hostname,
// and probeport is an optional port number for the probe.
func validateZombieHost(s string) error {

	// Reject an empty zombie host immediately.
	if s == "" {
		return fmt.Errorf("zombie host must not be empty")
	}

	// Split on the last colon to separate host from optional probeport.
	// Using LastIndex handles IPv6 addresses that contain colons.
	host := s
	port := ""
	if idx := strings.LastIndex(s, ":"); idx != -1 {

		// Only treat the part after the last colon as a port if it looks numeric.
		candidate := s[idx+1:]
		if isAllDigits(candidate) {
			host = s[:idx]
			port = candidate
		}
	}

	// Validate the host part contains only valid hostname/IP characters.
	if !reHostname.MatchString(host) {
		return fmt.Errorf("invalid zombie host %q — expected a hostname or IP address", host)
	}

	// Validate the optional probe port is within range.
	if port != "" {
		if errPort := validatePortNumber(port); errPort != nil {
			return fmt.Errorf("invalid zombie probe port: %s", errPort)
		}
	}

	// Return nil as everything went fine.
	return nil
}

// validateSinglePort validates that the string is a single port number (0–65535).
// Unlike validatePortList, this rejects ranges, comma-separated lists, and protocol prefixes.
func validateSinglePort(s string) error {

	// Reject non-numeric input before parsing.
	if !isAllDigits(s) {
		return fmt.Errorf("expected a single port number, got %q", s)
	}

	// Delegate to the shared port number range check.
	return validatePortNumber(s)
}

// validateFragmentRepeat validates the optional glued value for -f (e.g. -ff, -fff).
// Nmap accumulates 8 bytes of fragment size per -f, so any number of repeated 'f' is valid.
func validateFragmentRepeat(s string) error {

	// Accept any number of repeated 'f' characters (e.g. -ff, -fff, -ffff).
	for _, c := range s {
		if c != 'f' {
			return fmt.Errorf("-f accepts only repeated 'f' characters (e.g. -ff, -fff), got -%s glued", s)
		}
	}

	// Return nil as everything went fine.
	return nil
}

// validateVerbosityLevel validates the optional glued value for -v (e.g. -v2 or -vv).
func validateVerbosityLevel(s string) error {

	// Accept a single digit for explicit level setting.
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return nil
	}

	// Accept repeated 'v' characters from stacked invocations like -vv or -vvv.
	for _, c := range s {
		if c != 'v' {
			return fmt.Errorf("expected a verbosity level (0–9 or repeated v), got %q", s)
		}
	}

	// Return nil as everything went fine.
	return nil
}

// validateDebugLevel validates the optional glued value for -d (e.g. -d3 or -dd).
func validateDebugLevel(s string) error {

	// Accept a single digit for explicit level setting.
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return nil
	}

	// Accept repeated 'd' characters from stacked invocations like -dd or -ddd.
	for _, c := range s {
		if c != 'd' {
			return fmt.Errorf("expected a debug level (0–9 or repeated d), got %q", s)
		}
	}

	// Return nil as everything went fine.
	return nil
}

// validateAny accepts any non-empty value without further validation.
func validateAny(_ string) error { return nil }
