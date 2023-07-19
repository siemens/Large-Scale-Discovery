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
	"testing"
)

func TestCreatePassword(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		pwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"match", args{"password"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, _ := CreatePasswordHash(tt.args.pwd)
			if got := CheckPasswordHash(hash, tt.args.pwd); got != tt.wantErr {
				t.Errorf("CreatePasswordHash(): %v, want = '%v'", got, tt.wantErr)
			}
		})
	}
}
