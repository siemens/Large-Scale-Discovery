/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package brokerdb

import (
	"database/sql"
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"large-scale-discovery/_build"
	managerdb "large-scale-discovery/manager/database"
	"large-scale-discovery/utils"
	"strings"
	"time"
)

var brokerDb *gorm.DB // If desired public, code is most likely in the wrong package!

// Init opens the brokerdb from disk
func Init() error {

	// Open sqlite database
	// with busy timeout to avoid DB locks: https://github.com/mattn/go-sqlite3/issues/274
	// Busy timeout in milliseconds
	var errOpen error
	brokerDb, errOpen = gorm.Open(sqlite.Open("broker.sqlite?_busy_timeout=600000"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errOpen != nil {
		return fmt.Errorf("could not open manager db: %s", errOpen)
	}

	// Set DB log mode when development mode is enabled
	if _build.DevMode {
		brokerDb.Logger = brokerDb.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// Enable WAL mode for better concurrency
	brokerDb.Exec("PRAGMA journal_mode = WAL")

	// Enable foreign key support in SQLITE3 databases, where it is disabled by default -.-
	brokerDb.Exec("PRAGMA foreign_keys = ON;") // Required by SQLITE3 to enforce foreign key relations!!

	// Return nil as everything went fine
	return nil
}

// Close closes an open brokerdb
func Close() error {
	if brokerDb != nil {

		// Check for potential query optimizations and install them (to be done before closing connection)
		brokerDb.Exec("PRAGMA optimize") // https://www.sqlite.org/pragma.html#pragma_module_list

		// Retrieve and close sql db connection
		sqlDb, errDb := brokerDb.DB()
		if errDb != nil {
			return fmt.Errorf("could not retrieve underlying db connection: %s", errDb)
		}
		errClose := sqlDb.Close()
		if errClose != nil {
			return fmt.Errorf("could not close DB connection: %s", errClose)
		}
	}

	return nil
}

// AutoMigrate ScopeDb migrates the broker database's tables to the latest structure
func AutoMigrate() error {
	return brokerDb.AutoMigrate(&T_sub_input{})
}

// GetTarget queries a target from brokerdb by ID, across all scopes
func GetTarget(targetId uint64) (*T_sub_input, error) {

	// Prepare query result
	var subInput T_sub_input

	// Query entry from brokerdb
	errDb := brokerDb.Model(&subInput).
		Where("id = ?", targetId).
		First(&subInput).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return target from brokerdb
	return &subInput, nil
}

// DeleteTarget removes a given sub target from the brokerdb
func DeleteTarget(subTarget *T_sub_input) error {

	// Delete submodule target from brokerdb
	errDb := brokerDb.Delete(subTarget).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// SetTargetsStarted updates defined submodule targets in the brokerdb setting the scan_started timestamp
func SetTargetsStarted(subTargets []T_sub_input, subTargetIds []uint64, startTime sql.NullTime) error {

	// Update submodule target entries in brokerdb
	errDb := brokerDb.Model(&subTargets).
		Where("id in (?)", subTargetIds).
		Update("scan_started", startTime).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// CleanExceeded removes started scan tasks from the brokerdb that seem to be timed out
func CleanExceeded(scanScope *managerdb.T_scan_scope, label string, startedBefore time.Time) (int64, error) {

	// Delete submodule targets from brokerdb that did not return results
	db := brokerDb.
		Where("id_t_scan_scope = ?", scanScope.Id).
		Where("module = ?", label).
		Where("scan_started IS NOT NULL"). // Null would be lower than the given timestamp too...
		Where("scan_started < ?", startedBefore).
		Delete(&T_sub_input{})
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

// GetScopeSize checks whether the maximum backlog size of submodule targets (for a given scan scope) is reached
func GetScopeSize(idTScanScope uint64) (int64, error) {

	// Prepare query result
	var queueSize int64

	// Count entries in brokerdb for the given scope
	errDb := brokerDb.Model(&T_sub_input{}).
		Where("id_t_scan_scope = ?", idTScanScope).
		Where("scan_started IS NULL").
		Count(&queueSize).Error
	if errDb != nil {
		return 0, errDb
	}

	// Return queue size
	return queueSize, nil
}

// GetScopeTargets queries a given amount of submodule targets from brokerdb for the defined scan scope and module
func GetScopeTargets(idTScanScope uint64, subModule string, amount int) ([]T_sub_input, error) {

	// Prepare query result
	var modTargets []T_sub_input

	// Query amount of targets of given scan module from brokerdb
	errDb := brokerDb.
		Where("id_t_scan_scope = ?", idTScanScope).
		Where("module = ?", subModule).
		Where("scan_started IS NULL").
		Limit(amount).
		Find(&modTargets).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return found targets
	return modTargets, nil
}

// AddScopeTargets checks which submodule targets can be created from discovery results and writes them to the brokerdb.
func AddScopeTargets(logger scanUtils.Logger, scanScope *managerdb.T_scan_scope, subTargets []*T_sub_input) error {

	// Prepare slice for submodule targets that need to be created
	submoduleTargets := make([]T_sub_input, 0, len(subTargets)*2)

	// Prepare list of targets with port 445 open. Those targets should not be SMB crawled on port 139 again.
	addressesWithPort445 := make(map[string]struct{})
	for _, subTarget := range subTargets {
		if subTarget.Port == 445 {
			addressesWithPort445[subTarget.Address] = struct{}{}
		}
	}

	// Iterate target services returned by the discovery scan module
	for _, subTarget := range subTargets {

		// Skip sub target if port is sensitive one
		if utils.IntContained(subTarget.Port, scanScope.ScanSettings.SensitivePortsSlice) {
			continue
		}

		// Set scan scope ID and scans_tarted
		subTarget.IdTScanScope = scanScope.Id
		subTarget.ScanStarted = sql.NullTime{Valid: false} // Set to null scan started

		// Prepare submodule inputs from sub target, there might be multiple
		subInputs := generateSubmoduleInputs(logger, subTarget, &scanScope.ScanSettings, addressesWithPort445)

		// Get the submodules for this target
		submoduleTargets = append(
			submoduleTargets,
			subInputs...,
		)
	}

	// Check whether we have any submodule targets at all
	if len(submoduleTargets) < 1 {
		logger.Debugf("No sub targets to create")
		return nil
	}

	// Log creation
	logger.Debugf("Creating '%d' sub targets.", len(submoduleTargets))

	// Create all the submodule targets in the database. Create a transaction, so we either insert all targets or none.
	errTx := brokerDb.Transaction(func(txBroker *gorm.DB) error {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed PostgreSQLs limit of 65535 parameters
		errDb := brokerDb.
			Session(&gorm.Session{CreateBatchSize: MaxBatchSizeSubInput}).
			Create(&submoduleTargets).Error
		if errDb != nil {
			return errDb
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}

	logger.Debugf("Added %d targets to scan scope '%s'", len(submoduleTargets), scanScope.Name)

	// Return nil as everything went fine
	return nil
}

// CleanScopeTargets removes all scan targets that do not belong to an existing scan scope anymore
func CleanScopeTargets(remainingScopeIds []uint64) error {

	// Delete all entries of the associated scan scope
	errDb := brokerDb.
		Where("id_t_scan_scope NOT IN ?", remainingScopeIds).
		Delete(&T_sub_input{}).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// generateSubmoduleInputs receives a prefilled sub input template and decides which submodule target entries to generate
// from it, depending on its protocol, port, service attributes.
func generateSubmoduleInputs(
	logger scanUtils.Logger,
	subTarget *T_sub_input,
	scanSettings *managerdb.T_scan_settings,
	addressesWithPort445 map[string]struct{}, // List of addresses that have port 445 open, they shouldn't be SMB crawled on 139 too!
) []T_sub_input {

	// Prepare return slice
	subInputs := make([]T_sub_input, 0, 2)

	// Small closure that adds a new submodule target if the module is enabled
	appendSubmodule := func(label string) {

		// Skip queueing for disabled module
		instances, _ := scanSettings.MaxInstances(label) // Also returns negative number if module label is invalid
		if instances <= 0 {
			logger.Debugf("Skipping '%s' target because module is disabled.", label)
			return
		}

		// Append a copy of the generic target with the correct label set.
		subInput := *subTarget
		subInput.Module = label
		subInputs = append(subInputs, subInput)
	}

	// Decide TCP-only modules
	if subTarget.Protocol == "tcp" {

		// Check target is applicable for a banner scan (true by default)
		appendSubmodule(banner.Label)

		// Check target is applicable for an nfs scan
		if strings.Contains(subTarget.Service, "nfs") ||
			strings.Contains(subTarget.Service, "mountd") {
			appendSubmodule(nfs.Label)
		}

		// Check target is applicable for an smb scan
		if subTarget.Port == 445 {
			appendSubmodule(smb.Label)
		}
		if subTarget.Port == 139 { // Only add SMB sub input for port 139, if it isn't added for 445 too
			_, has445 := addressesWithPort445[subTarget.Address]
			if !has445 {
				appendSubmodule(smb.Label)
			}
		}

		// Check target is applicable for a webcrawler and webenum scan
		if strings.Contains(subTarget.Service, "http") ||
			strings.Contains(subTarget.ServiceProduct, "http") {
			appendSubmodule(webcrawler.Label)
			appendSubmodule(webenum.Label)
		}

		// Check target is applicable for an SSH scan
		if strings.Contains(subTarget.Service, "ssh") {
			appendSubmodule(ssh.Label)
		}

		// Check target is applicable for an ssl scan
		_, isValidSslPort := ssl.StartTlsPorts[subTarget.Port]
		if strings.Contains(subTarget.Service, "ssl") ||
			strings.Contains(subTarget.Service, "tls") ||
			strings.Contains(subTarget.Service, "https") || // searching serviceProduct for "https" delivers false positives
			strings.Contains(subTarget.ServiceProduct, "ssl") ||
			strings.Contains(subTarget.ServiceProduct, "tls") ||
			isValidSslPort {
			appendSubmodule(ssl.Label)
		}
	}

	// Return flags
	return subInputs
}
