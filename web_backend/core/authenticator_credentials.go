/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"strings"
)

// init automatically registers the authenticator implemented in this file. If you don't want this authenticator,
// just remove it. You can also add your own authenticator by adding a file with your dedicated implementation.
func init() {
	if _build.DevMode {
	} else {

		// Register authenticator for initialization
		authenticators = append(authenticators, NewAuthenticatorCredentials(nil))
	}
}

const passMinLength = 12

type AuthenticatorCredentials struct {
	domains      []string
	entryUrl     string
	registration bool // Whether users can sign up on their own
}

// NewAuthenticatorCredentials generates a new authenticator with the user domains it is responsible for. Everything
// else of the authenticator will be initialized later during core initialization, with the actual config values.
func NewAuthenticatorCredentials(domains []string) *AuthenticatorCredentials {
	return &AuthenticatorCredentials{
		domains: domains,
	}
}

// Domains returns the user domains this authenticator got registered for
func (a *AuthenticatorCredentials) Domains() []string {
	return a.domains
}

// Init initializes necessary routes and handlers to authenticate users. The argument "domain" restricts the authority
// of this authenticator to a specific user domain (if set), in order to:
//   - prevent impersonation of users by a rogue authentication provider.
//   - prevent authentication downgrades (e.g., OAUTH users using weaker authenticator).
//
// The following routes are deployed:
//   - /v1/auth/register
//   - /v1/auth/login
func (a *AuthenticatorCredentials) Init(conf map[string]interface{}) error {

	// Check whether required arguments are available and valid
	val, _ := conf["credentials_registration"]
	registration, _ := val.(bool)

	// Initialize authenticator
	a.registration = registration

	// Register the necessary authenticator endpoints (routes) to handle credentials authentication attempts.
	RegisterApiEndpointNoAuth("v1", "POST", "/auth/register", a.register())
	RegisterApiEndpointNoAuth("v1", "POST", "/auth/login", a.login())

	// Return nil as everything went fine
	return nil
}

// EntryUrl returns an URL to redirect to in order to initiate authentication, if required. Returns empty string if
// no redirect is required.
func (a *AuthenticatorCredentials) EntryUrl() string {
	return a.entryUrl
}

