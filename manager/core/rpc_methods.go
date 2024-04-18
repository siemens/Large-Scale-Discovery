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
	"database/sql"
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/manager/config"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/vburenin/nsync"
	"math"
	"strconv"
	"strings"
	"time"
)

var scopeTargetsUpdating = nsync.NewNamedMutex() // Block subsequent scan scope targets updates while previous is active

var ErrScopeUpdateOngoing = fmt.Errorf("synchronization of scan targets still ongoing")
var ErrViewNameExisting = fmt.Errorf("view name already existing")

// desensitize cleans sensitive data from a scan scope struct and it's db server struct, before data is
// returned to the RPC client. Such sensitive data may not leave the manager by default.
func desensitize(scanScope *database.T_scan_scope) {

	// Remove scope secret
	scanScope.Secret = ""

	// Remove database connection details
	scanScope.DbServer.Dialect = ""
	scanScope.DbServer.Host = ""
	scanScope.DbServer.Admin = ""
	scanScope.DbServer.Password = ""
	scanScope.DbServer.Args = ""
	// The port and public host are not sensitive and need to be presented to users
}

// Manager is used to implement the manager's RPC interfaces
type Manager struct{}

// SubscribeNotification can be called by a client in order to be notified about changes, e.g. in scan scope's details.
// The call will not be answered until changes become known. Basically the client's call will block until changes
// happen. It will then receive a notification. After processing the notification, it should subscribe *again* to
// receive subsequent/future notifications. The notifier will notify in intervals, so the client has a few seconds
// to process and re-subscribe.
func (s *Manager) SubscribeNotification(rpcArgs struct{}, rpcReply *ReplyNotification) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-SubscribeNotification", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client subscribing to scope changes.")

	// Block and wait for change notifications
	// If the client disappears, this RPC call will still stay alive until changes are received from the notifier, but
	// this is not expected to be an issue for now. There should not pile up too many stale RPC calls of this type and
	// they don't consume processing power while waiting.
	scopeUpdateEvents, err := scopeChangeNotifier.Receive()
	if err != nil {
		logger.Debugf("Notifier shutting down.")
		rpcReply.UpdateScopeIds = make([]uint64, 0, 1)
		rpcReply.RemainingScopeIds = make([]uint64, 0, 1)
		return nil // Silently return request, clients need to re-subscribe
	}

	// Log action
	logger.Debugf("Notifying client about %d scan scope changes", len(scopeUpdateEvents))

	// Translate to correct data type
	var scopeUpdateIds = make([]uint64, 0, len(scopeUpdateEvents))
	for _, val := range scopeUpdateEvents {

		// Cast to integer
		v := val.(uint64)

		// Append if > 0. Zero is not a valid scope ID, but might have been pushed to indicate a scan scope
		// got deleted, so some client refreshes data based on 'remainingScanScopeIds' sent along below.
		if v > 0 {
			scopeUpdateIds = append(scopeUpdateIds, v)
		}
	}

	// Get remaining scan scope IDs, and send them along with each notification, so that consumers can check
	// whether some scan scopes disappeared in order to initiate a cleanup.
	remainingScanScopeIds, errRemaining := database.GetScopeEntryIds()
	if errRemaining != nil {
		logger.Errorf("Could not query remaining scan scope IDs: %s", errRemaining)
		return errRemaining
	}

	// Prepare response
	rpcReply.UpdateScopeIds = scopeUpdateIds
	rpcReply.RemainingScopeIds = remainingScanScopeIds

	// Return nil as everything went fine
	return nil
}

