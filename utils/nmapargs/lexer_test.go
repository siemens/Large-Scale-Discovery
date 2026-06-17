/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package nmapargs_test

import "testing"

// TestLexerErrors verifies that malformed input strings are caught during lexing.
func TestLexerErrors(t *testing.T) {

	// Each case triggers a lexer-level error before flag validation begins.
	expectInvalid(t, `--script-args "unclosed`, "unclosed double quote")
	expectInvalid(t, `--script-args 'unclosed`, "unclosed single quote")
	expectInvalid(t, `-sS \`, "trailing backslash")
}

// TestLexerHappyPath verifies that well-formed input is tokenized correctly.
func TestLexerHappyPath(t *testing.T) {

	// Empty and whitespace-only inputs produce no tokens and should be accepted.
	expectValid(t, "")
	expectValid(t, "   ")
	expectValid(t, "\t\t")

	// Tabs as delimiters between valid flags.
	expectValid(t, "-sS\t-p\t22,80")

	// Backslash escapes produce literal characters inside values.
	expectValid(t, `-oN output\ file.txt`)

	// Double-quoted strings preserve internal spaces.
	expectValid(t, `--script-args "user=admin pass=test"`)

	// Single-quoted strings preserve internal spaces.
	expectValid(t, `--script-args 'user=admin pass=test'`)

	// Backslash escape inside double quotes.
	expectValid(t, `--script-args "path=C:\\tmp"`)
}
