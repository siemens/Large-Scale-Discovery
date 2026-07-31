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

// testConf is a minimal configuration struct used to exercise the generic Config[T] methods
// without pulling in component-specific dependencies or UnmarshalJSON side effects.
type testConf struct {
	Name     string        `json:"name"`
	Count    int           `json:"count"`
	Duration time.Duration `json:"duration"`
}

// defaultTestConfFactory returns a testConf with known default values for assertion.
func defaultTestConfFactory() testConf {
	return testConf{
		Name:     "default",
		Count:    42,
		Duration: 5 * time.Second,
	}
}

// randomTestFile returns a temporary file name with a random suffix to avoid collisions between tests.
func randomTestFile() string {
	return fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())
}

// TestInit verifies that Init creates a default config file when absent and loads it when present.
func TestInit(t *testing.T) {

	// Prepare a config instance for testing
	cfg := NewConfig[testConf]()

	// Prepare temporary file name for this test
	testFile := randomTestFile()

	// Seed the file so the "file-existing" case has something to load
	_ = cfg.Init(testFile, defaultTestConfFactory)

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
			path:          randomTestFile(),
			wantValidFile: true,
			wantErr:       true,
			wantErrStr:    "no configuration, created default in",
		},
		{
			name:          "file-existing",
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

			// Prepare cleanup for files created during this subtest
			if tt.path != testFile {
				defer func() { _ = os.Remove(tt.path) }()
			}

			// Execute test
			errInit := cfg.Init(tt.path, defaultTestConfFactory)
			if (errInit != nil) != tt.wantErr || (errInit != nil && !strings.HasPrefix(errInit.Error(), tt.wantErrStr)) {
				t.Errorf("Init() error = '%v', wantErr = '%v'", errInit, tt.wantErrStr)
			}
			if (scanUtils.IsValidFile(tt.path) == nil) != tt.wantValidFile {
				t.Errorf("Init() isValidFile = '%v', wantValidFile = '%v'", scanUtils.IsValidFile(tt.path), tt.wantValidFile)
			}
		})
	}
}

// TestGet verifies that Get returns the currently stored configuration.
func TestGet(t *testing.T) {

	// Prepare a config instance for testing
	cfg := NewConfig[testConf]()

	// Prepare and run test cases
	tests := []struct {
		name string
		want testConf
	}{
		{
			name: "empty-config",
			want: testConf{},
		},
		{
			name: "default-config",
			want: defaultTestConfFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Verify current config matches expected value
			if got := cfg.Get(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("Get() = '%v', want = '%v'", got, &tt.want)
			}
		})

		// Advance config state for the next test case
		conf := defaultTestConfFactory()
		cfg.Set(&conf)
	}
}

// TestLoad verifies that Load reads a config file and populates the configuration correctly.
func TestLoad(t *testing.T) {

	// Prepare a config instance for testing
	cfg := NewConfig[testConf]()

	// Prepare temporary file name for this test
	testFile := randomTestFile()

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare unit test data by saving a default config
	conf := defaultTestConfFactory()
	errSave := cfg.Save(&conf, testFile)
	if errSave != nil {
		t.Errorf("Load() Could not prepare test case: '%v'", errSave)
		return
	}

	// Marshal and unmarshal the expected config to populate computed fields set only during UnmarshalJSON
	wantConf := defaultTestConfFactory()
	wantBytes, errWantBytes := json.Marshal(wantConf)
	if errWantBytes != nil {
		t.Errorf("Load() Could not prepare test case: '%v'", errWantBytes)
		return
	}
	if errUnmarshal := json.Unmarshal(wantBytes, &wantConf); errUnmarshal != nil {
		t.Errorf("Load() Could not prepare test case: '%v'", errUnmarshal)
		return
	}

	// Prepare and run test cases
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantConf testConf
	}{
		{
			name:     "valid-config-file",
			path:     testFile,
			wantErr:  false,
			wantConf: wantConf,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Execute load
			if errLoad := cfg.Load(tt.path); (errLoad != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", errLoad, tt.wantErr)
			}

			// Verify loaded config matches expected value
			if got := cfg.Get(); !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("Load() =\n '%v',\n want =\n '%v'", got, &tt.wantConf)
			}
		})
	}
}

// TestSet verifies that Set stores the provided configuration.
func TestSet(t *testing.T) {

	// Prepare a config instance for testing
	cfg := NewConfig[testConf]()

	// Prepare and run test cases
	tests := []struct {
		name     string
		conf     testConf
		wantConf testConf
	}{
		{
			name:     "sample",
			conf:     defaultTestConfFactory(),
			wantConf: defaultTestConfFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Execute set
			cfg.Set(&tt.conf)

			// Verify stored config matches expected value
			if got := cfg.Get(); !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("Set() = '%v', want = '%v'", got, &tt.wantConf)
			}
		})
	}
}

// TestSave verifies that Save marshals the configuration to a JSON file without error.
func TestSave(t *testing.T) {

	// Prepare a config instance for testing
	cfg := NewConfig[testConf]()

	// Prepare temporary file name for this test
	testFile := randomTestFile()

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare and run test cases
	type args struct {
		conf testConf
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "first-save",
			args:    args{conf: defaultTestConfFactory(), path: testFile},
			wantErr: false,
		},
		{
			name:    "second-save",
			args:    args{conf: defaultTestConfFactory(), path: testFile},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Execute save
			if errSave := cfg.Save(&tt.args.conf, tt.args.path); (errSave != nil) != tt.wantErr {
				t.Errorf("Save() error = '%v', wantErr = '%v'", errSave, tt.wantErr)
			}
		})
	}
}
