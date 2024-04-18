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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"net"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

// MwInitializeRequest initializes a request by initializing a request-specific tagged logger, logging an initial
// message, preparing a request-context bound storage for common data and logging a final message after the request
// has been handled.
func MwInitializeRequest() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Remember time to log response time
		start := time.Now()

		// Generate UUID for context
		uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

		// Prepare request-specific tagged logger
		logger := log.GetLogger().Tagged(fmt.Sprintf("%s", uuid))

		// Log received request
		logger.Debugf("%s %s from %s.",
			context.Request.Method,
			context.Request.URL.String(),
			context.Request.RemoteAddr,
		)

		// Log middleware execution
		logger.Debugf("Preparing request storage.")

		// Initialize and attach request storage to request context
		SetContextStorage(context, &ContextStorage{
			Logger:      logger,
			CurrentUser: nil,
		})

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!

		// Log request completion
		logger.Debugf("Response sent (%s).", time.Since(start))
	}
}

func MwPanicHandler() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(context)

		// Log middleware execution
		logger.Debugf("Checking for request panics.")

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {

				// Check for a broken connection, as it is not really a condition that warrants a panic stack trace.
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							logger.Debugf("Connection closed unexpectedly.")
							context.Abort()
							return
						}
					}
				}

				// Get current request headers, censoring the sensitive authentication header
				httpRequest, _ := httputil.DumpRequest(context.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				// Put request string back together
				req := strings.Join(headers, "\n")
				req = strings.Trim(req, "\n") // Trim trailing double newline of requests
				req = strings.Replace(req, "\n", "\n\t\t| ", -1)

				// Log panic including associated request
				logger.Errorf("HTTP request panicked!\n\tRequest:\n\t\t| %s\n\tError:\n\t\t| %s", req, err)
				logger.Debugf("Aborting request.")

				// Abort further request execution and return error to client
				RespondInternalError(context)
			}
		}()

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!
	}
}

// MwSetCorsHeaders adds CORS-relevant response headers to every response. Headers can (but should not) be
// overwritten by actual handler function, if processed subsequently.
func MwSetCorsHeaders() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(context)

		// Log middleware execution
		logger.Debugf("Preparing CORS headers.")

		// Set REST relevant response headers
		context.Writer.Header().Set("Access-Control-Max-Age", "86400")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		context.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!
	}
}

// MwSetSecurityHeaders adds security-relevant response headers to every response. Headers can (but should not) be
// overwritten by actual handler function, if processed subsequently.
func MwSetSecurityHeaders() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(context)

		// Log middleware execution
		logger.Debugf("Preparing security headers.")

		// Get currently set response headers
		headers := context.Writer.Header()

		// Set "No Sniff" policy
		headers.Set("X-Content-Type-Options", "nosniff")

		// Set "DNS Prefetch Control" policy
		headers.Set("X-DNS-Prefetch-Control", "off")

		// Set frame guard
		headers.Set("X-Frame-Options", "DENY")

		// Set HSTS policy
		if _build.DevMode {
			// Strict transport security won't work with development certificate,
			// but just generate lots of warnings in the browser's developer bar.
		} else {
			o := 5184000
			op := "max-age=" + strconv.Itoa(o)
			op += "; includeSubDomains"
			headers.Set("Strict-Transport-Security", op)
		}

		// Set "IE No Open" policy
		headers.Set("X-Download-Options", "noopen")

		// Set XSS filter header
		headers.Set("X-XSS-Protection", "1; mode=block")

		// Set referrer policy
		headers.Set("Referrer-Policy", "no-referrer")

		// Set cache control policies
		headers.Set("Surrogate-Control", "no-store")
		headers.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")

		// Set content security policy
		policy := ""
		policy += fmt.Sprintf("%s %s; ", "default-src", "'self'")
		policy += fmt.Sprintf("%s %s; ", "script-src", "'self' 'unsafe-eval'")                               // 'unsafe-eval' required by KnockoutJS framework :(
		policy += fmt.Sprintf("%s %s; ", "font-src", "'self' data: https://fonts.gstatic.com")               // 'https://fonts.gstatic.com' required by SemanticUi
		policy += fmt.Sprintf("%s %s; ", "img-src", "'self' data:")                                          // 'data:' required by SemanticUi
		policy += fmt.Sprintf("%s %s; ", "style-src", "'self' 'unsafe-inline' https://fonts.googleapis.com") // 'https://fonts.googleapis.com' required by SemanticUi.
		policy = strings.TrimSuffix(policy, "; ")
		headers.Set("X-Webkit-CSP", policy)
		headers.Set("X-Content-Security-Policy", policy)
		headers.Set("Content-Security-Policy", policy)

		// Set custom security relevant response headers
		headers.Set("Server", "Apache") // Haha

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!
	}
}

