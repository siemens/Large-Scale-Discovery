/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
	"github.com/siemens/Large-Scale-Discovery/agent/core"
	broker "github.com/siemens/Large-Scale-Discovery/broker/core"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
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
	setupFlag := flag.Bool("setup", false, "Executes setup. Requires administrative privileges.")
	versionFlag := flag.Bool("version", false, "Prints build information.")

	// Declare linux specific command line argument
	flag.String("user", "", "The user to grant NFS sudoers rights.")

	// Parse command line arguments
	flag.Parse()

	// Print version information
	if *versionFlag {
		fmt.Printf("Agent:\n%s\n", "\t"+strings.Join(buildInfo(), "\n\t"))
		return
	}

	// Initialize configuration module
	errConf := config.Init("agent.conf")
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

	// Capture fatal runtime crashes to file
	log.SetCrashOutput(conf.Logging)

	// Make sure logger gets closed gracefully
	gracy.Register(func() {
		err := log.CloseGlobalLogger()
		if err != nil {
			fmt.Printf("could not close logger: %s\n", err)
		}
	})

	// Log start
	logger.Debugf("Starting agent.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Make agent print final message just before termination
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Agent terminated.")
	})

	// Log binary information
	for _, info := range buildInfo() {
		logger.Debugf("%s", info)
	}

	// Setup, if requested
	if *setupFlag {

		// Execute setup
		module, errModule := core.Setup()
		if errModule != nil {
			logger.Errorf("Setup of '%s' module failed: %s", module, errModule)
			return
		}

		// Check if setup was running with admin rights (depending on scan modules, it might be necessary)
		if scanUtils.IsElevated() {
			logger.Infof("Setup successful. Please restart without admin rights.")
		} else {
			logger.Infof("Setup successful.")
		}

		// Always terminate to make user start agent without setup flag
		return
	}

	// Verify scan module setup
	module, errModule := core.CheckSetup(logger)
	if errModule != nil {
		logger.Errorf("Could not initialize scan module '%s': %s", module, errModule)
		logger.Infof("Please check the configuration and run agent once with '-setup' argument!")
		return
	}

	// Verify scan module related config values set by the user
	errTest := core.CheckConfig(logger)
	if errTest != nil {
		logger.Errorf("Invalid configuration for %s", errTest)
		logger.Infof("Please check the configuration!")
		return
	}

	// Warn if agent is run with admin privileges
	if scanUtils.IsElevated() {
		logger.Errorf("Scan agent is running with admin rights!")
	}

	// Initialize asset inventory functionality of discovery module
	errInventory := discovery.InitInventories(logger, conf.Authentication.Inventories)
	if errInventory != nil {
		logger.Errorf("Could not initialize asset inventory: %s", errInventory)
		logger.Infof("Please check the configuration!")
		return
	}

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize agent
	errInit := core.Init(buildGitCommit, buildTimestamp)
	if errInit != nil {
		logger.Errorf("Could not initialize agent: %s", errInit)
		return
	}

	// Request scan tasks, execute and submit results
	core.Run()
}

func buildInfo() []string {
	return []string{
		fmt.Sprintf("Build Timestamp   : %s", buildTimestamp),
		fmt.Sprintf("Build GIT Commit  : %s", buildGitCommit[:8]),
		fmt.Sprintf("Build Go Version  : %s", buildGoVersion),
		fmt.Sprintf("Build OS/Arch     : %s", buildGoArch),
		fmt.Sprintf("Broker API Version: %s", broker.BrokerApiVersion.String()),
	}
}
