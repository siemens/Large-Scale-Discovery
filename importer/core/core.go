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
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/importer/config"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/vburenin/nsync"
)

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Importer context and cancellation function. Importer should terminate when the context is closed.
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests
var rpcScopeLock = nsync.NewNamedMutex()                                  // Named mutex to prevent parallel synchronization for the same scan scope
var scheduler gocron.Scheduler                                            // Scheduler with synchronization tasks that run in intervals (remote scan targets), rather than on-change (custom scan targets)

// Init initializes the importer and its dependencies
func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Initialize scheduler for synchronizations that do not run on-change, but in certain intervals (e.g. scan
	// scopes with remote data sources, where changes happen unattended)
	var errScheduler error
	scheduler, errScheduler = gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if errScheduler != nil {
		return fmt.Errorf("could not initialize scheduler: %w", errScheduler)
	}

	// Start scheduler
	scheduler.Start()

	// Initialize registered importers
	for type_, importer := range importers {
		errImporter := importer.Init(conf.Inventory)
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

	// Initialize RPC client manager facing.
	// The manager requires a shared secret to authorize the RPC connection.
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, conf.ManagerSsl, rpcRemoteCrt, conf.ManagerSecret)

	// Connect to manager but don't wait to start answering client requests. Connection attempt continues in background.
	_ = rpcClient.Connect(logger, true)

	// Return as everything went fine
	return nil
}

// Run synchronizes scan scopes until the importer context is terminated
func Run() error {

	// Only launch if shutdown is not already in progress
	if coreCtx.Err() != nil {
		return nil
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
		case _, ok := <-notificationReconnect:

			// Stop when the notification subscriber terminates
			if !ok {
				return nil
			}

			// Sync all scan scopes
			synchronizeAll()

		case notificationReply, ok := <-notification:

			// Stop when the notification subscriber terminates
			if !ok {
				return nil
			}

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
						logger.Infof("Unscheduling deleted scan scope %d.", id)
						scheduler.RemoveByTags(tag)
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

		// Close the importer context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Stop scheduler
		if scheduler != nil {
			errScheduler := scheduler.Shutdown()
			if errScheduler != nil {
				logger.Errorf("Could not shut down scheduler: %s", errScheduler)
			}
		}

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}

		// Close the scope db connections.
		errs := scopedb.CloseScopeDbs()
		for _, err := range errs {
			logger.Errorf("Could not close scope DB connection: %s", err)
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
	scopeEntries, errScopeEntries := manager.RpcGetScopes(logger, rpcClient)
	if errors.Is(errScopeEntries, utils.ErrRpcConnectivity) {
		logger.Debugf("Could not sync scan scopes, manager not reachable.")
		return // Will be retried when manager is reachable again
	} else if errScopeEntries != nil {
		logger.Warningf("Could not get scan scopes from manager: %s", errScopeEntries)
		return
	}

	// Update each scan scopes
	for _, scopeEntry := range scopeEntries {
		synchronize(scopeEntry.Id)
	}
}

// synchronize requests a specific scan scopes from the manager to synchronize its scope input targets if necessary
func synchronize(scopeId uint64) {

	// Get tagged logger
	logger := log.GetLogger()

	// Log action
	logger.Debugf("Processing scan scope (ID %d).", scopeId)

	// Request scope details from manager
	scopeEntry, errScopeEntry := manager.RpcGetScope(logger, rpcClient, scopeId)
	if errors.Is(errScopeEntry, utils.ErrRpcConnectivity) {
		logger.Debugf("Could not sync scan scope %d, manager not reachable.", scopeId)
		return // Will be retried when manager is reachable again
	} else if errScopeEntry != nil {
		logger.Warningf("Could not get scan scope %d from manager: %s", scopeId, errScopeEntry)
		return
	}

	// Decide whether to synchronize now or to update scheduled synchronization
	if scopeEntry.Type == "custom" {

		// Custom scan scopes don't need to be synchronized, they are maintained directly by the web backend via
		// the scope manager
		logger.Debugf("Skipping 'custom' scan scope '%s' (ID %d).", scopeEntry.Name, scopeEntry.Id)

	} else {

		// Update scheduled synchronization
		scheduleSynchronizeScanScope(scopeEntry)
	}
}

// scheduleSynchronizeScanScope schedules (or re-schedules) a certain scan scope for synchronization. Scan scopes
// populated with input targets from remote sources need to be schedules, as changes might have occurred.
func scheduleSynchronizeScanScope(scopeEntry managerdb.T_scan_scope) {

	// Get tagged logger
	logger := log.GetLogger()

	// Check if scan scope has suitable importer
	_, okImporter := importers[scopeEntry.Type]
	if !okImporter {
		logger.Warningf("Skipping scan scope '%s' (ID %d) of unknown type '%s'.", scopeEntry.Name, scopeEntry.Id, scopeEntry.Type)
		return
	}

	// Run immediate synchronization if scan scope is new
	if scopeEntry.LastSync.IsZero() {
		logger.Debugf("Initializing scan scope '%s' (ID %d).", scopeEntry.Name, scopeEntry.Id)
		go synchronizeScanScope(scopeEntry)
	}

	// Prepare unique tag for this scan scope
	tag := strconv.FormatUint(scopeEntry.Id, 10)

	// Remove potentially existing task
	scheduler.RemoveByTags(tag)

	// Check if scan scope has sync flag
	val, okVal := scopeEntry.Attributes["sync"]
	valBool, okType := val.(bool)
	if !okVal || !okType || !valBool {
		logger.Infof("Synchronization of scan scope '%s' (ID %d) disabled.", scopeEntry.Name, scopeEntry.Id)
		return
	}

	// Log action
	logger.Infof("Scheduling synchronization of scan scope '%s' (ID %d).", scopeEntry.Name, scopeEntry.Id)

	// Prepare task to execute (scheduler cannot take arguments directly)
	task := func() {
		synchronizeScanScope(scopeEntry)
	}

	// Schedule synchronization interval and task
	_, errSchedule := scheduler.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Saturday), gocron.NewAtTimes(gocron.NewAtTime(18, 0, 0))),
		gocron.NewTask(task),
		gocron.WithTags(tag),
	)
	if errSchedule != nil {
		logger.Errorf(
			"Could not schedule scan task '%s' (ID %d) for synchronization.", scopeEntry.Name, scopeEntry.Id)
	}
}

