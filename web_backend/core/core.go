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
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Backend context and context cancellation function. Backend should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests. This needs to be accessible by handler packages
var httpRouter *gin.Engine                                                // HTTP Router dispatching HTTP requests to associated request handlers
var apiEndpointsAdmin []string                                            // List of routes that are allowed for admins only
var apiEndpointsNoAuth []string                                           // List of routes that do not require authentication
var jwtSecret string                                                      // Secret used to sign authentication messages
var jwtAlgorithm string                                                   // Algorithm used to sign authentication messages
var httpServerCrt string                                                  // Certificate used by the listening RPC server
var httpServerKey string                                                  // Key used by the listening RPC server

// Jwt represents a JWT token and it's attributes, used for authentication
type Jwt struct {
	jwt.RegisteredClaims
	UserId   uint64 // The ID of the user this JWT token belongs to
	Revision uint   // User's logout count to invalidate previously issued JWT tokens ahead of time (JWT tokens not stored on the server-side like sessions!)
}

func Init() error {

	// Get config
	conf := config.GetConfig()

	// Prepare backend key and certificate path
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
		return fmt.Errorf("could not migrate backend db: %s", errMigrate)
	}

	// Set config values
	jwtSecret = conf.Jwt.Secret
	jwtAlgorithm = conf.Jwt.Algorithm

	// Initialize file for access log
	accessLogFile, errOpenFile := os.OpenFile("./logs/access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if errOpenFile != nil {
		return errOpenFile
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
	httpRouter.NoRoute(func(c *gin.Context) {
		respondError(c, http.StatusNotFound, "404 Not Found")
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

	// Middleware to compress all responses to minimize transmission times
	// This helps clients on slower connections a lot on larger responses, and might help you reduce bandwidth costs
	httpRouter.Use(gzip.Gzip(gzip.DefaultCompression))

	// Initialize authenticators
	errAuthenticators := initAuthenticators(conf.Authenticator)
	if errAuthenticators != nil {
		return errAuthenticators
	}

	// Initialize loaders
	errLoaders := initLoaders(conf.Loader)
	if errLoaders != nil {
		return errLoaders
	}

	// Return as everything went fine
	return nil
}

func InitManager() error {

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

	// Initialize RPC client manager facing
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, conf.ManagerAddressSsl, rpcRemoteCrt)

	// Connect to manager but don't wait to start answering client requests. Connection attempt continues in background.
	_ = rpcClient.Connect(logger, true)

	// Return as everything went fine
	return nil

}

// Run starts the gin web server
func Run() error {

	// Only launch if shutdown is not already in progress
	select {
	case <-coreCtx.Done():
		return nil
	default:
	}

	// Get config
	conf := config.GetConfig()

	// Serve web frontend (single-page JavaScript application)
	appBase := ""
	if _build.DevMode {

		// Prepare frontend paths
		pathSrc := "../web_frontend/src/"
		pathModules := "../web_frontend/node_modules/"

		// Check if frontend paths are valid
		errPath := scanUtils.IsValidFolder(pathSrc)
		if errPath != nil {
			return fmt.Errorf("invalid frontend path '%s'", pathSrc)
		}

		errPath2 := scanUtils.IsValidFolder(pathModules)
		if errPath2 != nil {
			return fmt.Errorf("invalid frontend path '%s'", pathModules)
		}

		// Set app base
		appBase = "/src/"

		// Serve web frontend from ./src folder in development mode.
		// This requires some other paths to be served too...
		httpRouter.Static(appBase, pathSrc)
		httpRouter.Static("/node_modules/", pathModules)
		httpRouter.StaticFile("/favicon.ico", filepath.Join(pathSrc, "favicon.ico"))

	} else {

		// Prepare frontend path
		pathDist := "../web_frontend/dist/"

		// Check if frontend paths are valid
		errPath := scanUtils.IsValidFolder(pathDist)
		if errPath != nil {
			return fmt.Errorf("invalid frontend path '%s'", pathDist)
		}

		// Set app base
		appBase = "/app/"

		// Serve web frontend from ./dist folder in production mode.
		// The dist folder should contain an optimized and minified version built from ./src with Gulp.
		httpRouter.Static(appBase, pathDist)
		httpRouter.StaticFile("/favicon.ico", filepath.Join(pathDist, "favicon.ico"))
	}

	// Redirect root requests to application
	// Due to a routes deficit in Gin, the root cannot be served directly
	httpRouter.GET("/", func(c *gin.Context) {

		// Attach query parameters if existing. Query parameter might be used by e.g. oauth callbacks to
		// pass error messages on to the web interface
		url := appBase
		if len(c.Request.URL.RawQuery) > 0 {
			url += "?" + c.Request.URL.RawQuery
		}

		// Redirect to app
		c.Redirect(http.StatusMovedPermanently, url)
	})

	// Prepare potential error
	var err error = nil

	// Start goroutine for HTTP endpoint
	if conf.ListenAddressHttp != "" {
		go func() {

			// Convert wildcard notation to suitable variant
			listen := conf.ListenAddressHttp
			if strings.HasPrefix(listen, "*:") {
				listen = strings.Replace(listen, "*:", ":", 1)
			}

			// Create TLS web server
			server := &http.Server{Addr: listen, Handler: httpRouter}

			// Start listening
			errServe := server.ListenAndServe()
			if errServe != nil {
				err = errServe
				coreCtxCancelFunc()
			}
		}()
	}

	// Start goroutine for HTTPS endpoint
	if conf.ListenAddressHttps != "" {
		go func() {

			// Convert wildcard notation to suitable variant
			listen := conf.ListenAddressHttps
			if strings.HasPrefix(listen, "*:") {
				listen = strings.Replace(listen, "*:", ":", 1)
			}

			// Create the TLS conf
			tlsConf := utils.TlsConfigFactory()

			// Create TLS web server
			server := &http.Server{Addr: listen, Handler: httpRouter, TLSConfig: tlsConf}

			// Start listening
			errServe := server.ListenAndServeTLS(httpServerCrt, httpServerKey)
			if errServe != nil {
				err = errServe
				coreCtxCancelFunc()
			}
		}()
	}

	// Block until termination request
	<-coreCtx.Done()

	// Return nil if everything went fine or webserver initialization error
	return err
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Close agent context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}

		// Make sure db gets closed on exit
		errClose := database.Close()
		if errClose != nil {
			logger.Errorf("Could not close backend db connection: '%s'", errClose)
		}
	})
}

