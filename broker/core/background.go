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
	"errors"
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/broker/brokerdb"
	"github.com/siemens/Large-Scale-Discovery/broker/memory"
	"github.com/siemens/Large-Scale-Discovery/broker/scopedb"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/vburenin/nsync"
	"math"
	"time"
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
	logger := log.GetLogger().Tagged(fmt.Sprintf("backgroundTasks"))
	logger.Infof("Starting background tasks.")

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Initialize scan scope notifications
	chNotification, chNotificationReconnect := manager.RpcSubscribeNotification(logger, rpcClient, coreCtx)

	// Initialize scan task cleanup interval
	chCleanExceeded := time.NewTicker(cleanupInterval)

	// Initialize scope stats submit interval
	chSubmitStats := time.NewTicker(submitInterval)

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

			logger.Debugf("Submitting latest scope stats.")
			go submitStats(logger)

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
		case notification := <-chNotification:

			// Execute cleanup of changed scan scopes asynchronously, so that we can directly go back to listening
			// for further changes...
			logger.Debugf("Cleaning outdated data from memory.")
			go cleanMemory(logger, notification)

		// Clear all scan scopes from memory, there might have been notifications missed
		case <-chNotificationReconnect:
			logger.Debugf("Removing all scan scopes from memory.")
			memory.ClearScopes()
		}
	}
}

// cleanMemory cleans scan stats and scope data of outdated/removed scan scopes from memory
func cleanMemory(logger scanUtils.Logger, notification manager.ReplyNotification) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Clean remaining scan targets of faded scan scopes form brokerdb
	errDel := brokerdb.CleanScopeTargets(notification.RemainingScopeIds)
	if errDel != nil {
		logger.Warningf("Could not clean remaining scan targets of faded scan scopes: %s", errDel)
	}

	// Get map of cached scan scopes for cleanup
	cachedScopes := memory.GetScopes()

	// Remove updated scan scopes from memory to force on-demand reload with latest data
	for _, updateScopeId := range notification.UpdateScopeIds {
		scanScope, cached := cachedScopes[updateScopeId]
		if cached {
			memory.RemoveScope(scanScope.Secret)
			logger.Infof("Scan scope '%s' (ID %d) removed from memory.", scanScope.Name, scanScope.Id)
		}
	}

	// Clean faded scan scopes form memory
	for cachedScopeId, cachedScope := range cachedScopes {
		if !utils.Uint64Contained(cachedScopeId, notification.RemainingScopeIds) {
			memory.RemoveScope(cachedScope.Secret)
			logger.Infof(
				"Scan scope '%s' (ID %d) cleaned from memory.", cachedScope.Name, cachedScope.Id)
		}
	}

	// Clean scan agent stats of expired scan scopes from memory
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

