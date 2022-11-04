/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package _test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// Settings necessary for some unit tests
var settings *Settings
var settingsErr error // Indicates if settings initialization failed
var once sync.Once

type Settings struct {
	PathDataDir  string // Path to sample data used by unit tests
	PathOpenssl  string // Path to the openSSL executable (needed for encrypted email)
	LogRecipient string // Recipient for log messages generated during unit test execution
	LdapHost     string // LDAP host to test LDAP queries with
	LdapUser     string // Username to query LDAP with
	LdapPassword string // Password to query LDAP with
}

func GetSettings() (*Settings, error) {

	// Initialize unit test settings if not done yet
	once.Do(func() {

		// Get absolute path to bin folder
		_, filename, _, _ := runtime.Caller(0)
		workingDir := filepath.Dir(filename)
		workingDir = filepath.Join(workingDir, "../", "_test")

		// Changes working directory to the bin folder.
		err := os.Chdir(workingDir)
		if err != nil {
			fmt.Println("Error ", err.Error())
		}

		// Create a new instance of the unit test settings, that might need to be adapted before running unit tests
		settings = &Settings{
			PathOpenssl:  "C:\\Program Files\\OpenSSL-Win64\\bin\\openssl.exe", // CONFIGURE BEFORE RUNNING UNIT TESTS
			LogRecipient: "user@domain.tld",                                    // CONFIGURE BEFORE RUNNING UNIT TESTS. Must be valid address or test will get stuck!
			LdapHost:     "",                                                   // must be set to enable respective LDAP unit tests!
			LdapUser:     "",                                                   // must be set to enable respective LDAP unit tests!
			LdapPassword: "",                                                   // must be set to enable respective LDAP unit tests!
		}

		// Check if settings are valid
		_, errPathOpenssl := exec.Command(settings.PathOpenssl).CombinedOutput()
		if errPathOpenssl != nil {
			settingsErr = fmt.Errorf("invalid OpenSSL path")
			return
		}

		// Set static values which should not be changed
		settings.PathDataDir = filepath.Join(workingDir, "data")
	})

	// Return previously initialized unit test settings
	return settings, settingsErr
}
