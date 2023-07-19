/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/vburenin/nsync"
	"large-scale-discovery/_build"
	"large-scale-discovery/broker/brokerdb"
	"large-scale-discovery/broker/config"
	"large-scale-discovery/broker/memory"
	"large-scale-discovery/log"
	manager "large-scale-discovery/manager/core"
	managerdb "large-scale-discovery/manager/database"
	"large-scale-discovery/utils"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Agent context and context cancellation function. Agent should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var startupTime time.Time                                                 // Time when the broker launched
var rpcServerCrt string                                                   // Certificate used by the listening RPC server
var rpcServerKey string                                                   // Key used by the listening RPC server
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests
var rpcScopeLock = nsync.NewNamedMutex()                                  // Named mutex to prevent parallel manager queries for the same scan scope

var errInvalidScopeSecret = fmt.Errorf("invalid secret")
var errScopeNotAvailable = fmt.Errorf("scan scope not available")

// Init initializes the broker and all of its parameters
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
	managerdb.SetConnectionName(fmt.Sprintf("Broker_%s_%s", localHostname, localIp))

	// Set maximum db connections according to configuration. This value must be below the amount configured in the
	// database itself, otherwise queries may fail. Also consider, that other applications might connect with multiple
	// parallel connections.
	managerdb.SetMaxConnectionsDefault(conf.DbConnections)

	// Prepare broker key and certificate path
	rpcServerCrt = filepath.Join("keys", "broker.crt")
	rpcServerKey = filepath.Join("keys", "broker.key")
	if _build.DevMode {
		rpcServerCrt = filepath.Join("keys", "broker_dev.crt")
		rpcServerKey = filepath.Join("keys", "broker_dev.key")
	}
	errServerCrt := scanUtils.IsValidFile(rpcServerCrt)
	if errServerCrt != nil {
		return errServerCrt
	}
	errServerKey := scanUtils.IsValidFile(rpcServerKey)
	if errServerKey != nil {
		return errServerKey
	}

	// Prepare RPC certificate path
	rpcRemoteCrt := filepath.Join("keys", "manager.crt")
	if _build.DevMode {
		rpcRemoteCrt = filepath.Join("keys", "manager_dev.crt")
	}
	errRemoteCrt := scanUtils.IsValidFile(rpcRemoteCrt)
	if errRemoteCrt != nil {
		return errRemoteCrt
	}

	// Initialize brokerdb
	errOpen := brokerdb.Init()
	if errOpen != nil {
		return fmt.Errorf("could not open brokerdb: %s", errOpen)
	}

	// Create or update the db tables
	errMigrate := brokerdb.AutoMigrate()
	if errMigrate != nil {
		return fmt.Errorf("could not migrate broker db: %s", errMigrate)
	}

	// Initialize RPC client manager facing
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, rpcRemoteCrt)

	// Connect to manager and wait for successful connection. Abort on shutdown request.
	success := rpcClient.Connect(logger, true)
	if !success {
		select {
		case <-coreCtx.Done():
			return nil
		case <-rpcClient.Established():
		}
	}

	// Launch background tasks to be executed on a regular basis, e.g.:
	// 	- listening for scope changes on the manager to update cached scope data
	//	- writing cached scan scope stats to the manager
	go backgroundTasks()

	// Set startup time. There is a certain time windows where scan agents can send accumulated scan results, before
	// timed out scans are cleaned up.
	startupTime = time.Now()

	// Register gob structures that will be sent via interface{}
	RegisterGobs()

	// Register broker as RPC receiver
	errRegister := rpc.Register(&Broker{})
	if errRegister != nil {
		return errRegister
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
		"Broker",
		rpcServerCrt,
		rpcServerKey,
		conf.ListenAddress,
	)
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Close agent context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}

		// Close the scope db connections
		errs := managerdb.CloseScopeDbs()
		for _, err := range errs {
			logger.Errorf("Could not close scope db connection: %s", err)
		}

		// Make sure broker db gets closed on exit
		errBroker := brokerdb.Close()
		if errBroker != nil {
			logger.Errorf("Could not close broker db connection: '%s'", errBroker)
		}
	})
}