// GetScope returns a specific scan scope to an RPC client
func (s *Manager) GetScope(rpcArgs *ArgsScopeId, rpcReply *ReplyScanScope) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetScope", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope '%d'.", rpcArgs.ScopeId)

	// Get scan scope for given secret
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not query scan scope '%d.", rpcArgs.ScopeId)
		return errScopeEntry
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	desensitize(scopeEntry)

	// Copy scan scope data into RPC response
	rpcReply.ScanScope = *scopeEntry

	// Log completion
	logger.Debugf("Scope returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetScopes returns all available scan scopes to an RPC client
func (s *Manager) GetScopes(rpcArgs *struct{}, rpcReply *ReplyScanScopes) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetScopes", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scopes.")

	// Find all scan scopes
	scopeEntries, errScopeEntries := database.GetScopeEntries()
	if errScopeEntries != nil {
		logger.Errorf("Could not query scan scopes: %s", errScopeEntries)
		return errScopeEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, scanScope := range scopeEntries {
		desensitize(&scanScope)
		scopeEntries[i] = scanScope
	}

	// Attach list of scan scopes to RPC response
	rpcReply.ScanScopes = scopeEntries

	// Log completion
	logger.Debugf("Scan scopes returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetScopesOf returns scan scopes for a given group ID
func (s *Manager) GetScopesOf(rpcArgs *ArgsGroupIds, rpcReply *ReplyScanScopes) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetScopesOf", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scopes of groups %s.", utils.JoinUint64(rpcArgs.GroupIds, ", "))

	// Get scan scopes of group
	scopeEntries, errScopeEntries := database.GetScopeEntriesOf(rpcArgs.GroupIds)
	if errScopeEntries != nil {
		logger.Errorf("Could not query scan scopes of %s.", utils.JoinUint64(rpcArgs.GroupIds, ", "))
		return errScopeEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, scanScope := range scopeEntries {
		desensitize(&scanScope)
		scopeEntries[i] = scanScope
	}

	// Copy scan scope data into RPC response
	rpcReply.ScanScopes = scopeEntries

	// Log completion
	logger.Debugf("Scan scopes returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// CreateScope creates a scan scope
func (s *Manager) CreateScope(rpcArgs *ArgsScopeDetails, rpcReply *ReplyScopeId) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-CreateScope", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get config
	conf := config.GetConfig()

	// Log action
	logger.Debugf("Client requesting scan scope creation.")

	// Get scope database entry
	serverEntry, errServerEntry := database.GetServerEntry(rpcArgs.DbServerId)
	if errServerEntry != nil {
		logger.Errorf("Could not get database server details to create scan scope: %s", errServerEntry)
		return errServerEntry
	}

	// Open server database
	serverDb, errServerHandle := database.GetServerDbHandle(logger, serverEntry)
	if errServerHandle != nil {
		logger.Errorf("Could not get database server handle to create scan scope: %s", errServerHandle)
		return errServerHandle
	}

	// Generate DB name
	dbName := strings.ToLower(shortuuid.New()[0:10]) // db names need to be lower case to avoid subsequent errors

	// Generate new scope secret
	scopeSecret, errSecret := utils.GenerateToken(utils.AlphaNumDash, 64)
	if errSecret != nil {
		logger.Errorf("Could not generate scope secret: %s", errSecret)
		return errSecret
	}

	// Create scope database and manager db entry
	scopeEntry, errCreateScope := database.XCreateScope(
		serverDb,
		serverEntry,
		rpcArgs.Name,
		dbName,
		rpcArgs.GroupId,
		rpcArgs.CreatedBy,
		scopeSecret,
		rpcArgs.Type,
		rpcArgs.Cycles,
		rpcArgs.CyclesRetention,
		rpcArgs.Attributes,
		conf.ScanDefaults,
	)
	if errCreateScope != nil {
		logger.Errorf("Could not create scan scope: %s", errCreateScope)
		return errCreateScope
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to initialize new scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Revoke public access rights from public, this can only be done with a direct connection to the database
	errRevoke := scopeDb.Exec(`REVOKE ALL ON schema public FROM public;`).Error
	if errRevoke != nil {
		return errRevoke
	}

	// Create tables in new scope db
	errMigrate := database.AutoMigrateScopeDb(scopeDb)
	if errMigrate != nil {
		logger.Errorf("Could not automigrate scan scope '%s' (ID %d)", scopeEntry.Name, scopeEntry.Id)
		return errMigrate
	}

	// Install trigram indexes
	errExtension := database.InstallTrigramIndices(scopeDb)
	if errExtension != nil {
		logger.Errorf(
			"Could not install database extensions for scan scope '%s' (ID %d)", scopeEntry.Name, scopeEntry.Id)
		return errExtension
	}

	// Create default view for new scan scope
	_, errCreateView := database.XCreateView(scopeDb, scopeEntry, "All", rpcArgs.CreatedBy, nil)
	if errCreateView != nil {

		// Log issue
		logger.Errorf(
			"Could not create default view on scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errCreateView,
		)
	}

	// Return scope ID of new scan scope
	rpcReply.ScopeId = scopeEntry.Id

	// Create notification about updated scan scope for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Log completion
	logger.Debugf("Scope created.")

	// Return nil to indicate successful RPC call
	return nil
}

// DeleteScope removes a scan scope including associated views and access grants from the scope db and the manager db
func (s *Manager) DeleteScope(rpcArgs *ArgsScopeId, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-DeleteScope", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scan scope deletion.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to delete scan scope: %s", errScopeEntry)
		return errScopeEntry
	}

	// Open server database
	serverDb, errServerHandle := database.GetServerDbHandle(logger, &scopeEntry.DbServer)
	if errServerHandle != nil {
		logger.Errorf(
			"Could not get database server handle to delete scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errServerHandle,
		)
		return errServerHandle
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to delete scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Delete scope on scope db and manager db
	errDelete := database.XDeleteScope(serverDb, scopeDb, scopeEntry)
	if errDelete != nil {
		logger.Errorf(
			"Could not delete scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errDelete,
		)
		return errDelete
	}

	// Create notification about updated scan scope for clients The remaining scan scope IDs are sent along with
	// every notification, so it is just necessary to trigger an empty one.
	scopeChangeNotifier.Send(uint64(0))

	// Log completion
	logger.Debugf("Scope deleted.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetScopeTargets queries the current list of scan targets from a scan scope
func (s *Manager) GetScopeTargets(rpcArgs *ArgsScopeId, rpcReply *ReplyScopeTargets) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetScopeTargets", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scan scope targets.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to query scan scope targets: %s", errScopeEntry)
		return errScopeEntry
	}

	// Check whether previous scope synchronization is still ongoing
	lockName := strconv.Itoa(int(scopeEntry.Id))
	acquired := scopeTargetsUpdating.TryLock(lockName)

	// Abort if previous scope synchronization is still ongoing
	if acquired {
		scopeTargetsUpdating.Unlock(lockName) // Release lock, obviously nothing going on
	} else {

		// Set flag to indicate ongoing scope synchronization and return
		logger.Debugf("Scope targets currently being synchronized.")
		rpcReply.Synchronization = true
		return nil
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to delete scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Query current list of scan scopes targets
	targets, errTargets := database.GetTargets(scopeDb)
	if errTargets != nil {
		logger.Errorf(
			"Could not query targets of scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errTargets,
		)
		return errTargets
	}

	// Set response values
	rpcReply.Targets = targets

	// Log completion
	logger.Debugf("Scope targets returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// ToggleScope enables/disables a scan scope. Disabled (paused) scan scopes are not processed by the broker. Scan
// agents will be able to complete running scan tasks, but not receive new ones.
func (s *Manager) ToggleScope(rpcArgs *ArgsScopeId, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-ToggleScope", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting to enable/disable scan scope.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to delete scan scope: %s", errScopeEntry)
		return errScopeEntry
	}

	// Toggle scan scope enabled attribute to enable/disable scan scope
	scopeEntry.Enabled = !scopeEntry.Enabled

	// Write changes
	_, errSave := scopeEntry.Save("enabled")
	if errSave != nil {
		logger.Errorf("Could not toggle scan scope: %s", errSave)
		return errSave
	}

	// Create notification about updated scan scope for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Log completion
	if scopeEntry.Enabled {
		logger.Debugf("Scope paused.")
	} else {
		logger.Debugf("Scope resumed.")
	}

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateScope updates scan scope details
func (s *Manager) UpdateScope(rpcArgs *ArgsScopeUpdate, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateScope", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope update.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.IdTScanScopes)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry '%d' to update: %s", rpcArgs.IdTScanScopes, errScopeEntry)
		return errScopeEntry
	}

	// Update scope database comment if necessary
	if scopeEntry.Name != rpcArgs.Name {

		// Open server database
		serverDb, errServerHandle := database.GetServerDbHandle(logger, &scopeEntry.DbServer)
		if errServerHandle != nil {
			logger.Errorf(
				"Could not get database server handle to update database comment '%s' (ID %d): %s",
				scopeEntry.Name,
				scopeEntry.Id,
				errServerHandle,
			)
			return errServerHandle
		}

		// Update database comment for users browsing accessible databases to distinguish them. The
		// randomly generated database names are not easy to tell apart.
		errComment := database.SetScopeDbComment(serverDb, scopeEntry.DbName, rpcArgs.Name)
		if errComment != nil {
			return fmt.Errorf("could not update database comment: %s", errComment)
		}
	}

	// Process scan scope attributes if passed
	if rpcArgs.Attributes != nil {

		// Safety check whether the same amount of attributes was supplied
		if len(scopeEntry.Attributes) != len(*rpcArgs.Attributes) {
			logger.Errorf(
				"Unequal scope attributes were supplied for scan scope '%s' (ID %d)",
				scopeEntry.Name,
				scopeEntry.Id,
			)
			return fmt.Errorf("scope attributes divergent")
		}

		// Safety check to prevent accidentally adding new
		for k := range *rpcArgs.Attributes {
			_, ok := scopeEntry.Attributes[k]
			if !ok {
				logger.Errorf(
					"Unequal scope attributes were supplied for scan scope '%s' (ID %d)",
					scopeEntry.Name,
					scopeEntry.Id,
				)
				return fmt.Errorf("scope attribute unknonw")
			}
		}

		// Update attribute
		scopeEntry.Attributes = *rpcArgs.Attributes
	}

	// Update attributes
	scopeEntry.Name = rpcArgs.Name
	scopeEntry.Cycles = rpcArgs.Cycles
	scopeEntry.CyclesRetention = rpcArgs.CyclesRetention

	// Save updated attributes
	saved, errSave := scopeEntry.Save("name", "cycles", "cycles_retention", "attributes")
	if errSave != nil {
		logger.Errorf("Could not update scan scope '%d': %s", rpcArgs.IdTScanScopes, errSave)
		return errSave
	}

	// Warn if there was more than one entry updated
	if saved != 1 {
		logger.Errorf("Updated an invalid amount of scopes: %d", saved)
	}

	// Create notification about updated scan scope for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Log completion
	logger.Debugf("Scope updated.")

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateScopeTargets updates scan scope details
func (s *Manager) UpdateScopeTargets(rpcArgs *ArgsTargetsUpdate, rpcReply *ReplyTargetsUpdate) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateScopeTargets", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope targets update.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.IdTScanScopes)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry '%d' to update: %s", rpcArgs.IdTScanScopes, errScopeEntry)
		return errScopeEntry
	}

	// Calculate amount of target addresses contained in target list (might be a mix of IPs/hostnames/network ranges)
	var totalTargets uint
	var targets []database.T_discovery
	for _, target := range rpcArgs.Targets {

		// Check if input is actually valid
		if target.Input == "" {
			logger.Warningf(
				"Empty input target dropped from inputs for '%s' (ID %d).",
				scopeEntry.Name,
				scopeEntry.Id,
			)
			continue
		} else if !scanUtils.IsValidAddress(target.Input) && !scanUtils.IsValidIpRange(target.Input) {
			logger.Warningf(
				"Invalid input target '%s' dropped from inputs for '%s' (ID %d).",
				target.Input,
				scopeEntry.Name,
				scopeEntry.Id,
			)
			continue
		}

		// Calculate and set input size
		count, errCount := utils.CountIpsInInput(target.Input)
		if errCount != nil {
			logger.Errorf("Could not calculate input size of '%s': %s", target.Input, errCount)
			return errCount
		}

		// Set input target size before inserting it into the database. Don't trust the RPC client
		target.InputSize = count

		// Append target to actual sanitized list
		targets = append(targets, target)

		// Count
		totalTargets += count
	}

	// Update scope size to new value
	scopeEntry.Size = totalTargets
	scopeEntry.LastSync = time.Now()

	// Save updated attributes
	saved, errSave := scopeEntry.Save("size", "last_sync")
	if errSave != nil {
		logger.Errorf(
			"Could not update size of scan scope %s ('%d'): %s", scopeEntry.Name, scopeEntry.Id, errSave)
		return errSave
	}

	// Warn if there was more than one entry updated
	if saved != 1 {
		logger.Errorf("Updated an invalid amount of scopes: %d", saved)
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to update scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Prepare lock name
	lockName := strconv.Itoa(int(scopeEntry.Id))

	// Acquire lock
	acquired := scopeTargetsUpdating.TryLock(lockName)

	// Abort action because update is currently ongoing
	if !acquired {
		return ErrScopeUpdateOngoing
	}

	// Prepare function to actually set targets, which can be launched either synchronously or asynchronously
	setTargets := func() error {

		// Make sure named mutex is released again to allow subsequent changes
		defer scopeTargetsUpdating.Unlock(lockName)

		// Execute update
		created, removed, updated, errSetTargets := database.SetTargets(scopeDb, targets)
		if errSetTargets != nil {
			logger.Errorf("Could not set inputs of scan scope '%s' (ID %d)", scopeEntry.Name, scopeEntry.Id)
			return fmt.Errorf("could not set scan scope inputs: %s", errSetTargets)
		} else {
			// Log some stats
			logger.Debugf("%d input targets created.", created)
			logger.Debugf("%d input targets removed.", removed)
			logger.Debugf("%d input targets updated.", updated)

			// Update reply
			rpcReply.Created = created
			rpcReply.Removed = removed
			rpcReply.Updated = updated
		}

		// Return nil as everything went fine
		return nil
	}

	// Continue with update in background
	if rpcArgs.Blocking {
		logger.Debugf("Synchronizing scope targets.")
		errSet := setTargets()
		if errSet != nil {
			rpcReply.Error = errSet.Error()
		}
		logger.Debugf("Scope targets updated.")
	} else {
		// Don't wait for actual update
		go func() { _ = setTargets() }()

		// Log background activity
		logger.Debugf("Synchronizing scope targets in the background.")
	}

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateSettings updates scan scope scan settings, depending on the arguments given
func (s *Manager) UpdateSettings(rpcArgs *ArgsSettingsUpdate, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateSettings", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope settings update.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.IdTScanScopes)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope settings entry '%d' to update: %s", rpcArgs.IdTScanScopes, errScopeEntry)
		return errScopeEntry
	}

	// Update scope's scan settings
	saved, errSave := scopeEntry.ScanSettings.SaveAll(&rpcArgs.ScanSettings)
	if errSave != nil {
		logger.Errorf("Could not update scope settings '%d': %s", rpcArgs.IdTScanScopes, errSave)
		return errSave
	}

	// Warn if there was more than one entry updated
	if saved != 1 {
		logger.Errorf("Updated an invalid amount of scope settings: %d", saved)
	}

	// Create notification about updated scan scope for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Log completion
	logger.Debugf("Scope settings updated.")

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateAgents updates scan agent stats in the manager's db.
func (s *Manager) UpdateAgents(rpcArgs *ArgsStatsUpdate, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateAgents", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope stats update.")

	// Prepare list of update entries
	for scopeId, scanAgents := range rpcArgs.ScanAgents {

		// Get scan scope for given secret
		scopeEntry, errScopeEntry := database.GetScopeEntry(scopeId)
		if errScopeEntry != nil {

			// Add scope ID to RPC reply to notify about vanished scan scopes for cleanup
			logger.Debugf("Scan scope '%d' does not exist anymore.", scopeId)
			continue // Proceed with next scan stats entry
		}

		// Open scope's database
		scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
		if errScopeHandle != nil {
			logger.Errorf("Could not open scope database: %s", errScopeHandle)
			return errScopeHandle
		}

		// Query scan progress from scope db
		total, done, active, failed, errProgress := database.GetProgress(scopeDb)
		if errProgress != nil {
			logger.Warningf("Could not query scan progress: %s", errProgress)
			return errProgress
		}

		// Prepare memory for percentage calculation
		const decimalPlaces = 4 // The amount of decimal places the rate should be floored to
		factor := math.Pow(10, decimalPlaces)

		// Let "done" be 100% if there are no targets in the scan scope yet
		d, a, f := 100., 0., 0.

		// Calculate updated progress values
		if total > 0 {
			d = float64(done) / float64(total) * 100
			a = float64(active) / float64(total) * 100
			f = float64(failed) / float64(total) * 100
		}

		// Update attributes
		scopeEntry.CycleDone = math.Floor(d*factor) / factor
		scopeEntry.CycleActive = math.Floor(a*factor) / factor
		scopeEntry.CycleFailed = math.Floor(f*factor) / factor

		// Save updated attributes
		_, errSave := scopeEntry.Save("cycle_done", "cycle_active", "cycle_failed")
		if errSave != nil {
			logger.Errorf("Could not update scope progress: %s", errSave)
			return errSave
		}

		// Update scan agent data. Existing scan agents are updated. Not existing ones created. None will be removed.
		errAgentsUpdate := database.UpdateScanAgents(scopeId, scanAgents)
		if errAgentsUpdate != nil {
			logger.Errorf("Could not update agent stats: %s", errAgentsUpdate)
			return errAgentsUpdate
		}
	}

	// Log completion
	logger.Debugf("Scope stats updated.")

	// Return nil to indicate successful RPC call
	return nil
}

