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
)

// Databases returns database servers configured in the admin interface available as scan scope storage.
// Different database servers might be owned by different customers.
var Databases = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Databases []managerdb.T_db_server `json:"databases"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(ctx)
			return
		}

		// Prepare memory for list of servers
		var dbServers []managerdb.T_db_server
		var errDbServers error

		// Request all database servers from manager
		dbServers, errDbServers = manager.RpcGetDatabases(logger, core.RpcClient())

		// Check for errors occurred while querying groups
		if errors.Is(errDbServers, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errDbServers != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Prepare response body
		body := responseBody{
			Databases: dbServers,
		}

		// Return response
		core.Respond(ctx, false, "Databases retrieved.", body)
	}
}

// DatabaseRemove deletes a certain database server, if no scan scopes currently use it
var DatabaseRemove = func() gin.HandlerFunc {

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

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(ctx)
			return
		}

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

		// Request manager to delete database server. Will only succeed if no scan scope is currently using it!
		errRpc := manager.RpcRemoveDatabase(logger, core.RpcClient(), req.Id)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil && errRpc.Error() == manager.ErrDatabaseInUse.Error() {
			core.Respond(ctx, true, "Database in use.", responseBody{})
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Database removed.", responseBody{})
	}
}

// DatabaseAddUpdate creates or updates database server details. If a database server ID is provided, an update will be performed.
var DatabaseAddUpdate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		ExistingDatabaseId *uint64 `json:"id"` // Set if EXISTING server shall be updated. PK of the database server element to identify associated entry for update
		Name               string  `json:"name"`
		Dialect            string  `json:"dialect"`
		Host               string  `json:"host"`
		HostPublic         string  `json:"host_public"`
		Port               int     `json:"port"`
		Admin              string  `json:"admin"`
		Password           string  `json:"password"`
		Args               string  `json:"args"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(ctx)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if all field values are defined
		if len(req.Name) <= 0 {
			core.Respond(ctx, true, "Invalid DB name.", responseBody{})
			return
		}
		if len(req.Dialect) <= 0 {
			core.Respond(ctx, true, "Invalid DB dialect.", responseBody{})
			return
		}
		if len(req.Host) <= 0 {
			core.Respond(ctx, true, "Invalid DB host.", responseBody{})
			return
		}
		if len(req.HostPublic) <= 0 {
			core.Respond(ctx, true, "Invalid DB host public.", responseBody{})
			return
		}
		if req.Port < 0 || req.Port > 65535 {
			core.Respond(ctx, true, "Invalid DB port.", responseBody{})
			return
		}
		if len(req.Admin) <= 0 {
			core.Respond(ctx, true, "Invalid DB admin.", responseBody{})
			return
		}
		if len(req.Password) != 0 && len(req.Password) < 10 {
			core.Respond(ctx, true, "Insufficient password complexity.", responseBody{})
			return
		}

		// Decide whether to create or update server
		createServer := req.ExistingDatabaseId == nil || *req.ExistingDatabaseId <= 0

		// Check if password is set in create mode
		if createServer && len(req.Password) == 0 {
			core.Respond(ctx, true, "Invalid DB password.", responseBody{})
			return
		}

		// Prepare some memory variables
		var respMsg string

		// Create scope
		if createServer {

			// Execute create on manager via RPC
			errRpc := manager.RpcAddUpdateDatabase(
				logger,
				core.RpcClient(),
				0,
				req.Name,
				req.Dialect,
				req.Host,
				req.HostPublic,
				req.Port,
				req.Admin,
				req.Password, // Will only be updated if not an empty string
				req.Args,
			)
			if errRpc != nil {
				logger.Warningf("Could not create server: %s", errRpc)
				core.RespondTemporaryError(ctx) // Return generic error information
				return
			}

			// Log event
			errEvent := database.NewEvent(
				contextUser,
				database.EventDatabaseAdd,
				fmt.Sprintf("Server: %s; DB: %s", req.Host, req.Name),
			)
			if errEvent != nil {
				logger.Errorf("Could not create event log: %s", errEvent)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Set response message
			respMsg = "Database added."
		}

		// Execute update of database server
		if !createServer {

			// Request manager to update database server
			errRpc := manager.RpcAddUpdateDatabase(
				logger,
				core.RpcClient(),
				*req.ExistingDatabaseId,
				req.Name,
				req.Dialect,
				req.Host,
				req.HostPublic,
				req.Port,
				req.Admin,
				req.Password, // Will only be updated if not an empty string
				req.Args,
			)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil && errRpc.Error() == manager.ErrDatabaseDuplicate.Error() {
				core.Respond(ctx, true, "Duplicate database.", responseBody{})
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}

			// Set response message
			respMsg = "Database updated."
		}

		// Return response
		core.Respond(ctx, false, respMsg, responseBody{})
		return
	}
}
