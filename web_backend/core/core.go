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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core/loaders"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

const JwtKindWeb = "Web"                       // JwtKindWeb identifies an interactive session token that is short-lived and automatically refreshed
const JwtKindApi = "API"                       // JwtKindApi identifies a long-lived API access token that is not automatically refreshed
const httpReadHeaderTimeout = 10 * time.Second // httpReadHeaderTimeout limits how long a client may take to send request headers
const httpIdleTimeout = 2 * time.Minute        // httpIdleTimeout limits how long an idle keep alive connection remains open
const httpShutdownTimeout = 10 * time.Second   // httpShutdownTimeout limits graceful shutdown work before connections are forced closed

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Application context and cancellation function. The backend should terminate when the context is closed.
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests. This needs to be accessible by handler packages
var httpRouter *gin.Engine                                                // HTTP router dispatching requests to associated handlers
var httpServer *http.Server                                               // Reference to the listening HTTP server instance
var httpsServer *http.Server                                              // Reference to the listening HTTPS server instance
var httpServersWg sync.WaitGroup                                          // Tracks the serving goroutines registered before shutdown
var accessLogFile *os.File                                                // Access log file owned by the backend lifecycle
var apiEndpointsAdmin []string                                            // List of routes that are allowed for admins only
var apiEndpointsNoAuth []string                                           // List of routes that do not require authentication
var jwtSecret string                                                      // Secret used to sign authentication messages
var jwtAlgorithm string                                                   // Algorithm used to sign authentication messages
var httpServerCrt string                                                  // Certificate used by the listening HTTPS server
var httpServerKey string                                                  // Key used by the listening HTTPS server

// Jwt represents a JWT token and its attributes, used for authentication
type Jwt struct {
	jwt.RegisteredClaims
	UserId   uint64 // The ID of the user this JWT token belongs to
	Revision uint   // For web tokens: user's LogoutCount. For API tokens: user's ApiTokenRevision
	Kind     string // "Web" = interactive session token, "API" = long-lived API access token
}

// Init initializes the backend and its dependencies
func Init() error {

	// Get config
	conf := config.GetConfig()

	// Prepare HTTP server certificate paths
	if conf.ListenAddressHttps != "" {
		httpServerCrt = filepath.Join("keys", "backend.crt")
		httpServerKey = filepath.Join("keys", "backend.key")
		if _build.DevMode {
			httpServerCrt = filepath.Join("keys", "backend_dev.crt")
			httpServerKey = filepath.Join("keys", "backend_dev.key")
		}
		errServerCrt := scanUtils.IsValidFile(httpServerCrt)
		if errServerCrt != nil {
			return errServerCrt
		}
		errServerKey := scanUtils.IsValidFile(httpServerKey)
		if errServerKey != nil {
			return errServerKey
		}
	}

	// Set GIN release mode before initializing
	if _build.DevMode {
	} else {
		gin.SetMode(gin.ReleaseMode) // Gin runs by default in debug mode. Needs to be switched to release mode
	}

	// Initialize database
	errOpen := database.Open()
	if errOpen != nil {
		return errOpen
	}

	// Create or update the db tables
	errMigrate := database.AutoMigrate()
	if errMigrate != nil {
		return fmt.Errorf("could not migrate backend database: %w", errMigrate)
	}

	// Set config values
	jwtSecret = conf.Jwt.Secret
	jwtAlgorithm = conf.Jwt.Algorithm

	// Initialize file for access log
	var errAccessLogFile error
	accessLogFile, errAccessLogFile = os.OpenFile("./logs/access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if errAccessLogFile != nil {
		return errAccessLogFile
	}

	// Set access log file. Needs to be set before engine is initialized.
	gin.DefaultWriter = io.MultiWriter(accessLogFile)

	// Disable Console Color, logs are written to a file
	gin.DisableConsoleColor()

	// Initialize GIN router. Initialize new one without default middleware, we have our custom one.
	httpRouter = gin.New()

	// Return 405 NOT ALLOWED if route is called with invalid method
	httpRouter.HandleMethodNotAllowed = true

	// Return a valid JSON body with HTTP error status code. Browser might try to parse response body as JSON.
	httpRouter.NoRoute(func(ctx *gin.Context) {
		respondError(ctx, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	})

	// Attach GIN default logger middleware, which will take care of access logging
	httpRouter.Use(gin.Logger())

	// Middleware to initialize request-context bound storage and logger. Mandatory for subsequent middleware.
	httpRouter.Use(MwInitializeRequest())

	// Middleware to intercept and log potential panics and tear down requests gracefully
	httpRouter.Use(MwPanicHandler())

	// Middleware to set CORS-relevant response headers
	httpRouter.Use(MwSetCorsHeaders()) // Sets CORS-relevant response headers

	// Middleware to set security relevant response headers
	httpRouter.Use(MwSetSecurityHeaders())

	// Middleware to takes care of options requests, which shell not run the actual request
	httpRouter.Use(MwAbortOptionsRequest())

	// Middleware to executes JWT authentication and some basic authorization checks
	httpRouter.Use(MwJwtAuthentication())

	// Middleware to enforce rate limits: IP-based for auth endpoints, user-based for authenticated API
	httpRouter.Use(MwRateLimit())

	// Middleware to compress all responses to minimize transmission times
	// This helps clients on slower connections a lot on larger responses, and might help you reduce bandwidth costs
	httpRouter.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedExtensions([]string{".zip"}),
	))

	// Initialize authenticators
	errAuthenticators := initAuthenticators(conf.Authenticator)
	if errAuthenticators != nil {
		return fmt.Errorf("could not initialize authenticators: %w", errAuthenticators)
	}

	// Initialize loaders
	errLoaders := loaders.Init(conf.Loader)
	if errLoaders != nil {
		return fmt.Errorf("could not initialize loaders: %w", errLoaders)
	}

	// Return as everything went fine
	return nil
}

