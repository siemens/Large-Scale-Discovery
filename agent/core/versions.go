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
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// toolDetectTimeout limits how long external tool-version detection may take
var toolDetectTimeout = 5 * time.Second

// detectNmap runs "nmap --version" via the supplied binary path with a bounded timeout and
// returns the parsed version string. Returns an error if the command fails to execute.
func detectNmap(nmapPath string) (string, error) {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errOut := exec.CommandContext(ctx, nmapPath, "--version").Output()
	if errOut != nil {
		return "", fmt.Errorf("could not run Nmap: %s", errOut)
	}

	// Parse and validate version from command output
	version := parseNmapVersion(string(out))

	// Return detected version
	return version, nil
}

// parseNmapVersion extracts the Nmap version token from "nmap --version" output.
// Expected first line format: "Nmap version X.YZ ( https://nmap.org )".
// Returns empty string when the output does not match the expected format.
func parseNmapVersion(output string) string {

	// Only the first line carries the version
	firstLine := strings.SplitN(output, "\n", 2)[0]
	fields := strings.Fields(firstLine)

	// Expect at least "Nmap version X.YZ"
	if len(fields) >= 3 &&
		strings.EqualFold(fields[0], "Nmap") &&
		strings.EqualFold(fields[1], "version") {
		return fields[2]
	}

	// Return empty string as version could not be determined
	return ""
}

// parseSslyzeVersion extracts the SSLyze version token from "sslyze -h" output.
// Searches every line for the argparse description anchor "SSLyze version X.Y.Z" emitted by
// CommandLineParser regardless of its position — the anchor sits below the usage block, so the
// parser cannot rely on a fixed line number, and this is the only externally observable version
// surface across all 5.x/6.x releases (--version does not exist).
// Returns empty string on unparseable or empty input.
func parseSslyzeVersion(output string) string {

	// Search every line for the description that argparse emits, e.g. "SSLyze version 5.1.3"
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 &&
			strings.EqualFold(fields[0], "SSLyze") &&
			strings.EqualFold(fields[1], "version") {
			return fields[2]
		}
	}

	// Return empty string as version could not be determined
	return ""
}

// parsePythonVersion extracts the Python version token from "python --version" output.
// Expected format: "Python X.Y.Z" (possibly followed by additional text).
// Returns empty string when the output does not match the expected format.
func parsePythonVersion(output string) string {

	// Only the first line carries the version
	firstLine := strings.SplitN(output, "\n", 2)[0]
	fields := strings.Fields(firstLine)

	// Expect at least "Python X.Y.Z"
	if len(fields) >= 2 && strings.EqualFold(fields[0], "Python") {
		return fields[1]
	}

	// Return empty string as version could not be determined
	return ""
}
