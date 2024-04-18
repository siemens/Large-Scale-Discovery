/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"encoding/json"
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
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

	// Parse Json
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

	// Serialize to Json
	file, errMarshal := json.MarshalIndent(conf, "", "    ")
	if errMarshal != nil {
		return errMarshal
	}

	// Write Json to file
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
	logging.File.Path = "./logs/importer.log"
	logging.Smtp.Connector.Subject = "Importer Error Log"

	// Prepare default settings for development
	if _build.DevMode {
		logging.Console.Level = zapcore.DebugLevel
	}

	// Generate importer config with default values
	conf := ImporterConfig{
		ManagerAddress: "localhost:2222",
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
	Logging        log.Settings           `json:"logging"`
	Importer       map[string]interface{} `json:"importer"` // Arbitrary arguments passed to importers. Flexible for own importer integrations.
}
