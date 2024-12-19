/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package config

import (
	"encoding/json"
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"path/filepath"
)

var templatePaths = Paths{
	Nmap:   "/usr/bin/nmap",
	Python: "/usr/bin/python3",
}

var templateAuthentication = Authentication{
	map[string]map[string]string{},
	templateCredentialsLdap,
	templateCredentialsWeb,
	templateCredentialsWeb,
}

var templateModules = Modules{
	Banner: ModuleBanner{
		Module: Module{
			MaxInstances: -1,
		},
	},
	Discovery: ModuleDiscovery{
		Module: Module{
			MaxInstances: -1,
		},
		LdapServerComment:  "If *no* LDAP server is configured, the respective scan target's domain will be queried. Cross-domain queries might only work with implicit LDAP authentication on Windows.",
		LdapServer:         "",
		BlacklistFile:      "",
		DomainOrderComment: "Sometimes there might be multiple DNS names discovered for a single host. With this grouped and ordered list of domains, you can force them into a deterministic order to promote the most plausible one. E.g. allows to prefer domain.local over domain.com.",
		DomainOrder: []string{
			"local",
			"sub1.local",
			"sub2.local",
			"third-party.com",
		},
	},
	Nfs: ModuleNfs{
		Module: Module{
			MaxInstances: -1,
		},
	},
	Ssh: ModuleSsh{
		Module: Module{
			MaxInstances: -1,
		},
	},
	Ssl: ModuleSsl{
		Module: Module{
			MaxInstances: -1,
		},
		Comment:              "SSL certificates will always be validated against default browser's trust stores. Additionally, they will be matched against the local OS' trust store, unless you want to set a custom one!",
		CustomTruststoreFile: "",
	},
	Webcrawler: ModuleWebcrawler{
		Module: Module{
			MaxInstances: -1,
		},
		Download:     false,
		DownloadPath: "",
		DownloadTypes: []string{
			"application/pdf", "application/msword", "application/vnd.ms-excel", "vnd.ms-excel.addin.macroEnabled.12",
			"vnd.ms-excel.sheet.binary.macroEnabled.12", "vnd.ms-excel.sheet.macroEnabled.12",
			"vnd.ms-excel.template.macroEnabled.12", "application/vnd.ms-word.document.macroEnabled.12",
			"vnd.ms-word.template.macroEnabled.12", "application/vnd.ms-word.template.macroEnabled.12",
		},
	},
	Webenum: ModuleWebenum{
		Module: Module{
			MaxInstances: -1,
		},
	},
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
	Inventories map[string]map[string]string `json:"inventories"` // Flexible configuration construct for asset inventory plugins
	Ldap        Credentials                  `json:"ldap"`        // Used by Discovery module for AD queries
	Webcrawler  Credentials                  `json:"webcrawler"`  // Used by webcrawler module for testing
	Webenum     Credentials                  `json:"webenum"`     // Used by webenum module for testing
}

type Modules struct {
	// Module specific configurations. Values set here, should only be required by the associated scan module or are
	// meant to override other generic values.
	Banner     ModuleBanner     `json:"banner"`
	Discovery  ModuleDiscovery  `json:"discovery"`
	Nfs        ModuleNfs        `json:"nfs"`
	Ssh        ModuleSsh        `json:"ssh"`
	Ssl        ModuleSsl        `json:"ssl"`
	Webcrawler ModuleWebcrawler `json:"webcrawler"`
	Webenum    ModuleWebenum    `json:"webenum"`
}

func (m *Modules) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux Modules
	var raw aux

	// Prepare default module values
	mod := Module{
		MaxInstances: -1,
	}

	// Set default value if no other value is present in the read Json file
	raw.Banner.Module = mod
	raw.Discovery.Module = mod
	raw.Nfs.Module = mod
	raw.Ssh.Module = mod
	raw.Ssl.Module = mod
	raw.Webcrawler.Module = mod
	raw.Webenum.Module = mod

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Update struct with de-serialized values
	*m = Modules(raw)

	// Return nil as everything went fine
	return nil
}

// ReadMaxInstances retrieves a module's max instances limit
func (m *Modules) ReadMaxInstances(label string) int {
	switch label {
	case discovery.Label:
		return m.Discovery.MaxInstances
	case banner.Label:
		return m.Banner.MaxInstances
	case nfs.Label:
		return m.Nfs.MaxInstances
	case ssl.Label:
		return m.Ssl.MaxInstances
	case ssh.Label:
		return m.Ssh.MaxInstances
	case webcrawler.Label:
		return m.Webcrawler.MaxInstances
	case webenum.Label:
		return m.Webenum.MaxInstances
	default:
		return -1
	} // Switch End
}
