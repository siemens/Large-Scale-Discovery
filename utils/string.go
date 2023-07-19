/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
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
