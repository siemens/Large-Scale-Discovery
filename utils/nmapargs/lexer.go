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

import "fmt"

// lex splits a raw Nmap argument string into tokens,
// respecting single quotes, double quotes, and backslash escapes.
// Returns an error if the input contains an unclosed quote or a trailing backslash.
func lex(input string) ([]string, error) {

	// tokens holds the final list of split arguments.
	var tokens []string

	// current accumulates runes for the token being built.
	var current []rune

	// inSingle, inDouble, and escaped track the current quoting/escape state.
	inSingle, inDouble, escaped := false, false, false

	// Walk every rune in the input and classify it.
	for _, ch := range input {
		switch {

		// If the previous rune was a backslash, this rune is always literal.
		case escaped:
			current = append(current, ch)
			escaped = false

		// A backslash outside single quotes begins an escape sequence.
		case ch == '\\' && !inSingle:
			escaped = true

		// A single quote outside double quotes toggles single-quote mode.
		case ch == '\'' && !inDouble:
			inSingle = !inSingle

		// A double quote outside single quotes toggles double-quote mode.
		case ch == '"' && !inSingle:
			inDouble = !inDouble

		// Unquoted whitespace is a token delimiter, flush current if non-empty.
		case (ch == ' ' || ch == '\t') && !inSingle && !inDouble:
			if len(current) > 0 {
				tokens = append(tokens, string(current))
				current = current[:0]
			}

		// Everything else is a literal character for the current token.
		default:
			current = append(current, ch)
		}
	}

	// An unclosed single quote is a syntax error.
	if inSingle {
		return nil, fmt.Errorf("unclosed single quote")
	}

	// An unclosed double quote is a syntax error.
	if inDouble {
		return nil, fmt.Errorf("unclosed double quote")
	}

	// A trailing backslash with no following character is a syntax error.
	if escaped {
		return nil, fmt.Errorf("trailing backslash")
	}

	// Flush any remaining runes as the final token.
	if len(current) > 0 {
		tokens = append(tokens, string(current))
	}

	// Return nil as everything went fine.
	return tokens, nil
}
