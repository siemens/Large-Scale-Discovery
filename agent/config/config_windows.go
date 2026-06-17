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

	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
)

var templatePaths = Paths{
	Nmap:   "./tools/nmap-7.92/nmap.exe",
	Sslyze: "./tools/sslyze-5.0.5/sslyze.exe",
}

var templateAuthentication = Authentication{
	map[string]map[string]string{},
	templateCredentialsLdap,
	templateCredentialsNuclei,
	templateCredentialsSmb,
	templateCredentialsWeb,
	templateCredentialsWeb,
}

var templateModules = Modules{
	Banner: ModuleBanner{
		MaxInstances: -1,
	},
	Discovery: ModuleDiscovery{
		MaxInstances:       -1,
		LdapServerComment:  "If *no* static LDAP server is configured, the target server will be dynamically derived from the discovered target's domain. If no credentials are configured, cross-domain queries might work on domain-joined Windows hosts via implicitly authentication. If credentials are configured, cross-domain queries might work OS and domain membership independent with explicit authentication via GSSAPI.",
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
		MaxInstances: -1,
	},
	Nuclei: ModuleNuclei{
		MaxInstances: -1,
	},
	Smb: ModuleSmb{
		MaxInstances: -1,
	},
	Ssh: ModuleSsh{
		MaxInstances: -1,
	},
	Ssl: ModuleSsl{
		MaxInstances:         -1,
		Comment:              "SSL certificates will always be validated against default browser's trust stores. Additionally, they will be matched against the local OS' trust store, unless you want to set a custom one!",
		CustomTruststoreFile: "",
	},
	Webcrawler: ModuleWebcrawler{
		MaxInstances: -1,
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
		MaxInstances: -1,
	},
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

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Paths
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
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

	// Copy loaded JSON values to actual config
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
	Ldap        CredentialsGssapi            `json:"ldap"`        // Used by Discovery module for AD queries
	Nuclei      CredentialsNuclei            `json:"nuclei"`      // Used by Nuclei modules for testing
	Smb         Credentials                  `json:"smb"`         // Used by SMB module for testing
	Webcrawler  Credentials                  `json:"webcrawler"`  // Used by webcrawler module for testing
	Webenum     Credentials                  `json:"webenum"`     // Used by webenum module for testing
}

type ModuleSmb struct {
	MaxInstances int `json:"max_instances"`
}

func (m *ModuleSmb) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux ModuleSmb
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
	*m = ModuleSmb(raw)

	// Return nil as everything went fine
	return nil
}

type Modules struct {
	// Module specific configurations. Values set here, should only be required by the associated scan module or are
	// meant to override other generic values.
	Banner     ModuleBanner     `json:"banner"`
	Discovery  ModuleDiscovery  `json:"discovery"`
	Nfs        ModuleNfs        `json:"nfs"`
	Nuclei     ModuleNuclei     `json:"nuclei"`
	Smb        ModuleSmb        `json:"smb"`
	Ssh        ModuleSsh        `json:"ssh"`
	Ssl        ModuleSsl        `json:"ssl"`
	Webcrawler ModuleWebcrawler `json:"webcrawler"`
	Webenum    ModuleWebenum    `json:"webenum"`
}

func (m *Modules) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux Modules
	var raw aux

	// Set default value if no other value is present in the read JSON file
	// This is necessary in case a whole branch is missing in the JSON, which would the respective child's Unmarshalling.
	raw.Banner.MaxInstances = -1
	raw.Discovery.MaxInstances = -1
	raw.Nfs.MaxInstances = -1
	raw.Nuclei.MaxInstances = -1
	raw.Smb.MaxInstances = -1
	raw.Ssh.MaxInstances = -1
	raw.Ssl.MaxInstances = -1
	raw.Webcrawler.MaxInstances = -1
	raw.Webenum.MaxInstances = -1

	// Unmarshal serialized JSON into temporary auxiliary structure
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
	case nuclei.Label:
		return m.Nuclei.MaxInstances
	case smb.Label:
		return m.Smb.MaxInstances
	case ssl.Label:
		return m.Ssl.MaxInstances
	case ssh.Label:
		return m.Ssh.MaxInstances
	case webcrawler.Label:
		return m.Webcrawler.MaxInstances
	case webenum.Label:
		return m.Webenum.MaxInstances
	default:
		return -1 // No limit configured
	} // Switch End
}
