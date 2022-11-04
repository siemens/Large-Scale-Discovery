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
	"reflect"
	"testing"
	"time"
)

func TestTimeInRange(t *testing.T) {
	type args struct {
		candidate time.Time
		start     string
		end       string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{time.Date(2020, 06, 30, 4, 0, 0, 0, time.UTC), "23:00", "05:00"}, true},
		{"", args{time.Date(2020, 06, 30, 23, 30, 0, 0, time.UTC), "23:00", "05:00"}, true},
		{"", args{time.Date(2020, 06, 30, 20, 0, 0, 0, time.UTC), "23:00", "05:00"}, false},
		{"", args{time.Date(2020, 06, 30, 11, 0, 0, 0, time.UTC), "10:00", "21:00"}, true},
		{"", args{time.Date(2020, 06, 30, 22, 0, 0, 0, time.UTC), "10:00", "21:00"}, false},
		{"", args{time.Date(2020, 06, 30, 03, 0, 0, 0, time.UTC), "10:00", "21:00"}, false},
		{"", args{time.Date(2020, 06, 30, 00, 0, 0, 0, time.UTC), "22:00", "02:00"}, true},
		{"", args{time.Date(2020, 06, 30, 10, 0, 0, 0, time.UTC), "10:00", "21:00"}, true},
		{"", args{time.Date(2020, 06, 30, 21, 0, 0, 0, time.UTC), "10:00", "21:00"}, true},
		{"", args{time.Date(2020, 06, 30, 06, 0, 0, 0, time.UTC), "23:00", "05:00"}, false},
		{"", args{time.Date(2020, 06, 30, 23, 0, 0, 0, time.UTC), "23:00", "05:00"}, true},
		{"", args{time.Date(2020, 06, 30, 05, 0, 0, 0, time.UTC), "23:00", "05:00"}, true},
		{"", args{time.Date(2020, 06, 30, 10, 0, 0, 0, time.UTC), "10:00", "21:00"}, true},
		{"", args{time.Date(2020, 06, 30, 21, 0, 0, 0, time.UTC), "10:00", "21:00"}, true},
		{"", args{time.Date(2020, 06, 30, 9, 0, 0, 0, time.UTC), "10:00", "10:00"}, false},
		{"", args{time.Date(2020, 06, 30, 11, 0, 0, 0, time.UTC), "10:00", "10:00"}, false},
		{"", args{time.Date(2020, 06, 30, 10, 0, 0, 0, time.UTC), "10:00", "10:00"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := "15:04"
			startTime, _ := time.Parse(format, tt.args.start)
			endTime, _ := time.Parse(format, tt.args.end)
			if got := TimeInRange(tt.args.candidate, startTime, endTime); got != tt.want {
				t.Errorf("TimeInRange() = %v, want %v", got, tt.want)
			} else {
				fmt.Println(tt.args.candidate.Format("15:04"), "between", tt.args.start, "-", tt.args.end, ":", got)
			}
		})
	}
}

