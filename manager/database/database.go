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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlog "gorm.io/gorm/logger"
)

// databaseShutdownTimeout limits how long database shutdown may take
const databaseShutdownTimeout = 10 * time.Second

// managerDb holds the manager database handle
var managerDb *gorm.DB

// Open initializes the manager database
func Open() error {

	// Prepare sqlite connection
	// Gorm might open additional connections, so it's important that
	// attributes are set in the DSN, rather than a subsequent PRAGMA query.
	// Attributes definition according to https://github.com/glebarez/go-sqlite.
	dsn := "file:manager.sqlite?" +
		"_pragma=journal_mode(WAL)&" + // WAL journal mode for better concurrency
		"_pragma=busy_timeout(60000)&" + // busy timeout to avoid instant database locked errors
		"_pragma=foreign_keys(1)" // with enforced foreign key constraints, otherwise the SQLITE database might ignore them

	// Open sqlite database
	db, errDb := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errDb != nil {
		return fmt.Errorf("could not open manager database: %s", errDb)
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
		return fmt.Errorf("could not apply PRAGMA settings to manager database: %w", errPragmas)
	}

	// Create or update the manager database tables
	errMigrate := db.AutoMigrate(
		&T_db_server{},
		&T_scan_scope{},
		&T_scan_setting{},
		&T_scan_agent{},
		&T_scope_view{},
		&T_view_grant{},
	)
	if errMigrate != nil {
		_ = utils.CloseDatabasePool(db)
		return fmt.Errorf("could not migrate manager database: %s", errMigrate)
	}

	// Assign db to global variable
	managerDb = db

	// Return nil as everything went fine
	return nil
}

// Close closes the manager database and attempts all maintenance and cleanup steps
func Close() error {

	// Detach the handle first so repeated shutdown calls remain harmless
	database := managerDb
	managerDb = nil
	if database == nil {
		return nil
	}

	// Bound maintenance so shutdown cannot wait indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), databaseShutdownTimeout)
	defer cancel()

	// Collect maintenance errors without skipping the mandatory pool close
	var closeErrors []error
	if errOptimize := database.WithContext(ctx).Exec("PRAGMA optimize").Error; errOptimize != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not optimize manager database: %w", errOptimize))
	}
	if errVacuum := database.WithContext(ctx).Exec("VACUUM").Error; errVacuum != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not vacuum manager database: %w", errVacuum))
	}

	// Close the pool even when either maintenance operation failed
	if errClose := utils.CloseDatabasePool(database); errClose != nil {
		closeErrors = append(closeErrors, errClose)
	}

	// Return every shutdown error after attempting all cleanup steps
	return errors.Join(closeErrors...)
}

