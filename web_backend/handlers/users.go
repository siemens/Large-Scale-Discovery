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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/siemens/Large-Scale-Discovery/_build"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/ZapSmtp/smtp"
)

// Users gets the list of existing users a user is allowed to see
// - all in case of admin
// - active users of the same company in case of user
var Users = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Users []database.T_user `json:"users"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Prepare memory for list of groups
		var userEntries = make([]database.T_user, 0) // Initialize empty slice to avoid returning nil to frontend
		var errUserEntries error

		// Query users depending on whether user is admin or not
		if contextUser.Admin {

			// Query all entries
			userEntries, errUserEntries = database.GetUsers()
			if errUserEntries != nil {
				logger.Errorf("Could not query existing users: %s", errUserEntries)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

		} else {

			// Query all entries
			var userEntriesUnfiltered []database.T_user
			userEntriesUnfiltered, errUserEntries = database.GetUsers()
			if errUserEntries != nil {
				logger.Errorf("Could not query existing users: %s", errUserEntries)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

			// Filter for enabled users of the same company
			userEntries = make([]database.T_user, 0, len(userEntriesUnfiltered)) // Resize slice to avoid redundant resizing
			for _, entry := range userEntriesUnfiltered {
				if entry.Active && entry.Company == contextUser.Company {
					userEntries = append(userEntries, entry)
				}
			}
		}

		// Prepare response body
		body := responseBody{
			Users: userEntries,
		}

		// Return response
		core.Respond(ctx, false, "Users retrieved.", body)
	}
}

// UserUpdate updates user details of a certain user, if the requesting user has admin rights
var UserUpdate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id      uint64  `json:"id"` // PK of the DB element to identify associated entry for update
		Active  *bool   `json:"active"`
		Admin   *bool   `json:"admin"`
		Demo    *bool   `json:"demo"`
		Name    *string `json:"name"`
		Surname *string `json:"surname"`
		Company *string `json:"company"`
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

		// Check if primary key is defined, otherwise gorm cannot update specific item
		if req.Id == 0 {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Get user to update
		userEntry, errUserEntry := database.GetUser(req.Id)
		if errUserEntry != nil {
			logger.Errorf("Could not query user: %s", errUserEntry)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user exists
		if userEntry == nil {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Add field to be updated, if contained in the request
		if req.Active != nil {

			// Execute if active state changed
			if *req.Active != userEntry.Active {

				// Add active value to update values
				userEntry.Active = *req.Active

				// Enable/Disable user on affected database servers
				var errRpc error
				if *req.Active {
					errRpc = manager.RpcEnableDbUser(logger, core.RpcClient(), userEntry.Email)
				} else {
					errRpc = manager.RpcDisableDbUser(logger, core.RpcClient(), userEntry.Email)
				}

				// Check enable/disable result
				if errors.Is(errRpc, utils.ErrRpcConnectivity) {
					core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
					return
				} else if errRpc != nil {
					core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
					return
				}
			}
		}
		if req.Admin != nil {
			userEntry.Admin = *req.Admin
		}
		if req.Demo != nil {
			userEntry.Demo = *req.Demo
		}
		if req.Name != nil {
			userEntry.Name = *req.Name
		}
		if req.Surname != nil {
			userEntry.Surname = *req.Surname
		}
		if req.Company != nil {
			userEntry.Company = *req.Company
		}

		// Save updated attributes
		saved, errSave := userEntry.Save("active", "admin", "demo", "name", "surname", "company")
		if errSave != nil {
			logger.Errorf("Could not update user entry: %s", errSave)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Prepare response
		var reqErr bool
		var reqMsg string
		if saved > 0 {
			reqErr = false
			reqMsg = "User updated."
		} else {
			reqErr = true
			reqMsg = "User not found."
		}

		// Return response
		core.Respond(ctx, reqErr, reqMsg, responseBody{})
	}
}

// UserDelete deletes a certain user, if the requesting user has admin rights
var UserDelete = func() gin.HandlerFunc {

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

		// Get user to delete
		userEntry, errUserEntry := database.GetUser(req.Id)
		if errUserEntry != nil {
			logger.Errorf("Could not query user: %s", errUserEntry)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user exists
		if userEntry == nil {
			core.Respond(ctx, true, "Invalid ID.", responseBody{})
			return
		}

		// Query list of current administrators
		administratorEntries, errAdministratorEntries := database.GetAdministrators()
		if errAdministratorEntries != nil {
			logger.Errorf("Could not query administrator users: %s", errAdministratorEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user is the only remaining administrator
		if userEntry.Admin && len(administratorEntries) <= 1 {
			core.Respond(ctx, true, "User is the only remaining administrator.", responseBody{})
			return
		}

		// Get groups the user is owner of
		groupEntries, errGroupEntries := database.GetGroupsOfUser(userEntry.Id)
		if errGroupEntries != nil {
			logger.Errorf("Could not query groups: %s", errGroupEntries)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user is the only remaining owner of any group
		for _, group := range groupEntries {
			if len(group.Ownerships) == 1 {
				core.Respond(ctx, true, "User is the last owner of a group.", responseBody{})
				return
			}
		}

		// Request views granted from manager
		scopeViews, errScopeViews := manager.RpcGetViewsGranted(logger, core.RpcClient(), userEntry.Email)
		if errors.Is(errScopeViews, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScopeViews != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Revoke access right from each view
		for _, view := range scopeViews {

			// Request manager to revoke view grants for given username
			errRpc := manager.RpcRevokeGrants(logger, core.RpcClient(), view.Id, userEntry.Email)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
				return
			}
		}

		// Cleanup user's group ownerships
		for _, ownership := range userEntry.Ownerships {
			errDeleteOwnership := ownership.Delete()
			if errDeleteOwnership != nil {
				logger.Errorf("Could not cleanup ownership entry: %s", errDeleteOwnership)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}
		}

		// Remove user from database
		errDelete := userEntry.Delete()
		if errDelete != nil {
			logger.Errorf("Could not delete entry: %s", errDelete)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, "User deleted.", responseBody{})
	}
}

// UserDetails returns the current user details of the authenticated user. This function is useful to sync
// cached user data in the application frontend with the latest data stored in the backend.
var UserDetails = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Id      uint64    `json:"id"`
		Email   string    `json:"email"`
		Name    string    `json:"name"`
		Surname string    `json:"surname"`
		Gender  string    `json:"gender"`
		Admin   bool      `json:"admin"`  // Whether the user has full admin rights (to control visible components, verification must be done on the backend)
		Owner   bool      `json:"owner"`  // Whether the user has scan scope management rights (to control visible components, verification must be done on the backend)
		Access  bool      `json:"access"` // Whether the user has at least one scope view granted (to control visible components, verification must be done on the backend)
		Demo    bool      `json:"demo"`
		Created time.Time `json:"created"`
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Request views granted from manager
		scopeViews, _ := manager.RpcGetViewsGranted(logger, core.RpcClient(), contextUser.Email)

		// Prepare response body
		body := responseBody{
			Id:      contextUser.Id,
			Email:   contextUser.Email,
			Name:    contextUser.Name,
			Surname: contextUser.Surname,
			Gender:  contextUser.Gender,
			Admin:   contextUser.Admin,
			Owner:   len(contextUser.Ownerships) > 0,
			Access:  len(scopeViews) > 0,
			Demo:    contextUser.Demo,
			Created: contextUser.Created,
		}

		// Return response
		core.Respond(ctx, false, "User profile retrieved.", body)
	}
}

// UserResetDbPassword resets the user's database password and updates it on all linked database servers
var UserResetDbPassword = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Password string `json:"password"` // Only returned if it couldn't be sent via encrypted e-mail
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Generate password
		dbPassword, errPassword := utils.GenerateToken(
			strings.Replace(utils.AlphaNumCaseSymbol, "?", "", -1), 15) // Drop question mark from character set, as gorm has issues with them
		if errPassword != nil {
			logger.Errorf("Could not generate user password: %s", errPassword)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Hash password into Postgresql format. Postgresql allows to directly set hashes (in Postgres-specific seeded
		// format), instead of cleartext user passwords. This way only need to store the hash of the user's password
		// locally for management and still use it to manage our database instances.
		dbPasswordHash, errDbPasswordHash := utils.HashScramSha256Postgres(dbPassword)
		if errDbPasswordHash != nil {
			logger.Errorf("Could not convert user password to Postgres hash: %s", errDbPasswordHash)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Request manager to update user password
		errRpc := manager.RpcUpdateDatabaseCredentials(logger, core.RpcClient(), contextUser.Email, dbPasswordHash)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(ctx) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errRpc != nil {
			core.RespondInternalError(ctx) // Return generic error information. Situation already logged!
			return
		}

		// Update attributes
		contextUser.DbPasswordHash = dbPasswordHash

		// Save updated attributes
		saved, errSave := contextUser.Save("db_password")
		if errSave != nil {
			logger.Errorf("Could not update user entry: %s", errSave)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Warn if there was more than one entry updated
		if saved != 1 {
			logger.Errorf("Updated an invalid amount of users: %d", saved)
		}

		// Prepare response body
		msg := ""
		body := responseBody{}

		// Enable encryption by setting user certificate, if available
		if _build.DevMode {

			// Log action
			logger.Infof("Skipping user e-mail notification during development.")
			logger.Infof("Set '%s' as development DB password for user '%s'.", dbPassword, contextUser.Email)

			// Set response message
			msg = "Database password set."

			// Expose new password once through web interface in a non-persistent way
			body.Password = dbPassword

		} else if len(contextUser.Certificate) > 0 {

			// Log action
			logger.Debugf("Sending new database password to requesting user via e-mail.")

			// Set response message
			msg = "Database password sent via e-mail."

			// Prepare mail values
			subject := "Large-Scale Discovery - Temporary Credentials"
			message := fmt.Sprintf("Here are your personal and TEMPORARY database credentials.\n"+
				"For details please visit %s.\n\n"+
				"Username:\t%s\n"+
				"Password:\t%s\n\n"+
				"This credentials are valid for limited time on all granted scope views!\n\n"+
				"Via the web interface, you can:\n"+
				"   - Request a new personal and temporary database password\n"+
				"   - Find database connection details to access scan results\n"+
				"   - See the scan progress of a certain scan scope\n",
				ctx.Request.Host, // Prepare dynamically, website might be accessed via different domains
				contextUser.Email,
				dbPassword,
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
					"Could not send new database credentials to user '%s': %s",
					contextUser.Email,
					errMail,
				)
				core.Respond(ctx, true, "Could not e-mail new database credentials.", responseBody{})
				return
			}
		} else {

			// Log action
			logger.Debugf("Returning new database password to requesting user via web interface.")

			// Set response message
			msg = "Database password set."

			// Expose new password once through web interface in a non-persistent way
			body.Password = dbPassword
		}

		// Log event
		errEvent := database.NewEvent(contextUser, database.EventDbPassword, "")
		if errEvent != nil {
			logger.Errorf("Could not create event log: %s", errEvent)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, msg, body)
	}
}

// UserFeedback sends an email to the system admin with the submitted feedback
var UserFeedback = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	// Define expected response structure
	type responseBody struct {
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

		// Check authentication
		if contextUser == nil {
			core.RespondAuthError(ctx)
			return
		}

		// Initialize sanitizer
		b := bluemonday.StrictPolicy()

		// Sanitize message data
		req.Subject = b.Sanitize(req.Subject)
		req.Message = b.Sanitize(req.Message)

		// Abort if message is not valid
		if len(req.Subject) == 0 {
			core.Respond(ctx, true, "Invalid message subject.", responseBody{})
			return
		}
		if len(req.Message) == 0 {
			core.Respond(ctx, true, "Invalid message message.", responseBody{})
			return
		}

		// Prepare values
		subject := fmt.Sprintf("LSD Message: %s", req.Subject)
		message := fmt.Sprintf(
			"User:\t\t%s %s\nMail:\t\t%s\nMessage:\t\t%s",
			contextUser.Name,
			contextUser.Surname,
			contextUser.Email,
			req.Message,
		)

		// Send email with contact message
		errSend := smtp.SendMail(
			smtpConnection.Server,
			smtpConnection.Port,
			smtpConnection.Username,
			smtpConnection.Password,
			smtpConnection.Sender,
			smtpConnection.Recipients,
			smtpConnection.EncryptionCerts,
			subject,
			[]byte(message),
			nil,
			"",
			nil,
			nil,
			false,
		)
		if errSend != nil {
			logger.Errorf(
				"Could not send contact mail: %s. \nFrom: %s %s (%s)\nSubject: %s\nMessage: %s",
				errSend,
				contextUser.Name,
				contextUser.Surname,
				contextUser.Email,
				req.Subject,
				req.Message,
			)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, "Thank you for your message!", responseBody{})
	}
}

// UserApiToken issues a long-lived API access token for the currently authenticated user,
// increments the user's ApiTokenRevision to invalidate any previously issued API tokens,
// and delivers the new token via S/MIME-encrypted e-mail. If no S/MIME certificate is available
// for the user, the token is returned once in the response body as a fallback.
var UserApiToken = func(smtpConnection *utils.Smtp) gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Token     string    `json:"token"`      // Only populated as a fallback when no S/MIME certificate is on file
		ExpiresAt time.Time `json:"expires_at"` // Expiry of the issued token
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)
		if contextUser == nil {
			core.RespondAuthError(ctx)
			return
		}

		// Check if user has rights to perform action
		if contextUser.Demo {
			core.RespondAuthError(ctx)
			return
		}

		// Bump the API token revision counter - this atomically invalidates any previously
		// issued API token for this user, so a user can always "rotate" by requesting a new
		// one. Must be saved before minting the new token so the new token's revision claim
		// matches the stored value.
		contextUser.ApiTokenRevision += 1
		if _, errSave := contextUser.Save("api_token_revision"); errSave != nil {
			logger.Errorf("Could not update API token revision for user '%s': %s", contextUser.Email, errSave)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Get config
		conf := config.GetConfig()

		// Mint the new API token
		token, expiresAt := core.CreateApiJwt(contextUser.Id, contextUser.ApiTokenRevision, conf.Jwt.ExpiryApi)

		// Prepare response body. Token is only populated as a fallback below.
		body := responseBody{
			ExpiresAt: expiresAt,
		}
		msg := ""

		// Branch on deployment mode / certificate availability
		if _build.DevMode {

			// Log action
			logger.Infof("Skipping API token e-mail notification during development.")
			logger.Infof("Issued API token for user '%s' valid until %s.",
				contextUser.Email, expiresAt.Format(time.RFC3339))

			// Set response message
			msg = "API token generated."
			body.Token = token

		} else if len(contextUser.Certificate) > 0 {

			// Log action
			logger.Debugf("Sending new API token to user '%s' via e-mail.", contextUser.Email)

			// Prepare mail values
			subject := "Large-Scale Discovery - API Access Token"
			message := fmt.Sprintf(
				"Here is your personal API access token for Large-Scale Discovery.\n"+
					"For details please visit %s.\n\n"+
					"Token:\t%s\n"+
					"Valid until:\t%s\n\n"+
					"This token replaces any previously issued API token for your account.\n\n"+
					"Via the web interface, you can:\n"+
					"   - Request a new API access token\n"+
					"   - Find API documentation and usage examples\n"+
					"   - Manage your scan scopes and scan targets\n",
				ctx.Request.Host,
				token,
				expiresAt.Format("2006-01-02 15:04:05"),
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
					"Could not send API token to user '%s': %s",
					contextUser.Email,
					errMail,
				)
				core.Respond(ctx, true, "Could not send API token via e-mail.", responseBody{})
				return
			}

			// Set response message
			msg = "API token sent via e-mail."

		} else {

			// Fallback: no certificate on file, show the token once in the web UI
			logger.Debugf("Returning new API token to user '%s' via web interface.", contextUser.Email)

			// Set response message
			msg = "API token generated."
			body.Token = token
		}

		// Log event
		errEvent := database.NewEvent(
			contextUser,
			database.EventApiToken,
			"expires="+expiresAt.Format(time.RFC3339),
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