// getScanScope retrieves a scan scope from the local memory, if stored. If not stored, an RPC request is sent to
// the manager. Only one goroutine at a time is allowed to request a certain scan scope from the manager. Parallel
// goroutines will wait for it, search memory again, and potentially try requesting from the manager again.
// If there is an RPC connection issue, the primary goroutine (requesting the scan scope from the manager) will keep
// trying, until the RPC request succeeded.
// If the scan scope was neither found in memory, nor on the manager, an errScopeNotAvailable error is returned.
// ATTENTION: Scan results can be dropped if a scan scope is unknown, but not if it just couldn't be retrieved!
// ATTENTION: Make sure client's don't pile up subsequent requests, as long as the previous ones didn't return!
func getScanScope(logger scanUtils.Logger, scopeSecret string) (*managerdb.T_scan_scope, error) {

	// Search scan scope in memory
	cachedScope, cached := memory.GetScope(scopeSecret)
	if cached {
		return &cachedScope, nil
	}

	// Acquire lock if it is not yet taken. Allow parallel requests for different scope secrets.
	acquired := rpcScopeLock.TryLock(scopeSecret)

	// If another goroutine is already requesting the scan scope, this one can wait and try again.
	if !acquired {

		// Log action
		logger.Debugf("Waiting for primary RPC request for scan scope '%s...'.", scopeSecret[0:5])

		// Wait for primary request to be completed. This will block the current goroutine until the primary RPC
		// request to the manager went through (waiting and re-trying on connectivity issues).
		rpcScopeLock.Lock(scopeSecret)

		// Make sure lock is released when done
		defer rpcScopeLock.Unlock(scopeSecret)

		// Check if scan scope is now available in memory, the primary RPC request might have retrieved it.
		cachedScope, cached = memory.GetScope(scopeSecret)
		if cached {
			return &cachedScope, nil
		}

		// Return error because scan scope secret is unknown
		logger.Infof("Invalid scope secret '%s...'.", scopeSecret[0:5])
		return nil, errInvalidScopeSecret
	}

	// Make sure lock is released when done
	defer rpcScopeLock.Unlock(scopeSecret)

	// Log action
	logger.Debugf("Requesting scan scope '%s...'.", scopeSecret[0:5])

	// Get config
	conf := config.GetConfig()

	// Prepare memory for RPC result
	var scanScope managerdb.T_scan_scope
	var errRpc error

	// Keep trying until RPC request (connection) succeeded. Subsequent RPC requests for the same scan scope
	// will block and wait for this to complete.
	for {

		// Request scope with full details from manager
		scanScope, errRpc = manager.RpcGetScopeFull(logger, rpcClient, conf.ManagerPrivilegeSecret, scopeSecret)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {

			// Wait for RPC re-connection and retry
			select {
			case <-coreCtx.Done():
				return nil, errScopeNotAvailable // Abort, due to shutdown request
			case <-rpcClient.Established():
				continue // Retry
			}

		} else if errRpc != nil && errRpc.Error() == manager.ErrInvalidPrivilege.Error() { // Errors received from RPC lose their original type!!
			return nil, errScopeNotAvailable
		} else if errRpc != nil {
			return nil, errScopeNotAvailable // Was already reported by the rpcClient.
		}

		// Abort if used scan scope secret didn't return a valid scan scope. The manager does not return an error if
		// the scope secret was invalid, because it would trigger a critical log message in the RPC client, although
		// it's legit.
		if scanScope.Id == 0 {
			logger.Infof("Invalid scope secret '%s...'.", scopeSecret[0:5])
			return nil, errInvalidScopeSecret
		}

		// Add loaded scan scope to memory for future usage
		logger.Debugf("Loaded scan scope '%s' (ID %d).", scanScope.Name, scanScope.Id)
		memory.AddScope(scopeSecret, scanScope)

		// Return scan scope as retrieved from the manager
		return &scanScope, nil
	}
}

