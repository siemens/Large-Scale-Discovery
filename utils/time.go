/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"fmt"
	"strings"
	"time"
)

// TimeInRange decides whether a given time is between two clocks, ignoring dates.
func TimeInRange(candidate time.Time, start time.Time, end time.Time) bool {

	// Get rid of date, we only care about time
	candidate = time.Date(0, 1, 1, candidate.Hour(), candidate.Minute(), 0, 0, time.UTC)

	// Pares string time
	start = time.Date(0, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	end = time.Date(0, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)

	// Standard check if start time lower end time
	if start.Before(end) {
		return !candidate.Before(start) && !candidate.After(end)
	}

	// Edge case start time equal to end time
	if start.Equal(end) {
		return candidate.Equal(start)
	}

	// Inverse check, if start time greater end time
	startAfterCandidate := start.After(candidate)
	endBeforeCandidate := end.Before(candidate)
	if !startAfterCandidate || !endBeforeCandidate {
		return true
	} else {
		return false
	}
}

// JoinWeekdays converts weekday integers into string and concatenates them using the given delimiter
func JoinWeekdays(days []time.Weekday, delimiter string) string {
	return strings.Trim(strings.Join(strings.Split(fmt.Sprintf("%d", days), " "), delimiter), "[]")
}

// UniqueWeekdays gets rid of redundant elements
func UniqueWeekdays(elements []time.Weekday) []time.Weekday {

	// Use map to record duplicates as we find them.
	encountered := map[time.Weekday]bool{}
	var result []time.Weekday

	// Iterate elements and add them to the new slice if they were not seen before
	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}

	// Return the new slice.
	return result
}

// TimezonesBetween calculates timezones that have a current time between earliest and latest. Each slice
// returned represents a from-to timezone range. Returns one timezone range if the time spawn does not cross midnight,
// or two otherwise. Timezone ranges are better than a simple list of timezones, because they also cover potential/rare
// half or three quarter step time zones, such as, 11.5. Returns the complete timezone range if start clock equals
// end clock.
func TimezonesBetween(startClock string, endClock string, invalidDays []time.Weekday) [][]int {
	return timezonesBetween(time.Now(), startClock, endClock, invalidDays)
}

// timezonesBetween is the private interface to allow internal unit tests assuming an arbitrary time
func timezonesBetween(now time.Time, startClock string, endClock string, invalidDays []time.Weekday) [][]int {

	// Covert current time to UTC-0
	nowUtc := now.UTC()

	// Pares string time
	timeFormat := "15:04"
	startTime, _ := time.Parse(timeFormat, startClock)
	endTime, _ := time.Parse(timeFormat, endClock)

	// Prepare weekday lookup map
	invalid := make(map[time.Weekday]struct{})
	for _, invalidDay := range invalidDays {
		invalid[invalidDay] = struct{}{}
	}

	// Test timezones
	tzRanges := make([][]int, 0, 2)
	tzRange := make([]int, 0, 2)
	for tz := -12; tz <= 12; tz++ {

		// Get current time in timezone 'tz'
		nowInTz := nowUtc.Add(time.Hour * time.Duration(tz))

		// Check if timezone is in range
		var tzInRange bool
		if startTime.Equal(endTime) {
			tzInRange = true // Assume whole timezone range if start time equals end time
		} else {
			tzInRange = TimeInRange(nowInTz, startTime, endTime)
		}

		// Drop timezone if it is within an invalid day of the week
		_, invalidDay := invalid[nowInTz.Weekday()]
		if tzInRange && invalidDay {
			tzInRange = false
		}

		// Start or end range
		if tzInRange && len(tzRange) == 0 {
			tzRange = append(tzRange, tz)
		} else if !tzInRange && len(tzRange) == 1 {
			tzRange = append(tzRange, tz-1)
			tzRanges = append(tzRanges, tzRange)
			tzRange = make([]int, 0, 2)
		} else if tz == 12 && len(tzRange) == 1 {
			tzRange = append(tzRange, tz)
			tzRanges = append(tzRanges, tzRange)
			tzRange = make([]int, 0, 2)
		}

		// Test output for debugging
		// val := ""
		// if tzInRange {
		// val = "YES"
		// }
		// fmt.Printf("%3d  %s-%s ? %s (%s)\t: %s\n", tz, startClock, endClock, nowInTz.Format("15:04"), nowInTz.Weekday(), val)
	}

	// Test output for debugging
	// fmt.Println("Timezone ranges:", tzRanges)

	// Return timezone threshold
	return tzRanges
}
