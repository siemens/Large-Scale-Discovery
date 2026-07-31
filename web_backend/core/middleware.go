/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// MwInitializeRequest initializes a request by initializing a request-specific tagged logger, logging an initial
// message, preparing a request-context bound storage for common data and logging a final message after the request
// has been handled.
func MwInitializeRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Remember time to log response time
		start := time.Now()

		// Generate UUID for context
		uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

		// Prepare request-specific tagged logger
		logger := log.GetLogger().Tagged(uuid)

		// Log received request
		if ctx.RemoteIP() == ctx.ClientIP() {
			logger.Debugf("%s %s from %s.",
				ctx.Request.Method,
				ctx.Request.URL.String(),
				ctx.Request.RemoteAddr,
			)
		} else {
			logger.Debugf("%s %s from %s (Client IP: %s).",
				ctx.Request.Method,
				ctx.Request.URL.String(),
				ctx.Request.RemoteAddr,
				ctx.ClientIP(), // Attention, client IP extracted from HTTP headers might be spoofed
			)
		}

		// Attention, client IP extracted from HTTP headers might be spoofed
		var clientHeaders []string
		xff := ctx.GetHeader("X-Forwarded-For")
		if xff != "" {
			clientHeaders = append(clientHeaders, fmt.Sprintf("X-Forwarded-For=%s", xff))
		}
		xri := ctx.GetHeader("X-Real-IP")
		if xri != "" {
			clientHeaders = append(clientHeaders, fmt.Sprintf("X-Real-IP=%s", xri))
		}
		fwd := ctx.GetHeader("Forwarded")
		if fwd != "" {
			clientHeaders = append(clientHeaders, fmt.Sprintf("Forwarded=%s", fwd))
		}
		if len(clientHeaders) > 0 {
			logger.Debugf("Remote IP headers: %s", strings.Join(clientHeaders, "; "))
		}

		// Log middleware execution
		logger.Debugf("Preparing request storage.")

		// Initialize and attach request storage to request context
		SetContextStorage(ctx, &ContextStorage{
			Logger:      logger,
			CurrentUser: nil,
		})

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!

		// Log request completion
		logger.Debugf("Response sent (%s).", time.Since(start))
	}
}

// MwPanicHandler recovers request panics, logs diagnostic context, and returns a safe error response
func MwPanicHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Log middleware execution
		logger.Debugf("Checking for request panics.")

		// Recover from panics
		defer func() {
			if r := recover(); r != nil {

				// Check for a broken connection, as it is not really a condition that warrants a panic stack trace.
				if ne, ok := r.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							logger.Debugf("Connection closed unexpectedly.")
							ctx.Abort()
							return
						}
					}
				}

				// Get current request headers, censoring the sensitive authentication header
				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
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
				logger.Errorf(
					"HTTP request panicked!\n\tRequest:\n\t\t| %s\n\tError:\n\t\t| %s\n\tStacktrace:\n\t\t| %s",
					req,
					r,
					scanUtils.StacktraceIndented("\t\t| "),
				)
				logger.Debugf("Aborting request.")

				// Abort further request execution and return error to client
				RespondInternalError(ctx)
			}
		}()

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}

// MwSetCorsHeaders adds CORS-relevant response headers to every response. Headers can (but should not) be
// overwritten by actual handler function, if processed subsequently.
func MwSetCorsHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Log middleware execution
		logger.Debugf("Preparing CORS headers.")

		// Set REST relevant response headers
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		ctx.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}

// MwSetSecurityHeaders adds security-relevant response headers to every response. Headers can (but should not) be
// overwritten by actual handler function, if processed subsequently.
func MwSetSecurityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Log middleware execution
		logger.Debugf("Preparing security headers.")

		// Get currently set response headers
		headers := ctx.Writer.Header()

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

		// Explicitly disable the legacy XSS auditor — it was removed from Chrome/Firefox
		// and can introduce XSS vulnerabilities in old IE. CSP handles this properly.
		headers.Set("X-XSS-Protection", "0")

		// Set referrer policy
		headers.Set("Referrer-Policy", "no-referrer")

		// Set cache control policies
		headers.Set("Surrogate-Control", "no-store")
		headers.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")

		// Generate a per-request nonce for PrimeVue's runtime <style> injection and
		// for the OIDC callback inline script. The nonce is shared via gin context so
		// other handlers can embed it in responses.
		var nonceBuf [16]byte
		if _, errNonce := rand.Read(nonceBuf[:]); errNonce != nil {
			logger.Errorf("Could not generate CSP nonce: %v", errNonce)
			respondError(ctx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			ctx.Abort()
			return
		}
		nonce := hex.EncodeToString(nonceBuf[:])
		ctx.Set("csp-nonce", nonce)

		// Set content security policy.
		// script-src: 'self' only — the Vite build uses vue.runtime.esm-bundler (no runtime
		//   template compiler) and @intlify/unplugin-vue-i18n pre-compiles translations at
		//   build time, so no new Function() calls remain; 'unsafe-eval' is not needed.
		// style-src: 'self' for the Vite-bundled CSS + nonce for PrimeVue's runtime
		//   <style> injection (Aura theme engine injects component styles at runtime).
		// style-src-attr: PrimeVue sets inline style attributes on elements (e.g. DataTable
		//   column widths, overlay positioning). Nonces cannot cover style attributes, so
		//   'unsafe-inline' is scoped to style-src-attr only, keeping <style> elements
		//   nonce-gated via style-src.
		// form-action: restricts where forms may submit; not covered by default-src fallback.
		// worker-src: no web workers or service workers used; explicit 'none' is safer than
		//   relying on the default-src fallback chain.
		policy := ""
		policy += fmt.Sprintf("%s %s; ", "default-src", "'self'")
		policy += fmt.Sprintf("%s %s; ", "script-src", "'self'")
		policy += fmt.Sprintf("%s %s 'nonce-%s'; ", "style-src", "'self'", nonce)
		policy += fmt.Sprintf("%s %s; ", "style-src-attr", "'unsafe-inline'")
		policy += fmt.Sprintf("%s %s; ", "connect-src", "'self'")
		policy += fmt.Sprintf("%s %s; ", "img-src", "'self' data:")
		policy += fmt.Sprintf("%s %s; ", "font-src", "'self'")
		policy += fmt.Sprintf("%s %s; ", "object-src", "'none'")
		policy += fmt.Sprintf("%s %s; ", "base-uri", "'self'")
		policy += fmt.Sprintf("%s %s; ", "form-action", "'self'")
		policy += fmt.Sprintf("%s %s; ", "worker-src", "'none'")
		policy += fmt.Sprintf("%s %s; ", "frame-ancestors", "'none'")
		policy = strings.TrimSuffix(policy, "; ")
		headers.Set("Content-Security-Policy", policy)

		// Set custom security relevant response headers
		headers.Set("Server", "Apache") // Haha

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}

