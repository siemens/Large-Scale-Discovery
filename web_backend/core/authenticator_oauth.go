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
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"strings"
	"time"
)

// init automatically registers the authenticator implemented in this file. If you don't want this authenticator,
// just remove it. You can also add your own authenticator by adding a file with your dedicated implementation.
func init() {
	if _build.DevMode {
	} else {

		// Define authenticator, it will be initialized later by core
		a := NewAuthenticatorOauth("oauth", nil)

		// Register authenticator for initialization
		authenticators = append(authenticators, a)
	}
}

//
// This SSO authenticator is following the OAUTH "Authorization Code Grant/Flow"!
//

/*

--- Terminologies ----

Resource Owner: 		You are the owner of your identity, your data, and any actions that can be performed with your accounts.
Client: 				The application that wants to access data or perform actions on behalf of the "Resource Owner".
Authorization Server: 	The application that knows the "Resource Owner", where the "Resource Owner" already has an account.
Resource Server: 		The Application Programming Interface (API) or service the "Client" wants to use on behalf of the "Resource Owner" (This application).
Redirect URI: 			The URL the "Authorization Server" will redirect the "Resource Owner" back to after granting permission to the "Client".
						This is sometimes referred to as the “Callback URL.”
Response Type: 			The type of information the "Client" expects to receive. The most common "Response Type" is code,
						where the "Client" expects an "Authorization Code".
Scope: 					These are the granular permissions the "Client" wants, such as access to data or to perform actions.
Consent: 				The "Authorization Server" takes the "Scopes" the "Client" is requesting, and verifies with the "Resource Owner" whether or not they want to give the "Client" permission
Client ID: 				This ID is used to identify the "Client" on the "Authorization Server".
Client Secret: 			This is a secret password that only the "Client" and "Authorization Server" know.
						This allows them to securely share information privately behind the scenes.
Authorization Code: 	A short-lived temporary code the "Client" gives the "Authorization Server" in exchange for an "Access Token".
Access Token: 			The access key or session token the "Client" will use to communicate with the "Resource Server".
						This is like a badge or key card that gives the "Client" permission to request data or perform actions with the "Resource Server" on your behalf.
						In this application's context, the valid "Access token" is used to obtain an application-specific valid JWT "Session Token".
Session Token:			The JWT "Session Token" is not part of the oauth protocol, but the token generated by this application in order to authenticate subsequent requests.


---- Authorization Code Grant Workflow Description ----

The Authorization Code Grant, or Code Flow, is the most widely used oauth flow and also used in this package. To obtain
a token using code flow, the "Client sends an authorization request to the oauth server by simply redirecting the
browser to the server. The oauth server makes sure that the user is authenticated. When the user approves, a
short-lived code is issued to the "Client. This code can be considered a one time password, or a nonce. The "Client
receives this code, and can now use it in an authenticated backend call and exchange it for an application
"Session Token".

One thing to mention here is that the user only will enter its credentials at the oauth server. The user won’t have to
give the credentials to the app, it simply enters them to the server it already knows and trusts. This is one thing
that oauth set out to solve.

An additional benefit is that the token owner passes the browser, which makes it harder to steal, and since the call to
exchange the token is authenticated, the server can be sure that it delivers the token to the correct "Client.

Usually, the code flow will also allow you to receive a "Refresh Token", which allows the "Client to get new
"Access Tokens" without involving the user. Even after the "Access Token" is expired. However, the refresh token is not
used as part of this package. Instead, once the user (user-agent) is authenticated, the "Access Token" is dismissed and
exchanged for an application-specific JWT "Session Token". Validity and refresh of it is managed by the application
itself.


---- WORKFLOW ----

1. The "Resource Owner" (user/client/browser), wants to authenticate against the backend server.
2. The backend redirects the browser to the "Authorization Server" and includes with the request the "Client ID"
and "Redirect URI".
3. The "Authorization Server" verifies who you are, and if necessary prompts for a login.
4. The "Authorization Server" redirects the user-agent back to backend server using the "Redirect URI" along with an "Authorization Code".
5. The backend server contacts the "Authorization Server" directly and sends its "Client ID", "Client Secret" and the "Authorization Code", which was received from the user (user-agent).
6. The "Authorization Server" verifies the data and responds with an "Access Token".
7. The backend server checks the "Access Token" validity by using the "Authorization Server" certificate key.
8. The backend server generates an application-specific JWT and stores it in the user's browser session storage. This JWT "Session Token" is used for subsequent authentication.
9. The user is redirected to the LSD root page and is now able to login with the JWT "Session Token" held within the browser's session storage.


---- OPEN ID CONNECT  ----

As part of the Authorization Code Grant/Flow we are using the OPEN ID CONNECT protocol.

Oauth 2.0 is designed only for authorization, for granting access to data and features from one application to another.
OpenID Connect (OIDC) is a thin layer that sits on top of oauth 2.0 that adds login and profile information
about the person who is logged in. Establishing a login session is often referred to as authentication, and information
about the person logged in is called identity. When an "Authorization Server" supports OIDC, it is sometimes called an
identity provider, since it provides information about the "Resource Owner" (user).

OpenID Connect enables scenarios where one login can be used across multiple applications, also known as single
sign-on (SSO).


---- Java Web Token ----

An ID Token is a specifically formatted string of characters known as a JSON Web Token, or JWT.
A JWT may look like gibberish, but the "Client can extract information embedded in the JWT such as a user ID,
name, log-in time, expiration,... The data inside the ID Token are called claims.

With OIDC, there’s also a standard way the "Client can request additional identity information from the
"Authorization Server", such as their email address, using the "Access Token".

*/