// RegisterApiEndpoint registers a new route, which requires authentication.
func RegisterApiEndpoint(version string, method string, path string, handlers ...gin.HandlerFunc) {
	relUrl := GenerateRelativeUrl(version, strings.Trim(path, "/"))
	httpRouter.Handle(method, relUrl, handlers...)
}

// RegisterApiEndpointAdmin registers a new route, which should not require authentication.
func RegisterApiEndpointAdmin(version string, method string, path string, handlers ...gin.HandlerFunc) {
	relUrl := GenerateRelativeUrl(version, strings.Trim(path, "/"))
	apiEndpointsAdmin = append(apiEndpointsAdmin, relUrl)
	RegisterApiEndpoint(version, method, path, handlers...)
}

// RegisterApiEndpointNoAuth registers a new route, which requires admin privileges.
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
func OwnerOrAdmin(groupId uint64, user *database.T_user) bool {

	// Allow if user is admin
	if user.Admin {
		return true
	}

	// Get user's group ownerships to see if he has access rights (map for more efficient lookups)
	groups := make(map[uint64]struct{}, len(user.Ownerships))
	for _, ownership := range user.Ownerships {
		groups[ownership.Group.Id] = struct{}{}
	}

	// Check if given group ID is contained in user's owned groups
	if _, ok := groups[groupId]; ok {
		return true
	}

	// Return false otherwise
	return false
}
