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
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var timespans9To5 = Timespans{
	Timespan{ // Monday
		StartDay:  "1",
		StartTime: "09:00", // 9
		EndDay:    "1",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Tuesday
		StartDay:  "2",
		StartTime: "09:00", // 9
		EndDay:    "2",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Wednesday
		StartDay:  "3",
		StartTime: "09:00", // 9
		EndDay:    "3",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Thursday
		StartDay:  "4",
		StartTime: "09:00", // 9
		EndDay:    "4",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Friday
		StartDay:  "5",
		StartTime: "09:00", // 9
		EndDay:    "5",     // to
		EndTime:   "17:00", // 5
	},
}
var timespansOverlapping = Timespans{
	Timespan{ // Monday
		StartDay:  "1",
		StartTime: "12:00", // 12
		EndDay:    "1",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Monday
		StartDay:  "1",
		StartTime: "00:00", // 0
		EndDay:    "1",     // to
		EndTime:   "04:00", // 4
	},
	Timespan{ // Tuesday
		StartDay:  "2",
		StartTime: "09:00", // 9
		EndDay:    "2",     // to
		EndTime:   "17:00", // 5
	},
	Timespan{ // Saturday - Monday
		StartDay:  "0",
		StartTime: "09:00", // 9
		EndDay:    "1",     // to
		EndTime:   "09:00", // 9
	},
	Timespan{ // Tuesday - Friday
		StartDay:  "2",
		StartTime: "09:00", // 9
		EndDay:    "5",     // to
		EndTime:   "09:00", // 9
	},
	Timespan{ // Wednesday - Thursday
		StartDay:  "2",
		StartTime: "09:00", // 9
		EndDay:    "5",     // to
		EndTime:   "17:00", // 5
	},
}

