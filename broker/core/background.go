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
	"errors"
	"math"
	"time"

	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/broker/database"
	"github.com/siemens/Large-Scale-Discovery/broker/memory"
	brokerscopedb "github.com/siemens/Large-Scale-Discovery/broker/scopedb"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/vburenin/nsync"
)

var cleanupGracePeriod = time.Minute * 20 // Period after broker launch, in which timed out scans are not cleaned up to allow scan agents to save accumulated scan results
var cleanupInterval = time.Minute * 5     // Interval in which cleanup of exceeded scan tasks will execute
var submitInterval = time.Second * 6      // Interval in which scan agent stats will be transferred to the manager. Interval should be longer than agent request interval!
var rpcStatsLock = nsync.NewTryMutex()    // Named mutex to prevent parallel manager submits of scope stats

// backgroundTasks initializes and handles background tasks to be executed on a regular basis.
//   - scan scope changes: 	Respective scan scopes must be removed from the local memory. They will be re-loaded
//     from the manager the next time they are needed. In case of connectivity issues *all*
//     scan scopes should be wiped, because there might have been unobserved changes.
//   - sending scope stats: 	The broker is collecting stats about scan scopes and scan agents and storing
//     them in memory. This data needs to be written to the manager regularly
func backgroundTasks() {

	// Get tagged logger
	logger := log.GetLogger().Tagged("backgroundTasks")
	logger.Infof("Starting background tasks.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Initialize scan scope notifications
	chNotification, chNotificationReconnect := manager.RpcSubscribeNotification(logger, rpcClient, coreCtx)

	// Initialize scan task cleanup interval
	chCleanExceeded := time.NewTicker(cleanupInterval)
	defer chCleanExceeded.Stop()

	// Initialize scope stats submit interval
	chSubmitStats := time.NewTicker(submitInterval)
	defer chSubmitStats.Stop()

	// Loop and handle background tasks until context expires
	for {

		// Keep waiting for a background event to handle
		select {

		// Cancellation signal
		case <-coreCtx.Done():

			// Return nil as everything went fine
			logger.Debugf("Shutting down.")
			return

		// Submit the latest scan scope stats to the manager
		case <-chSubmitStats.C:

			// Log action
			logger.Debugf("Submitting latest scope stats.")

			// Launch action in background. Wait for it on shutdown.
			wgBackgroundTask.Add(1)
			go func() {
				defer wgBackgroundTask.Done()
				submitStats(logger)
			}()

		// Submit the latest scan scope stats to the manager
		case <-chCleanExceeded.C:

			// Skip cleanup if broker is not yet up for a while. If the broker was down, we want to give scan
			// agents a chance to return accumulated/cached scan results.
			if time.Since(startupTime) < cleanupGracePeriod {
				logger.Debugf("Skipping cleanup of exceeded scan tasks due to grace period.")
			} else {
				logger.Debugf("Cleaning exceeded scan tasks.")
				cleanExceeded(logger)
			}

		// Clear notified scan scope from memory
		case notification, ok := <-chNotification:

			// Stop when the notification subscriber terminates
			if !ok {
				return
			}

			// Log action
			logger.Debugf("Cleaning outdated data from memory.")

			// Execute cleanup of changed scan scopes asynchronously, so that we can directly
			// go back to listening for further changes. Wait for it on shutdown.
			wgBackgroundTask.Add(1)
			go func(notification manager.ReplyNotification) {
				defer wgBackgroundTask.Done()
				cleanMemory(logger, notification)
			}(notification)

		// Clear all scan scopes from memory, there might have been notifications missed
		case _, ok := <-chNotificationReconnect:

			// Stop when the notification subscriber terminates
			if !ok {
				return
			}

			// Log action
			logger.Debugf("Removing all scan scopes from memory.")

			// Clear scopes from memory
			memory.ClearScopes()
		}
	}
}

// cleanMemory cleans scan stats and scope data of outdated/removed scan scopes from memory
func cleanMemory(logger scanUtils.Logger, notification manager.ReplyNotification) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Clean remaining scan targets of faded scan scopes form broker database
	errDel := database.CleanScopeTargets(notification.RemainingScopeIds)
	if errDel != nil {
		logger.Warningf("Could not clean remaining scan targets of faded scan scopes: %s", errDel)
	}

	// Get map of cached scan scopes for cleanup
	scopeEntries := memory.GetScopes()

	// Remove updated scan scopes from memory to force on-demand reload with latest data
	for _, updateScopeId := range notification.UpdateScopeIds {
		scopeEntry, cached := scopeEntries[updateScopeId]
		if cached {
			memory.RemoveScope(scopeEntry.Secret)
			logger.Infof("Scan scope '%s' (ID %d) removed from memory.", scopeEntry.Name, scopeEntry.Id)
		}
	}

	// Clean faded scan scopes form memory
	for scopeEntryId, scopeEntry := range scopeEntries {
		if !utils.Uint64Contained(scopeEntryId, notification.RemainingScopeIds) {
			memory.RemoveScope(scopeEntry.Secret)
			logger.Infof(
				"Scan scope '%s' (ID %d) cleaned from memory.", scopeEntry.Name, scopeEntry.Id)
		}
	}

	// Clean scan agent stats of removed scan scopes from memory
	for _, agentStat := range memory.GetAgents() {
		if !utils.Uint64Contained(agentStat.IdTScanScope, notification.RemainingScopeIds) {
			memory.RemoveAgent(agentStat.Name, agentStat.Host, "", agentStat.IdTScanScope) // Don't use IP for identification as it might be a dynamic one
			logger.Infof(
				"Scan agent stats of '%s-%s-%s' removed from memory.",
				agentStat.Name,
				agentStat.Host,
				agentStat.Ip,
			)
		}
	}
}

