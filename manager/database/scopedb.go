/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"database/sql"
	"errors"
	"fmt"
	escape "github.com/segmentio/go-pg-escape"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const dbMaxIdle = 30
const dbIdleTimeout = time.Minute * 5

// Application name to be set on the database connection (visible in the DB)
var connectionName = "Golang"

// Number of maximum parallel db connections. Postgres default max is 100 but there are connections reserve for superuser and there may be multiple applications connecting
var dbMaxConn = 20

// Map of db connection handles. Db connection handles are used to administrate databases on a database server and independent of scan scopes
var dbsLock sync.RWMutex            // Lock to control access to db handles to avoid duplicate creation of the same
var dbs = make(map[uint64]*gorm.DB) // Map of db handles (one handle per database server). We want to keep the connection open in order to let the sql driver handle connection pooling internally.

// Map of scope db connection handles. Scope db connection handles are used to edit data of a specific scan scope and cannot be used to create or delete scan scopes
var scopeDbsLock sync.RWMutex            // Lock to control access to scope db handles to avoid duplicate creation of the same
var scopeDbs = make(map[uint64]*gorm.DB) // Map of scope db handles (one handle per scope). We want to keep the connection open in order to let the sql driver handle connection pooling internally.

var ErrDbExists = errors.New("database already existing")

// ErrInvalidCharacter is a custom error for handling input parameters that contain characters that are not whitelisted
type ErrInvalidCharacter struct {
	ParamName string
	Value     string
}

func (e ErrInvalidCharacter) Error() string {
	return fmt.Sprintf("only alphanumeric and \"_\" values are allowed in %s: %s", e.ParamName, e.Value)
}

// DbCredentials contains credentials of a db user to scope db
type DbCredentials struct {
	Username string // Email address (primary key) of the user
	Password string
}

// SetConnectionName allows to set the value to use as an application name. This will only affect new database handles.
func SetConnectionName(name string) {
	name = strings.Replace(name, " ", "_", -1) // Connection name may not contain space
	connectionName = name
}

// SetMaxConnectionsDefault allows to change the default maximum database connections value. This will only affect new
// database handles.
func SetMaxConnectionsDefault(amount int) {
	dbMaxConn = amount
}

