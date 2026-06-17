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

func TestInit(t *testing.T) {

	// Retrieve test settings once to set working directory
	_ = _test.GetSettings()

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
			name:          "file-existing", // File will be created by the first test.
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

			// Cleanup if file will be created
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

func TestGetConfig(t *testing.T) {

	// Set the initial global agent config, in case all test cases get executed at once it will be already set.
	Set(&ManagerConfig{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want ManagerConfig
	}{
		{
			name: "empty-managerConfig",
			want: ManagerConfig{},
		},
		{
			name: "default-managerConfig", // Default conf will be loaded after first unittest.
			want: defaultManagerConfigFactory(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		conf := defaultManagerConfigFactory()
		Set(&conf)
	}
}

func TestLoad(t *testing.T) {

	// Retrieve test settings once to set working directory
	_ = _test.GetSettings()

	// Prepare temporary file name for this test
	testFile := fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare unit test data
	conf := defaultManagerConfigFactory()
	errSave := Save(&conf, testFile)
	if errSave != nil {
		t.Errorf("Load() Could not prepare test case: '%v'", errSave)
		return
	}

	// Marshal and unmarshal the expected config to populate computed fields (e.g. Duration) set only during UnmarshalJSON.
	wantConf := defaultManagerConfigFactory()
	wantBytes, errMarshal := json.Marshal(wantConf)
	if errMarshal != nil {
		t.Errorf("Load() Could not prepare test case: '%v'", errMarshal)
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
		wantConf ManagerConfig
	}{
		{
			name:     "",
			path:     testFile,
			wantErr:  false,
			wantConf: wantConf,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errLoad := Load(tt.path); (errLoad != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", errLoad, tt.wantErr)
			}
			if c := GetConfig(); !reflect.DeepEqual(c, &tt.wantConf) {
				t.Errorf("Load() = '%v', want = '%v'", c, &tt.wantConf)
			}
		})
	}
}

func TestSet(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name string
		conf ManagerConfig
		want ManagerConfig
	}{
		{
			name: "sample",
			conf: defaultManagerConfigFactory(),
			want: defaultManagerConfigFactory(),
		},
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

	// Retrieve test settings once to set working directory
	_ = _test.GetSettings()

	// Prepare temporary file name for this test
	testFile := fmt.Sprintf("test_%d.json", rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare and run test cases
	type args struct {
		conf ManagerConfig
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "first-save",
			args:    args{conf: defaultManagerConfigFactory(), path: testFile},
			wantErr: false,
		},
		{
			name:    "second-save",
			args:    args{conf: defaultManagerConfigFactory(), path: testFile},
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