// NewCycle initializes a new scan cycle for a given scan scope. All scan progress will be reset,
// but existing results will be kept.
func (s *Manager) NewCycle(rpcArgs *ArgsScopeId, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-NewCycle", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting initialization of new scan cycle.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry '%d' to initialize scan cycle: %s", rpcArgs.ScopeId, errScopeEntry)
		return errScopeEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to initialize new scan cycle for '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Execute scan cycle initialization
	errNewCycle := database.XCycleScope(logger, scopeDb, scopeEntry)
	if errNewCycle != nil {
		logger.Errorf(
			"Could not initialize new scan cycle for '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errNewCycle,
		)
		return errNewCycle
	}

	// Create notification about incremented scan cycle for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Log completion
	logger.Debugf("New scan cycle initialized.")

	// Return nil to indicate successful RPC call
	return nil
}

// ResetInput reset a scan scope's input target to put it back into queue
func (s *Manager) ResetInput(rpcArgs *ArgsTargetReset, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-ResetInput", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope target reset.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to reset scope target: %s", errScopeEntry)
		return errScopeEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to initialize new scan scope '%s' (ID %d): %s",
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Define update values
	values := map[string]interface{}{
		"scan_started":  sql.NullTime{},
		"scan_finished": sql.NullTime{},
		"scan_status":   scanUtils.StatusWaiting,
	}

	// For discovery scans, the timeout must be calculated dynamically, based on the actual input size
	errReset := scopeDb.Model(&database.T_discovery{}).
		Where("input = ?", rpcArgs.Input).
		Updates(values).Error
	if errReset != nil {
		return errReset
	}

	// Log completion
	logger.Debugf("Scope target reset.")

	// Return nil to indicate successful RPC call
	return nil
}

