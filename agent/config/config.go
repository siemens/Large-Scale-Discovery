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

var agentConfig = &AgentConfig{} // Global configuration
var agentConfigLock sync.Mutex   // Lock required to avoid simultaneous requesting/updating of config
var templateCredentialsLdap = CredentialsGssapi{
	Credentials: Credentials{
		Comment:  "If *no* explicit LDAP credentials are configured, implicit authentication will be tried on Windows. Implicit authentication does not work on Linux and queries would be skipped.",
		Domain:   "",
		User:     "",
		Password: "",
	},
	DisableGssapi: false,
}
var templateCredentialsNuclei = CredentialsNuclei{
	Comment:  "If *no* explicit credentials are configured, templates requiring these variables will be skipped.",
	User:     "",
	Password: "",
}

//lint:ignore U1000 used in Windows-only builds
var templateCredentialsSmb = Credentials{
	Comment:  "If *no* explicit SMB credentials are configured, implicit authentication will be tried on Windows. SMB crawling is not supported on Linux and would be skipped.",
	Domain:   "",
	User:     "",
	Password: "",
}
var templateCredentialsWeb = Credentials{
	Comment:  "If *no* explicit credentials are configured, authentication will be skipped. Works both, on Windows and Linux.",
	Domain:   "",
	User:     "",
	Password: "",
}

