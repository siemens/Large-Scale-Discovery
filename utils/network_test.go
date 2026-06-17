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

func TestCountIpsInInput(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		subnet string
	}
	tests := []struct {
		name    string
		args    args
		want    uint
		wantErr bool
	}{
		{"empty-host", args{subnet: ""}, 0, false},
		{"single-host", args{subnet: "test.domain.tld"}, 1, false},
		{"single-ip", args{subnet: "10.10.0.1"}, 1, false},
		{"single-subnet", args{subnet: "10.10.0.0/32"}, 1, false},
		{"point-to-point", args{subnet: "10.10.0.0/31"}, 2, false},
		{"small-subnet", args{subnet: "10.10.0.0/30"}, 4, false},
		{"medium-subnet", args{subnet: "10.10.0.0/24"}, 256, false},
		{"large-subnet", args{subnet: "10.10.10.0/22"}, 1024, false},
		{"huge-subnet", args{subnet: "10.10.0.0/17"}, 32768, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountIpsInInput(tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountIpsInInput() error = '%v', wantErr = '%v'", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountIpsInInput() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

func TestSplitNetwork(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name       string
		network    string
		targetSize uint32
		want       []string
		wantErr    bool
	}{
		{"equal-size1", "10.0.0.0/32", 1, []string{"10.0.0.0/32"}, false},
		{"equal-size2", "10.0.0.0/31", 2, []string{"10.0.0.0/31"}, false},
		{"equal-size3", "10.0.0.0/22", 1024, []string{"10.0.0.0/22"}, false},
		{"bigger-size1", "10.0.0.0/22", 2048, []string{"10.0.0.0/22"}, false},
		{"bigger-size2", "10.0.0.0/22", 4096, []string{"10.0.0.0/22"}, false},
		{"lower-size1", "10.0.0.0/22", 512, []string{"10.0.0.0/23", "10.0.2.0/23"}, false},
		{"lower-size2", "10.0.0.0/31", 1, []string{"10.0.0.0/32", "10.0.0.1/32"}, false},
		{"lower-size3", "10.0.0.0/22", 64, []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26", "10.0.1.0/26", "10.0.1.64/26", "10.0.1.128/26", "10.0.1.192/26", "10.0.2.0/26", "10.0.2.64/26", "10.0.2.128/26", "10.0.2.192/26", "10.0.3.0/26", "10.0.3.64/26", "10.0.3.128/26", "10.0.3.192/26"}, false},

		{"invalid-lower-size1", "10.0.0.0/32", 0, nil, true},

		/*
				Is 10.0.10.10/28 valid?
			    	✅ As a host IP in a /28 subnet: Yes, it is a valid IP within 10.0.10.0/28.
			    	❌ As a network address: No, 10.0.10.10 is not the base (network) address of the subnet. The network address would be 10.0.10.0/28.

				Nmap will scan every IP address for which the first <numbits> are the same as for the reference
				IP or hostname given. For example, 192.168.10.0/24 would scan the 256 hosts between 192.168.10.0
				(binary: 11000000 10101000 00001010 00000000) and 192.168.10.255 (binary: 11000000 10101000 00001010 11111111),
				inclusive. 192.168.10.40/24 would scan exactly the same targets
		*/
		{"lower-size-wrong-base1", "10.0.10.10/28", 4, []string{"10.0.10.0/30", "10.0.10.4/30", "10.0.10.8/30", "10.0.10.12/30"}, false},
		{"lower-size-wrong-base2", "10.0.1.2/30", 1, []string{"10.0.1.0/32", "10.0.1.1/32", "10.0.1.2/32", "10.0.1.3/32"}, false},

		{"invalid-input1", "sub.domain.tld", 0, nil, true},
		{"invalid-input2", "10.12.11.128", 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitNetworkIpV4(tt.network, tt.targetSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitNetworkIpV4() error = '%v', wantErr = '%v'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitNetworkIpV4() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
