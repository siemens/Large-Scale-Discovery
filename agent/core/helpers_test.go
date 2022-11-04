/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"reflect"
	"testing"
)

// An assertion function to be used in testing
func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if reflect.DeepEqual(a, b) {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

// TestInputResolve tests the inputResolve function against expected results
func TestInputResolve(t *testing.T) {

	// Prepare test inputs
	tests := []string{"127.0.0.1", "2001:db8:3333:4444:5555:6666:7777:8888", "::1234:5678", "10.10.10.0/24",
		"2001:db8:abcd:0012::0/64", "nonexisting.host.com", "text-lb.esams.wikimedia.org", "invalid input"}

	// Prepare expected results
	type checkResults struct {
		IPv4s []string
		IPv6s []string
	}
	expectedResults := []checkResults{{IPv4s: []string{"127.0.0.1"}, IPv6s: nil},
		{IPv4s: nil, IPv6s: []string{"2001:db8:3333:4444:5555:6666:7777:8888"}},
		{IPv4s: nil, IPv6s: []string{"::1234:5678"}},
		{IPv4s: []string{"10.10.10.0/24"}, IPv6s: nil},
		{IPv4s: nil, IPv6s: []string{"2001:db8:abcd:0012::0/64"}},
		{IPv4s: nil, IPv6s: nil},
		{IPv4s: []string{"91.198.174.192"}, IPv6s: nil},
		{IPv4s: nil, IPv6s: nil},
	}

	// Run test checks
	for i, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			// Run function
			ipv4s, ipv6s := resolveInput(tt)

			// Test asserts
			assertEqual(t, ipv4s, expectedResults[i].IPv4s, fmt.Sprintf("Wrong resolved IPv4s for %s", tt))
			assertEqual(t, ipv6s, expectedResults[i].IPv6s, fmt.Sprintf("Wrong resolved IPv6s for %s", tt))
		})
	}
}