// GetServerDbHandle establishes a database connection, independent of a scan scope, configures and returns it. This
// type of connection is required to administrate (create or remove) scan scope databases on a certain the database
// server.
func GetServerDbHandle(logger scanUtils.Logger, dbServer *T_db_server) (*gorm.DB, error) {

	if dbServer == nil {
		return nil, fmt.Errorf("invalid DB connection information")
	}

	// We need to lock this function as the access to the map is not thread safe. We use a Read/Write lock as it's
	// expected that we have a lot of lookups for already established connections and just occasionally need to add a
	// new connection to the map. We could also use a concurrent map, but this seems like an overkill as long as the
	// number of entries does not get too big.
	dbsLock.RLock()

	// Check if a DB connection is already established
	if serverDb, ok := dbs[dbServer.Id]; ok && serverDb != nil {

		// Get the underlying sql connection and make sure that the connection is still alive
		sqlDb, errSql := serverDb.DB()
		if errSql == nil {
			errPing := sqlDb.Ping()
			if errPing == nil {
				dbsLock.RUnlock()
				return serverDb, nil
			}
		} else {
			return nil, errSql
		}
	}

	// Unlock the read lock, as we want to write now.
	dbsLock.RUnlock()
	dbsLock.Lock()
	defer dbsLock.Unlock()

	// Establish a new connection
	logger.Debugf("Establishing connection to DB server '%s' (ID %d).", dbServer.Name, dbServer.Id)

	// Prepare args variable
	var dsn string

	// Prepare connection string
	if dbServer.Dialect == "postgres" {
		dsn = fmt.Sprintf(
			"host='%s' port='%d' user='%s' dbname='%s' password='%s' application_name=%s %s", // application name cannot be quoted, hence cannot contain space
			dbServer.Host,
			dbServer.Port,
			dbServer.Admin,
			"postgres",
			dbServer.Password,
			connectionName,
			dbServer.Args,
		)
	} else {
		return nil, fmt.Errorf("dialect '%s' not supported", dbServer.Dialect)
	}

	// Open database. The database driver for postgres (pgx) uses prepared statements by default. We therefore do
	// not need to enable prepared statements in gorm.
	serverDb, errDB := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errDB != nil {
		return nil, errDB
	}

	// Set gorm charset
	serverDb.Set("gorm:table_options", "charsset=utf8")

	// Set development setting
	if _build.DevMode {
		serverDb.Logger = serverDb.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// We'll let the database/sql package handle the connection pooling, by setting up the number of open and idle
	// connections as well as the lifetime of connections.
	sqlDb, errSqlDb := serverDb.DB()
	if errSqlDb != nil {
		return nil, fmt.Errorf("could not get underlying sql database connection: %s", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(dbMaxConn)
	sqlDb.SetMaxIdleConns(dbMaxIdle)
	sqlDb.SetConnMaxIdleTime(dbIdleTimeout)

	// Safe the connection for consecutive handle requests
	dbs[dbServer.Id] = serverDb

	// Return successful database connection
	return serverDb, nil
}

// GetScopeDbHandle establishes a database connection to a scan scope, configures and returns it. This type of
// connection is required to edit data of a specific scan scope.
func GetScopeDbHandle(logger scanUtils.Logger, scanScope *T_scan_scope) (*gorm.DB, error) {

	if scanScope == nil || scanScope.DbServer.Id == 0 {
		return nil, fmt.Errorf("invalid DB connection information")
	}

	// We need to lock this function as the access to the map is not thread safe. We use a Read/Write lock as it's
	// expected that we have a lot of lookups for already established connections and just occasionally need to add a
	// new connection to the map. We could also use a concurrent map, but this seems like an overkill as long as the
	// number of entries does not get too big.
	scopeDbsLock.RLock()

	// Check if a DB connection is already established
	if scopeDbHandle, ok := scopeDbs[scanScope.Id]; ok && scopeDbHandle != nil {

		// Make sure that the connection is still alive
		// Get the underlying sql connection and make sure that the connection is still alive
		sqlDb, errSql := scopeDbHandle.DB()
		if errSql == nil {
			errPing := sqlDb.Ping()
			if errPing == nil {
				scopeDbsLock.RUnlock()
				return scopeDbHandle, nil
			}
		} else {
			return nil, errSql
		}
	}

	// Unlock the read lock, as we want to write now.
	scopeDbsLock.RUnlock()
	scopeDbsLock.Lock()
	defer scopeDbsLock.Unlock()

	// Establish a new connection
	logger.Debugf("Establishing connection to scan scope '%s' ('%s').", scanScope.Name, scanScope.DbName)

	// Prepare args variable
	var dsn string

	// Prepare connection string
	if scanScope.DbServer.Dialect == "postgres" {
		dsn = fmt.Sprintf(
			"host='%s' port='%d' user='%s' dbname='%s' password='%s' application_name=%s %s", // application name cannot be quoted, hence cannot contain space
			scanScope.DbServer.Host,
			scanScope.DbServer.Port,
			scanScope.DbServer.Admin,
			scanScope.DbName,
			scanScope.DbServer.Password,
			connectionName,
			scanScope.DbServer.Args,
		)
	} else {
		return nil, fmt.Errorf("dialect '%s' not supported", scanScope.DbServer.Dialect)
	}

	// Open scope database. The database driver for postgres (pgx) uses prepared statements by default. We therefore do
	// not need to enable prepared statements in gorm.
	scopeDbHandle, errDB := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errDB != nil {
		return nil, errDB
	}

	// Set gorm charset
	scopeDbHandle.Set("gorm:table_options", "charset=utf8")

	// Set development setting
	if _build.DevMode {
		scopeDbHandle.Logger = scopeDbHandle.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// We'll let the database/sql package handle the connection pooling, by setting up the number of open and idle
	// connections as well as the lifetime of connections.
	sqlDb, errSqlDb := scopeDbHandle.DB()
	if errSqlDb != nil {
		return nil, fmt.Errorf("could not get get underlying sql database connection: %s", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(dbMaxConn)
	sqlDb.SetMaxIdleConns(dbMaxIdle)
	sqlDb.SetConnMaxIdleTime(dbIdleTimeout)

	// Safe the connection for consecutive handle requests
	scopeDbs[scanScope.Id] = scopeDbHandle

	// Return successful database connection
	return scopeDbHandle, nil
}

// CloseScopeDbs will close all the open scope DB connections, it has to be called once(!) when the program finishes.
// This function will probably be called by the manager main routine on exit.
func CloseScopeDbs() []error {

	// Prepare memory for error
	errs := make([]error, 0, 1)

	// Acquire lock, as we also need to remove closed
	scopeDbsLock.RLock()
	defer scopeDbsLock.RUnlock()

	// Iterate scope databases
	for id, db := range scopeDbs {

		// Skip empty uninitialized databases connections
		if db == nil {
			continue
		}

		// Retrieve and close sql db connection
		sqlDb, errDb := db.DB()
		if errDb != nil {
			errs = append(errs, fmt.Errorf("could not retrieve underlying db connection with ID '%d': %s", id, errDb))
			continue
		}
		errClose := sqlDb.Close()
		if errClose != nil {
			errs = append(errs, fmt.Errorf("could not close DB connection with ID '%d': %s", id, errClose))
		}
	}

	// Drop references to all (now closed) DB handles
	scopeDbs = make(map[uint64]*gorm.DB)

	// Return errors if there were some
	if len(errs) != 0 {
		return errs
	}

	// Return nil as everything went fine
	return nil
}

// InstallTrigramIndices installs the trigram index database extension on a given database and applies it to
// selected table columns
func InstallTrigramIndices(scopeDb *gorm.DB) error {

	// Install Trigram index extension and apply it to common full-text searched columns
	return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// Create database
		errDb := txScopeDb.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm;`).Error
		if errDb != nil {
			return errDb
		}

		// Define full-text search indices. Please note, some simple indices are defined in the model definition
		// via Gorm struct tags.
		// ATTENTION: These indices tend to get huge in storage terms and need to be maintained on every update
		indices := []string{
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_hosts_other_names 			ON t_discovery_hosts 	USING GIN (other_names 			gin_trgm_ops)`,

			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_other_names 		ON t_discovery_services USING GIN (other_names 			gin_trgm_ops)`,        // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_os_admin_users 	ON t_discovery_services USING GIN (os_admin_users 		gin_trgm_ops)`,    // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_os_rdp_users 		ON t_discovery_services USING GIN (os_rdp_users 		gin_trgm_ops)`,       // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_asset_company 	ON t_discovery_services USING GIN (asset_company 		gin_trgm_ops)`,      // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_asset_department 	ON t_discovery_services USING GIN (asset_department 	gin_trgm_ops)`, // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_asset_owner 		ON t_discovery_services USING GIN (asset_owner 			gin_trgm_ops)`,        // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_input_manager 	ON t_discovery_services USING GIN (input_manager 		gin_trgm_ops)`,      // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_input_contact 	ON t_discovery_services USING GIN (input_contact 		gin_trgm_ops)`,      // Views are joining t_discovery_services data a lot
			`CREATE INDEX IF NOT EXISTS trgm_t_discovery_services_ad_managed_by 	ON t_discovery_services USING GIN (ad_managed_by 		gin_trgm_ops)`,      // Views are joining t_discovery_services data a lot

			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_share 						ON t_smb_files 			USING GIN (share 				gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_path 						ON t_smb_files 			USING GIN (path 				gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_name 						ON t_smb_files 			USING GIN (name 				gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_properties					ON t_smb_files 			USING GIN (properties			gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_extension					ON t_smb_files 			USING GIN (extension			gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_smb_files_mime						ON t_smb_files 			USING GIN (mime					gin_trgm_ops)`,

			`CREATE INDEX IF NOT EXISTS trgm_t_webcrawler_pages_html_content 		ON t_webcrawler_pages 	USING GIN (html_content 		gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_webcrawler_pages_response_headers 	ON t_webcrawler_pages 	USING GIN (response_headers 	gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_webcrawler_pages_raw_links 			ON t_webcrawler_pages 	USING GIN (raw_links 			gin_trgm_ops)`,

			`CREATE INDEX IF NOT EXISTS trgm_t_webenum_results_html_content 		ON t_webenum_results 	USING GIN (html_content 		gin_trgm_ops)`,
			`CREATE INDEX IF NOT EXISTS trgm_t_webenum_results_response_headers 	ON t_webenum_results 	USING GIN (response_headers 	gin_trgm_ops)`,
		}

		// Apply trigram defined indices
		for _, index := range indices {
			errDb = txScopeDb.Exec(index).Error
			if errDb != nil {
				return errDb
			}
		}

		// Return nil as everything went fine
		return nil
	})
}

// AutomigrateScanScopes initializes the scope databases
func AutomigrateScanScopes(logger scanUtils.Logger) error {

	// Find all scan scopes
	scopeEntries, errScopeEntries := GetScopeEntries()
	if errScopeEntries != nil {
		logger.Warningf("Could not query scan scopes: %s", errScopeEntries)
		return errScopeEntries
	}

	// Iterate over all scan scopes to auto-migrate databases
	for _, scanScope := range scopeEntries {

		// Open scan scope database
		scopeDb, errHandle := GetScopeDbHandle(logger, &scanScope)
		if errHandle != nil {
			return fmt.Errorf(
				"could not migrate scan scope '%s' ('%s'): %s", scanScope.Name, scanScope.DbName, errHandle)
		}

		// Log action
		logger.Infof("Migrating scan scope '%s' ('%s').", scanScope.Name, scanScope.DbName)

		// Auto-migrate scan scope database
		errMigrate := AutoMigrateScopeDb(scopeDb)
		if errMigrate != nil {
			return errMigrate
		}
	}

	// Get scope views
	viewEntries, errViewEntries := GetViewEntries()
	if errViewEntries != nil {
		logger.Warningf("Could not query scan scope views: %s", errViewEntries)
		return errViewEntries
	}

	// Iterate over all scan scope views and migrate them
	for _, viewEntry := range viewEntries {

		// Log action
		logger.Infof(
			"Migrating scan scope view '%s' of scan scope '%s' ('%s').",
			viewEntry.Name,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.DbName,
		)

		// Open scan scope database
		scopeDb, errHandle := GetScopeDbHandle(logger, &viewEntry.ScanScope)
		if errHandle != nil {
			return fmt.Errorf(
				"could not migrate scope views of scan scope '%s' ('%s'): %s",
				viewEntry.ScanScope.Name,
				viewEntry.ScanScope.DbName,
				errHandle,
			)
		}

		// Execute rebuild of view as transaction
		errTxScopeDb := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {
			return rebuildScopeView(txScopeDb, &viewEntry)
		})

		// Abort process if inner transaction failed already
		if errTxScopeDb != nil {
			return errTxScopeDb
		}
	}

	// Return nil as everything went fine
	return nil
}

// SetScopeDbComment set a comment on a scope database, making it easier to distinguish databases
func SetScopeDbComment(serverDb *gorm.DB, dbName string, dbComment string) error {

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(`COMMENT ON DATABASE %I IS %L;`, dbName, dbComment)
	if errQuery != nil {
		return errQuery
	}

	// Drop view
	return serverDb.Exec(query).Error
}

// EnableDatabaseCredentials disables a user, but leaves access rights untouched
func EnableDatabaseCredentials(serverDb *gorm.DB, username string) error {

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(`ALTER USER %I WITH LOGIN;`, username)
	if errQuery != nil {
		return errQuery
	}

	// Enable user
	errDb := serverDb.Exec(query).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// DisableDatabaseCredentials disables a user, but leaves access rights untouched
func DisableDatabaseCredentials(serverDb *gorm.DB, username string) error {

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(`ALTER USER %I WITH NOLOGIN;`, username)
	if errQuery != nil {
		return errQuery
	}

	// Kill existing connections
	errKill := killUserConnections(serverDb, username)
	if errKill != nil {
		return errKill
	}

	// Disable user
	errDb := serverDb.Exec(query).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// UpdateDatabaseCredentials updates a user's password on the database server
func UpdateDatabaseCredentials(serverDb *gorm.DB, username string, password string, expiry time.Time) error {

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(
		`ALTER USER %I WITH ENCRYPTED PASSWORD %L VALID UNTIL %L;`,
		username,
		password,
		expiry.Format(time.RFC3339),
	)
	if errQuery != nil {
		return errQuery
	}

	// Update user
	errDb := serverDb.Exec(query).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

func AutoMigrateScopeDb(scopeDb *gorm.DB) error {

	allModels := []interface{}{
		&T_discovery{},
		&T_discovery_host{},
		&T_discovery_script{},
		&T_discovery_service{},
		&T_banner{},
		&T_nfs{},
		&T_nfs_file{},
		&T_smb{},
		&T_smb_file{},
		&T_ssh{},
		&T_ssl{},
		&T_ssl_certificate{},
		&T_ssl_cipher{},
		&T_ssl_issue{},
		&T_webcrawler{},
		&T_webcrawler_vhost{},
		&T_webcrawler_page{},
		&T_webenum{},
		&T_webenum_results{},
		&T_sql_log{},
	}

	// Create or update db tables.
	if errMigrate := scopeDb.AutoMigrate(allModels...); errMigrate != nil {
		return fmt.Errorf("could not auto migrate: %s\n", errMigrate)
	}

	// Remove "NOT NULL" constraint from t_discovery_hosts that got created automatically
	// It is not desired, because t_discovery entries referenced by t_discovery_hosts may vanish over time
	errConstraint := scopeDb.Exec(`ALTER TABLE t_discovery_hosts ALTER COLUMN id_t_discovery DROP NOT NULL;`).Error
	if errConstraint != nil {
		return fmt.Errorf("could not drop standard constraint: %s", errConstraint)
	}

	// Return nil as everything went fine
	return nil
}

// SetTargets changes current scan scope targets to the given list (inserting new, deleting vanished and
// updating remaining ones)
func SetTargets(scopeDb *gorm.DB, targets []T_discovery) (created uint64, removed uint64, updated uint64, err error) {

	// Retrieve all input targets currently stored in the scope db
	var existingScopeEntries []T_discovery
	errDb := scopeDb.Model(&T_discovery{}).
		Find(&existingScopeEntries).Error
	if errDb != nil {
		return 0, 0, 0, fmt.Errorf("could not load existing inputs: %s", errDb)
	}

	// Transform existing entries into searchable data structure
	var existingInputs = make(map[string]T_discovery, len(existingScopeEntries))
	for _, existingEntry := range existingScopeEntries {
		existingInputs[existingEntry.Input] = existingEntry
	}

	// Transform new entries into searchable data structure
	var newInputs = make(map[string]T_discovery, len(targets))
	for _, target := range targets {
		newInputs[target.Input] = target
	}

	// Compare new list with existing inputs to calculate new/vanished/updated values
	createEntries, removeEntries, updateEntries := mergeInputs(existingInputs, newInputs)

	// Start transaction on the scoped db. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// TODO test what happens when broker tries to complete (update) a t_discovery service, which might
		// have been deleted by the importer in the meantime

		// Insert missing targets
		missingTargets := make([]T_discovery, 0, len(createEntries))
		for _, createEntry := range createEntries {

			// Calculate and set input size
			count, errCount := utils.CountIpsInInput(createEntry.Input)
			if errCount != nil {
				// Rollback everything if we can't insert something
				return fmt.Errorf("could not count input size: %s", errCount)
			}
			createEntry.InputSize = count
			missingTargets = append(missingTargets, createEntry)
		}

		// Execute the actual insert.
		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb2 := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: MaxBatchSizeDiscovery}).
			Create(&missingTargets).Error
		if errDb2 != nil {
			// Rollback everything if we can't insert something
			return fmt.Errorf("could not insert input: %s", errDb2)
		}

		// Delete removed targets
		var deleteIds []uint64
		for _, removeEntry := range removeEntries {
			deleteIds = append(deleteIds, removeEntry.Id)
		}
		if len(deleteIds) > 0 {
			errDb3 := txScopeDb.Delete(&T_discovery{}, deleteIds).Error
			if errDb3 != nil {
				// Rollback everything if we can't insert something
				return fmt.Errorf("could not remove input: %s", errDb3)
			}
		}

		// Update metadata of remaining targets
		for _, updateEntry := range updateEntries {

			// Execute update
			txScopeDb.
				Model(&updateEntry).
				Select(
					"enabled",
					"priority",
					"timezone",
					"lat",
					"lng",
					"postal_address",
					"input_network",
					"input_country",
					"input_location",
					"input_routing_domain",
					"input_zone",
					"input_purpose",
					"input_company",
					"input_department",
					"input_manager",
					"input_contract",
					"input_comment",
				).
				Omit( // Exclude data that should not be updated, e.g. state data that might currently be used by the broker
					"id",
					"input",
					"input_size",
					"scan_count",
					"scan_started",
					"scan_finished",
					"scan_status",
					"scan_ip",
					"scan_hostname",
				).
				Updates(&updateEntry)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return transaction error if transaction failed
	if errTx != nil {
		return 0, 0, 0, fmt.Errorf("could not set scope inputs: %s", errTx)
	}

	// Return nil as everything went fine
	return uint64(len(createEntries)), uint64(len(removeEntries)), uint64(len(updateEntries)), nil
}

// GetTargets queries the current inputs from the scan scope
func GetTargets(scopeDb *gorm.DB) ([]T_discovery, error) {

	// Prepare memory for query result
	var targets []T_discovery

	// Query total inputs
	errDb := scopeDb.Model(&T_discovery{}).
		Order("id").
		Find(&targets).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return counts
	return targets, nil
}

// GetProgress queries a given scope db for completed, failed and running scan inputs (progress)
func GetProgress(scopeDb *gorm.DB) (total int64, done int64, running int64, failed int64, err error) {

	// Query total inputs
	errDb := scopeDb.Model(&T_discovery{}).
		Where("enabled IS TRUE").
		Count(&total).Error
	if errDb != nil {
		err = errDb
		return
	}

	// Query done inputs
	errDb2 := scopeDb.Model(&T_discovery{}).
		Where("scan_started IS NOT NULL").
		Where("scan_finished IS NOT NULL").
		Where("enabled IS TRUE").
		Where(
			"scan_status = ? OR scan_status = ? OR scan_status = ? OR scan_status = ?",
			scanUtils.StatusCompleted,
			scanUtils.StatusDeadline,
			scanUtils.StatusNotReachable,
			scanUtils.StatusSkipped,
		).
		Count(&done).Error
	if errDb2 != nil {
		err = errDb2
		return
	}

	// Query running inputs
	errDb3 := scopeDb.Model(&T_discovery{}).
		Where("scan_started IS NOT NULL").
		Where("scan_finished IS NULL").
		Where("enabled IS TRUE").
		Count(&running).Error
	if errDb3 != nil {
		err = errDb3
		return
	}

	// Query failed inputs
	errDb4 := scopeDb.Model(&T_discovery{}).
		Where("scan_started IS NOT NULL").
		Where("scan_finished IS NOT NULL").
		Where("enabled IS TRUE").
		Where("scan_status = ? OR scan_status = ?", scanUtils.StatusFailed, scanUtils.StatusProxyError).
		Count(&failed).Error
	if errDb4 != nil {
		err = errDb4
		return
	}

	// Return counts
	return
}

// ResetFailed resets the scan status of failed scan inputs in order to trigger a rescan within the current scan cycle
func ResetFailed(scopeDb *gorm.DB) error {

	// Execute cycle reset by resetting all scan input states
	errScopeDb := scopeDb.Model(&T_discovery{}).
		Where("scan_started IS NOT NULL").
		Where("scan_finished IS NOT NULL").
		Where("scan_status = ?", scanUtils.StatusFailed).
		Updates(map[string]interface{}{
			"scan_started":  sql.NullTime{},
			"scan_finished": sql.NullTime{},
			"scan_status":   scanUtils.StatusWaiting,
		}).Error
	if errScopeDb != nil {
		return fmt.Errorf("could not reset failed scan targets in scope db: %s", errScopeDb)
	}

	// Return nil as everything went fine
	return nil
}

// createDatabaseCredentials creates a new user in the database, if not yet existing
func createDatabaseCredentials(serverDb *gorm.DB, username string, password string, expiry time.Time, connections int) error {

	// Convert connection value to string to build query
	c := strconv.Itoa(connections)

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(
		`CREATE USER %I WITH LOGIN NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION ENCRYPTED PASSWORD %L VALID UNTIL %L CONNECTION LIMIT %s;`,
		username,
		password,
		expiry.Format(time.RFC3339),
		c,
	)
	if errQuery != nil {
		return errQuery
	}

	// Create user
	errDb := serverDb.Exec(query).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// deleteDatabaseCredentials removes a given user from database server. If a user is still referenced by some table,
// it can't be deleted. In case of such error, the error will be ignored and the user remains.
func deleteDatabaseCredentials(txServerDb *gorm.DB, username string) {

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, _ := escape.Escape(`DROP USER IF EXISTS %I;`, username)

	// Prepare savepoint within transaction. Dropping the user might fail if the user is still required for some
	// other tables or databases. In that case we just want to ignore the error and continue as if it didn't happen.
	txServerDb.SavePoint("sp1")

	// Drop user
	errDb := txServerDb.Exec(query).Error
	if errDb != nil {
		txServerDb.RollbackTo("sp1")
	}
}

// killUserConnections kills active database sessions of a given user. This should be called in certain cases,
// where a user should not be able to continue using the database (e.g. user gets deleted or disabled). Make sure
// re-connection is prevented, before calling, otherwise connections might be re-established automatically.
func killUserConnections(serverDb *gorm.DB, username string) error {
	return serverDb.Exec(`
		SELECT
			pg_terminate_backend(pid)
		FROM
			pg_stat_activity
		WHERE
			usename = ?
	`, username).Error
}

// createScopeDb creates a new scan database in the scope db. This action cannot run within a transaction!
func createScopeDb(serverDb *gorm.DB, name string, comment string) error {

	// Count databases with given name
	var dbCount int64
	errDb := serverDb.
		Table("pg_catalog.pg_database").
		Where("datname = ?", name).
		Count(&dbCount).Error
	if errDb != nil {
		return errDb
	}

	// Check if db exists
	if dbCount > 0 {
		return ErrDbExists
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlCreate, errSqlCreate := escape.Escape(`CREATE DATABASE %I WITH ENCODING='UTF8';`, name)
	if errSqlCreate != nil {
		return errSqlCreate
	}

	// Create database
	errDb2 := serverDb.Exec(sqlCreate).Error
	if errDb2 != nil {
		return errDb2
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlRevoke, errSqlRevoke := escape.Escape(`REVOKE ALL ON DATABASE %I FROM PUBLIC;`, name)
	if errSqlRevoke != nil {
		return errSqlRevoke
	}

	// Deny database connection for public
	errDb3 := serverDb.Exec(sqlRevoke).Error
	if errDb3 != nil {
		return errDb3
	}

	// Update database comment for users browsing accessible databases to distinguish them. The
	// randomly generated database names are not easy to tell apart.
	errComment := SetScopeDbComment(serverDb, name, comment)
	if errComment != nil {
		return fmt.Errorf("could not update database comment: %s", errComment)
	}

	// Return nil as everything went fine
	return nil
}

// deleteScopeDb removes a scan scope database from a database server. This action cannot run within a transaction!
// ATTENTION: This action cannot be rolled back!
func deleteScopeDb(serverDb *gorm.DB, name string) error {

	// Disconnect all connected clients
	errDb := serverDb.Exec(`
		SELECT 
			pg_terminate_backend(pg_stat_activity.pid)
		FROM 
			pg_stat_activity
		WHERE 
			pg_stat_activity.datname = ? AND pid <> pg_backend_pid();
	`, name).Error
	if errDb != nil {
		return errDb
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	query, errQuery := escape.Escape(`DROP DATABASE %I;`, name)
	if errQuery != nil {
		return errQuery
	}

	// Drop database
	errDb = serverDb.Exec(query).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// createScopeView creates a set of view tables belonging new scope view (there are multiple view tables per view:
// hosts, services, smb, nfs,...).
func createScopeView(txScopeDb *gorm.DB, name string, filters map[string][]string) ([]string, error) {

	// Prepare storage for generated view names
	var viewNames []string

	// Prepare gigantic where clause based on defined filters
	subClauses := make([]string, 0, len(filters))
	for column, values := range filters {
		values = scanUtils.TrimToLower(values)
		if len(values) > 0 {
			clauseValues := "'" + strings.Join(values, "', '") + "'"
			clause := fmt.Sprintf("LOWER(%s) IN (%s)", column, clauseValues)
			subClauses = append(subClauses, clause)
		}
	}

	// Combine them as a OR clause
	whereClause := strings.Join(subClauses, " OR ")

	// Create views
	for view, definition := range viewDefinitions {

		// Define view name
		viewName, errSanitize := sanitizeViewName(name + "_" + view)
		if errSanitize != nil {
			return nil, fmt.Errorf("could not sanitize view name: %s", errSanitize)
		}

		// Remember view name
		viewNames = append(viewNames, viewName)

		// Build create query from template
		query := strings.Replace(definition, "?", `"`+viewName+`"`, -1) // Cannot be done with prepared statements

		// Strip trailing semicolons if available (because the where clause is going to get attached
		query = strings.TrimRight(query, "\r\n\t ;")

		// Attach where clause
		if len(whereClause) > 0 {
			query = query + " WHERE " + whereClause
		}

		// Create view
		errDb := txScopeDb.Exec(query).Error
		if errDb != nil {
			return nil, errDb
		}
	}

	// Return nil as everything went fine
	return viewNames, nil
}

// deleteScopeView removes a set of view tables belonging to a scope view (there are multiple view  tables per view:
// hosts, services, smb, nfs,...).
func deleteScopeView(txScopeDb *gorm.DB, viewTableNames []string) error {

	// Iterate views and drop them
	for _, viewName := range viewTableNames {

		// Build escaped query manually, as it can't be executed as a prepared statement
		// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
		query, errQuery := escape.Escape(`DROP VIEW IF EXISTS %I;`, viewName)
		if errQuery != nil {
			return errQuery
		}

		// Drop view
		errDb := txScopeDb.Exec(query).Error
		if errDb != nil {
			return errDb
		}
	}

	// Return nil as everything went fine
	return nil
}

// rebuildScopeView rebuilds a scan scope views in the scope db based on the last definition and re-grants
// authorized users. This might be necessary, if the view definition has changed, e.g. if columns were
// introduced, removed or changed in their order. It can also be used to pull access rights straight to match
// the actual configuration.
func rebuildScopeView(txScopeDb *gorm.DB, viewEntry *T_scope_view) error {

	// Return if there are no views tables to be updated
	if viewEntry.ViewNames == "" {
		return nil
	}

	// Generate list of view table names from manager db
	viewTableNames := utils.ToSlice(viewEntry.ViewNames, ",")

	// Drop view
	errDelete := deleteScopeView(txScopeDb, viewTableNames)
	if errDelete != nil {
		return fmt.Errorf("could not rebuild view: %s", errDelete)
	}

	// Cast filters types
	filters := make(map[string][]string)
	for key, val := range viewEntry.Filters {
		interfaceValues := val.([]interface{})
		stringValues := make([]string, 0, len(interfaceValues))
		for _, interfaceValue := range interfaceValues {
			stringValues = append(stringValues, interfaceValue.(string))
		}
		filters[key] = stringValues
	}

	// Create view
	_, errCreate := createScopeView(txScopeDb, viewEntry.Name, filters)
	if errCreate != nil {
		return fmt.Errorf("could not rebuild view: %s", errCreate)
	}

	// Iterate granted users
	for _, grant := range viewEntry.Grants {
		errGrant := grantScopeView(
			txScopeDb,
			viewEntry,
			DbCredentials{Username: grant.Username},
			time.Now(),
			0, // User should already exist and doesn't need to be created/configured
		)
		if errGrant != nil {
			return fmt.Errorf("could not rebuild view: %s", errGrant)
		}
	}

	// Return nil as everything went fine
	return nil
}

// grantScopeView grants access for a credentials set to a set of view tables belonging to a scope view (there
// are multiple view tables per view: hosts, services, smb, nfs,...).
// ATTENTION: A credentials set may either be user specific or a random access token
// ATTENTION: If the credentials set is not existing on the given database server, it will be created. That's why
//
//	the hashed user's password needs to be passed in too.
func grantScopeView(
	txScopeDb *gorm.DB,
	viewEntry *T_scope_view,
	credentials DbCredentials,
	expiry time.Time,
	connections int,
) error {

	// Check if user exists
	var userCount int64
	errDb := txScopeDb.
		Table("pg_roles").
		Where("rolname = ?", credentials.Username).
		Count(&userCount).Error
	if errDb != nil {
		return errDb
	}

	// Create user if not existing on database server
	if userCount == 0 {

		// Create credentials
		errCreate := createDatabaseCredentials(
			txScopeDb, credentials.Username, credentials.Password, expiry, connections)
		if errCreate != nil {
			return errCreate
		}
	}

	// By default, every user belongs to the role "public" and is thereby allowed to create tables in the default
	// schema. In order to prevent this, the rights of "public" have to be revoked from the default schema. This
	// action requires a direct connection to the respective database, so it could not be done during database creation.
	// As a work-around, this is done now before the first user is granted [... and, as a side-effect re-executed
	// every time another user is granted]
	errDb2 := txScopeDb.Exec(`REVOKE ALL ON schema public FROM public;`).Error
	if errDb2 != nil {
		return errDb2
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlGrant3, errSqlGrant3 := escape.Escape(
		`GRANT CONNECT ON DATABASE %I TO %I`,
		viewEntry.ScanScope.DbName,
		credentials.Username,
	)
	if errSqlGrant3 != nil {
		return errSqlGrant3
	}

	// Grant connect right
	errDb3 := txScopeDb.Exec(sqlGrant3).Error
	if errDb3 != nil {
		return errDb3
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlGrant4, errSqlGrant4 := escape.Escape(`GRANT USAGE ON SCHEMA public TO %I`, credentials.Username)
	if errSqlGrant4 != nil {
		return errSqlGrant4
	}

	// Grant usage on schema
	errDb4 := txScopeDb.Exec(sqlGrant4).Error
	if errDb4 != nil {
		return errDb4
	}

	// Get list of view table names from manager db
	viewTableNames := utils.ToSlice(viewEntry.ViewNames, ",")

	// Grant rights to all views
	for _, viewName := range viewTableNames {

		// Build escaped query manually, as it can't be executed as a prepared statement
		// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
		sqlGrant5, errSqlGrant5 := escape.Escape(`GRANT SELECT ON %I TO %I`, viewName, credentials.Username)
		if errSqlGrant5 != nil {
			return errSqlGrant5
		}

		// Create right on view
		errDb5 := txScopeDb.Exec(sqlGrant5).Error
		if errDb5 != nil {
			return errDb5
		}
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlGrant6, errSqlGrant6 := escape.Escape(`GRANT SELECT ON t_discovery TO %I`, credentials.Username)
	if errSqlGrant6 != nil {
		return errSqlGrant6
	}

	// Grant right on t_discovery, so the user can check input and progress
	errDb6 := txScopeDb.Exec(sqlGrant6).Error
	if errDb6 != nil {
		return errDb6
	}

	// Return nil as everything went fine
	return nil
}

// revokeScopeView revokes user access from a set of view tables belonging to a scope view (there are multiple view
// tables per view: hosts, services, smb, nfs,...).
// ATTENTION: If cleanUser is true, the given user and remaining rights are completely removed from the database server.
func revokeScopeView(txScopeDb *gorm.DB, dbName, username string, viewTableNames []string) error {

	// Revoke rights from all views
	for _, viewTable := range viewTableNames {

		// Build escaped query manually, as it can't be executed as a prepared statement
		// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
		sqlRevoke, errSqlRevoke := escape.Escape(`REVOKE ALL ON %I FROM %I`, viewTable, username)
		if errSqlRevoke != nil {
			return errSqlRevoke
		}

		// Revoke right from view table
		errDb := txScopeDb.Exec(sqlRevoke).Error // Cannot be done with prepared statement
		if errDb != nil {
			return errDb
		}
	}

	// Build escaped query manually, as it can't be executed as a prepared statement
	// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
	sqlRevoke2, errSqlRevoke2 := escape.Escape(`REVOKE ALL ON t_discovery FROM %I`, username)
	if errSqlRevoke2 != nil {
		return errSqlRevoke2
	}

	// Revoke right from t_discovery
	errDb2 := txScopeDb.Exec(sqlRevoke2).Error
	if errDb2 != nil {
		return errDb2
	}

	// Count if the user still has any select rights left on the scope db
	var rightsCount int64
	errDb3 := txScopeDb.
		Table("information_schema.role_table_grants").
		Where("grantee = ?", username).
		Where("privilege_type = ?", "SELECT").
		Count(&rightsCount).Error
	if errDb3 != nil {
		return errDb3
	}

	// Clean up remaining rights from the related scope db
	if rightsCount == 0 {

		// Build escaped query manually, as it can't be executed as a prepared statement
		// ATTENTION: This is tailored for Postgres databases and might not be safe with others!
		sqlRevoke4, errSqlRevoke4 := escape.Escape(`
			REVOKE ALL ON DATABASE %I FROM %I;
			REVOKE ALL ON SCHEMA public FROM %I;
			REVOKE ALL ON ALL TABLES IN SCHEMA public FROM %I;
			REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM %I;
		`, dbName, username, username, username, username)
		if errSqlRevoke4 != nil {
			return errSqlRevoke4
		}

		errDb4 := txScopeDb.Exec(sqlRevoke4).Error
		if errDb4 != nil {
			return errDb4
		}
	}

	// Return nil as everything went fine
	return nil
}

// sanitizeViewName strips illegal view name characters and transforms the name into unified format
func sanitizeViewName(name string) (string, error) {

	// Unify prefix (originally user input)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove all none alphanumeric characters
	reg, errReg := regexp.Compile("[^a-zA-Z0-9_]+")
	if errReg != nil {
		return "", errReg
	}
	name = reg.ReplaceAllString(name, "")
	return name, nil
}

// mergeInputs determines which entries shall be created/removed/updated comparing a new list with an old one
func mergeInputs(existingInputs map[string]T_discovery, newInputs map[string]T_discovery) (
	create []T_discovery, remove []T_discovery, update []T_discovery) {

	// Find added values that need to be inserted
	for k, v := range newInputs {
		_, exists := existingInputs[k]
		if !exists {
			create = append(create, v)
		}
	}

	// Find removed values that need to be deleted
	for k, v := range existingInputs {
		_, exists := newInputs[k]
		if !exists {
			remove = append(remove, v)
		}
	}

	// Find remaining entries that need to be updated
	for k, v := range newInputs {
		existing, exists := existingInputs[k]
		if exists {

			// Update database ID in current entry from existing entry
			v.Id = existing.Id

			// Add entry to list of update entries
			update = append(update, v)
		}
	}

	// Return sorted lists
	return create, remove, update
}