// ConnectManager initializes the RPC client and starts connecting to the manager
func ConnectManager() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Prepare RPC certificate path
	rpcRemoteCrt := filepath.Join("keys", "manager.crt")
	if _build.DevMode {
		rpcRemoteCrt = filepath.Join("keys", "manager_dev.crt")
	}
	errRemoteCrt := scanUtils.IsValidFile(rpcRemoteCrt)
	if errRemoteCrt != nil {
		return errRemoteCrt
	}

	// Register gob structures that will be sent via interface{}
	manager.RegisterGobs()

	// Initialize RPC client manager facing.
	// The manager requires a shared secret to authorize the RPC connection.
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, conf.ManagerSsl, rpcRemoteCrt, conf.ManagerSecret)

	// Connect to manager but don't wait to start answering client requests. Connection attempt continues in background.
	_ = rpcClient.Connect(logger, true)

	// Return as everything went fine
	return nil

}

// Run starts the Gin web server
func Run() error {

	// Only launch if shutdown is not already in progress
	if coreCtx.Err() != nil {
		return nil
	}

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Serve the compiled Vue frontend in development and production
	frontendDir := "web_frontend"
	pathDist := filepath.Join("..", frontendDir, "dist")
	if errPath := scanUtils.IsValidFolder(pathDist); errPath != nil {

		// Development may start its API while Vite prepares the first build, but production requires the browser application
		if _build.DevMode {
			logger.Warningf(
				"Could not find frontend dist folder at '%s'. Run 'npm run build' or 'npm run build:watch' in '%s/'.",
				pathDist,
				frontendDir,
			)
		} else {
			return fmt.Errorf("could not validate frontend path '%s': %w", pathDist, errPath)
		}
	} else {

		// Serve static frontend assets
		httpRouter.StaticFile("/favicon.ico", filepath.Join(pathDist, "favicon.ico"))
		httpRouter.Static("/assets", filepath.Join(pathDist, "assets"))
		httpRouter.Static("/fonts", filepath.Join(pathDist, "fonts"))
		httpRouter.Static("/img", filepath.Join(pathDist, "img"))

		// Serve the agent download package maintained separately by the deployment
		httpRouter.Static("/app/downloads", filepath.Join(pathDist, "downloads"))

		// Preserve JSON 404s below the REST prefix and serve the SPA shell for browser navigation and deep links.
		// The per-request CSP nonce is inserted into the meta tag used by PrimeVue's runtime style injection.
		indexPath := filepath.Join(pathDist, "index.html")
		httpRouter.NoRoute(func(ctx *gin.Context) {
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/v1") {
				respondError(ctx, http.StatusNotFound, http.StatusText(http.StatusNotFound))
				return
			}
			if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
				respondError(ctx, http.StatusNotFound, http.StatusText(http.StatusNotFound))
				return
			}
			content, errContent := os.ReadFile(indexPath)
			if errContent != nil {
				respondError(ctx, http.StatusNotFound, http.StatusText(http.StatusNotFound))
				return
			}
			if nonce := ctx.GetString("csp-nonce"); nonce != "" {
				content = []byte(strings.ReplaceAll(string(content), "__CSP_NONCE__", nonce))
			}
			ctx.Data(http.StatusOK, "text/html; charset=utf-8", content)
		})
	}

	// Collect fatal listener errors without sharing mutable state between serving goroutines
	chServerError := make(chan error, 2)

	// Prepare and start the HTTP endpoint
	if conf.ListenAddressHttp != "" {

		// Prepare HTTP config
		listen := conf.ListenAddressHttp
		if strings.HasPrefix(listen, "*:") {
			listen = strings.Replace(listen, "*:", ":", 1)
		}

		// Prepare HTTP server
		httpServer = &http.Server{
			Addr:              listen,
			Handler:           httpRouter,
			ReadHeaderTimeout: httpReadHeaderTimeout,
			IdleTimeout:       httpIdleTimeout,
		}

		// Start HTTP endpoint
		httpServersWg.Add(1)
		go func() {
			defer httpServersWg.Done()
			errServe := httpServer.ListenAndServe()
			if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
				chServerError <- errServe
			}
			coreCtxCancelFunc()
		}()
	}

	// Prepare and start the HTTPS endpoint
	if conf.ListenAddressHttps != "" {

		// Prepare HTTPS config
		listen := conf.ListenAddressHttps
		if strings.HasPrefix(listen, "*:") {
			listen = strings.Replace(listen, "*:", ":", 1)
		}

		// Prepare HTTPS server
		httpsServer = &http.Server{
			Addr:              listen,
			Handler:           httpRouter,
			TLSConfig:         utils.TlsConfigFactory(),
			ReadHeaderTimeout: httpReadHeaderTimeout,
			IdleTimeout:       httpIdleTimeout,
		}

		// Start HTTPS endpoint
		httpServersWg.Add(1)
		go func() {
			defer httpServersWg.Done()
			errServe := httpsServer.ListenAndServeTLS(httpServerCrt, httpServerKey)
			if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
				chServerError <- errServe
			}
			coreCtxCancelFunc()
		}()
	}

	// Block until termination request
	<-coreCtx.Done()

	// Return a fatal listener error if one caused termination
	select {
	case errServer := <-chServerError:
		return errServer
	default:
		return nil
	}
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Prepare helper function to softly terminate HTTP servers
		softCloseHttpServer := func(logger scanUtils.Logger, serverName string, server *http.Server) {

			// Skip endpoints that were not configured
			if server == nil {
				return
			}

			// Bound graceful request draining so shutdown cannot wait indefinitely
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
			errShutdown := server.Shutdown(shutdownCtx)
			shutdownCancel()
			if errShutdown == nil {
				return
			}
			logger.Errorf("Could not shut down '%s' server: %s", serverName, errShutdown)

			// Force listener and connection closure after graceful draining failed
			if errClose := server.Close(); errClose != nil && !errors.Is(errClose, http.ErrServerClosed) {
				logger.Errorf("Could not force close '%s' server: %s", serverName, errClose)
			}
		}

		// Stop HTTP servers
		softCloseHttpServer(logger, "HTTP", httpServer)
		softCloseHttpServer(logger, "HTTPS", httpsServer)
		httpServersWg.Wait()

		// Cancel context-bound dependencies after active HTTP handlers have finished
		coreCtxCancelFunc()

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}

		// Make sure db gets closed on exit
		errClose := database.Close()
		if errClose != nil {
			logger.Errorf("Could not close backend database connection: '%s'", errClose)
		}

		// Release the access log file after all request handlers have stopped writing to it
		if accessLogFile != nil {
			if errLogClose := accessLogFile.Close(); errLogClose != nil {
				logger.Errorf("Could not close access log: '%s'", errLogClose)
			}
			accessLogFile = nil
		}
	})
}

