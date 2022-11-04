/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lithammer/shortuuid"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/vburenin/nsync"
	"gorm.io/gorm"
	"large-scale-discovery/broker/brokerdb"
	"large-scale-discovery/broker/memory"
	"large-scale-discovery/broker/scopedb"
	"large-scale-discovery/log"
	managerdb "large-scale-discovery/manager/database"
	"large-scale-discovery/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

var scopeModuleLock = nsync.NewNamedMutex() // Named mutex to prevent parallel changes on the same scope and module data

var errorGeneric = fmt.Errorf("RPC endpoint not available") // Generic error returned to the scan agent, which may not contain sensitive details

// Broker is used to implement the broker's RPC interfaces
type Broker struct{}

// RequestScanTasks processes scan task requests received from agents
func (s *Broker) RequestScanTasks(rpcArgs *ArgsGetScanTask, rpcReply *ReplyGetScanTask) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-RequestScanTasks", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get scan scope data by secret (will be taken from memory, if available).
	scanScope, errScanScope := getScanScope(logger, rpcArgs.ScopeSecret)
	if errors.Is(errScanScope, errInvalidScopeSecret) {
		return errScanScope // Let the scan agent know that the secret was invalid
	} else if errors.Is(errScanScope, errScopeNotAvailable) {
		return nil // Pretend there are currently no scan tasks available
	} else if scanScope == nil {
		logger.Warningf("Scan scope should be available but is not.")
		return nil
	}

	// Log action
	logger.Infof(
		"'%s' (%s/%s) requesting task for '%s' (ID %d).",
		rpcArgs.Name,
		rpcArgs.Host,
		rpcArgs.Ip,
		scanScope.Name,
		scanScope.Id,
	)

	// Prepare map of scan counts
	tasks := make(map[string]int, len(rpcArgs.ModuleData))

	// Iterate scan modules and log active tasks
	for _, agentModule := range rpcArgs.ModuleData {

		// Add module task count to map of scan counts
		tasks[agentModule.Label] = agentModule.ActiveTasks

		// Get maximum instances for scan module
		maxInstances, errInstances := scanScope.ScanSettings.MaxInstances(agentModule.Label)
		if errInstances != nil {
			logger.Errorf("Could not get max instances for '%s': %s", agentModule.Label, errInstances)
			return errorGeneric // Error message returned to agent
		}

		// Log actual number
		logger.Infof(
			"\t%3d / %3d active '%s' tasks.",
			agentModule.ActiveTasks,
			maxInstances,
			agentModule.Label,
		)
	}

	// Update cached agent stats
	memory.UpdateAgent(
		rpcArgs.Name,
		rpcArgs.Host,
		rpcArgs.Ip,
		scanScope.Id,
		tasks,
		rpcArgs.SystemData,
	)

	// Don't feed new scan tasks, while scan scope is paused
	if !scanScope.Enabled {

		// Return nil to indicate successful RPC call
		logger.Debugf("Scope currently paused.")
		return nil
	}

	// Create a WaitGroup to ensure that the database is up-to-date before the new targets are returned
	var wg sync.WaitGroup

	// Iterate scan modules and return task for available slots
	for _, agentModule := range rpcArgs.ModuleData {

		// Get maximum instances for scan module
		maxInstances, errInstances := scanScope.ScanSettings.MaxInstances(agentModule.Label)
		if errInstances != nil {
			logger.Errorf("Could not get max instances for '%s': %s", agentModule.Label, errInstances)
			return errorGeneric // Error message returned to agent
		}

		// Calculate available scan slots
		availableSlots := maxInstances - agentModule.ActiveTasks

		// Skip module if no slots available
		if availableSlots <= 0 {
			continue
		}

		// Query targets and build RPC response
		if agentModule.Label == discovery.Label { // Handle request for discovery scan scan targets

			// Query targets and fill RPC response with response data
			if availableSlots > 0 {
				errFeed := feedDiscovery(
					logger,
					scanScope,
					availableSlots,
					rpcArgs.Ip,
					rpcArgs.Host,
					rpcReply,
				)
				if errFeed != nil {
					logger.Errorf(
						"Could not feed discovery tasks for '%s' (ID %d): %s",
						scanScope.Name,
						scanScope.Id,
						errFeed,
					)
					return errorGeneric // Error message returned to agent
				}
			}

		} else { // Handle request for submodule scan targets

			// Query targets and fill RPC response with response data
			if availableSlots > 0 {
				errFeed := feedSubmodule(
					logger,
					scanScope,
					agentModule.Label,
					availableSlots,
					rpcArgs.Ip,
					rpcArgs.Host,
					&wg,
					rpcReply,
				)
				if errFeed != nil {
					logger.Errorf(
						"Could not feed '%s' tasks for '%s' (ID %d): %s",
						agentModule.Label,
						scanScope.Name,
						scanScope.Id,
						errFeed,
					)
					return errorGeneric // Error message returned to agent
				}
			}
		}
	}

	// Wait for all routines to finish
	wg.Wait()

	// Return nil to indicate successful RPC call
	return nil
}

