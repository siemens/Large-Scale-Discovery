/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"context"
	"fmt"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/manager/config"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

// scopeChangeNotifierInterval defines how often scope-change notifications are evaluated
const scopeChangeNotifierInterval = time.Second * 1

var scopeChangeNotifier *utils.Notifier                                   // Batches scan scope change notifications
var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Manager context and cancellation function. Manager should terminate when the context is closed.
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var rpcAdmissionLock sync.Mutex                                           // Serializes RPC admission with shutdown
var rpcInflight sync.WaitGroup                                            // Tracks RPC handlers admitted before shutdown, and background tasks they launch
var scopeDbPrepareLock sync.Mutex                                         // A lock making sure that the scope database preparation is not interrupted by a shutdown request
var rpcServerCrt string                                                   // Certificate used by the listening RPC server
var rpcServerKey string                                                   // Key used by the listening RPC server

// Init initializes the manager and all of its parameters
func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Look up manager IP
	localIp, errLocalIp := utils.GetLocalIp()
	if errLocalIp != nil {
		return fmt.Errorf("could not read local ip: %s", errLocalIp)
	}

	// Look up manager hostname
	localHostname, errLocalHostname := os.Hostname()
	if errLocalHostname != nil {
		return fmt.Errorf("could not read local hostname: %s", errLocalHostname)
	}

	// Set the name to use in DB connections
	scopedb.SetConnectionName(fmt.Sprintf("Manager_%s_%s", localHostname, localIp))

	// Set maximum db connections according to configuration. This value must be below the amount configured in the
	// database itself, otherwise queries may fail. Also consider, that other applications might connect with multiple
	// parallel connections.
	scopedb.SetMaxConnectionsDefault(conf.Database.Connections)

	// Prepare manager key and certificate path
	if conf.ListenSsl {
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
	}

	// Initialize manager database
	errDb := database.Open()
	if errDb != nil {
		return fmt.Errorf("could not initialize manager database: %s", errDb)
	}

	// Wrap database prepare code into function to make use of local defer statement
	if errScopeDbPrepare := func() error {

		// Acquire DB prepare lock to prevent "Shutdown()" from interrupting
		scopeDbPrepareLock.Lock()
		defer scopeDbPrepareLock.Unlock()

		// Write development sample data to the database
		if _build.DevMode {
			logger.Infof("Deploying sample data.")
			errDev := XDeploySampleData(logger, conf.ScanDefaults)
			if errDev != nil {
				return fmt.Errorf("could not write development sample data: %s", errDev)
			}
			logger.Infof("Deploying sample data completed.")
		}

		// Automigrate all scan scope tables
		errScopeDb := scopedb.AutomigrateScanScopes(logger)
		if errScopeDb != nil {
			return fmt.Errorf("could not initialize scope database: %s", errScopeDb)
		}

		// Return nil as everything went fine
		return nil
	}(); errScopeDbPrepare != nil {
		return errScopeDbPrepare
	}

	// Initialize scope change notifier that can be subscribed to in order to receive a signal about changed scopes
	scopeChangeNotifier = utils.NewNotifier(scopeChangeNotifierInterval, func(a interface{}, b interface{}) bool {
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
	if coreCtx.Err() != nil {
		return nil
	}

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Start serving RPC connections. Each connection must present a valid shared secret before it is served, closing
	// the unauthenticated-RPC trust boundary.
	return utils.ServeRpc(
		logger,
		coreCtx,
		"Manager",
		rpcServerCrt,
		rpcServerKey,
		conf.ListenAddress,
		conf.ListenSecrets,
	)
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Close core context while holding the admission lock so no RPC handler can pass its shutdown check without
		// being registered in the wait group before shutdown starts waiting
		rpcAdmissionLock.Lock()
		coreCtxCancelFunc()
		rpcAdmissionLock.Unlock()

		// Wait for scope database preparation if currently running
		scopeDbPrepareLock.Lock()
		defer scopeDbPrepareLock.Unlock()

		// Shut down scope change notifier
		if scopeChangeNotifier != nil {
			scopeChangeNotifier.Shutdown()
		}

		// Wait for admitted RPC handlers and their background tasks before closing resources
		rpcInflight.Wait()

		// Close scope pools only after every borrower has released them
		errsScope := scopedb.CloseScopeDbs()
		for _, errClose := range errsScope {
			logger.Errorf("Could not close scope database connection: '%s'", errClose)
		}

		// Close the administrative database server connections
		errsServer := scopedb.CloseServerDbs()
		for _, errClose := range errsServer {
			logger.Errorf("Could not close database server connection: '%s'", errClose)
		}

		// Make sure db gets closed on exit
		errManager := database.Close()
		if errManager != nil {
			logger.Errorf("Could not close manager database connection: '%s'", errManager)
		}
	})
}

func beginRpc() bool {

	// Acquire admission lock and make sure it is released again
	rpcAdmissionLock.Lock()
	defer rpcAdmissionLock.Unlock()

	// Check core context validity
	if coreCtx.Err() != nil {
		return false
	}

	// Increase RPC inflight counter
	rpcInflight.Add(1)

	// Signal access
	return true
}
