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
	"github.com/gin-gonic/gin"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"time"
)

// SqlLogs requests event logs based on a given filter, if the user is an administrator
var SqlLogs = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		DbName string    `json:"db_name"`
		Since  time.Time `json:"since"`
	}

	// Define expected response structure
	type responseBody struct {
		Logs []database.T_sql_log `json:"logs"`
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

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Query related logs
		sqlLogs, errSql := manager.RpcGetSqlLogs(logger, core.RpcClient(), req.DbName, req.Since)
		if errSql != nil {
			logger.Errorf("Could not query SQL logs: %s", errSql)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Transform durations into strings for the web frontend
		for i := 0; i < len(sqlLogs); i++ {
			sqlLogs[i].QueryDurationString = time.Duration(sqlLogs[i].QueryDuration).String()
			sqlLogs[i].TotalDurationString = time.Duration(sqlLogs[i].TotalDuration).String()
		}

		// Prepare response body
		body := responseBody{
			Logs: sqlLogs,
		}

		// Return response
		core.Respond(ctx, false, "SQL logs returned.", body)
	}
}
