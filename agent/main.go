/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package main

import (
	"flag"
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"large-scale-discovery/agent/config"
	"large-scale-discovery/agent/core"
	"large-scale-discovery/log"
	"large-scale-discovery/utils"
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

	// Declare command line arguments
	setupFlag := flag.Bool("setup", false, "Executes setup. Requires administrative privileges.")

	// Declare linux specific command line argument
	flag.String("user", "", "The user to grant NFS sudoers rights.")

	// Parse command line arguments
	flag.Parse()

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

	// Make sure logger gets closed gracefully
	gracy.Register(func() {
		err := log.CloseGlobalLogger()
		if err != nil {
			fmt.Printf("could not close logger: %s\n", err)
		}
	})

	// Make agent print final message just before termination
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Agent terminated.")
	})

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
		}
	}()

	// Setup, if requested
	if *setupFlag {

		// Execute setup
		errSetup, errModule := core.Setup()
		if errSetup != nil {
			logger.Errorf("Setup of '%s' module failed: %s", errModule, errSetup)
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
	errCheck, errModule := core.CheckSetup()
	if errCheck != nil {
		logger.Errorf("Could not initialize scan module '%s': %s", errModule, errCheck)
		logger.Infof("Please check the configuration and run agent once with '-setup' argument!")
		return
	}

	// Verify scan module related config values set by the user
	errTest := core.CheckConfig()
	if errTest != nil {
		logger.Errorf("Invalid configuration for %s", errTest)
		logger.Infof("Please check the configuration!")
		return
	}

	// Warn if agent is run with admin privileges
	if scanUtils.IsElevated() {
		logger.Errorf("Scan agent is running with admin rights!")
	}

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize agent
	errInit := core.Init()
	if errInit != nil {
		logger.Errorf("Could not initialize agent: %s", errInit)
		return
	}

	// Request scan tasks, execute and submit results
	core.Run()
}
