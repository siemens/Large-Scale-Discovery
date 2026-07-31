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

	"github.com/noneymous/PgProxy/pgproxy"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils/config"
)

// configPgProxy stores the loaded PostgreSQL proxy configuration
var configPgProxy = config.NewConfig[PgProxy]()

// Init initializes the configuration module and loads a JSON configuration. If the file does not exist, a default
// configuration is generated.
func Init(confPath string) error { return configPgProxy.Init(confPath, defaultPgProxyConfigFactory) }

// Load reads a configuration from a file, deserializes it, and sets it as the global configuration.
func Load(confPath string) error { return configPgProxy.Load(confPath) }

// Set sets a passed configuration as the global configuration.
func Set(conf *PgProxy) { configPgProxy.Set(conf) }

// Save serializes a given configuration and writes it to a file.
func Save(conf *PgProxy, path string) error { return configPgProxy.Save(conf, path) }

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *PgProxy { return configPgProxy.Get() }

func defaultPgProxyConfigFactory() PgProxy {

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

	// Prepare default PgProxy config
	conf := PgProxy{
		ManagerAddress: managerAddress,
		ManagerSsl:     true, // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
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
	return conf
}

// PgProxy is the root configuration object tying all PgProxy settings together.
type PgProxy struct {
	ManagerAddress string        `json:"manager_address"`
	ManagerSsl     bool          `json:"manager_ssl"`    // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
	ManagerSecret  string        `json:"manager_secret"` // Token to authorize RPC connections to invoke manager RPC methods.
	Port           uint          `json:"port"`
	ForceSsl       bool          `json:"force_ssl"`
	DefaultSni     bool          `json:"default_sni"`
	Snis           []pgproxy.Sni `json:"snis"`
	Logging        log.Settings  `json:"logging"`
}

// UnmarshalJSON reads JSON, validates values, and populates the configuration.
func (conf *PgProxy) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux PgProxy
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	errUnmarshal := json.Unmarshal(b, &raw)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Check values
	if len(raw.ManagerSecret) == 0 {
		return fmt.Errorf("manager secret required")
	}
	if raw.Port > 65535 { // uint cannot be < 0 anyway
		return fmt.Errorf("invalid port")
	}

	// Update struct with de-serialized values
	*conf = PgProxy(raw)

	// Return nil as everything went fine
	return nil
}
