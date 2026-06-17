/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package nmapargs

import "testing"

// TestIsFlag verifies that isFlag correctly distinguishes flags from non-flag tokens.
func TestIsFlag(t *testing.T) {

	// Cases where the token is a valid flag prefix.
	flagCases := []struct {
		name  string
		input string
	}{
		{"long flag", "--top-ports"},
		{"short lowercase", "-p"},
		{"short uppercase", "-T"},
		{"double dash prefix", "--"},
		{"short two-char", "-sS"},
		{"three-char output", "-oN"},
	}

	// Each flag case must return true.
	for _, tc := range flagCases {
		t.Run(tc.name, func(t *testing.T) {
			if !isFlag(tc.input) {
				t.Errorf("isFlag(%q) = false, want true", tc.input)
			}
		})
	}

	// Cases where the token is not a flag.
	nonFlagCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"single dash", "-"},
		{"negative number", "-1"},
		{"negative large number", "-443"},
		{"bare word", "target.example.com"},
		{"number", "80"},
		{"dash digit", "-0"},
	}

	// Each non-flag case must return false.
	for _, tc := range nonFlagCases {
		t.Run(tc.name, func(t *testing.T) {
			if isFlag(tc.input) {
				t.Errorf("isFlag(%q) = true, want false", tc.input)
			}
		})
	}
}

// TestMatchShortFlag verifies that matchShortFlag finds the longest matching key
// and correctly splits the glued value.
func TestMatchShortFlag(t *testing.T) {

	// Table of inputs with expected flag key and glued value.
	cases := []struct {
		name     string
		token    string
		wantKey  string
		wantGlue string
	}{
		{"three-char match oN", "-oN", "-oN", ""},
		{"three-char match with glued value", "-oN/tmp/out.txt", "-oN", "/tmp/out.txt"},
		{"two-char match sS", "-sS", "-sS", ""},
		{"two-char match with glued value", "-p22,80", "-p", "22,80"},
		{"optional flag no value", "-PS", "-PS", ""},
		{"optional flag with ports", "-PS22,443", "-PS", "22,443"},
		{"timing flag glued", "-T4", "-T", "4"},
		{"unknown flag", "-zQ", "", ""},
		{"single char not in registry", "-Z", "", ""},
	}

	// Run each case and compare key and glued value.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// Match the token against the short flag registry.
			gotKey, gotGlue := matchShortFlag(tc.token)
			if gotKey != tc.wantKey || gotGlue != tc.wantGlue {
				t.Errorf("matchShortFlag(%q) = (%q, %q), want (%q, %q)",
					tc.token, gotKey, gotGlue, tc.wantKey, tc.wantGlue)
			}
		})
	}
}

// TestConsumeArg verifies that consumeArg returns the correct argument value
// from inline syntax, the next token, or reports missing.
func TestConsumeArg(t *testing.T) {

	// Inline value via '=' should be returned directly.
	t.Run("inline value", func(t *testing.T) {

		// Simulate --top-ports=200 where hasInline is true.
		val, ok := consumeArg([]string{"--top-ports=200"}, 0, true, "200")
		if !ok || val != "200" {
			t.Errorf("consumeArg(inline) = (%q, %v), want (\"200\", true)", val, ok)
		}
	})

	// Next token is a non-flag value.
	t.Run("next token is value", func(t *testing.T) {

		// Simulate --top-ports 100 where the value is the next token.
		tokens := []string{"--top-ports", "100", "-sS"}
		val, ok := consumeArg(tokens, 0, false, "")
		if !ok || val != "100" {
			t.Errorf("consumeArg(next token) = (%q, %v), want (\"100\", true)", val, ok)
		}
	})

	// Next token is itself a flag — no argument available.
	t.Run("next token is flag", func(t *testing.T) {

		// Simulate --top-ports -sS where the next token looks like a flag.
		tokens := []string{"--top-ports", "-sS"}
		val, ok := consumeArg(tokens, 0, false, "")
		if ok {
			t.Errorf("consumeArg(next is flag) = (%q, true), want (\"\", false)", val)
		}
	})

	// No next token exists — last token in the slice.
	t.Run("no next token", func(t *testing.T) {

		// Simulate --top-ports at the end of input with nothing following.
		tokens := []string{"--top-ports"}
		val, ok := consumeArg(tokens, 0, false, "")
		if ok {
			t.Errorf("consumeArg(no next) = (%q, true), want (\"\", false)", val)
		}
	})

	// Next token is a negative number, which is not a flag.
	t.Run("next token is negative number", func(t *testing.T) {

		// Simulate --data-length -1 where -1 is not a flag.
		tokens := []string{"--data-length", "-1"}
		val, ok := consumeArg(tokens, 0, false, "")
		if !ok || val != "-1" {
			t.Errorf("consumeArg(negative number) = (%q, %v), want (\"-1\", true)", val, ok)
		}
	})

	// Inline empty string is still a valid consumed value.
	t.Run("inline empty value", func(t *testing.T) {

		// Simulate --flag= where hasInline is true but value is empty.
		val, ok := consumeArg([]string{"--flag="}, 0, true, "")
		if !ok || val != "" {
			t.Errorf("consumeArg(inline empty) = (%q, %v), want (\"\", true)", val, ok)
		}
	})
}