// ResetSecret reset the secret of a given scan scope used to associate scan agents
func (s *Manager) ResetSecret(rpcArgs *ArgsScopeId, rpcReply *ReplyScopeSecret) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-ResetSecret", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scope secret reset.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to reset scope secret: %s", errScopeEntry)
		return errScopeEntry
	}

	// Generate new scope secret
	scopeSecret, errSecret := utils.GenerateToken(utils.AlphaNumDash, 64)
	if errSecret != nil {
		logger.Errorf("Could not generate scope secret: %s", errSecret)
		return errSecret
	}

	// Reset attributes
	scopeEntry.Secret = scopeSecret

	// Save reset attributes
	_, errSave := scopeEntry.Save("secret")
	if errSave != nil {
		logger.Errorf("Could not save new scope secret: %s", errSave)
		return errSave
	}

	// Create notification about reset scan scope for clients
	scopeChangeNotifier.Send(scopeEntry.Id)

	// Set scope secret in reply. Disclose only once during reset. The consumer can decide on whether to notify
	// the user. The manager is isolated and does not interact with users.
	rpcReply.ScopeSecret = scopeSecret

	// Log completion
	logger.Debugf("Scope secret reset.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetViews returns all available scope views to an RPC client
func (s *Manager) GetViews(rpcArgs *struct{}, rpcReply *ReplyScopeViews) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetViewEntries", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting views.")

	// Find all views
	viewEntries, errViewEntries := database.GetViewEntries()
	if errViewEntries != nil {
		logger.Errorf("Could not query scope views: %s", errViewEntries)
		return errViewEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, scopeView := range viewEntries {

		// Remove sensitive contents
		desensitize(&scopeView.ScanScope)

		viewEntries[i] = scopeView
	}

	// Attach list of scope views to RPC response
	rpcReply.ScopeViews = viewEntries

	// Log completion
	logger.Debugf("Views returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetViewsOf returns views of scan scopes for given group ids
func (s *Manager) GetViewsOf(rpcArgs *ArgsGroupIds, rpcReply *ReplyScopeViews) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetViewsOf", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting views of groups %s.", utils.JoinUint64(rpcArgs.GroupIds, ", "))

	// Get views of group
	viewEntries, errViewEntries := database.GetViewEntriesOf(rpcArgs.GroupIds)
	if errViewEntries != nil {
		logger.Errorf("Could not query views of %s.", utils.JoinUint64(rpcArgs.GroupIds, ", "))
		return errViewEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, scopeView := range viewEntries {

		// Remove sensitive contents
		desensitize(&scopeView.ScanScope)

		viewEntries[i] = scopeView
	}

	// Copy view data into RPC response
	rpcReply.ScopeViews = viewEntries

	// Log completion
	logger.Debugf("Views returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetViewsGranted returns views a user has granted access to
func (s *Manager) GetViewsGranted(rpcArgs *ArgsUsername, rpcReply *ReplyScopeViews) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetViewsGranted", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting granted views of user '%s'.", rpcArgs.Username)

	// Get views of user
	viewEntries, errViewEntries := database.GetViewsGranted(rpcArgs.Username)
	if errViewEntries != nil {
		logger.Errorf("Could not query views of user '%s'.", rpcArgs.Username)
		return errViewEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, scopeView := range viewEntries {

		// Remove sensitive contents
		desensitize(&scopeView.ScanScope)

		viewEntries[i] = scopeView
	}

	// Copy view data into RPC response
	rpcReply.ScopeViews = viewEntries

	// Log completion
	logger.Debugf("Views returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetView returns view for given view id
func (s *Manager) GetView(rpcArgs *ArgsViewId, rpcReply *ReplyScopeView) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetView", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting view '%d'.", rpcArgs.ViewId)

	// Get view entry based on view id
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not query view %d.", rpcArgs.ViewId)
		return errViewEntry
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	desensitize(&viewEntry.ScanScope)

	// Copy view data into RPC response
	rpcReply.ScopeViews = *viewEntry

	// Log completion
	logger.Debugf("View returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// CreateView creates a view on a scan scope with optional filters. Users can be granted access to views only.
func (s *Manager) CreateView(rpcArgs *ArgsViewDetails, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-CreateView", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting view creation.")

	// Get scope entry
	scopeEntry, errScopeEntry := database.GetScopeEntry(rpcArgs.ScopeId)
	if errScopeEntry != nil {
		logger.Errorf("Could not get scope entry to create view: %s", errScopeEntry)
		return errScopeEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, scopeEntry)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to create view '%s' on scope '%s' (ID %d): %s",
			rpcArgs.ViewName,
			scopeEntry.Name,
			scopeEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Check if same name is already existing on this scan scope
	exists, errExists := database.ViewExists(scopeEntry.Id, rpcArgs.ViewName)
	if errExists != nil {
		logger.Errorf(
			"Could not check existence of view '%s' on scope '%s' (ID %d): %s",
			rpcArgs.ViewName,
			scopeEntry.Name,
			scopeEntry.Id,
			errExists,
		)
		return errExists
	}
	if exists {
		return ErrViewNameExisting
	}

	// Create view on scope db and manager db
	_, errCreate := database.XCreateView(scopeDb, scopeEntry, rpcArgs.ViewName, rpcArgs.CreatedBy, rpcArgs.Filters)
	if errCreate != nil {
		logger.Errorf(
			"Could not create view '%s' on scan scope '%s' (ID %d): %s",
			rpcArgs.ViewName,
			scopeEntry.Name,
			scopeEntry.Id,
			errCreate,
		)
		return errCreate
	}

	// Log completion
	logger.Debugf("Scan view created.")

	// Return nil to indicate successful RPC call
	return nil
}

// DeleteView removes a scope view including associated access grants from the scope db and the manager db
func (s *Manager) DeleteView(rpcArgs *ArgsViewId, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-DeleteView", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting view deletion.")

	// Get view entry
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not get view entry to delete view: %s", errViewEntry)
		return errViewEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, &viewEntry.ScanScope)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to delete view tables '%s' (ID %s) of scope '%s' (ID %d): %s",
			viewEntry.Name,
			rpcArgs.ViewId,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Delete view in scope db and manager db
	errDelete := database.XDeleteView(scopeDb, viewEntry)
	if errDelete != nil {
		logger.Errorf(
			"Could not delete scope view '%s' (ID %d) of scan scope '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errDelete,
		)
		return errDelete
	}

	// Log completion
	logger.Debugf("Scan view deleted.")

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateView updates scope view details
func (s *Manager) UpdateView(rpcArgs *ArgsViewUpdate, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateView", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting view update.")

	// Get view entry
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not get view entry to update view: %s", errViewEntry)
		return errViewEntry
	}

	// Update view name
	viewEntry.Name = rpcArgs.Name

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, &viewEntry.ScanScope)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to update view tables '%s' (ID %s) of scope '%s' (ID %d): %s",
			viewEntry.Name,
			rpcArgs.ViewId,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Rename view in scope db and manager db
	errUpdate := database.XUpdateView(scopeDb, viewEntry)
	if errUpdate != nil {
		logger.Errorf(
			"Could not update scope view '%s' (ID %d) of scan scope '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errUpdate,
		)
		return errUpdate
	}

	// Log completion
	logger.Debugf("View updated.")

	// Return nil to indicate successful RPC call
	return nil
}

