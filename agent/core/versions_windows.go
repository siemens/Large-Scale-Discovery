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
	"fmt"
	"os/exec"

	"github.com/siemens/Large-Scale-Discovery/agent/config"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows/registry"
)

// versions runs Nmap, Npcap, and SSLyze version detection in parallel using the binary paths from
// agent config and returns the discovered version strings. Returns an error if any tool command
// fails to execute, which aborts agent startup. Python is not required on Windows, so versionPython
// is always empty.
func versions(paths config.Paths) (versionNmap, versionNpcap, versionSslyze, versionPython string, err error) {

	// Run all detections in parallel to bound total startup delay
	var g errgroup.Group
	var errVersionNmap, errVersionSslyze error
	g.Go(func() error {
		versionNmap, errVersionNmap = detectNmap(paths.Nmap)
		return nil // collect errors below to run all detections
	})
	g.Go(func() error {
		versionSslyze, errVersionSslyze = detectSslyze(paths.Sslyze)
		return nil // collect errors below to run all detections
	})

	// Read the optional registry value while the external commands are running
	versionNpcap = detectNpcap()
	_ = g.Wait()

	// Return error if something went wrong
	if errVersionNmap != nil {
		return "", "", "", "", errVersionNmap
	}
	if errVersionSslyze != nil {
		return "", "", "", "", errVersionSslyze
	}

	// Return detected versions
	return versionNmap, versionNpcap, versionSslyze, "", nil
}

// detectNpcap reads the Npcap display version from its Windows uninstall entry
func detectNpcap() string {

	// Check both registry locations used by 32-bit and native Npcap installers
	paths := []string{
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\NpcapInst`,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NpcapInst`,
	}
	for _, path := range paths {

		// Open the uninstall entry and try the next registry location if unavailable
		key, errKey := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if errKey != nil {
			continue
		}

		// Read the version shown by Windows Apps and Features
		version, _, errGet := key.GetStringValue("DisplayVersion")
		_ = key.Close()
		if errGet == nil && version != "" {
			return version
		}
	}

	// Return an empty version when Npcap is absent or has no uninstall metadata
	return ""
}

// detectSslyze runs "sslyze -h" via the supplied executable path with a bounded timeout
// and returns the parsed version string. Returns an error if the command fails to execute.
func detectSslyze(sslyzePath string) (string, error) {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errOut := exec.CommandContext(ctx, sslyzePath, "-h").Output()
	if errOut != nil {
		return "", fmt.Errorf("could not run SSLyze: %s", errOut)
	}

	// Parse and validate version from command output
	version := parseSslyzeVersion(string(out))

	// Return detected version
	return version, nil
}
