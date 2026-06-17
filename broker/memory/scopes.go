/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package memory

import (
	"github.com/orcaman/concurrent-map/v2"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
)

var scanScopes = cmap.New[*managerdb.T_scan_scope]() // Mapping of scan scope IDs (as string) with scan scope structs (containing database details, scan settings, ...). The map key must be something never-changing to avoid race conditions on change.

// AddScope adds a scan scope to memory, using its secret as an identifier. Scan scope is overwritten, if same
// secret was already used as a key. Usually the T_scan_scope struct returned by the manager should not contain the
// scope secret. So it needs to be passed as a dedicated argument.
func AddScope(scopeSecret string, scope managerdb.T_scan_scope) {
	scanScopes.Set(scopeSecret, &scope)
}

// RemoveScope removes a certain scan scope from memory, if existing
func RemoveScope(scopeSecret string) {
	scanScopes.Remove(scopeSecret)
}

// GetScope returns a copy of a scan scope as currently stored in memory and a flag whether the requested scan
// scope existed
func GetScope(scopeSecret string) (managerdb.T_scan_scope, bool) {
	scanScope, exists := scanScopes.Get(scopeSecret)
	if exists {
		return *scanScope, exists
	} else {
		return managerdb.T_scan_scope{}, exists
	}
}

// GetScopes returns a copied list of all stored scan scopes
func GetScopes() map[uint64]managerdb.T_scan_scope {

	// Grab cached scope items
	cachedScopeItems := scanScopes.Items()

	// Prepare memory for copy
	cachedScopes := make(
		map[uint64]managerdb.T_scan_scope,
		len(cachedScopeItems),
	)

	// Iterate cached scope items and copy data over
	for _, item := range cachedScopeItems {
		scanScope := *item
		cachedScopes[scanScope.Id] = scanScope
	}

	// Return copied map of scan scopes
	return cachedScopes
}

// ClearScopes removes all currently stored scan scopes from memory (active ones might be immediately reloaded by
// some other goroutine when required)
func ClearScopes() {
	for key := range scanScopes.Items() {
		scanScopes.Remove(key)
	}
}
