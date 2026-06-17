/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package handlers

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/utils/nmapargs"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/ZapSmtp/smtp"
)

type Connection struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
}

// Scope is a helper struct holding the values to be returned by this handler
type Scope struct {
	managerdb.T_scan_scope

	// Expand with additional information useful for the web frontend
	GroupName    string                   `json:"group_name"`    // Additional information for the web frontend
	Connection   Connection               `json:"connection"`    // The scope's current settings. Can be omitted if user should not see them.
	ScanSettings managerdb.T_scan_setting `json:"scan_settings"` // The scope's current settings. Can be omitted if user should not see them.
	ScanAgents   []managerdb.T_scan_agent `json:"scan_agents"`   // The scope's last seen agents. Can be omitted if user should not see them.
}

// ScopesResponse defines the expected response structure for the Scopes handler
type ScopesResponse struct {
	Scopes       []Scope `json:"scopes"`        // List of scan scopes owned by the current user
	AllowCustom  bool    `json:"allow_custom"`  // Whether the user is allowed to create custom scan scopes
	AllowNetwork bool    `json:"allow_network"` // Whether the user is allowed to create network scan scopes
	AllowAsset   bool    `json:"allow_asset"`   // Whether the user is allowed to create asset scan scopes
}

// Scopes returns scopes owned by the current user (all in case of admin)
// @Summary      List scopes
// @Description  Returns all scan scopes owned by the current user.
// @Tags         scopes
// @Produce      json
// @Success      200  {object}  ScopesResponse  "Scopes retrieved"
// @Failure      401  "Unauthorized"
// @Failure      503  "Service Unavailable"
// @Security     BearerAuth
// @Router       /scopes [get]
func Scopes() gin.HandlerFunc {

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Prepare memory for list of scopes
		var scanScopes []managerdb.T_scan_scope
		var errScanScopes error

		// Prepare memory for user rights for scope creation
		var allowCustom bool
		var allowNetwork bool
		var allowAsset bool

		// Query groups, depending on whether user is admin or not
		if contextUser.Admin {

			// Request all scan scopes from manager
			scanScopes, errScanScopes = manager.RpcGetScopes(logger, core.RpcClient())

			// Set user rights for scope creation
			allowCustom = true
			allowNetwork = true
			allowAsset = true

		} else {

			// Get current user's ownerships
			groups := make([]uint64, 0, 3)
			for _, ownership := range contextUser.Ownerships {
				groups = append(groups, ownership.Group.Id)

				// Check if user is allowed to create scan scopes of certain kinds
				if ownership.Group.AllowCustom {
					allowCustom = true
				}
				if ownership.Group.AllowNetwork {
					allowNetwork = true
				}
				if ownership.Group.AllowAsset {
					allowAsset = true
				}
			}

			// Request owned scan scopes from manager
			scanScopes, errScanScopes = manager.RpcGetScopesOf(logger, core.RpcClient(), groups)

		}

		// Check for errors occurred while querying groups
		if errors.Is(errScanScopes, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScopes != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Query groups, required to enrich a scope's group name
		groupEntries, errGroupEntries := database.GetGroups()
		if errGroupEntries != nil {
			logger.Warningf("Could not query groups: %s", errGroupEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Translate list into map for efficient lookups
		groupsDict := make(map[uint64]database.T_group, len(groupEntries))
		for _, group := range groupEntries {
			groupsDict[group.Id] = group
		}

		// Prepare customized list of scan scopes
		var scopes = make([]Scope, 0, 3) // Initialize empty slice to avoid returning nil to frontend
		for _, scanScope := range scanScopes {

			// Check whether group exists. Scope are stored on the manager, while groups are stored locally. There
			// should not be a discrepancy.
			group, ok := groupsDict[scanScope.IdTGroup]
			if !ok {
				logger.Warningf(
					"An unknown group ('%d') is set as the owner of scan scope '%s' ('%s').",
					scanScope.IdTGroup,
					scanScope.Name,
					scanScope.DbName,
				)
			}

			// Create and append scan scope to response
			scopes = append(scopes, Scope{
				T_scan_scope: scanScope,
				GroupName:    group.Name,
				Connection: Connection{
					Host:     scanScope.DbServer.HostPublic,
					Port:     scanScope.DbServer.Port,
					Database: scanScope.DbName,
				},

				ScanSettings: scanScope.ScanSettings,
				ScanAgents:   scanScope.ScanAgents,
			})
		}

		// Prepare response body
		body := ScopesResponse{
			Scopes:       scopes,
			AllowCustom:  allowCustom,
			AllowNetwork: allowNetwork,
			AllowAsset:   allowAsset,
		}

		// Return response
		core.Respond(ctx, false, "Scopes retrieved.", body)
	}
}

// ScopeDelete deletes a certain scope, if the user has ownership rights (or is admin)
var ScopeDelete = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id uint64 `json:"id"` // PK of the DB element to identify associated entry for update
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

		// Check if primary key is defined, otherwise gorm cannot update specific entry
		if req.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
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

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to delete scan scope and associated data
		errRpc := manager.RpcDeleteScope(logger, core.RpcClient(), req.Id)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Scope deleted.", responseBody{})
	}
}

