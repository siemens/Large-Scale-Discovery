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
	"fmt"
	"large-scale-discovery/importer/config"
	"large-scale-discovery/importer/core"
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

	// Initialize config
	errConf := config.Init("importer.conf")
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

	// Make agent print final message on exit
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Importer terminated.")
	})

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
		}
	}()

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize importer
	errInit := core.Init()
	if errInit != nil {
		logger.Errorf("Could not initialize importer: %s", errInit)
		return
	}

	// Check for changed targets in scan scopes and update related scope db
	errRun := core.Run()
	if errRun != nil {
		logger.Errorf("Could not run importer: %s", errRun)
	}
}
