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

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/manager/config"
	"github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

// Build information accessible via -version
var buildGitCommit = "dev12345"                       // Git commit hash identifying this binary. Injected by the build command.
var buildTimestamp = "0001-01-01T00:00:00+00:00"      // Timestamp when this binary was built. Injected by the build command.
var buildGoVersion = runtime.Version()                // Go version used to build this binary.
var buildGoArch = runtime.GOOS + "/" + runtime.GOARCH // OS and architecture targeted by this binary.

// main is the application entry point
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
		fmt.Printf("Manager:\n%s\n", "\t"+strings.Join(buildInfo(), "\n\t"))
		return
	}

	// Initialize configuration
	errConf := config.Init("manager.conf")
	if errConf != nil {
		fmt.Printf("Could not load configuration: %s\n", errConf)
		return
	}

	// Get config
	conf := config.GetConfig()

	// Initialize logger
	logger, errLogger := log.InitGlobalLogger(conf.Logging)
	if errLogger != nil {
		fmt.Printf("Could not initialize logger: %s\n", errLogger)
		return
	}

	// Capture fatal runtime crashes (concurrent map writes, stack overflows, etc.) to file
	log.SetCrashOutput(conf.Logging)

	// Make sure logger gets closed on exit
	gracy.Register(func() {
		if errClose := log.CloseGlobalLogger(); errClose != nil {
			fmt.Printf("Could not close logger: %s\n", errClose)
		}
	})

	// Log start
	logger.Debugf("Starting manager.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Print the final component message just before termination
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Manager terminated.")
	})

	// Log binary information
	for _, info := range buildInfo() {
		logger.Debugf("%s", info)
	}

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize manager
	errInit := core.Init()
	if errInit != nil {
		logger.Errorf("Could not initialize manager: %s", errInit)
		return
	}

	// Listen and serve connections
	errServe := core.Run()
	if errServe != nil {
		logger.Errorf("Could not serve manager RPC: %s", errServe)
	}
}

// buildInfo returns the metadata embedded in this binary
func buildInfo() []string {
	return []string{
		fmt.Sprintf("Build Timestamp   : %s", buildTimestamp),
		fmt.Sprintf("Build GIT Commit  : %s", buildGitCommit[:8]),
		fmt.Sprintf("Build Go Version  : %s", buildGoVersion),
		fmt.Sprintf("Build OS/Arch     : %s", buildGoArch),
	}
}
