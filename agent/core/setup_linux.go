/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"io/ioutil"
	"large-scale-discovery/agent/config"
	"large-scale-discovery/log"
	"os"
	"path/filepath"
)

var osTruststorePaths = []string{
	"/etc/ssl/certs/",
}

// checkConfigDependant tests OS specific configuration values by trying to initialize scan modules with them. This allows to
// discover invalid configurations at startup, instead of during runtime. Dynamic target arguments are replaced by
// dummy data.
func checkConfigDependant() error {

	// Dummy scan target arguments
	dummyLogger := log.GetLogger().Tagged("checkConfigDependant")
	dummyTarget := "127.0.0.1"
	dummyPort := 0
	dummyOtherNames := []string{"a", "b"}

	// Get config
	conf := config.GetConfig()

	// Decide truststore for SSL test
	var sslyzeAdditionalTruststore string
	if len(conf.Modules.Ssl.CustomTruststoreFile) == 0 {
		sslyzeAdditionalTruststore = SslOsTruststoreFile
	} else {
		sslyzeAdditionalTruststore = conf.Modules.Ssl.CustomTruststoreFile
	}

	// Run Ssl test
	_, errSsl := ssl.NewScanner(
		dummyLogger,
		conf.Paths.Python,
		sslyzeAdditionalTruststore, // The ssl scan module will validate this path
		dummyTarget,
		dummyPort,
		dummyOtherNames,
	)
	if errSsl != nil {
		return fmt.Errorf("'%s': %s", ssl.Label, errSsl)
	}

	// Return nil if everything went fine
	return nil
}

// Linux implementation OS trust store generation
func generateTruststoreOs(truststoreOutputFile string) error {

	// Prepare OS trust store file
	outputFile, errOpen := os.OpenFile(truststoreOutputFile, os.O_CREATE|os.O_WRONLY, 0660)
	if errOpen != nil {
		return fmt.Errorf("could not create OS trust store '%s': %s", truststoreOutputFile, errOpen)
	}

	// Make sure file gets closed on exit
	defer func() { _ = outputFile.Close() }()

	// Iterate common Linux directories with system-wide CA certificates
	for _, osTruststorePath := range osTruststorePaths {

		// Skip if directory is not existing
		errValidate := scanUtils.IsValidFolder(osTruststorePath)
		if errValidate == nil {
			continue
		}

		// Read directory
		files, errReadDir := ioutil.ReadDir(osTruststorePath)
		if errReadDir != nil {
			return fmt.Errorf("could not read trust store directory '%s': %s", osTruststorePath, errReadDir)
		}

		// Iterate directory files
		for _, f := range files {
			if !f.IsDir() {

				// Build path
				filePath := filepath.Join(osTruststorePath, f.Name())

				// Read file
				data, errRead := ioutil.ReadFile(filePath)
				if errRead != nil {
					continue
				}

				// Write to OS trust store
				_, errWrite := outputFile.Write(data)
				if errWrite != nil {
					return fmt.Errorf("could not write OS trust store '%s': %s", osTruststorePath, errWrite)
				}
			}
		}
	}

	// Return nil as everything went fine
	return nil
}