// cleanExceeded cleans exceeded scans by removing their target entries from the brokerdb and updating their info
// entries in the scope db
func cleanExceeded(logger scanUtils.Logger) {

	// Define cleanup procedure
	fnCleanup := func(scanScope *managerdb.T_scan_scope) {

		// Log potential panics before letting them move on
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
				panic(r)
			}
		}()

		// Open scope's database
		scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
		if errHandle != nil {
			logger.Errorf("Could not open scope database: %s", errHandle)
			return
		}

		// Log cleaning of exceeded module targets
		logger.Debugf("Cleaning exceeded '%s' scans.", discovery.Label)

		// Read '--min-hostgroup' value
		nmapHostgroup := extractMinHostgroup(scanScope.ScanSettings.DiscoveryNmapArgs)
		if nmapHostgroup == 0 {
			nmapHostgroup = defaultNmapMinHostgroup
		}

		// Extract '--host-timeout' value
		nmapHosttimeoutMinutes := extractHostTimeoutMinutes(scanScope.ScanSettings.DiscoveryNmapArgs)
		if nmapHosttimeoutMinutes == 0 {
			nmapHosttimeoutMinutes = defaultNmapHosttimeoutMinutes
		}

		// Clean exceeded discovery scan inputs in scope db, setting status and scan_finished timestamp
		countDiscovery, errCleanDiscovery := scopedb.CleanExceededDiscovery(
			scopeDb,
			nmapHostgroup,
			int(math.Ceil(nmapHosttimeoutMinutes)),
		)
		if errCleanDiscovery != nil {
			logger.Warningf(
				"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from scope db: %s",
				discovery.Label,
				scanScope.Name,
				scanScope.Id,
				errCleanDiscovery,
			)
		} else if countDiscovery > 0 {
			logger.Infof(
				"Set %d exceeded '%s' scans of scope '%s' (ID %d) to '%s' in scope db.",
				countDiscovery,
				discovery.Label,
				scanScope.Name,
				scanScope.Id,
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
				maxScanTimeMinutes = scanScope.ScanSettings.NfsScanTimeoutMinutes
			case smb.Label:
				maxScanTimeMinutes = scanScope.ScanSettings.SmbScanTimeoutMinutes
			case ssl.Label:
				maxScanTimeMinutes = scanScope.ScanSettings.SslScanTimeoutMinutes
			case ssh.Label:
				maxScanTimeMinutes = scanScope.ScanSettings.SshScanTimeoutMinutes
			case webcrawler.Label:
				maxScanTimeMinutes = scanScope.ScanSettings.WebcrawlerScanTimeoutMinutes
			case webenum.Label:
				maxScanTimeMinutes = scanScope.ScanSettings.WebenumScanTimeoutMinutes
			default:
				logger.Errorf("Invalid module label '%s'.", label)
				continue
			}

			// Calculate threshold for scans that ran into their timeout but didn't return results
			maxScanTime := time.Minute * time.Duration(maxScanTimeMinutes)
			maxScanTime = maxScanTime + (time.Minute * 10) // Add buffer for potential networking/processing delays
			startedBefore := time.Now().Add(-maxScanTime)

			// Update exceeded module scans in scope db, setting status and scan_finished timestamp
			countScopedb, errCleanScopedb := scopedb.CleanExceeded(scopeDb, label, startedBefore)
			if errCleanScopedb != nil {
				logger.Warningf(
					"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from scope db: %s",
					label,
					scanScope.Name,
					scanScope.Id,
					errCleanScopedb,
				)
			} else if countScopedb > 0 {
				logger.Infof(
					"Set %d exceeded '%s' scans of scope '%s' (ID %d) to '%s' in scope db.",
					countScopedb,
					label,
					scanScope.Name,
					scanScope.Id,
					scanUtils.StatusFailed,
				)
			}

			// Remove exceeded module inputs from brokerdb
			countBrokerdb, errCleanBrokerdb := brokerdb.CleanExceeded(scanScope, label, startedBefore)
			if errCleanBrokerdb != nil {
				logger.Warningf(
					"Could not clean exceeded '%s' scans of scope '%s' (ID %d) from broker db: %s",
					label,
					scanScope.Name,
					scanScope.Id,
					errCleanBrokerdb,
				)
			} else if countBrokerdb > 0 {
				logger.Infof(
					"Removed %d exceeded '%s' scans of scope '%s' (ID %d) from broker db.",
					countBrokerdb,
					label,
					scanScope.Name,
					scanScope.Id,
				)
			}
		}
	}

	// Iterate and clean all scan scopes
	for _, scanScope := range memory.GetScopes() {
		go fnCleanup(&scanScope)
	}
}

// submitStats queries current scan scope progress and submits it together with other scan scope sats to the manager.
// There it will be persisted and made available for other components, e.g. the web backend for display.
func submitStats(logger scanUtils.Logger) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
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
		agentStats := memory.GetAgents()

		// Clear memory so that new entries can accumulate already for the next execution
		// and to minimize possible race conditions of agent requests between GetAgents() and ClearAgents()
		memory.ClearAgents()

		// Group scan agent stats by scan scope ID
		scopeAgents := make(map[uint64][]managerdb.T_scan_agent)
		for _, agent := range agentStats {

			// Get scope specific slice
			scanAgents, ok := scopeAgents[agent.IdTScanScope]

			// Initialize scan agent slice for scan scope if not yet existing
			if !ok {
				scopeAgents[agent.IdTScanScope] = make([]managerdb.T_scan_agent, 0, 1)
				scanAgents, _ = scopeAgents[agent.IdTScanScope]
			}

			// Attach new scan agent statistics
			scanAgents = append(scanAgents, agent)

			// Update referenced scopes
			scopeAgents[agent.IdTScanScope] = scanAgents
		}

		// Send updated agent stats to manager via RPC. Additional scan progress stats will be queried and attached by
		// the manager during this call.
		errRpc := manager.RpcUpdateAgents(logger, rpcClient, scopeAgents)
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {
			logger.Debugf("Skipped submitting scan stats, due to connectivity issues.")
			return
		} else if errRpc != nil {
			logger.Warningf("Could not submit scan stats: %s", errRpc)
			return
		}
	}
}
