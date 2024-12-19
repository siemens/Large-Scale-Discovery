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
	"github.com/noneymous/PgProxy/pgproxy"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/log"
	"os"
	"path/filepath"
	"sync"
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

	// Parse Json
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
	logging.File.Path = "./logs/pgproxy.log"
	logging.Smtp.Connector.Subject = "PgProxy Error Log"

	// Prepare default config values
	managerAddress := "localhost:2222"

	// Prepare default agent config
	c := PgProxyConfig{
		ManagerAddress:    managerAddress,
		ManagerAddressSsl: true, // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
		Port:              5432,
		ForceSsl:          false,
		DefaultSni:        false,
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
	ManagerAddress    string        `json:"manager_address"`
	ManagerAddressSsl bool          `json:"manager_address_ssl"` // Encrypted endpoint be used, unless within a secure network or with a TLS load balancer is in front.
	Port              uint          `json:"port"`
	ForceSsl          bool          `json:"force_ssl"`
	DefaultSni        bool          `json:"default_sni"`
	Snis              []pgproxy.Sni `json:"snis"`
	Logging           log.Settings  `json:"logging"`
}

func (p *PgProxyConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux PgProxyConfig
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check values
	if raw.Port < 0 || raw.Port > 65535 {
		return fmt.Errorf("invalid port")
	}

	// Update struct with de-serialized values
	*p = PgProxyConfig(raw)

	// Return nil as everything went fine
	return nil
}
