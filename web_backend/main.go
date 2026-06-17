/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

// @title           Large-Scale Discovery API
// @version         1.0
// @description     Network scanning solution for information gathering in large IT/OT network environments.
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost
// @BasePath        /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/Large-Scale-Discovery/web_backend/handlers"
)

// Build information accessible via -version
var buildGitCommit = "dev12345"                       // Git commit hash identifying the version of this scan agent. Injected by the build command.
var buildTimestamp = "0001-01-01T00:00:00+00:00"      // Timestamp when this agent was built. Injected by the build command.
var buildGoVersion = runtime.Version()                // Golang version used during building of the agent.
var buildGoArch = runtime.GOOS + "/" + runtime.GOARCH // Golang version used during building of the agent.

// main application entry point
func main() {

	// Introduce Gracy to take care of cleanup/shutdown actions on interrupt
	gracy := utils.NewGracy()

	// Register Gracy as the interrupt handler in duty
	gracy.Promote()

	// We paid Gracy, let her execute nevertheless (e.g. if in case of panic rather than interrupt)
	defer gracy.Shutdown()

	// Declare command line arguments
	versionFlag := flag.Bool("version", false, "Prints build information.")

	// Parse command line arguments
	flag.Parse()

	// Print version information
	if *versionFlag {
		fmt.Printf("Backend:\n%s\n", "\t"+strings.Join(buildInfo(), "\n\t"))
		return
	}

	// Initialize configuration
	errConf := config.Init("backend.conf")
	if errConf != nil {
		fmt.Println("Could not load configuration:", errConf)
		return
	}

	// Get config
	conf := config.GetConfig()

	// Initialize logger
	logger, errLog := log.InitGlobalLogger(conf.Logging)
	if errLog != nil {
		fmt.Println("could not initialize logger: ", errLog)
		return
	}

	// Capture fatal runtime crashes (concurrent map writes, stack overflows, etc.) to file
	log.SetCrashOutput(conf.Logging)

	// Make sure logger gets closed on exit
	gracy.Register(func() {
		err := log.CloseGlobalLogger()
		if err != nil {
			fmt.Printf("could not close logger: %s\n", err)
		}
	})

	// Log start
	logger.Debugf("Starting backend.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Make backend print final message on exit
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Backend terminated.")
	})

	// Log binary information
	for _, info := range buildInfo() {
		logger.Debugf("%s", info)
	}

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize database, http router and middleware, authenticators and loaders
	errInit := core.Init()
	if errInit != nil {
		logger.Errorf("Could not initialize backend: %s", errInit)
		return
	}

	// Initialize RPC connection to manager
	errConnectManager := core.ConnectManager()
	if errConnectManager != nil {
		logger.Errorf("Could not initialize connection: %s", errConnectManager)
		return
	}

	// Write development sample data to the database
	if _build.DevMode {
		errDev := database.DeploySampleData()
		if errDev != nil {
			logger.Warningf("Could not write development sample data: %s", errDev)
			return
		}
	}

	// Copy same SMTP configuration from logger for backend e-mailing.
	smtpConfig := conf.Logging.Smtp.Connector

	// Initialize REST API endpoints
	registerApiEndpointsV1(&smtpConfig)

	// Start the backend web server. This will start serving two interfaces:
	//   - The REST API mounted under /api
	//   - The web application frontend under /
	// The web frontend is served from ../web_frontend/src/ in dev mode and from ../web_frontend/dist/ otherwise.
	errLaunch := core.Run()
	if errLaunch != nil {
		logger.Errorf("Could not launch backend: %s", errLaunch)
	}
}