func TestTimezonesActive(t *testing.T) {

	// Prepare static timezone for unit tests
	locUtc2 := time.FixedZone("UTC+2", 2*60*60)

	// Prepare and run test cases
	tests := []struct {
		name      string
		time      time.Time
		timespans Timespans
		want      []int
	}{
		{
			name: "simple-no-hour",
			time: time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:10",
					EndDay:    "1",
					EndTime:   "09:11",
				},
			},
			want: []int{},
		},

		//
		// Test cases for start, middle and end cases within simple one-hour timespans
		//
		{
			name: "simple-same-hour-early-beginning",
			time: time.Date(2025, 10, 12, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "01:00",
				},
			},
			want: []int{3},
		},
		{
			name: "simple-same-hour-early-middle",
			time: time.Date(2025, 10, 13, 0, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "01:00",
				},
			},
			want: []int{2},
		},
		{
			name: "simple-same-hour-early-end",
			time: time.Date(2025, 10, 13, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "01:00",
				},
			},
			want: []int{1},
		},
		{
			name: "simple-same-hour-noon-beginning",
			time: time.Date(2025, 10, 13, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "10:00",
				},
			},
			want: []int{3},
		},
		{
			name: "simple-same-hour-noon-middle",
			time: time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "10:00",
				},
			},
			want: []int{2},
		},
		{
			name: "simple-same-hour-noon-end",
			time: time.Date(2025, 10, 13, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "10:00",
				},
			},
			want: []int{1},
		},
		{
			name: "simple-same-hour-late-beginning",
			time: time.Date(2025, 10, 13, 22, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "23:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{3},
		},
		{
			name: "simple-same-hour-late-middle",
			time: time.Date(2025, 10, 13, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "23:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{2},
		},
		{
			name: "simple-same-hour-late-end",
			time: time.Date(2025, 10, 14, 0, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "23:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{1},
		},

		//
		// Test cases for start, middle and end cases within a simple same-day timespan
		//
		{
			name: "simple-same-day-beginning-1",
			time: time.Date(2025, 10, 13, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name: "simple-same-day-beginning-2",
			time: time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "simple-same-day-beginning-3",
			time: time.Date(2025, 10, 13, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "simple-same-day-middle",
			time: time.Date(2025, 10, 13, 13, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{-2, -1, 0, 1, 2, 3, 4, 5},
		},
		{
			name: "simple-same-day-end-1",
			time: time.Date(2025, 10, 13, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{-5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name: "simple-same-day-end-2",
			time: time.Date(2025, 10, 13, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{-6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name: "simple-same-day-end-3",
			time: time.Date(2025, 10, 13, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "1",
					EndTime:   "17:00",
				},
			},
			want: []int{-7, -6, -5, -4, -3, -2, -1, 0},
		},

		//
		// Test cases for start, middle and end cases within a simple full-day timespan
		//
		{
			name: "simple-one-day-beginning-1",
			time: time.Date(2025, 10, 12, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-day-beginning-2",
			time: time.Date(2025, 10, 13, 0, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-day-beginning-3",
			time: time.Date(2025, 10, 13, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-day-middle",
			time: time.Date(2025, 10, 13, 14, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		{
			name: "simple-one-day-end-1",
			time: time.Date(2025, 10, 13, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name: "simple-one-day-end-2",
			time: time.Date(2025, 10, 13, 23, 59, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name: "simple-one-day-end-3",
			time: time.Date(2025, 10, 14, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "00:00",
					EndDay:    "1",
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0},
		},

		//
		// Test cases for start, middle and end cases within a simple full-week timespan
		//
		{
			name: "simple-one-week-beginning-1",
			time: time.Date(2025, 10, 11, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-beginning-2",
			time: time.Date(2025, 10, 12, 0, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-beginning-3",
			time: time.Date(2025, 10, 12, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-middle",
			time: time.Date(2025, 10, 15, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-end-1",
			time: time.Date(2025, 10, 18, 22, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-end-2",
			time: time.Date(2025, 10, 18, 23, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "simple-one-week-end-3",
			time: time.Date(2025, 10, 19, 0, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "0", // Sunday
					StartTime: "00:00",
					EndDay:    "6", // Monday
					EndTime:   "23:59",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},

		//
		// Test cases for start, middle and end cases within a simple one-day timespan
		//
		{
			name: "dayspan-simple-beginning-1",
			time: time.Date(2025, 10, 13, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-simple-beginning-2",
			time: time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-simple-beginning-3",
			time: time.Date(2025, 10, 13, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-simple-middle",
			time: time.Date(2025, 10, 14, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-simple-end-1",
			time: time.Date(2025, 10, 14, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name: "dayspan-simple-end-2",
			time: time.Date(2025, 10, 14, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name: "dayspan-simple-end-3",
			time: time.Date(2025, 10, 14, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1",
					StartTime: "09:00",
					EndDay:    "2",
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0},
		},

		//
		// Test cases for start, middle and end cases within a complex timespan where start day integer > end day integer
		//
		{
			name: "dayspan-complex-week-beginning-1",
			time: time.Date(2025, 10, 13, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-week-beginning-2",
			time: time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-week-beginning-3",
			time: time.Date(2025, 10, 13, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-week-middle",
			time: time.Date(2025, 10, 16, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-week-end-1",
			time: time.Date(2025, 10, 19, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name: "dayspan-complex-week-end-2",
			time: time.Date(2025, 10, 19, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name: "dayspan-complex-week-end-3",
			time: time.Date(2025, 10, 19, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "1", // Monday
					StartTime: "09:00",
					EndDay:    "0", // Sunday
					EndTime:   "17:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0},
		},

		//
		// Test cases for some more complex timespans with overflowing day frames
		//
		{
			name: "dayspan-complex-day-beginning-1",
			time: time.Date(2025, 10, 18, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-day-beginning-2",
			time: time.Date(2025, 10, 18, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-day-beginning-3",
			time: time.Date(2025, 10, 18, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "dayspan-complex-day-middle",
			time: time.Date(2025, 10, 19, 1, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{-6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "dayspan-complex-day-end-1",
			time: time.Date(2025, 10, 19, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name: "dayspan-complex-day-end-2",
			time: time.Date(2025, 10, 19, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name: "dayspan-complex-day-end-3",
			time: time.Date(2025, 10, 19, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: Timespans{
				Timespan{
					StartDay:  "6", // Saturday
					StartTime: "17:00",
					EndDay:    "0", // Sunday
					EndTime:   "09:00",
				},
			},
			want: []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0},
		},

		//
		// Test cases for some slices of timespans
		//
		{
			name:      "slice-week-monday-beginning-1",
			time:      time.Date(2025, 10, 13, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:      "slice-week-monday-beginning-2",
			time:      time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:      "slice-week-monday-beginning-3",
			time:      time.Date(2025, 10, 13, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:      "slice-week-monday-end-1",
			time:      time.Date(2025, 10, 13, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name:      "slice-week-monday-end-2",
			time:      time.Date(2025, 10, 13, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name:      "slice-week-monday-end-3",
			time:      time.Date(2025, 10, 13, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-7, -6, -5, -4, -3, -2, -1, 0},
		},
		{
			name:      "slice-week-wednesday-beginning-1",
			time:      time.Date(2025, 10, 15, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:      "slice-week-wednesday-beginning-2",
			time:      time.Date(2025, 10, 15, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:      "slice-week-wednesday-beginning-3",
			time:      time.Date(2025, 10, 15, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:      "slice-week-wednesday-end-1",
			time:      time.Date(2025, 10, 15, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name:      "slice-week-wednesday-end-2",
			time:      time.Date(2025, 10, 15, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name:      "slice-week-wednesday-end-3",
			time:      time.Date(2025, 10, 15, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-7, -6, -5, -4, -3, -2, -1, 0},
		},
		{
			name:      "slice-week-friday-beginning-1",
			time:      time.Date(2025, 10, 17, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:      "slice-week-friday-beginning-2",
			time:      time.Date(2025, 10, 17, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:      "slice-week-friday-beginning-3",
			time:      time.Date(2025, 10, 17, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:      "slice-week-friday-end-1",
			time:      time.Date(2025, 10, 17, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-5, -4, -3, -2, -1, 0, 1, 2},
		},
		{
			name:      "slice-week-friday-end-2",
			time:      time.Date(2025, 10, 17, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-6, -5, -4, -3, -2, -1, 0, 1},
		},
		{
			name:      "slice-week-friday-end-3",
			time:      time.Date(2025, 10, 17, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{-7, -6, -5, -4, -3, -2, -1, 0},
		},
		{
			name:      "slice-week-saturday-beginning-1",
			time:      time.Date(2025, 10, 18, 8, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},
		{
			name:      "slice-week-saturday-beginning-2",
			time:      time.Date(2025, 10, 18, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},
		{
			name:      "slice-week-saturday-beginning-3",
			time:      time.Date(2025, 10, 18, 10, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},
		{
			name:      "slice-week-saturday-end-1",
			time:      time.Date(2025, 10, 18, 16, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},
		{
			name:      "slice-week-saturday-end-2",
			time:      time.Date(2025, 10, 18, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},
		{
			name:      "slice-week-saturday-end-3",
			time:      time.Date(2025, 10, 18, 18, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespans9To5), // Monday - Friday, 9:00 - 17:00
			want:      []int{},
		},

		//
		// Test cases for some complex slices of timespans with all kinds and overlaps
		//
		{
			name:      "slice-complex-overlapping-1",
			time:      time.Date(2025, 10, 13, 4, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 10, 11, 12},
		},
		{
			name:      "slice-complex-overlapping-2",
			time:      time.Date(2025, 10, 13, 9, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 5, 6, 7, 8, 9},
		},
		{
			name:      "slice-complex-overlapping-3",
			time:      time.Date(2025, 10, 13, 17, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{-12, -11, -10, -9, -8, -7, -3, -2, -1, 0, 1},
		},
		{
			name:      "slice-complex-overlapping-4",
			time:      time.Date(2025, 10, 15, 12, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name:      "slice-complex-overlapping-5",
			time:      time.Date(2025, 10, 17, 15, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{-12, -11, -10, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3},
		},
		{
			name:      "slice-complex-overlapping-6",
			time:      time.Date(2025, 10, 18, 12, 0, 1, 0, locUtc2).UTC(),
			timespans: shuffle(timespansOverlapping),
			want:      []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timezonesActive(tt.time, tt.timespans)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timezonesActive() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

func shuffle(timespans Timespans) Timespans {
	for i := range timespans {
		j := rand.Intn(i + 1)
		timespans[i], timespans[j] = timespans[j], timespans[i]
	}
	return timespans
}