// ScopeTargetsRequest defines the expected request structure for the ScopeTargets handler
type ScopeTargetsRequest struct {
	Id uint64 `json:"id" example:"42"` // Scope ID to get the targets for
}

// ScopeTargetsResponse defines the expected response structure for the ScopeTargets handler
type ScopeTargetsResponse struct {
	Synchronization bool                    `json:"synchronization"` // Flag indicating whether previous synchronization is still ongoing (no targets in that case)
	Targets         []managerdb.T_discovery `json:"targets"`         // Only returned if no synchronization currently ongoing
}

// ScopeTargets returns a scope's inputs if owned by the current user (all in case of admin)
// @Summary      Get scope targets
// @Description  Returns the list of scan targets (inputs) for a specific scope. Requires ownership.
// @Tags         scopes
// @Accept       json
// @Produce      json
// @Param        request body ScopeTargetsRequest true "Scope targets request"
// @Success      200  {object}  ScopeTargetsResponse  "Scope targets retrieved"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      503  "Service Unavailable"
// @Security     BearerAuth
// @Router       /scope/targets [post]
func ScopeTargets() gin.HandlerFunc {

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Declare expected request struct
		var req ScopeTargetsRequest

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if primary key is defined, otherwise gorm cannot update specific entry
		if req.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", ScopeTargetsResponse{})
			return
		}

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
		if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScope != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check ID to make sure it existed in the DB
		if scanScope.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", ScopeTargetsResponse{})
			return
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request scope details from manager
		synchronization, targets, errScanScopeTargets := manager.RpcGetScopeTargets(logger, core.RpcClient(), req.Id)
		if errors.Is(errScanScopeTargets, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScopeTargets != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Prepare response body
		body := ScopeTargetsResponse{
			Synchronization: synchronization,
			Targets:         targets,
		}

		// Return response
		core.Respond(ctx, false, "Scope targets retrieved.", body)
	}
}

// ScopeResetFailed resets the scan status of failed scan inputs in order to trigger a rescan within the current scan cycle
var ScopeResetFailed = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id uint64 `json:"id"` // PK of the DB element to identify associated entry for update
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

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
		if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScope != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to reset failed scan targets
		errRpc := manager.RpcResetFailed(
			logger,
			core.RpcClient(),
			scanScope.Id,
		)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Reset failed scan targets.", responseBody{})
	}
}

// ScopeNewCycle initializes a new scan cycle causing all scan targets to be scanned again. Results of ongoing scans
// will be dropped, but existing ones will remain.
var ScopeNewCycle = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id uint64 `json:"id"` // PK of the DB element to identify associated entry for update
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

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
		if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScope != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to increment scan cycle
		errRpc := manager.RpcNewCycle(
			logger,
			core.RpcClient(),
			scanScope.Id,
		)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "New scan cycle initialized.", responseBody{})
	}
}

// ScopeTogglePause enabled/disables a scan scope. Disabled (paused) scan scopes are not processed by the broker. Scan
// agents will be able to complete running scan tasks, but not receive new ones.
var ScopeTogglePause = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id uint64 `json:"id"` // PK of the DB element to identify associated entry for update
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

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
		if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScope != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to update scan scope
		errRpc := manager.RpcToggleScope(
			logger,
			core.RpcClient(),
			scanScope.Id,
		)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		if scanScope.Enabled {
			core.Respond(ctx, false, "Scope paused.", responseBody{})
		} else {
			core.Respond(ctx, false, "Scope resumed.", responseBody{})
		}
	}
}

// ScopeUpdateSettings updates scan settings of a certain scope, if the user has ownership rights (or is admin)
var ScopeUpdateSettings = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id           uint64                   `json:"id"` // PK of the DB element to identify associated entry for update
		ScanSettings managerdb.T_scan_setting `json:"scan_settings"`
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

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
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

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Validate nmap pre-scan args if changed
		if req.ScanSettings.DiscoveryNmapArgsPrescan != scanScope.ScanSettings.DiscoveryNmapArgsPrescan {

			// Validate Nmap pre-scan arguments with dangerous-flag blocking.
			validationResult := nmapargs.Validate(
				req.ScanSettings.DiscoveryNmapArgsPrescan,
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsInput),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsDatabase),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsOutput),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsSpoof),
			)

			// Check if validation was successful
			if !validationResult.Valid {
				logger.Warningf(
					"Invalid Nmap pre-scan args '%s': %s",
					req.ScanSettings.DiscoveryNmapArgsPrescan,
					strings.Join(validationResult.Errors, ", "),
				)
				core.Respond(ctx, true, fmt.Sprintf("Invalid Nmap pre-scan args: %s", validationResult.Errors[0]), responseBody{})
				return
			}
		}

		// Validate nmap args if changed
		if req.ScanSettings.DiscoveryNmapArgs != scanScope.ScanSettings.DiscoveryNmapArgs {

			// Validate Nmap scan arguments with dangerous-flag blocking.
			validationResult := nmapargs.Validate(
				req.ScanSettings.DiscoveryNmapArgs,
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsInput),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsDatabase),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsOutput),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsScript),
				nmapargs.WithBlockFlags(nmapargs.BlockFlagsSpoof),
			)

			// Check if validation was successful
			if !validationResult.Valid {
				logger.Warningf(
					"Invalid Nmap args '%s': %s",
					req.ScanSettings.DiscoveryNmapArgs,
					strings.Join(validationResult.Errors, ", "),
				)
				core.Respond(ctx, true, fmt.Sprintf("Invalid Nmap args: %s", validationResult.Errors[0]), responseBody{})
				return
			}
		}

		// Request manager to update scan settings
		errRpc := manager.RpcUpdateSettings(logger, core.RpcClient(), req.Id, req.ScanSettings)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Scope settings updated.", responseBody{})
	}
}