// SubmitScanResult processes scan results received from agents
func (s *Broker) SubmitScanResult(rpcArgs *ArgsSaveScanResult, rpcReply *struct{}) error {

	// Process results based on their type
	switch rpcArgs.Result.(type) {
	case discovery.Result:
		// Save discovery scan result
		go saveDiscoveryResult(rpcArgs)
	default:
		// Save submodule result
		go saveSubResult(rpcArgs)
	}

	// Return nil to indicate successful RPC call
	return nil
}

// lock acquires an access lock for a given scan scope and module to prevent concurrent access to the same
// data segment in broker db and scope db. The scope's ID is used as an identifier, because it never changes, while the
// scope secret might change.
func lock(scanScopeId uint64, module string) {
	scopeModuleLock.Lock(strconv.FormatUint(scanScopeId, 10) + module)
}

// unlock releases an access lock for a given scan scope and module
func unlock(scanScopeId uint64, module string) {
	scopeModuleLock.Unlock(strconv.FormatUint(scanScopeId, 10) + module)
}

// feedDiscovery queries discovery scan targets (for a given scan scope) from *database* and adds them to the RPC response
// returned to the requesting the scan agent. In contrast to submodules, which have stored scan tasks in the local
// brokerdb, the discovery scan module takes targets directly from the database (t_discovery table).
func feedDiscovery(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	amount int,
	scanIp string,
	scanHostname string,
	rpcReply *ReplyGetScanTask,
) error {

	// Query current queue size to make sure submodules are not falling behind
	queueSize, errSize := brokerdb.GetScopeSize(scanScope.Id)
	if errSize != nil {
		return errSize
	}

	// Do not feed further discovery scan tasks until submodules could catch up
	if queueSize > 10000 { // Targets are easily piled up if a few discovery scan instances send results
		logger.Infof("Queue size exceeded for '%s' (ID %d). %d targets queued.", scanScope.Name, scanScope.Id, queueSize)
		return nil
	}

	// Acquire brokerdb access for related requests (same scope and scan type)
	lock(scanScope.Id, discovery.Label)

	// Unlock module access for the scope
	defer unlock(scanScope.Id, discovery.Label)

	// Check whether a new scan cycle should be initialized
	if scanScope.Cycles {
		_, errCycle := checkCycle(logger, scanScope)
		if errCycle != nil {
			return errCycle
		}
	}

	// Get timezone ranges that are currently within the configured working hours range
	timezoneRanges := utils.TimezonesBetween(
		scanScope.ScanSettings.DiscoveryTimeEarliest,
		scanScope.ScanSettings.DiscoveryTimeLatest,
		scanScope.ScanSettings.DiscoverySkipDaysSlice,
	)

	// Get discovery scan targets directly from scope db (t_discovery) and block them for other agent's requests
	newTargets, errTargets := scopedb.GetBlockDiscoveryTargets(
		logger,
		scanScope,
		amount,
		timezoneRanges,
		scanIp,
		scanHostname,
	)
	if errTargets != nil {
		return errTargets
	}

	// Return if no targets were found
	if len(newTargets) == 0 {
		logger.Debugf("Currently no targets available for '%s'.", discovery.Label)
		return nil
	}

	// Log action
	logger.Infof("Feeding %d '%s' tasks.", len(newTargets), discovery.Label)

	// Inject default '--min-hostgroup' argument if not present. A fixed value is required to estimate
	// the maximum scan duration.
	if v := extractMinHostgroup(scanScope.ScanSettings.DiscoveryNmapArgs); v == 0 {
		scanScope.ScanSettings.DiscoveryNmapArgs = strings.Join([]string{
			scanScope.ScanSettings.DiscoveryNmapArgs,
			"--min-hostgroup",
			strconv.Itoa(defaultNmapMinHostgroup),
		}, " ")
	}

	// Inject default '--host-timeout' argument if not present. Some fixed value is required to estimate
	// the maximum scan duration.
	if v := extractHostTimeoutMinutes(scanScope.ScanSettings.DiscoveryNmapArgs); v == 0 {
		scanScope.ScanSettings.DiscoveryNmapArgs = strings.Join([]string{
			scanScope.ScanSettings.DiscoveryNmapArgs,
			"--host-timeout",
			strconv.Itoa(defaultNmapHosttimeoutMinutes) + "m",
		}, " ")
	}

	// Add new scan targets to RPC reply
	for _, target := range newTargets {

		// Prepare scan task
		scanTask := ScanTask{
			Label:        discovery.Label,
			Id:           target.Id,
			Target:       target.Input,
			ScanSettings: scanScope.ScanSettings,
		}

		// Append scan task to list of tasks
		rpcReply.ScanTasks = append(rpcReply.ScanTasks, scanTask)
	}

	// Return nil as everything went fine
	return nil
}

