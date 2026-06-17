/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// UniqueInts gets rid of redundant elements
func UniqueInts(elements []int) []int {

	// Use map to record duplicates as we find them.
	var encountered = make(map[int]bool, len(elements))
	var result = make([]int, 0, len(elements))

	// Iterate elements and add them to the new slice if they were not seen before
	for _, v := range elements {
		if encountered[v] {
			// Do not add duplicate.
		} else {
			encountered[v] = true      // Record this element as an encountered element.
			result = append(result, v) // Append to result slice.
		}
	}

	// Return the new slice.
	return result
}

// IntContained checks whether a given int value is contained within one or multiple given slices
func IntContained(candidate int, slices ...[]int) bool {

	// Translate integers into map for faster lookups
	items := make(map[int]struct{})
	for _, slice := range slices {
		for _, item := range slice {
			items[item] = struct{}{}
		}
	}

	// Search items for candidate
	_, ok := items[candidate]
	if ok {
		return true
	}

	// Return false as item was not found in candidates
	return false
}

// Uint64Contained checks whether a given int64 value is contained within one or multiple given slices
func Uint64Contained(candidate uint64, slices ...[]uint64) bool {

	// Translate integers into map for faster lookups
	items := make(map[uint64]struct{})
	for _, slice := range slices {
		for _, item := range slice {
			items[item] = struct{}{}
		}
	}

	// Search items for candidate
	_, ok := items[candidate]
	if ok {
		return true
	}

	// Return false as item was not found in candidates
	return false
}

// JoinInt converts a slice of ints into strings and concatenates them using the given delimiter
func JoinInt(ints []int, delimiter string) string {
	return strings.Trim(strings.Join(strings.Split(fmt.Sprint(ints), " "), delimiter), "[]")
}

// JoinUint64 converts a slice of int64's into strings and concatenates them using the given delimiter
func JoinUint64(uints []uint64, delimiter string) string {
	return strings.Trim(strings.Join(strings.Split(fmt.Sprint(uints), " "), delimiter), "[]")
}

// GetIntValue safely converts an interface to int
func GetIntValue(val interface{}) int {

	// Handle nil case
	if val == nil {
		return 0
	}

	// Try to convert based on type
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:

		// Parse string to integer
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return intVal
	default:
		return 0
	}
}
