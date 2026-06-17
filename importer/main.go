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
	"github.com/siemens/Large-Scale-Discovery/importer/config"
	"github.com/siemens/Large-Scale-Discovery/importer/core"
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
	versionFlag := flag.Bool("version", false, "Prints build information.")

	// Parse command line arguments
	flag.Parse()

	// Print version information
	if *versionFlag {
		fmt.Printf("Importer:\n%s\n", "\t"+strings.Join(buildInfo(), "\n\t"))
		return
	}

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
	logger.Debugf("Starting importer.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Make agent print final message on exit
	gracy.Register(func() {
		time.Sleep(time.Microsecond) // Make sure this message is written last, in case of race condition
		logger.Debugf("Importer terminated.")
	})

	// Log binary information
	for _, info := range buildInfo() {
		logger.Debugf("%s", info)
	}

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

func buildInfo() []string {
	return []string{
		fmt.Sprintf("Build Timestamp   : %s", buildTimestamp),
		fmt.Sprintf("Build GIT Commit  : %s", buildGitCommit[:8]),
		fmt.Sprintf("Build Go Version  : %s", buildGoVersion),
		fmt.Sprintf("Build OS/Arch     : %s", buildGoArch),
	}
}