// MwAbortOptionsRequest aborts OPTIONS requests, as they should not get executed in the backend
func MwAbortOptionsRequest() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(context)

		// Log middleware execution
		logger.Debugf("Checking for OPTIONS request.")

		// Tell clients that there is no OPTIONS request available
		context.Writer.Header().Set("access-control-allow-methods", "POST, GET")

		// Abort OPTIONS requests
		if context.Request.Method == "OPTIONS" {
			logger.Debugf("Aborted OPTIONS request.")
			context.AbortWithStatus(200)
			return
		}

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!
	}
}

// MwJwtAuthentication is a middleware validating every request for valid authentication details
func MwJwtAuthentication() gin.HandlerFunc {
	return func(context *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(context)

		// Log middleware execution
		logger.Debugf("Executing authentication middleware.")

		// Stop validation if route does not require authentication
		if !strings.HasPrefix(context.Request.URL.Path, "/api/") {
			logger.Debugf("Static file accessible without authentication.")
			context.Next() // Success, step to next middleware plugin, if existing
			return
		} else if scanUtils.StrContained(context.Request.URL.Path, apiEndpointsNoAuth) {
			logger.Debugf("API endpoint accessible without authentication.")
			context.Next() // Success, step to next middleware plugin, if existing
			return
		} else {
			logger.Debugf("API endpoint requires authentication.")
		}

		// Retrieve the authorization header from the request
		tokenHeader := context.Request.Header.Get("Authorization")

		// Stop validation if authorization header is missing
		if tokenHeader == "" {
			logger.Debugf("Authentication failed. Authentication token is missing.")
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		// Stop validation if token format does not meet expectation (`Bearer {token-body}`)
		splits := strings.Split(tokenHeader, " ")
		if len(splits) != 2 {
			logger.Debugf("Authentication failed. Authentication token is malformed.")
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		// Read and parse token
		assertion := &Jwt{}
		tokenPart := splits[1] // Grab the token part, what we are truly interested in
		token, err := jwt.ParseWithClaims(tokenPart, assertion, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		// Stop validation if token could not be parsed
		if err != nil {
			logger.Debugf("Authentication failed. Authentication token exceeded: %s", err)
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		// Stop validation if token is invalid. Maybe not signed on this server.
		if !token.Valid {
			logger.Debugf("Authentication failed. Authentication token is invalid.")
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		// Make sure there is a valid SAML ID associated with the token
		if assertion.UserId <= 0 {
			logger.Warningf("Authentication failed. Authentication token not assignable.")
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		userEntry, errUserEntry := database.GetUser(assertion.UserId)
		if errUserEntry != nil {
			logger.Warningf("Authentication failed. Could not query user: %s", errUserEntry)
			RespondInternalError(context) // Return generic error information
			return
		}

		// Stop if user is invalid
		if userEntry == nil {
			logger.Warningf("Authentication failed. User with ID '%d' is not existing.", assertion.UserId)
			RespondAuthError(context)
			return
		}

		// Abort login if JWT revision does not match user's logout count (revision is increased on each logout)
		if assertion.Revision != userEntry.LogoutCount {
			logger.Warningf("Authentication failed. Token revision invalid.")
			RespondAuthError(context)
			return // Return, without processing further middleware plugins
		}

		// Log authentication
		logger.Debugf("Authenticated as '%s'.", userEntry.Email)

		// Stop if user is disabled
		if !userEntry.Active {
			logger.Debugf("Authorization failed. User is disabled.")
			RespondAuthError(context)
			return
		}

		// Validate admin rights, if necessary
		if userEntry.Admin {
			logger.Debugf("Authorization successful as role administrator.")
		} else {
			if scanUtils.StrContained(context.Request.URL.Path, apiEndpointsAdmin) {
				logger.Debugf("Authorization failed. Admin privileges required.")
				RespondAuthError(context)
				return
			} else {
				logger.Debugf("Authentication successful as user.")
			}
		}

		// Set user in current request context on server side. Handlers can use this data subsequently.
		getContextStorage(context).CurrentUser = userEntry

		// Success, step to next middleware plugin, if existing
		context.Next() // Should only be called in middleware handlers!
	}
}
