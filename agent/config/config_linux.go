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
	"path/filepath"
)

var templatePaths = Paths{
	Nmap:   "/usr/bin/nmap",
	Python: "/usr/bin/python3",
}

var templateAuthentication = Authentication{
	templateCredentialsLdap,
	templateCredentialsWeb,
	templateCredentialsWeb,
}

type Paths struct {
	// Paths to executables, e.g. of third party tools. Use complete paths to executables to avoid cross-platform
	// issues. E.g. on Linux the Python executable files might have different names (python, python3.7, python3.8)
	NmapDir   string `json:"-"`
	Nmap      string `json:"nmap"`
	PythonDir string `json:"-"`
	Python    string `json:"python"`
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
	if raw.Python == "" {
		return fmt.Errorf("Python path not set")
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

	python := raw.Python
	if !filepath.IsAbs(python) {
		var errAbs error
		python, errAbs = filepath.Abs(python)
		if errAbs != nil {
			return errAbs
		}
	}

	// Do input validation
	errNmap := scanUtils.IsValidExecutable(nmap, "-h") // args required on linux
	if errNmap != nil {
		return errNmap
	}
	errPython := scanUtils.IsValidExecutable(python, "-h") // args required on linux
	if errPython != nil {
		return errPython
	}

	// Calculate Nmap and Python dir
	nmapDir := filepath.Dir(nmap)
	errNmapDir := scanUtils.IsValidFolder(nmapDir)
	if errNmapDir != nil {
		return errNmapDir
	}

	pythonDir := filepath.Dir(python)
	errPythonDir := scanUtils.IsValidFolder(pythonDir)
	if errPythonDir != nil {
		return errPythonDir
	}

	// Copy loaded Json values to actual config
	*p = Paths{
		NmapDir:   nmapDir,
		Nmap:      nmap,
		PythonDir: pythonDir,
		Python:    python,
	}

	// Return nil as everything is valid
	return nil
}

type Authentication struct {
	Ldap       Credentials `json:"ldap"`       // Used by Discovery module for AD queries
	Webcrawler Credentials `json:"webcrawler"` // Used by webcrawler module for testing
	Webenum    Credentials `json:"webenum"`    // Used by webenum module for testing
}
