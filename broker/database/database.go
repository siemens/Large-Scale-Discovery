/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/_build"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// databaseShutdownTimeout limits how long database shutdown may take
const databaseShutdownTimeout = 10 * time.Second

// brokerDb holds the broker database handle
var brokerDb *gorm.DB

// Open opens the broker database from disk
func Open() error {

	// Prepare sqlite connection
	// Gorm might open additional connections, so it's important that
	// attributes are set in the DSN, rather than a subsequent PRAGMA query.
	// Attributes definition according to https://github.com/glebarez/go-sqlite.
	dsn := "file:broker.sqlite?" +
		"_pragma=journal_mode(WAL)&" + // WAL journal mode for better concurrency
		"_pragma=busy_timeout(60000)&" + // busy timeout to avoid instant database locked errors
		"_pragma=foreign_keys(1)" // with enforced foreign key constraints, otherwise the SQLITE database might ignore them

	// Open sqlite database
	db, errDb := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errDb != nil {
		return fmt.Errorf("could not open broker database: %s", errDb)
	}

	// Set DB log mode when development mode is enabled
	if _build.DevMode {
		db.Logger = db.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// Check if PRAGMA settings were applied correctly
	errPragmas := utils.CheckPragmas(db, map[string]interface{}{
		"journal_mode": "wal", // string
		"busy_timeout": 60000, // int (ms)
		"foreign_keys": 1,     // int
	})
	if errPragmas != nil {
		_ = utils.CloseDatabasePool(db)
		return fmt.Errorf("could not apply PRAGMA settings to broker database: %w", errPragmas)
	}

	// Assign db to global variable
	brokerDb = db

	// Return nil as everything went fine
	return nil
}

// Close closes the broker database and attempts all maintenance and cleanup steps.
func Close() error {

	// Detach the handle first so repeated shutdown calls remain harmless
	database := brokerDb
	brokerDb = nil
	if database == nil {
		return nil
	}

	// Bound maintenance so shutdown cannot wait indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), databaseShutdownTimeout)
	defer cancel()

	// Collect maintenance errors without skipping the mandatory pool close
	var closeErrors []error
	if errOptimize := database.WithContext(ctx).Exec("PRAGMA optimize").Error; errOptimize != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not optimize broker database: %w", errOptimize))
	}
	if errVacuum := database.WithContext(ctx).Exec("VACUUM").Error; errVacuum != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not vacuum broker database: %w", errVacuum))
	}

	// Close the pool even when either maintenance operation failed
	if errClose := utils.CloseDatabasePool(database); errClose != nil {
		closeErrors = append(closeErrors, errClose)
	}

	// Return every shutdown error after attempting all cleanup steps
	return errors.Join(closeErrors...)
}

// AutoMigrate ScopeDb migrates the broker database's tables to the latest structure
func AutoMigrate() error {
	return brokerDb.AutoMigrate(&T_sub_input{})
}

