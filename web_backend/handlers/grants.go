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
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/ZapSmtp/smtp"
)

// Grant is a helper struct holding the values to be returned by this handler
type Grant struct {
	managerdb.T_view_grant

	// Expand with additional information useful for the web frontend
	UserCreated    time.Time `json:"user_created"`
	UserLastLogin  time.Time `json:"user_last_login"`
	UserCompany    string    `json:"user_company"`
	UserDepartment string    `json:"user_department"`
	UserIsOwner    bool      `json:"user_is_owner"` // Flag indicating whether this granted user is also an owner of the associated scan scope
	UserIsAdmin    bool      `json:"user_is_admin"` // Flag indicating whether this granted user is also an administrator
}

// ViewGrantToken generates a none user bound access token with a prolonged validity time frame for a given view, if the
// user has ownership rights on the related scope (or is admin)
var ViewGrantToken = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		ViewId      uint64 `json:"view_id"`
		Description string `json:"description"`
		ExpiryDays  uint   `json:"expiry_days"`
	}

	// Define expected response structure
	type responseBody struct {
		Username string `json:"username"` // Only returned if it couldn't be sent via encrypted e-mail
		Password string `json:"password"` // Only returned if it couldn't be sent via encrypted e-mail
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

		// Validate access token expiry value
		if req.ExpiryDays <= 0 { // Maximum allowed value is set and checked by the manager itself
			core.Respond(ctx, true, "Invalid token expiry duration.", responseBody{})
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

		// Check if user has rights to update view
		if !core.OwnerOrAdmin(scopeView.ScanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request new scope access token grant from manager
		tokenUsername, tokenPassword, errToken := manager.RpcGrantToken(
			logger,
			core.RpcClient(),
			scopeView.Id,
			req.Description,
			contextUser.Email,
			time.Hour*24*time.Duration(req.ExpiryDays),
		)
		if errors.Is(errToken, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errToken != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Prepare response body
		msg := ""
		body := responseBody{}

		// Send access token to user via encrypted e-mail
		if _build.DevMode {

			// Log action
			logger.Infof("Skipping user e-mail notification during development.")
			logger.Infof("Generated development access token '%s:%s'.", tokenUsername, tokenPassword)

			// Set response message
			msg = "Access token generated."

			// Expose new credentials once through web interface in a non-persistent way
			body.Username = tokenUsername
			body.Password = tokenPassword

		} else if len(contextUser.Certificate) > 0 {

			// Log action
			logger.Debugf("Sending new access token to issuer via e-mail.")

			// Prepare mail values
			subject := "Large-Scale Discovery - Access Token"
			message := fmt.Sprintf("You generated an access token.\n"+
				"For details please visit %s.\n\n"+
				"Access Token Username:\t%s\n"+
				"Access Token Password:\t%s\n\n"+
				"Via the web interface, you can:\n"+
				"   - Revoke access tokens\n"+
				"   - Request a personal and momentary database password\n"+
				"   - Find database connection details to access scan results\n"+
				"   - See the scan progress of a certain scan scope\n",
				ctx.Request.Host, // Prepare dynamically, website might be accessed via different domains
				tokenUsername,
				tokenPassword,
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
					"Could not send new access token to user '%s': %s",
					contextUser.Email,
					errMail,
				)
				core.Respond(ctx, true, "Could not e-mail new access token.", responseBody{})
				return
			}
		} else {

			// Log action
			logger.Debugf("Returning new access token to issuer via web interface.")

			// Set response message
			msg = "Access token generated."

			// Expose new credentials once through web interface in a non-persistent way
			body.Username = tokenUsername
			body.Password = tokenPassword
		}

		// Log event
		errEvent := database.NewEvent(
			contextUser,
			database.EventViewToken,
			fmt.Sprintf("Token: %s; View: %s", req.Description, scopeView.Name),
		)
		if errEvent != nil {
			logger.Errorf("Could not create event log: %s", errEvent)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, msg, body)
	}
}

// ViewGrantUsers updates access rights of a view and sets them to the given list of users (adding new ones, removing outdated
// ones), if the user has ownership rights on the related scope (or is admin)
var ViewGrantUsers = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		ViewId uint64   `json:"view_id"`
		Users  []string `json:"users"`
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

		// Check if user has rights to update view
		if !core.OwnerOrAdmin(scopeView.ScanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Unify values and remove duplicates
		userEmails := scanUtils.UniqueStrings(scanUtils.TrimToLower(req.Users))

		// Prepare map of desired users
		usersDesired := make(map[string]struct{}, len(userEmails))
		for _, userEmail := range userEmails {
			usersDesired[userEmail] = struct{}{}
		}

		// Prepare list of users to revoke
		var usernamesRevoke []string
		for _, grantEntry := range scopeView.Grants {
			if grantEntry.IsUser { // Ignore access token grant types
				if _, desired := usersDesired[grantEntry.Username]; !desired {
					usernamesRevoke = append(usernamesRevoke, grantEntry.Username)
				}
			}
		}

		// Prepare list of credentials to grant
		var usernamesGrant []managerdb.DbCredentials
		for _, userEmail := range userEmails {

			// Check if input is valid e-mail address
			if !utils.IsPlausibleEmail(userEmail) {
				core.Respond(ctx, true, fmt.Sprintf("Invalid user '%s'.", userEmail), responseBody{})
				return
			}

			// Query given user from DB
			userEntry, errUserEntry := database.GetUserByMail(userEmail)
			if errUserEntry != nil {
				logger.Errorf("Could not query user '%s': %s", userEmail, errUserEntry)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Initialize user if missing
			var userToGrant *database.T_user
			if userEntry == nil {

				// Get appropriate loader
				loader := core.GetLoader(logger, userEmail)

				// Generate new user
				newUser := database.NewUser(userEmail, "", "", "", "")

				// Prepare user object for creation
				// ATTENTION: RefreshUser might update user attributes, but does not yet commit them!
				errTemporary, errPublic, errInternal := loader.RefreshUser(logger, newUser)

				// Abort if there was an error and return response based on error kind
				if len(errPublic) > 0 {
					logger.Debugf("Could not auto-load user '%s' from source: %s", userEmail, errPublic)
					core.Respond(ctx, true, errPublic, responseBody{})
					return
				} else if errTemporary != nil {
					logger.Warningf("Could not auto-load user '%s' from source: %s", userEmail, errTemporary)
					core.RespondTemporaryError(ctx)
					return
				} else if errInternal != nil {
					logger.Errorf("Could not auto-load user '%s' from source: %s", userEmail, errInternal)
					core.RespondInternalError(ctx) // Return generic error information
					return
				}

				// Make user password user, by setting a (non-functional) password, if there is no dedicated
				// authenticator configured. The user will need to do a password reset to activate its account.
				if core.EntryUrl(newUser.Email) == "" {
					newUser.Password = sql.NullString{
						String: "-",
						Valid:  true,
					}
				}

				// Create new user object in DB
				err := newUser.Create()
				if err != nil {
					logger.Errorf("Could not auto-create user '%s': %s", userEmail, err)
					core.RespondInternalError(ctx) // Return generic error information
					return
				}

				// Prepare mail values
				subject := "Large-Scale Discovery - Welcome Inside"
				message := fmt.Sprintf("You were granted access to some scan scope results.\n"+
					"For details please visit %s.\n\n"+
					"Via the web interface, you can:\n"+
					"   - Request your personal and momentary database password\n"+
					"   - Find database connection details to access scan results\n"+
					"   - See the scan progress of a certain scan scope\n",
					ctx.Request.Host, // Prepare dynamically, website might be accessed via different domains
				)

				// Enable encryption by setting user certificate, if available
				var recipientCerts [][]byte
				if len(newUser.Certificate) > 0 {
					recipientCerts = [][]byte{newUser.Certificate}
				}

				// Send new access notification to user via encrypted e-mail
				if _build.DevMode {
					logger.Infof("Skipping user e-mail notification during development.")
				} else {
					logger.Debugf("Sending welcome message to new user via e-mail.")
					errMail := smtp.SendMail(
						smtpConnection.Server,
						smtpConnection.Port,
						smtpConnection.Username,
						smtpConnection.Password,
						smtpConnection.Sender,
						[]mail.Address{{Name: newUser.Name + " " + newUser.Surname, Address: newUser.Email}},
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
							"Could not send welcome message to user '%s': %s",
							newUser.Email,
							errMail,
						)
						core.Respond(ctx, true, "Could not e-mail welcome message.", responseBody{})
						return
					}
				}

				// Assign to outer variable
				userToGrant = newUser

			} else {
				userToGrant = userEntry
			}

			// Check if user is currently active. Don't allow granting disabled user
			if !userToGrant.Active {
				core.Respond(
					ctx,
					true,
					fmt.Sprintf("'%s' is disabled.", userToGrant.Email),
					responseBody{},
				)
				return
			}

			// Append user to list of users to grant
			usernamesGrant = append(usernamesGrant, managerdb.DbCredentials{
				Username: userToGrant.Email,
				Password: userToGrant.DbPasswordHash,
			})
		}

		// Request manager to revoke view grants for given usernames
		errRpc := manager.RpcRevokeGrants(logger, core.RpcClient(), scopeView.Id, usernamesRevoke...)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Request manager to grant view for given username
		errRpc2 := manager.RpcGrantUsers(logger, core.RpcClient(), scopeView.Id, usernamesGrant, contextUser.Email)
		if errors.Is(errRpc2, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc2 != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Prepare list of newly added users
		usernamesPrevious := make([]string, 0, len(scopeView.Grants))
		for _, grant := range scopeView.Grants {
			usernamesPrevious = append(usernamesPrevious, grant.Username)
		}

		// Log event
		for _, grant := range usernamesGrant {

			// Skip if user was granted before
			if scanUtils.StrContained(grant.Username, usernamesPrevious) {
				continue
			}

			// Create event
			errEvent := database.NewEvent(
				contextUser,
				database.EventViewGrant,
				fmt.Sprintf("User: %s; View: %s", grant.Username, scopeView.Name),
			)
			if errEvent != nil {
				logger.Errorf("Could not create event log: %s", errEvent)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}
		}

		// Return response
		core.Respond(ctx, false, "Updated view access.", responseBody{})
	}
}

// ViewGrantRevoke revokes grant from scope view, if the requesting user has ownership rights on the related scope (or
// is admin). Grant may be a user bound access right or an access token based access right.
var ViewGrantRevoke = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		ViewId   uint64 `json:"view_id"`
		Username string `json:"username"`
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

		// Check if user has rights to update view
		if !core.OwnerOrAdmin(scopeView.ScanScope.IdTGroup, contextUser) {
			core.RespondAuthError(ctx)
			return
		}

		// Request manager to revoke view grants for given usernames
		errRpc := manager.RpcRevokeGrants(logger, core.RpcClient(), scopeView.Id, req.Username)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Return response
		core.Respond(ctx, false, "Revoked access.", responseBody{})
	}
}
