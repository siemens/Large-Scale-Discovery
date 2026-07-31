/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/siemens/Large-Scale-Discovery/_test"
)

// TestMain initializes a shared in-process SQLite manager database and tears it down after all tests complete
func TestMain(m *testing.M) {

	// GetSettings sets the working directory to _test/ so all test-created files are isolated there.
	_test.GetSettings()

	// Create a per-run subdirectory within _test/ so the SQLite file is isolated and cleaned up on exit.
	tmpDir, errTmpDir := os.MkdirTemp(".", "lsd-manager-db-test-*")
	if errTmpDir != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not create temp dir: %v\n", errTmpDir)
		os.Exit(1)
	}
	if errChdir := os.Chdir(tmpDir); errChdir != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not enter temp dir: %v\n", errChdir)
		os.Exit(1)
	}

	// Open the manager database before running any tests
	if errOpen := Open(); errOpen != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not open manager database: %v\n", errOpen)
		os.Exit(1)
	}

	// Run the package tests before releasing database and filesystem resources
	code := m.Run()
	_ = Close()
	_ = os.Chdir("..")
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