// oauthStateName is used in the communication between the user-agent and Authentication Server. It is used to identify
// the authentication session in step 2 of the authentication process for a given user.
const oauthStateName = "oauth"

// oauthStateExpiry is the expiration time of the validity of the authentication token.
// It can happen that there is a significant delay between during authentication with the Authentication Server and
// the user - for instance if the user has to find his credentials. The validity duration of the token does not have
// a significant impact on the security of the system.
const oauthStateExpiry = 10 * time.Minute

// openIDConfig contains all endpoint configuration items required for the proper communication between the
// "Authorization Server" and the backend server. It is used to unmarshal the configuration request received from the
// configuration endpoint of "Authorization Server". In case the "Authorization Server" changes any endpoint
// configuration, the backend server would automatically reconfigure at restart. In order to ensure proper
// configuration of the "Authorization Server" endpoints, the oauth_config_url configuration variable
// must be valid in the backend configuration file
type openIDConfig struct {
	Issuer                                                    string   `json:"issuer"`
	AuthorizationEndpoint                                     string   `json:"authorization_endpoint"`
	TokenEndpoint                                             string   `json:"token_endpoint"`
	RevocationEndpoint                                        string   `json:"revocation_endpoint"`
	UserinfoEndpoint                                          string   `json:"userinfo_endpoint"`
	IntrospectionEndpoint                                     string   `json:"introspection_endpoint"`
	JwksURI                                                   string   `json:"jwks_uri"`
	RegistrationEndpoint                                      string   `json:"registration_endpoint"`
	PingRevokedSrisEndpoint                                   string   `json:"ping_revoked_sris_endpoint"`
	PingEndSessionEndpoint                                    string   `json:"ping_end_session_endpoint"`
	DeviceAuthorizationEndpoint                               string   `json:"device_authorization_endpoint"`
	ScopesSupported                                           []string `json:"scopes_supported"`
	ClaimsSupported                                           []string `json:"claims_supported"`
	ResponseTypesSupported                                    []string `json:"response_types_supported"`
	ResponseModesSupported                                    []string `json:"response_modes_supported"`
	GrantTypesSupported                                       []string `json:"grant_types_supported"`
	SubjectTypesSupported                                     []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported                          []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported                         []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported                []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	ClaimTypesSupported                                       []string `json:"claim_types_supported"`
	ClaimsParameterSupported                                  bool     `json:"claims_parameter_supported"`
	RequestParameterSupported                                 bool     `json:"request_parameter_supported"`
	RequestURIParameterSupported                              bool     `json:"request_uri_parameter_supported"`
	RequestObjectSigningAlgValuesSupported                    []string `json:"request_object_signing_alg_values_supported"`
	IDTokenEncryptionAlgValuesSupported                       []string `json:"id_token_encryption_alg_values_supported"`
	IDTokenEncryptionEncValuesSupported                       []string `json:"id_token_encryption_enc_values_supported"`
	BackchannelAuthenticationEndpoint                         string   `json:"backchannel_authentication_endpoint"`
	BackchannelTokenDeliveryModesSupported                    []string `json:"backchannel_token_delivery_modes_supported"`
	BackchannelAuthenticationRequestSigningAlgValuesSupported []string `json:"backchannel_authentication_request_signing_alg_values_supported"`
	BackchannelUserCodeParameterSupported                     bool     `json:"backchannel_user_code_parameter_supported"`
}

