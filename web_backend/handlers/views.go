/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"regexp"
)

// View is a helper struct holding the values to be returned by this handler
type View struct {
	managerdb.T_scope_view

	// Expand with additional information useful for the web frontend
	Grants    []Grant `json:"grants"`     // The current list of tokens with access.
	ScanScope Scope   `json:"scan_scope"` // The view's scope details including scan status. Can be omitted if user should not see them.
}

// Views returns views owned by the current user (all in case of admin)
var Views = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Views []View `json:"views"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Prepare memory for list of views
		var scopeViews []managerdb.T_scope_view
		var errScopeViews error

		// Query groups depending on whether user is admin or not
		if contextUser.Admin {

			// Request all scope view from manager
			scopeViews, errScopeViews = manager.RpcGetViews(logger, core.RpcClient())

		} else {

			// Get current user's ownerships
			groups := make([]uint64, 0, 3)
			for _, ownership := range contextUser.Ownerships {
				groups = append(groups, ownership.Group.Id)
			}

			// Request owned scope view from manager
			scopeViews, errScopeViews = manager.RpcGetViewsOf(logger, core.RpcClient(), groups)
		}

		// Check for errors occurred while querying groups
		if errors.Is(errScopeViews, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScopeViews != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Lookup users, required to enrich user's companies
		userEntries, errUserEntries := database.GetUsers()
		if errUserEntries != nil {
			logger.Errorf("Could not query users: %s", errUserEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Query groups, required to enrich a view's group name
		groupEntries, errGroupEntries := database.GetGroups()
		if errGroupEntries != nil {
			logger.Errorf("Could not query groups: %s", errGroupEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Translate list into map for efficient lookups
		groupsDict := make(map[uint64]database.T_group, len(groupEntries))
		for _, group := range groupEntries {
			groupsDict[group.Id] = group
		}

		// Translate list into map for efficient lookups
		usersDict := make(map[string]database.T_user, len(userEntries))
		for _, user := range userEntries {
			usersDict[user.Email] = user
		}

		// Prepare customized list of views
		var views = make([]View, 0, 3) // Initialize empty slice to avoid returning nil to frontend
		for _, scopeView := range scopeViews {

			// Check whether group exists. Scope are stored on the manager, while groups are stored locally. There
			// should not be a discrepancy.
			group, ok := groupsDict[scopeView.ScanScope.IdTGroup]
			if !ok {
				logger.Warningf(
					"An unknown group ('%d') is set as the owner of view '%s' on scan scope '%s' ('%s').",
					scopeView.ScanScope.IdTGroup,
					scopeView.Name,
					scopeView.ScanScope.Name,
					scopeView.ScanScope.DbName,
				)
			}

			// Prepare list of emails of granted users
			var grants = make([]Grant, 0, 3)
			for _, grantEntry := range scopeView.Grants {

				// Prepare memory for whether grant entry is also referencing a scan scope owner
				owner := false

				// Ignore potentially existing access token
				if grantEntry.IsUser {

					// Check whether user exists. Scopes are stored on the manager, while groups are stored locally.
					// There should not be a discrepancy.
					userEntry, errUserEntry := database.GetUserByMail(grantEntry.Username)
					if errUserEntry != nil {
						logger.Errorf("Could not query user: %s", errUserEntry)
						core.RespondInternalError(ctx) // Return generic error information
						return
					}

					// Warn about corrupted data
					if userEntry == nil {
						logger.Warningf(
							"An unknown user ('%s') has access to view '%s' on scan scope '%s' ('%s').",
							grantEntry.Username,
							scopeView.Name,
							scopeView.ScanScope.Name,
							scopeView.ScanScope.DbName,
						)
					}

					// Check if granted user is scope owner too
					for _, owners := range group.Ownerships {
						if owners.User.Email == grantEntry.Username {
							owner = true
							break
						}
					}
				}

				// Add to list of granted token
				grants = append(grants, Grant{
					T_view_grant:   grantEntry,
					UserCreated:    usersDict[grantEntry.Username].Created,
					UserLastLogin:  usersDict[grantEntry.Username].LastLogin,
					UserCompany:    usersDict[grantEntry.Username].Company,
					UserDepartment: usersDict[grantEntry.Username].Department,
					UserIsAdmin:    usersDict[grantEntry.Username].Admin,
					UserIsOwner:    owner,
				})
			}

			// Remove list of targets as it might be big causing delays and is not required here
			delete(scopeView.ScanScope.Attributes, "targets")

			// Create and append scope view to response
			views = append(views, View{
				T_scope_view: scopeView,
				Grants:       grants,
				ScanScope: Scope{
					T_scan_scope: scopeView.ScanScope,
					GroupName:    group.Name,

					// Connection: currently not required when requesting views for configuration
					// ScanSettings: Don't return scan settings attributes. Sensitive!
				},
			})
		}

		// Prepare response body
		body := responseBody{
			Views: views,
		}

		// Return response
		core.Respond(ctx, false, "Views retrieved.", body)
	}
}

// ViewsGranted returns a list of views the current user has access rights to
var ViewsGranted = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Views []View `json:"views"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Request views granted from manager
		scopeViews, errScopeViews := manager.RpcGetViewsGranted(logger, core.RpcClient(), contextUser.Email)
		if errors.Is(errScopeViews, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScopeViews != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Query groups, required to enrich a view's group name
		groupEntries, errGroupEntries := database.GetGroups()
		if errGroupEntries != nil {
			logger.Errorf("Could not query groups: %s", errGroupEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Translate list into map for efficient lookups
		groupsDict := make(map[uint64]database.T_group, len(groupEntries))
		for _, group := range groupEntries {
			groupsDict[group.Id] = group
		}

		// Prepare customized list of views
		var views = make([]View, 0, 3) // Initialize empty slice to avoid returning nil to frontend
		for _, scopeView := range scopeViews {

			// Check whether group exists. Scope are stored on the manager, while groups are stored locally. There
			// should not be a discrepancy.
			group, ok := groupsDict[scopeView.ScanScope.IdTGroup]
			if !ok {
				logger.Warningf(
					"An unknown group ('%d') is set as the owner of view '%s' on scan scope '%s' ('%s').",
					scopeView.ScanScope.IdTGroup,
					scopeView.Name,
					scopeView.ScanScope.Name,
					scopeView.ScanScope.DbName,
				)
				continue // Misconfiguration, log critical error and skip from display
			}

			// Remove list of targets as it might be big causing delays and is not required here
			delete(scopeView.ScanScope.Attributes, "targets")

			// Create and append scope view to response
			views = append(views, View{
				T_scope_view: scopeView,
				Grants:       nil, // Currently not required when requesting ones own (granted) views
				ScanScope: Scope{
					T_scan_scope: scopeView.ScanScope,
					GroupName:    group.Name,
					Connection: Connection{
						Host:     scopeView.ScanScope.DbServer.HostPublic,
						Port:     scopeView.ScanScope.DbServer.Port,
						Database: scopeView.ScanScope.DbName,
					},

					// ScanSettings: Don't return scan settings attributes. Sensitive!
				},
			})
		}

		// Prepare response body
		body := responseBody{
			Views: views,
		}

		// Return response
		core.Respond(ctx, false, "Views retrieved.", body)
	}
}

// ViewCreate creates a view on a scope, if the user has ownership rights on the related scope (or is admin)
var ViewCreate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		ScopeId                   uint64    `json:"scope_id"`
		ViewName                  string    `json:"view_name"`
		FilterInputTargets        *[]string `json:"filter_input_targets"`
		FilterInputCountries      *[]string `json:"filter_input_countries"`
		FilterInputLocations      *[]string `json:"filter_input_locations"`
		FilterInputRoutingDomains *[]string `json:"filter_input_routing_domains"`
		FilterInputZones          *[]string `json:"filter_input_zones"`
		FilterInputPurposes       *[]string `json:"filter_input_purposes"`
		FilterInputCompanies      *[]string `json:"filter_input_companies"`
		FilterInputDepartments    *[]string `json:"filter_input_departments"`
		FilterInputManagers       *[]string `json:"filter_input_managers"`
		FilterInputContacts       *[]string `json:"filter_input_contacts"`
		FilterInputComments       *[]string `json:"filter_input_comments"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.ScopeId)
		if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScope != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check ID to make sure it existed in the DB
		if scanScope.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Check if user has rights to create view
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Check if view name is already existing on same scope
		for _, scopeView := range scanScope.ScopeViews {
			if scopeView.Name == req.ViewName {
				core.Respond(ctx, true, "Duplicate view name on scan scope.", responseBody{})
				return
			}
		}

		// Get group to create scope for
		groupEntry, errGroupEntry := database.GetGroup(scanScope.IdTGroup)
		if errGroupEntry != nil {
			logger.Errorf("Could not query group: %s", errGroupEntry)
			core.RespondInternalError(ctx)
			return
		}

		// Check if group exists
		if groupEntry == nil {
			logger.Errorf(
				"An unknown group ('%d') is set as the owner of scan scope '%s' ('%s').",
				scanScope.IdTGroup,
				scanScope.Name,
				scanScope.DbName,
			)
			core.RespondInternalError(ctx)
			return
		}

		// Request owned scan scopes from manager
		scopes, errScopes := manager.RpcGetScopesOf(logger, core.RpcClient(), []uint64{groupEntry.Id})
		if errScopes != nil {
			logger.Errorf("Could not query scopes of group '%d': %s", groupEntry.Id, errScopes)
			core.RespondTemporaryError(ctx)
			return
		}

		// Count views on group's scopes
		cntViews := 0
		for _, scope := range scopes {
			cntViews += len(scope.ScopeViews)
		}

		// Check whether limits are exceeded
		if groupEntry.MaxViews >= 0 && cntViews >= groupEntry.MaxViews {
			core.Respond(ctx, true, "View limit reached.", responseBody{})
			return
		}

		// Compile simple regex to filter characters other than insensitive alpha numeric ones and space
		reg, errRegex := regexp.Compile("[^a-zA-Z0-9-_./ ]+")
		if errRegex != nil {
			logger.Errorf("Could not compile regex to check view filters: %s", errRegex)
			core.RespondInternalError(ctx) // Return generic error information.
			return
		}

		// Prepare sanitization function removing
		fnSanitize := func(strings *[]string) ([]string, error) {

			// Skip sanitization if slice is not defined
			if strings == nil {
				return nil, nil
			}

			// Iterate values and generate sanitized slice of strings
			var stringsSanitized = make([]string, 0, len(*strings))
			for _, str := range *strings {

				// Return with error if invalid character is detected
				match := reg.FindString(str) // Search for invalid characters
				if len(match) > 0 {
					return nil, fmt.Errorf("%s", str)
				}

				// Add sanitized value
				if len(str) > 0 && str != "*" && str != "-" {
					stringsSanitized = append(stringsSanitized, str)
				}
			}

			// Return sanitized slice of strings
			return stringsSanitized, nil
		}

		// Prepare filter definition. Each key is a scopedb column name and a slice of valid values
		var filters = make(map[string][]string, 12)
		var values []string
		var errSanitize error
		if req.FilterInputTargets != nil {
			values, errSanitize = fnSanitize(req.FilterInputTargets)
			if errSanitize == nil && len(values) > 0 {
				filters["input"] = values
			}
		}
		if errSanitize == nil && req.FilterInputCountries != nil {
			values, errSanitize = fnSanitize(req.FilterInputCountries)
			if errSanitize == nil && len(values) > 0 {
				filters["input_country"] = values
			}
		}
		if errSanitize == nil && req.FilterInputLocations != nil {
			values, errSanitize = fnSanitize(req.FilterInputLocations)
			if errSanitize == nil && len(values) > 0 {
				filters["input_location"] = values
			}
		}
		if errSanitize == nil && req.FilterInputRoutingDomains != nil {
			values, errSanitize = fnSanitize(req.FilterInputRoutingDomains)
			if errSanitize == nil && len(values) > 0 {
				filters["input_routing_domain"] = values
			}
		}
		if errSanitize == nil && req.FilterInputZones != nil {
			values, errSanitize = fnSanitize(req.FilterInputZones)
			if errSanitize == nil && len(values) > 0 {
				filters["input_zone"] = values
			}
		}
		if errSanitize == nil && req.FilterInputPurposes != nil {
			values, errSanitize = fnSanitize(req.FilterInputPurposes)
			if errSanitize == nil && len(values) > 0 {
				filters["input_purpose"] = values
			}
		}
		if errSanitize == nil && req.FilterInputCompanies != nil {
			values, errSanitize = fnSanitize(req.FilterInputCompanies)
			if errSanitize == nil && len(values) > 0 {
				filters["input_company"] = values
			}
		}
		if errSanitize == nil && req.FilterInputDepartments != nil {
			values, errSanitize = fnSanitize(req.FilterInputDepartments)
			if errSanitize == nil && len(values) > 0 {
				filters["input_department"] = values
			}
		}
		if errSanitize == nil && req.FilterInputManagers != nil {
			values, errSanitize = fnSanitize(req.FilterInputManagers)
			if errSanitize == nil && len(values) > 0 {
				filters["input_manager"] = values
			}
		}
		if errSanitize == nil && req.FilterInputContacts != nil {
			values, errSanitize = fnSanitize(req.FilterInputContacts)
			if errSanitize == nil && len(values) > 0 {
				filters["input_contact"] = values
			}
		}
		if errSanitize == nil && req.FilterInputComments != nil {
			values, errSanitize = fnSanitize(req.FilterInputComments)
			if errSanitize == nil && len(values) > 0 {
				filters["input_comment"] = values
			}
		}

		// Abort with error message
		if errSanitize != nil {
			core.Respond(ctx, true, fmt.Sprintf("Illegal filter value '%s'.", errSanitize), responseBody{})
			return
		}

		// Request view creation from manager
		errRpc := manager.RpcCreateView(logger, core.RpcClient(), scanScope.Id, req.ViewName, contextUser.Email, filters)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "View created.", responseBody{})
	}
}

// ViewDelete deletes a certain view on a scope, if the user has ownership rights on the related scope (or is admin)
var ViewDelete = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		ViewId uint64 `json:"id"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Request scope view from manager
		scopeView, errScopeView := manager.RpcGetView(logger, core.RpcClient(), req.ViewId)
		if errors.Is(errScopeView, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScopeView != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check ID to make sure it existed in the DB
		if scopeView.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Check if user has rights to delete view
		if !core.OwnerOrAdmin(scopeView.ScanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request view deletion from manager
		errRpc := manager.RpcDeleteView(logger, core.RpcClient(), scopeView.Id)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "View deleted.", responseBody{})
	}
}

// ViewUpdate updates view details, if the user has ownership rights (or is admin)
var ViewUpdate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id   uint64  `json:"id"` // PK of the DB element to identify associated entry for update
		Name *string `json:"name"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if primary key is defined, otherwise gorm cannot update specific item
		if req.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		/// Request scope view from manager
		scopeView, errScopeView := manager.RpcGetView(logger, core.RpcClient(), req.Id)
		if errors.Is(errScopeView, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScopeView != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check ID to make sure it existed in the DB
		if scopeView.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scopeView.ScanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Decide new or old name
		name := scopeView.Name
		if req.Name != nil {
			name = *req.Name
		}

		// Execute update on manager via RPC
		errRpc := manager.RpcUpdateView(logger, core.RpcClient(), req.Id, name)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "View updated.", responseBody{})
	}
}