// TestCheckNumericRanges verifies that checkNumericRanges detects inverted min/max pairs
// and accepts valid or absent pairs.
func TestCheckNumericRanges(t *testing.T) {

	// Valid range where min < max produces no errors.
	t.Run("valid min less than max", func(t *testing.T) {

		// Set --min-rate below --max-rate.
		vals := map[string]float64{"--min-rate": 10, "--max-rate": 500}
		errs := checkNumericRanges(nil, vals)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	// Equal values are allowed (min == max).
	t.Run("equal values allowed", func(t *testing.T) {

		// Set --min-parallelism equal to --max-parallelism.
		vals := map[string]uint64{"--min-parallelism": 10, "--max-parallelism": 10}
		errs := checkNumericRanges(vals, nil)
		if len(errs) != 0 {
			t.Errorf("expected no errors for equal values, got %v", errs)
		}
	})

	// Inverted range where min > max produces an error.
	t.Run("inverted range", func(t *testing.T) {

		// Set --min-rate above --max-rate.
		vals := map[string]float64{"--min-rate": 500, "--max-rate": 10}
		errs := checkNumericRanges(nil, vals)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
	})

	// Only one side of a pair is present — no error expected.
	t.Run("only min present", func(t *testing.T) {

		// Set --min-hostgroup without its counterpart.
		vals := map[string]uint64{"--min-hostgroup": 100}
		errs := checkNumericRanges(vals, nil)
		if len(errs) != 0 {
			t.Errorf("expected no errors when only min present, got %v", errs)
		}
	})

	// Empty map produces no errors.
	t.Run("empty map", func(t *testing.T) {

		// No numeric values provided at all.
		errs := checkNumericRanges(map[string]uint64{}, map[string]float64{})
		if len(errs) != 0 {
			t.Errorf("expected no errors for empty map, got %v", errs)
		}
	})

	// Multiple inverted pairs produce multiple errors.
	t.Run("multiple inverted pairs", func(t *testing.T) {

		// Invert both parallelism and hostgroup pairs.
		vals := map[string]uint64{
			"--min-parallelism": 500,
			"--max-parallelism": 10,
			"--min-hostgroup":   200,
			"--max-hostgroup":   50,
		}
		errs := checkNumericRanges(vals, nil)
		if len(errs) != 2 {
			t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
		}
	})
}

// TestParseTimeDuration verifies the Nmap time string to millisecond conversion.
func TestParseTimeDuration(t *testing.T) {

	// Table of inputs with expected millisecond values.
	cases := []struct {
		name    string
		input   string
		wantMs  float64
		wantErr bool
	}{
		{"milliseconds", "100ms", 100, false},
		{"seconds", "3s", 3000, false},
		{"minutes", "2m", 120000, false},
		{"hours", "1h", 3600000, false},
		{"bare number defaults to seconds", "5", 5000, false},
		{"fractional seconds", "0.5s", 500, false},
		{"fractional hours", "0.5h", 1800000, false},
		{"invalid input", "abc", 0, true},
	}

	// Run each case and compare the result.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// Parse the time duration string.
			gotMs, gotErr := parseTimeDuration(tc.input)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("parseTimeDuration(%q) error = %v, wantErr = %v", tc.input, gotErr, tc.wantErr)
			}

			// Compare the millisecond value when no error is expected.
			if !tc.wantErr && gotMs != tc.wantMs {
				t.Errorf("parseTimeDuration(%q) = %g, want %g", tc.input, gotMs, tc.wantMs)
			}
		})
	}
}

// TestCheckTimeRanges verifies that time-based min/max range checks work correctly
// after converting values to a common unit.
func TestCheckTimeRanges(t *testing.T) {

	// Valid range where min < max produces no errors.
	t.Run("valid ordering", func(t *testing.T) {

		// 100ms < 5000ms (5s).
		vals := map[string]float64{"--min-rtt-timeout": 100, "--max-rtt-timeout": 5000}
		errs := checkTimeRanges(vals)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	// Equal values are allowed.
	t.Run("equal values", func(t *testing.T) {

		// Both set to 1000ms.
		vals := map[string]float64{"--min-rtt-timeout": 1000, "--max-rtt-timeout": 1000}
		errs := checkTimeRanges(vals)
		if len(errs) != 0 {
			t.Errorf("expected no errors for equal values, got %v", errs)
		}
	})

	// Inverted range where min > max produces an error.
	t.Run("inverted range", func(t *testing.T) {

		// 5000ms > 100ms.
		vals := map[string]float64{"--min-rtt-timeout": 5000, "--max-rtt-timeout": 100}
		errs := checkTimeRanges(vals)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
	})

	// Only one side present produces no errors.
	t.Run("only min present", func(t *testing.T) {

		// Only --scan-delay set.
		vals := map[string]float64{"--scan-delay": 500}
		errs := checkTimeRanges(vals)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	// Empty map produces no errors.
	t.Run("empty map", func(t *testing.T) {

		// No time values provided at all.
		errs := checkTimeRanges(map[string]float64{})
		if len(errs) != 0 {
			t.Errorf("expected no errors for empty map, got %v", errs)
		}
	})
}