// ScopeCreateUpdateCustomRequest defines the expected request structure for the ScopeCreateUpdateCustom handler
type ScopeCreateUpdateCustomRequest struct {
	ExistingScopeId *uint64                  `json:"scope_id" example:"42"`        // Set if EXISTING scope shall be updated
	GroupId         *uint64                  `json:"group_id" example:"1"`         // Set if NEW scope shall be created
	Name            string                   `json:"name" example:"My Scope"`      // Name of the scan scope
	Ot              bool                     `json:"ot"`                           // Whether this is an OT discovery scope
	Cycles          bool                     `json:"cycles"`                       // Whether to enable cyclic scanning
	CyclesRetention int                      `json:"cycles_retention" example:"3"` // Amount of scan cycles to keep. -1 to keep all. 0 not allowed.
	Targets         *[]managerdb.T_discovery `json:"targets"`                      // List of scan targets to set
}

// ScopeCreateUpdateCustomResponse defines the expected response structure for the ScopeCreateUpdateCustom handler
type ScopeCreateUpdateCustomResponse struct {
	Warnings []string `json:"warnings"` // List of warnings generated during target processing
}

// ScopeCreateUpdateCustom creates or updates a scan scope configuration. If a group ID is supplied, a new scan scope will
// be created. If a scope ID is provided, an update will be performed. Only executes, if the user has ownership rights
// @Summary      Create or update custom scope
// @Description  Creates a new custom scan scope or updates an existing one with scan targets. Supply group_id for creation or scope_id for update. Requires ownership.
// @Tags         scopes
// @Accept       json
// @Produce      json
// @Param        request body ScopeCreateUpdateCustomRequest true "Custom scope request"
// @Success      200  {object}  ScopeCreateUpdateCustomResponse  "Scope created or updated"
// @Failure      400  "Bad Request"
// @Failure      401  "Unauthorized"
// @Failure      503  "Service Unavailable"
// @Security     BearerAuth
// @Router       /scope/update/custom [post]
func ScopeCreateUpdateCustom() gin.HandlerFunc {

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Declare expected request struct
		var req ScopeCreateUpdateCustomRequest

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Abort request because it is invalid
		if (req.ExistingScopeId == nil && req.GroupId == nil) || (req.ExistingScopeId != nil && req.GroupId != nil) {
			logger.Debugf("Scope ID to update or group ID to create scope required.")
			core.Respond(ctx, true, "Scope ID or group ID required.", ScopeCreateUpdateCustomResponse{})
			return
		}

		// Check if scope name is defined
		if len(req.Name) == 0 {
			core.Respond(ctx, true, "Invalid scope name.", ScopeCreateUpdateCustomResponse{})
			return
		}

		// Don't allow 0 retention cycles, but change it to -1 (keep all). In case of a bug, cycle retention would
		// be unintentionally zero, causing all scan results (outside the current scan cycle) to be wiped.
		if req.CyclesRetention <= 0 {
			req.CyclesRetention = -1
		}

		// Decide whether to create or update scope
		createScope := req.ExistingScopeId == nil || *req.ExistingScopeId <= 0

		// Prepare some memory variables
		var scopeId uint64               // To be filled later, in update mode
		var scopeGroup *database.T_group // To be set later
		var respMsg string

		// Prepare creation of new scan scope
		if createScope {

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(*req.GroupId)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", ScopeCreateUpdateCustomResponse{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		} else { // Prepare update of existing scan scope

			// Request scope details from manager
			scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), *req.ExistingScopeId)
			if errors.Is(errScanScope, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errScanScope != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Check ID to make sure it existed in the DB
			if scanScope.Id == 0 {
				core.Respond(ctx, true, "Invalid ID.", ScopeCreateUpdateCustomResponse{})
				return
			}

			// Remember existing scope ID
			scopeId = scanScope.Id

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(scanScope.IdTGroup)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", ScopeCreateUpdateCustomResponse{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scopeGroup.Id, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Check if user is allowed to create scan scopes of this kind
		if !contextUser.Admin && !scopeGroup.AllowCustom {
			core.RespondAuthError(ctx)
			return
		}

		// Request owned scan scopes from manager
		groupScopes, errScopes := manager.RpcGetScopesOf(logger, core.RpcClient(), []uint64{scopeGroup.Id})
		if errScopes != nil {
			logger.Warningf("Could not query scopes of group %d: %s", scopeGroup.Id, errScopes)
			core.RespondTemporaryError(ctx)
			return
		}

		// Create scope
		if createScope {

			// Check if new scope can be created
			if scopeGroup.MaxScopes >= 0 && len(groupScopes) >= scopeGroup.MaxScopes {
				core.Respond(ctx, true, "Scope limit reached.", ScopeCreateUpdateCustomResponse{})
				return
			}

			// Chose database server by ID to use for the new scan scope
			dbServerId := scopeGroup.DbServerId // Use other database server if assigned for scope group
			if dbServerId <= 0 {
				dbServerId = 1 // Use default database server is no other is specified
			}

			// Execute create on manager via RPC
			createId, errCreate := manager.RpcCreateScope(
				logger,
				core.RpcClient(),
				dbServerId,
				req.Name,
				scopeGroup.Id,
				contextUser.Email,
				req.Ot,
				req.Cycles,
				req.CyclesRetention,
				"custom",
				nil,
			)
			if errCreate != nil {
				logger.Warningf("Could not create scan scope: %s", errCreate)
				core.RespondTemporaryError(ctx) // Return generic error information
				return
			}

			// Log event
			errEvent := database.NewEvent(
				contextUser,
				database.EventScopeCreate,
				fmt.Sprintf("Scope: %s", req.Name),
			)
			if errEvent != nil {
				logger.Errorf("Could not create event log: %s", errEvent)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Remember created scope ID
			scopeId = createId

			// Set response message
			respMsg = "Scope created."
		}

		// Prepare working memory
		var warnLarge = uint(0)
		var warnSplits = uint(0)

		// Validate and sanitize scope targets and synchronize them with the scopedb
		if req.Targets != nil {

			// Prepare working memory
			var totalTargets = uint(0)

			// Prepare actual list of targets to hold sanitized data
			targets := make([]managerdb.T_discovery, 0, len(*req.Targets))

			// Sanitize and count new scope targets
			for _, target := range *req.Targets {

				// OT discovery scans do not have a preset scan target. The target depend on the agent's local
				// network interfaces and is selected by the scan agent. However, a single generic target entry
				// is generated for OT discovery scans, so that users can enter target attributes, such as
				// location information, etc. The input column is back-filled in later by the broker with data
				// fed back by the agent.
				if req.Ot {
					// Input and input size should not be changed by user input and is prevented by the frontend.
					// It's not important and for simplicity reasons currently not enforced by the backend, since
					// the broker will always correct it when it receives the actual scan data from the scan agent.
					// The scope owner's scope size limits are not fully effective for this kind of scan scope
					// anyway, since the target size only becomes known with the scan results.
					targets = append(targets, target)
					break
				}

				// Check if target input is valid
				if !scanUtils.IsValidAddress(target.Input) && !scanUtils.IsValidIpRange(target.Input) {
					core.Respond(
						ctx,
						true,
						fmt.Sprintf("Invalid scope target '%s'.", target.Input),
						ScopeCreateUpdateCustomResponse{},
					)
					return
				}

				// Calculate and set input size
				count, errCount := utils.CountIpsInInput(target.Input)
				if errCount != nil {
					core.Respond(
						ctx,
						true,
						fmt.Sprintf("Invalid scope target '%s'.", target.Input),
						ScopeCreateUpdateCustomResponse{},
					)
					return
				}

				// Split into smaller targets if necessary or directly append to list of targets
				if count > utils.NetworkSizeSplit {

					// Warn on overly large inputs
					if count > utils.NetworkSizeSkip {
						warnLarge += 1
					}

					// Warn about split
					warnSplits += 1

					// Split into smaller subnets if too big for Nmap
					subnets, errSubnets := utils.SplitNetworkIpV4(target.Input, utils.NetworkSizeSplit) // Create ideal bunches of 1024 if splitting is necessary
					if errSubnets != nil {
						logger.Errorf("Could not split network '%s': %s", target.Input, errSubnets)
						core.RespondInternalError(ctx) // Return generic error information
						return
					}

					// Generate target entries for split network
					// Set correct target size
					for _, subnet := range subnets {

						t := target // Copy target struct
						t.Input = subnet

						// Count
						totalTargets += utils.NetworkSizeSplit

						// Append to targets
						targets = append(targets, t)
					}

				} else {

					// Count
					totalTargets += count

					// Append to targets
					targets = append(targets, target)
				}
			}

			// Check whether group has sufficient limits left
			if totalTargets > 0 && scopeGroup.MaxTargets >= 0 {

				// Calculate current total amount of targets configured by the group
				count := uint(0)
				for _, groupScope := range groupScopes {
					if scopeId > 0 && scopeId != groupScope.Id { // Skip current group because it's targets will be replaced
						count += groupScope.Size
					}
				}
				count += totalTargets

				// Check if group limit is exceeded
				if int(count) >= scopeGroup.MaxTargets {
					core.Respond(ctx, true, "Target limit reached.", ScopeCreateUpdateCustomResponse{})
					return
				}
			}

			// Deploy scope targets in scopedb via manager. The manager will update the scan scope targets in the
			// background and return an RPC response immediately (if blocking=false).
			// ATTENTION: Another targets update for the same scan scope will fail until the previous one is completed.
			_, errRpc := manager.RpcUpdateScopeTargets(
				logger,
				core.RpcClient(),
				scopeId,
				targets,
				false,
			)
			if errRpc != nil && errRpc.Error() == manager.ErrScopeUpdateOngoing.Error() { // Errors received from RPC lose their original type!!
				core.Respond(ctx, true, "Synchronization of scan targets still ongoing.", ScopeCreateUpdateCustomResponse{})
				return
			} else if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}
		}

		// Execute update of scan scope
		if !createScope {

			// Request manager to update scan scope
			errRpc := manager.RpcUpdateScope(
				logger,
				core.RpcClient(),
				scopeId,
				req.Name,
				req.Cycles,
				req.CyclesRetention,
				nil,
			)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Set response message
			respMsg = "Scope updated."
		}

		// Prepare warnings
		var warnings []string
		if warnSplits > 0 {
			warnings = append(warnings, fmt.Sprintf("%d large subnet(s) splitted.", warnSplits))
		}
		if warnLarge > 0 {
			warnings = append(warnings, fmt.Sprintf("%d subnet(s) were extremely large.", warnLarge))
		}

		// Prepare response body
		body := ScopeCreateUpdateCustomResponse{
			Warnings: warnings,
		}

		// Return response
		core.Respond(ctx, false, respMsg, body)
	}
}

