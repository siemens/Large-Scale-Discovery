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

import "testing"

func TestValidPassword(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		password       string
		minLength      int
		requiresLower  bool
		requiresUpper  bool
		requiresNumber bool
		requiresSymbol bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"length", args{"-", 1, false, false, false, false}, true},
		{"lower", args{"a", 1, true, false, false, false}, true},
		{"upper", args{"A", 1, false, true, false, false}, true},
		{"number", args{"0", 1, false, false, true, false}, true},
		{"symbol", args{"!", 1, false, false, false, true}, true},
		{"all+length", args{"aA0!", 4, true, true, true, true}, true},
		{"all", args{"aA0!", 0, true, true, true, true}, true},
		{"all+chinese", args{"aA0!", 0, true, true, true, true}, true},

		{"missing-lower", args{"A0!", 0, true, true, true, true}, false},
		{"missing-upper", args{"a0!", 0, true, true, true, true}, false},
		{"missing-number", args{"aA!", 0, true, true, true, true}, false},
		{"missing-symbol", args{"aA0", 0, true, true, true, true}, false},
		{"missing-length", args{"aA0!", 5, true, true, true, true}, false},

		{"missing-lower-chinese", args{"A0本", 0, true, true, true, true}, false},
		{"missing-upper-chinese", args{"a0本", 0, true, true, true, true}, false},
		{"missing-number-chinese", args{"aA本", 0, true, true, true, true}, false},
		{"missing-symbol-chinese", args{"aA0", 0, true, true, true, true}, false},
		{"missing-length-chinese", args{"aA0本", 5, true, true, true, true}, false},

		{"long-none-required", args{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 0, false, false, false, false}, true},
		{"long-all-required", args{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 10, true, true, true, true}, true},

		{"symbol-detection1", args{"!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 0, true, false, false, false}, false},
		{"symbol-detection2", args{"!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 0, false, true, false, false}, false},
		{"symbol-detection3", args{"!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 0, false, false, true, false}, false},
		{"symbol-detection4", args{"!\"§$%&/()=?`´+*~#'-_.:,;µ><|^°本", 0, false, false, false, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPassword(tt.args.password, tt.args.minLength, tt.args.requiresLower, tt.args.requiresUpper, tt.args.requiresNumber, tt.args.requiresSymbol); got != tt.want {
				t.Errorf("ValidPassword() = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}
