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
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
)

// BackendSettings returns some application settings that might be relevant for the frontend. E.g. whether the registration
// of users is enabled, etc...
var BackendSettings = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct{}

	// Define expected response structure
	type responseBody struct { // Access tokens will be automatically appended on success
		DevelopmentLogin        bool                     `json:"development_login"`
		CredentialsRegistration bool                     `json:"credentials_registration"`
		FrontendLinks           map[string][]config.Link `json:"frontend_links"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get config
		conf := config.GetConfig()

		// Get credentials enabled setting
		val, _ := conf.Authenticator["credentials_registration"]
		credentialsRegistration, _ := val.(bool)

		// Respond with required authenticator. Empty string indicating arbitrary or default authenticator.
		core.Respond(
			ctx,
			false,
			"",
			responseBody{
				DevelopmentLogin:        _build.DevMode,
				CredentialsRegistration: credentialsRegistration,
				FrontendLinks:           conf.FrontendLinks,
			},
		)
	}
}

// BackendAuthenticator checks whether a redirect to a special authenticator is required for the given e-mail address.
// If no special authenticator is required, and empty string is returned. "development" may be returned in development
// mode to notify the frontend that no password is required. It's up to the frontend to decide where to redirect
// the user, to ask the password from the user or to re-route the authentication request.
var BackendAuthenticator = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Email string `json:"email"`
	}

	// Define expected response structure
	type responseBody struct { // Access tokens will be automatically appended on success
		EntryUrl string `json:"entry_url"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Make sure Email address is set, which is the primary user identifier
		if req.Email == "" {
			logger.Debugf("No login e-mail address supplied.")
			core.Respond(ctx, true, "E-mail address required.", responseBody{})
			return
		}

		// Check if received email address is plausible
		if !utils.IsPlausibleEmail(req.Email) {
			logger.Warningf("Could not authenticate invalid email address '%s'.", req.Email)
			core.RespondAuthError(ctx)
			return
		}

		// Check whether the user should be redirected to a different URL
		entryUrl := core.EntryUrl(req.Email)
		if len(entryUrl) > 0 {
			logger.Debugf("Authentication redirect to '%s' required.", entryUrl)
		}

		// Respond with required authenticator. Empty string indicating arbitrary or default authenticator.
		core.Respond(
			ctx,
			false,
			"",
			responseBody{
				EntryUrl: entryUrl,
			},
		)
	}
}

// BackendLogout increments the user's logout counter, which is contained in every JWT token and allows access for previously
// issued JWT tokens.
var BackendLogout = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Update attributes
		contextUser.LogoutCount += 1

		// Save updated attributes
		_, errSave := contextUser.Save("logout_count")
		if errSave != nil {
			logger.Errorf("Could not update user's logout count: %s", errSave)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Unset current user, otherwise core.Respond() will re-generate and return a valid authentication token!
		core.UnsetContextUser(ctx)

		// Return response
		logger.Debugf("Logout successful.")
		core.Respond(ctx, false, "Authentication successful.", responseBody{})
	}
}
