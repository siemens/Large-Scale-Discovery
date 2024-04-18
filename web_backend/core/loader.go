/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"strings"
)

var loaders []Loader                // List of registered loaders to be initialized
var loadersLookup map[string]Loader // Map of initialized loader module, referenced by user domain for quick lookup

type Loader interface {
	Domains() []string
	Init(conf map[string]interface{}) error
	RefreshUser(logger scanUtils.Logger, user *database.T_user) (errTemporary error, errInternal error, errPublic string) // May update the passed user struct, but does not yet commit changes
}

// GetLoader retrieves the appropriate loader for the given domain or falls back to the default loader
func GetLoader(logger scanUtils.Logger, userEmail string) Loader {

	// Check if input can be processed
	if strings.Count(userEmail, "@") == 1 {

		// Extract user domain from e-mail address
		userDomain := userEmail[strings.LastIndex(userEmail, "@"):]

		// Lookup loader
		l, ok := loadersLookup[userDomain]

		// Return associated loader if dedicated one is available
		if ok {
			logger.Debugf("Using dedicated loader for '%s'.", userDomain)
			return l
		}
	}

	// Get default loader
	logger.Debugf("Using default loader.")
	l, _ := loadersLookup[""]

	// Return selected loader
	return l
}

// initLoaders initializes registered loaders
func initLoaders(confLoaders map[string]interface{}) error {

	// Prepare memory for temporary list of initialized loaders
	// After initialization, all user domains will be start with @, for security reasons
	loadersLookup = make(map[string]Loader, len(loaders))

	// Initialize registered loaders
	for _, loader := range loaders {

		// Initialize as default loader, if it isn't configured for specific domains
		if len(loader.Domains()) == 0 {

			// Check if default loader is already registered
			_, loaderExists := loadersLookup[""]
			if loaderExists {
				return fmt.Errorf("multiple default loaders configured")
			}

			// Register loader as the default loader
			loadersLookup[""] = loader
		} else {

			// Add reference to loader lookup map for each of the loader's responsible user domains
			for _, loaderDomains := range loader.Domains() {

				// Abort if empty value (reserved for default loader) was specified among user domains
				if len(loaderDomains) == 0 {
					return fmt.Errorf("default loader cannot be registerd along with other user domains")
				}

				// Check if domain is formatted correctly
				if !strings.HasPrefix(loaderDomains, "@") {
					return fmt.Errorf("loader user domains must start with @")
				}

				// Make sure there wasn't another @ contained
				if strings.Count(loaderDomains, "@") != 1 {
					return fmt.Errorf(
						"invalid loader user domain '%s'", loaderDomains)
				}

				// Check if loader for user domain is already registered
				_, loaderExists := loadersLookup[loaderDomains]
				if loaderExists {
					return fmt.Errorf("multiple '%s' loaders configured", loaderDomains)
				}

				// Register user domain with reference to this loaer
				loadersLookup[loaderDomains] = loader
			}
		}

		// Initialize loader
		errLoader := loader.Init(confLoaders)
		if errLoader != nil {
			return fmt.Errorf(
				"could not initialize loader for '%s': %s",
				strings.Join(loader.Domains(), ", "),
				errLoader,
			)
		}
	}

	// Return nil as everything went fine
	return nil
}