// GetDatabaseEntries queries the manager database for all database servers
func GetDatabaseEntries() ([]T_db_server, error) {

	// Prepare database server memory
	var dbServerEntries []T_db_server

	// Find described database server
	errDb := managerDb.
		Find(&dbServerEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return dbServerEntries, nil
}

// GetDatabaseEntry queries the manager database for the database server by ID.
// If no entry is found, a nil pointer is returned, make sure to check it!
func GetDatabaseEntry(dbServerId uint64) (*T_db_server, error) {

	// Prepare memory for result
	var dbServer = T_db_server{}

	// Execute query
	errDb := managerDb.
		Preload("ScanScopes").
		Where("id = ?", dbServerId).
		First(&dbServer).Error
	if errors.Is(errDb, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if errDb != nil {
		return nil, errDb
	}

	// Return result
	return &dbServer, nil
}

// UpdateDatabaseEntry updates data of existing database servers in the database or creates new ones. None will be removed.
func UpdateDatabaseEntry(dbServerEntry *T_db_server) error {

	// Decide list of attributes to update
	cols := []string{"name", "dialect", "host", "host_public", "port", "admin", "args"}
	if len(dbServerEntry.Password) > 0 { // Only set new password if defined. Otherwise, keep old one!
		cols = append(cols, "password")
	}

	// Update columns to new value on `id` conflict.
	errDb := managerDb.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns(cols), // Unfortunately we can not omit columns
		}).
		Create(&dbServerEntry).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// RemoveDatabaseEntry removes a database server from the manager database
func RemoveDatabaseEntry(dbServerId uint64) error {

	// Prepare memory for result
	dbServerEntry := T_db_server{
		Id: dbServerId,
	}

	errDb := dbServerEntry.Delete()
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// GetScopeEntries queries the manager database for all available scan scopes including their database server details and
// scan settings. It returns an empty slice and NO error if no entry is found.
func GetScopeEntries() ([]T_scan_scope, error) {

	// Prepare memory for result
	var scopeEntries []T_scan_scope

	// Find all scan scopes
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Find(&scopeEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return nil as everything went fine
	return scopeEntries, nil
}

// GetScopeEntryIds retrieves a list of remaining scan scope IDs.
// It returns an empty slice and NO error if no entry is found.
func GetScopeEntryIds() ([]uint64, error) {

	// Prepare memory for result
	var scopeEntries []T_scan_scope

	// Find all scan scopes
	errDb := managerDb.
		Find(&scopeEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Prepare list of scan scope IDs
	var scanScopeIds []uint64
	for _, scopeEntry := range scopeEntries {
		scanScopeIds = append(scanScopeIds, scopeEntry.Id)
	}

	// Return list of existing scan scope IDs
	return scanScopeIds, nil
}

// GetScopeEntriesOf queries the manager database for scope entries of a given group by group ID.
// It returns an empty slice and NO error if no entry is found.
func GetScopeEntriesOf(groupIds []uint64) ([]T_scan_scope, error) {

	// Return empty slice directly
	if len(groupIds) < 1 {
		return []T_scan_scope{}, nil
	}

	// Prepare memory for result
	var scopeEntries []T_scan_scope

	// Execute query
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Preload("ScopeViews").
		Where("id_t_group IN (?)", groupIds).
		Find(&scopeEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return result
	return scopeEntries, nil
}

// GetScopeEntry queries the manager database for the scope entry by ID.
// If no entry is found, a nil pointer is returned, make sure to check it!
func GetScopeEntry(scopeId uint64) (*T_scan_scope, error) {

	// Prepare memory for result
	var scopeEntry T_scan_scope

	// Execute query
	errDb := managerDb.
		Preload("ScanSettings").
		Preload("ScanAgents").
		Preload("DbServer").
		Preload("ScopeViews").
		Preload("ScopeViews.Grants").
		Where("id = ?", scopeId).
		First(&scopeEntry).Error
	if errors.Is(errDb, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if errDb != nil {
		return nil, errDb
	}

	// Return result
	return &scopeEntry, nil
}

// GetScopeEntryBySecret queries the manager database for the scope entry by secret.
// If no entry is found, a nil pointer is returned, make sure to check it!
func GetScopeEntryBySecret(secret string) (*T_scan_scope, error) {

	// Prepare memory for result
	var scopeEntry T_scan_scope

	// Execute query
	db := managerDb.
		Preload("ScanSettings").
		Preload("DbServer").
		Where("secret = ?", secret).
		Limit(1).
		Find(&scopeEntry) // .Limit(1).Find() instead of First() to suppress gorm.ErrRecordNotFound and related error log message if an invalid scope secret was set on a scan agent
	if db.Error != nil {
		return nil, db.Error
	}

	// Return not found error if entry could not be found
	if db.RowsAffected != 1 {
		return nil, nil
	}

	// Return result
	return &scopeEntry, nil
}

// GetScopeEntryByName returns the scan scope including its database server for a given scan scope database name.
// If no entry is found, a nil pointer is returned, make sure to check it!
func GetScopeEntryByName(dbName string) (*T_scan_scope, error) {

	// Prepare memory for result
	var scopeEntry T_scan_scope

	// Execute query
	errDb := managerDb.
		Preload("DbServer").
		Where("db_name = ?", dbName).
		First(&scopeEntry).Error
	if errors.Is(errDb, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if errDb != nil {
		return nil, errDb
	}

	// Return result
	return &scopeEntry, nil
}

// GetAgentEntries queries the manager database for all scan agents, grouped by scan scope.
// It returns an empty slice and NO error if no entry is found.
func GetAgentEntries() ([]T_scan_agent, error) {

	// Prepare memory for result
	var agentEntries []T_scan_agent

	// Find all scan scopes
	errDb := managerDb.
		Order("id_t_scan_scope asc").
		Preload("ScanScope").
		Find(&agentEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return nil as everything went fine
	return agentEntries, nil
}

// CreateAgentEntries persists new scan agent entries
func CreateAgentEntries(agentEntries []T_scan_agent) error {

	// Persist the entries without exposing the package database handle
	return managerDb.Create(&agentEntries).Error
}

// UpdateScanAgents updates data of existing scan agents in the database or creates new ones. None will be removed
func UpdateScanAgents(scopeId uint64, agentEntries []T_scan_agent) error {

	// Return early on empty slice
	if len(agentEntries) < 1 {
		return nil
	}

	// Prepare transaction to update entries within
	return managerDb.Transaction(func(txManagerDb *gorm.DB) error {

		// Make sure that the agents have the scopeId set. Copy the scan agents, to not alter the input
		agents := make([]T_scan_agent, len(agentEntries)) // Prefill with empty agent structs
		for i, a := range agentEntries {
			agents[i] = a
			agents[i].IdTScanScope = scopeId
		}

		// Update columns to new value on `id_t_scan_scope`, `name` and `host` conflict.
		errDb := txManagerDb.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id_t_scan_scope"}, {Name: "name"}, {Name: "host"}},
				DoUpdates: clause.AssignmentColumns([]string{"ip", "shared", "limits", "last_seen", "tasks", "cpu_rate", "memory_rate", "platform", "platform_family", "platform_version", "build_commit", "build_timestamp", "api_version", "version_nmap", "version_npcap", "version_sslyze", "version_python"}), // Unfortunately we can not omit columns
			}).
			Create(&agents).Error
		if errDb != nil {
			return errDb
		}

		// Return nil as everything went fine
		return nil
	})
}

// GetViewEntries queries the manager database for all view entries. It returns an empty slice and NO error if no view
// entry is found at all.
// ATTENTION: Contains sensitive information (e.g. DbServer details) which may not be yielded to clients!
func GetViewEntries() ([]T_scope_view, error) {

	// Prepare memory for result
	var viewEntries []T_scope_view

	// Find all scope views
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.DbServer").
		Preload("Grants").
		Find(&viewEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return
	return viewEntries, nil
}

// GetViewEntriesOf queries the manager database for all view entries of a given list of group IDs.
// It returns an empty slice and NO error if no view entry is found.
func GetViewEntriesOf(groupIds []uint64) ([]T_scope_view, error) {

	// Return empty slice directly
	if len(groupIds) < 1 {
		return []T_scope_view{}, nil
	}

	// Prepare memory for result
	var viewEntries []T_scope_view

	// Execute query
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.ScanSettings").
		Preload("ScanScope.ScanAgents").
		Preload("Grants").
		Joins("JOIN t_scan_scopes on t_scope_views.id_t_scan_scope = t_scan_scopes.id").
		Where("id_t_group IN (?)", groupIds).
		Find(&viewEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return result
	return viewEntries, nil
}

// GetViewsGranted queries the manager database for all view entries a user has granted access.
// It returns an empty slice and NO error if no view entry is found.
// ATTENTION: Contains sensitive information (e.g. DbServer details) which may not be yielded to clients!
func GetViewsGranted(username string) ([]T_scope_view, error) {

	// Prepare memory for result
	var viewEntries []T_scope_view

	// Execute query
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.ScanSettings").
		Preload("ScanScope.ScanAgents").
		Preload("ScanScope.DbServer").
		Preload("Grants").
		Joins("JOIN T_view_grants on t_scope_views.id = T_view_grants.id_t_scope_view").
		Where("username = ?", username).
		Find(&viewEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return result
	return viewEntries, nil
}

// GetViewEntry queries the manager database for the view entry by ID.
// If no entry is found, a nil pointer is returned, make sure to check it!
// ATTENTION: Contains sensitive information (e.g. DbServer details) which may not be yielded to clients!
func GetViewEntry(viewId uint64) (*T_scope_view, error) {

	// Prepare memory for result
	var viewEntry T_scope_view

	// Execute query
	errDb := managerDb.
		Preload("ScanScope").
		Preload("ScanScope.DbServer").
		Preload("Grants").
		Where("id = ?", viewId).
		First(&viewEntry).Error
	if errors.Is(errDb, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if errDb != nil {
		return nil, errDb
	}

	// Return result
	return &viewEntry, nil
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

// CreateDatabaseEntry creates a new database server entry in the manager database
func CreateDatabaseEntry(
	dbName string, // A name that can be assigned to the server as a human understandable identifier
	dbDialect string,
	dbServer string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbHostPublic string,
	dbArgs string,
) (*T_db_server, error) {

	// Create the entry through the initialized manager database
	return createDatabaseEntry(
		managerDb,
		dbName,
		dbDialect,
		dbServer,
		dbPort,
		dbUser,
		dbPassword,
		dbHostPublic,
		dbArgs,
	)
}

// createDatabaseEntry creates a database server entry through the supplied internal handle
func createDatabaseEntry(
	handle *gorm.DB,
	dbName string,
	dbDialect string,
	dbServer string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbHostPublic string,
	dbArgs string,
) (*T_db_server, error) {

	// Prepare development database server
	dbServerEntry := &T_db_server{
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
	errDb := handle.Create(&dbServerEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return entry
	return dbServerEntry, nil
}

// GetDatabaseEntryByName queries the manager database for the database server entry by name.
// If no entry is found, a nil pointer is returned, make sure to check it!
func GetDatabaseEntryByName(name string) (*T_db_server, error) {

	// Prepare database server memory
	var dbServer = T_db_server{}

	// Find described database server
	db := managerDb.
		Where("name = ?", name).
		Limit(1).
		Find(&dbServer) // .Limit(1).Find() instead of First() to suppress gorm.ErrRecordNotFound and related error log message if an invalid scope secret was set on a scan agent
	if db.Error != nil {
		return nil, db.Error
	}

	// Return not found error if entry could not be found
	if db.RowsAffected != 1 {
		return nil, nil
	}

	// Return
	return &dbServer, nil
}

// CreateScopeEntry creates a scan scope entry in the manager database to track a scope database
func CreateScopeEntry(
	dbServerEntry *T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	splitSize uint32,
	attributes utils.JsonMap,
	size uint, // Amount of target addresses deployed in this scope
	settingsEntry T_scan_setting,
) (*T_scan_scope, error) {

	// Create the scope through the initialized manager database
	return createScopeEntry(
		managerDb,
		dbServerEntry,
		name,
		dbName,
		groupId,
		createdBy,
		secret,
		scopeType,
		cycles,
		cyclesRetention,
		splitSize,
		attributes,
		size,
		settingsEntry,
	)
}

// createScopeEntry creates a scan scope entry through the supplied internal handle
func createScopeEntry(
	handle *gorm.DB,
	dbServerEntry *T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	splitSize uint32,
	attributes utils.JsonMap,
	size uint,
	settingsEntry T_scan_setting,
) (*T_scan_scope, error) {

	// Make sure, attribute table column isn't null
	if attributes == nil {
		attributes = utils.JsonMap{}
	}

	// Prepare new scan scope
	scopeEntry := T_scan_scope{
		IdTDbServer:     dbServerEntry.Id,
		Name:            name,
		DbName:          dbName,
		IdTGroup:        groupId,
		Created:         time.Now(),
		CreatedBy:       createdBy,
		Secret:          secret,
		Type:            scopeType,
		Cycles:          cycles,
		CyclesRetention: cyclesRetention,
		SplitSize:       splitSize,
		Attributes:      attributes,
		Size:            size,
		Cycle:           1,
		CycleStarted:    time.Now(),
		CycleActive:     0,
		CycleDone:       0,
		CycleFailed:     0,
		CycleQueue:      utils.JsonMap{},
		ScanSettings:    settingsEntry,
	}

	// Create scan scope
	errDb := handle.Create(&scopeEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	scopeEntry.DbServer = *dbServerEntry

	// Return nil as everything went fine
	return &scopeEntry, nil
}

// UpdateScopeInstances updates the maximum module instances per scan agent of a scan scope. It returns
// gorm.ErrRecordNotFound if no entry is found.
func UpdateScopeInstances(scanScopeId uint64, instances map[string]uint32) error {

	// Update the settings through the initialized manager database
	return updateScopeInstances(managerDb, scanScopeId, instances)
}

// updateScopeInstances updates scan settings through the supplied internal handle
func updateScopeInstances(handle *gorm.DB, scanScopeId uint64, instances map[string]uint32) error {

	// Query scan scope by ID
	var settingsEntry T_scan_setting
	errDb := handle.
		Where("id_t_scan_scope = ?", scanScopeId).
		First(&settingsEntry).Error
	if errDb != nil {
		return errDb
	}

	// Update max instances for every given module
	for module, moduleInstances := range instances {

		// Set module instances and select equivalent db column
		switch module {
		case discovery.Label:
			settingsEntry.MaxInstancesDiscovery = moduleInstances
		case banner.Label:
			settingsEntry.MaxInstancesBanner = moduleInstances
		case nfs.Label:
			settingsEntry.MaxInstancesNfs = moduleInstances
		case nuclei.Label:
			settingsEntry.MaxInstancesNuclei = moduleInstances
		case smb.Label:
			settingsEntry.MaxInstancesSmb = moduleInstances
		case ssh.Label:
			settingsEntry.MaxInstancesSsh = moduleInstances
		case ssl.Label:
			settingsEntry.MaxInstancesSsl = moduleInstances
		case webcrawler.Label:
			settingsEntry.MaxInstancesWebcrawler = moduleInstances
		case webenum.Label:
			settingsEntry.MaxInstancesWebenum = moduleInstances
		default:
			return fmt.Errorf("unknown module '%s'", module)
		}
	}

	// Save updated attributes
	errDb2 := handle.Save(&settingsEntry).Error
	if errDb2 != nil {
		return errDb2
	}

	// Return nil as everything went fine
	return nil
}

// CreateViewEntry creates a scope view entry in the manager database, to track scope database views
func CreateViewEntry(
	scopeEntry *T_scan_scope,
	name string,
	createdBy string,
	filters map[string][]string,
	viewTableNames []string,
) (*T_scope_view, error) {

	// Create the view through the initialized manager database
	return createViewEntry(managerDb, scopeEntry, name, createdBy, filters, viewTableNames)
}

// createViewEntry creates a scope view entry through the supplied internal handle
func createViewEntry(
	handle *gorm.DB,
	scopeEntry *T_scan_scope,
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
	viewEntry := T_scope_view{
		IdTScanScope: scopeEntry.Id,
		Name:         name,
		Created:      time.Now(),
		CreatedBy:    createdBy,
		Filters:      f,
		ViewNames:    strings.Join(viewTableNames, ","),
	}

	// Create view entry in manager database to keep track
	errDb := handle.Create(&viewEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	viewEntry.ScanScope = *scopeEntry

	// Return nil as everything went fine
	return &viewEntry, nil
}

// CreateGrantEntry creates a grant entry in the manager database, to track scope database access rights.
// A grant entry can either represent access rights for a dedicated user or for a not user bound access token.
func CreateGrantEntry(
	viewEntry *T_scope_view,
	isUser bool,
	username string,
	createdBy string,
	expiry time.Time,
	description string,
) (*T_view_grant, error) {

	// Create the grant through the initialized manager database
	return createGrantEntry(managerDb, viewEntry, isUser, username, createdBy, expiry, description)
}

// createGrantEntry creates a view grant entry through the supplied internal handle
func createGrantEntry(
	handle *gorm.DB,
	viewEntry *T_scope_view,
	isUser bool,
	username string,
	createdBy string,
	expiry time.Time,
	description string,
) (*T_view_grant, error) {

	// Prepare new scope view
	grantEntry := T_view_grant{
		IdTScopeView: viewEntry.Id,
		IsUser:       isUser,
		Username:     username,
		Created:      time.Now(),
		CreatedBy:    createdBy,
		Expiry:       expiry,
		Description:  description,
	}

	// Create view entry in manager database to keep track
	errDb := handle.Create(&grantEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Add referenced and known data to returned (newly created) struct, to avoid unnecessary lookup query
	grantEntry.ScopeView = *viewEntry

	// Return nil as everything went fine
	return &grantEntry, nil
}

// databaseCredentialsRequired determines whether a credentials set (user credentials or access token) is still
// required on a given database.
func databaseCredentialsRequired(txManagerDb *gorm.DB, username string, dbServerId uint64) (bool, error) {

	// Check if valid database server ID was passed
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

// DeleteAgent removes a scan agent from the manager database
func DeleteAgent(agentId uint64) error {

	// Prepare memory for result
	agentEntry := T_scan_agent{
		Id: agentId,
	}

	errDb := agentEntry.Delete()
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}
