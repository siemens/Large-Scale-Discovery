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
	"testing"
)

func TestIntContained(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		candidate int
		slices    [][]int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"one-slice-contained", args{1111, [][]int{{1, 2, 3, 1111}}}, true},
		{"one-slice-not-contained", args{1111, [][]int{{1, 2, 3, 4}}}, false},

		{"multiple-slices-contained", args{1111, [][]int{{1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 1111}, {1, 2, 3, 4}}}, true},
		{"multiple-slices-not-contained", args{1111, [][]int{{1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 5}, {1, 2, 3, 4}}}, false},

		{"known-1", args{222, [][]int{{222, 555, 666}, {33333, 777, 888}}}, true},
		{"known-2", args{222, [][]int{{222, 555, 666}, {222, 555, 666}}}, true},
		{"known-3", args{33333, [][]int{{}, {33333, 777, 888}}}, true},
		{"unknown-1", args{4444444, [][]int{{222, 555, 666}, {33333, 777, 888}}}, false},
		{"unknown-2", args{1111, [][]int{{222, 555, 666}, {33333, 777, 888}}}, false},
		{"unknown-3", args{1111, [][]int{{}, {33333, 777, 888}}}, false},
		{"unknown-4", args{1111, [][]int{{}, {}}}, false},
		{"unknown-5", args{1111, [][]int{{}}}, false},
		{"unknown-6", args{1111, [][]int{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntContained(tt.args.candidate, tt.args.slices...); got != tt.want {
				t.Errorf("UintContained() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

func TestUint64Contained(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		candidate uint64
		slices    [][]uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"one-slice-contained", args{1111, [][]uint64{{1, 2, 3, 1111}}}, true},
		{"one-slice-not-contained", args{1111, [][]uint64{{1, 2, 3, 4}}}, false},

		{"multiple-slices-contained", args{1111, [][]uint64{{1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 1111}, {1, 2, 3, 4}}}, true},
		{"multiple-slices-not-contained", args{1111, [][]uint64{{1, 2, 3, 4}, {1, 2, 3, 4}, {1, 2, 3, 5}, {1, 2, 3, 4}}}, false},

		{"known-1", args{222, [][]uint64{{222, 555, 666}, {33333, 777, 888}}}, true},
		{"known-2", args{222, [][]uint64{{222, 555, 666}, {222, 555, 666}}}, true},
		{"known-3", args{33333, [][]uint64{{}, {33333, 777, 888}}}, true},
		{"unknown-1", args{4444444, [][]uint64{{222, 555, 666}, {33333, 777, 888}}}, false},
		{"unknown-2", args{1111, [][]uint64{{222, 555, 666}, {33333, 777, 888}}}, false},
		{"unknown-3", args{1111, [][]uint64{{}, {33333, 777, 888}}}, false},
		{"unknown-4", args{1111, [][]uint64{{}, {}}}, false},
		{"unknown-5", args{1111, [][]uint64{{}}}, false},
		{"unknown-6", args{1111, [][]uint64{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uint64Contained(tt.args.candidate, tt.args.slices...); got != tt.want {
				t.Errorf("Uint64Contained() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