type AuthenticatorOauth struct {
	name                 string
	domains              []string
	entryUrl             string
	allowedDepartments   []string
	oauthConfig          *oauth2.Config
	oauthGetLatestKeySet func() (jwk.Set, error)
	oauthClaimsMapping   map[string]string
}

// NewAuthenticatorOauth generates a new authenticator with the user domains it is responsible for. Everything
// else of the authenticator will be initialized later during core initialization, with the actual config values.
func NewAuthenticatorOauth(name string, domains []string) *AuthenticatorOauth {
	return &AuthenticatorOauth{
		name:    name,
		domains: domains, // Store domains including @ prefix
	}
}

// Domains returns the user domains this authenticator got registered for
func (a *AuthenticatorOauth) Domains() []string {
	return a.domains
}

// Init initializes necessary routes and handlers to authenticate users. The argument "domain" restricts the authority
// of this authenticator to a specific user domain (if set), in order to:
//   - prevent impersonation of users by a rogue authentication provider.
//   - prevent authentication downgrades (e.g., OAUTH users using weaker authenticator).
//
// The following routes are deployed, where * is the name of the authenticator:
//   - /v1/auth/*/redirect
//   - /v1/auth/*/callback
func (a *AuthenticatorOauth) Init(conf map[string]interface{}) error {

	// Check if config exists
	if conf == nil {
		return fmt.Errorf("authenticator configuration missing")
	}

	// Extract authenticator attributes for this authenticator by name
	authConf, okAuthConf := conf[a.name]
	if !okAuthConf {
		return fmt.Errorf("authenticator configuration empty")
	}
	conf = authConf.(map[string]interface{})

	// Check whether required arguments are available and valid
	val, okVal := conf["departments"]
	valDepartments, okValDepartments := val.([]interface{})
	if okVal && !okValDepartments {
		return fmt.Errorf("invalid list of whitelisted departments")
	}
	allowedDepartments := make([]string, 0, len(valDepartments))
	for _, dep := range valDepartments {
		cVal, okCVal := dep.(string)
		if !okCVal || len(cVal) == 0 {
			return fmt.Errorf("invalid department whitelist value")
		}
		allowedDepartments = append(allowedDepartments, cVal)
	}
	val, okVal = conf["oauth_config_url"]
	oauthConfUrl, okOauthConfUrl := val.(string)
	if !okVal || !okOauthConfUrl || !strings.HasPrefix(strings.ToLower(oauthConfUrl), "https://") {
		return fmt.Errorf("invalid oauth config URL")
	}
	val, okVal = conf["oauth_client_id"]
	oauthClientId, okOauthClientId := val.(string)
	if !okVal || !okOauthClientId || len(oauthClientId) < 3 {
		return fmt.Errorf("invalid oauth client ID")
	}
	val, okVal = conf["oauth_client_secret"]
	oauthClientSecret, okOauthClientSecret := val.(string)
	if !okVal || !okOauthClientSecret || len(oauthClientSecret) < 3 {
		return fmt.Errorf("invalid oauth client secret")
	}
	val, okVal = conf["oauth_scope"]
	oauthScope, okOauthScope := val.(string)
	if !okVal || !okOauthScope {
		return fmt.Errorf("invalid oauth scope")
	}
	val, okVal = conf["oauth_public_key_url"]
	oauthPublicKeyUrl, okOauthPublicKeyUrl := val.(string)
	if !okVal || !okOauthPublicKeyUrl || !strings.HasPrefix(strings.ToLower(oauthPublicKeyUrl), "https://") {
		return fmt.Errorf("invalid oauth public key URL")
	}

	// Verify oauth claims mapping is plausible
	val, okVal = conf["oauth_claims_mapping"]
	oauthClaimsMapping, okOauthClaimsMapping := val.(map[string]interface{})
	if !okVal || !okOauthClaimsMapping {
		return fmt.Errorf("missing oauth claims mapping")
	}
	oauthClaims := make(map[string]string, len(oauthClaimsMapping))
	for _, claim := range []string{"user_mail", "user_name", "user_surname", "user_company", "user_department"} {
		c, okC := oauthClaimsMapping[claim]
		cVal, okCVal := c.(string)
		if claim == "user_mail" && (!okC || !okCVal || len(cVal) == 0) {
			return fmt.Errorf("missing oauth claims mapping for '%s'", claim)
		}
		oauthClaims[claim] = cVal
	}

	// Prepare memory for oauth config struct
	confOpenID := openIDConfig{}

	// Read oauth config from oauth config URL
	confJson, errGet := http.Get(oauthConfUrl)
	if errGet != nil {
		return fmt.Errorf("could not request oauth config: %s", errGet)
	}
	marshaled, errRead := io.ReadAll(confJson.Body)
	if errRead != nil {
		return fmt.Errorf("could not request oauth config: %s", errRead)
	}
	errUnmarshal := json.Unmarshal(marshaled, &confOpenID)
	if errUnmarshal != nil {
		return fmt.Errorf("could not request oauth config: %s", errUnmarshal)
	}

	// Initialize oauth config
	oauthConfig := oauth2.Config{
		ClientID:     oauthClientId,     // Oauth application ID on the authentication server
		ClientSecret: oauthClientSecret, // Client secret
		Endpoint: oauth2.Endpoint{ // Authentication endpoints
			AuthURL:  confOpenID.AuthorizationEndpoint, // Client authentication
			TokenURL: confOpenID.TokenEndpoint,         // Exchange access token for JWT
		},
		RedirectURL: GenerateRelativeUrl("v1", "/auth/"+a.name+"/callback"), // Client is redirected here after authenticating against the authentication server
		Scopes:      []string{oauthScope},
	}

	// Prepare context
	ctx := context.Background()

	// Prepare public key url cache
	c, errC := jwk.NewCache(ctx, httprc.NewClient())
	if errC != nil {
		return fmt.Errorf("could not create cache for public key URL: %s", errC)
	}

	// Register public key url in cache to keep it refreshed automatically in the background
	errReg := c.Register(ctx, oauthPublicKeyUrl)
	if errReg != nil {
		return fmt.Errorf("could not register public key URL: %s", errReg)
	}

	// Execute initial fetch of public keys
	if _, errRefresh := c.Refresh(ctx, oauthPublicKeyUrl); errRefresh != nil {
		return fmt.Errorf("could not fetch public key set from oauth provider: %s", errRefresh)
	}

	// Prepare parametrized get function to retrieve cached and regularly updated key set
	fnGetKeySet := func() (jwk.Set, error) {
		return c.Lookup(ctx, oauthPublicKeyUrl)
	}

	// Prepare redirect URL
	redirectUrl := "/auth/" + a.name + "/redirect"

	// Initialize the authenticator
	a.entryUrl = GenerateRelativeUrl("v1", redirectUrl)
	a.allowedDepartments = allowedDepartments
	a.oauthConfig = &oauthConfig
	a.oauthGetLatestKeySet = fnGetKeySet
	a.oauthClaimsMapping = oauthClaims

	// Register the necessary authenticator endpoints (routes) to handle oauth authentication attempts.
	RegisterApiEndpointNoAuth("v1", "GET", redirectUrl, a.redirect())
	RegisterApiEndpointNoAuth("v1", "GET", "/auth/"+a.name+"/callback", a.login())

	// Return nil as everything went fine
	return nil
}

