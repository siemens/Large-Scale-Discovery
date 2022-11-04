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

	// Set the initial global broker config, in case all test cases get executed at once it will be already set.
	Set(&BrokerConfig{})

	// Prepare and run test cases
	tests := []struct {
		name string
		want BrokerConfig
	}{
		{"empty-brokerConfig", BrokerConfig{}},
		{"default-brokerConfig", defaultBrokerConfigFactory()}, // Default conf will be loaded after first unittest
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetConfig(); !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetConfig() = '%v', want = '%v'", got, &tt.want)
			}
		})

		conf := defaultBrokerConfigFactory()
		Set(&conf)
	}
}

func TestLoad(t *testing.T) {

	// Prepare cleanup
	defer func() { _ = os.Remove(testFile) }()

	// Prepare unit test data
	conf := defaultBrokerConfigFactory()
	err := Save(&conf, testFile)
	if err != nil {
		t.Errorf("Load() Could not prepare test case %v", err)
		return
	}

	// Prepare and run test cases
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantConf BrokerConfig
	}{
		{"", testFile, false, defaultBrokerConfigFactory()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Load(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = '%v', wantErr = '%v'", err, tt.wantErr)
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
		conf BrokerConfig
		want BrokerConfig
	}{
		{"sample", defaultBrokerConfigFactory(), defaultBrokerConfigFactory()},
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
		conf BrokerConfig
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"first-save", args{defaultBrokerConfigFactory(), testFile}, false},
		{"second-save", args{defaultBrokerConfigFactory(), testFile}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Save(&tt.args.conf, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
		})
	}
}
