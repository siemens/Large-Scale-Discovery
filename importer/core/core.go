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
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/importer/config"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/vburenin/nsync"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Importer context and context cancellation function. Importer should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests
var rpcScopeLock = nsync.NewNamedMutex()                                  // Named mutex to prevent parallel synchronization for the same scan scope
var scheduler *gocron.Scheduler                                           // Scheduler with synchronization tasks that run in intervals (remote scan targets), rather than on-change (custom scan targets)

func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Initialize scheduler for synchronizations that do not run on-change, but in certain intervals (e.g. scan
	// scopes with remote data sources, where changes happen unattended)
	scheduler = gocron.NewScheduler(time.UTC)

	// Start scheduler
	scheduler.StartAsync()

	// Initialize registered importers
	for type_, importer := range importers {
		errImporter := importer.Init(conf.Importer)
		if errImporter != nil {
			return fmt.Errorf("importer initialization '%s' failed: %s", type_, errImporter)
		}
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

	// Register gob structures that will be sent via interface{}
	manager.RegisterGobs()

	// Initialize RPC client manager facing
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, rpcRemoteCrt)

	// Connect to manager but don't wait to start answering client requests. Connection attempts continue in background.
	_ = rpcClient.Connect(logger, true)

	// Return as everything went fine
	return nil
}

// Run starts the gin web server
func Run() error {

	// Only launch if shutdown is not already in progress
	select {
	case <-coreCtx.Done():
		return nil
	default:
	}

	// Get tagged logger
	logger := log.GetLogger()

	// Sync all scan scopes
	synchronizeAll() // Sync all scan scopes

	// Subscribe for notifications
	notification, notificationReconnect := manager.RpcSubscribeNotification(logger, rpcClient, coreCtx)

	// Keep waiting for scope changes on the manager to check whether synchronization is necessary
	for {
		select {
		case <-coreCtx.Done(): // Cancellation signal
			// Return nil as everything went fine
			return nil
		case <-notificationReconnect:
			synchronizeAll() // Sync all scan scopes
		case notificationReply := <-notification:

			// Update synchronization of changed scan scopes
			for _, updateScopeId := range notificationReply.UpdateScopeIds {
				synchronize(updateScopeId)
			}

			// Remove synchronization of deleted scan scopes
			for _, job := range scheduler.Jobs() {
				for _, tag := range job.Tags() {
					id, errId := strconv.ParseUint(tag, 10, 64)
					if errId != nil {
						logger.Errorf("Could not parse scheduler job tag to scope ID '%s'.", tag)
					}
					if !utils.Uint64Contained(id, notificationReply.RemainingScopeIds) {
						logger.Infof("Un-scheduling deleted scan scope '%d'.", id)
						_ = scheduler.RemoveByTag(tag)
					}
				}
			}
		}
	}
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Close agent context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Stop scheduler
		scheduler.Stop()

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}

		// Close the scope db connections.
		errs := managerdb.CloseScopeDbs()
		for _, err := range errs {
			logger.Errorf("Could not close scope db connection: %s", err)
		}
	})
}

// synchronizeAll requests all existing scan scopes from the manager to synchronize the scan scope input targets of
// each of them where necessary
func synchronizeAll() {

	// Get tagged logger
	logger := log.GetLogger()

	// Log action
	logger.Debugf("Synchronizing all scan scopes.")

	// Request all scan scopes from manager
	scanScopes, errScanScopes := manager.RpcGetScopes(logger, rpcClient)
	if errors.Is(errScanScopes, utils.ErrRpcConnectivity) {
		logger.Debugf("Could not sync scan scopes, manager not reachable.")
		return // Will be retried when manager is reachable again
	} else if errScanScopes != nil {
		logger.Warningf("Could not get scan scopes from manager: %s", errScanScopes)
		return
	}

	// Update each scan scopes
	for _, scanScope := range scanScopes {
		synchronize(scanScope.Id)
	}
}

// synchronize requests a specific scan scopes from the manager to synchronize its scope input targets if necessary
func synchronize(scopeId uint64) {

	// Get tagged logger
	logger := log.GetLogger()

	// Log action
	logger.Debugf("Processing scan scope (ID %d).", scopeId)

	// Request scope details from manager
	scanScope, errScanScopes := manager.RpcGetScope(logger, rpcClient, scopeId)
	if errors.Is(errScanScopes, utils.ErrRpcConnectivity) {
		logger.Debugf("Could not sync scan scope %d, manager not reachable.", scopeId)
		return // Will be retried when manager is reachable again
	} else if errScanScopes != nil {
		logger.Warningf("Could not get scan scope %d from manager: %s", scopeId, errScanScopes)
		return
	}

	// Decide whether to synchronize now or to update scheduled synchronization
	if scanScope.Type == "custom" {

		// Custom scan scopes don't need to be synchronized, they are maintained directly by the web backend via
		// the scope manager
		logger.Debugf("Skipping 'custom' scan scope '%s' (ID %d).", scanScope.Name, scanScope.Id)

	} else {

		// Update scheduled synchronization
		scheduleSynchronizeScanScope(scanScope)
	}
}

