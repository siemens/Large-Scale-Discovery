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

// configBroker stores the loaded broker configuration
var configBroker = config.NewConfig[Broker]()

// Init initializes the configuration module and loads a JSON configuration. If the file does not exist, a default
// configuration is generated.
func Init(confPath string) error { return configBroker.Init(confPath, defaultBrokerConfigFactory) }

// Load reads a configuration from a file, deserializes it, and sets it as the global configuration.
func Load(confPath string) error { return configBroker.Load(confPath) }

// Set sets a passed configuration as the global configuration.
func Set(conf *Broker) { configBroker.Set(conf) }

// Save serializes a given configuration and writes it to a file.
func Save(conf *Broker, path string) error { return configBroker.Save(conf, path) }

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *Broker { return configBroker.Get() }

func defaultBrokerConfigFactory() Broker {

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
	conf := Broker{
		ListenAddress:          "localhost:3333",
		ListenSsl:              true, // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
		ManagerAddress:         "localhost:2222",
		ManagerSsl:             true, // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
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

// Broker is the root configuration object tying all broker settings together.
type Broker struct {
	ListenAddress          string       `json:"listen_address"`
	ListenSsl              bool         `json:"listen_ssl"` // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
	ManagerAddress         string       `json:"manager_address"`
	ManagerSsl             bool         `json:"manager_ssl"`              // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
	ManagerSecret          string       `json:"manager_secret"`           // Token to authorize RPC connections to invoke manager RPC methods.
	ManagerSecretPrivilege string       `json:"manager_secret_privilege"` // Tokens to authorize sensitive RPC requests, to query full scope details, including scope secret and the scope's database credentials.
	DbConnections          int          `json:"db_connections"`
	Logging                log.Settings `json:"logging"`
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (conf *Broker) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Broker
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
	if len(raw.ManagerSecretPrivilege) == 0 {
		return fmt.Errorf("manager privilege secret required")
	}

	// Copy loaded JSON values to actual config
	*conf = Broker(raw)

	// Return nil as everything is valid
	return nil
}
