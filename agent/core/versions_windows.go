//go:build windows

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
	"golang.org/x/sys/windows/registry"
)

var toolDetectTimeout = 5 * time.Second // Maximum time to wait for a tool version command

// versions runs nmap, npcap, and sslyze version detection in parallel using the binary paths
// from agent config and returns the discovered version strings. npcap is detected via the
// Windows registry following the GoScans discovery.CheckNpcap() pattern.
// All detections respect a 5-second timeout; empty strings are returned for any tool that is
// unavailable or fails to respond.
func versions(paths config.Paths) (versionNmap, versionNpcap, versionSslyze string) {

	// Run all detections in parallel to bound total startup delay
	var g errgroup.Group
	g.Go(func() error {
		versionNmap = detectNmap(paths.Nmap)
		return nil
	})
	g.Go(func() error {
		versionNpcap = detectNpcap()
		return nil
	})
	g.Go(func() error {
		versionSslyze = detectSslyze(paths.Sslyze)
		return nil
	})
	_ = g.Wait()

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

// detectNpcap reads the Npcap version from the Windows registry, following the GoScans
// discovery.CheckNpcap() pattern (SOFTWARE\WOW6432Node\Npcap and SOFTWARE\Npcap).
// Returns empty string when Npcap is not installed.
func detectNpcap() string {

	// Try the WOW6432Node path first, then the native path, mirroring Npcap's own installer
	paths := []string{
		`SOFTWARE\WOW6432Node\Npcap`,
		`SOFTWARE\Npcap`,
	}
	for _, path := range paths {

		// Open the registry key; skip on error and try the next path
		k, errOpen := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if errOpen != nil {
			continue
		}

		// Read the default value, which holds the version string
		val, _, errGet := k.GetStringValue("")
		_ = k.Close()
		if errGet == nil && val != "" {
			return val
		}
	}

	// Return empty string as Npcap was not found in the registry
	return ""
}

// detectSslyze runs "sslyze --version" via the supplied executable path with a bounded timeout
// and returns the parsed version string. Returns empty string when the binary is absent or times out.
func detectSslyze(sslyzePath string) string {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errRun := exec.CommandContext(ctx, sslyzePath, "--version").Output()
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