// GetTarget queries a target from broker database by ID, across all scopes
func GetTarget(targetId uint64) (*T_sub_input, error) {

	// Prepare memory for result
	var subTargetEntry T_sub_input

	// Query entry from broker database
	errDb := brokerDb.Model(&subTargetEntry).
		Where("id = ?", targetId).
		First(&subTargetEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return target from broker database
	return &subTargetEntry, nil
}

// DeleteTarget removes a given sub-target from the broker database
func DeleteTarget(subTargetEntry *T_sub_input) error {

	// Delete submodule target from broker database
	errDb := brokerDb.Delete(subTargetEntry).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// SetTargetsStarted updates defined submodule targets in the broker database setting the scan_started timestamp
func SetTargetsStarted(subTargetEntries []T_sub_input, subInputIds []uint64, startTime sql.NullTime) error {

	// Update submodule target entries in broker database
	errDb := brokerDb.Model(&subTargetEntries).
		Where("id in (?)", subInputIds).
		Update("scan_started", startTime).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// CleanExceeded removes started scan tasks from the broker database that seem to be timed out
func CleanExceeded(scopeEntry *managerdb.T_scan_scope, label string, startedBefore time.Time) (int64, error) {

	// Delete submodule targets from broker database that did not return results
	db := brokerDb.
		Where("id_t_scan_scope = ?", scopeEntry.Id).
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

	// Prepare memory for result
	var queueSize int64

	// Count entries in broker database for the given scope
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

// GetScopeStats returns per-module counts of queued (not yet started) targets for a given scan scope
func GetScopeStats(idTScanScope uint64) (utils.JsonMap, error) {

	// Query module counts grouped by module
	type moduleCount struct {
		Module string
		Count  int64
	}
	var counts []moduleCount
	errDb := brokerDb.Model(&T_sub_input{}).
		Select("module, COUNT(*) as count").
		Where("id_t_scan_scope = ?", idTScanScope).
		Group("module").
		Find(&counts).Error
	if errDb != nil {
		return nil, errDb
	}

	// Build result map
	result := make(utils.JsonMap, len(counts))
	for _, c := range counts {
		result[c.Module] = c.Count
	}

	// Return nil as everything went fine
	return result, nil
}

// GetScopeTargets queries a given amount of submodule targets from broker database for the defined scan scope and module
func GetScopeTargets(
	logger scanUtils.Logger,
	idTScanScope uint64,
	subModule string,
	amount int,
	timezonesActive []int, // List of timezone currently relevant (within configured working hours)
) ([]T_sub_input, error) {

	// Return if no timezones suite the current configuration
	if len(timezonesActive) == 0 {
		logger.Infof("No timezones within the configured working hours.")
		return []T_sub_input{}, nil
	}

	// Prepare memory for result
	var subTargetEntries []T_sub_input

	// Search for new targets
	query := brokerDb.
		Where("id_t_scan_scope = ?", idTScanScope).
		Where("module = ?", subModule).
		Where("scan_started IS NULL")

	// Attach timezone condition
	query = query.Where("timezone IN (?)", timezonesActive)

	// Add constraints
	query = query.Limit(amount)

	// Execute query
	errDb := query.Find(&subTargetEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return found targets
	return subTargetEntries, nil
}

// AddScopeTargets checks which submodule targets can be created from discovery results and writes them to the database.
func AddScopeTargets(logger scanUtils.Logger, scopeEntry *managerdb.T_scan_scope, subTargetEntriesPotential []*T_sub_input) error {

	// Prepare slice for submodule targets that need to be created
	subTargetEntries := make([]T_sub_input, 0, len(subTargetEntriesPotential)*2)

	// Prepare list of targets with port 445 open. Those targets should not be SMB crawled on port 139 again.
	addressesWithPort445 := make(map[string]struct{})
	for _, subTargetEntryPotential := range subTargetEntriesPotential {
		if subTargetEntryPotential.Port == 445 {
			addressesWithPort445[subTargetEntryPotential.Address] = struct{}{}
		}
	}

	// Iterate target services returned by the discovery scan module
	for _, subTargetEntryPotential := range subTargetEntriesPotential {

		// Skip submodule target if port is sensitive one
		if utils.IntContained(subTargetEntryPotential.Port, scopeEntry.ScanSettings.SensitivePortsSlice) {
			continue
		}

		// Set scan scope ID and scans_tarted
		subTargetEntryPotential.IdTScanScope = scopeEntry.Id
		subTargetEntryPotential.ScanStarted = sql.NullTime{Valid: false} // Set to null scan started

		// Prepare submodule inputs from sub-target, there might be multiple
		subTargetEntriesGenerated := generateSubmoduleInputs(
			logger,
			subTargetEntryPotential,
			&scopeEntry.ScanSettings,
			addressesWithPort445,
		)

		// Get the submodules for this target
		subTargetEntries = append(
			subTargetEntries,
			subTargetEntriesGenerated...,
		)
	}

	// Check whether we have any submodule targets at all
	if len(subTargetEntries) < 1 {
		logger.Debugf("No sub-targets to create")
		return nil
	}

	// Log creation
	logger.Debugf("Creating %d sub-targets.", len(subTargetEntries))

	// Create all the submodule targets in the database. Create a transaction, so we either insert all targets or none.
	errTx := brokerDb.Transaction(func(txBroker *gorm.DB) error {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := brokerDb.
			Session(&gorm.Session{CreateBatchSize: MaxBatchSizeSubInput}).
			Create(&subTargetEntries).Error
		if errDb != nil {
			return errDb
		}
		return nil
	})
	if errTx != nil {
		return errTx
	}

	logger.Debugf("Added %d targets to scan scope '%s'", len(subTargetEntries), scopeEntry.Name)

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

// CleanScopeTargetsDisabled queries which modules are currently queued for a scope and removes
// targets for any module that is now disabled (maxInstances <= 0) in the given scan settings.
func CleanScopeTargetsDisabled(logger scanUtils.Logger, scopeEntry *managerdb.T_scan_scope) error {

	// Query distinct module labels currently queued for this scope
	var queuedLabels []string
	errDb := brokerDb.
		Model(&T_sub_input{}).
		Where("id_t_scan_scope = ?", scopeEntry.Id).
		Distinct("module").
		Pluck("module", &queuedLabels).Error // Pluck scans a result into a slice
	if errDb != nil {
		return errDb
	}

	// Check each queued module against current settings and collect disabled ones
	var disabledLabels []string
	for _, label := range queuedLabels {
		maxInstances, _ := scopeEntry.ScanSettings.MaxInstances(label)
		if maxInstances <= 0 {
			disabledLabels = append(disabledLabels, label)
		}
	}

	// Return if there is nothing to clean up
	if len(disabledLabels) == 0 {
		return nil
	}

	// Log action
	logger.Infof(
		"Cleaning queued but disabled sub module targets %v from scan scope '%s' (ID %d).",
		"'"+strings.Join(disabledLabels, "' / '")+"'",
		scopeEntry.Name,
		scopeEntry.Id,
	)

	// Delete queued targets for disabled modules
	errDel := brokerDb.
		Where("id_t_scan_scope = ?", scopeEntry.Id).
		Where("module IN ?", disabledLabels).
		Where("scan_started IS NULL"). // Let the executing ones finish. Exceeded ones will be cleaned up by the normal cleanup routine.
		Delete(&T_sub_input{}).Error
	if errDel != nil {
		return errDel
	}

	// Return nil as everything went fine
	return nil
}

// generateSubmoduleInputs receives a prefilled sub input template and decides which submodule target entries to generate
// from it, depending on its protocol, port, service attributes.
func generateSubmoduleInputs(
	logger scanUtils.Logger,
	subTargetEntryTemplate *T_sub_input,
	settingsEntry *managerdb.T_scan_setting,
	addressesWithPort445 map[string]struct{}, // List of addresses that have port 445 open, they shouldn't be SMB crawled on 139 too!
) []T_sub_input {

	// Prepare return slice
	subTargetEntries := make([]T_sub_input, 0, 2)

	// Small closure that adds a new submodule target if the module is enabled
	appendSubmodule := func(label string) {

		// Skip queueing for disabled module
		maxInstances, _ := settingsEntry.MaxInstances(label) // Also returns negative number if module label is invalid
		if maxInstances <= 0 {
			logger.Debugf("Skipping '%s' target because module is disabled.", label)
			return
		}

		// Append a copy of the generic target with the correct label set.
		subTargetEntry := *subTargetEntryTemplate
		subTargetEntry.Module = label
		subTargetEntries = append(subTargetEntries, subTargetEntry)
	}

	// Decide TCP-only modules
	if subTargetEntryTemplate.Protocol == "tcp" {

		// Check target is applicable for a banner scan (true by default)
		appendSubmodule(banner.Label)

		// Check target is applicable for nuclei scan (true by default)
		appendSubmodule(nuclei.Label)

		// Check target is applicable for an nfs scan
		if strings.Contains(subTargetEntryTemplate.Service, "nfs") ||
			strings.Contains(subTargetEntryTemplate.Service, "mountd") {
			appendSubmodule(nfs.Label)
		}

		// Check target is applicable for an smb scan
		if subTargetEntryTemplate.Port == 445 {
			appendSubmodule(smb.Label)
		}
		if subTargetEntryTemplate.Port == 139 { // Only add SMB sub input for port 139, if it isn't added for 445 too
			_, has445 := addressesWithPort445[subTargetEntryTemplate.Address]
			if !has445 {
				appendSubmodule(smb.Label)
			}
		}

		// Check target is applicable for a webcrawler and webenum scan
		if strings.Contains(subTargetEntryTemplate.Service, "http") ||
			strings.Contains(subTargetEntryTemplate.ServiceProduct, "http") {
			appendSubmodule(webcrawler.Label)
			appendSubmodule(webenum.Label)
		}

		// Check target is applicable for an SSH scan
		if strings.Contains(subTargetEntryTemplate.Service, "ssh") {
			appendSubmodule(ssh.Label)
		}

		// Check target is applicable for an ssl scan
		_, isValidSslPort := ssl.StartTlsPorts[subTargetEntryTemplate.Port]
		if strings.Contains(subTargetEntryTemplate.Service, "ssl") ||
			strings.Contains(subTargetEntryTemplate.Service, "tls") ||
			strings.Contains(subTargetEntryTemplate.Service, "https") || // searching serviceProduct for "https" delivers false positives
			strings.Contains(subTargetEntryTemplate.ServiceProduct, "ssl") ||
			strings.Contains(subTargetEntryTemplate.ServiceProduct, "tls") ||
			isValidSslPort {
			appendSubmodule(ssl.Label)
		}
	}

	// Return flags
	return subTargetEntries
}
