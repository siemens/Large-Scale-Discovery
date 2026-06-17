/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"regexp"
	"strings"
)

// TrimToUpper converts slice elements to upper case and trim whitespaces
func TrimToUpper(slice []string) []string {
	var trimmedLowerSlice []string
	for _, item := range slice {
		trimmedLowerSlice = append(trimmedLowerSlice, strings.ToUpper(strings.TrimSpace(item)))
	}
	return trimmedLowerSlice
}

// RemoveFromSlice removes a given element (and potential duplicates) from a slice and returns a new slice
func RemoveFromSlice(list []string, s string) []string {

	var retList []string

	// Generate new slice dropping requested strings
	for _, current := range list {
		if current == s {
			continue
		} else {
			retList = append(retList, current)
		}
	}

	// Return new filtered slice
	return retList
}

// ValidUtf8String returns a valid utf-8 string by replacing all invalid byte sequences with a hardcoded replacement
// character. Additionally, it gets rid of trailing null bytes and replaces none-trailing null bytes.
func ValidUtf8String(str string) string {

	// Trim all trailing NULL characters from sequence (might come from allocated empty buffers)
	str = strings.TrimRight(str, "\x00")

	// Insert a '•' for non-trailing null characters
	str = strings.Replace(str, "\x00", "•", -1)

	// Insert a '�' for the invalid runes (aka. utf8.RuneError)
	return strings.ToValidUTF8(str, "�")
}

// ToValidUtf8String returns a valid utf-8 string by replacing all invalid byte sequences with a hardcoded replacement
// character. Additionally, it gets rid of trailing null bytes and replaces none-trailing null bytes.
func ToValidUtf8String(b []byte) string {

	// Convert the byte slice to a string
	str := string(b)

	// Insert a '�' for the invalid runes (aka. utf8.RuneError), remove trailing null bytes and replace none-trailing
	// null bytes with "•".
	return ValidUtf8String(str)
}

// SanitizeCommaSeparated normalizes user-provided text into a clean and beautified comma-separated string.
// The comma separated values have a single whitespace after each comma, for beautified display and automatic
// line breaks in graphical interfaces.
//
// It performs the following sanitization steps:
//   - Replaces carriage returns, newlines, and tabs with commas
//   - Removes whitespace surrounding commas
//   - Collapses multiple consecutive commas into a single comma
//   - Removes leading and trailing commas and whitespaces
//
// Examples:
//
//	"foo bar,\n bar\tbaz"    -> "foo bar,bar,baz"
//	",,,foo , bar,,,"   -> "foo,bar"
func SanitizeCommaSeparated(val string) string {

	// Normalize line breaks/tabs into commas
	val = strings.ReplaceAll(val, "\r", ",")
	val = strings.ReplaceAll(val, "\n", ",")
	val = strings.ReplaceAll(val, "\t", ",")

	// Normalize spaces around commas
	var spaceCommaRe = regexp.MustCompile(`\s*,\s*`)
	val = spaceCommaRe.ReplaceAllString(val, ",")

	// Collapse duplicate commas
	var multiCommaRe = regexp.MustCompile(`,+`)
	val = multiCommaRe.ReplaceAllString(val, ",")

	// Remove leading/trailing commas
	val = strings.Trim(val, ",")

	// Remove leading/trailing spaces
	val = strings.Trim(val, " ")

	// Beautify into string with one nice space after a comma
	vals := strings.Split(val, ",")
	val = strings.Join(vals, ", ")

	// Return result
	return val
}

// SanitizeToSlice splits a string by separator and sanitizes single values by triming spaces.
// Returns an empty slice if the input string was empty.
// Golang's original function would return a slice with one entry being an empty string.
func SanitizeToSlice(s, sep string) []string {

	// Return empty slice if string is empty
	if len(s) == 0 {
		return []string{}
	}

	// Split by comma
	parts := strings.Split(s, sep)

	// Trim spaces that might have been used to beautify the comma separated string
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, p)
	}

	// Return result
	return result
}
