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
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"time"
)

// Events requests events of a certain type, if the user is an administrator
var Events = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Event database.Event `json:"event"`
		Since *time.Time     `json:"since"`
	}

	// Define expected response structure
	type responseBody struct {
		Events []database.T_event `json:"events"`
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

		// Initialize standard time, if none was supplied
		if req.Since == nil {
			req.Since = &time.Time{}
		}

		// Query related events
		events, errEvents := database.GetEvents(req.Event, *req.Since)
		if errEvents != nil {
			logger.Errorf("Could not query events: %s", errEvents)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Prepare response body
		body := responseBody{
			Events: events,
		}

		// Return response
		core.Respond(ctx, false, "Events returned.", body)
	}
}