// ScopeCreateUpdateNetworks creates or updates a scan scope configuration to be imported from a network
// inventory. If a group ID is supplied, a new scan scope will be created. If a scope ID is provided, an update will
// be performed. Only executes, if the user has ownership rights (or is admin).
// ATTENTION: This only creates a scan scope with an import definition. Scan input targets are not inserted into the
//
//	scan scope database. The importer component takes care of maintaining scan input targets according to
//	this scan scope import definition.
var ScopeCreateUpdateNetworks = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		ExistingScopeId      *uint64  `json:"scope_id"` // Set if EXISTING scope shall be updated. PK of the DB scope element to identify associated entry for update
		GroupId              *uint64  `json:"group_id"` // Set if NEW scope shall be created. PK of DB group element to create scope for
		Type                 string   `json:"type"`     // Type flag relevant for the importer to decide which repository to load from
		Name                 string   `json:"name"`
		Cycles               bool     `json:"cycles"`
		CyclesRetention      int      `json:"cycles_retention"` // >=1 for the amount of scan cycles to keep. -1 to keep all scan results. 0 is not allowed for safety reasons!
		Sync                 bool     `json:"sync"`
		AssetCompanies       []string `json:"asset_companies"`
		AssetDepartments     []string `json:"asset_departments"`
		AssetRoutingDomains  []string `json:"asset_routing_domains"`
		AssetZones           []string `json:"asset_zones"`
		AssetPurposes        []string `json:"asset_purposes"`
		AssetCountries       []string `json:"asset_countries"`
		AssetLocations       []string `json:"asset_locations"`
		AssetContacts        []string `json:"asset_contacts"`
		AssetExcludeKeywords []string `json:"asset_exclude_keywords"`
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

		// Abort request because it is invalid
		if (req.ExistingScopeId == nil && req.GroupId == nil) || (req.ExistingScopeId != nil && req.GroupId != nil) {
			logger.Debugf("Scope ID to update or group ID to create scope required.")
			core.Respond(ctx, true, "Scope ID or group ID required.", responseBody{})
			return
		}

		// Check if scope name is defined
		if len(req.Name) == 0 {
			core.Respond(ctx, true, "Invalid scope name.", responseBody{})
			return
		}

		// Don't allow 0 retention cycles, but change it to -1 (keep all). In case of a bug, cycle retention would
		// be unintentionally zero, causing all scan results (outside the current scan cycle) to be wiped.
		if req.CyclesRetention <= 0 {
			req.CyclesRetention = -1
		}

		// Prepare some memory variables
		createScope := req.ExistingScopeId == nil || *req.ExistingScopeId <= 0
		updateScope := !createScope
		scopeId := uint64(0)             // To be filled later, in update mode
		var scopeGroup *database.T_group // To be set later

		// Prepare creation of new scan scope
		if createScope {

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(*req.GroupId)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", responseBody{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		} else { // Prepare update of existing scan scope

			// Request scope details from manager
			scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), *req.ExistingScopeId)
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

			// Remember existing scope ID
			scopeId = scanScope.Id

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(scanScope.IdTGroup)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", responseBody{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scopeGroup.Id, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Check if user is allowed to create scan scopes of this kind
		if !contextUser.Admin && !scopeGroup.AllowNetwork {
			core.RespondAuthError(ctx)
			return
		}

		// Request owned scan scopes from manager
		groupScopes, errScopes := manager.RpcGetScopesOf(logger, core.RpcClient(), []uint64{scopeGroup.Id})
		if errScopes != nil {
			logger.Warningf("Could not query scopes of group %d: %s", scopeGroup.Id, errScopes)
			core.RespondTemporaryError(ctx)
			return
		}

		// Check if new scope can be created
		if createScope {
			if scopeGroup.MaxScopes >= 0 && len(groupScopes) >= scopeGroup.MaxScopes {
				core.Respond(ctx, true, "Scope limit reached.", responseBody{})
				return
			}
		}

		// Unify values and remove duplicates
		assetCompanies := scanUtils.UniqueStrings(req.AssetCompanies)
		assetDepartments := scanUtils.UniqueStrings(req.AssetDepartments)
		assetRoutingDomains := scanUtils.UniqueStrings(req.AssetRoutingDomains)
		assetZones := scanUtils.UniqueStrings(req.AssetZones)
		assetPurposes := scanUtils.UniqueStrings(req.AssetPurposes)
		assetCountries := scanUtils.UniqueStrings(req.AssetCountries)
		assetLocations := scanUtils.UniqueStrings(req.AssetLocations)
		assetContacts := scanUtils.UniqueStrings(req.AssetContacts)
		assetExcludeKeywords := scanUtils.UniqueStrings(req.AssetExcludeKeywords)

		// Define attributes to store with scan scope (don't just pass on arbitrary user input!)
		attributes := utils.JsonMap{
			"sync":                   req.Sync,
			"asset_companies":        assetCompanies,
			"asset_departments":      assetDepartments,
			"asset_routing_domains":  assetRoutingDomains,
			"asset_zones":            assetZones,
			"asset_purposes":         assetPurposes,
			"asset_countries":        assetCountries,
			"asset_locations":        assetLocations,
			"asset_contacts":         assetContacts,
			"asset_exclude_keywords": assetExcludeKeywords,
		}

		// Execute creation of scan scope
		if createScope {

			// Prepare scope type name
			scopeType := "networks"
			if req.Type != "" {
				scopeType = req.Type
			}

			// Chose database server by ID to use for the new scan scope
			dbServerId := scopeGroup.DbServerId // Use other database server if assigned for scope group
			if dbServerId <= 0 {
				dbServerId = 1 // Use default database server is no other is specified
			}

			// Execute create on manager via RPC
			_, errCreate := manager.RpcCreateScope(
				logger,
				core.RpcClient(),
				dbServerId,
				req.Name,
				scopeGroup.Id,
				contextUser.Email,
				false,
				req.Cycles,
				req.CyclesRetention,
				scopeType,
				attributes,
			)
			if errCreate != nil {
				logger.Warningf("Could not create scan scope: %s", errCreate)
				core.RespondTemporaryError(ctx) // Return generic error information
				return
			}

			// Log event
			errEvent := database.NewEvent(
				contextUser,
				database.EventScopeCreate,
				fmt.Sprintf("Scope: %s", req.Name),
			)
			if errEvent != nil {
				logger.Errorf("Could not create event log: %s", errEvent)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Return response
			core.Respond(ctx, false, "Scope created.", responseBody{})
		}

		// Execute update of scan scope
		if updateScope {

			// Request manager to update scan scope
			errRpc := manager.RpcUpdateScope(
				logger,
				core.RpcClient(),
				scopeId,
				req.Name,
				req.Cycles,
				req.CyclesRetention,
				&attributes,
			)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Return response
			core.Respond(ctx, false, "Scope updated.", responseBody{})
		}
	}
}