// checkCycle checks whether a cycle has completed and triggers the initialization of a new one
func checkCycle(logger scanUtils.Logger, scanScope *managerdb.T_scan_scope) (initialized bool, err error) {

	// Skip cycle check for the first few minutes after startup because agents might not have connected
	// and the amount of currently active scan scope discovery tasks is unclear.
	if time.Since(startupTime) < cleanupGracePeriod {
		logger.Debugf("Skipping new cycle check due to grace period.")
		return false, nil
	}

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		return false, fmt.Errorf("could not open scope database: %s", errHandle)
	}

	// Check if there is anything in this scope yet
	var contentCount int64
	errDb := scopeDb.Model(&managerdb.T_discovery{}).
		Limit(1). // Fastest way to find out if there is something in the table
		Count(&contentCount).Error
	if errDb != nil {
		return false, fmt.Errorf("could not count total targets in scope db: %s", errDb)
	}

	// Skip initialization if there is nothing to initialize
	if contentCount == 0 {
		logger.Debugf("Current scan cycle has no '%s' scan targets yet.", discovery.Label)
		return false, nil
	}

	// Count active discovery scan tasks of the current scan scope
	activeTasks := 0
	recentAgents := memory.GetAgents()
	for _, recentAgent := range recentAgents {

		// Only check scan scope's agents
		if recentAgent.IdTScanScope == scanScope.Id {

			// Iterate scan task modules
			for module, moduleCount := range recentAgent.Tasks {

				// Count active discovery task
				if module == discovery.Label {
					activeTasks += moduleCount.(int)
				}
			}
		}
	}

	// Count remaining scan targets in the current scan cycle
	var remainingTasks int64
	errDb2 := scopeDb.Model(&managerdb.T_discovery{}).
		Where("scan_started IS NULL"). // Not active elements only
		Where("enabled IS TRUE").      // Enabled elements only
		Limit(1).                      // Finding a single case is already enough to decide
		Count(&remainingTasks).Error
	if errDb2 != nil {
		return false, fmt.Errorf("could not count remaining targets in scope db: %s", errDb2)
	}

	// Skip initialization if there are discovery inputs left
	if remainingTasks > 0 {
		logger.Debugf("Current scan cycle has remaining '%s' scan targets.", discovery.Label)
		return false, nil
	}

	// Count incomplete scan targets in the current scan cycle
	var incompleteTasks int64
	errDb3 := scopeDb.Model(&managerdb.T_discovery{}).
		Where("scan_finished IS NULL"). // Not active elements only
		Where("enabled IS TRUE").       // Enabled elements only
		Limit(1).                       // Finding a single case is already enough to decide
		Count(&incompleteTasks).Error
	if errDb3 != nil {
		return false, fmt.Errorf("could not count incomplete targets in scope db: %s", errDb3)
	}

	// Skip initialization if there are incomplete targets and active tasks.
	// After a certain timeout an incomplete target will be marked as completed by the background task cleanExceeded(),
	// so the scan scope can also enter a new scan cycle if one or more scan agents stop responding and are stuck in
	// memory with an invalid discovery scan task count greater 0.
	if incompleteTasks > 0 && activeTasks > 0 {
		logger.Debugf(
			"Current scan cycle has incomplete '%s' scan targets and active scan tasks.", discovery.Label)
		return false, nil
	}

	// Log cycle reset
	logger.Infof("Initiating new scan cycle for '%s' (ID: %d).", scanScope.Name, scanScope.Id)

	// Execute scan cycle reset via manager
	errNewCycle := manager.RpcNewCycle(logger, rpcClient, scanScope.Id)
	if errNewCycle != nil {
		return false, fmt.Errorf(
			"could not initialize new cycle for scope '%s' (ID %d): %s",
			scanScope.Name,
			scanScope.Id,
			errNewCycle,
		)
	}

	// Return initialized flag or error
	return true, nil
}