func Test_timezonesCurrentlyBetween(t *testing.T) {
	nowTemplate := time.Now().UTC()
	type args struct {
		now         time.Time
		earliest    string
		latest      string
		invalidDays []time.Weekday
	}
	tests := []struct {
		name           string
		args           args
		timezoneRanges [][]int
	}{
		{
			"24h-range-all",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 9, 0, 0, 0, time.UTC),
				earliest: "00:00",
				latest:   "00:00",
			},
			[][]int{{-12, 12}},
		},

		{
			"simple-all",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 9, 0, 0, 0, time.UTC),
				earliest: "00:00",
				latest:   "23:59",
			},
			[][]int{{-12, 12}},
		},
		{
			"simple-all",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 0, 0, 0, 0, time.UTC),
				earliest: "00:00",
				latest:   "23:59",
			},
			[][]int{{-12, 12}},
		},
		{
			"simple-all",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 23, 59, 0, 0, time.UTC),
				earliest: "00:00",
				latest:   "23:59",
			},
			[][]int{{-12, 12}},
		},

		{
			"simple-now-before-start",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 8, 59, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{1, 7}},
		},
		{
			"simple-now-at-start",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 9, 0, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{0, 7}},
		},
		{
			"simple-now-at-end",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 16, 0, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{-7, 0}},
		},
		{
			"simple-now-after-end",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 16, 1, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{-7, -1}},
		},

		{
			"simple-now-early",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 1, 0, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{-12, -9}, {8, 12}},
		},

		{
			"simple-now-late",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 23, 0, 0, 0, time.UTC),
				earliest: "09:00",
				latest:   "16:00",
			},
			[][]int{{-12, -7}, {10, 12}},
		},

		{
			"late-range-now-early",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 1, 0, 0, 0, time.UTC),
				earliest: "23:00",
				latest:   "00:00",
			},
			[][]int{{-2, -1}},
		},
		{
			"late-range-now-late",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 23, 0, 0, 0, time.UTC),
				earliest: "23:00",
				latest:   "00:00",
			},
			[][]int{{0, 1}},
		},

		{
			"early-range-now-early",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 1, 0, 0, 0, time.UTC),
				earliest: "01:00",
				latest:   "02:00",
			},
			[][]int{{0, 1}},
		},
		{
			"early-range-now-late",
			args{
				now:      time.Date(nowTemplate.Year(), nowTemplate.Month(), nowTemplate.Day(), 23, 0, 0, 0, time.UTC),
				earliest: "01:00",
				latest:   "02:00",
			},
			[][]int{{2, 3}},
		},

		//
		// Some tests with weekend days excluded
		//

		{
			"all-except-weekend-starting-thursday",
			args{
				now:         time.Date(2020, 07, 02, 18, 0, 0, 0, time.UTC),
				earliest:    "00:00",
				latest:      "23:59",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{-12, 12}},
		},
		{
			"all-except-weekend-starting-thursday",
			args{
				now:         time.Date(2020, 07, 02, 18, 0, 0, 0, time.UTC),
				earliest:    "00:00",
				latest:      "00:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{-12, 12}},
		},
		{
			"all-except-weekend-starting-friday",
			args{
				now:         time.Date(2020, 7, 3, 18, 0, 0, 0, time.UTC),
				earliest:    "00:00",
				latest:      "23:59",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{-12, 5}},
		},
		{
			"all-except-weekend-starting-friday",
			args{
				now:         time.Date(2020, 7, 3, 18, 0, 0, 0, time.UTC),
				earliest:    "00:00",
				latest:      "00:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{-12, 5}},
		},

		{
			"some-except-weekend-starting-friday",
			args{
				now:         time.Date(2020, 7, 3, 21, 0, 0, 0, time.UTC),
				earliest:    "21:00",
				latest:      "03:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{0, 2}}, // Timezone +3 is 00:00 o'clock, which is already Saturday!
		},
		{
			"some-except-weekend-starting-sunday",
			args{
				now:         time.Date(2020, 7, 5, 21, 0, 0, 0, time.UTC),
				earliest:    "21:00",
				latest:      "03:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{3, 6}},
		},
		{
			"some-except-weekend-starting-sunday",
			args{
				now:         time.Date(2020, 7, 6, 6, 0, 0, 0, time.UTC),
				earliest:    "21:00",
				latest:      "03:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{{-6, -3}},
		},

		{
			"none-because-saturday",
			args{
				now:         time.Date(2020, 7, 4, 9, 0, 0, 0, time.UTC),
				earliest:    "09:00",
				latest:      "16:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{},
		},
		{
			"none-because-saturday",
			args{
				now:         time.Date(2020, 7, 4, 21, 0, 0, 0, time.UTC),
				earliest:    "09:00",
				latest:      "16:00",
				invalidDays: []time.Weekday{time.Saturday, time.Sunday},
			},
			[][]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timezoneRanges := timezonesBetween(tt.args.now, tt.args.earliest, tt.args.latest, tt.args.invalidDays)
			if !reflect.DeepEqual(timezoneRanges, tt.timezoneRanges) {
				t.Errorf("timezonesBetween() timezoneRanges = %v, want %v", timezoneRanges, tt.timezoneRanges)
			}
		})
	}
}