// GrantToken creates a new access token and grants it access rights to the given view and returns
// the username and password representing the access token
func (s *Manager) GrantToken(rpcArgs *ArgsGrantToken, rpcReply *ReplyCredentials) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GrantToken", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get config
	conf := config.GetConfig()

	// Log action
	logger.Debugf(
		"Client requesting to generate access token for scope view '%d'.",
		rpcArgs.ViewId,
	)

	// Get requested view entry
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not get view entry to generate access token: %s", errViewEntry)
		return errViewEntry
	}

	// Validate expiry date for maximum allowed value
	if rpcArgs.Expiry > conf.Database.TokenExpiry {
		return fmt.Errorf(
			"access token expiry time may not exceed %d days",
			int(math.Floor(conf.Database.TokenExpiry.Hours()/24)),
		)
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, &viewEntry.ScanScope)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to generate access token for scope view '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Grant user on scope db and manager db in a transactional manner
	generatedUsername, generatedPassword, errGrant := database.XGrantToken(
		scopeDb,
		viewEntry,
		rpcArgs.CreatedBy,
		rpcArgs.Expiry,
		rpcArgs.Description,
		conf.Database.ConnectionsClient,
	)
	if errGrant != nil {
		logger.Errorf(
			"Could not generate access token for scope view '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			errGrant,
		)
		return errGrant
	}

	// Set return values
	rpcReply.Username = generatedUsername
	rpcReply.Password = generatedPassword

	// Log completion
	logger.Debugf("Generated access token for scope view.")

	// Return nil to indicate successful RPC call
	return nil
}