// EntryUrl returns an URL to redirect to in order to initiate authentication, if required. Returns empty string if
// no redirect is required.
func (a *AuthenticatorOauth) EntryUrl() string {
	return a.entryUrl
}

// redirect redirects the user to the authentication provider
func (a *AuthenticatorOauth) redirect() gin.HandlerFunc {

	// Return request handling function
	return func(ctx *gin.Context) {

		// Prepare memory for state
		stateBytes := make([]byte, 16)

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Create random token for preserving state
		n, errSeed := rand.Read(stateBytes)
		if errSeed != nil || n != len(stateBytes) {
			logger.Errorf("Could not generate random state token: %s", errSeed)
			RespondInternalError(ctx)
			return
		}

		// Convert to string value for inclusion into the cookie
		oauthState := base64.URLEncoding.EncodeToString(stateBytes)

		// Prepare oauth state cookie
		cookie := http.Cookie{
			Name:     oauthStateName,
			Value:    oauthState,
			Expires:  time.Now().Add(oauthStateExpiry),
			HttpOnly: true,
			Secure:   true,
		}

		// Attach HTTP cookie to HTTP response headers
		http.SetCookie(ctx.Writer, &cookie)

		// Prepare absolute callback URI
		callbackUrl := "https://" + ctx.Request.Host + a.oauthConfig.RedirectURL // Prepare dynamically, website might be accessed via different domains

		// Generate the redirect URL. This is the callback URL used in the second step of the flow
		// the endpoint is configured in oauth.json
		consentUrl := a.oauthConfig.AuthCodeURL(
			oauthState,
			oauth2.SetAuthURLParam("redirect_uri", callbackUrl), // .AuthCodeURL() requires absolute redirect URL to generate valid callback
		)

		// Send client to the authentication server
		ctx.Redirect(http.StatusTemporaryRedirect, consentUrl)
	}
}

