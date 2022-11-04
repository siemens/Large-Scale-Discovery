/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
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
	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/ZapSmtp/smtp"
	"large-scale-discovery/_build"
	manager "large-scale-discovery/manager/core"
	"large-scale-discovery/utils"
	"large-scale-discovery/web_backend/core"
	"large-scale-discovery/web_backend/database"
	"net/mail"
)

// Groups returns a list of groups the requesting user is owner of
// - all in case of admin
var Groups = func() gin.HandlerFunc {

	// Define expected response structure
	type responseBody struct {
		Groups []database.T_group `json:"groups"`
	}

	// Return request handling function
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(context)

		// Get user from context storage
		contextUser := core.GetContextUser(context)

		// Prepare memory for list of groups
		var groupEntries = make([]database.T_group, 0, 3) // Initialize empty slice to avoid returning nil to frontend
		var errGroupEntries error

		// Query groups, depending on whether user is admin or not
		if contextUser.Admin {

			// Query all groups
			groupEntries, errGroupEntries = database.GetGroups()
			if errGroupEntries != nil {
				logger.Errorf("Could not query existing groups: %s", errGroupEntries)
				core.RespondInternalError(context) // Return generic error information
				return
			}

		} else {

			// Query user's memberships
			groupEntries, errGroupEntries = database.GetGroupsOfUser(contextUser.Id)
			if errGroupEntries != nil {
				logger.Errorf("Could not query user's groups: %s", errGroupEntries)
				core.RespondInternalError(context) // Return generic error information
				return
			}
		}

		// Prepare response body
		body := responseBody{
			Groups: groupEntries,
		}

		// Return response
		core.Respond(context, false, "Groups of user retrieved.", body)
	}
}