// register creates a new user for credential authentication
func (a *AuthenticatorCredentials) register() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Email    string  `json:"email"` // User ID
		Password string  `json:"password"`
		Name     *string `json:"name"`
		Surname  *string `json:"surname"`
	}

	// Define expected response structure
	type responseBody struct{}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Abort registration if disabled
		if !a.registration {
			logger.Debugf("Credentials registration is disabled.")
			Respond(ctx, true, "Credentials registration is disabled.", responseBody{})
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if received email address is plausible
		if !utils.IsPlausibleEmail(req.Email) {
			logger.Debugf("Invalid e-mail address.")
			Respond(ctx, true, "Invalid e-mail address.", responseBody{})
			return
		}

		// Log login attempt
		logger.Debugf("Credentials user '%s' trying to register.", req.Email)

		// Check whether this is the right authenticator for this user
		errAuthenticator := authenticatorAllowed(req.Email, a)
		if errAuthenticator != nil {
			logger.Debugf("Credentials authenticator not responsible for '%s'.", req.Email)
			Respond(ctx, true, "Invalid registration method.", responseBody{})
			return
		}

		// Query for user with given email address
		userEntry, errUserEntry := database.GetUserByMail(req.Email)
		if errUserEntry != nil {
			logger.Errorf("Could not query existing users: %s", errUserEntry)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user with same e-mail address already exists
		if userEntry != nil {
			logger.Debugf("E-mail address already registered.")
			Respond(ctx, true, "E-mail address already registered.", responseBody{})
			return
		}

		// Check password strength
		if !utils.ValidPassword(
			req.Password,
			passMinLength,
			true,
			true,
			true,
			true,
		) {
			logger.Debugf("Insufficient password complexity.")
			Respond(ctx, true, "Insufficient password complexity.", responseBody{})
			return
		}

		// Calculate password hash
		passwordHash, errHash := utils.CreatePasswordHash(req.Password)
		if errHash != nil {
			logger.Errorf("Could not calculate password hash: %s", errHash)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		userEmail := req.Email
		name := ""
		if req.Name != nil {
			name = *req.Name
		}
		surname := ""
		if req.Surname != nil {
			surname = *req.Surname
		}

		// Generate new credentials user
		user := database.NewUser(userEmail, userEmail, "", name, surname)
		user.Password = sql.NullString{
			String: passwordHash,
			Valid:  true,
		}

		// Get appropriate loader
		loader := GetLoader(logger, user.Email)

		// Update user
		if loader != nil {

			// ATTENTION: RefreshUser might update user attributes, but does not yet commit them!
			errTemporary, errInternal, errPublic := loader.RefreshUser(logger, user)

			// Abort if there was an error and return response based on error kind
			if len(errPublic) > 0 {
				if errInternal != nil {
					logger.Debugf("Invalid registration data for user '%s': %s", req.Email, errInternal)
				} else {
					logger.Debugf(
						"Invalid registration data for user '%s': %s",
						req.Email,
						strings.ToLower(strings.Trim(errPublic, ".")),
					)
				}
				Respond(ctx, true, errPublic, responseBody{})
				return
			} else if errTemporary != nil {
				logger.Warningf("Could not update user data for user '%s': %s", req.Email, errTemporary)
				RespondTemporaryError(ctx)
				return
			} else if errInternal != nil {
				logger.Errorf("Could not load user data for user '%s': %s", req.Email, errInternal)
				RespondInternalError(ctx) // Return generic error information
				return
			}
		}

		// Create new user
		errCreate := user.Create()
		if errCreate != nil {
			logger.Errorf("Could not create new user '%s': %s", userEmail, errCreate)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Return response
		logger.Debugf("User created.")
		Respond(ctx, false, "User created.", responseBody{})
	}
}

// login takes the email address and password from the request, validates them and return a valid
// access token required to authenticate the user on subsequent REST requests.
func (a *AuthenticatorCredentials) login() gin.HandlerFunc {

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Define expected response structure
	type responseBody struct{} // Access tokens will be automatically appended on success

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Make sure Email address is set, which is the primary user identifier
		if req.Email == "" {
			logger.Debugf("No login e-mail address supplied.")
			Respond(ctx, true, "E-mail address required.", responseBody{})
			return
		}

		// Make sure Email address is set, which is the primary user identifier
		if req.Password == "" {
			logger.Debugf("No login password supplied.")
			Respond(ctx, true, "Password required.", responseBody{})
			return
		}

		// Check if received email address is plausible
		if !utils.IsPlausibleEmail(req.Email) {
			logger.Warningf("Could not authenticate invalid email address '%s'.", req.Email)
			RespondAuthError(ctx)
			return
		}

		// Log login attempt
		logger.Debugf("Credentials user '%s' trying to authenticate.", req.Email)

		// Check whether this is the right authenticator for this user
		errAuthenticator := authenticatorAllowed(req.Email, a)
		if errAuthenticator != nil {
			logger.Debugf("Credentials authenticator not responsible for '%s'.", req.Email)
			Respond(ctx, true, "Invalid user.", responseBody{})
			return
		}

		// Query for user with given email address
		userEmail := req.Email
		userEntry, errUserEntry := database.GetUserByMail(userEmail)
		if errUserEntry != nil {
			logger.Errorf("Could not query existing users: %s", errUserEntry)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user exists
		if userEntry == nil {
			logger.Debugf("Unknown login e-mail address.")
			Respond(ctx, true, "Invalid user.", responseBody{})
			return
		}

		// Check if user can authenticate via password
		if len(userEntry.Password.String) == 0 || len(userEntry.SsoId.String) != 0 {
			logger.Debugf("User not a credentials user.")
			Respond(ctx, true, "Invalid user.", responseBody{})
			return
		}

		// Validate user password
		success := utils.CheckPasswordHash(userEntry.Password.String, req.Password)
		if success != nil {
			logger.Debugf("Authentication failed with invalid credentials.")
			Respond(ctx, true, "Invalid user.", responseBody{})
			return
		}

		// Get appropriate loader
		loader := GetLoader(logger, userEntry.Email)

		// Update user
		if loader != nil {

			// ATTENTION: RefreshUser might update user attributes, but does not yet commit them!
			errTemporary, errInternal, errPublic := loader.RefreshUser(logger, userEntry)

			// Abort if there was an error and return response based on error kind
			if len(errPublic) > 0 {
				if errInternal != nil {
					logger.Errorf("Unexpected invalid user data for user '%s': %s", req.Email, errInternal)
				} else {
					logger.Errorf(
						"Unexpected invalid user data for user '%s': %s",
						req.Email,
						strings.ToLower(strings.Trim(errPublic, ".")),
					)
				}
				Respond(ctx, true, errPublic, responseBody{})
				return
			} else if errTemporary != nil {
				logger.Warningf("Could not update user data for user '%s': %s", req.Email, errTemporary)
				// Let existing user pass based on cached data - might need urgent access to turn something off
			} else if errInternal != nil {
				logger.Errorf("Could not load user data for user '%s': %s", req.Email, errInternal)
				// Let existing user pass based on cached data - might need urgent access to turn something off
			}
		}

		// Do some important login checks, update login-related user attributes and write any changes
		// ATTENTION: Commits all changed user attributes!
		errUser, errLogin := doLogin(logger, userEntry, nil, ctx.Request.Host)
		if errUser != nil {
			logger.Debugf("Login not allowed for user '%s': '%s'.", userEmail, errUser)
			Respond(ctx, true, "Invalid user.", responseBody{})
			return

		} else if errLogin != nil {
			logger.Errorf("Could not login user '%s': %s", userEmail, errLogin)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Log successful login
		logger.Debugf("User '%s' logged in.", userEntry.Email)

		// Update request context with current user. It will be used by "Respond()" to generate and attach a fresh
		// access token for this user to the response.
		SetContextStorage(ctx, &ContextStorage{
			Logger:      logger,
			CurrentUser: userEntry,
		})

		// Return response
		logger.Debugf("Authentication successful.")
		Respond(ctx, false, "Authentication successful.", responseBody{})
	}
}