// registerApiEndpointsV1 registers REST endpoints and some authentication requirements for API version 1
func registerApiEndpointsV1(smtpConfig *utils.Smtp) {

	// API version string
	v := "v1"

	//
	// Unauthenticated routes
	//
	core.RegisterApiEndpointNoAuth(v, "GET", "/backend/settings", handlers.BackendSettings())
	core.RegisterApiEndpointNoAuth(v, "POST", "/backend/authenticator", handlers.BackendAuthenticator())

	//
	// USER authenticated routes
	//
	core.RegisterApiEndpoint(v, "GET", "/swagger/doc.json", handlers.ApiDocs())
	core.RegisterApiEndpoint(v, "POST", "/backend/logout", handlers.BackendLogout())

	core.RegisterApiEndpoint(v, "GET", "/users", handlers.Users())              // Returns list of already registered users. All for administrators, a filtered/restricted list for other users
	core.RegisterApiEndpoint(v, "GET", "/user/details", handlers.UserDetails()) // Returns the current user's details
	core.RegisterApiEndpoint(v, "POST", "/user/password", handlers.UserResetDbPassword(smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/user/api-token", handlers.UserApiToken(smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/user/feedback", handlers.UserFeedback(smtpConfig))

	core.RegisterApiEndpoint(v, "GET", "/groups", handlers.Groups()) // Returns a list of groups the user is owner of (all in case of admin)

	core.RegisterApiEndpoint(v, "GET", "/scopes", handlers.Scopes()) // Returns a list of scopes the user is owner of (all in case of admin)
	core.RegisterApiEndpoint(v, "POST", "/scope/rescan", handlers.ScopeResetFailed())
	core.RegisterApiEndpoint(v, "POST", "/scope/cycle", handlers.ScopeNewCycle())
	core.RegisterApiEndpoint(v, "POST", "/scope/pause", handlers.ScopeTogglePause())
	core.RegisterApiEndpoint(v, "POST", "/scope/secret", handlers.ScopeResetSecret(smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/scope/delete", handlers.ScopeDelete())
	core.RegisterApiEndpoint(v, "POST", "/scope/targets", handlers.ScopeTargets())
	core.RegisterApiEndpoint(v, "POST", "/scope/target/reset", handlers.ScopeResetInput())
	core.RegisterApiEndpoint(v, "POST", "/scope/update/custom", handlers.ScopeCreateUpdateCustom())
	core.RegisterApiEndpoint(v, "POST", "/scope/update/networks", handlers.ScopeCreateUpdateNetworks())
	core.RegisterApiEndpoint(v, "POST", "/scope/update/assets", handlers.ScopeCreateUpdateAssets())
	core.RegisterApiEndpoint(v, "POST", "/scope/update/settings", handlers.ScopeUpdateSettings())

	core.RegisterApiEndpoint(v, "GET", "/agents", handlers.Agents())
	core.RegisterApiEndpoint(v, "POST", "/agent/delete", handlers.AgentDelete())

	core.RegisterApiEndpoint(v, "GET", "/views", handlers.Views())                // Returns a list of views a user is owner of (all in case of admin)
	core.RegisterApiEndpoint(v, "GET", "/views/granted", handlers.ViewsGranted()) // Returns a list of views a user has access rights to
	core.RegisterApiEndpoint(v, "POST", "/view/create", handlers.ViewCreate())
	core.RegisterApiEndpoint(v, "POST", "/view/delete", handlers.ViewDelete())
	core.RegisterApiEndpoint(v, "POST", "/view/update", handlers.ViewUpdate())
	core.RegisterApiEndpoint(v, "POST", "/view/grant/token", handlers.ViewGrantToken(smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/view/grant/users", handlers.ViewGrantUsers(smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/view/grant/revoke", handlers.ViewGrantRevoke())

	//
	// ADMIN authenticated routes
	//
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/events", handlers.Events())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/sql", handlers.SqlLogs())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/user/update", handlers.UserUpdate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/user/delete", handlers.UserDelete())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/create", handlers.GroupCreate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/update", handlers.GroupUpdate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/delete", handlers.GroupDelete())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/assign", handlers.GroupAssign(smtpConfig))

	core.RegisterApiEndpointAdmin(v, "GET", "/admin/databases", handlers.Databases())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/database/update", handlers.DatabaseAddUpdate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/database/remove", handlers.DatabaseRemove())
}

func buildInfo() []string {
	return []string{
		fmt.Sprintf("Build Timestamp   : %s", buildTimestamp),
		fmt.Sprintf("Build GIT Commit  : %s", buildGitCommit[:8]),
		fmt.Sprintf("Build Go Version  : %s", buildGoVersion),
		fmt.Sprintf("Build OS/Arch     : %s", buildGoArch),
	}
}
