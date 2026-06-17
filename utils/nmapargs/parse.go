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
	"strconv"
	"strings"
)

// isFlag returns true when s looks like an Nmap flag (starts with '-' followed
// by a letter or another '-'). This correctly excludes negative numbers such as -1.
func isFlag(s string) bool {
	if len(s) < 2 {
		return false
	}

	// The two possible numeric flags are matched exactly.
	if s == "-4" || s == "-6" {
		return true
	}

	// Every other flag starts with '-' followed by a letter or a second dash.
	return s[0] == '-' && (s[1] == '-' || (s[1] >= 'A' && s[1] <= 'Z') || (s[1] >= 'a' && s[1] <= 'z'))
}

// matchShortFlag finds the longest matching flag key in shortFlags for the
// given token and returns (flagKey, gluedValue). Returns ("", "") if unknown.
func matchShortFlag(token string) (string, string) {

	// Try three-character keys first (-oN, -oX, -oG, -oA, -oS).
	if len(token) >= 3 {
		key := token[:3]
		if _, ok := shortFlags[key]; ok {
			return key, token[3:]
		}
	}

	// Fall back to two-character keys (-sS, -PS, -T, -p, etc.).
	if len(token) >= 2 {
		key := token[:2]
		if _, ok := shortFlags[key]; ok {
			return key, token[2:]
		}
	}

	// Return empty strings when the token does not match any known short flag.
	return "", ""
}

// consumeArg returns the argument value for a long flag.
// When hasInline is true the value comes from --flag=value syntax.
// Otherwise, the function peeks at the next token in the slice.
func consumeArg(tokens []string, i int, hasInline bool, inlineVal string) (string, bool) {

	// If the value was provided inline via '=', use it directly.
	if hasInline {
		return inlineVal, true
	}

	// Otherwise, peek at the next token if it exists and is not itself a flag.
	if i+1 < len(tokens) && !isFlag(tokens[i+1]) {
		return tokens[i+1], true
	}

	// No argument could be found.
	return "", false
}

// checkNumericRanges validates that every min/max pair satisfies min <= max.
func checkNumericRanges(intVals map[string]uint64, floatVals map[string]float64) []string {

	// pairs lists every (min-flag, max-flag) combination that must be ordered.
	intPairs := [][2]string{
		{"--min-hostgroup", "--max-hostgroup"},
		{"--min-parallelism", "--max-parallelism"},
	}

	// Check each pair and report when the minimum exceeds the maximum.
	var errs []string
	for _, pair := range intPairs {
		minKey, maxKey := pair[0], pair[1]
		minVal, hasMin := intVals[minKey]
		maxVal, hasMax := intVals[maxKey]

		// Only report when both flags were actually provided.
		if hasMin && hasMax && minVal > maxVal {
			errs = append(errs, fmt.Sprintf(
				"%s (%d) must not be greater than %s (%d)",
				minKey, minVal, maxKey, maxVal,
			))
		}
	}

	floatPairs := [][2]string{
		{"--min-rate", "--max-rate"},
	}

	// Check each pair and report when the minimum exceeds the maximum.
	for _, pair := range floatPairs {
		minKey, maxKey := pair[0], pair[1]
		minVal, hasMin := floatVals[minKey]
		maxVal, hasMax := floatVals[maxKey]

		// Only report when both flags were actually provided.
		if hasMin && hasMax && minVal > maxVal {
			errs = append(errs, fmt.Sprintf(
				"%s (%g) must not be greater than %s (%g)",
				minKey, minVal, maxKey, maxVal,
			))
		}
	}

	// Return reported errors.
	return errs
}

// parseTimeDuration converts an Nmap time string to milliseconds.
// Nmap supports suffixes ms (milliseconds), s (seconds), m (minutes), h (hours).
// A bare number without suffix is treated as seconds, matching Nmap's default behaviour.
func parseTimeDuration(s string) (float64, error) {

	// Determine the suffix and extract the numeric part.
	var numStr string
	var multiplier float64
	switch {
	case strings.HasSuffix(s, "ms"):
		numStr = s[:len(s)-2]
		multiplier = 1
	case strings.HasSuffix(s, "h"):
		numStr = s[:len(s)-1]
		multiplier = 3600000
	case strings.HasSuffix(s, "m"):
		numStr = s[:len(s)-1]
		multiplier = 60000
	case strings.HasSuffix(s, "s"):
		numStr = s[:len(s)-1]
		multiplier = 1000
	default:
		// No suffix means seconds in Nmap.
		numStr = s
		multiplier = 1000
	}

	// Parse the numeric portion as a float.
	val, errParse := strconv.ParseFloat(numStr, 64)
	if errParse != nil {
		return 0, fmt.Errorf("invalid time value %q", s)
	}

	// Return the duration converted to milliseconds.
	return val * multiplier, nil
}

// checkTimeRanges validates that every time-based min/max pair satisfies min <= max
// after converting both values to a common unit (milliseconds).
func checkTimeRanges(timeVals map[string]float64) []string {

	// timePairs lists every (min-flag, max-flag) combination that must be ordered.
	timePairs := [][2]string{
		{"--min-rtt-timeout", "--max-rtt-timeout"},
		{"--initial-rtt-timeout", "--max-rtt-timeout"},
		{"--scan-delay", "--max-scan-delay"},
	}

	// Check each pair and report when the minimum exceeds the maximum.
	var errs []string
	for _, pair := range timePairs {
		minKey, maxKey := pair[0], pair[1]
		minVal, hasMin := timeVals[minKey]
		maxVal, hasMax := timeVals[maxKey]

		// Only report when both flags were actually provided.
		if hasMin && hasMax && minVal > maxVal {
			errs = append(errs, fmt.Sprintf(
				"%s must not be greater than %s (after unit conversion)",
				minKey, maxKey,
			))
		}
	}

	// Return reported errors.
	return errs
}
