/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
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

// scopeEntries stores scan-scope state in memory
var scopeEntries = cmap.New[*managerdb.T_scan_scope]()

// AddScope adds a scan scope to memory, using its secret as an identifier. Scan scope is overwritten, if same
// secret was already used as a key. Usually the T_scan_scope struct returned by the manager should not contain the
// scope secret. So it needs to be passed as a dedicated argument.
func AddScope(scopeSecret string, scopeEntry managerdb.T_scan_scope) {
	scopeEntries.Set(scopeSecret, &scopeEntry)
}

// RemoveScope removes a certain scan scope from memory, if existing
func RemoveScope(scopeSecret string) {
	scopeEntries.Remove(scopeSecret)
}

// GetScope returns a copy of a scan scope as currently stored in memory and a flag whether the requested scan
// scope existed
func GetScope(scopeSecret string) (managerdb.T_scan_scope, bool) {
	scopeEntry, exists := scopeEntries.Get(scopeSecret)
	if exists {
		return *scopeEntry, exists
	} else {
		return managerdb.T_scan_scope{}, exists
	}
}

// GetScopes returns a copied list of all stored scan scopes
func GetScopes() map[uint64]managerdb.T_scan_scope {

	// Grab cached scope items
	scopeEntriesCached := scopeEntries.Items()

	// Prepare memory for copy
	scopeEntriesCopied := make(
		map[uint64]managerdb.T_scan_scope,
		len(scopeEntriesCached),
	)

	// Iterate cached scope items and copy data over
	for _, scopeEntryCached := range scopeEntriesCached {
		scopeEntry := *scopeEntryCached // Copy
		scopeEntriesCopied[scopeEntry.Id] = scopeEntry
	}

	// Return copied map of scan scopes
	return scopeEntriesCopied
}

// ClearScopes removes all currently stored scan scopes from memory (active ones might be immediately reloaded by
// some other goroutine when required)
func ClearScopes() {
	for key := range scopeEntries.Items() {
		scopeEntries.Remove(key)
	}
}
