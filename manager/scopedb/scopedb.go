/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package scopedb

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/go-pg-escape"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

var connectionName = "LSD"                                // Application name shown on database connections
var dbMaxConn = 5                                         // Maximum parallel connections per database pool
var dbMaxIdle = 2                                         // Maximum idle database connections
var dbIdleTimeout = time.Minute * 5                       // Maximum lifetime of an idle database connection
var dbMigrationReapInterval = time.Second * 5             // Interval at which idle sessions blocking a migration are reaped
var dbMigrationIdleGrace = time.Minute * 3                // Grace before an idle-in-transaction session blocking a migration is terminated
var dbsLock sync.RWMutex                                  // Guards the database-server handle cache
var dbs = make(map[uint64]*gorm.DB)                       // Caches one pooled handle per database server
var scopeDbsLock sync.RWMutex                             // Guards the scope-database handle cache
var scopeDbs = make(map[uint64]*gorm.DB)                  // Caches one pooled handle per scope database
var ErrDbExists = errors.New("database already existing") // Indicates that a requested database already exists

// ErrInvalidCharacter is a custom error for handling input parameters that contain characters that are not whitelisted
type ErrInvalidCharacter struct {
	ParamName string
	Value     string
}

func (e ErrInvalidCharacter) Error() string {
	return fmt.Sprintf("only alphanumeric and \"_\" values are allowed in %s: %s", e.ParamName, e.Value)
}

