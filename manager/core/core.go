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
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/manager/config"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const ScopeChangeNotifierInterval = time.Second * 1

var scopeChangeNotifier *utils.Notifier
var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Agent context and context cancellation function. Agent should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var scopedbPrepareLock sync.Mutex                                         // A lock making sure that the scope db preparation is not interrupted by a shutdown request
var rpcServerCrt string                                                   // Certificate used by the listening RPC server
var rpcServerKey string                                                   // Key used by the listening RPC server

// Init initializes the manager and all of its parameters
func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Lookup agent IP
	localIp, errIp := utils.GetLocalIp()
	if errIp != nil {
		return fmt.Errorf("could not read local ip: %s", errIp)
	}

	// Lookup agent hostname
	localHostname, errHostname := os.Hostname()
	if errHostname != nil {
		return fmt.Errorf("could not read local hostname: %s", errHostname)
	}

	// Set the name to use in DB connections
	database.SetConnectionName(fmt.Sprintf("Manager_%s_%s", localHostname, localIp))

	// Set maximum db connections according to configuration. This value must be below the amount configured in the
	// database itself, otherwise queries may fail. Also consider, that other applications might connect with multiple
	// parallel connections.
	database.SetMaxConnectionsDefault(conf.Database.Connections)

	// Prepare manager key and certificate path
	rpcServerCrt = filepath.Join("keys", "manager.crt")
	rpcServerKey = filepath.Join("keys", "manager.key")
	if _build.DevMode {
		rpcServerCrt = filepath.Join("keys", "manager_dev.crt")
		rpcServerKey = filepath.Join("keys", "manager_dev.key")
	}
	errCrt := scanUtils.IsValidFile(rpcServerCrt)
	if errCrt != nil {
		return errCrt
	}
	errKey := scanUtils.IsValidFile(rpcServerKey)
	if errKey != nil {
		return errKey
	}

	// Initialize manager database
	errDb := database.OpenManagerDb()
	if errDb != nil {
		return fmt.Errorf("could not initialize manager db: %s", errDb)
	}

	// Wrap database prepare code into function to make use of local defer statement
	if errScopedbPrepare := func() error {

		// Acquire DB prepare lock to prevent "Shutdown()" from interrupting
		scopedbPrepareLock.Lock()
		defer scopedbPrepareLock.Unlock()

		// Write development sample data to the database
		if _build.DevMode {
			logger.Infof("Deploying sample data.")
			errDev := database.XDeploySampleData(logger, conf.ScanDefaults)
			if errDev != nil {
				return fmt.Errorf("could not write development sample data: %s", errDev)
			}
			logger.Infof("Deploying sample data completed.")
		}

		// Automigrate all scan scope tables
		errScopeDb := database.AutomigrateScanScopes(logger)
		if errScopeDb != nil {
			return fmt.Errorf("could not initialize scope db: %s", errScopeDb)
		}

		// Return nil as everything went fine
		return nil
	}(); errScopedbPrepare != nil {
		return errScopedbPrepare
	}

	// Initialize scope change notifier that can be subscribed to in order to receive a signal about changed scopes
	scopeChangeNotifier = utils.NewNotifier(ScopeChangeNotifierInterval, func(a interface{}, b interface{}) bool {
		return a == b
	})

	// Register gob structures that will be sent via interface{}
	RegisterGobs()

	// Register manager as RPC receiver
	errRegister := rpc.Register(&Manager{})
	if errRegister != nil {
		return fmt.Errorf("could not register RPC service: %s", errRegister)
	}

	// Return nil as everything went fine
	return nil
}

// Run loops to accept and process RPC connections until the core context is terminated
func Run() error {

	// Only launch if shutdown is not already in progress
	select {
	case <-coreCtx.Done():
		return nil
	default:
	}

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Start serving RPC connections
	return utils.ServeRpc(
		logger,
		coreCtx,
		"Manager",
		rpcServerCrt,
		rpcServerKey,
		conf.ListenAddress,
	)
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Wait for scope DB preparation if currently running
		scopedbPrepareLock.Lock()
		defer scopedbPrepareLock.Unlock()

		// Shut down scope change notifier
		if scopeChangeNotifier != nil {
			scopeChangeNotifier.Shutdown()
		}

		// Close agent context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Close the scope db connections.
		errsScope := database.CloseScopeDbs()
		for _, err := range errsScope {
			logger.Errorf("Could not close scope db connection: '%s'", err)
		}

		// Make sure db gets closed on exit
		errManager := database.CloseManagerDb()
		if errManager != nil {
			logger.Errorf("Could not close manager db connection: '%s'", errManager)
		}
	})
}
