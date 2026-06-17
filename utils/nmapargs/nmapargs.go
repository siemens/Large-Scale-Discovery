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

import (
	"fmt"
	"strconv"
	"strings"
)

// Result is the output of Validate, containing the validation verdict and any error messages.
type Result struct {
	Valid  bool
	Errors []string
}

// Validate parses and validates a raw Nmap argument string.
// It never invokes Nmap itself. Returns Valid=true only when the argument
// string represents a correct, non-contradictory Nmap invocation.
// Optional functional options may be passed to enable additional checks such as flag blocking.
func Validate(args string, optionFunctions ...OptionFunc) Result {

	// Apply all functional options to build the final configuration.
	var opts options
	for _, optionFunction := range optionFunctions {
		optionFunction(&opts)
	}

	// Lex the raw string into individual tokens.
	tokens, errLex := lex(args)
	if errLex != nil {
		return Result{Valid: false, Errors: []string{errLex.Error()}}
	}

	// Build a set of blocked flags from the resolved config for O(1) lookup.
	var blockedFlags = make(map[string]struct{})
	for _, flag := range opts.blockedFlags {
		blockedFlags[flag] = struct{}{}
	}

	// errs accumulates every validation problem found.
	var errs []string

	// flags tracks which Nmap flags appeared in the input.
	var flags = make(map[string]bool)

	// singleUseFlags lists flags that Nmap allows only once per invocation.
	singleUseFlags := map[string]bool{"-p": true, "--exclude-ports": true}

	// numericValues records parsed numeric argument values for later range checks.
	var numericValues = make(map[string]uint64)

	// floatValues stores parsed float arguments for rate range checks.
	var floatValues = make(map[string]float64)

	// timeValues stores parsed time-duration arguments (in milliseconds) for cross-unit range checks.
	var timeValues = make(map[string]float64)

	// Walk every token in the input and dispatch to the appropriate handler.
	i := 0
	for i < len(tokens) {
		token := tokens[i]

		// Skip empty tokens that may result from the lexer.
		if token == "" {
			i++
			continue
		}

		// Long flags (--flag or --flag=value).
		if strings.HasPrefix(token, "--") {

			// Split on the first '=' to detect inline values.
			name, inlineVal, hasInline := strings.Cut(token, "=")

			// Look up the flag definition in the registry.
			def, known := longFlags[name]
			if !known {
				errs = append(errs, fmt.Sprintf("unknown flag %q", name))
				i++
				continue
			}

			// Reject flags that are blocked by policy before any further processing.
			if _, blocked := blockedFlags[name]; blocked {
				errs = append(errs, fmt.Sprintf("'%s' is blocked by policy", name))

				// Skip the argument token if this flag expects one, to avoid a spurious bare-word error.
				if def.kind == argRequired && !hasInline && i+1 < len(tokens) && !isFlag(tokens[i+1]) {
					i++
				}
				i++
				continue
			}

			// Reject duplicate use of flags that Nmap allows only once.
			if singleUseFlags[name] && flags[name] {
				errs = append(errs, fmt.Sprintf("%s may only be specified once", name))
			}

			// Record that this flag was seen.
			flags[name] = true

			switch def.kind {

			case argNone:
				// A no-argument flag must not carry an inline value.
				if hasInline {
					errs = append(errs, fmt.Sprintf(
						"%s takes no argument, but = value %q was given", name, inlineVal,
					))
				}

			case argRequired:
				// Consume the argument from inline or the next token.
				val, ok := consumeArg(tokens, i, hasInline, inlineVal)
				if !ok {
					errs = append(errs, fmt.Sprintf("%s requires an argument", name))
					i++
					continue
				}

				// If the value came from the next token, advance the index.
				if !hasInline {
					i++
				}

				// Run the flag-specific validator on the argument value.
				if def.validator != nil {
					if validatorErr := def.validator(val); validatorErr != nil {
						errs = append(errs, fmt.Sprintf("%s: %s", name, validatorErr))
					}
				}

				// Store integer values so range checks can compare them later.
				if n, conversionErr := strconv.ParseUint(val, 10, 64); conversionErr == nil {
					numericValues[name] = n
				}

				// Store float values separately for rate range checks.
				if f, conversionErr := strconv.ParseFloat(val, 64); conversionErr == nil {
					floatValues[name] = f
				}

				// Store time-duration values for cross-unit range comparison.
				if reTime.MatchString(val) {
					if ms, timeErr := parseTimeDuration(val); timeErr == nil {
						timeValues[name] = ms
					}
				}

			case argOptional:
				// Validate the inline value if one was provided.
				if hasInline && inlineVal != "" && def.validator != nil {
					if validatorErr := def.validator(inlineVal); validatorErr != nil {
						errs = append(errs, fmt.Sprintf("%s: %s", name, validatorErr))
					}
				}
			}

			i++
			continue
		}

		// Short flags (-X, -XX, or -XXvalue).
		if strings.HasPrefix(token, "-") && len(token) >= 2 {

			// Match the token against the short flag registry.
			flagKey, gluedVal := matchShortFlag(token)
			if flagKey == "" {
				errs = append(errs, fmt.Sprintf("unknown flag %q", token))
				i++
				continue
			}

			// Look up the definition for further processing.
			def := shortFlags[flagKey]

			// Reject flags that are blocked by policy before any further processing.
			if _, blocked := blockedFlags[flagKey]; blocked {
				errs = append(errs, fmt.Sprintf("'%s' is blocked by policy", flagKey))

				// Skip the argument token if this flag expects one, to avoid a spurious bare-word error.
				if def.kind == argRequired && gluedVal == "" && i+1 < len(tokens) && !isFlag(tokens[i+1]) {
					i++
				}
				i++
				continue
			}

			// Reject duplicate use of flags that Nmap allows only once.
			if singleUseFlags[flagKey] && flags[flagKey] {
				errs = append(errs, fmt.Sprintf("%s may only be specified once", flagKey))
			}

			// Record the flag as seen.
			flags[flagKey] = true

			switch def.kind {

			case argNone:
				// A no-argument flag must not have trailing characters.
				if gluedVal != "" {
					errs = append(errs, fmt.Sprintf(
						"%s takes no argument, but got trailing %q", flagKey, gluedVal,
					))
				}

			case argRequired:
				// Determine the argument value from glued characters or the next token.
				var val string
				if gluedVal != "" {
					val = gluedVal
				} else if i+1 < len(tokens) && !isFlag(tokens[i+1]) {
					i++
					val = tokens[i]
				} else {
					errs = append(errs, fmt.Sprintf("%s requires an argument", flagKey))
					i++
					continue
				}

				// Run the flag-specific validator on the argument value.
				if def.validator != nil {
					if validatorErr := def.validator(val); validatorErr != nil {
						errs = append(errs, fmt.Sprintf("%s: %s", flagKey, validatorErr))
					}
				}

				// Store integer values for later range checks.
				if n, conversionErr := strconv.ParseUint(val, 10, 64); conversionErr == nil {
					numericValues[flagKey] = n
				}

				// Store float values separately for rate range checks.
				if f, conversionErr := strconv.ParseFloat(val, 64); conversionErr == nil {
					floatValues[flagKey] = f
				}

				// Store time-duration values for cross-unit range comparison.
				if reTime.MatchString(val) {
					if ms, timeErr := parseTimeDuration(val); timeErr == nil {
						timeValues[flagKey] = ms
					}
				}

			case argOptional:
				// Validate the glued value if one was provided.
				if gluedVal != "" && def.validator != nil {
					if validatorErr := def.validator(gluedVal); validatorErr != nil {
						errs = append(errs, fmt.Sprintf("%s: %s", flagKey, validatorErr))
					}
				}
			}

			i++
			continue
		}

		// Bare tokens are rejected — targets are supplied separately by the caller.
		errs = append(errs, fmt.Sprintf(
			"%q is not allowed, targets are specified separately",
			token,
		))
		i++
	}

	// Validate that every min/max pair satisfies min <= max.
	errs = append(errs, checkNumericRanges(numericValues, floatValues)...)

	// Validate that time-based min/max pairs are correctly ordered after unit conversion.
	errs = append(errs, checkTimeRanges(timeValues)...)

	// Run all conflict rules against the collected flag set.
	for _, rule := range allRules {
		errs = append(errs, rule(flags)...)
	}

	// Return the validation result.
	return Result{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}
