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
	"os"
	"path/filepath"
	"sync"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"go.uber.org/zap/zapcore"
)

var importerConfig = &ImporterConfig{} // Global configuration
var importerConfigLock sync.Mutex      // Lock required to avoid simultaneous requesting/updating of config

// Init initializes the configuration module and loads a JSON configuration. If JSON is not existing, a default
// configuration will be generated.
func Init(configFile string) error {
	if errFile := scanUtils.IsValidFile(configFile); errFile != nil {
		defaultConf := defaultImporterConfigFactory()
		errSave := Save(&defaultConf, configFile)
		if errSave != nil {
			return fmt.Errorf("could not initialize configuration in '%s': %s", configFile, errSave)
		} else {
			return fmt.Errorf("no configuration, created default in '%s'", configFile)
		}
	} else {
		errLoad := Load(configFile)
		if errLoad != nil {
			return errLoad
		} else {
			return nil
		}
	}
}

// Load reads a configuration from a file, de-serializes it and sets it as the global configuration
func Load(path string) error {

	// Lock global config before initializing an update
	importerConfigLock.Lock()
	defer importerConfigLock.Unlock()

	// Prepare new config, don't work on the global values
	newConfig := &ImporterConfig{}

	// Read file content
	rawJson, errLoad := os.ReadFile(path)
	if errLoad != nil {
		return errLoad
	}

	// Parse JSON
	errUnmarshal := json.Unmarshal(rawJson, newConfig)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Replace global configuration with new one
	importerConfig = newConfig

	// Return nil to indicate successful config update
	return nil
}

// Set sets a passed configuration as the global configuration
func Set(conf *ImporterConfig) {

	// Lock global config before initializing an update
	importerConfigLock.Lock()
	defer importerConfigLock.Unlock()

	// Replace global configuration with new one
	importerConfig = conf
}

// Save serializes a given configuration and writes it to a file
func Save(conf *ImporterConfig, path string) error {

	// Lock global config, because the given config pointer might be the global config
	importerConfigLock.Lock()
	defer importerConfigLock.Unlock()

	// Serialize to JSON
	file, errMarshal := json.MarshalIndent(conf, "", "    ")
	if errMarshal != nil {
		return errMarshal
	}

	// Write JSON to file
	errWrite := os.WriteFile(path, file, 0644)
	if errWrite != nil {
		return errWrite
	}

	// Return nil to indicate successful storage
	return nil
}

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *ImporterConfig {

	// The global configuration might get updated regularly to allow user updating settings without aborting scans
	importerConfigLock.Lock()
	defer importerConfigLock.Unlock()

	// Return current global configuration
	return importerConfig
}

func defaultImporterConfigFactory() ImporterConfig {

	// Prepare default logging settings and adapt for importer
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = filepath.Join("logs", "importer.log")
	logging.Smtp.Connector.Subject = "Importer Error Log"

	// Prepare default settings for development
	managerSecret := ""
	if _build.DevMode {
		managerSecret = "dev_secret"
		logging.Console.Level = zapcore.DebugLevel
	}

	// Generate importer config with default values
	conf := ImporterConfig{
		ManagerAddress: "localhost:2222",
		ManagerSsl:     true, // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
		ManagerSecret:  managerSecret,
		Logging:        logging,
		Importer:       map[string]interface{}{ // Flexible map of arguments as needed by integrated importers
		},
	}

	// Return generated config
	return conf
}

//
// JSON structure of configuration
//

type ImporterConfig struct {
	// The root configuration object tying all configuration segments together.
	ManagerAddress string                 `json:"manager_address"`
	ManagerSsl     bool                   `json:"manager_ssl"`    // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
	ManagerSecret  string                 `json:"manager_secret"` // Token to authorize RPC connections to invoke manager RPC methods.
	Logging        log.Settings           `json:"logging"`
	Importer       map[string]interface{} `json:"importer"` // Arbitrary arguments passed to importers. Flexible for own importer integrations.
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (c *ImporterConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ImporterConfig
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Do input validation
	if len(raw.ManagerSecret) == 0 {
		return fmt.Errorf("manager secret required")
	}

	// Copy loaded JSON values to actual config
	*c = ImporterConfig(raw)

	// Return nil as everything is valid
	return nil
}
