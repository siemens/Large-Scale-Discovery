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
			v, errV := strconv.Atoi(strings.ToLower(splits[i+1]))
			if errV == nil {

				// Return value from arguments
				return v
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
				ms, errMs := strconv.Atoi(strings.ReplaceAll(hostTimeout, "ms", ""))
				if errMs == nil {

					// Return value from arguments in minutes
					return float64(ms) / 1000 / 60
				}
			} else if strings.HasSuffix(hostTimeout, "s") {
				s, errS := strconv.Atoi(strings.ReplaceAll(hostTimeout, "s", ""))
				if errS == nil {

					// Return value from arguments in minutes
					return float64(s) / 60
				}
			} else if strings.HasSuffix(hostTimeout, "m") {
				m, errM := strconv.Atoi(strings.ReplaceAll(hostTimeout, "m", ""))
				if errM == nil {

					// Return value from arguments in minutes
					return float64(m)
				}
			} else if strings.HasSuffix(hostTimeout, "h") {
				h, errH := strconv.Atoi(strings.ReplaceAll(hostTimeout, "h", ""))
				if errH == nil {

					// Return value from arguments in minutes
					return float64(h) * 60
				}
			} else { // no unit means seconds
				s, errS := strconv.Atoi(hostTimeout)
				if errS == nil {

					// Return value from arguments in minutes
					return float64(s) / 60
				}
			}
		}
	}

	// Return empty value as value is not set in arguments
	return 0
}
