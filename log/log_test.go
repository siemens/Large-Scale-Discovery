/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package log

import (
	"os"
	"testing"

	"github.com/siemens/Large-Scale-Discovery/_test"
	"go.uber.org/zap/zapcore"
)

var settings = Settings{
	Console: &ConsoleHandler{
		Level: zapcore.DebugLevel,
	},
	File: &FileHandler{
		Level:   zapcore.DebugLevel,
		Path:    "mytest.log",
		SizeMb:  100,
		History: 10,
	},
	Smtp: nil,
}

// TestNewLogger checks the basic functionality of the logger.
func TestNewLogger(t *testing.T) {

	// Retrieve test settings once to set working directory
	_ = _test.GetSettings()

	// Prepare cleanup
	defer func() { _ = os.Remove(settings.File.Path) }()

	// Get new independent (NOT THE GLOBAL) logger
	testLogger, err := InitGlobalLogger(settings)
	if err != nil {
		t.Errorf("could not initialize global logger")
		return
	}

	defer func() {
		errClose := CloseGlobalLogger()
		if errClose != nil {
			t.Errorf("could not close global logger")
		}
	}()

	// Send some test log messages
	for i := 0; i < 10; i++ {
		go func(i int) { testLogger.Debugf("Debug message async %d", i) }(i)
	}
	testLogger.Debugf("Debug message.")
	testLogger.Infof("Info message.")
	func(logger Logger) {
		taggedTestLogger := logger.Tagged("tagged logger")
		taggedTestLogger.Debugf("Tagged debug message.")
	}(testLogger)
	testLogger.Warningf("Warning message.")
	testLogger.Errorf("Error message.")
}
