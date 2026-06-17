/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
)

// init automatically registers this import type during importer initialization. This way, you can add your custom
// importers without changing the core code.
func init() {
	importers["assets"] = &ImporterAssets{}
}

type ImporterAssets struct{}

// Init initializes the importer, which is happening during application launch
func (i *ImporterAssets) Init(conf map[string]interface{}) error {

	// TODO Implement, initialization of connection or verification of connection details to remote repository

	// Return nil as everything went fine
	return nil
}

// Import retrieves a list of targets as it should be deployed in a scopedb. This list is compared to the currently
// deployed one. The core will split it into lists of targets to be created, removed and updated. The scopedb will
// be changed accordingly.
func (i *ImporterAssets) Import(
	filters Filters, // Filters set for the scan scope to select desired assets from repository
) ([]managerdb.T_discovery, error) {

	// TODO Implement, loading list of scan inputs from remote repository
	// Optional synchronization settings (filters) can be found in scanScope.Attributes[...]

	// Return results or error
	return nil, nil
}
