/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"encoding/json"
	scanUtils "github.com/siemens/GoScans/utils"
	"os"
	"reflect"
	"strings"
	"testing"
)

const testFile = "test.json"

// checkAndChangeWd checks whether the current working directory is the requires ".../_bin" path. This is need in order
// for the checks during the unmarshal process to succeed, because unmarshalling does test whether the paths configured
// in the config file do exist.
func checkAndChangeWd(t *testing.T) {

	// Get the working directory.
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get the current working directory: '%v'", err)
	}

	// Check if the working directory has the correct suffix. This is not an ideal check but still better than nothing.
	if !strings.HasSuffix(dir, "_bin") {

		// Change to the correct working directory - current location should be ".../agent/config" directory.
		errCh := os.Chdir("../../_bin")
		if errCh != nil {
			t.Errorf("Could not change to the correct working directory: '%v'", errCh)
			return
		}
	}
}

func TestInit(t *testing.T) {

	// Set the correct working directory if needed.
	checkAndChangeWd(t)

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare and run test cases
	tests := []struct {
		name          string
		path          string
		wantValidFile bool
		wantErr       bool
		wantErrStr    string
	}{
		{"file-not-existing", testFile, true, true, "no configuration, created default in 'test.json'"},
		{"file-existing", testFile, true, false, ""}, // File will be created by the first test.
		{"path-not-existing", "nonExistingPath/conf.config", false, true, "could not initialize configuration in 'nonExistingPath/conf.config': open nonExistingPath/conf.config: The system cannot find the path specified."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errInit := Init(tt.path)
			if (errInit != nil) != tt.wantErr || (errInit != nil && errInit.Error() != tt.wantErrStr) {
				t.Errorf("Init() error = '%v', wantErr = '%v'", errInit, tt.wantErrStr)
			}

			if (scanUtils.IsValidFile(tt.path) == nil) != tt.wantValidFile {
				t.Errorf("Init() isValidFile = '%v', wantValidFile = '%v'", scanUtils.IsValidFile(tt.path), tt.wantValidFile)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {

	// Set the initial global agent config, in case all test cases get executed at once it will be already set.
	Set(&AgentConfig{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want AgentConfig
	}{
		{"empty-agentConfig", AgentConfig{}},
		{"default-agentConfig", defaultAgentConfigFactory()}, // Default conf will be loaded after first unittest
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		conf := defaultAgentConfigFactory()
		Set(&conf)
	}
}

func TestLoad(t *testing.T) {

	// Set the correct working directory if needed.
	checkAndChangeWd(t)

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare unit test data
	conf := defaultAgentConfigFactory()
	errSave := Save(&conf, testFile)
	if errSave != nil {
		t.Errorf("Load() Could not prepare test case '%v'", errSave)
		return
	}

	// There are some fields that only get set during the unmarshal process. As these might vary between the different
	// operating systems, the appropriate default configuration should be marshalled und unmarshalled.
	defaultConf := defaultAgentConfigFactory()
	confBytes, errMarshal := json.Marshal(defaultConf)
	if errMarshal != nil {
		t.Errorf("Load() Could not prepare test case '%v'", errMarshal)
		return
	}
	if errUnmarshal := json.Unmarshal(confBytes, &defaultConf); errUnmarshal != nil {
		t.Errorf("Load() Could not prepare test case '%v'", errUnmarshal)
		return
	}

	// Prepare and run test cases
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantConf AgentConfig
	}{
		{"", testFile, false, defaultConf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Load(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
			if c := GetConfig(); !reflect.DeepEqual(c, &tt.wantConf) {
				t.Errorf("Load() =\n '%v',\n want =\n '%v'", c, &tt.wantConf)
			}
		})
	}
}

func TestSet(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name string
		conf AgentConfig
		want AgentConfig
	}{
		{"sample", defaultAgentConfigFactory(), defaultAgentConfigFactory()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(&tt.conf)
			if c := GetConfig(); !reflect.DeepEqual(c, &tt.want) {
				t.Errorf("SetFile() = '%v', want = '%v'", c, &tt.want)
			}
		})
	}
}

func TestSave(t *testing.T) {

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare and run test cases
	type args struct {
		conf AgentConfig
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"first-save", args{defaultAgentConfigFactory(), testFile}, false},
		{"second-save", args{defaultAgentConfigFactory(), testFile}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Save(&tt.args.conf, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
		})
	}
}