// ScopeCreateUpdateAssets creates or updates a scan scope configuration to be imported from an asset
// inventory. If a group ID is supplied, a new scan scope will be created. If a scope ID is provided, an update will
// be performed. Only executes, if the user has ownership rights (or is admin).
// ATTENTION: This only creates a scan scope with an import definition. Scan input targets are not inserted into the
//
//	scan scope database. The importer component takes care of maintaining scan input targets according to
//	this scan scope import definition.
var ScopeCreateUpdateAssets = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		ExistingScopeId  *uint64  `json:"scope_id"` // Set if EXISTING scope shall be updated. PK of the DB scope element to identify associated entry for update
		GroupId          *uint64  `json:"group_id"` // Set if NEW scope shall be created. PK of DB group element to create scope for
		Type             string   `json:"type"`     // Type flag relevant for the importer to decide which repository to load from
		Name             string   `json:"name"`
		Cycles           bool     `json:"cycles"`
		CyclesRetention  int      `json:"cycles_retention"` // >=1 for the amount of scan cycles to keep. -1 to keep all scan results. 0 is not allowed for safety reasons!
		Sync             bool     `json:"sync"`
		AssetType        string   `json:"asset_type"`
		AssetCompanies   []string `json:"asset_companies"`
		AssetDepartments []string `json:"asset_departments"`
		AssetCountries   []string `json:"asset_countries"`
		AssetLocations   []string `json:"asset_locations"`
		AssetContacts    []string `json:"asset_contacts"`
		AssetCritical    string   `json:"asset_critical"`
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

		// Abort request because it is invalid
		if (req.ExistingScopeId == nil && req.GroupId == nil) || (req.ExistingScopeId != nil && req.GroupId != nil) {
			logger.Debugf("Scope ID to update or group ID to create scope required.")
			core.Respond(ctx, true, "Scope ID or group ID required.", responseBody{})
			return
		}

		// Check if scope name is defined
		if len(req.Name) == 0 {
			core.Respond(ctx, true, "Invalid scope name.", responseBody{})
			return
		}

		// Don't allow 0 retention cycles, but change it to -1 (keep all). In case of a bug, cycle retention would
		// be unintentionally zero, causing all scan results (outside the current scan cycle) to be wiped.
		if req.CyclesRetention <= 0 {
			req.CyclesRetention = -1
		}

		// Prepare some memory variables
		createScope := req.ExistingScopeId == nil || *req.ExistingScopeId <= 0
		scopeId := uint64(0)             // To be filled later, in update mode
		var scopeGroup *database.T_group // To be set later

		// Prepare creation of new scan scope
		if createScope {

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(*req.GroupId)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", responseBody{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		} else { // Prepare update of existing scan scope

			// Request scope details from manager
			scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), *req.ExistingScopeId)
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

			// Remember existing scope ID
			scopeId = scanScope.Id

			// Get group to create scope for
			groupEntry, errGroupEntry := database.GetGroup(scanScope.IdTGroup)
			if errGroupEntry != nil {
				logger.Warningf("Could not query group: %s", errGroupEntry)
				core.RespondInternalError(ctx)
				return
			}

			// Check if group exists
			if groupEntry == nil {
				core.Respond(ctx, true, "Invalid group.", responseBody{})
				return
			}

			// Select group entry
			scopeGroup = groupEntry
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scopeGroup.Id, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Check if user is allowed to create scan scopes of this kind
		if !contextUser.Admin && !scopeGroup.AllowAsset {
			core.RespondAuthError(ctx)
			return
		}

		// Request owned scan scopes from manager
		groupScopes, errScopes := manager.RpcGetScopesOf(logger, core.RpcClient(), []uint64{scopeGroup.Id})
		if errScopes != nil {
			logger.Warningf("Could not query scopes of group %d: %s", scopeGroup.Id, errScopes)
			core.RespondTemporaryError(ctx)
			return
		}

		// Check if new scope can be created
		if createScope {
			if scopeGroup.MaxScopes >= 0 && len(groupScopes) >= scopeGroup.MaxScopes {
				core.Respond(ctx, true, "Scope limit reached.", responseBody{})
				return
			}
		}

		// Check some passed values for plausibility
		if !scanUtils.StrContained(req.AssetType, []string{"Any", "Server", "Network", "Client"}) {
			core.Respond(ctx, true, "Invalid asset type value.", responseBody{})
			return
		}
		if !scanUtils.StrContained(req.AssetCritical, []string{"Any", "Yes", "No"}) {
			core.Respond(ctx, true, "Invalid critical value.", responseBody{})
			return
		}

		// Unify values and remove duplicates
		assetCompanies := scanUtils.UniqueStrings(req.AssetCompanies)
		assetDepartments := scanUtils.UniqueStrings(req.AssetDepartments)
		assetCountries := scanUtils.UniqueStrings(req.AssetCountries)
		assetLocations := scanUtils.UniqueStrings(req.AssetLocations)
		assetContacts := scanUtils.UniqueStrings(req.AssetContacts)

		// Define attributes to store with scan scope (don't just pass on arbitrary user input!)
		attributes := utils.JsonMap{
			"sync":              req.Sync,
			"asset_type":        strings.TrimSpace(req.AssetType),
			"asset_companies":   assetCompanies,
			"asset_departments": assetDepartments,
			"asset_countries":   assetCountries,
			"asset_locations":   assetLocations,
			"asset_contacts":    assetContacts,
			"asset_critical":    strings.TrimSpace(req.AssetCritical),
		}

		// Execute creation of scan scope
		if createScope {

			// Prepare scope type name
			scopeType := "assets"
			if req.Type != "" {
				scopeType = req.Type
			}

			// Chose database server by ID to use for the new scan scope
			dbServerId := scopeGroup.DbServerId // Use other database server if assigned for scope group
			if dbServerId <= 0 {
				dbServerId = 1 // Use default database server is no other is specified
			}

			// Execute create on manager via RPC
			_, errCreate := manager.RpcCreateScope(
				logger,
				core.RpcClient(),
				dbServerId,
				req.Name,
				scopeGroup.Id,
				contextUser.Email,
				false,
				req.Cycles,
				req.CyclesRetention,
				scopeType,
				attributes,
			)
			if errCreate != nil {
				logger.Warningf("Could not create scan scope: %s", errCreate)
				core.RespondTemporaryError(ctx) // Return generic error information
				return
			}

			// Log event
			errEvent := database.NewEvent(
				contextUser,
				database.EventScopeCreate,
				fmt.Sprintf("Scope: %s", req.Name),
			)
			if errEvent != nil {
				logger.Errorf("Could not create event log: %s", errEvent)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Return response
			core.Respond(ctx, false, "Scope created.", responseBody{})
		}

		// Execute update of scan scope
		if !createScope {

			// Request manager to update scan scope
			errRpc := manager.RpcUpdateScope(
				logger,
				core.RpcClient(),
				scopeId,
				req.Name,
				req.Cycles,
				req.CyclesRetention,
				&attributes,
			)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Return response
			core.Respond(ctx, false, "Scope updated.", responseBody{})
		}
	}
}

