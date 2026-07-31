/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"strconv"
	"strings"
)

const defaultNmapMinHostgroup = 64            // Default minimum amount of IPs for Nmap to be processed in parallel. Used if none is specified in the Nmap arguments.
const defaultNmapHosttimeoutMinutes = 60 * 12 // Default maximum amount of time in hours for Nmap for a single IP/hostgroup. Used if none is specified in the Nmap arguments.

// Extract '--min-hostgroup' value from Nmap argument string, if present, or return 0
func extractMinHostgroup(args string) int {

	// Read Nmap args into slice
	splits := strings.Split(args, " ")

	// Iterate slice and find value for '--min-hostgroup'
	for i := 0; i < len(splits); i++ {
		if strings.ToLower(splits[i]) == "--min-hostgroup" {

			// Convert value to integer
			hostGroup, errHostGroup := strconv.Atoi(strings.ToLower(splits[i+1]))
			if errHostGroup == nil {

				// Return value from arguments
				return hostGroup
			}
		}
	}

	// Return empty value as value is not set in arguments
	return 0
}

// Extract '--host-timeout' value from Nmap argument string, if present, or return 0
func extractHostTimeoutMinutes(args string) float64 {

	// Read Nmap args into slice
	splits := strings.Split(args, " ")

	// Iterate slice and find value for '--host-timeout'
	for i := 0; i < len(splits); i++ {
		if strings.ToLower(splits[i]) == "--host-timeout" {

			// Extract host timeout value in Nmap format
			hostTimeout := strings.ToLower(splits[i+1])

			// Convert hot timeout value from Nmap format to integer minutes
			if strings.HasSuffix(hostTimeout, "ms") {
				milliseconds, errMilliseconds := strconv.Atoi(strings.ReplaceAll(hostTimeout, "ms", ""))
				if errMilliseconds == nil {

					// Return value from arguments in minutes
					return float64(milliseconds) / 1000 / 60
				}
			} else if strings.HasSuffix(hostTimeout, "s") {
				seconds, errSeconds := strconv.Atoi(strings.ReplaceAll(hostTimeout, "s", ""))
				if errSeconds == nil {

					// Return value from arguments in minutes
					return float64(seconds) / 60
				}
			} else if strings.HasSuffix(hostTimeout, "m") {
				minutes, errMinutes := strconv.Atoi(strings.ReplaceAll(hostTimeout, "m", ""))
				if errMinutes == nil {

					// Return value from arguments in minutes
					return float64(minutes)
				}
			} else if strings.HasSuffix(hostTimeout, "h") {
				hours, errHours := strconv.Atoi(strings.ReplaceAll(hostTimeout, "h", ""))
				if errHours == nil {

					// Return value from arguments in minutes
					return float64(hours) * 60
				}
			} else { // no unit means seconds
				seconds, errSeconds := strconv.Atoi(hostTimeout)
				if errSeconds == nil {

					// Return value from arguments in minutes
					return float64(seconds) / 60
				}
			}
		}
	}

	// Return empty value as value is not set in arguments
	return 0
}