// GrantUsers adds necessary access rights to a scope view (leaves existing ones)
func (s *Manager) GrantUsers(rpcArgs *ArgsGrantUsers, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GrantUsers", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get config
	conf := config.GetConfig()

	// Log action
	logger.Debugf(
		"Client requesting to grant access to scope view '%d' for '%d' additional users.",
		rpcArgs.ViewId,
		len(rpcArgs.DbCredentials),
	)

	// Get view entry
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not get view entry to grant access for user: %s", errViewEntry)
		return errViewEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, &viewEntry.ScanScope)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to grant access for users to scope view '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Prepare list of already granted usernames
	granted := make(map[string]struct{}, len(viewEntry.Grants))
	for _, grantEntry := range viewEntry.Grants {
		if grantEntry.IsUser { // Ignore access token grant types
			granted[grantEntry.Username] = struct{}{}
		}
	}

	// Prepare list of users excluding already set ones
	dbCredentials := make([]database.DbCredentials, 0, len(rpcArgs.DbCredentials))
	for _, credentials := range rpcArgs.DbCredentials {
		if _, existing := granted[credentials.Username]; !existing {
			dbCredentials = append(dbCredentials, credentials)
		}
	}

	// Grant user on scope db and manager db in a transactional manner
	errGrant := database.XGrantUsers(
		scopeDb,
		viewEntry,
		rpcArgs.GrantedBy,
		dbCredentials,
		conf.Database.ConnectionsClient,
	)
	if errGrant != nil {
		logger.Errorf(
			"Could not grant user to scope view '%s' (ID %d): %s", viewEntry.Name, viewEntry.Id, errGrant)
		return errGrant
	}

	// Log completion
	logger.Debugf("Granted user to scope view.")

	// Return nil to indicate successful RPC call
	return nil
}

