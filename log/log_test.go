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
	"github.com/siemens/Large-Scale-Discovery/_test"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"go.uber.org/zap/zapcore"
	"net/mail"
	"os"
	"testing"
	"time"
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
	Smtp: &SmtpHandler{
		Level:                zapcore.DebugLevel,
		LevelPriority:        zapcore.DebugLevel,
		DelayMinutes:         int((time.Hour * 24).Minutes()), // 1 Day
		DelayPriorityMinutes: int((time.Minute * 5).Minutes()),
		Connector: utils.Smtp{
			Server:   "mail.domain.com",
			Port:     42,
			Username: "",
			Password: "",
			Subject:  "Default Log",
			Sender:   mail.Address{Name: "User", Address: "user@domain.tld"},

			Recipients: []mail.Address{
				{Name: "User1", Address: "user2@domain.tld"},
				{Name: "User1", Address: "user2@domain.tld"},
			},
			OpensslPath:         opensslPath,
			SignatureCertPath:   "",
			SignatureKeyPath:    "",
			EncryptionCertPaths: []string{},
			TempDir:             ""},
	},
}

// TestNewLogger checks the basic functionality of the logger.
func TestNewLogger(t *testing.T) {

	// Retrieve test settings
	testSettings, errSettings := _test.GetSettings()
	if errSettings != nil {
		t.Errorf("Invalid test settings: %s", errSettings)
		return
	} else if settings.Smtp == nil {
		t.Log("Smtp capabilities won't be tested")
	} else if testSettings.LogRecipient == "" || testSettings.LogRecipient == "user@domain.tld" {
		t.Error("Invalid smtp configuration")
		return
	}

	// Prepare cleanup
	defer func() { _ = os.Remove(settings.File.Path) }()

	// Get new independent (NOT THE GLOBAL) logger
	testLogger, err := InitGlobalLogger(settings)
	if err != nil {
		t.Errorf("unable to initialize global logger")
		return
	}

	defer func() {
		err := CloseGlobalLogger()
		if err != nil {
			t.Errorf("unable to close global logger")
		}
	}()

	// Prepare cleanup
	defer func() { _ = os.Remove(settings.File.Path) }()

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
