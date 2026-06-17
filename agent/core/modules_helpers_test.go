/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"testing"
)

// TestClassifyInput verifies classifyInput categorises IP addresses, subnets, hostnames, and invalid inputs correctly.
func TestClassifyInput(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name      string
		input     string
		wantIPv4s bool
		wantIPv6s bool
	}{
		{
			name:      "ipv4-address",
			input:     "127.0.0.1",
			wantIPv4s: true,
			wantIPv6s: false,
		},
		{
			name:      "ipv6-address",
			input:     "2001:db8:3333:4444:5555:6666:7777:8888",
			wantIPv4s: false,
			wantIPv6s: true,
		},
		{
			name:      "ipv6-short",
			input:     "::1234:5678",
			wantIPv4s: false,
			wantIPv6s: true,
		},
		{
			name:      "ipv4-subnet",
			input:     "10.10.10.0/24",
			wantIPv4s: true,
			wantIPv6s: false,
		},
		{
			name:      "ipv6-subnet",
			input:     "2001:db8:abcd:0012::0/64",
			wantIPv4s: false,
			wantIPv6s: true,
		},
		{
			name:      "nonexisting-host",
			input:     "nonexisting.host.com",
			wantIPv4s: false,
			wantIPv6s: false,
		},
		{
			name:      "invalid-input",
			input:     "invalid input",
			wantIPv4s: false,
			wantIPv6s: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIPv4s, gotIPv6s := classifyInput(tt.input)
			if gotIPv4s != tt.wantIPv4s {
				t.Errorf("classifyInput() IPv4s = '%v', want = '%v'", gotIPv4s, tt.wantIPv4s)
			}
			if gotIPv6s != tt.wantIPv6s {
				t.Errorf("classifyInput() IPv6s = '%v', want = '%v'", gotIPv6s, tt.wantIPv6s)
			}
		})
	}

	// DNS resolution is environment-specific (IPv4-only vs dual-stack). Only assert that
	// a real hostname resolves to at least one address family.
	t.Run("existing-host", func(t *testing.T) {
		gotIPv4s, gotIPv6s := classifyInput("text-lb.esams.wikimedia.org")
		if !gotIPv4s && !gotIPv6s {
			t.Errorf("classifyInput(%q) = IPv4s:%v IPv6s:%v, want at least one true", "text-lb.esams.wikimedia.org", gotIPv4s, gotIPv6s)
		}
	})
}