// RevokeGrants removes an access grant (may be a user bound grant or a none user bound access token)
// from the scope db and the manager db
func (s *Manager) RevokeGrants(rpcArgs *ArgsRevokeGrants, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-RevokeGrants", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf(
		"Client requesting to revoke access from scope view '%d' for user '%s'.",
		rpcArgs.ViewId,
		rpcArgs.Usernames,
	)

	// Get view entry
	viewEntry, errViewEntry := database.GetViewEntry(rpcArgs.ViewId)
	if errViewEntry != nil {
		logger.Errorf("Could not get view entry to revoke access: %s", errViewEntry)
		return errViewEntry
	}

	// Open scope's database
	scopeDb, errScopeHandle := database.GetScopeDbHandle(logger, &viewEntry.ScanScope)
	if errScopeHandle != nil {
		logger.Errorf(
			"Could not get database scope handle to revoke access for user '%s' from scope view '%s' (ID %d): %s",
			rpcArgs.Usernames,
			viewEntry.Name,
			viewEntry.Id,
			errScopeHandle,
		)
		return errScopeHandle
	}

	// Revoke user in a transaction on scope db and manager db
	errRevoke := database.XRevokeGrants(scopeDb, viewEntry, rpcArgs.Usernames)
	if errRevoke != nil {
		logger.Errorf(
			"Could not revoke user from scope view '%s' (ID %d): %s", viewEntry.Name, viewEntry.Id, errRevoke)
		return errRevoke
	}

	// Log completion
	logger.Debugf("Revoked user from scope view.")

	// Return nil to indicate successful RPC call
	return nil
}

// UpdateServerCredentials iterates database servers the user has access rights to and updates the user's password
func (s *Manager) UpdateServerCredentials(rpcArgs *ArgsCredentials, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-UpdateServerCredentials", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get config
	conf := config.GetConfig()

	// Get view grants of user to find out which database servers are affected
	viewEntries, errViewEntries := database.GetViewsGranted(rpcArgs.Username)
	if errViewEntries != nil {
		logger.Errorf("Could not get view entries of the users: %s", errViewEntries)
		return errViewEntries
	}

	// Get database servers affected
	var executedDbServers []uint64
	for _, scopeView := range viewEntries {

		// Skip if DB server was already updated
		if utils.Uint64Contained(scopeView.ScanScope.DbServer.Id, executedDbServers) {
			continue
		}

		// Open server database
		serverDb, errServerHandle := database.GetServerDbHandle(logger, &scopeView.ScanScope.DbServer)
		if errServerHandle != nil {
			logger.Errorf("Could not get database server handle to update user password: %s", errServerHandle)
			return errServerHandle
		}

		// Calculate expiry time
		expiryTime := time.Now().Add(conf.Database.PasswordExpiry)

		// Update password with limited validity time frame
		errSet := database.UpdateServerCredentials(serverDb, rpcArgs.Username, rpcArgs.Password, expiryTime)
		if errSet != nil {
			logger.Errorf(
				"Could not update user password on database server '%s' (ID %d): %s",
				scopeView.ScanScope.DbServer.Name,
				scopeView.ScanScope.DbServer.Id,
				errSet,
			)
			return errSet
		}

		// Remember DB server
		executedDbServers = append(executedDbServers, scopeView.ScanScope.DbServer.Id)

		// No need to update the view in the slice, as we haven't modified anything
	}

	// Log completion
	logger.Debugf("Updated user password on DB servers.")

	// Return nil to indicate successful RPC call
	return nil
}