// DbCredentials contains credentials of a db user to scope database
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
func GetServerDbHandle(logger scanUtils.Logger, dbServerEntry *database.T_db_server) (*gorm.DB, error) {

	// Reject calls without the metadata required to identify a cache entry
	if dbServerEntry == nil {
		return nil, fmt.Errorf("invalid DB connection information")
	}

	// We need to lock this function as the access to the map is not thread safe. We use a Read/Write lock as it's
	// expected that we have a lot of lookups for already established connections and just occasionally need to add a
	// new connection to the map. We could also use a concurrent map, but this seems like an overkill as long as the
	// number of entries does not get too big.
	dbsLock.RLock()

	// Check if a DB connection is already established
	serverDb := dbs[dbServerEntry.Id]
	dbsLock.RUnlock()
	if serverDb != nil {

		// Get the underlying sql connection and make sure that the connection is still alive
		sqlDb, errSqlDb := serverDb.DB()
		if errSqlDb != nil {
			return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
		}
		if errPing := sqlDb.Ping(); errPing == nil {
			return serverDb, nil
		}
	}

	// Serialize connection creation and check the cache again because another caller may have populated it while
	// this caller was waiting for the write lock
	dbsLock.Lock()
	defer dbsLock.Unlock()
	if serverDb = dbs[dbServerEntry.Id]; serverDb != nil {
		sqlDb, errSqlDb := serverDb.DB()
		if errSqlDb != nil {
			return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
		}
		if errPing := sqlDb.Ping(); errPing == nil {
			return serverDb, nil
		}
	}

	// Establish a new connection
	logger.Debugf("Establishing connection to DB server '%s' (ID %d).", dbServerEntry.Name, dbServerEntry.Id)

	// Prepare args variable
	var dsn string

	// Prepare connection string
	if dbServerEntry.Dialect == "postgres" {
		dsn = fmt.Sprintf(
			"host='%s' port='%d' user='%s' dbname='%s' password='%s' application_name=%s %s", // application name cannot be quoted, hence cannot contain space
			dbServerEntry.Host,
			dbServerEntry.Port,
			dbServerEntry.Admin,
			"postgres",
			dbServerEntry.Password,
			connectionName,
			dbServerEntry.Args,
		)
	} else {
		return nil, fmt.Errorf("dialect '%s' not supported", dbServerEntry.Dialect)
	}

	// Open database. The database driver for postgres (pgx) uses prepared statements by default. We therefore do
	// not need to enable prepared statements in gorm.
	newServerDb, errNewServerDb := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errNewServerDb != nil {
		_ = utils.CloseDatabasePool(newServerDb)
		return nil, fmt.Errorf("could not open new database connection: %s", errNewServerDb)
	}

	// Set gorm charset
	newServerDb.Set("gorm:table_options", "charsset=utf8")

	// Set development setting
	if _build.DevMode {
		newServerDb.Logger = newServerDb.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// We'll let the database/sql package handle the connection pooling, by setting up the number of open and idle
	// connections as well as the lifetime of connections.
	sqlDb, errSqlDb := newServerDb.DB()
	if errSqlDb != nil {
		_ = utils.CloseDatabasePool(newServerDb)
		return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(dbMaxConn)
	sqlDb.SetMaxIdleConns(dbMaxIdle)
	sqlDb.SetConnMaxIdleTime(dbIdleTimeout)

	// Replace the stale cached handle and release its underlying pool
	staleServerDb := dbs[dbServerEntry.Id]
	dbs[dbServerEntry.Id] = newServerDb
	if staleServerDb != nil {
		if errClose := utils.CloseDatabasePool(staleServerDb); errClose != nil {
			logger.Warningf("Could not close stale DB server handle with ID %d: %s", dbServerEntry.Id, errClose)
		}
	}

	// Return successful database connection
	return newServerDb, nil
}

// GetScopeDbHandle establishes a database connection to a scan scope, configures and returns it. This type of
// connection is required to edit data of a specific scan scope.
func GetScopeDbHandle(logger scanUtils.Logger, scopeEntry *database.T_scan_scope) (*gorm.DB, error) {

	// Reject calls without the metadata required to identify a cache entry
	if scopeEntry == nil || scopeEntry.DbServer.Id == 0 {
		return nil, fmt.Errorf("invalid DB connection information")
	}

	// We need to lock this function as the access to the map is not thread safe. We use a Read/Write lock as it's
	// expected that we have a lot of lookups for already established connections and just occasionally need to add a
	// new connection to the map. We could also use a concurrent map, but this seems like an overkill as long as the
	// number of entries does not get too big.
	scopeDbsLock.RLock()

	// Check if a DB connection is already established
	scopeDbHandle := scopeDbs[scopeEntry.Id]
	scopeDbsLock.RUnlock()
	if scopeDbHandle != nil {

		// Make sure that the connection is still alive
		// Get the underlying sql connection and make sure that the connection is still alive
		sqlDb, errSqlDb := scopeDbHandle.DB()
		if errSqlDb != nil {
			return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
		}
		if errPing := sqlDb.Ping(); errPing == nil {
			return scopeDbHandle, nil
		}
	}

	// Serialize connection creation and check the cache again because another caller may have populated it while
	// this caller was waiting for the write lock
	scopeDbsLock.Lock()
	defer scopeDbsLock.Unlock()
	if scopeDbHandle = scopeDbs[scopeEntry.Id]; scopeDbHandle != nil {
		sqlDb, errSqlDb := scopeDbHandle.DB()
		if errSqlDb != nil {
			return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
		}
		if errPing := sqlDb.Ping(); errPing == nil {
			return scopeDbHandle, nil
		}
	}

	// Establish a new connection
	logger.Debugf("Establishing connection to scan scope '%s' ('%s').", scopeEntry.Name, scopeEntry.DbName)

	// Prepare args variable
	var dsn string

	// Prepare connection string
	if scopeEntry.DbServer.Dialect == "postgres" {
		dsn = fmt.Sprintf(
			"host='%s' port='%d' user='%s' dbname='%s' password='%s' application_name=%s %s", // application name cannot be quoted, hence cannot contain space
			scopeEntry.DbServer.Host,
			scopeEntry.DbServer.Port,
			scopeEntry.DbServer.Admin,
			scopeEntry.DbName,
			scopeEntry.DbServer.Password,
			connectionName,
			scopeEntry.DbServer.Args,
		)
	} else {
		return nil, fmt.Errorf("dialect '%s' not supported", scopeEntry.DbServer.Dialect)
	}

	// Open scope database. The database driver for postgres (pgx) uses prepared statements by default. We therefore do
	// not need to enable prepared statements in gorm.
	newScopeDbHandle, errNewScopeDbHandle := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errNewScopeDbHandle != nil {
		_ = utils.CloseDatabasePool(newScopeDbHandle)
		return nil, fmt.Errorf("could not open new database connection: %s", errNewScopeDbHandle)
	}

	// Set gorm charset
	newScopeDbHandle.Set("gorm:table_options", "charset=utf8")

	// Set development setting
	if _build.DevMode {
		newScopeDbHandle.Logger = newScopeDbHandle.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// We'll let the database/sql package handle the connection pooling, by setting up the number of open and idle
	// connections as well as the lifetime of connections.
	sqlDb, errSqlDb := newScopeDbHandle.DB()
	if errSqlDb != nil {
		_ = utils.CloseDatabasePool(newScopeDbHandle)
		return nil, fmt.Errorf("could not get underlying sql database connection: %w", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(dbMaxConn)
	sqlDb.SetMaxIdleConns(dbMaxIdle)
	sqlDb.SetConnMaxIdleTime(dbIdleTimeout)

	// Replace the stale cached handle and release its underlying pool
	staleScopeDb := scopeDbs[scopeEntry.Id]
	scopeDbs[scopeEntry.Id] = newScopeDbHandle
	if staleScopeDb != nil {
		if errClose := utils.CloseDatabasePool(staleScopeDb); errClose != nil {
			logger.Warningf("Database handle with ID %d %s", scopeEntry.Id, errClose)
		}
	}

	// Return successful database connection
	return newScopeDbHandle, nil
}

// CloseScopeDbs closes all scope database handles
func CloseScopeDbs() []error {
	return closeDbHandles(&scopeDbsLock, &scopeDbs)
}

// CloseServerDbs closes all administrative database handles
func CloseServerDbs() []error {
	return closeDbHandles(&dbsLock, &dbs)
}

// closeDbHandles detaches and closes all handles in the map
func closeDbHandles(
	lock *sync.RWMutex,
	dbHandles *map[uint64]*gorm.DB,
) []error {

	// Detach every owned handle atomically so repeated close calls remain harmless
	lock.Lock()
	handles := *dbHandles
	*dbHandles = make(map[uint64]*gorm.DB)
	lock.Unlock()

	// Prepare memory for errors
	errs := make([]error, 0, 1)

	// Close all handles
	for id, dbHandle := range handles {
		if errClose := utils.CloseDatabasePool(dbHandle); errClose != nil {
			errs = append(errs, fmt.Errorf("database handle with ID %d %w", id, errClose))
		}
	}

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
	scopeEntries, errScopeEntries := database.GetScopeEntries()
	if errScopeEntries != nil {
		logger.Warningf("Could not query scan scopes: %s", errScopeEntries)
		return errScopeEntries
	}

	// Iterate over all scan scopes to auto-migrate databases
	for _, scopeEntry := range scopeEntries {

		// Open scan scope database
		scopeDb, errScopeDb := GetScopeDbHandle(logger, &scopeEntry)
		if errScopeDb != nil {
			return fmt.Errorf(
				"could not migrate scan scope '%s' ('%s'): %s", scopeEntry.Name, scopeEntry.DbName, errScopeDb)
		}

		// Log action
		logger.Infof("Migrating scan scope '%s' ('%s').", scopeEntry.Name, scopeEntry.DbName)

		// Auto-migrate scan scope database
		errMigrate := AutoMigrateScopeDb(scopeDb)
		if errMigrate != nil {
			return errMigrate
		}
	}

	// Get scope views
	viewEntries, errViewEntries := database.GetViewEntries()
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
		scopeDb, errScopeDb := GetScopeDbHandle(logger, &viewEntry.ScanScope)
		if errScopeDb != nil {
			return fmt.Errorf(
				"could not migrate scope views of scan scope '%s' ('%s'): %s",
				viewEntry.ScanScope.Name,
				viewEntry.ScanScope.DbName,
				errScopeDb,
			)
		}

		// Execute rebuild of view as transaction
		errTxScopeDb := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {
			return RebuildScopeView(txScopeDb, &viewEntry)
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

	// Kill existing connections. User requesting fresh password, so there would most likely not be any active
	// sessions anymore. Kill them otherwise to prevent potentially hijacked sessions from being kept alive.
	errKill := killUserConnections(serverDb, username)
	if errKill != nil {
		return errKill
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
		&T_nuclei{},
		&T_nuclei_result{},
		&T_smb{},
		&T_smb_file{},
		&T_ssh{},
		&T_ssl{},
		&T_ssl_cipher{},
		&T_ssl_certificate{},
		&T_ssl_setting{},
		&T_ssl_issue{},
		&T_webcrawler{},
		&T_webcrawler_vhost{},
		&T_webcrawler_page{},
		&T_webenum{},
		&T_webenum_results{},
		&T_sql_log{},
	}

	// Start a background reaper clearing idle sessions that block the migration's lock acquisition. Active queries are
	// never touched, so the migration waits for legitimate work to finish, while dead or idle-in-transaction leftovers
	// that would otherwise stall it indefinitely get cleared out of the way once queued behind them.
	stopReaper := make(chan struct{})
	reaperDone := make(chan struct{})
	go func() {

		// Signal completion so the migration can wait for a clean reaper shutdown
		defer close(reaperDone)

		// Reap periodically until the migration signals it is done acquiring and releasing its locks
		ticker := time.NewTicker(dbMigrationReapInterval)
		defer ticker.Stop()
		grace := int(dbMigrationIdleGrace.Seconds())
		for {
			select {
			case <-stopReaper:
				return
			case <-ticker.C:

				// Ignore reaping errors, the next tick retries and the migration keeps waiting for its turn
				_ = reapMigrationBlockers(scopeDb, grace)
			}
		}
	}()

	// Create or update db tables, waiting indefinitely for the required locks while the reaper clears idle blockers
	errMigrate := scopeDb.AutoMigrate(allModels...)

	// Stop the reaper and wait for it to return before evaluating the migration result
	close(stopReaper)
	<-reaperDone

	// Abort if the migration itself failed
	if errMigrate != nil {
		//lint:ignore ST1005 I want it this way in this case for better visibility
		return fmt.Errorf("could not auto migrate: %s\n", errMigrate)
	}

	// Check whether the automatically created "NOT NULL" constraint on t_discovery_hosts is still present. Testing
	// first avoids taking an ACCESS EXCLUSIVE lock on every startup for a change that has already been applied.
	var isNullable string
	errCheck := scopeDb.Raw(`SELECT is_nullable FROM information_schema.columns WHERE table_name = 't_discovery_hosts' AND column_name = 'id_t_discovery';`).Scan(&isNullable).Error
	if errCheck != nil {
		return fmt.Errorf("could not check standard constraint: %s", errCheck)
	}

	// Remove the "NOT NULL" constraint if still set, it is not desired because t_discovery entries referenced by
	// t_discovery_hosts may vanish over time
	if isNullable == "NO" {
		errConstraint := scopeDb.Exec(`ALTER TABLE t_discovery_hosts ALTER COLUMN id_t_discovery DROP NOT NULL;`).Error
		if errConstraint != nil {
			return fmt.Errorf("could not drop standard constraint: %s", errConstraint)
		}
	}

	// Return nil as everything went fine
	return nil
}

// reapMigrationBlockers terminates idle database sessions that are blocking a pending schema migration. Sessions that
// are actively executing a query are never terminated; they are allowed to complete naturally while the migration waits
// its turn behind them. A session stuck "idle in transaction" is granted a grace period to resume before it is
// terminated, whereas an already aborted transaction is terminated immediately as it will never proceed on its own.
func reapMigrationBlockers(scopeDb *gorm.DB, graceSeconds int) error {

	// Terminate idle sessions blocking one of our own connections that is currently waiting for a lock
	query := `
		SELECT pg_terminate_backend(blocker.pid)
		FROM pg_stat_activity blocker
		JOIN pg_stat_activity waiter ON blocker.pid = ANY (pg_blocking_pids(waiter.pid))
		WHERE waiter.wait_event_type = 'Lock'
		  AND waiter.application_name = ?
		  AND blocker.pid <> pg_backend_pid()
		  AND (
		        blocker.state = 'idle in transaction (aborted)'
		     OR (blocker.state = 'idle in transaction' AND now() - blocker.state_change > make_interval(secs => ?))
		  );`

	// Execute reaping statement, targeting only idle blockers of our own lock-waiting connections
	errReap := scopeDb.Exec(query, connectionName, graceSeconds).Error
	if errReap != nil {
		return errReap
	}

	// Return nil as everything went fine
	return nil
}

// SetTargets changes current scan scope targets to the given list (inserting new, deleting vanished and
// updating remaining ones)
func SetTargets(scopeDb *gorm.DB, targetEntries []T_discovery) (created uint64, removed uint64, updated uint64, err error) {

	// Retrieve all input targets currently stored in the scope database
	var targetEntriesLoaded []T_discovery
	errDb := scopeDb.Model(&T_discovery{}).
		Find(&targetEntriesLoaded).Error
	if errDb != nil {
		return 0, 0, 0, fmt.Errorf("could not load existing inputs: %s", errDb)
	}

	// Transform existing entries into searchable data structure
	var targetEntriesExisting = make(map[string]T_discovery, len(targetEntriesLoaded))
	for _, targetEntryExisting := range targetEntriesLoaded {
		targetEntriesExisting[targetEntryExisting.Input] = targetEntryExisting
	}

	// Transform new entries into searchable data structure
	var targetEntriesNew = make(map[string]T_discovery, len(targetEntries))
	for _, targetEntry := range targetEntries {
		targetEntriesNew[targetEntry.Input] = targetEntry
	}

	// Compare new list with existing inputs to calculate new/vanished/updated values
	targetEntriesCreate, targetEntriesRemove, targetEntriesUpdate := mergeInputs(targetEntriesExisting, targetEntriesNew)

	// Start transaction on the scope database. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// TODO test what happens when broker tries to complete (update) a t_discovery service, which might
		// have been deleted by the importer in the meantime

		// Execute the actual insert.
		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb2 := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: MaxBatchSizeDiscovery}).
			Create(&targetEntriesCreate).Error
		if errDb2 != nil {
			// Rollback everything if we can't insert something
			return fmt.Errorf("could not insert input: %s", errDb2)
		}

		// Delete removed targets
		var deleteIds []uint64
		for _, targetEntryRemove := range targetEntriesRemove {
			deleteIds = append(deleteIds, targetEntryRemove.Id)
		}
		if len(deleteIds) > 0 {
			errDb3 := txScopeDb.Delete(&T_discovery{}, deleteIds).Error
			if errDb3 != nil {
				// Rollback everything if we can't insert something
				return fmt.Errorf("could not remove input: %s", errDb3)
			}
		}

		// Update metadata of remaining targets
		for _, targetEntryUpdate := range targetEntriesUpdate {

			// Execute update
			txScopeDb.
				Model(&targetEntryUpdate).
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
					"input_contact",
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
				Updates(&targetEntryUpdate)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return transaction error if transaction failed
	if errTx != nil {
		return 0, 0, 0, fmt.Errorf("could not set scope inputs: %s", errTx)
	}

	// Return nil as everything went fine
	return uint64(len(targetEntriesCreate)), uint64(len(targetEntriesRemove)), uint64(len(targetEntriesUpdate)), nil
}

// GetTargets queries the current inputs from the scan scope
func GetTargets(scopeDb *gorm.DB) ([]T_discovery, error) {

	// Prepare memory for query result
	var targetEntries []T_discovery

	// Query total inputs
	errDb := scopeDb.Model(&T_discovery{}).
		Order("id").
		Find(&targetEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return counts
	return targetEntries, nil
}

// GetProgress queries a given scope database for completed, failed and running scan inputs (progress)
func GetProgress(scopeDb *gorm.DB) (total int64, done int64, running int64, failed int64, errProgress error) {

	// Query total inputs
	errDb := scopeDb.Model(&T_discovery{}).
		Where("enabled IS TRUE").
		Count(&total).Error
	if errDb != nil {
		errProgress = errDb
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
		errProgress = errDb2
		return
	}

	// Query running inputs
	errDb3 := scopeDb.Model(&T_discovery{}).
		Where("scan_started IS NOT NULL").
		Where("scan_finished IS NULL").
		Where("enabled IS TRUE").
		Count(&running).Error
	if errDb3 != nil {
		errProgress = errDb3
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
		errProgress = errDb4
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
		return fmt.Errorf("could not reset failed scan targets in scope database: %s", errScopeDb)
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

// DeleteDatabaseCredentials removes a given user from database server. If a user is still referenced by some table,
// it can't be deleted. In case of such error, the error will be ignored and the user remains.
func DeleteDatabaseCredentials(txServerDb *gorm.DB, username string) {

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
// reconnection is prevented before calling; otherwise, connections might be re-established automatically.
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

// CreateScopeDb creates a new scan database in the scope database. This action cannot run within a transaction!
func CreateScopeDb(serverDb *gorm.DB, name string, comment string) error {

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

// DeleteScopeDb removes a scan scope database from a database server. This action cannot run within a transaction!
// ATTENTION: This action cannot be rolled back!
func DeleteScopeDb(serverDb *gorm.DB, name string) error {

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

// CreateScopeView creates a set of view tables belonging new scope view (there are multiple view tables per view:
// hosts, services, smb, nfs,...).
func CreateScopeView(txScopeDb *gorm.DB, name string, filters map[string][]string) ([]string, error) {

	// Prepare storage for generated view names
	var viewNames []string

	// Prepare gigantic where clause based on defined filters. Values and column names are escaped as Postgres
	// literals and identifiers to prevent SQL injection through direct RPC callers
	subClauses := make([]string, 0, len(filters))
	for column, values := range filters {
		values = scanUtils.TrimToLower(values)
		if len(values) > 0 {

			// Escape filter values as Postgres literals and column name as identifier
			escapedValues := make([]string, 0, len(values))
			for _, value := range values {
				escaped, _ := escape.Escape("%L", value)
				escapedValues = append(escapedValues, escaped)
			}
			columnEscaped, _ := escape.Escape("%I", column)
			subClauses = append(subClauses, fmt.Sprintf("LOWER(%s) IN (%s)", columnEscaped, strings.Join(escapedValues, ", ")))
		}
	}

	// Combine them as a OR clause
	whereClause := strings.Join(subClauses, " OR ")

	// Create views
	for view, definition := range viewDefinitions {

		// Define view name
		viewName, errViewName := SanitizeViewName(name + "_" + view)
		if errViewName != nil {
			return nil, fmt.Errorf("could not sanitize view name: %s", errViewName)
		}

		// Drop view if it was already existing to rebuild it
		errDelete := DeleteScopeView(txScopeDb, viewName)
		if errDelete != nil {
			return nil, fmt.Errorf("could not remove original view: %s", errDelete)
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

	// Sort views alphabetically
	slices.Sort(viewNames)

	// Return nil as everything went fine
	return viewNames, nil
}

// DeleteScopeView removes a set of view tables belonging to a scope view (there are multiple view  tables per view:
// hosts, services, smb, nfs,...).
func DeleteScopeView(txScopeDb *gorm.DB, viewTableNames ...string) error {

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

// RebuildScopeView rebuilds a scan scope views in the scope database based on the last definition and re-grants
// authorized users. This might be necessary, if the view definition has changed, e.g. if columns were
// introduced, removed or changed in their order. It can also be used to pull access rights straight to match
// the actual configuration.
func RebuildScopeView(txScopeDb *gorm.DB, viewEntry *database.T_scope_view) error {

	// Return if there are no views tables to be updated
	if viewEntry.ViewNames == "" {
		return nil
	}

	// Generate list of view table names from manager database
	viewTableNames := utils.SanitizeToSlice(viewEntry.ViewNames, ",")

	// Drop previously known view tables. They might now have other suffixes, if changed in the program code.
	errDelete := DeleteScopeView(txScopeDb, viewTableNames...)
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

	// Create view. Existing ones with the same name will be removed and re-created in the process.
	// Might happen, if new scope views were introduced but not added to the view's 'view_names' column.
	updatedViewTableNames, errUpdatedViewTableNames := CreateScopeView(txScopeDb, viewEntry.Name, filters)
	if errUpdatedViewTableNames != nil {
		return fmt.Errorf("could not rebuild view: %s", errUpdatedViewTableNames)
	}

	// Store updated list of view table names
	viewEntry.ViewNames = strings.Join(updatedViewTableNames, ",")
	_, errSave := viewEntry.Save("view_names")
	if errSave != nil {
		return fmt.Errorf("could not rebuild view: %s", errSave)
	}

	// Iterate granted users
	for _, grant := range viewEntry.Grants {
		errGrant := GrantScopeView(
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

// GrantScopeView grants access for a credentials set to a set of view tables belonging to a scope view (there
// are multiple view tables per view: hosts, services, smb, nfs,...).
// ATTENTION: A credentials set may either be user specific or a random access token
// ATTENTION: If the credentials set is not existing on the given database server, it will be created. That's why
//
//	the hashed user's password needs to be passed in too.
func GrantScopeView(
	txScopeDb *gorm.DB,
	viewEntry *database.T_scope_view,
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

	// Get list of view table names from manager database
	viewTableNames := utils.SanitizeToSlice(viewEntry.ViewNames, ",")

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

// RevokeScopeView revokes user access from a set of view tables belonging to a scope view (there are multiple view
// tables per view: hosts, services, smb, nfs,...).
// ATTENTION: If cleanUser is true, the given user and remaining rights are completely removed from the database server.
func RevokeScopeView(txScopeDb *gorm.DB, dbName, username string, viewTableNames []string) error {

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

	// Count if the user still has any select rights left on the scope database
	var rightsCount int64
	errDb3 := txScopeDb.
		Table("information_schema.role_table_grants").
		Where("grantee = ?", username).
		Where("privilege_type = ?", "SELECT").
		Count(&rightsCount).Error
	if errDb3 != nil {
		return errDb3
	}

	// Clean up remaining rights from the related scope database
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

// SanitizeViewName strips illegal view name characters and transforms the name into unified format
func SanitizeViewName(name string) (string, error) {

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
func mergeInputs(targetEntriesExisting map[string]T_discovery, targetEntriesNew map[string]T_discovery) (
	targetEntriesCreate []T_discovery, targetEntriesRemove []T_discovery, targetEntriesUpdate []T_discovery) {

	// Find added values that need to be inserted
	for k, targetEntry := range targetEntriesNew {
		_, exists := targetEntriesExisting[k]
		if !exists {
			targetEntriesCreate = append(targetEntriesCreate, targetEntry)
		}
	}

	// Find removed values that need to be deleted
	for k, targetEntry := range targetEntriesExisting {
		_, exists := targetEntriesNew[k]
		if !exists {
			targetEntriesRemove = append(targetEntriesRemove, targetEntry)
		}
	}

	// Find remaining entries that need to be updated
	for k, targetEntry := range targetEntriesNew {
		targetEntryExisting, exists := targetEntriesExisting[k]
		if exists {

			// Update database ID in current entry from existing entry
			targetEntry.Id = targetEntryExisting.Id

			// Add entry to list of update entries
			targetEntriesUpdate = append(targetEntriesUpdate, targetEntry)
		}
	}

	// Return sorted lists
	return targetEntriesCreate, targetEntriesRemove, targetEntriesUpdate
}
