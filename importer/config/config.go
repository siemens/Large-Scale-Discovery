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
	"path/filepath"

	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils/config"
	"go.uber.org/zap/zapcore"
)

// configImporter stores the loaded importer configuration
var configImporter = config.NewConfig[Importer]()

// Init initializes the configuration module and loads a JSON configuration. If the file does not exist, a default
// configuration is generated.
func Init(confPath string) error { return configImporter.Init(confPath, defaultImporterConfigFactory) }

// Load reads a configuration from a file, deserializes it, and sets it as the global configuration.
func Load(confPath string) error { return configImporter.Load(confPath) }

// Set sets a passed configuration as the global configuration.
func Set(conf *Importer) { configImporter.Set(conf) }

// Save serializes a given configuration and writes it to a file.
func Save(conf *Importer, path string) error { return configImporter.Save(conf, path) }

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *Importer { return configImporter.Get() }

func defaultImporterConfigFactory() Importer {

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
	conf := Importer{
		ManagerAddress: "localhost:2222",
		ManagerSsl:     true, // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
		ManagerSecret:  managerSecret,
		Logging:        logging,
		Inventory:      map[string]interface{}{ // Flexible map of arguments as needed by integrated inventory sources
		},
	}

	// Return generated config
	return conf
}

//
// JSON structure of configuration
//

// Importer is the root configuration object tying all importer settings together.
type Importer struct {
	ManagerAddress string                 `json:"manager_address"`
	ManagerSsl     bool                   `json:"manager_ssl"`    // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
	ManagerSecret  string                 `json:"manager_secret"` // Token to authorize RPC connections to invoke manager RPC methods.
	Logging        log.Settings           `json:"logging"`
	Inventory      map[string]interface{} `json:"inventory"` // Arbitrary arguments passed to importers. Flexible for own inventory integrations.
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (conf *Importer) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Importer
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	errUnmarshal := json.Unmarshal(b, &raw)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Do input validation
	if len(raw.ManagerSecret) == 0 {
		return fmt.Errorf("manager secret required")
	}

	// Copy loaded JSON values to actual config
	*conf = Importer(raw)

	// Return nil as everything is valid
	return nil
}
