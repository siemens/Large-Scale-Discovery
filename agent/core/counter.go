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
	"github.com/orcaman/concurrent-map/v2"
)

var moduleLabels []string
var scopeCounter map[string]moduleCount // Map of scan scopes referencing ConcurrentMaps of module counts

type moduleCount struct { // ConcurrentMap of module to count
	cmap.ConcurrentMap[string, int]
}

func newModuleCount(labels ...string) moduleCount {

	// Initialize module counter
	mC := moduleCount{
		cmap.New[int](),
	}

	// Populate module counter and set counts to 0
	for _, label := range labels {
		mC.Set(label, 0)
	}

	// Return module counter
	return mC
}

func (mC *moduleCount) increment(label string) error {

	// Get interface to map value
	count, ok := mC.Get(label)

	// Abort if module name is not existing
	if !ok {
		return fmt.Errorf("unknwon module")
	}

	// Cast and increase the counter
	mC.Set(label, count+1)

	// Return nil as everything went fine
	return nil
}

func (mC *moduleCount) decrement(label string) {

	// Get interface to map value
	count, ok := mC.Get(label)

	// Abort if module name is not existing
	if !ok {
		return
	}

	// Cast and decrease the counter
	if count > 0 {
		mC.Set(label, count-1)
	}
}

// InitScopeCounters initializes module counters for every scan scope and scan module
func InitScopeCounters(scopeSecrets []string, labels []string) {

	// Remember initialized module labels
	moduleLabels = labels

	// Initialize module counts
	scopeCounter = make(map[string]moduleCount, len(scopeSecrets))
	for _, scopeSecret := range scopeSecrets {
		scopeCounter[scopeSecret] = newModuleCount(labels...)
	}
}

func IncrementModuleCount(scopeSecret string, label string) error {

	// Get scope's module counter
	mC, ok := scopeCounter[scopeSecret]

	// Abort if module name is not existing
	if !ok {
		return fmt.Errorf("unknwon scope counter")
	}

	// Increment counter and return result
	return mC.increment(label)
}

func DecrementModuleCount(scopeSecret string, label string) {

	// Get scope's module counter
	mC, ok := scopeCounter[scopeSecret]

	// Abort if module name is not existing
	if !ok {
		return
	}

	// Decrement counter
	mC.decrement(label)
}

// GetModuleLabels returns a slice of module labels initialized
func GetModuleLabels() []string {
	return moduleLabels
}

// GetTotalInstanceCounts returns a map of modules with their module count across all scan scopes.
func GetTotalInstanceCounts() map[string]int {

	// Prepare result map
	result := make(map[string]int, len(moduleLabels))

	// Iterate scan scopes, read module counts and populate result map
	for _, mC := range scopeCounter {
		for module := range mC.IterBuffered() {
			result[module.Key] = module.Val
		}
	}

	// Return result map
	return result
}

// GetScopeInstanceCounts returns a map of modules with their module count for a given scan scope.
func GetScopeInstanceCounts(scopeSecret string) map[string]int {

	// Get module count ConcurrentMap if existing
	mC, ok := scopeCounter[scopeSecret]
	if !ok {
		return nil
	}

	// Prepare result map
	result := make(map[string]int, len(moduleLabels))

	// Iterate scope modules, read module counts and populate result map
	for module := range mC.IterBuffered() {
		result[module.Key] = module.Val
	}

	// Return result map
	return result
}