// RegisterApiEndpoint registers a new route, which requires authentication.
func RegisterApiEndpoint(version string, method string, path string, handlers ...gin.HandlerFunc) {
	relUrl := GenerateRelativeUrl(version, strings.Trim(path, "/"))
	httpRouter.Handle(method, relUrl, handlers...)
}

// RegisterApiEndpointAdmin registers a new route, which requires admin privileges.
func RegisterApiEndpointAdmin(version string, method string, path string, handlers ...gin.HandlerFunc) {
	relUrl := GenerateRelativeUrl(version, strings.Trim(path, "/"))
	apiEndpointsAdmin = append(apiEndpointsAdmin, relUrl)
	RegisterApiEndpoint(version, method, path, handlers...)
}

// RegisterApiEndpointNoAuth registers a new route, which does not require authentication.
func RegisterApiEndpointNoAuth(version string, method string, path string, handlers ...gin.HandlerFunc) {
	relUrl := GenerateRelativeUrl(version, strings.Trim(path, "/"))
	apiEndpointsNoAuth = append(apiEndpointsNoAuth, relUrl)
	RegisterApiEndpoint(version, method, path, handlers...)
}

// GenerateRelativeUrl generates a relative URL for an endpoint
func GenerateRelativeUrl(version string, path string) string {
	return fmt.Sprintf("/api/%s/%s", version, strings.Trim(path, "/"))
}

// RpcClient exposes the RPC client to external packages
func RpcClient() *utils.Client {
	return rpcClient
}

// OwnerOrAdmin determines whether the user has sufficient rights to edit a group's elements (e.g. scan scope), which
// is the case if the user has an ownership relation or is an administrator.
func OwnerOrAdmin(groupId uint64, userEntry *database.T_user) bool {

	// Allow if user is admin
	if userEntry.Admin {
		return true
	}

	// Get user's group ownerships to see if he has access rights (map for more efficient lookups)
	groups := make(map[uint64]struct{}, len(userEntry.Ownerships))
	for _, ownership := range userEntry.Ownerships {
		groups[ownership.Group.Id] = struct{}{}
	}

	// Check if given group ID is contained in user's owned groups
	if _, ok := groups[groupId]; ok {
		return true
	}

	// Return false otherwise
	return false
}
