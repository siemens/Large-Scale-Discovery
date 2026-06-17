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

	"github.com/noneymous/PgProxy/pgproxy"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
)

var pgProxyConf = &PgProxyConfig{}
var pgProxyConfLock sync.Mutex // Lock required to avoid simultaneous requesting/updating of config

// Init initializes the configuration module and loads a JSON configuration. If JSON is not existing, a default
// configuration will be generated.
func Init(configFile string) error {
	if errFile := scanUtils.IsValidFile(configFile); errFile != nil {
		defaultConf := defaultPgProxyConfigFactory()
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
	pgProxyConfLock.Lock()
	defer pgProxyConfLock.Unlock()

	// Prepare new config, don't work on the global values
	newConfig := &PgProxyConfig{}

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
	pgProxyConf = newConfig

	// Return nil to indicate successful config update
	return nil
}

// Save serializes a given configuration and writes it to a file
func Save(conf *PgProxyConfig, path string) error {

	// Lock global config, because the given config pointer might be the global config
	pgProxyConfLock.Lock()
	defer pgProxyConfLock.Unlock()

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
func GetConfig() *PgProxyConfig {

	// The global configuration might get updated regularly to allow user updating settings without aborting scans
	pgProxyConfLock.Lock()
	defer pgProxyConfLock.Unlock()

	// Return current global configuration
	return pgProxyConf
}

func defaultPgProxyConfigFactory() PgProxyConfig {

	// Prepare default logging settings and adapt for agent
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = filepath.Join("logs", "pgproxy.log")
	logging.Smtp.Connector.Subject = "PgProxy Error Log"

	// Prepare default config values
	managerAddress := "localhost:2222"
	managerSecret := ""
	if _build.DevMode {
		managerSecret = "dev_secret"
	}

	// Prepare default agent config
	c := PgProxyConfig{
		ManagerAddress: managerAddress,
		ManagerSsl:     true, // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
		ManagerSecret:  managerSecret,
		Port:           5432,
		ForceSsl:       false,
		DefaultSni:     false,
		Snis: []pgproxy.Sni{
			{
				CertPath: filepath.Join("keys", "pgproxy_dev.crt"),
				KeyPath:  filepath.Join("keys", "pgproxy_dev.key"),
				Database: pgproxy.Database{
					Host:    "localhost",
					Port:    5432,
					SslMode: "prefer",
				},
			},
		},
		Logging: logging,
	}

	// Return generated config
	return c
}

type PgProxyConfig struct {
	ManagerAddress string        `json:"manager_address"`
	ManagerSsl     bool          `json:"manager_ssl"`    // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
	ManagerSecret  string        `json:"manager_secret"` // Token to authorize RPC connections to invoke manager RPC methods.
	Port           uint          `json:"port"`
	ForceSsl       bool          `json:"force_ssl"`
	DefaultSni     bool          `json:"default_sni"`
	Snis           []pgproxy.Sni `json:"snis"`
	Logging        log.Settings  `json:"logging"`
}

func (p *PgProxyConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux PgProxyConfig
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check values
	if len(raw.ManagerSecret) == 0 {
		return fmt.Errorf("manager secret required")
	}
	if raw.Port > 65535 { // uint cannot be < 0 anyway
		return fmt.Errorf("invalid port")
	}

	// Update struct with de-serialized values
	*p = PgProxyConfig(raw)

	// Return nil as everything went fine
	return nil
}