// MwAbortOptionsRequest aborts OPTIONS requests because they should not reach backend handlers
func MwAbortOptionsRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Log middleware execution
		logger.Debugf("Checking for OPTIONS request.")

		// Tell clients that there is no OPTIONS request available
		ctx.Writer.Header().Set("access-control-allow-methods", "POST, GET")

		// Abort OPTIONS requests
		if ctx.Request.Method == "OPTIONS" {
			logger.Debugf("Aborted OPTIONS request.")
			ctx.AbortWithStatus(http.StatusOK)
			return
		}

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}

// MwJwtAuthentication validates protected requests and stores the authenticated user in their context
func MwJwtAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := GetContextLogger(ctx)

		// Log middleware execution
		logger.Debugf("Executing authentication middleware.")

		// Stop validation if route does not require authentication
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			logger.Debugf("Static file accessible without authentication.")
			ctx.Next() // Success, step to next middleware plugin, if existing
			return
		} else if scanUtils.StrContained(ctx.Request.URL.Path, apiEndpointsNoAuth) {
			logger.Debugf("API endpoint accessible without authentication.")
			ctx.Next() // Success, step to next middleware plugin, if existing
			return
		} else {
			logger.Debugf("API endpoint requires authentication.")
		}

		// Retrieve the authorization header from the request
		tokenHeader := ctx.Request.Header.Get("Authorization")

		// Stop validation if authorization header is missing
		if tokenHeader == "" {
			logger.Debugf("Authentication failed. Authentication token is missing.")
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		// Stop validation if token format does not meet expectation (`Bearer {token-body}`)
		splits := strings.Split(tokenHeader, " ")
		if len(splits) != 2 {
			logger.Debugf("Authentication failed. Authentication token is malformed.")
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		// Read and parse token
		assertion := &Jwt{}
		tokenPart := splits[1] // Grab the token part, what we are truly interested in
		token, errToken := jwt.ParseWithClaims(tokenPart, assertion, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		// Stop validation if token could not be parsed
		if errToken != nil {
			logger.Debugf("Authentication failed. Authentication token exceeded: %s", errToken)
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		// Stop validation if token is invalid. Maybe not signed on this server.
		if !token.Valid {
			logger.Debugf("Authentication failed. Authentication token is invalid.")
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		// Make sure there is a valid SAML ID associated with the token
		if assertion.UserId <= 0 {
			logger.Warningf("Authentication failed. Authentication token not assignable.")
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		userEntry, errUserEntry := database.GetUser(assertion.UserId)
		if errUserEntry != nil {
			logger.Warningf("Authentication failed. Could not query user: %s", errUserEntry)
			RespondInternalError(ctx) // Return generic error information
			return
		}

		// Check if user entry exists
		if userEntry == nil {
			logger.Warningf("Authentication failed. User with ID %d is not existing.", assertion.UserId)
			RespondAuthError(ctx)
			return
		}

		// Abort login if JWT revision does not match the user's matching counter. Web tokens are
		// validated against LogoutCount (bumped on each logout), API tokens against ApiTokenRevision
		// (bumped whenever a new API token is issued).
		var expectedRevision uint
		if assertion.Kind == JwtKindApi {
			expectedRevision = userEntry.ApiTokenRevision
		} else {
			expectedRevision = userEntry.LogoutCount
		}

		if assertion.Revision != expectedRevision {
			logger.Debugf("Authentication failed. Outdated authentication token.")
			RespondAuthError(ctx)
			return // Return, without processing further middleware plugins
		}

		// Log authentication
		logger.Debugf("Authenticated as '%s'.", userEntry.Email)

		// Stop if user is disabled
		if !userEntry.Active {
			logger.Debugf("Authorization failed. User is disabled.")
			RespondAuthError(ctx)
			return
		}

		// Validate admin rights, if necessary
		if userEntry.Admin {
			logger.Debugf("Authorization successful as role administrator.")
		} else {
			if scanUtils.StrContained(ctx.Request.URL.Path, apiEndpointsAdmin) {
				logger.Debugf("Authorization failed. Admin privileges required.")
				RespondAuthError(ctx)
				return
			} else {
				logger.Debugf("Authentication successful as user.")
			}
		}

		// Set user in current request context on server side. Handlers can use this data subsequently.
		storage := getContextStorage(ctx)
		storage.CurrentUser = userEntry
		storage.CurrentJwtKind = assertion.Kind

		// Success, step to next middleware plugin, if existing
		ctx.Next() // Should only be called in middleware handlers!
	}
}