// feedSubmodule queries scan tasks (for a certain scan scope and submodule) from local brokerdb and returns them to
// the scan agent.
func feedSubmodule(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	label string,
	amount int,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
	rpcReply *ReplyGetScanTask,
) error {

	// Acquire brokerdb access for the scope and module to prevent multiple agents from manipulating the same data
	lock(scanScope.Id, label)

	// Unlock module access for the scope and module
	defer unlock(scanScope.Id, label)

	// Query next submodule targets from brokerdb
	newTargets, errTargets := brokerdb.GetScopeTargets(scanScope.Id, label, amount)
	if errTargets != nil {
		return errTargets
	}

	// Return if no targets were found
	if len(newTargets) == 0 {
		logger.Debugf("Currently no targets available for '%s'.", label)
		return nil
	}

	// Log action
	logger.Infof("Feeding %d '%s' tasks.", len(newTargets), label)

	// Store target IDs from brokerdb and database
	var brokerDbIds []uint64
	var scopeDbIds []uint64
	for _, target := range newTargets {
		brokerDbIds = append(brokerDbIds, target.Id)                // Store id from broker db (T_sub_input -> id)
		scopeDbIds = append(scopeDbIds, target.IdTDiscoveryService) // Store id from scope db (t_discovery_services -> id)
	}

	// Get time of starting this scans to make sure brokerdb and scope db values are in sync
	scanStartTime := sql.NullTime{Time: time.Now(), Valid: true}

	// Update scan_started timestamp in brokerdb
	errSet := brokerdb.SetTargetsStarted(newTargets, brokerDbIds, scanStartTime)
	if errSet != nil {
		return errSet
	}

	// Create initial entry in the database, to indicate running scan (might be interesting for the user), also
	// increment the WaitGroup to ensure that the database is up-to-date
	wg.Add(1)
	switch label {
	case banner.Label:
		go scopedb.PrepareBannerResults(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case nfs.Label:
		go scopedb.PrepareNfsResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case smb.Label:
		go scopedb.PrepareSmbResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case ssl.Label:
		go scopedb.PrepareSslResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case ssh.Label:
		go scopedb.PrepareSshResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case webcrawler.Label:
		go scopedb.PrepareWebcrawlerResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	case webenum.Label:
		go scopedb.PrepareWebenumResult(logger, scanScope, scopeDbIds, scanStartTime, scanIp, scanHostname, wg)
	default:
		// Unexpected label, log and decrement the WaitGroup as no routine was started
		logger.Debugf("Unknown label: %s", label)
		wg.Done()
	}

	// Add new scan targets to RPC reply
	for _, target := range newTargets {

		// Append scan task to list of tasks. ScanTask is a generic struct, the agent will pick the data required by
		// the specific scan task.
		rpcReply.ScanTasks = append(rpcReply.ScanTasks, ScanTask{
			Label:          label,
			Id:             target.Id, // Send the id/pk from brokerdb, it will be returned back with the result
			Target:         target.Address,
			Protocol:       target.Protocol,
			Port:           target.Port,
			OtherNames:     strings.Split(target.OtherNames, scopedb.DbValueSeparator),
			Service:        target.Service,
			ServiceProduct: target.ServiceProduct,
			ScanSettings:   scanScope.ScanSettings,
		})
	}

	// Return nil as everything went fine
	return nil
}

func saveDiscoveryResult(rpcArgs *ArgsSaveScanResult) {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-saveDiscoveryResult", uuid))

	// Recover potential panics to avoid broker termination
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Unexpected error while saving discovery result: %s", r)
		}
	}()

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get scan scope data by secret (will be taken from memory, if available)
	scanScope, errScanScope := getScanScope(logger, rpcArgs.ScopeSecret)
	if errScanScope != nil {
		logger.Errorf(
			"'%s' (%s/%s) sent result '%d' with scope secret '%s...', but could not be processed: %s",
			rpcArgs.Name,
			rpcArgs.Host,
			rpcArgs.Ip,
			rpcArgs.Id,
			rpcArgs.ScopeSecret[0:5],
			errScanScope,
		)
		return
	}

	// Log action
	logger.Infof(
		"'%s' (%s/%s) sent discovery result '%d' of '%s' (ID %d).",
		rpcArgs.Name,
		rpcArgs.Host,
		rpcArgs.Ip,
		rpcArgs.Id,
		scanScope.Name,
		scanScope.Id,
	)

	// Cast data to according data type
	nmapRes := rpcArgs.Result.(discovery.Result)

	// Something went wrong on the agent
	if nmapRes.Exception {
		logger.Errorf(
			"Discovery scan '%d' of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
			rpcArgs.Name,
			rpcArgs.Host,
			rpcArgs.Ip,
			nmapRes.Status,
		)
		return
	}

	// Process discovery result and save results to scope db and sub targets to brokerdb
	errSave := scopedb.SaveDiscoveryResult(logger, scanScope, rpcArgs.Id, &nmapRes)
	if errSave != nil {
		logger.Errorf(
			"Could not save discovery result of scan scope '%s' (ID %d): %s. Dropping result.",
			scanScope.Name,
			scanScope.Id,
			errSave,
		)
	} else {
		logger.Debugf("Result saved successfully.")
	}
}

