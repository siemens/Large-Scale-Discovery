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
	"encoding/json"
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"large-scale-discovery/_build"
	"large-scale-discovery/log"
	"large-scale-discovery/utils"
	"path/filepath"
	"sync"
	"time"
)

var webConfig = &WebConfig{} // Global configuration
var webConfigLock sync.Mutex // Lock required to avoid simultaneous requesting/updating of config

// Init initializes the configuration module and loads a JSON configuration. If JSON is not existing, a default
// configuration will be generated.
func Init(configFile string) error {
	if errFile := scanUtils.IsValidFile(configFile); errFile != nil {
		defaultConf := defaultWebConfigFactory()
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
	webConfigLock.Lock()
	defer webConfigLock.Unlock()

	// Prepare new config, don't work on the global values
	newConfig := &WebConfig{}

	// Read file content
	rawJson, errLoad := ioutil.ReadFile(path)
	if errLoad != nil {
		return errLoad
	}

	// Parse Json
	errUnmarshal := json.Unmarshal(rawJson, newConfig)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Replace global configuration with new one
	webConfig = newConfig

	// Return nil to indicate successful config update
	return nil
}

// Set sets a passed configuration as the global configuration
func Set(conf *WebConfig) {

	// Lock global config before initializing an update
	webConfigLock.Lock()
	defer webConfigLock.Unlock()

	// Replace global configuration with new one
	webConfig = conf
}

// Save serializes a given configuration and writes it to a file
func Save(conf *WebConfig, path string) error {

	// Lock global config, because the given config pointer might be the global config
	webConfigLock.Lock()
	defer webConfigLock.Unlock()

	// Serialize to Json
	file, errMarshal := json.MarshalIndent(conf, "", "    ")
	if errMarshal != nil {
		return errMarshal
	}

	// Write Json to file
	errWrite := ioutil.WriteFile(path, file, 0644)
	if errWrite != nil {
		return errWrite
	}

	// Return nil to indicate successful storage
	return nil
}

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *WebConfig {

	// The global configuration might get updated regularly to allow user updating settings without aborting scans
	webConfigLock.Lock()
	defer webConfigLock.Unlock()

	// Return current global configuration
	return webConfig
}

func defaultWebConfigFactory() WebConfig {

	// Prepare default logging settings and adapt web
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = "./logs/backend.log"
	logging.Smtp.Connector.Subject = "Backend Error Log"

	// Define default values
	frontendUrl := "https://domain.tld"
	jwtLifetime := time.Minute * 20  // 20 minutes
	jwtJifetimeMax := time.Hour * 12 // 12 hours until fresh authentication is required
	jwtSecret, _ := utils.GenerateToken(utils.AlphaNumCaseSymbol, 64)

	// Prepare default settings for development
	if _build.DevMode {
		frontendUrl = "https://localhost"
		logging.Console.Level = zapcore.DebugLevel
	}

	// Prepare default web config
	conf := WebConfig{
		ManagerAddress: "localhost:2222",
		ListenAddress:  "localhost:443",
		FrontendUrl:    frontendUrl,
		Jwt: Jwt{
			Secret:         jwtSecret,
			Algorithm:      "HS256",
			ExpiryMinutes:  int(jwtLifetime.Minutes()),
			Expiry:         jwtLifetime,
			RefreshMinutes: int(jwtJifetimeMax.Minutes()),
			Refresh:        jwtJifetimeMax,
		},
		Authenticator: map[string]interface{}{ // Flexible map of arguments as needed by integrated authenticators
			"credentials_registration": false,
			"oauth_public_key_url":     "https://sso.domain.tld/ext/oauth/jwks",
			"oauth_config_url":         "https://sso.domain.tld/.well-known/openid-configuration",
			"oauth_redirect_url":       "https://application.domain.tld/api/v1/auth/oauth/callback",
			"oauth_client_id":          "",
			"oauth_client_secret":      "",
		},
		Loader: map[string]interface{}{ // Flexible map of arguments as needed by integrated connectors
			"ldap_certificate_path": filepath.Join("keys", "scdasldap.siemens.net.pem"),
			"ldap_host":             "scdasldap.siemens.net",
			"ldap_port":             float64(636), // will be float after loading from JSON. Must be float for unit test to succeed.
			"ldap_user":             "",
			"ldap_password":         "",
			"ldap_timeout_seconds":  float64(5), // will be float after loading from JSON. Must be float for unit test to succeed.
		},
		Logging: logging,
	}

	// Return generated config
	return conf
}

//
// JSON structure of configuration
//

type WebConfig struct {
	// The root configuration object tying all configuration segments together.
	ManagerAddress string                 `json:"manager_address"`
	ListenAddress  string                 `json:"listen_address"`
	FrontendUrl    string                 `json:"frontend_url"` // Valid URL that can be used by users to access the web interface. Will also be included e.g. in e-mails sent by the backend.
	Jwt            Jwt                    `json:"jwt"`
	Logging        log.Settings           `json:"logging"`
	Authenticator  map[string]interface{} `json:"authenticator"` // Arbitrary arguments passed to authenticators. Flexible for own authenticator integrations.
	Loader         map[string]interface{} `json:"loader"`        // Arbitrary arguments passed to connectors. Flexible for own connector integrations.
}

func (c *WebConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux WebConfig
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Copy loaded Json values to actual config
	*c = WebConfig(raw)

	// Return nil as everything is valid
	return nil
}

type Jwt struct {
	// LDAP connection information
	Secret         string        `json:"secret"`          // Secret to encrypt JWT token with
	Algorithm      string        `json:"algorithm"`       // Algorithm to encrypt JWT token with
	ExpiryMinutes  int           `json:"expiry_minutes"`  // Max time a JWT token is valid
	Expiry         time.Duration `json:"-"`               //
	RefreshMinutes int           `json:"refresh_minutes"` // Max time a JWT token can be refreshed
	Refresh        time.Duration `json:"-"`               //
}

func (j *Jwt) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux Jwt
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Copy loaded Json values to actual config
	*j = Jwt(raw)

	// Set unserializable values
	j.Expiry = time.Duration(raw.ExpiryMinutes) * time.Minute
	j.Refresh = time.Duration(raw.RefreshMinutes) * time.Minute

	// Return nil as everything is valid
	return nil
}