// cleanExceeded removes exceeded broker targets and marks their scope-database entries failed
func cleanExceeded(logger scanUtils.Logger) {

	// Define cleanup procedure
	fnCleanup := func(scopeEntry *managerdb.T_scan_scope) {

		// Log potential panics before letting them move on
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
				panic(r)
			}
		}()

		// Open the scope database through the shared manager package
		scopeDb, errScopeDb := scopedb.GetScopeDbHandle(logger, scopeEntry)
		if errScopeDb != nil {
			logger.Errorf("Could not open scope database: %s", errScopeDb)
			return
		}

		// Log cleaning of exceeded module targets
		logger.Debugf("Cleaning exceeded '%s' scans.", discovery.Label)

		// Read '--min-hostgroup' value
		nmapHostgroup := extractMinHostgroup(scopeEntry.ScanSettings.DiscoveryNmapArgs)
		if nmapHostgroup == 0 {
			nmapHostgroup = defaultNmapMinHostgroup
		}

		// Extract '--host-timeout' value
		nmapHosttimeoutMinutes := extractHostTimeoutMinutes(scopeEntry.ScanSettings.DiscoveryNmapArgs)
		if nmapHosttimeoutMinutes == 0 {
			nmapHosttimeoutMinutes = defaultNmapHosttimeoutMinutes
		}

		// Clean exceeded discovery scan inputs in scope database, setting status and scan_finished timestamp
		countDiscovery, errCountDiscovery := brokerscopedb.CleanExceededDiscovery(
			scopeDb,
			nmapHostgroup,
			int(math.Ceil(nmapHosttimeoutMinutes)),
		)
		if errCountDiscovery != nil {
			logger.Warningf(
				"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from scope database: %s",
				discovery.Label,
				scopeEntry.Name,
				scopeEntry.Id,
				errCountDiscovery,
			)
		} else if countDiscovery > 0 {
			logger.Infof(
				"Set %d exceeded '%s' scans of scope '%s' (ID %d) to '%s' in scope database.",
				countDiscovery,
				discovery.Label,
				scopeEntry.Name,
				scopeEntry.Id,
				scanUtils.StatusFailed,
			)
		}

		// Iterate scan modules to clean exceeded scans
		for _, label := range []string{
			banner.Label,
			nfs.Label,
			smb.Label,
			ssl.Label,
			ssh.Label,
			webcrawler.Label,
			webenum.Label,
		} {

			// Log cleaning of exceeded module targets
			logger.Debugf("Cleaning exceeded '%s' scans.", label)

			// Get max scan time configured for scan module
			var maxScanTimeMinutes int
			switch label {
			case banner.Label:
				maxScanTimeMinutes = 20 // Banner scan module does not have a configurable timeout because it should not take long
			case nfs.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.NfsScanTimeoutMinutes
			case nuclei.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.NucleiScanTimeoutMinutes
			case smb.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.SmbScanTimeoutMinutes
			case ssl.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.SslScanTimeoutMinutes
			case ssh.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.SshScanTimeoutMinutes
			case webcrawler.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.WebcrawlerScanTimeoutMinutes
			case webenum.Label:
				maxScanTimeMinutes = scopeEntry.ScanSettings.WebenumScanTimeoutMinutes
			default:
				logger.Errorf("Invalid module label '%s'.", label)
				continue
			}

			// Calculate threshold for scans that ran into their timeout but didn't return results
			maxScanTime := time.Minute * time.Duration(maxScanTimeMinutes)
			maxScanTime = maxScanTime + (time.Minute * 10) // Add buffer for potential networking/processing delays
			startedBefore := time.Now().Add(-maxScanTime)

			// Update exceeded module scans in scope database, setting status and scan_finished timestamp
			countScopeDb, errCountScopeDb := brokerscopedb.CleanExceeded(scopeDb, label, startedBefore)
			if errCountScopeDb != nil {
				logger.Warningf(
					"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from scope database: %s",
					label,
					scopeEntry.Name,
					scopeEntry.Id,
					errCountScopeDb,
				)
			} else if countScopeDb > 0 {
				logger.Infof(
					"Set %d exceeded '%s' scans of scope '%s' (ID %d) to '%s' in scope database.",
					countScopeDb,
					label,
					scopeEntry.Name,
					scopeEntry.Id,
					scanUtils.StatusFailed,
				)
			}

			// Remove exceeded module inputs from broker database
			countBrokerDb, errCountBrokerDb := database.CleanExceeded(scopeEntry, label, startedBefore)
			if errCountBrokerDb != nil {
				logger.Warningf(
					"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from broker database: %s",
					label,
					scopeEntry.Name,
					scopeEntry.Id,
					errCountBrokerDb,
				)
			} else if countBrokerDb > 0 {
				logger.Infof(
					"Removed %d exceeded '%s' scans of scope '%s' (ID %d) from broker database.",
					countBrokerDb,
					label,
					scopeEntry.Name,
					scopeEntry.Id,
				)
			}
		}
	}

	// Iterate and clean all scan scopes. Wait for it on shutdown.
	for _, scopeEntry := range memory.GetScopes() {
		wgBackgroundTask.Add(1)
		go func(scopeEntry managerdb.T_scan_scope) {
			defer wgBackgroundTask.Done()
			fnCleanup(&scopeEntry)
		}(scopeEntry)
	}
}

