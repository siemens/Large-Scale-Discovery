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
	"github.com/microcosm-cc/bluemonday"
	"github.com/siemens/Large-Scale-Discovery/_build"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/ZapSmtp/smtp"
	"net/mail"
	"strings"
	"time"
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
		var userEntries = make([]database.T_user, 0, 0) // Initialize empty slice to avoid returning nil to frontend
		var errUserEntriesUnfiltered error

		// Query users depending on whether user is admin or not
		if contextUser.Admin {

			// Query all entries
			userEntries, errUserEntriesUnfiltered = database.GetUsers()
			if errUserEntriesUnfiltered != nil {
				logger.Errorf("Could not query existing users: %s", errUserEntriesUnfiltered)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}

		} else {

			// Query all entries
			var userEntriesUnfiltered []database.T_user
			userEntriesUnfiltered, errUserEntriesUnfiltered = database.GetUsers()
			if errUserEntriesUnfiltered != nil {
				logger.Errorf("Could not query existing users: %s", errUserEntriesUnfiltered)
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

// UserUpdate updates user details of a certain user, if the requesting user has has admin rights
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

// UserDelete deletes a certain user, if the requesting user has has admin rights
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

		// Prepare mail values
		subject := fmt.Sprintf("Large-Scale Discovery - Temporary Credentials")
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

		// Enable encryption by setting user certificate, if available
		var encCert [][]byte
		if len(contextUser.Certificate) > 0 {
			encCert = [][]byte{contextUser.Certificate}
		}

		// Send new token to user via encrypted e-mail
		if _build.DevMode {
			logger.Infof("Skipping user e-mail notification during development.")
			logger.Infof("Set '%s' as development DB password for user '%s'.", dbPassword, contextUser.Email)
		} else {
			logger.Debugf("Sending new database password to requesting user via e-mail.")
			errMail := smtp.SendMail3(
				smtpConnection.Server,
				smtpConnection.Port,
				smtpConnection.Username,
				smtpConnection.Password,
				smtpConnection.Sender,
				[]mail.Address{{Name: contextUser.Name + " " + contextUser.Surname, Address: contextUser.Email}},
				subject,
				message,
				smtpConnection.OpensslPath,
				smtpConnection.SignatureCert,
				smtpConnection.SignatureKey,
				encCert,
				"",
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
		}

		// Log event
		errEvent := database.NewEvent(contextUser, database.EventDbPassword, "")
		if errEvent != nil {
			logger.Errorf("Could not create event log: %s", errEvent)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		core.Respond(ctx, false, "Database password sent via E-mail.", responseBody{})
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
			core.Respond(ctx, true, "Invalid feedback subject.", responseBody{})
			return
		}
		if len(req.Message) == 0 {
			core.Respond(ctx, true, "Invalid feedback message.", responseBody{})
			return
		}

		// Prepare values
		subject := fmt.Sprintf("LSD Feedback: %s", req.Subject)
		message := fmt.Sprintf(
			"User:\t\t%s %s\nMail:\t\t%s\nMessage:\t\t%s",
			contextUser.Name,
			contextUser.Surname,
			contextUser.Email,
			req.Message,
		)

		// Send email with feedback
		errSend := smtp.SendMail3(
			smtpConnection.Server,
			smtpConnection.Port,
			smtpConnection.Username,
			smtpConnection.Password,
			smtpConnection.Sender,
			smtpConnection.Recipients,
			subject,
			message,
			"",
			nil,
			nil,
			smtpConnection.EncryptionCerts,
			"",
		)
		if errSend != nil {
			logger.Errorf(
				"Could not send feedback mail: %s. \nFrom: %s %s (%s)\nSubject: %s\nMessage: %s",
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
		core.Respond(ctx, false, "Thank you for your feedback!", responseBody{})
	}
}
