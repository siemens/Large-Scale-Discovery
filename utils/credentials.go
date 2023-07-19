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
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"
)

const AlphaNumCaseSymbol = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-_!?§$%&/()=*+#,;.:<>|°^"
const AlphaNumDash = "abcdefghijklmnopqrstuvwxyz0123456789-"
const AlphaNum = "abcdefghijklmnopqrstuvwxyz0123456789"

// ValidPassword checks whether a given string matches password requirements
func ValidPassword(
	password string,
	minLength int,
	requiresLower bool,
	requiresUpper bool,
	requiresNumber bool,
	requiresSymbol bool,
) bool {

	// Prepare flags
	hasLower := false
	hasUpper := false
	hasNumber := false
	hasSymbol := false

	// Count string length (characters, not bytes)
	n := 0
	for range password {
		n++
	}

	// Check min password length
	if n < minLength {
		return false
	}

	for _, char := range password {

		// Convert rune to string
		charStr := string(char)

		// Check if character is an integer
		if _, err := strconv.Atoi(charStr); err == nil {
			hasNumber = true
			continue
		}

		// Check if character is symbol
		if char < 'A' || char > 'z' {
			hasSymbol = true
			continue
		}

		// Check if character is lower case
		if char >= 'a' && char <= 'z' {
			hasLower = true
			continue
		}

		// Check if character is uppwer case
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			continue
		}
	}

	// Check if all prerequisites are fulfilled
	if requiresLower && !hasLower {
		return false
	}
	if requiresUpper && !hasUpper {
		return false
	}
	if requiresNumber && !hasNumber {
		return false
	}
	if requiresSymbol && !hasSymbol {
		return false
	}

	// Return true if no issue was found
	return true
}

// GenerateToken generates a random string based allowed letters and a given length
func GenerateToken(letters string, length int) (string, error) {

	// Prepare characters to use
	chars := []rune(letters)
	max := big.NewInt(int64(len(chars)))

	// Build random string
	var b strings.Builder
	for i := 0; i < length; i++ {

		// Get random int between 0 and len(chars)
		randInt, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		// Append randomly chosen rune to string builder
		b.WriteRune(chars[randInt.Int64()])
	}

	// Convert to string
	token := b.String()

	// Return random token
	return token, nil
}