// DisableDbCredentials iterates database servers the user has access rights to and disables it's account
func (s *Manager) DisableDbCredentials(rpcArgs *ArgsUsername, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-DisableDbCredentials", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get view grants of user to find out which database servers are affected
	viewEntries, errViewEntries := database.GetViewsGranted(rpcArgs.Username)
	if errViewEntries != nil {
		logger.Errorf("Could not get view entries of the users: %s", errViewEntries)
		return errViewEntries
	}

	// Get database servers affected
	var executedDbServers []uint64
	for _, scopeView := range viewEntries {

		// Skip if DB server was already updated
		if utils.Uint64Contained(scopeView.ScanScope.DbServer.Id, executedDbServers) {
			continue
		}

		// Open server database
		serverDb, errServerHandle := database.GetServerDbHandle(logger, &scopeView.ScanScope.DbServer)
		if errServerHandle != nil {
			logger.Errorf("Could not get database server handle to disable user: %s", errServerHandle)
			return errServerHandle
		}

		// Disable user
		errSet := database.DisableServerCredentials(serverDb, rpcArgs.Username)
		if errSet != nil {
			logger.Errorf(
				"Could not disable user on database server '%s' (ID %d): %s",
				scopeView.ScanScope.DbServer.Name,
				scopeView.ScanScope.DbServer.Id,
				errSet,
			)
			return errSet
		}

		// Remember DB server
		executedDbServers = append(executedDbServers, scopeView.ScanScope.DbServer.Id)

		// No need to update the view in the slice, as we haven't modified anything
	}

	// Log completion
	logger.Debugf("User disabled on DB servers.")

	// Return nil to indicate successful RPC call
	return nil
}

// EnableDbCredentials iterates database servers the user has access rights to and enables it's account
func (s *Manager) EnableDbCredentials(rpcArgs *ArgsUsername, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-EnableDbCredentials", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get view grants of user to find out which database servers are affected
	viewEntries, errViewEntries := database.GetViewsGranted(rpcArgs.Username)
	if errViewEntries != nil {
		logger.Errorf("Could not get view entries of the users: %s", errViewEntries)
		return errViewEntries
	}

	// Get database servers affected
	var executedDbServers []uint64
	for _, scopeView := range viewEntries {

		// Skip if DB server was already updated
		if utils.Uint64Contained(scopeView.ScanScope.DbServer.Id, executedDbServers) {
			continue
		}

		// Open server database
		serverDb, errServerHandle := database.GetServerDbHandle(logger, &scopeView.ScanScope.DbServer)
		if errServerHandle != nil {
			logger.Errorf("Could not get database server handle to enable user: %s", errServerHandle)
			return errServerHandle
		}

		// Enable user
		errSet := database.EnableServerCredentials(serverDb, rpcArgs.Username)
		if errSet != nil {
			logger.Errorf(
				"Could not enable user on database server '%s' (ID %d): %s",
				scopeView.ScanScope.DbServer.Name,
				scopeView.ScanScope.DbServer.Id,
				errSet,
			)
			return errSet
		}

		// Remember DB server
		executedDbServers = append(executedDbServers, scopeView.ScanScope.DbServer.Id)

		// No need to update the view in the slice, as we haven't modified anything
	}

	// Log completion
	logger.Debugf("User enabled on DB servers.")

	// Return nil to indicate successful RPC call
	return nil
}

// GetAgents returns all available scan agents, referencing their associated scan scope
func (s *Manager) GetAgents(rpcArgs *struct{}, rpcReply *ReplyScanAgents) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetAgents", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scan agents.")

	// Find all scan agents
	agentEntries, errAgentEntries := database.GetAgentEntries()
	if errAgentEntries != nil {
		logger.Errorf("Could not query scan agents: %s", errAgentEntries)
		return errAgentEntries
	}

	// IMPORTANT: Clean sensitive scan scope data from return data
	for i, agent := range agentEntries {
		desensitize(&agent.ScanScope)

		agentEntries[i] = agent
	}

	// Attach list of scan agents to RPC response
	rpcReply.ScanAgents = agentEntries

	// Log completion
	logger.Debugf("Scan agents returned.")

	// Return nil to indicate successful RPC call
	return nil
}

// DeleteAgent removes a scan agent stats entry from the manager db
func (s *Manager) DeleteAgent(rpcArgs *ArgsAgentId, rpcReply *struct{}) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-DeleteAgent", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Log action
	logger.Debugf("Client requesting scan agent deletion.")

	// Delete scope on scope db and manager db
	errDelete := database.DeleteAgent(rpcArgs.AgentId)
	if errDelete != nil {
		logger.Errorf(
			"Could not delete scan agent with ID %d: %s",
			rpcArgs.AgentId,
			errDelete,
		)
		return errDelete
	}

	// Log completion
	logger.Debugf("Scan agent deleted.")

	// Return nil to indicate successful RPC call
	return nil
}