// synchronizeScanScope synchronize a given scan scope's input targets if necessary. The synchronization of a given
// scan scope will be skipped if it is already ongoing.
func synchronizeScanScope(scopeEntry managerdb.T_scan_scope) {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(uuid)

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Acquire lock if it is not yet taken. Allow parallel requests for different scope secrets.
	scopeIdAsStr := strconv.FormatUint(scopeEntry.Id, 10)
	acquired := rpcScopeLock.TryLock(scopeIdAsStr)

	// Proceed with scan scope synchronization if it was not already in progress
	if !acquired {
		logger.Infof("Synchronizing scan scope '%s' (ID %d) already in progress.", scopeEntry.Name, scopeEntry.Id)
		return
	}

	// Make sure lock is released again
	defer rpcScopeLock.Unlock(scopeIdAsStr)

	// Get importer to user for this scan scope
	importer, okImporter := importers[scopeEntry.Type]
	if !okImporter {
		logger.Warningf("Skipping unknown scan scope type '%s'.", scopeEntry.Type)
		return
	}

	// Log action
	logger.Infof("Synchronizing scan scope '%s' (ID %d).", scopeEntry.Name, scopeEntry.Id)

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Infof(
			"Synchronizing scan scope '%s' (ID %d) took %s.", scopeEntry.Name, scopeEntry.Id, time.Since(start))
	}()

	// Transform JSON data from scan scope attributes to filters struct
	filters, errFilters := ParseFilters(scopeEntry.Attributes)
	if errFilters != nil {
		logger.Errorf(
			"Could not extract scan scope filters from scan scope attributes for scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errFilters,
		)
		return
	}

	// Get the latest list of scan inputs
	targetEntries, errTargetEntries := importer.Import(filters, scopeEntry.SplitSize)
	if errTargetEntries != nil {
		logger.Errorf(
			"Could not load inputs for scan scope '%s' (ID %d): %s", scopeEntry.Name, scopeEntry.Id, errTargetEntries)
		return
	}
	logger.Infof("%5d input targets received from '%s'.", len(targetEntries), scopeEntry.Type)

	// Deploy scope targets in scopedb via manager. The manager will update the scan scope targets in the
	// background and return an RPC response immediately (if blocking=false).
	// ATTENTION: Another targets update for the same scan scope will fail until the previous one is completed.
	rpcResult, errRpcResult := manager.RpcUpdateScopeTargets(
		logger,
		rpcClient,
		scopeEntry.Id,
		targetEntries,
		true,
	)
	if errors.Is(errRpcResult, manager.ErrScopeUpdateOngoing) {
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d), another synchronization still ongoing.", scopeEntry.Name, scopeEntry.Id)
		return
	} else if errors.Is(errRpcResult, utils.ErrRpcConnectivity) {
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d), manager not reachable.", scopeEntry.Name, scopeEntry.Id)
		return
	} else if errRpcResult != nil {
		logger.Errorf("Could not synchronize scan scope '%s' (ID %d): %s", scopeEntry.Name, scopeEntry.Id, errRpcResult)
		return
	}

	// Log some stats
	logger.Infof("%5d input targets created.", rpcResult.Created)
	logger.Infof("%5d input targets removed.", rpcResult.Removed)
	logger.Infof("%5d input targets updated.", rpcResult.Updated)
}
