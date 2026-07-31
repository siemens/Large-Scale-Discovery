/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/utils/config"
	"go.uber.org/zap/zapcore"
)

// configWeb stores the loaded web backend configuration
var configWeb = config.NewConfig[Web]()

// Init initializes the configuration module and loads a JSON configuration. If the file does not exist, a default
// configuration is generated.
func Init(confPath string) error { return configWeb.Init(confPath, defaultWebConfigFactory) }

// Load reads a configuration from a file, deserializes it, and sets it as the global configuration.
func Load(confPath string) error { return configWeb.Load(confPath) }

// Set sets a passed configuration as the global configuration.
func Set(conf *Web) { configWeb.Set(conf) }

// Save serializes a given configuration and writes it to a file.
func Save(conf *Web, path string) error { return configWeb.Save(conf, path) }

// GetConfig returns a pointer to the current global configuration.
func GetConfig() *Web { return configWeb.Get() }

// defaultWebConfigFactory returns a new backend configuration with safe defaults
func defaultWebConfigFactory() Web {

	// Prepare default logging settings and adapt for backend
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = filepath.Join("logs", "backend.log")
	logging.Smtp.Connector.Subject = "Backend Error Log"

	// Define default values
	frontendLinks := map[string][]Link{
		"menu": {},
		"repositories": {
			Link{
				Icon:   "",
				Text:   "Active Directory",
				Href:   "#",
				Target: "",
			},
			Link{
				Icon:   "",
				Text:   "Asset Inventory",
				Href:   "#",
				Target: "",
			},
			Link{
				Icon:   "",
				Text:   "Network Inventory",
				Href:   "#",
				Target: "",
			},
		},
		"tooling": {
			Link{
				Icon:   "",
				Text:   "Nmap Port Scanner",
				Href:   "https://nmap.org/",
				Target: "_blank",
			},
			Link{
				Icon:   "",
				Text:   "SSLyze SSL Scanner",
				Href:   "https://github.com/nabla-c0d3/sslyze",
				Target: "_blank",
			},
		},
	}
	jwtLifetime := time.Minute * 20              // 20 minutes
	jwtLifetimeMax := time.Hour * 12             // 12 hours until fresh authentication is required
	accessTokenMaxExpiry := time.Hour * 24 * 365 // Maximum allowed lifetime for a view access token (enforced before passing to manager)
	jwtSecret, _ := utils.GenerateToken(utils.AlphaNumCaseSymbol, 64)

	// Prepare default settings for development
	managerSecret := ""
	if _build.DevMode {
		managerSecret = "dev_secret"
		logging.Console.Level = zapcore.DebugLevel
	}

	// Prepare default web config
	conf := Web{
		ListenAddressHttps: "localhost:443",
		ListenAddressHttp:  "", // Leave empty to disable HTTP endpoint
		ManagerAddress:     "localhost:2222",
		ManagerSsl:         true, // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
		ManagerSecret:      managerSecret,
		FrontendLinks:      frontendLinks,
		RateLimit: RateLimit{
			AuthRequestsPerMinute: 10,
			ApiRequestsPerMinute:  120,
		},
		Jwt: Jwt{
			Secret:         jwtSecret,
			Algorithm:      "HS256",
			ExpiryMinutes:  int(jwtLifetime.Minutes()),
			Expiry:         jwtLifetime,
			RefreshMinutes: int(jwtLifetimeMax.Minutes()),
			Refresh:        jwtLifetimeMax,
			ExpiryApiDays:  365,
		},
		ExpiryTokenMaxDays: int(accessTokenMaxExpiry.Hours() / 24),
		ExpiryTokenMax:     accessTokenMaxExpiry,
		Authenticator: map[string]interface{}{ // Flexible map of arguments as needed by integrated authenticators
			"credentials_registration": false,
			"oauth": map[string]interface{}{ // A name for the oauth authenticator so that there can be multiple ones initialized for different customers
				"oauth_public_key_url": "https://sso.domain.tld/ext/oauth/jwks",
				"oauth_config_url":     "https://sso.domain.tld/.well-known/openid-configuration",
				"oauth_client_id":      "",
				"oauth_client_secret":  "",
				"oauth_scope":          "",
				"oauth_claims_mapping": map[string]string{
					"user_mail":       "mail",
					"user_name":       "given_name",
					"user_surname":    "family_name",
					"user_company":    "company",
					"user_department": "org_code",
					"user_gid":        "gid",
				},
			},
		},
		Loader: map[string]interface{}{ // Flexible map of arguments as needed by integrated connectors
			"ldap_certificate_path": "",
			"ldap_host":             "",
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

// Web is the root configuration object tying all backend settings together.
type Web struct {
	ListenAddressHttps string                 `json:"listen_address_https"` // Leave empty to disable TLS endpoint
	ListenAddressHttp  string                 `json:"listen_address_http"`  // Leave empty to disable HTTP; HTTPS should remain enabled unless a TLS terminator protects the connection.
	ManagerAddress     string                 `json:"manager_address"`
	ManagerSsl         bool                   `json:"manager_ssl"`    // TLS should remain enabled unless a trusted network or TLS terminator protects the connection.
	ManagerSecret      string                 `json:"manager_secret"` // Token to authorize RPC connections to invoke manager RPC methods.
	FrontendLinks      map[string][]Link      `json:"frontend_links"`
	RateLimit          RateLimit              `json:"rate_limit"`
	Jwt                Jwt                    `json:"jwt"`
	ExpiryTokenMaxDays int                    `json:"expiry_token_max_days"` // Maximum allowed lifetime of a view access token (enforced before passing to manager)
	ExpiryTokenMax     time.Duration          `json:"-"`                     //
	Authenticator      map[string]interface{} `json:"authenticator"`         // Arbitrary arguments passed to authenticators. Flexible for own authenticator integrations.
	Loader             map[string]interface{} `json:"loader"`                // Arbitrary arguments passed to connectors. Flexible for own connector integrations.
	Logging            log.Settings           `json:"logging"`
}

// UnmarshalJSON applies defaults before decoding a backend configuration
func (conf *Web) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Web
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

	// Do input validation
	if raw.ExpiryTokenMaxDays <= 0 {
		return fmt.Errorf("invalid maximum access token expiry duration")
	}

	// Copy loaded JSON values to actual config
	*conf = Web(raw)

	// Set unserializable values
	conf.ExpiryTokenMax = time.Duration(raw.ExpiryTokenMaxDays) * 24 * time.Hour

	// Return nil as everything is valid
	return nil
}

// Link defines a frontend navigation link supplied by the backend
type Link struct {
	Icon   string `json:"icon"`
	Text   string `json:"text"`
	Href   string `json:"href"`
	Target string `json:"target"`
}

// RateLimit configures request limits for an endpoint group
type RateLimit struct {
	AuthRequestsPerMinute int `json:"auth_requests_per_minute"` // AuthRequestsPerMinute limits login/auth endpoint requests per source IP. Zero disables the limit.
	ApiRequestsPerMinute  int `json:"api_requests_per_minute"`  // ApiRequestsPerMinute limits general authenticated API requests per user. Zero disables the limit.
}

// Jwt configures JSON Web Token creation and refresh behavior
type Jwt struct {
	Secret         string        `json:"secret"`          // Secret to encrypt JWT token with
	Algorithm      string        `json:"algorithm"`       // Algorithm to encrypt JWT token with
	ExpiryMinutes  int           `json:"expiry_minutes"`  // Max time a JWT token is valid
	Expiry         time.Duration `json:"-"`               //
	RefreshMinutes int           `json:"refresh_minutes"` // Max time a JWT token can be refreshed
	Refresh        time.Duration `json:"-"`               //
	ExpiryApiDays  int           `json:"expiry_api_days"` // Max time a user-issued API token is valid
	ExpiryApi      time.Duration `json:"-"`               //
}

// UnmarshalJSON applies defaults before decoding JSON Web Token settings
func (j *Jwt) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Jwt
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	errUnmarshal := json.Unmarshal(b, &raw)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Copy loaded JSON values to actual config
	*j = Jwt(raw)

	// Set unserializable values
	j.Expiry = time.Duration(raw.ExpiryMinutes) * time.Minute
	j.Refresh = time.Duration(raw.RefreshMinutes) * time.Minute
	j.ExpiryApi = time.Duration(raw.ExpiryApiDays) * 24 * time.Hour

	// Return nil as everything is valid
	return nil
}
