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
	"os"
	"testing"

	_test "github.com/siemens/Large-Scale-Discovery/_test"
	"github.com/siemens/Large-Scale-Discovery/log"
	"go.uber.org/zap/zapcore"
)

// TestMain initializes a minimal silent logger so that functions calling log.GetLogger() do not panic.
func TestMain(m *testing.M) {

	// GetSettings sets cwd to _test/ so all test-created files are isolated there
	_ = _test.GetSettings()

	// Initialize a minimal silent logger so functions calling log.GetLogger() do not panic
	_, _ = log.InitGlobalLogger(log.Settings{
		Console: &log.ConsoleHandler{
			Enabled: false,
			Level:   zapcore.InfoLevel,
		},
	})

	// Run all tests
	os.Exit(m.Run())
}
