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
	"sync"

	scanUtils "github.com/siemens/GoScans/utils"
)

// Config is a generic, thread-safe configuration holder that encapsulates the Init/Load/Save/Set/Get
// boilerplate shared across all components.
type Config[T any] struct {
	conf *T
	mu   sync.Mutex
}

// NewConfig creates a Config initialised with the zero value of T.
func NewConfig[T any]() *Config[T] {
	return &Config[T]{
		conf: new(T),
	}
}

// Init loads the configuration from configFile. If the file does not exist, a default configuration is
// created via defaultFactory, saved to configFile, and an error is returned prompting the user to review it.
func (config *Config[T]) Init(path string, defaultFactory func() T) error {
	if errFile := scanUtils.IsValidFile(path); errFile != nil {
		defaultConf := defaultFactory()
		errSave := config.Save(&defaultConf, path)
		if errSave != nil {
			return fmt.Errorf("could not initialize configuration in '%s': %s", path, errSave)
		} else {
			return fmt.Errorf("no configuration, created default in '%s'", path)
		}
	} else {
		errLoad := config.Load(path)
		if errLoad != nil {
			return errLoad
		} else {
			return nil
		}
	}
}

// Load reads a configuration from a file, deserializes it, and sets it as the active configuration.
func (config *Config[T]) Load(path string) error {

	// Lock config before initializing an update
	config.mu.Lock()
	defer config.mu.Unlock()

	// Prepare new config, don't work on the active value
	configNew := new(T)

	// Read file content
	rawJson, errRawJson := os.ReadFile(path)
	if errRawJson != nil {
		return errRawJson
	}

	// Parse JSON
	errUnmarshal := json.Unmarshal(rawJson, configNew)
	if errUnmarshal != nil {
		return errUnmarshal
	}

	// Replace active configuration with new one
	config.conf = configNew

	// Return nil to indicate successful config update
	return nil
}

// Set replaces the active configuration with the provided one.
func (config *Config[T]) Set(conf *T) {

	// Lock config before initializing an update
	config.mu.Lock()
	defer config.mu.Unlock()

	// Replace active configuration with new one
	config.conf = conf
}

// Save serializes a given configuration and writes it to a file.
func (config *Config[T]) Save(conf *T, path string) error {

	// Lock config, because the given config pointer might be the active config
	config.mu.Lock()
	defer config.mu.Unlock()

	// Serialize to JSON
	file, errFile := json.MarshalIndent(conf, "", "    ")
	if errFile != nil {
		return errFile
	}

	// Write JSON to file
	errWrite := os.WriteFile(path, file, 0644)
	if errWrite != nil {
		return errWrite
	}

	// Return nil to indicate successful storage
	return nil
}

// Get returns a pointer to the current active configuration.
func (config *Config[T]) Get() *T {

	// The configuration might get updated regularly to allow user updating settings without aborting scans
	config.mu.Lock()
	defer config.mu.Unlock()

	// Return current active configuration
	return config.conf
}
