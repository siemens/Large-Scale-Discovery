/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
)

// ScopeAgents groups visible scan agents by scope and identifies the current user's ownership
type ScopeAgents struct {
	ScopeId        uint64                   `json:"scope_id"`
	ScopeName      string                   `json:"scope_name"`
	ScopeDb        string                   `json:"scope_db"`
	ScopeOwnership bool                     `json:"scope_ownership"`
	Agents         []managerdb.T_scan_agent `json:"agents"`
}

// Agents returns scan agents data grouped by scan scope
var Agents = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		ScopeAgents []ScopeAgents `json:"scope_agents"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Demo users always receive an empty list
		// They could be granted, but should not be allowed to use their grants.
		if contextUser.Demo {
			core.Respond(ctx, false, "Agents retrieved.", responseBody{ScopeAgents: []ScopeAgents{}})
			return
		}

		// Request views granted from manager
		accessibleViews, _ := manager.RpcGetViewsGranted(logger, core.RpcClient(), contextUser.Email)

		// Prepare list of owned groups
		ownedGroups := make(map[uint64]struct{}, len(contextUser.Ownerships))
		for _, ownership := range contextUser.Ownerships {
			ownedGroups[ownership.Group.Id] = struct{}{}
		}

		// Prepare list of accessible scan scopes based on scope views
		accessibleScopeIds := make(map[uint64]struct{}, len(accessibleViews))
		for _, scopeView := range accessibleViews {
			accessibleScopeIds[scopeView.IdTScanScope] = struct{}{}
		}

		// Prepare memory for list of scopes
		var agentEntries []managerdb.T_scan_agent

		// Query scan agents according to user's role
		if contextUser.Admin || len(ownedGroups) > 0 || len(accessibleScopeIds) > 0 {

			// Query all scan agents before applying user-specific access restrictions
			agentEntriesUnfiltered, errAgentEntriesUnfiltered := manager.RpcGetAgents(logger, core.RpcClient())
			if errors.Is(errAgentEntriesUnfiltered, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errAgentEntriesUnfiltered != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Return all scan agents to administrators without further filtering
			if contextUser.Admin {
				agentEntries = agentEntriesUnfiltered
			} else {

				// Filter scan agents to scopes owned by or accessible to the user
				for _, scanAgent := range agentEntriesUnfiltered {
					_, hasOwnership := ownedGroups[scanAgent.ScanScope.IdTGroup]
					_, hasGrant := accessibleScopeIds[scanAgent.ScanScope.Id]
					if hasOwnership || hasGrant {
						agentEntries = append(agentEntries, scanAgent)
					}
				}
			}
		}

		// Prepare customized list of scan agents
		var agentStats = make([]ScopeAgents, 0, 4) // Initialize empty slice to avoid returning nil to frontend
		for _, agentEntry := range agentEntries {

			// Append scan agent to scan scope's list (cannot be done in a for range loop, because it passes copies
			// of the agents slice)
			found := false
			for i := 0; i < len(agentStats); i++ {
				if agentStats[i].ScopeId == agentEntry.ScanScope.Id {
					agentStats[i].Agents = append(agentStats[i].Agents, agentEntry)
					found = true
					break
				}
			}

			if !found {

				// Introduce new list for new scan scope and include the user's ownership
				_, hasOwnership := ownedGroups[agentEntry.ScanScope.IdTGroup]
				agentStats = append(agentStats, ScopeAgents{
					ScopeId:        agentEntry.ScanScope.Id,
					ScopeName:      agentEntry.ScanScope.Name,
					ScopeDb:        agentEntry.ScanScope.DbName,
					ScopeOwnership: hasOwnership,
					Agents:         []managerdb.T_scan_agent{agentEntry},
				})
			}
		}

		// Prepare response body
		body := responseBody{
			ScopeAgents: agentStats,
		}

		// Return response
		core.Respond(ctx, false, "Agents retrieved.", body)
	}
}

// AgentDelete deletes a certain scan agent stats entry, if the user has ownership rights on the associated scan
// scope (or is admin)
var AgentDelete = func() gin.HandlerFunc {

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

		// Prepare memory for list of agents
		scanAgents, errScanAgents := manager.RpcGetAgents(logger, core.RpcClient())

		// Check for errors occurred while querying groups
		if errors.Is(errScanAgents, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanAgents != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Lookup scan scope the agent belongs to
		var scopeEntry managerdb.T_scan_scope
		for _, scanAgent := range scanAgents {
			if scanAgent.Id == req.Id {
				scopeEntry = scanAgent.ScanScope
				break
			}
		}

		// Check if user has rights to update scope
		if !core.OwnerOrAdmin(scopeEntry.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to delete scan agent stats and associated data
		errRpc := manager.RpcDeleteAgent(logger, core.RpcClient(), req.Id)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Scan agent deleted.", responseBody{})
	}
}
