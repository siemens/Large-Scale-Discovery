/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_test"
)

// TestMain ensures the working directory is set to _test/ before any tests run.
func TestMain(m *testing.M) {

	// GetSettings sets cwd to _test/ so all test-created files are isolated there
	_ = _test.GetSettings()
	os.Exit(m.Run())
}

// TestInit verifies that Init creates a default config file when absent and loads it when present.
func TestInit(t *testing.T) {

	// Skip when required executables are absent. The config UnmarshalJSON validates their paths.
	if _, err := exec.LookPath("nmap"); err != nil {
		t.Skip("Integration test skipped: nmap not available")
		return
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Integration test skipped: python3 not available")
		return
	}

	// Prepare temporary file name for this test
	testFile := fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

	// Init test config
	_ = Init(testFile)

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
		{
			name:          "file-not-existing",
			path:          fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63()),
			wantValidFile: true,
			wantErr:       true,
			wantErrStr:    "no configuration, created default in",
		},
		{
			name:          "file-existing", // File will be created by the first test
			path:          testFile,
			wantValidFile: true,
			wantErr:       false,
			wantErrStr:    "",
		},
		{
			name:          "path-not-existing",
			path:          "nonExistingPath/conf.config",
			wantValidFile: false,
			wantErr:       true,
			wantErrStr:    "could not initialize configuration in 'nonExistingPath/conf.config': open nonExistingPath/conf.config: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Prepare cleanup
			if tt.path != testFile {
				defer func() { _ = os.Remove(tt.path) }()
			}

			// Execute test
			errInit := Init(tt.path)
			if (errInit != nil) != tt.wantErr || (errInit != nil && !strings.HasPrefix(errInit.Error(), tt.wantErrStr)) {
				t.Errorf("Init() error = '%v', wantErr = '%v'", errInit, tt.wantErrStr)
			}
			if (scanUtils.IsValidFile(tt.path) == nil) != tt.wantValidFile {
				t.Errorf("Init() isValidFile = '%v', wantValidFile = '%v'", scanUtils.IsValidFile(tt.path), tt.wantValidFile)
			}
		})
	}
}

// TestGetConfig verifies that GetConfig returns the currently stored global agent configuration.
func TestGetConfig(t *testing.T) {

	// Set the initial global agent config, in case all test cases get executed at once it will be already set
	Set(&AgentConfig{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want AgentConfig
	}{
		{
			name: "empty-agent-config",
			want: AgentConfig{},
		},
		{
			name: "default-agent-config", // Default conf will be loaded after first unittest
			want: defaultAgentConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		// Advance the global config state for the next test case
		conf := defaultAgentConfigFactory()
		Set(&conf)
	}
}

// TestLoad verifies that Load reads a config file and populates the global agent configuration correctly.
func TestLoad(t *testing.T) {

	// Skip when required executables are absent. The config UnmarshalJSON validates their paths.
	if _, err := exec.LookPath("nmap"); err != nil {
		t.Skip("Integration test skipped: nmap not available")
		return
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Integration test skipped: python3 not available")
		return
	}

	// Prepare temporary file name for this test
	testFile := fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

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
	// operating systems, the appropriate default configuration should be marshaled and unmarshaled.
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
		{
			name:     "valid-config-file",
			path:     testFile,
			wantErr:  false,
			wantConf: defaultConf,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Load(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("Load() =\n '%v',\n want =\n '%v'", got, &tt.wantConf)
			}
		})
	}
}

// TestSet verifies that Set stores the provided configuration as the global agent config.
func TestSet(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name string
		conf AgentConfig
		want AgentConfig
	}{
		{
			name: "sample",
			conf: defaultAgentConfigFactory(),
			want: defaultAgentConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(&tt.conf)
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("Set() = '%v', want = '%v'", got, &tt.want)
			}
		})
	}
}

// TestSave verifies that Save marshals the agent configuration to a JSON file without error.
func TestSave(t *testing.T) {

	// Prepare temporary file name for this test
	testFile := fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

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
		{
			name:    "first-save",
			args:    args{conf: defaultAgentConfigFactory(), path: testFile},
			wantErr: false,
		},
		{
			name:    "second-save",
			args:    args{conf: defaultAgentConfigFactory(), path: testFile},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Save(&tt.args.conf, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
		})
	}
}
