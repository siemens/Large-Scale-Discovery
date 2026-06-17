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

	_test "github.com/siemens/Large-Scale-Discovery/_test"
)

// TestMain initializes a shared in-process SQLite manager DB and tears it down after all tests complete.
func TestMain(m *testing.M) {

	// GetSettings sets the working directory to _test/ so all test-created files are isolated there.
	_test.GetSettings()

	// Create a per-run subdirectory within _test/ so the SQLite file is isolated and cleaned up on exit.
	tmpDir, errTmp := os.MkdirTemp(".", "lsd-manager-db-test-*")
	if errTmp != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not create temp dir: %v\n", errTmp)
		os.Exit(1)
	}
	_ = os.Chdir(tmpDir)

	// Open the manager DB before running any tests.
	if errOpen := OpenManagerDb(); errOpen != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not open manager db: %v\n", errOpen)
		os.Exit(1)
	}

	code := m.Run()

	_ = CloseManagerDb()
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
