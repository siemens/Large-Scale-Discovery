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

var brokerConfig = &BrokerConfig{} // Global configuration
var brokerConfigLock sync.Mutex    // Lock required to avoid simultaneous requesting/updating of config

// Init initializes the configuration module and loads a JSON configuration. If JSON is not existing, a default
// configuration will be generated.
func Init(configFile string) error {
	if errFile := scanUtils.IsValidFile(configFile); errFile != nil {
		defaultConf := defaultBrokerConfigFactory()
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
	brokerConfigLock.Lock()
	defer brokerConfigLock.Unlock()

	// Prepare new config, don't work on the global values
	newConfig := &BrokerConfig{}

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
	brokerConfig = newConfig

	// Return nil to indicate successful config update
	return nil
}

// Set sets a passed configuration as the global configuration
func Set(conf *BrokerConfig) {

	// Lock global config before initializing an update
	brokerConfigLock.Lock()
	defer brokerConfigLock.Unlock()

	// Replace global configuration with new one
	brokerConfig = conf
}

// Save serializes a given configuration and writes it to a file
func Save(conf *BrokerConfig, path string) error {

	// Lock global config, because the given config pointer might be the global config
	brokerConfigLock.Lock()
	defer brokerConfigLock.Unlock()

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
func GetConfig() *BrokerConfig {

	// The global configuration might get updated regularly to allow user updating settings without aborting scans
	brokerConfigLock.Lock()
	defer brokerConfigLock.Unlock()

	// Return current global configuration
	return brokerConfig
}

func defaultBrokerConfigFactory() BrokerConfig {

	// Prepare default logging settings and adapt for broker
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = filepath.Join("logs", "broker.log")
	logging.Smtp.Connector.Subject = "Broker Error Log"

	// Prepare default settings for development
	secret := ""
	secretPrivilege := ""
	if _build.DevMode {
		secret = "dev_secret"
		secretPrivilege = "dev_secret"
		logging.Console.Level = zapcore.DebugLevel
	}

	// Prepare default broker config
	conf := BrokerConfig{
		ListenAddress:          "localhost:3333",
		ListenSsl:              true, // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
		ManagerAddress:         "localhost:2222",
		ManagerSsl:             true, // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
		ManagerSecret:          secret,
		ManagerSecretPrivilege: secretPrivilege,
		DbConnections:          30,
		Logging:                logging,
	}

	// Return generated config
	return conf
}

//
// JSON structure of configuration
//

type BrokerConfig struct {
	// The root configuration object tying all configuration segments together.
	ListenAddress          string       `json:"listen_address"`
	ListenSsl              bool         `json:"listen_ssl"` // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
	ManagerAddress         string       `json:"manager_address"`
	ManagerSsl             bool         `json:"manager_ssl"`              // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
	ManagerSecret          string       `json:"manager_secret"`           // Token to authorize RPC connections to invoke manager RPC methods.
	ManagerSecretPrivilege string       `json:"manager_secret_privilege"` // Tokens to authorize sensitive RPC requests, to query full scope details, including scope secret and the scope's database credentials.
	DbConnections          int          `json:"db_connections"`
	Logging                log.Settings `json:"logging"`
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (c *BrokerConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux BrokerConfig
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
	if len(raw.ManagerSecretPrivilege) == 0 {
		return fmt.Errorf("manager privilege secret required")
	}

	// Copy loaded JSON values to actual config
	*c = BrokerConfig(raw)

	// Return nil as everything is valid
	return nil
}
