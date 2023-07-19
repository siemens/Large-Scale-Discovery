/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
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
	"path/filepath"
)

var templatePaths = Paths{
	Nmap:   "./tools/nmap-7.91/nmap.exe",
	Sslyze: "./tools/sslyze-5.0.5/sslyze.exe",
}

var templateAuthentication = Authentication{
	map[string]map[string]string{},
	templateCredentialsLdap,
	templateCredentialsSmb,
	templateCredentialsWeb,
	templateCredentialsWeb,
}

type Paths struct {
	// issues. E.g. on Linux the Python executable files might have different names (python, python3.7, python3.8)
	// Paths to executables, e.g. of third party tools. Use complete paths to executables to avoid cross-platform
	NmapDir   string `json:"-"`
	Nmap      string `json:"nmap"`
	SslyzeDir string `json:"-"`
	Sslyze    string `json:"sslyze"`
}

func (p *Paths) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux Paths
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check if paths are set
	if raw.Nmap == "" {
		return fmt.Errorf("Nmap path not set")
	}
	if raw.Sslyze == "" {
		return fmt.Errorf("SSLyze path not set")
	}

	// Convert the paths to an absolute ones if necessary
	nmap := raw.Nmap
	if !filepath.IsAbs(nmap) {
		var errAbs error
		nmap, errAbs = filepath.Abs(nmap)
		if errAbs != nil {
			return errAbs
		}
	}

	sslyze := raw.Sslyze
	if !filepath.IsAbs(sslyze) {
		var errAbs error
		sslyze, errAbs = filepath.Abs(sslyze)
		if errAbs != nil {
			return errAbs
		}
	}

	// Do input validation
	errNmap := scanUtils.IsValidExecutable(nmap, "-h") // args required
	if errNmap != nil {
		return errNmap
	}

	errSslyze := scanUtils.IsValidExecutable(sslyze, "-h") // args required
	if errSslyze != nil {
		return errSslyze
	}

	// Calculate Nmap and SSLyze dir
	nmapDir := filepath.Dir(nmap)
	errNmapDir := scanUtils.IsValidFolder(nmapDir)
	if errNmapDir != nil {
		return errNmapDir
	}

	sslyzeDir := filepath.Dir(sslyze)
	errSslyzeDir := scanUtils.IsValidFolder(sslyzeDir)
	if errSslyzeDir != nil {
		return errSslyzeDir
	}

	// Copy loaded Json values to actual config
	*p = Paths{
		NmapDir:   nmapDir,
		Nmap:      nmap,
		SslyzeDir: sslyzeDir,
		Sslyze:    sslyze,
	}

	// Return nil as everything is valid
	return nil
}

type Authentication struct {
	Inventories map[string]map[string]string `json:"inventories"` // Flexible configuration construct for asset inventory plugins
	Ldap        Credentials                  `json:"ldap"`        // Used by Discovery module for AD queries
	Smb         Credentials                  `json:"smb"`         // Used by SMB module for testing
	Webcrawler  Credentials                  `json:"webcrawler"`  // Used by webcrawler module for testing
	Webenum     Credentials                  `json:"webenum"`     // Used by webenum module for testing
}