// submitStats queries current scan scope progress and submits it together with other scan scope sats to the manager.
// There it will be persisted and made available for other components, e.g. the web backend for display.
func submitStats(logger scanUtils.Logger) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Acquire lock if it is not yet taken
	acquired := rpcStatsLock.TryLockTimeout(submitInterval) // Let second run wait but don't pile up multiples

	// Execute if lock not already ongoing
	if acquired {

		// Make sure lock is released again
		defer rpcStatsLock.Unlock()

		// Get copy of map of agent stats
		agents := memory.GetAgents()

		// Clear memory so that new entries can accumulate already for the next execution
		// and to minimize possible race conditions of agent requests between GetAgents() and ClearAgents().
		// Memory needs to be cleared to avoid sending stats for old outdated scan agents that don't exist anymore.
		memory.ClearAgents()

		// Group scan agent stats by scan scope ID
		agentsByScope := make(map[uint64][]managerdb.T_scan_agent)
		for _, agent := range agents {
			agentsByScope[agent.IdTScanScope] = append(agentsByScope[agent.IdTScanScope], agent)
		}

		// Get known scan scopes from memory
		scopeEntries := memory.GetScopes()

		// Collect per-module queue stats for each scope
		statsByScope := make(map[uint64]utils.JsonMap)
		for scopeEntryId := range scopeEntries {
			scopeStats, errScopeStats := database.GetScopeStats(scopeEntryId)
			if errScopeStats != nil {
				logger.Warningf("Could not query scope stats of scan scope %d: %s", scopeEntryId, errScopeStats)
				continue
			}
			statsByScope[scopeEntryId] = scopeStats
		}

		// Send updated agent and queue stats to manager via RPC. Additional scan progress stats will be queried
		// and attached by the manager during this call.
		errRpc := manager.RpcUpdateScopeStats(logger, rpcClient, agentsByScope, statsByScope)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			logger.Debugf("Skipped submitting scan stats, due to connectivity issues.")
			return
		} else if errRpc != nil {
			logger.Warningf("Could not submit scan stats: %s", errRpc)
			return
		}
	}
}