// login takes the email address and password from the request, validates them and return a valid
// access token required to authenticate the user on subsequent REST requests.
func (a *AuthenticatorOauth) login() gin.HandlerFunc {

	// Prepare unified error message
	authErrRedirect := fmt.Sprintf("/?error=%s", "Unauthorized")
	tempErrRedirect := fmt.Sprintf("/?error=%s", "Temporary")
	internalErrRedirect := fmt.Sprintf("/?error=%s", "Unexpected")

	// Return request handling function with enclosed public key for verification
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Check for authentication request errors
		errReq := ctx.Request.FormValue("error")
		errReqDesc := ctx.Request.FormValue("error_description")
		if errReq != "" {
			logger.Warningf("Authentication request invalid: %s: %s", errReq, errReqDesc)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Get oauth state cookie
		seed, errSeed := ctx.Request.Cookie(oauthStateName)
		if errSeed != nil {
			logger.Debugf("Authentication data incomplete: %s", errSeed)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Verify user-agent carries over the cookie from step 1
		seedFromState := ctx.Request.FormValue("state")
		if seedFromState != seed.Value {
			logger.Debugf("Authentication data with unexpected seed: %s", errSeed)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Get the authentication code
		code := ctx.Request.FormValue("code")
		if code == "" {
			logger.Debugf("Authentication data without code.")
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Prepare absolute callback URI
		callbackUrl := "https://" + ctx.Request.Host + a.oauthConfig.RedirectURL // Prepare dynamically, website might be accessed via different domains

		// Extract the JWT token
		oauthToken, errOauthToken := a.oauthConfig.Exchange(
			context.Background(),
			code,
			oauth2.SetAuthURLParam("redirect_uri", callbackUrl), // .Exchange() requires absolute redirect URL for verification
		)
		if errOauthToken != nil {
			logger.Warningf("Authentication data invalid: %s", errOauthToken) // Bug or attack
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Validate the JWT token
		if !oauthToken.Valid() {
			logger.Warningf("Authentication token invalid.") // Bug or attack
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Get current key set from cache, which is regularly updated
		keySet, errKeySet := a.oauthGetLatestKeySet()
		if errKeySet != nil {
			logger.Errorf("Could not get key set: %s", errKeySet) // Bug or attack
			ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
			return
		}

		// Parse the validated JWT for user specific data
		// If you are using Azure AD and cannot validate the JWT signature:
		// https://stackoverflow.com/questions/70900067/parse-validate-jwt-token-from-azuread-in-golang
		oauthJwtToken, errOauthJwtToken := jwt.Parse(
			[]byte(oauthToken.AccessToken),
			jwt.WithValidate(true), // Validate JWT token attributes like expiry
			jwt.WithKeySet( // Use keys from public key URL to validate JWT token signature
				keySet,
				jws.WithInferAlgorithmFromKey(true)), // Try plausible algorithms, as publik key URL does not describe applicable algorithms
		)
		if errOauthJwtToken != nil {
			logger.Errorf("Could not parse access token: %s", errOauthJwtToken)
			ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
			return
		}

		// Prepare user attributes
		var userEmail string
		var userName string
		var userSurname string
		var userCompany string
		var userDepartment string

		// Log available claims for easier debugging
		var msg []string
		for _, k := range oauthJwtToken.Keys() {
			var errGet error
			switch {
			case k == jwt.AudienceKey:
				var val []string
				errGet = oauthJwtToken.Get(k, &val)
				msg = append(msg, fmt.Sprintf("%s: %s", k, val))
			case k == jwt.ExpirationKey || k == jwt.IssuedAtKey || k == jwt.NotBeforeKey:
				var val time.Time
				errGet = oauthJwtToken.Get(k, &val)
				msg = append(msg, fmt.Sprintf("%s: %s", k, val))
			case k == "amr":
				var val []interface{}
				errGet = oauthJwtToken.Get(k, &val)
				msg = append(msg, fmt.Sprintf("%s: %v", k, val))
			default:
				var val string
				errGet = oauthJwtToken.Get(k, &val)
				msg = append(msg, fmt.Sprintf("%s: %s", k, val))
			}
			if errGet != nil {
				logger.Warningf("Could not log claim '%s': %s", k, errGet)
			}
		}
		logger.Debugf("Available claims: \n\t%s", strings.Join(msg, "\n\t"))

		// Extract desired claim values. User's email is required to authenticate the user.
		// The other data is needed to create a new user.
		errGet := oauthJwtToken.Get(a.oauthClaimsMapping["user_mail"], &userEmail)
		if errGet != nil {
			logger.Errorf("Invalid claim '%s' to extract user mail: %s", a.oauthClaimsMapping["user_mail"], errGet)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		} else if !utils.IsPlausibleEmail(userEmail) {
			logger.Errorf("Invalid value '%s' for claim '%s'.", userEmail, a.oauthClaimsMapping["user_mail"])
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}
		if a.oauthClaimsMapping["user_name"] != "" {
			errGet = oauthJwtToken.Get(a.oauthClaimsMapping["user_name"], &userName)
			if errGet != nil {
				logger.Warningf("Invalid claim '%s' to extract user name: %s", a.oauthClaimsMapping["user_name"], errGet)
			}
		}
		if a.oauthClaimsMapping["user_surname"] != "" {
			errGet = oauthJwtToken.Get(a.oauthClaimsMapping["user_surname"], &userSurname)
			if errGet != nil {
				logger.Warningf("Invalid claim '%s' to extract user surname: %s", a.oauthClaimsMapping["user_surname"], errGet)
			}
		}
		if a.oauthClaimsMapping["user_company"] != "" {
			errGet = oauthJwtToken.Get(a.oauthClaimsMapping["user_company"], &userCompany)
			if errGet != nil {
				logger.Warningf("Invalid claim '%s' to extract user company: %s", a.oauthClaimsMapping["user_company"], errGet)
			}
		}
		if a.oauthClaimsMapping["user_department"] != "" {
			errGet = oauthJwtToken.Get(a.oauthClaimsMapping["user_department"], &userDepartment)
			if errGet != nil {
				logger.Warningf("Invalid claim '%s' to extract user department: %s", a.oauthClaimsMapping["user_department"], errGet)
			}
		}

		// Try to generate fallback company from email address if possible
		if userCompany == "" {

			// Check if domain of user mail is within scope of this authenticator
			// Otherwise, don't select the domain, because it might be something global (gmail.com, outlook.com,...)
			// and we don't want to leak e-mail addresses. Users of the same company can see each other in the web
			// interface!
			domain := strings.SplitN(userEmail, "@", 2)[1]
			if scanUtils.StrContained("@"+domain, a.domains) {
				userCompany = domain
			}
		}

		// Check if received email address is plausible
		if !utils.IsPlausibleEmail(userEmail) {
			logger.Warningf("Could not authenticate invalid email address '%s'.", userEmail)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Log login attempt
		logger.Debugf("Oauth user '%s' trying to authenticate.", userEmail)

		// Check whether this is the right authenticator for this user
		errAuthenticator := authenticatorAllowed(userEmail, a)
		if errAuthenticator != nil {
			logger.Errorf("Oauth %s.", errAuthenticator)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		}

		// Query for user with given email address
		userEntry, errUserEntry := database.GetUserByMail(userEmail)
		if errUserEntry != nil {
			logger.Errorf("Could not query existing users: %s", errUserEntry)
			ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
			return
		}

		// Check if user exists
		var userExisting = userEntry != nil

		// Prepare memory for user
		var user *database.T_user

		// Check if user exists
		if userExisting {

			// Take user from query result
			user = userEntry

			// Check if user can authenticate via oauth
			if len(user.SsoId.String) == 0 || len(user.Password.String) != 0 {
				logger.Debugf("User not an oauth user.")
				ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
				return
			}
		} else {

			// Generate new oauth user
			user = database.NewUser(userEmail, userCompany, userDepartment, userName, userSurname)
			user.SsoId = sql.NullString{ // Set something, just in case the loader subsequently doesn't
				String: userEmail,
				Valid:  true,
			}
		}

		// Get appropriate loader
		loader := GetLoader(logger, user.Email)

		// Update user
		// ATTENTION: RefreshUser might update user attributes, but does not yet commit them!
		errTemporary, errInternal, errPublic := loader.RefreshUser(logger, user)

		// Abort if there was an error and return response based on error kind
		if len(errPublic) > 0 {

			// Decide how to log situation based on whether the user was already registered or is new
			if userExisting {
				if errInternal != nil {
					logger.Errorf("Unexpected invalid user data for user '%s': %s", user.Email, errInternal)
				} else {
					logger.Errorf(
						"Unexpected invalid user data for user '%s': %s",
						user.Email,
						strings.ToLower(strings.Trim(errPublic, ".")),
					)
				}
			} else {
				if errInternal != nil {
					logger.Debugf("Invalid registration data for user '%s': %s", user.Email, errInternal)
				} else {
					logger.Debugf(
						"Invalid registration data for user '%s': %s",
						user.Email,
						strings.ToLower(strings.Trim(errPublic, ".")),
					)
				}
			}

			// Respond with error
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		} else if errTemporary != nil {
			logger.Warningf("Could not update user data for user '%s': %s", user.Email, errTemporary)
			if userExisting {
				// Let existing user pass based on cached data - might need urgent access to turn something off
			} else {
				ctx.Redirect(http.StatusTemporaryRedirect, tempErrRedirect)
				return
			}
		} else if errInternal != nil {
			logger.Errorf("Could not load user data for user '%s': %s", user.Email, errInternal)
			if userExisting {
				// Let existing user pass based on cached data - might need urgent access to turn something off
			} else {
				ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
				return
			}
		}

		// Create new user if it wasn't existing
		if !userExisting {
			errCreate := user.Create()
			if errCreate != nil {
				logger.Errorf("Could not create new user '%s': %s", userEmail, errCreate)
				ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
				return
			}
		}

		// Update login-related user attributes and write changes
		// ATTENTION: Commits all changed user attributes!
		errUser, errLogin := doLogin(logger, user, a.allowedDepartments, ctx.Request.Host)
		if errUser != nil {
			logger.Debugf("Login not allowed for user '%s': '%s'.", userEmail, errUser)
			ctx.Redirect(http.StatusTemporaryRedirect, authErrRedirect)
			return
		} else if errLogin != nil {
			logger.Errorf("Could not login user '%s': %s", userEmail, errLogin)
			ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
			return
		}

		// Tell client to interpret response as HTML
		ctx.Writer.Header().Add("Content-Type", "text/html")

		// Lower content security policy level for this response, as it is necessary to execute some inline JS
		ctx.Writer.Header().Del("Content-Security-Policy")
		ctx.Writer.Header().Add("Content-Security-Policy", "default-src 'unsafe-inline'")

		// Create a JWT token in order to log in the user after authentication
		jwtToken, expires := createJwt(user.Id, user.LogoutCount)
		token := Token{
			AccessToken: jwtToken,
			Expire:      expires,
		}

		// Marshall token struct into JSON string
		tokenJson, errJson := json.Marshal(token)
		if errJson != nil {
			ctx.Redirect(http.StatusTemporaryRedirect, internalErrRedirect)
			return
		}

		// Return JavaScript snippet placing the JWT in the session storage and redirect the user-agent index page.
		js := fmt.Sprintf(`
			<html>
			<script>
				sessionStorage.setItem("token",  JSON.stringify(%s));
				window.location.href = "/";
			</script>
			</html>
		`, tokenJson)

		// Respond with HTML/JavaScript page preparing the client's session store
		ctx.String(200, js)
	}
}
