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
	"crypto/tls"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/noneymous/PgProxy/pgproxy"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/pgproxy/config"
	"github.com/siemens/Large-Scale-Discovery/pgproxy/core"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

// Build information accessible via -version
var buildGitCommit = "dev12345"                       // Git commit hash identifying the version of this scan agent. Injected by the build command.
var buildTimestamp = "0001-01-01T00:00:00+00:00"      // Timestamp when this agent was built. Injected by the build command.
var buildGoVersion = runtime.Version()                // Golang version used during building of the agent.
var buildGoArch = runtime.GOOS + "/" + runtime.GOARCH // Golang version used during building of the agent.

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
		fmt.Printf("PgProxy:\n%s\n", "\t"+strings.Join(buildInfo(), "\n\t"))
		return
	}

	// Initialize config
	errConf := config.Init("pgproxy.conf")
	if errConf != nil {
		fmt.Printf("Could not load configuration: %s.\n", errConf)
		return
	}

	// Get config
	conf := config.GetConfig()

	// Initialize logger
	logger, errLog := log.InitGlobalLogger(conf.Logging)
	if errLog != nil {
		fmt.Printf("Could not initialize logger: %s.\n", errLog)
		return
	}

	// Capture fatal runtime crashes (concurrent map writes, stack overflows, etc.) to file
	log.SetCrashOutput(conf.Logging)

	// Make sure logger gets closed on exit
	gracy.Register(func() {
		err := log.CloseGlobalLogger()
		if err != nil {
			fmt.Printf("Could not close logger: %s.\n", err)
		}
	})

	// Log start
	logger.Debugf("Starting PgProxy.")

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
		logger.Debugf("PgProxy terminated.")
	})

	// Log binary information
	for _, info := range buildInfo() {
		logger.Debugf("%s", info)
	}

	// Make sure core gets shut down gracefully
	gracy.Register(core.Shutdown)

	// Initialize RPC connection to manager
	errConnectManager := core.ConnectManager()
	if errConnectManager != nil {
		logger.Errorf("Could not initialize connection: %s", errConnectManager)
		return
	}

	// Prepare some reasonable TLS config
	tlsConf := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		MaxVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			// Limit cipher suites to secure ones https://ciphersuite.info/cs/
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}

	// Initialize PgProxy
	pgProxy, errPgProxy := pgproxy.Init(logger, conf.Port, tlsConf, conf.ForceSsl, conf.DefaultSni)
	if errPgProxy != nil {
		logger.Errorf("Could not initialize PgProxy: %s.", errPgProxy)
		return
	}

	// Prepare wait group for RPC requests running in the background
	wg := &sync.WaitGroup{}

	// Register monitoring function
	pgProxy.RegisterMonitoring(func(
		loggerPgProxy scanUtils.Logger, // Internal logger from PgProxy, within the context of a client connection
		dbName string,
		dbUser string,
		dbTables []string,
		query string,
		queryResults int,
		queryStart time.Time,
		queryEndExecution time.Time,
		queryEndTotal time.Time,
		clientName string,
	) error {

		// Indent lines
		logMsg := "    " + strings.Join(strings.Split(query, "\n"), "\n    ")

		// Filter queries targeting postgres default database
		if dbName == "postgres" {
			loggerPgProxy.Debugf("Not logging query against database 'postgres':\n%s", logMsg)
			return nil
		}

		// Filter queries calling functions or something not a relevant table
		if len(dbTables) == 0 {
			loggerPgProxy.Debugf("Not logging query without target table:\n%s", logMsg)
			return nil
		}

		// Filter queries targeting pg tables
		for _, table := range dbTables {
			if strings.HasPrefix(strings.ToLower(table), "pg_") {
				loggerPgProxy.Debugf("Not logging query against table 'pg_*':\n%s", logMsg)
				return nil
			}
		}

		// Filter certain kind of irrelevant queries
		s := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(query, " ", ""), "\n", ""))
		if strings.HasPrefix(s, "set") {
			loggerPgProxy.Debugf("Not logging SET query:\n%s", logMsg)
			return nil
		} else if s == "selectversion()" {
			loggerPgProxy.Debugf("Not logging VERSION query:\n%s", logMsg)
			return nil
		}

		// Log query with stats
		loggerPgProxy.Debugf("Query of user '%s' ran %s and returned %d row(s) in %s: \n%s", dbUser, queryEndExecution.Sub(queryStart), queryResults, queryEndTotal.Sub(queryStart), logMsg)

		// Submit log entry in the background
		wg.Add(1)
		go func() {

			// Decrement wait group counter at the end
			defer wg.Done()

			// Send log data to manager for storage within associated scan scope database
			errCreate := manager.RpcCreateSqlLogs(
				loggerPgProxy,
				core.RpcClient(),
				dbName,
				dbUser,
				strings.Join(dbTables, "\n"),
				query,
				queryResults,
				queryStart,
				queryEndExecution.Sub(queryStart),
				queryEndTotal.Sub(queryStart),
				clientName,
			)
			if errCreate != nil {
				loggerPgProxy.Warningf("Could not submit log data: %s", errCreate)
			}
		}()

		// Return from monitoring function
		return nil
	})

	// Make sure core gets shut down gracefully
	gracy.Register(pgProxy.Stop)

	// Register proxy interfaces and routes
	errAdd := pgProxy.RegisterSni(conf.Snis...)
	if errAdd != nil {
		logger.Errorf("Could not add PgProxy SNI: %s.", errAdd)
		return
	}

	// Listen and serve connections
	logger.Debugf("PgProxy running.")
	pgProxy.Serve()

	// Wait for ongoing goroutines submitting log data
	wg.Wait()
}

func buildInfo() []string {
	return []string{
		fmt.Sprintf("Build Timestamp   : %s", buildTimestamp),
		fmt.Sprintf("Build GIT Commit  : %s", buildGitCommit[:8]),
		fmt.Sprintf("Build Go Version  : %s", buildGoVersion),
		fmt.Sprintf("Build OS/Arch     : %s", buildGoArch),
	}
}
