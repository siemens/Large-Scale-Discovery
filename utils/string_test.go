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
	"reflect"
	"testing"
)

func TestRemoveFromSlice(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		slice []string
		s     string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"empty", args{[]string{}, "3"}, nil},
		{"no-occurrence", args{[]string{"1", "2"}, "3"}, []string{"1", "2"}},
		{"one-occurrence", args{[]string{"1", "2", "3"}, "3"}, []string{"1", "2"}},
		{"multiple-occurrences", args{[]string{"3", "1", "3", "3", "2", "3"}, "3"}, []string{"1", "2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveFromSlice(tt.args.slice, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveFromSlice() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

func TestToValidUtf8String(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"trailing-nulls", args{[]byte{97, 97, 0, 0}}, "aa"},
		{"none-trailing-nulls", args{[]byte{97, 0, 0, 0, 97}}, "a•••a"},
		{"trailing-none-trailing-nulls", args{[]byte{97, 0, 0, 0, 97, 0, 0, 0, 0, 0, 0, 0}}, "a•••a"},
		{"leading-nulls", args{[]byte{0, 0, 0, 97}}, "•••a"},
		{"only-nulls", args{[]byte{0, 0, 0}}, ""},
		{"no-nulls", args{[]byte{97, 97, 97}}, "aaa"},

		{"null-buffer-1", args{[]byte("\000\000\000\x00\x00\x00\u0000\u0000\u0000")}, ""},
		{"null-buffer-2", args{[]byte{0, 0, 0, 0, 0, 0}}, ""},
		{"empty-buffer", args{[]byte{}}, ""},
		{"valid-bytes-1", args{[]byte{'h', 'i'}}, "hi"},
		{"valid-bytes-2", args{[]byte{65, 66, 67}}, "ABC"},
		{"invalid-bytes", args{[]byte{'¿'}}, "�"},
		{"trailing-null-chars-1", args{[]byte{'n', '¿', 'l', 'l', 0, 0}}, "n�ll"},
		{"trailing-null-chars-2", args{[]byte{65, 66, 67, 0, 0}}, "ABC"},
		{"non-trailing-null-chars", args{[]byte{'n', 0, 0, 'l', 'l', 0}}, "n••ll"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToValidUtf8String(tt.args.b); got != tt.want {
				t.Errorf("ToValidUtf8String() = %v, want %v", got, tt.want)
			}
		})
	}
}
