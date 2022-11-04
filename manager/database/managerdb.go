/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlog "gorm.io/gorm/logger"
	"large-scale-discovery/_build"
	"large-scale-discovery/utils"
	"strings"
	"time"
)

var managerDb *gorm.DB // If desired public, code is most likely in the wrong package!

// OpenManagerDb initializes the manager db
func OpenManagerDb() error {

	// Initialize management database
	// With busy timeout to avoid DB locks: https://github.com/mattn/go-sqlite3/issues/274
	// Busy timeout in milliseconds
	var errOpen error
	managerDb, errOpen = gorm.Open(sqlite.Open("manager.sqlite?_busy_timeout=600000"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errOpen != nil {
		return fmt.Errorf("could not open manager db: %s", errOpen)
	}

	// Set DB log mode when development mode is enabled
	if _build.DevMode {
		managerDb.Logger = managerDb.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// Enable WAL mode for better concurrency
	managerDb.Exec("PRAGMA journal_mode = WAL")

	// Enable foreign key support in SQLITE3 databases, where it is disabled by default -.-
	managerDb.Exec("PRAGMA foreign_keys = ON;") // Required by SQLITE3 to enforce foreign key relations!!

	// Create or update the management db tables
	errMigrate := managerDb.AutoMigrate(
		&T_db_server{},
		&T_scan_scope{},
		&T_scan_settings{},
		&T_scan_agent{},
		&T_scope_view{},
		&T_view_grant{},
	)
	if errMigrate != nil {
		return fmt.Errorf("could not migrate manager db: %s", errMigrate)
	}

	// Return nil as everything went fine
	return nil
}

// CloseManagerDb closes the manager db
func CloseManagerDb() error {
	if managerDb != nil {

		// Check for potential query optimizations and install them (to be done before closing connection)
		managerDb.Exec("PRAGMA optimize") // https://www.sqlite.org/pragma.html#pragma_module_list

		// Retrieve and close sql db connection
		sqlDb, errDb := managerDb.DB()
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

// GetServerEntry queries the manager db for the database server by ID. It returns nil and a gorm.ErrRecordNotFound
// if no entry is found (check with errors.Is(...)).
func GetServerEntry(serverId uint64) (*T_db_server, error) {

	// Prepare db server memory
	var dbServer = T_db_server{}

	// Find described db server
	errDb := managerDb.
		Where("id = ?", serverId).
		First(&dbServer).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return &dbServer, nil
}

// GetScopeEntries queries the manager db for all available scan scopes including their db server details and
// scan settings. It returns an empty slice and NO error if no entry is found.
func GetScopeEntries() ([]T_scan_scope, error) {

	// Prepare query result
	var scanScopes []T_scan_scope

	// Find all scan scopes
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Find(&scanScopes).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return nil as everything went fine
	return scanScopes, nil
}

// GetScopeEntryIds retrieves a list of remaining scan scope IDs. It returns an empty slice and NO error if no entry
// is found.
func GetScopeEntryIds() ([]uint64, error) {

	// Prepare query result
	var scanScopes []T_scan_scope

	// Find all scan scopes
	errDb := managerDb.
		Find(&scanScopes).Error
	if errDb != nil {
		return nil, errDb
	}

	// Prepare list of scan scope IDs
	var scanScopeIds []uint64
	for _, scanScope := range scanScopes {
		scanScopeIds = append(scanScopeIds, scanScope.Id)
	}

	// Return list of existing scan scope IDs
	return scanScopeIds, nil
}

// GetScopeEntriesOf queries the manager db for scope entries of a given group by group ID. It returns an empty slice
// and NO error if no entry is found.
func GetScopeEntriesOf(groupIds []uint64) ([]T_scan_scope, error) {

	// Return empty slice directly
	if len(groupIds) < 1 {
		return []T_scan_scope{}, nil
	}

	// Prepare query result
	var scanScopes []T_scan_scope

	// Get the requested scan scope
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Preload("ScopeViews").
		Where("id_t_group IN (?)", groupIds).
		Find(&scanScopes).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return scanScopes, nil
}

// GetScopeEntry queries the manager db for the scope entry by ID. It returns an empty slice and a
// gorm.ErrRecordNotFound if no entry is found (check with errors.Is(...)).
func GetScopeEntry(scopeId uint64) (*T_scan_scope, error) {

	// Prepare query result
	var scanScope T_scan_scope

	// Get the requested scan scope
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Preload("ScopeViews").
		Preload("ScopeViews.Grants").
		Where("id = ?", scopeId).
		First(&scanScope).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return &scanScope, nil
}

// GetScopeEntryBySecret queries the manager db for the scope entry by secret. Returns nil and a gorm.ErrRecordNotFound
// if no scan scope could be found with the given secret (check with errors.Is(...)).
func GetScopeEntryBySecret(secret string) (*T_scan_scope, error) {

	// Prepare query result
	var scanScope T_scan_scope

	// Get the requested scan scope
	db := managerDb.
		Preload("ScanSettings").
		Preload("DbServer").
		Where("secret = ?", secret).
		Limit(1).
		Find(&scanScope) // .Limit(1).Find() instead of First() to suppress gorm.ErrRecordNotFound and related error log message if an invalid scope secret was set on a scan agent
	if db.Error != nil {
		return nil, db.Error
	}

	// Return not found error if entry could not be found
	if db.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}

	// Return scope entry
	return &scanScope, nil
}

// GetAgentEntries queries the manager db for all scan agents, grouped by scan scope. It returns an empty slice and
// NO error if no entry is found.
func GetAgentEntries() ([]T_scan_agent, error) {

	// Prepare query result
	var scanAgents []T_scan_agent

	// Find all scan scopes
	errDb := managerDb.
		Order("id_t_scan_scope asc").
		Preload("ScanScope").
		Find(&scanAgents).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return nil as everything went fine
	return scanAgents, nil
}

// UpdateScanAgents updates data of existing scan agents in the database or creates new ones. None will be removed.
func UpdateScanAgents(scopeId uint64, scanAgents []T_scan_agent) error {

	// Return early on empty slice
	if len(scanAgents) < 1 {
		return nil
	}

	// Prepare transaction to update entries within
	return managerDb.Transaction(func(txManagerDb *gorm.DB) error {

		// Make sure that the agents have the scopeId set. Copy the scan agents, to not alter the input
		agents := make([]T_scan_agent, len(scanAgents))
		for i, a := range scanAgents {
			agents[i] = a
			agents[i].IdTScanScope = scopeId
		}

		// Update columns to new value on `id_t_scan_scope`, `name`, `host`, and `ip` conflict.
		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed PostgreSQLs limit of 65535 parameters
		errDb := txManagerDb.
			Session(&gorm.Session{CreateBatchSize: MaxBatchSizeScanAgent}).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id_t_scan_scope"}, {Name: "name"}, {Name: "host"}},
				DoUpdates: clause.AssignmentColumns([]string{"ip", "last_seen", "tasks", "cpu_rate", "memory_rate", "platform", "platform_family", "platform_version"}), // Unfortunately we can not omit columns
			}).
			Create(&agents).Error
		if errDb != nil {
			return errDb
		}

		// Return nil as everything went fine
		return nil
	})
}

// GetViewEntries queries the manager db for all view entries. It returns an empty slice and NO error if no view
// entry is found at all.
func GetViewEntries() ([]T_scope_view, error) {

	// Prepare query result
	var scopeViews []T_scope_view

	// Find all scope views
	errDb := managerDb.
		Preload("ScanScope").
		Preload("Grants").
		Find(&scopeViews).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return scopeViews, nil
}

// GetViewEntriesOf queries the manager db for all view entries of a given list of group IDs. It returns an empty
// slice and NO error if no view entry is found.
func GetViewEntriesOf(groupIds []uint64) ([]T_scope_view, error) {

	// Return empty slice directly
	if len(groupIds) < 1 {
		return []T_scope_view{}, nil
	}

	// Prepare query result
	var scopeViews []T_scope_view

	// Get the requested scan scope
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.ScanSettings").
		Preload("ScanScope.ScanAgents").
		Preload("Grants").
		Joins("JOIN t_scan_scopes on t_scope_views.id_t_scan_scope = t_scan_scopes.id").
		Where("id_t_group IN (?)", groupIds).
		Find(&scopeViews).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return scopeViews, nil
}

// GetViewsGranted queries the manager db for all view entries a user has granted access. It returns an empty slice and
// NO error if no view entry is found.
func GetViewsGranted(username string) ([]T_scope_view, error) {

	// Prepare query result
	var scopeViews []T_scope_view

	// Get the requested scan scope
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.ScanSettings").
		Preload("ScanScope.ScanAgents").
		Preload("ScanScope.DbServer").
		Preload("Grants").
		Joins("JOIN T_view_grants on t_scope_views.id = T_view_grants.id_t_scope_view").
		Where("username = ?", username).
		Find(&scopeViews).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return scopeViews, nil
}

// GetView queries the manager db for the view entry by ID. It returns nil and a gorm.ErrRecordNotFound if no
// entry is found (check with errors.Is(...)).
func GetViewEntry(viewId uint64) (*T_scope_view, error) {

	// Prepare query result
	var scopeView T_scope_view

	// Get the requested view entry
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.DbServer").
		Preload("Grants").
		Where("id = ?", viewId).
		First(&scopeView).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return view entry
	return &scopeView, nil
}

// ViewExists checks whether a given view name on a scan scope does already exist.
func ViewExists(scanScopeId uint64, viewName string) (bool, error) {

	// Query scan scope by ID
	var viewsCount int64
	errDb := managerDb.
		Model(&T_scope_view{}).
		Where("id_t_scan_scope = ? AND name = ? ", scanScopeId, viewName).
		Count(&viewsCount).Error
	if errDb != nil {
		return false, errDb
	}

	// Evaluate query
	if viewsCount != 0 {
		return true, nil
	}
	return false, nil
}

// createServerEntry creates a new db server entry in the manager db.
func createServerEntry(
	txManagerDb *gorm.DB,
	dbName string, // A name that can be assigned to the server as a human understandable identifier
	dbDialect string,
	dbServer string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbHostPublic string,
	dbArgs string,
) (*T_db_server, error) {

	// Prepare development database server
	dbEntry := &T_db_server{
		Name:       dbName,
		Dialect:    dbDialect,
		Host:       dbServer,
		Port:       dbPort,
		Admin:      dbUser,
		Password:   dbPassword,
		HostPublic: dbHostPublic,
		Args:       dbArgs,
	}

	// Create scan scope
	errDb := txManagerDb.Create(&dbEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return entry
	return dbEntry, nil
}

// getServerEntryByName queries the manager db for the database server entry by name. It returns nil and a
// gorm.ErrRecordNotFound if no entry is found (check with errors.Is(...)).
func getServerEntryByName(name string) (*T_db_server, error) {

	// Prepare db server memory
	var dbServer = T_db_server{}

	// Find described db server
	db := managerDb.
		Where("name = ?", name).
		Limit(1).
		Find(&dbServer) // .Limit(1).Find() instead of First() to suppress gorm.ErrRecordNotFound and related error log message if an invalid scope secret was set on a scan agent
	if db.Error != nil {
		return nil, db.Error
	}

	// Return not found error if entry could not be found
	if db.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}

	// Return
	return &dbServer, nil
}

// createScopeEntry creates a scan scope entry in the manager db, to track scope db databases
func createScopeEntry(
	txManagerDb *gorm.DB,
	dbServer *T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	attributes utils.JsonMap,
	size uint, // Amount of target addresses deployed in this scope
	scanSettings T_scan_settings,
) (*T_scan_scope, error) {

	// Make sure, attribute table column isn't null
	if attributes == nil {
		attributes = utils.JsonMap{}
	}

	// Prepare new scan scope
	scanScope := T_scan_scope{
		IdTDbServer:     dbServer.Id,
		Name:            name,
		DbName:          dbName,
		IdTGroup:        groupId,
		Created:         time.Now(),
		CreatedBy:       createdBy,
		Secret:          secret,
		Type:            scopeType,
		Cycles:          cycles,
		CyclesRetention: cyclesRetention,
		Attributes:      attributes,
		Size:            size,
		Cycle:           1,
		CycleStarted:    time.Now(),
		CycleActive:     0,
		CycleDone:       0,
		CycleFailed:     0,
		ScanSettings:    scanSettings,
	}

	// Create scan scope
	errDb := txManagerDb.Create(&scanScope).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	scanScope.DbServer = *dbServer

	// Return nil as everything went fine
	return &scanScope, nil
}

// updateScopeInstances updates the maximum module instances per scan agent of a scan scope. It returns an
// gorm.ErrRecordNotFound if no entry is found.
func updateScopeInstances(txManagerDb *gorm.DB, scanScopeId uint64, instances map[string]uint32) error {

	// Query scan scope by ID
	var scanSettings T_scan_settings
	errDb := txManagerDb.
		Where("id_t_scan_scope = ?", scanScopeId).
		First(&scanSettings).Error
	if errDb != nil {
		return errDb
	}

	// Update max instances for every given module
	for module, instances := range instances {

		// Set module instances and select equivalent db column
		switch module {
		case discovery.Label:
			scanSettings.MaxInstancesDiscovery = instances
		case banner.Label:
			scanSettings.MaxInstancesBanner = instances
		case nfs.Label:
			scanSettings.MaxInstancesNfs = instances
		case smb.Label:
			scanSettings.MaxInstancesSmb = instances
		case ssh.Label:
			scanSettings.MaxInstancesSsh = instances
		case ssl.Label:
			scanSettings.MaxInstancesSsl = instances
		case webcrawler.Label:
			scanSettings.MaxInstancesWebcrawler = instances
		case webenum.Label:
			scanSettings.MaxInstancesWebenum = instances
		default:
			return fmt.Errorf("unknown module '%s'", module)
		}
	}

	// Save updated attributes
	errDb2 := txManagerDb.Save(&scanSettings).Error
	if errDb2 != nil {
		return errDb2
	}

	// Return nil as everything went fine
	return nil
}

// createViewEntry creates a scope view entry in the manager db, to track scope db views
func createViewEntry(
	txManagerDb *gorm.DB,
	scanScope *T_scan_scope,
	name string,
	createdBy string,
	filters map[string][]string,
	viewTableNames []string,
) (*T_scope_view, error) {

	// Convert data structure
	f := make(utils.JsonMap)
	for k, v := range filters {
		f[k] = v
	}

	// Prepare new scope view
	scopeView := T_scope_view{
		IdTScanScope: scanScope.Id,
		Name:         name,
		Created:      time.Now(),
		CreatedBy:    createdBy,
		Filters:      f,
		ViewNames:    strings.Join(viewTableNames, ","),
	}

	// Create view entry in manager db to keep track
	errDb := txManagerDb.Create(&scopeView).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	scopeView.ScanScope = *scanScope

	// Return nil as everything went fine
	return &scopeView, nil
}

// createGrantEntry creates a grant entry in the manager db, to track scope db access rights. A grant entry can
// either represent access rights for a dedicated user or for a not user bound access token.
func createGrantEntry(
	txManagerDb *gorm.DB,
	view *T_scope_view,
	isUser bool,
	username string,
	createdBy string,
	expiry time.Time,
	description string,
) (*T_view_grant, error) {

	// Prepare new scope view
	grantEntry := T_view_grant{
		IdTScopeView: view.Id,
		IsUser:       isUser,
		Username:     username,
		Created:      time.Now(),
		CreatedBy:    createdBy,
		Expiry:       expiry,
		Description:  description,
	}

	// Create view entry in manager db to keep track
	errDb := txManagerDb.Create(&grantEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	grantEntry.ScopeView = *view

	// Return nil as everything went fine
	return &grantEntry, nil
}

// serverCredentialsRequired determines whether a credentials set (user credentials or access token) is still
// required on a given database.
func serverCredentialsRequired(txManagerDb *gorm.DB, username string, dbServerId uint64) (bool, error) {

	// Check if valid DB server ID was passed
	if dbServerId == 0 {
		return true, fmt.Errorf("invalid database server ID")
	}

	// Prepare count result
	var entryCount int64

	// Search associated grant entries
	errDb := txManagerDb.Model(&T_view_grant{}).
		Joins("JOIN t_scope_views on t_view_grants.id_t_scope_view = t_scope_views.id").
		Joins("JOIN t_scan_scopes on t_scope_views.id_t_scan_scope = t_scan_scopes.id").
		Where("username = ? AND id_t_db_server = ?", username, dbServerId).
		Count(&entryCount).Error
	if errDb != nil {
		return true, errDb
	}

	// Return true if a remaining entry was found
	if entryCount > 0 {
		return true, nil
	}

	// Return
	return false, nil
}

// DeleteAgent removes a scan agent from the manager db
func DeleteAgent(agentId uint64) error {

	// Prepare query result
	scanAgent := T_scan_agent{
		Id: agentId,
	}

	errDb := scanAgent.Delete()
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}