// scheduleSynchronizeScanScope schedules (or re-schedules) a certain scan scope for synchronization. Scan scopes
// populated with input targets from remote sources need to be schedules, as changes might have occurred.
func scheduleSynchronizeScanScope(scanScope managerdb.T_scan_scope) {

	// Get tagged logger
	logger := log.GetLogger()

	// Check if scan scope has suitable importer
	_, okImporter := importers[scanScope.Type]
	if !okImporter {
		logger.Warningf("Skipping scan scope '%s' (ID %d) of unknown type '%s'.", scanScope.Name, scanScope.Id, scanScope.Type)
		return
	}

	// Run immediate synchronization if scan scope is new
	if scanScope.LastSync.IsZero() {
		logger.Debugf("Initializing scan scope '%s' (ID %d).", scanScope.Name, scanScope.Id)
		go synchronizeScanScope(scanScope)
	}

	// Prepare unique tag for this scan scope
	tag := strconv.FormatUint(scanScope.Id, 10)

	// Remove potentially existing task
	_ = scheduler.RemoveByTag(tag)

	// Check if scan scope has sync flag
	val, okVal := scanScope.Attributes["sync"]
	valBool, okType := val.(bool)
	if !okVal || !okType || !valBool {
		logger.Infof("Synchronization of scan scope '%s' (ID %d) disabled.", scanScope.Name, scanScope.Id)
		return
	}

	// Log action
	logger.Infof("Scheduling synchronization of scan scope '%s' (ID %d).", scanScope.Name, scanScope.Id)

	// Prepare task to execute (scheduler cannot take arguments directly)
	task := func() {
		synchronizeScanScope(scanScope)
	}

	// Schedule synchronization interval and task
	_, errSchedule := scheduler.Every(1).Saturday().At("18:00").Tag(tag).Do(task)
	if errSchedule != nil {
		logger.Errorf(
			"Could not schedule scan task '%s' (ID %d) for synchronization.", scanScope.Name, scanScope.Id)
	}
}

// synchronizeScanScope synchronize a given scan scope's input targets if necessary. The synchronization of a given
// scan scope will be skipped if it is already ongoing.
func synchronizeScanScope(scanScope managerdb.T_scan_scope) {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(uuid)

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Acquire lock if it is not yet taken. Allow parallel requests for different scope secrets.
	scopeIdAsStr := strconv.FormatUint(scanScope.Id, 10)
	acquired := rpcScopeLock.TryLock(scopeIdAsStr)

	// Proceed with scan scope synchronization if it was not already in progress
	if !acquired {
		logger.Infof("Synchronizing scan scope '%s' (ID %d) already in progress.", scanScope.Name, scanScope.Id)
		return
	}

	// Make sure lock is released again
	defer rpcScopeLock.Unlock(scopeIdAsStr)

	// Get importer to user for this scan scope
	importer, okImporter := importers[scanScope.Type]
	if !okImporter {
		logger.Warningf("Skipping unknown scan scope type '%s'.", scanScope.Type)
		return
	}

	// Log action
	logger.Infof("Synchronizing scan scope '%s' (ID %d).", scanScope.Name, scanScope.Id)

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Infof(
			"Synchronizing scan scope '%s' (ID %d) took %s.", scanScope.Name, scanScope.Id, time.Since(start))
	}()

	// Transform JSON data from scan scope attributes to filters struct
	filters, errFilters := ParseFilters(scanScope.Attributes)
	if errFilters != nil {
		logger.Errorf(
			"Could not extract scan scope filters from scan scope attributes for scan scope '%s' (ID %d): %s",
			scanScope.Name,
			scanScope.Id,
			errFilters,
		)
		return
	}

	// Get the latest list of scan inputs
	targets, errTargets := importer.Import(filters)
	if errTargets != nil {
		logger.Errorf(
			"Could not load inputs for scan scope '%s' (ID %d): %s", scanScope.Name, scanScope.Id, errTargets)
		return
	}
	logger.Infof("%5d input targets received from '%s'.", len(targets), scanScope.Type)

	// Deploy scope targets in scopedb via manager. The manager will update the scan scope targets in the
	// background and return an RPC response immediately (if blocking=false).
	// ATTENTION: Another targets update for the same scan scope will fail until the previous one is completed.
	rpcResult, errRpc := manager.RpcUpdateScopeTargets(
		logger,
		rpcClient,
		scanScope.Id,
		targets,
		true,
	)
	if errRpc != nil && errRpc.Error() == manager.ErrScopeUpdateOngoing.Error() { // Errors received from RPC lose their original type!!
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d), previous synchronization still ongoing.", scanScope.Name, scanScope.Id)
		return
	} else if errors.Is(errRpc, utils.ErrRpcConnectivity) {
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d), manager not reachable.", scanScope.Name, scanScope.Id)
		return
	} else if errRpc != nil {
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d): %s", scanScope.Name, scanScope.Id, errRpc)
		return
	}

	// Log some stats
	logger.Infof("%5d input targets created.", rpcResult.Created)
	logger.Infof("%5d input targets removed.", rpcResult.Removed)
	logger.Infof("%5d input targets updated.", rpcResult.Updated)
}
