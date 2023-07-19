/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"github.com/davecgh/go-spew/spew"
	scanUtils "github.com/siemens/GoScans/utils"
	"os"
	"reflect"
	"testing"
)

const testFile = "test.json"

func TestInit(t *testing.T) {

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

	// Set the initial global web config, in case all test cases get executed at once it will be already set.
	Set(&WebConfig{})

	// Prepare and run test cases
	tests := []struct {
		name     string
		wantConf WebConfig
	}{
		{"empty-webConfig", WebConfig{}},
		{"default-webConfig", defaultWebConfigFactory()}, // Default conf will be loaded after first unittest
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfig()
			got.Jwt.Secret = ""         // Remove random data from struct
			tt.wantConf.Jwt.Secret = "" // Remove random data from struct
			if !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.wantConf)
			}
		})

		conf := defaultWebConfigFactory()
		Set(&conf)
	}
}

func TestLoad(t *testing.T) {

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare unit test data
	conf := defaultWebConfigFactory()
	errSave := Save(&conf, testFile)
	if errSave != nil {
		t.Errorf("Load() Could not prepare test case '%v'", errSave)
		return
	}

	// Prepare and run test cases
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantConf WebConfig
	}{
		{"", testFile, false, defaultWebConfigFactory()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errLoad := Load(tt.path); (errLoad != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", errLoad, tt.wantErr)
			}

			got := GetConfig()
			got.Jwt.Secret = ""         // Remove random data from struct
			tt.wantConf.Jwt.Secret = "" // Remove random data from struct

			spew.Dump(got)
			spew.Dump(tt.wantConf)

			if !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("Load() =\n '%v',\n want =\n '%v'", got, &tt.wantConf)
			}
		})
	}
}

func TestSet(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name     string
		conf     WebConfig
		wantConf WebConfig
	}{
		{"sample", defaultWebConfigFactory(), defaultWebConfigFactory()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(&tt.conf)

			got := GetConfig()
			got.Jwt.Secret = ""         // Remove random data from struct
			tt.wantConf.Jwt.Secret = "" // Remove random data from struct
			if !reflect.DeepEqual(got, &tt.wantConf) {
				t.Errorf("SetFile() = '%v', want = '%v'", got, &tt.wantConf)
			}
		})
	}
}

func TestSave(t *testing.T) {

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare and run test cases
	type args struct {
		conf WebConfig
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"first-save", args{defaultWebConfigFactory(), testFile}, false},
		{"second-save", args{defaultWebConfigFactory(), testFile}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Save(&tt.args.conf, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
		})
	}
}