// ScopeResetInput resets the scan status flags of a scope input to add it back to queue again
var ScopeResetInput = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		ScopeId uint64 `json:"scope_id"` // Scope ID to identify associated entry
		Input   string `json:"input"`
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

		// Check if primary key is defined, otherwise gorm cannot update specific entry
		if req.ScopeId == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
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

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to reset a scan scope's input entry
		errRpc := manager.RpcResetInput(logger, core.RpcClient(), req.ScopeId, req.Input)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Scope target reset.", responseBody{})
	}
}

// ScopeResetSecret sets a new scope secret used to associate scan agents, if the user has ownership rights (or is admin)
var ScopeResetSecret = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id uint64 `json:"id"` // PK of the DB element to identify associated entry for update
	}

	// Define expected response structure
	type responseBody struct {
		Secret string `json:"secret"` // Only returned if it couldn't be sent via encrypted e-mail
	}

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

		// Check if primary key is defined, otherwise gorm cannot update specific entry
		if req.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Request scope details from manager
		scanScope, errScanScope := manager.RpcGetScope(logger, core.RpcClient(), req.Id)
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

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to update scope secret
		newToken, errNewToken := manager.RpcResetSecret(logger, core.RpcClient(), req.Id)
		if errors.Is(errNewToken, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errNewToken != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Prepare response body
		msg := ""
		body := responseBody{}

		// Send new scope secret to user via encrypted e-mail
		if _build.DevMode {
			logger.Infof("Skipping user e-mail notification during development.")
			logger.Infof("Set '%s' as development scope secret for scan scope '%s'.", newToken, scanScope.Name)

			// Set response message
			msg = "Scope secret set."

			// Expose new password once through web interface in a non-persistent way
			body.Secret = newToken
		} else if len(contextUser.Certificate) > 0 {

			// Log action
			logger.Debugf("Sending new scope secret to requesting user via e-mail.")

			// Set response message
			msg = "Scope secret sent via e-mail."

			// Prepare mail values
			subject := fmt.Sprintf("New secret for scan scope '%s'", scanScope.Name)
			message := fmt.Sprintf("Scan Scope:\t%s\n"+
				"Scope Secret:\t%s\n\n"+
				"Scan scopes can be configured at %s.",
				scanScope.Name,
				newToken,
				ctx.Request.Host, // Prepare dynamically, website might be accessed via different domains
			)

			// Send new token to user via encrypted e-mail
			recipientCerts := [][]byte{contextUser.Certificate}
			errMail := smtp.SendMail(
				smtpConnection.Server,
				smtpConnection.Port,
				smtpConnection.Username,
				smtpConnection.Password,
				smtpConnection.Sender,
				[]mail.Address{{Name: contextUser.Name + " " + contextUser.Surname, Address: contextUser.Email}},
				recipientCerts,
				subject,
				[]byte(message),
				nil,
				smtpConnection.OpensslPath,
				smtpConnection.SignatureCert,
				smtpConnection.SignatureKey,
				false,
			)
			if errMail != nil {
				logger.Errorf(
					"Could not send new secret of scan scope %d to user '%s': %s",
					req.Id,
					contextUser.Email,
					errMail,
				)
				core.Respond(ctx, true, "Could not e-mail new scope secret.", responseBody{})
				return
			}
		} else {

			// Log action
			logger.Debugf("Returning new scope secret to requesting user via web interface.")

			// Set response message
			msg = "Scope secret set."

			// Expose new password once through web interface in a non-persistent way
			body.Secret = newToken
		}

		// Log event
		errEvent := database.NewEvent(contextUser, database.EventScopeSecret, "")
		if errEvent != nil {
			logger.Errorf("Could not create event log: %s", errEvent)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, msg, body)
	}
}
