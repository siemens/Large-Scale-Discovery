/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package main

import (
	"fmt"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
	"github.com/siemens/Large-Scale-Discovery/web_backend/handlers"
	"time"
)

// main application entry point
func main() {

	// Introduce Gracy to take care of cleanup/shutdown actions on interrupt
	gracy := utils.NewGracy()

	// Register Gracy as the interrupt handler in duty
	gracy.Promote()

	// We paid Gracy, let her execute nevertheless (e.g. if in case of panic rather than interrupt)
	defer gracy.Shutdown()

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

	// Make sure logger gets closed on exit
	gracy.Register(func() {
		err := log.CloseGlobalLogger()
		if err != nil {
			fmt.Printf("could not close logger: %s\n", err)
		}
	})

	// Make backend print final message on exit
	defer func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Backend terminated.")
	}()

	// Catch potential panics to gracefully log issue with stacktrace
	gracy.Register(func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
		}
	})

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize database, http router and middleware, authenticators and loaders
	errInit := core.Init()
	if errInit != nil {
		logger.Errorf("Could not initialize backend: %s", errInit)
		return
	}

	// Initialize RPC connection to manager
	errInitManager := core.InitManager()
	if errInitManager != nil {
		logger.Errorf("Could not initialize connection: %s", errInitManager)
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
	registerApiEndpointsV1(conf.FrontendUrl, &smtpConfig)

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
func registerApiEndpointsV1(frontendUrl string, smtpConfig *utils.Smtp) {

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
	core.RegisterApiEndpoint(v, "POST", "/backend/logout", handlers.BackendLogout())

	core.RegisterApiEndpoint(v, "GET", "/users", handlers.Users())              // Returns list of already registered users. All for administrators, a filtered/restricted list for other users
	core.RegisterApiEndpoint(v, "GET", "/user/details", handlers.UserDetails()) // Returns the current user's details
	core.RegisterApiEndpoint(v, "POST", "/user/reset", handlers.UserResetDbPassword(frontendUrl, smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/user/feedback", handlers.UserFeedback(smtpConfig))

	core.RegisterApiEndpoint(v, "GET", "/groups", handlers.Groups()) // Returns a list of groups the user is owner of (all in case of admin)

	core.RegisterApiEndpoint(v, "GET", "/scopes", handlers.Scopes()) // Returns a list of scopes the user is owner of (all in case of admin)
	core.RegisterApiEndpoint(v, "POST", "/scope/cycle", handlers.ScopeNewCycle())
	core.RegisterApiEndpoint(v, "POST", "/scope/pause", handlers.ScopeTogglePause())
	core.RegisterApiEndpoint(v, "POST", "/scope/secret", handlers.ScopeResetSecret(frontendUrl, smtpConfig))
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
	core.RegisterApiEndpoint(v, "POST", "/view/grant/token", handlers.ViewGrantToken(frontendUrl, smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/view/grant/users", handlers.ViewGrantUsers(frontendUrl, smtpConfig))
	core.RegisterApiEndpoint(v, "POST", "/view/grant/revoke", handlers.ViewGrantRevoke())

	//
	// ADMIN authenticated routes
	//
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/events", handlers.Events())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/user/update", handlers.UserUpdate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/user/delete", handlers.UserDelete())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/create", handlers.GroupCreate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/update", handlers.GroupUpdate())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/delete", handlers.GroupDelete())
	core.RegisterApiEndpointAdmin(v, "POST", "/admin/group/assign", handlers.GroupAssign(frontendUrl, smtpConfig))
}