func saveSubResult(rpcArgs *ArgsSaveScanResult) {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-saveSubResult", uuid))

	// Recover potential panics to avoid broker termination
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Unexpected error while saving sub result: %s", r)
		}
	}()

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get scan scope data by secret (will be taken from memory, if available)
	scanScope, errScanScope := getScanScope(logger, rpcArgs.ScopeSecret)
	if errScanScope != nil {
		logger.Errorf(
			"'%s' (%s/%s) sent result '%d' with scope secret '%s...', but could not be processed: %s",
			rpcArgs.Name,
			rpcArgs.Host,
			rpcArgs.Ip,
			rpcArgs.Id,
			rpcArgs.ScopeSecret[0:5],
			errScanScope,
		)
		return
	}

	// Log action
	logger.Debugf(
		"'%s' (%s/%s) sent result '%d' of '%s' (ID %d).",
		rpcArgs.Name,
		rpcArgs.Host,
		rpcArgs.Ip,
		rpcArgs.Id,
		scanScope.Name,
		scanScope.Id,
	)

	// Query related target from brokerdb, it also contains the related scope db service ID
	subInput, errGet := brokerdb.GetTarget(rpcArgs.Id)
	if errors.Is(errGet, gorm.ErrRecordNotFound) { // Check if entry didn't exist
		logger.Infof("Scan task '%d' got cleaned up already. Dropping result.", rpcArgs.Id)
		return
	} else if errGet != nil {
		logger.Errorf("Could not query broker for scan target '%d': %s. Dropping result.", rpcArgs.Id, errGet)
		return
	}

	// Define error variable set in case of a database error
	var errSave error

	// Process results based on their result/module type
	var label string
	switch rpcArgs.Result.(type) {
	case banner.Result: // Result from the banner submodule
		label = banner.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		bannerRes := rpcArgs.Result.(banner.Result)

		// Something went wrong on the agent
		if bannerRes.Exception {
			logger.Errorf(
				"'%s' scan '%d' of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				rpcArgs.Id,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				bannerRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveBannerResult(logger, scanScope, subInput.IdTDiscoveryService, &bannerRes)

	case nfs.Result:
		label = nfs.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		nfsRes := rpcArgs.Result.(nfs.Result)

		// Something went wrong on the agent
		if nfsRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				nfsRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveNfsResult(logger, scanScope, subInput.IdTDiscoveryService, &nfsRes)

	case smb.Result:
		label = smb.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		smbRes := rpcArgs.Result.(smb.Result)

		// Something went wrong on the agent
		if smbRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				smbRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveSmbResult(logger, scanScope, subInput.IdTDiscoveryService, &smbRes)

	case ssh.Result:
		label = ssh.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		sshRes := rpcArgs.Result.(ssh.Result)

		// Something went wrong on the agent
		if sshRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				sshRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveSshResult(logger, scanScope, subInput.IdTDiscoveryService, &sshRes)

	case ssl.Result:
		label = ssl.Label
		logger.Infof(
			"Processing '%s' result '%d' for"+
				" '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		sslRes := rpcArgs.Result.(ssl.Result)

		// Something went wrong on the agent
		if sslRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				sslRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveSslResult(logger, scanScope, subInput.IdTDiscoveryService, &sslRes)

	case webcrawler.Result:
		label = webcrawler.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		webcrawlerRes := rpcArgs.Result.(webcrawler.Result)

		// Something went wrong on the agent
		if webcrawlerRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				webcrawlerRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveWebcrawlerResult(logger, scanScope, subInput.IdTDiscoveryService, &webcrawlerRes)

	case webenum.Result:
		label = webenum.Label
		logger.Infof(
			"Processing '%s' result '%d' for '%s' (ID %d).",
			label,
			rpcArgs.Id,
			scanScope.Name,
			scanScope.Id,
		)

		// Cast data to according data type
		webenumRes := rpcArgs.Result.(webenum.Result)

		// Something went wrong on the agent
		if webenumRes.Exception {
			logger.Errorf(
				"'%s' scan of scan scope '%s' (ID %d) crashed unexpectedly on '%s' (%s/%s): %s",
				label,
				scanScope.Name,
				scanScope.Id,
				rpcArgs.Name,
				rpcArgs.Host,
				rpcArgs.Ip,
				webenumRes.Status,
			)
			return
		}

		// Write scan result to database
		errSave = scopedb.SaveWebenumResult(logger, scanScope, subInput.IdTDiscoveryService, &webenumRes)

	default: // Unknown data type
		logger.Errorf(
			"Unknown result type '%s' for '%s' (ID %d). Dropping result.",
			scanScope.Name,
			scanScope.Id,
			rpcArgs,
		)
		return
	}

	// Do cleanup if saving results succeeded
	if errSave != nil {
		logger.Errorf(
			"Could not save '%s' result for '%s' (ID: %d): %s. Dropping result.",
			label,
			scanScope.Name,
			scanScope.Id,
			errSave,
		)
	} else {
		logger.Debugf("Result saved successfully.")

		// Remove target from cache
		errRem := brokerdb.DeleteTarget(subInput)
		if errRem != nil {
			// Rollback is not necessary at this point ... just warn and investigate & fix if it happens
			logger.Errorf("Could not prune target '%d' from cache: %s", rpcArgs.Id, errRem)
		}
	}
}
