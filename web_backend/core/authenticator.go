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
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	scanUtils "github.com/siemens/GoScans/utils"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"strings"
	"time"
)

var authenticators []Authenticator                // List of registered authenticators to be initialized
var authenticatorsLookup map[string]Authenticator // Map of initialized authentication module, referenced by user domain for quick lookup

type Authenticator interface {
	Domains() []string
	Init(conf map[string]interface{}) error
	EntryUrl() string
}

// EntryUrl checks whether there is a certain authentication entry URL the user should be redirected to,
// e.g. a page initializing SSO with a third party authentication provider.
func EntryUrl(authenticatingEmail string) string {

	// Extract user domain from e-mail address
	userDomain := authenticatingEmail[strings.LastIndex(authenticatingEmail, "@"):]

	// Get authenticator responsible for user domain in order to read associated redirect URL, if set
	authenticator, hasAuthenticator := authenticatorsLookup[userDomain]

	// Get default authenticator
	if !hasAuthenticator {
		authenticator, _ = authenticatorsLookup[""]
	}

	// Return authenticator's redirect URL if set
	return authenticator.EntryUrl()
}

// initAuthenticators initializes registered authenticators and adds them to the lookup table
func initAuthenticators(confAuth map[string]interface{}) error {

	// Initialize map of authenticators.
	// After initialization, all user domains will be start with @, for security reasons
	authenticatorsLookup = make(map[string]Authenticator, len(authenticators))

	// Initialize registered authenticators
	for _, authenticator := range authenticators {

		// Initialize as default authenticator, if it isn't configured for specific domains
		if len(authenticator.Domains()) == 0 {

			// Check if default authenticator is already registered
			_, authenticatorExists := authenticatorsLookup[""]
			if authenticatorExists {
				return fmt.Errorf("multiple default authenticators configured")
			}

			// Register authenticator as the default authenticator
			authenticatorsLookup[""] = authenticator
		} else {

			// Add reference to authenticators lookup map for each of the authenticator's responsible user domains
			for _, authenticatorDomains := range authenticator.Domains() {

				// Abort if empty value (reserved for default authenticator) was specified among user domains
				if len(authenticatorDomains) == 0 {
					return fmt.Errorf("default authenticator cannot be registerd along with other user domains")
				}

				// Check if domain is formatted correctly
				if !strings.HasPrefix(authenticatorDomains, "@") {
					return fmt.Errorf("authenticator user domains must start with @")
				}

				// Make sure there wasn't another @ contained
				if strings.Count(authenticatorDomains, "@") != 1 {
					return fmt.Errorf(
						"invalid authenticator user domain '%s'", authenticatorDomains)
				}

				// Check if authenticator for user domain is already registered
				_, authenticatorExists := authenticatorsLookup[authenticatorDomains]
				if authenticatorExists {
					return fmt.Errorf("multiple '%s' authenticators configured", authenticatorDomains)
				}

				// Register user domain with reference to this authenticator
				authenticatorsLookup[authenticatorDomains] = authenticator
			}
		}

		// Initialize authenticator
		errAuthenticator := authenticator.Init(confAuth)
		if errAuthenticator != nil {
			return fmt.Errorf(
				"could not initialize authenticator for '%s': %s",
				strings.Join(authenticator.Domains(), ", "),
				errAuthenticator,
			)
		}
	}

	// Verify that there is a default authenticator
	if _, ok := authenticatorsLookup[""]; !ok {
		return fmt.Errorf("no default authenticator registered")
	}

	// Return nil as everything went fine
	return nil
}

// authenticatorAllowed checks whether given user (e-mail address) is allowed to use a certain authenticator.
// Checks whether:
//   - user is allowed to use this authenticator
//   - user has a better suitable authenticator
func authenticatorAllowed(userEmail string, authenticator Authenticator) error {

	// Extract user domain from e-mail address
	userDomain := userEmail[strings.LastIndex(userEmail, "@"):]

	// Check if user uses its dedicated authenticator
	dedicatedAuthenticator, hasDedicated := authenticatorsLookup[userDomain]
	if hasDedicated && authenticator != dedicatedAuthenticator {
		return fmt.Errorf("authenticator not allowed for '%s'", userEmail)
	}

	// Don't allow other than the default authenticator, if there is no dedicated authenticator
	defaultAuthenticator, _ := authenticatorsLookup[""]
	if !hasDedicated && authenticator != defaultAuthenticator {
		return fmt.Errorf("authenticator not responsible for '%s'", userEmail)
	}

	// Return nil if user is allowed to use authenticator
	return nil
}

// doLogin increments some login attributes related to the user. Furthermore, it commits any changes on the
// user struct.
// Returns userError, if something is wrong with the user.
// Returns internalError, if something else went wrong.
func doLogin(
	logger scanUtils.Logger,
	user *database.T_user,
	allowedDepartments []string,
	vhost string,
) (errUser, errInternal error) {
	// Do some plausibility check
	if user == nil || user.Email == "" || user.Id == 0 || user.Company == "" || user.Created.IsZero() {
		return nil, fmt.Errorf("invalid user struct")
	}

	// Update attributes
	user.LastLogin = time.Now()

	// Check if user is member of an allowed company department
	if len(allowedDepartments) > 0 {
		user.Demo = true
		for _, dep := range allowedDepartments {
			dep = strings.ToUpper(dep)
			depUser := strings.ToUpper(user.Department)
			if depUser == dep || strings.HasPrefix(depUser, dep+" ") {
				user.Demo = false
			}
		}
	}

	// Save updated attributes. Additional attributes outside of this function may have been updated by a loader
	// plugin, so save should update all user attributes, except the sensitive/internal ones (ID, e-mail, password,
	// admin, db password). A loader may also have set a user to inactive.
	_, errSave := user.Save(
		"ssoid", "company", "department", "last_login", "active", "name", "surname", "gender", "demo", "certificate")
	if errSave != nil {
		return nil, errSave
	}

	// Check if user is set active
	if !user.Active {

		// Disable user on databases if it turned inactive
		errToggle := manager.RpcDisableDbUser(logger, rpcClient, user.Email)
		if errToggle != nil {
			logger.Warningf("Could not disable DB user '%s' during login.", user.Email)
		}

		// Return error because user is invalid
		return fmt.Errorf("user is disabled"), nil
	}

	// Log event
	errEvent := database.NewEvent(user, database.EventLogin, vhost)
	if errEvent != nil {
		logger.Errorf("Could not create event log: %s", errEvent)
		return nil, errEvent
	}

	// Return no errors, indicating user may pass
	return nil, nil
}

// createJwt generates a JWT access token for a given user
func createJwt(userId uint64, revision uint) (string, time.Time) {

	// Get config
	conf := config.GetConfig()

	// Define token expiry
	expires := time.Now().Add(conf.Jwt.Expiry)

	// Prepare JWT token
	tk := &Jwt{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()), // time necessary to generate varying tokens
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
		UserId:   userId,
		Revision: revision, // Tag is not a secret
	}

	// Sign JWT token
	// ATTENTION: The signed token is NOT encrypted, it's just encoded and whatever is contained can be
	//            recovered/read by the client!
	token := jwt.NewWithClaims(jwt.GetSigningMethod(jwtAlgorithm), tk)
	tokenSigned, _ := token.SignedString([]byte(jwtSecret))

	// Return generated token
	return tokenSigned, expires
}
