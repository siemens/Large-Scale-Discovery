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
	"encoding/json"
	"strconv"
	"time"
)

var TimeFormat = "15:04"

type Timespans []Timespan

func (t *Timespans) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

type Timespan struct {
	StartDay  string `json:"startDay"`
	StartTime string `json:"startTime"`
	EndDay    string `json:"endDay"`
	EndTime   string `json:"endTime"`
}

func (t *Timespan) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

// TimezonesActive calculates timezones that have a current time between earliest and latest. Each slice
// returned represents a from-to timezone range. Returns one timezone range if the time spawn does not cross midnight,
// or two otherwise. Timezone ranges are better than a simple list of timezones, because they also cover potential/rare
// half or three-quarter step time zones, such as, 11.5. Returns the complete timezone range if start clock equals
// end clock.
func TimezonesActive(timespans Timespans) []int {
	return timezonesActive(time.Now().UTC(), timespans)
}

// timezonesActive is the private interface to allow internal unit tests assuming an arbitrary time
func timezonesActive(nowAtUtc0 time.Time, timespans Timespans) []int {

	// Convert to UTC-0 if it is not already
	nowAtUtc0 = nowAtUtc0.UTC()

	// Test timezones
	timezones := make([]int, 0, 25)
	for tz := -12; tz <= 12; tz++ {

		// Get current time in timezone 'tz'
		nowAtUtc := nowAtUtc0.Add(time.Hour * time.Duration(tz))

		// Iterate timespans
		for _, timespan := range timespans {

			// Cast timespan values to time types
			startDayInt, _ := strconv.ParseInt(timespan.StartDay, 10, 64)
			endDayInt, _ := strconv.ParseInt(timespan.EndDay, 10, 64)

			// Prepare parsed day and time values
			timespanStartDay := time.Weekday(startDayInt)
			timespanEndDay := time.Weekday(endDayInt)
			timespanStartTime, _ := time.Parse(TimeFormat, timespan.StartTime)
			timespanEndTime, _ := time.Parse(TimeFormat, timespan.EndTime)

			// Check if
			nowAfterStart := isTimeAfter(nowAtUtc, timespanStartTime)
			nowBeforeEnd := isTimeAfter(timespanEndTime, nowAtUtc)

			// Shift offset of the timestamp and the start and end day
			if timespanStartDay > timespanEndDay {

				// Calculate offset
				offset := 7 - timespanStartDay

				// Shift by offset
				nowAtUtc = nowAtUtc.Add(time.Hour * 24 * time.Duration(offset))
				timespanStartDay = timespanStartDay + offset
				timespanEndDay = timespanEndDay + offset

				// Fix overflown values
				if timespanStartDay > 6 {
					timespanStartDay -= 7
				}
				if timespanEndDay > 6 {
					timespanEndDay -= 7
				}
			}

			// End before start, which means scan runs across week boundary
			if timespanStartDay == timespanEndDay {

				// Timezone within this timespan definition
				if nowAtUtc.Weekday() >= timespanStartDay && nowAfterStart && nowBeforeEnd && nowAtUtc.Weekday() <= timespanEndDay {
					timezones = append(timezones, tz)
				}
			} else if timespanStartDay < timespanEndDay { // Simple timespan case within same week

				// Timezone within this timespan definition
				dayAtStart := nowAtUtc.Weekday() == timespanStartDay
				dayAtEnd := nowAtUtc.Weekday() == timespanEndDay
				dayBetween := nowAtUtc.Weekday() > timespanStartDay && nowAtUtc.Weekday() < timespanEndDay
				if (dayAtStart && nowAfterStart) ||
					(dayAtEnd && nowBeforeEnd) ||
					dayBetween {
					timezones = append(timezones, tz)
				}

			} else { // Timespan running across border day 6 re-entering at day 0

				// Case should not exist, since the offset was shifted above to overcome this case
				continue

			}
		}
	}

	// Return collected timezones
	return UniqueInts(timezones)
}

func isTimeAfter(t1, t2 time.Time) bool {
	h1, m1, s1 := t1.Clock()
	h2, m2, s2 := t2.Clock()

	if h1 != h2 {
		return h1 > h2
	}
	if m1 != m2 {
		return m1 > m2
	}
	return s1 > s2
}