// GroupCreate creates a user group, if the requesting user has has admin rights
var GroupCreate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Name       string `json:"name"`
		MaxScopes  int    `json:"max_scopes"`
		MaxViews   int    `json:"max_views"`
		MaxTargets int    `json:"max_targets"`
		MaxUsers   int    `json:"max_owners"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(context)

		// Get user from context storage
		contextUser := core.GetContextUser(context)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(context)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := context.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Prepare new group object
		newGroup := database.T_group{
			Name:       req.Name,
			CreatedBy:  contextUser.Email,
			MaxScopes:  req.MaxScopes,
			MaxViews:   req.MaxViews,
			MaxTargets: req.MaxTargets,
			MaxOwners:  req.MaxUsers,
		}

		// Create new group object
		err := newGroup.Create()
		if err != nil {
			logger.Errorf("Could not create group: %s", err)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Return response
		core.Respond(context, false, "Group created.", responseBody{})
	}
}

// GroupUpdate updates group details of a certain group, if the requesting user has has admin rights
var GroupUpdate = func() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id         uint64  `json:"id"` // PK of the DB element to identify associated entry for update
		Name       *string `json:"name"`
		MaxScopes  *int    `json:"max_scopes"`
		MaxViews   *int    `json:"max_views"`
		MaxTargets *int    `json:"max_targets"`
		MaxUsers   *int    `json:"max_owners"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(context)

		// Get user from context storage
		contextUser := core.GetContextUser(context)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(context)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := context.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if primary key is defined, otherwise gorm cannot update specific item
		if req.Id == 0 {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Get group to update
		groupEntry, errGroupEntry := database.GetGroup(req.Id)
		if errGroupEntry != nil {
			logger.Errorf("Could not query group: %s", errGroupEntry)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if group exists
		if groupEntry == nil {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Update attributes, if contained in the request
		if req.Name != nil {
			groupEntry.Name = *req.Name
		}
		if req.MaxScopes != nil {
			if *req.MaxScopes < 0 {
				groupEntry.MaxScopes = -1
			} else {
				groupEntry.MaxScopes = *req.MaxScopes
			}
		}
		if req.MaxViews != nil {
			if *req.MaxViews < 0 {
				groupEntry.MaxViews = -1
			} else {
				groupEntry.MaxViews = *req.MaxViews
			}
		}
		if req.MaxTargets != nil {
			if *req.MaxTargets < 0 {
				groupEntry.MaxTargets = -1
			} else {
				groupEntry.MaxTargets = *req.MaxTargets
			}
		}
		if req.MaxUsers != nil {
			if *req.MaxUsers < 0 {
				groupEntry.MaxOwners = -1
			} else {
				groupEntry.MaxOwners = *req.MaxUsers
			}
		}

		// Save updated attributes
		saved, errSave := groupEntry.Save("name", "max_scopes", "max_views", "max_targets", "max_owners")
		if errSave != nil {
			logger.Errorf("Could not update group entry: %s", errSave)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Prepare response
		var reqErr bool
		var reqMsg string
		if saved > 0 {
			reqErr = false
			reqMsg = "Group updated."
		} else {
			reqErr = true
			reqMsg = "Group not found."
		}

		// Return response
		core.Respond(context, reqErr, reqMsg, responseBody{})
	}
}

// GroupDelete deletes a certain group, if the requesting user has has admin rights
var GroupDelete = func() gin.HandlerFunc {

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
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(context)

		// Get user from context storage
		contextUser := core.GetContextUser(context)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(context)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := context.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if primary key is defined, otherwise gorm cannot update specific group
		if req.Id == 0 {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Get group to delete
		groupEntry, errGroupEntry := database.GetGroup(req.Id)
		if errGroupEntry != nil {
			logger.Errorf("Could not query group: %s", errGroupEntry)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if group exists
		if groupEntry == nil {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Request owned scan scopes from manager
		scanScopes, errScanScopes := manager.RpcGetScopesOf(logger, core.RpcClient(), []uint64{groupEntry.Id})
		if errors.Is(errScanScopes, utils.ErrRpcConnectivity) {
			core.RespondTemporaryError(context) // Return temporary error because of connection issues. Situation already logged!
			return
		} else if errScanScopes != nil {
			core.RespondInternalError(context) // Return generic error information. Situation already logged!
			return
		}

		// Remove each scope and associated data (views, access rights, users,...)
		for _, groupScope := range scanScopes {

			// Request manager to delete scan scope and associated data
			errRpc := manager.RpcDeleteScope(logger, core.RpcClient(), groupScope.Id)
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {
				core.RespondTemporaryError(context) // Return temporary error because of connection issues. Situation already logged!
				return
			} else if errRpc != nil {
				core.RespondInternalError(context) // Return generic error information. Situation already logged!
				return
			}
		}

		// Execute update in database
		errDelete := groupEntry.Delete()
		if errDelete != nil {
			logger.Errorf("Could not delete group: %s", errDelete)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Return response
		core.Respond(context, false, "Group deleted.", responseBody{})
	}
}

// GroupAssign sets the owners (administrators) of a given group, if the requesting user has has admin rights
var GroupAssign = func(
	frontendUrl string,
	smtpConnection *utils.Smtp,
) gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Id     uint64   `json:"id"`     // PK of the DB element to identify associated entry for update
		Owners []string `json:"owners"` // List of e-mail addresses referencing users that should be set as group admins
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(context)

		// Get user from context storage
		contextUser := core.GetContextUser(context)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(context)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := context.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if primary key is defined, otherwise gorm cannot update specific group
		if req.Id == 0 {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Get group to update
		groupEntry, errGroupEntry := database.GetGroup(req.Id)
		if errGroupEntry != nil {
			logger.Errorf("Could not query group: %s", errGroupEntry)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Check if group exists
		if groupEntry == nil {
			core.Respond(context, true, "Invalid ID.", responseBody{})
			return
		}

		// Unify values and remove duplicates
		ownerEmails := scanUtils.UniqueStrings(scanUtils.TrimToLower(req.Owners))

		// Prepare new set of owners
		var newOwners []database.T_user

		// Query referenced users
		for _, ownerEmail := range ownerEmails {

			// Check if input is valid e-mail address
			if !utils.IsPlausibleEmail(ownerEmail) {
				core.Respond(context, true, "Invalid owner.", responseBody{})
				return
			}

			// Query given owner from DB
			userEntry, errUserEntry := database.GetUserByMail(ownerEmail)
			if errUserEntry != nil {
				logger.Errorf("Could not query existing user: %s", errUserEntry)
				core.RespondInternalError(context) // Return generic error information
				return
			}

			// Initialize user if missing
			var owner database.T_user
			if userEntry == nil {

				// Get appropriate loader
				loader := core.GetLoader(logger, ownerEmail)

				// Generate new user
				newUser := database.NewUser(ownerEmail, "", "", "", "")

				// Prepare user object for creation
				// ATTENTION: RefreshUser might update user attributes, but does not yet commit them!
				errTemporary, errInternal, errPublic := loader.RefreshUser(logger, newUser)

				// Abort if there was an error and return response based on error kind
				if len(errPublic) > 0 {
					logger.Debugf("Could not auto-create user '%s': %s", ownerEmail, errPublic)
					core.Respond(context, true, errPublic, responseBody{})
					return
				} else if errTemporary != nil {
					logger.Warningf("Could not auto-create user '%s' from source: %s", ownerEmail, errTemporary)
					core.RespondTemporaryError(context)
					return
				} else if errInternal != nil {
					logger.Errorf("Could not auto-create user '%s': %s", ownerEmail, errInternal)
					core.RespondInternalError(context) // Return generic error information
					return
				}

				// If the loader did neither set a password, nor a SSO ID, make the user a credentials type of user,
				// by setting a (non-functional) password. The user will need to do a password reset to activate
				// its account.
				if len(newUser.Password.String) == 0 && len(newUser.SsoId.String) == 0 {
					newUser.Password = sql.NullString{
						String: "-",
						Valid:  true,
					}
				}

				// Create new user object in DB
				err := newUser.Create()
				if err != nil {
					logger.Errorf("Could not auto-create user: %s", err)
					core.RespondInternalError(context) // Return generic error information
					return
				}

				// Prepare mail values
				subject := fmt.Sprintf("Large-Scale Discovery - Ownership")
				message := fmt.Sprintf("You were granted ownership of some scan group.\n"+
					"For details please visit %s.\n\n"+
					"Via the web interface, you can:\n"+
					"   - Manage assigned scan scope\n"+
					"   - See the scan progress of a certain scan scope\n"+
					"   - Find database connection details to access scan results\n"+
					"   - Request your personal and momentary database password\n"+
					"   - See the scan progress of a certain scan scope\n",
					frontendUrl)

				// Enable encryption by setting user certificate, if available
				var encCert [][]byte
				if len(newUser.Certificate) > 0 {
					encCert = [][]byte{newUser.Certificate}
				}

				// Send new token to user via encrypted e-mail
				if _build.DevMode {
					logger.Infof("Skipping user e-mail notification during development.")
				} else {
					logger.Debugf("Sending scope ownership notification to user via e-mail.")
					errMail := smtp.SendMail3(
						smtpConnection.Server,
						smtpConnection.Port,
						smtpConnection.Username,
						smtpConnection.Password,
						smtpConnection.Sender,
						[]mail.Address{{Name: newUser.Name + " " + newUser.Surname, Address: newUser.Email}},
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
							"Could not send initial database credentials to user '%s': %s",
							newUser.Email,
							errMail,
						)
						core.Respond(context, true, "Could not e-mail initial database credentials.", responseBody{})
						return
					}
				}

				// Assign to outer variable
				owner = *newUser

			} else {
				owner = *userEntry
			}

			// Add user object to list of new users
			newOwners = append(newOwners, owner)
		}

		// Check if total amount of group owners would not exceed maximum
		if groupEntry.MaxOwners >= 0 && len(groupEntry.Ownerships)+len(newOwners) > groupEntry.MaxOwners {
			core.Respond(context, true, "Maximum group owners exceeded.", responseBody{})
			return
		}

		// Apply new admin users to group
		errUpdate := groupEntry.UpdateOwners(newOwners)
		if errUpdate != nil {
			logger.Errorf("Could not update owners: %s", errUpdate)
			core.RespondInternalError(context) // Return generic error information
			return
		}

		// Return response
		core.Respond(context, false, "Owners updated.", responseBody{})
	}
}
