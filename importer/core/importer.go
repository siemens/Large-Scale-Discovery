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
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"strings"
)

var importers = make(map[string]Importer) // Map of importer names referencing the responsible importer module.

type Importer interface {
	Init(conf map[string]interface{}) error
	Import(filters Filters) ([]managerdb.T_discovery, error)
}

// Filters describes various possible strings that can be applied to filter the entries returned by the respective
// repository. Not all filters might apply to all repositories.
type Filters struct {
	Countries   []string
	Locations   []string
	Companies   []string
	Departments []string
	Contacts    []string

	RoutingDomains  []string
	Zones           []string
	Purposes        []string
	ExcludeKeywords []string

	Type     string
	Critical string
}

func ParseFilters(scanScopeSettings utils.JsonMap) (Filters, error) {

	// Read filters from scan scope attributes
	filterCountries, errCountries := readSliceAttribute(scanScopeSettings, "asset_countries")
	if errCountries != nil {
		return Filters{}, errCountries
	}
	filterLocations, errLocations := readSliceAttribute(scanScopeSettings, "asset_locations")
	if errLocations != nil {
		return Filters{}, errLocations
	}
	filterCompanies, errCompanies := readSliceAttribute(scanScopeSettings, "asset_companies")
	if errCompanies != nil {
		return Filters{}, errCompanies
	}
	filterDepartments, errDepartments := readSliceAttribute(scanScopeSettings, "asset_departments")
	if errDepartments != nil {
		return Filters{}, errDepartments
	}
	filterContacts, errContacts := readSliceAttribute(scanScopeSettings, "asset_contacts")
	if errContacts != nil {
		return Filters{}, errContacts
	}
	filterRoutingDomains, okRoutingDomains := readSliceAttribute(scanScopeSettings, "asset_routing_domains")
	if okRoutingDomains != nil {
		return Filters{}, okRoutingDomains
	}
	filterZones, errZones := readSliceAttribute(scanScopeSettings, "asset_zones")
	if errZones != nil {
		return Filters{}, errZones
	}
	filterPurposes, errPurposes := readSliceAttribute(scanScopeSettings, "asset_purposes")
	if errPurposes != nil {
		return Filters{}, errPurposes
	}
	filterExcludeKeywords, errExcludeKeywords := readSliceAttribute(scanScopeSettings, "asset_exclude_keywords")
	if errExcludeKeywords != nil {
		return Filters{}, errExcludeKeywords
	}

	var filterType string
	var okType bool
	val, ok := scanScopeSettings["asset_type"]
	if ok {
		filterType, okType = val.(string)
		if !okType {
			return Filters{}, fmt.Errorf("filter attribute 'asset_type' not valid")
		}
	}

	var filterCritical string
	var okCritical bool
	val, ok = scanScopeSettings["asset_critical"]
	if ok {
		filterCritical, okCritical = val.(string)
		if !okCritical {
			return Filters{}, fmt.Errorf("filter attribute 'asset_critical' not valid")
		}
	}

	// Prepare settings struct and return it
	// Convert all values to upper case for later matching, should be case in-sensitive matching
	return Filters{
		Countries:   utils.TrimToUpper(filterCountries),
		Locations:   utils.TrimToUpper(filterLocations),
		Departments: utils.TrimToUpper(filterDepartments),
		Companies:   utils.TrimToUpper(filterCompanies),
		Contacts:    utils.TrimToUpper(filterContacts),

		RoutingDomains:  utils.TrimToUpper(filterRoutingDomains),
		Zones:           utils.TrimToUpper(filterZones),
		Purposes:        utils.TrimToUpper(filterPurposes),
		ExcludeKeywords: utils.TrimToUpper(filterExcludeKeywords),

		Type:     strings.ToUpper(strings.TrimSpace(filterType)),
		Critical: strings.ToUpper(strings.TrimSpace(filterCritical)),
	}, nil
}

// Define helper function to convert interface to original slice of strings
func readSliceAttribute(attributes utils.JsonMap, key string) ([]string, error) {

	// Access attribute value
	val, _ := attributes[key]

	// Return empty slice if value is nil
	if val == nil {
		return []string{}, nil
	}

	// Cast to slice of values
	slice, okSlice := val.([]interface{})
	if !okSlice {
		return nil, fmt.Errorf("filter attribute '%s' not a slice", key)
	}

	// Prepare slice of result values
	result := make([]string, 0, len(slice))

	// Cast interface values in slice to strings
	for _, i := range slice {
		value, okString := i.(string)
		if !okString {
			return nil, fmt.Errorf("filter attribute '%s' not a slice of strings", key)
		}
		result = append(result, value)
	}

	// Return slice of strings
	return result, nil
}

// substrContained searches a value for a list of substrings, whether any of them is contained as a substring
func substrContained(value string, substrings ...[]string) bool {
	for _, slice := range substrings {
		for _, substring := range slice {
			if strings.Contains(value, substring) {
				return true
			}
		}
	}
	return false
}
