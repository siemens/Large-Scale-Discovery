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

func TestCountIpsInInput(t *testing.T) {
	type args struct {
		subnet string
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{"Single Host", args{subnet: "test.domain.tld"}, 1, false},
		{"Single Ip", args{subnet: "10.10.0.1"}, 1, false},
		{"Single Subnet", args{subnet: "10.10.0.0/32"}, 1, false},
		{"Point2Point", args{subnet: "10.10.0.0/31"}, 2, false},
		{"Small Subnet", args{subnet: "10.10.0.0/30"}, 4, false},
		{"Medium Subnet", args{subnet: "10.10.0.0/24"}, 256, false},
		{"Large Subnet", args{subnet: "10.10.10.0/22"}, 1024, false},
		{"Huge Subnet", args{subnet: "10.10.0.0/17"}, 32768, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountIpsInInput(tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountIpsInInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountIpsInInput() got = %v, want %v", got, tt.want)
			}
		})
	}
}