// Init initializes the configuration module and loads a JSON configuration. If JSON is not existing, a default
// configuration will be generated.
func Init(configFile string) error {
	if errFile := scanUtils.IsValidFile(configFile); errFile != nil {
		defaultConf := defaultAgentConfigFactory()
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
	agentConfigLock.Lock()
	defer agentConfigLock.Unlock()

	// Prepare new config, don't work on the global values
	newConfig := &AgentConfig{}

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
	agentConfig = newConfig

	// Return nil to indicate successful config update
	return nil
}

// Set sets a passed configuration as the global configuration
func Set(conf *AgentConfig) {

	// Lock global config before initializing an update
	agentConfigLock.Lock()
	defer agentConfigLock.Unlock()

	// Replace global configuration with new one
	agentConfig = conf
}

// Save serializes a given configuration and writes it to a file
func Save(conf *AgentConfig, path string) error {

	// Lock global config, because the given config pointer might be the global config
	agentConfigLock.Lock()
	defer agentConfigLock.Unlock()

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
func GetConfig() *AgentConfig {

	// The global configuration might get updated regularly to allow user updating settings without aborting scans
	agentConfigLock.Lock()
	defer agentConfigLock.Unlock()

	// Return current global configuration
	return agentConfig
}

func defaultAgentConfigFactory() AgentConfig {

	// Prepare default logging settings and adapt for agent
	logging := log.DefaultLogSettingsFactory()
	logging.File.Path = filepath.Join("logs", "agent.log")
	logging.Smtp.Connector.Subject = "Agent Error Log"

	// Prepare default settings for development
	var scopeSecrets = make([]string, 0)
	if _build.DevMode {
		scopeSecrets = []string{"dev_secret"}
		logging.Console.Level = zapcore.DebugLevel
	}

	// Prepare default agent config
	conf := AgentConfig{
		BrokerAddress:  "localhost:3333",
		ScopeSecrets:   scopeSecrets,
		Paths:          templatePaths,
		Authentication: templateAuthentication,
		Logging:        logging,
		Modules:        templateModules,
	}

	// Return generated config
	return conf
}

//
// JSON structure of configuration
//

type Credentials struct {
	Comment  string `json:"comment"`
	Domain   string `json:"domain"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type CredentialsGssapi struct {
	Credentials
	DisableGssapi bool `json:"disable_gssapi"`
}
type CredentialsNuclei struct {
	Comment  string `json:"comment"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ModuleBanner struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleBanner) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleBanner
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Update struct with de-serialized values
	*m = ModuleBanner(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleDiscovery struct {
	MaxInstances int `json:"max_instances"`
	// Discovery-specific configuration values.
	LdapServerComment  string   `json:"ldap_server_comment"`
	LdapServer         string   `json:"ldap_server"`
	BlacklistFile      string   `json:"blacklist_file"`
	DomainOrderComment string   `json:"domain_order_comment"`
	DomainOrder        []string `json:"domain_order"`
}

func (m *ModuleDiscovery) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleDiscovery
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Do input validation
	if raw.BlacklistFile != "" {
		errBlacklist := scanUtils.IsValidFile(raw.BlacklistFile)
		if errBlacklist != nil {
			return fmt.Errorf("blacklist %s", errBlacklist) // results in e.g. "blacklist path not a file"
		}
	}
	if raw.LdapServer != "" && !scanUtils.IsValidHostname(raw.LdapServer) && !scanUtils.IsValidIp(raw.LdapServer) {
		return fmt.Errorf("hostname or IP expected as LDAP server")
	}

	// Update struct with de-serialized values
	*m = ModuleDiscovery(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleNfs struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleNfs) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleNfs
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Update struct with de-serialized values
	*m = ModuleNfs(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleNuclei struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleNuclei) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleNuclei
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Update struct with de-serialized values
	*m = ModuleNuclei(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleSsh struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleSsh) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleSsh
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Update struct with de-serialized values
	*m = ModuleSsh(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleSsl struct {
	MaxInstances int `json:"max_instances"`
	// Ssl-specific configuration values.
	Comment              string `json:"custom_truststore_file_comment"`
	CustomTruststoreFile string `json:"custom_truststore_file"` // Path to custom trust store. Otherwise, OS one is used
}

func (m *ModuleSsl) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleSsl
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Do input validation
	if raw.CustomTruststoreFile != "" {
		errTruststore := scanUtils.IsValidFile(raw.CustomTruststoreFile)
		if errTruststore != nil {
			return fmt.Errorf("trust store %s", errTruststore) // results in e.g. "truststore path not a file"
		}
	}

	// Update struct with de-serialized values
	*m = ModuleSsl(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleWebcrawler struct {
	MaxInstances int `json:"max_instances"`
	// Webcrawler-specific configuration values.
	Download      bool     `json:"download_files"` // Whether to download downloadable contents
	DownloadPath  string   `json:"download_path"`  // Path to the folder to download files to. If empty the working directory is chosen.
	DownloadTypes []string `json:"download_types"` // Response content types to download
}

func (m *ModuleWebcrawler) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleWebcrawler
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Do input validation
	if raw.DownloadPath != "" {
		if errPath := scanUtils.IsValidFolder(raw.DownloadPath); errPath != nil {
			return fmt.Errorf("download path: %s", errPath) // results in e.g. "download path: path not a folder"
		}
	}

	// Update struct with de-serialized values
	*m = ModuleWebcrawler(raw)

	// Return nil as everything went fine
	return nil
}

type ModuleWebenum struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleWebenum) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleWebenum
	var raw aux

	// Set default value if no other value is present in the read JSON file
	raw.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Fix invalid settings
	if raw.MaxInstances < -1 {
		raw.MaxInstances = -1
	}

	// Update struct with de-serialized values
	*m = ModuleWebenum(raw)

	// Return nil as everything went fine
	return nil
}

type AgentConfig struct {
	// The root configuration object tying all configuration segments together.
	BrokerAddress  string         `json:"broker_address"`
	ScopeSecrets   []string       `json:"scope_secrets"`
	Paths          Paths          `json:"paths"`
	Authentication Authentication `json:"authentication"`
	Logging        log.Settings   `json:"logging"`
	Modules        Modules        `json:"modules"`
	OtInterfaces   []string       `json:"ot_interfaces"` // The network interface to run the OT scans on, e.g. "eth0". Only relevant for OT scan scopes. Empty list scans all non-loopback interfaces.
}

func (c *AgentConfig) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux AgentConfig
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check if there is at least one scope secret defined
	if len(raw.ScopeSecrets) == 0 {
		return fmt.Errorf("no scope secret")
	}

	// Validate defined scope secrets
	for _, secret := range raw.ScopeSecrets {
		if len(secret) < 10 {
			return fmt.Errorf("invalid scope secret")
		}
	}

	// Update struct with de-serialized values
	*c = AgentConfig(raw)

	// Return nil as everything went fine
	return nil
}
