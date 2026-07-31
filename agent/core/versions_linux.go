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
	"fmt"
	"os/exec"

	"github.com/siemens/Large-Scale-Discovery/agent/config"
	"golang.org/x/sync/errgroup"
)

// versions runs Nmap, SSLyze, and python version detection in parallel using the binary paths from
// agent config and returns the discovered version strings. Returns an error if any tool command
// fails to execute, which aborts agent startup. Npcap is not required on Linux, so versionNpcap
// is always empty.
func versions(paths config.Paths) (versionNmap, versionNpcap, versionSslyze, versionPython string, err error) {

	// Run all detections in parallel to bound total startup delay
	var g errgroup.Group
	var errVersionNmap, errVersionSslyze, errVersionPython error
	g.Go(func() error {
		versionNmap, errVersionNmap = detectNmap(paths.Nmap)
		return nil // collect errors below to run all detections
	})
	g.Go(func() error {
		versionSslyze, errVersionSslyze = detectSslyze(paths.Python)
		return nil // collect errors below to run all detections
	})
	g.Go(func() error {
		versionPython, errVersionPython = detectPython(paths.Python)
		return nil // collect errors below to run all detections
	})
	_ = g.Wait()

	// Return error if something went wrong
	if errVersionNmap != nil {
		return "", "", "", "", errVersionNmap
	}
	if errVersionSslyze != nil {
		return "", "", "", "", errVersionSslyze
	}
	if errVersionPython != nil {
		return "", "", "", "", errVersionPython
	}

	// Return detected versions
	return versionNmap, "", versionSslyze, versionPython, nil
}

// detectSslyze runs "python -m sslyze -h" via the supplied python path with a bounded
// timeout and returns the parsed version string. Returns an error if the command fails to execute.
func detectSslyze(pythonPath string) (string, error) {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errOut := exec.CommandContext(ctx, pythonPath, "-m", "sslyze", "-h").Output()
	if errOut != nil {
		return "", fmt.Errorf("could not run SSLyze: %s", errOut)
	}

	// Parse and validate version from command output
	version := parseSslyzeVersion(string(out))

	// Return detected version
	return version, nil
}

// detectPython runs "<pythonPath> --version" with a bounded timeout and returns the parsed
// Python version string. Returns an error if the command fails to execute.
func detectPython(pythonPath string) (string, error) {

	// Run with a bounded context to prevent blocking agent startup
	ctx, cancel := context.WithTimeout(context.Background(), toolDetectTimeout)
	defer cancel()
	out, errOut := exec.CommandContext(ctx, pythonPath, "--version").Output()
	if errOut != nil {
		return "", fmt.Errorf("could not run Python: %s", errOut)
	}

	// Parse and validate version from command output
	version := parsePythonVersion(string(out))

	// Return detected version
	return version, nil
}
