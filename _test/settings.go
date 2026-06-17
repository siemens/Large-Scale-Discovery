/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package _test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Settings necessary for some unit tests
var settings *Settings
var once sync.Once

type Settings struct {
	LdapHost     string // LDAP host to test LDAP queries with
	LdapUser     string // Username to query LDAP with
	LdapPassword string // Password to query LDAP with
	PgHost       string // PostgreSQL host for integration tests; leave empty to skip PG tests
	PgPort       int    // PostgreSQL port (default: 5432)
	PgUser       string // PostgreSQL admin user (default: postgres)
	PgPassword   string // PostgreSQL admin password
}

// GetSettings returns test settings that should be used by unit tests.
// Invalid settings will be changed to empty values.
// Unit test should decide themselves which of these settings are mandatory and check their availability.
// Unit tests should always run as comprehensive as possible with the current configuration.
func GetSettings() *Settings {

	// Initialize unit test settings if not done yet
	once.Do(func() {

		// Get absolute path to bin folder
		_, filename, _, _ := runtime.Caller(0) // File path of _test.settings.go
		workingDir := filepath.Dir(filename)   // Dir path of _test

		///////////////////////////////////////////////////////////////////////
		// CONFIGURE BEFORE RUNNING UNIT TESTS TO INCREASE COVERAGE ==========>
		// EVERYTHING THAT IS NOT SET CORRECTLY WILL LEAD TO SKIPPED UNIT TESTS
		///////////////////////////////////////////////////////////////////////
		ldapHost := ""       // must be set to enable respective unit tests!
		ldapUser := ""       // must be set to enable respective unit tests!
		ldapPassword := ""   // must be set to enable respective unit tests!
		pgHost := ""         // must be set to enable PostgreSQL integration tests!
		pgPort := 5432       // default PostgreSQL port
		pgUser := "postgres" // default PostgreSQL admin user
		pgPassword := ""     // set if authentication is required
		// proxy, _ = url.Parse("http://127.0.0.1:8080") // ATTENTION: Responses might look different via proxy!!
		///////////////////////////////////////////////////////////////////////
		// <========== CONFIGURE BEFORE RUNNING UNIT TESTS TO INCREASE COVERAGE
		///////////////////////////////////////////////////////////////////////

		// Changes working directory to the bin folder.
		err := os.Chdir(workingDir)
		if err != nil {
			panic(fmt.Sprintf("could not set working directory for unit tests: %s", err))
		}

		// Create a new instance of the unit test settings, that might need to be adapted before running unit tests
		settings = &Settings{
			LdapHost:     ldapHost,
			LdapUser:     ldapUser,
			LdapPassword: ldapPassword,
			PgHost:       pgHost,
			PgPort:       pgPort,
			PgUser:       pgUser,
			PgPassword:   pgPassword,
		}
	})

	// Return previously initialized unit test settings
	return settings
}
