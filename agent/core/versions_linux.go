//go:build linux

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
	"os/exec"
	"strings"
	"time"

	"github.com/siemens/Large-Scale-Discovery/agent/config"
	"golang.org/x/sync/errgroup"
)

var toolDetectTimeout = 5 * time.Second // Maximum time to wait for a tool version command

// versions runs nmap and sslyze version detection in parallel using the binary paths from
// agent config and returns the discovered version strings. npcapVer is always empty on Linux.
// On Linux, sslyze is invoked as a Python module ("python -m sslyze"), mirroring setup_linux.go.
// All detections respect a 5-second timeout; empty strings are returned for any tool that is
// unavailable or fails to respond.
func versions(paths config.Paths) (versionNmap, versionNpcap, versionSslyze string) {

	// Run both detections in parallel to bound total startup delay
	var g errgroup.Group
	g.Go(func() error {
		versionNmap = detectNmap(paths.Nmap)
		return nil
	})
	g.Go(func() error {
		versionSslyze = detectSslyze(paths.Python)
		return nil
	})
	_ = g.Wait()

	// npcap is Windows-only; always empty on Linux
	versionNpcap = ""

	// Return nil as everything went fine
	return
}

// detectNmap runs "nmap --version" via the supplied binary path with a bounded timeout and
// returns the parsed nmap version string. Returns empty string when the binary is absent or times out.
func detectNmap(nmapPath string) string {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errRun := exec.CommandContext(ctx, nmapPath, "--version").Output()
	if errRun != nil {
		return ""
	}

	// Return nil as everything went fine
	return parseNmapVersion(string(out))
}

// detectSslyze runs "python -m sslyze --version" via the supplied python path with a bounded
// timeout and returns the parsed version string. Returns empty string when the binary is absent
// or times out.
func detectSslyze(pythonPath string) string {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errRun := exec.CommandContext(ctx, pythonPath, "-m", "sslyze", "--version").Output()
	if errRun != nil {
		return ""
	}

	// Return nil as everything went fine
	return parseSslyzeVersion(string(out))
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

// parseSslyzeVersion extracts the SSLyze version token from "sslyze --version" output.
// Expected output is a single line containing just the semver string, e.g. "5.2.0".
// Returns empty string when the output contains spaces (unexpected format) or is empty.
func parseSslyzeVersion(output string) string {

	// Take the first line and strip whitespace
	firstLine := strings.TrimSpace(strings.SplitN(output, "\n", 2)[0])

	// A bare version token has no spaces
	if firstLine != "" && !strings.ContainsAny(firstLine, " \t") {
		return firstLine
	}

	// Return empty string as version could not be determined
	return ""
}
