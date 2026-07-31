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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_test"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// TestMain initializes the SQLite metadata database required by scopedb integration tests
func TestMain(m *testing.M) {

	// Isolate every file created by this package below the shared _test directory
	_test.GetSettings()
	tmpDir, errTmpDir := os.MkdirTemp(".", "lsd-manager-scopedb-test-*")
	if errTmpDir != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not create temp dir: %v\n", errTmpDir)
		os.Exit(1)
	}
	if errChdir := os.Chdir(tmpDir); errChdir != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not enter temp dir: %v\n", errChdir)
		os.Exit(1)
	}

	// Open the real manager metadata database used by scopedb lookups
	if errOpen := database.Open(); errOpen != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not open manager database: %v\n", errOpen)
		os.Exit(1)
	}

	// Run the package tests before releasing database and filesystem resources
	code := m.Run()
	_ = database.Close()
	_ = os.Chdir("..")
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

// test_openSqliteDb opens a real in-process database for cache lifecycle tests
func test_openSqliteDb(t *testing.T) (*gorm.DB, *sql.DB) {

	// Mark setup failures as originating from the calling test
	t.Helper()

	// Open an independent healthy database handle
	dbHandle, errDbHandle := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if errDbHandle != nil {
		t.Fatalf("test_openSqliteDb() error = '%v', want = 'nil'", errDbHandle)
	}
	sqlDb, errSqlDb := dbHandle.DB()
	if errSqlDb != nil {
		t.Fatalf("test_openSqliteDb() DB() error = '%v', want = 'nil'", errSqlDb)
	}

	// Keep cleanup safe when the cache lifecycle already closed the pool
	t.Cleanup(func() {
		_ = sqlDb.Close()
	})

	// Return both layers so tests can verify the pool lifecycle
	return dbHandle, sqlDb
}

// test_openRecoverableStaleDb opens a pool whose failed ping can recover while it remains open
func test_openRecoverableStaleDb(t *testing.T) (*gorm.DB, *sql.DB, string) {

	// Mark setup failures as originating from the calling test
	t.Helper()

	// Require an existing SQLite file so the initial ping deterministically fails
	dbPath := filepath.Join(t.TempDir(), "stale.sqlite")
	dsn := "file:" + filepath.ToSlash(dbPath) + "?mode=rw"
	sqlDb, errSqlDb := sql.Open(sqlite.DriverName, dsn)
	if errSqlDb != nil {
		t.Fatalf("test_openRecoverableStaleDb() error = '%v', want = 'nil'", errSqlDb)
	}
	t.Cleanup(func() {
		_ = sqlDb.Close()
	})
	if errPing := sqlDb.Ping(); errPing == nil {
		t.Fatal("test_openRecoverableStaleDb() Ping() error = 'nil', want = 'missing file error'")
	}

	// Wrap the real SQL pool in the same GORM handle shape used by the cache
	dbHandle := &gorm.DB{Config: &gorm.Config{ConnPool: sqlDb}}
	return dbHandle, sqlDb, dbPath
}

// test_cacheDbHandle installs an isolated cache entry and removes every handle left under its ID
func test_cacheDbHandle(
	t *testing.T,
	lock *sync.RWMutex,
	activeHandles *map[uint64]*gorm.DB,
	id uint64,
	dbHandle *gorm.DB,
) {

	// Mark setup failures as originating from the calling test
	t.Helper()

	// Reserve a unique cache ID without disturbing state owned by another test
	lock.Lock()
	if _, activeExists := (*activeHandles)[id]; activeExists {
		lock.Unlock()
		t.Fatalf("test_cacheDbHandle() ID = '%d', want = 'unused'", id)
	}
	(*activeHandles)[id] = dbHandle
	lock.Unlock()

	// Detach and close only the handle owned by this test
	t.Cleanup(func() {
		if !lock.TryLock() {
			t.Errorf("test_cacheDbHandle() could not reacquire the cache lock during cleanup")
			return
		}
		activeHandle := (*activeHandles)[id]
		delete(*activeHandles, id)
		lock.Unlock()

		_ = utils.CloseDatabasePool(activeHandle)
	})
}

// test_assertRwLockReleased verifies a handle lookup did not retain its cache read lock
func test_assertRwLockReleased(t *testing.T, lock *sync.RWMutex) {

	// Acquire the write lock without blocking so a leaked read lock is observable
	t.Helper()
	if lock.TryLock() {
		lock.Unlock()
		return
	}

	t.Fatal("cache lock state = 'read-locked', want = 'unlocked'")
}

// makeDiscovery returns a T_discovery with the given input value set.
func makeDiscovery(input string) T_discovery {
	targetEntry := T_discovery{}
	targetEntry.Input = input
	return targetEntry
}

// TestSanitizeViewName verifies view names are lower-cased and non-alphanumeric characters become underscores.
func TestSanitizeViewName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "plain-upper",
			input:   "MyView",
			want:    "myview",
			wantErr: false,
		},
		{
			name:    "space-to-underscore",
			input:   "My View",
			want:    "my_view",
			wantErr: false,
		},
		{
			name:    "hyphen-to-underscore",
			input:   "My-View",
			want:    "my_view",
			wantErr: false,
		},
		{
			name:    "digits-preserved",
			input:   "my-view-2024",
			want:    "my_view_2024",
			wantErr: false,
		},
		{
			name:    "specials-stripped",
			input:   "a!@#$b",
			want:    "ab",
			wantErr: false,
		},
		{
			name:    "underscores-kept",
			input:   "__keep_underscores__",
			want:    "__keep_underscores__",
			wantErr: false,
		},
		{
			name:    "all-upper",
			input:   "UPPER CASE",
			want:    "upper_case",
			wantErr: false,
		},
		{
			name:    "mixed",
			input:   "mixed Case-AND-Spaces",
			want:    "mixed_case_and_spaces",
			wantErr: false,
		},
		{
			name:    "only-spaces",
			input:   "   ",
			want:    "___",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGot := SanitizeViewName(tt.input)
			if (errGot != nil) != tt.wantErr {
				t.Errorf("sanitizeViewName(%q) error = '%v', wantErr = '%v'", tt.input, errGot, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sanitizeViewName(%q) = '%v', want = '%v'", tt.input, got, tt.want)
			}
		})
	}
}

// TestMergeInputs verifies mergeInputs correctly classifies entries as create, remove, or update.
func TestMergeInputs(t *testing.T) {
	tests := []struct {
		name       string
		existing   map[string]T_discovery
		newInputs  map[string]T_discovery
		wantCreate int
		wantRemove int
		wantUpdate int
	}{
		{
			name:       "all-new",
			existing:   map[string]T_discovery{},
			newInputs:  map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1"), "192.0.2.2": makeDiscovery("192.0.2.2")},
			wantCreate: 2,
			wantRemove: 0,
			wantUpdate: 0,
		},
		{
			name:       "all-removed",
			existing:   map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1"), "192.0.2.2": makeDiscovery("192.0.2.2")},
			newInputs:  map[string]T_discovery{},
			wantCreate: 0,
			wantRemove: 2,
			wantUpdate: 0,
		},
		{
			name:       "all-updated",
			existing:   map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1")},
			newInputs:  map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1")},
			wantCreate: 0,
			wantRemove: 0,
			wantUpdate: 1,
		},
		{
			name:       "mixed",
			existing:   map[string]T_discovery{"keep": makeDiscovery("keep"), "remove": makeDiscovery("remove")},
			newInputs:  map[string]T_discovery{"keep": makeDiscovery("keep"), "add": makeDiscovery("add")},
			wantCreate: 1,
			wantRemove: 1,
			wantUpdate: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetEntriesCreate, targetEntriesRemove, targetEntriesUpdate := mergeInputs(tt.existing, tt.newInputs)
			if len(targetEntriesCreate) != tt.wantCreate {
				t.Errorf("mergeInputs() create = '%v', want = '%v'", len(targetEntriesCreate), tt.wantCreate)
			}
			if len(targetEntriesRemove) != tt.wantRemove {
				t.Errorf("mergeInputs() remove = '%v', want = '%v'", len(targetEntriesRemove), tt.wantRemove)
			}
			if len(targetEntriesUpdate) != tt.wantUpdate {
				t.Errorf("mergeInputs() update = '%v', want = '%v'", len(targetEntriesUpdate), tt.wantUpdate)
			}
		})
	}
}

// TestSetConnectionName verifies SetConnectionName sanitizes spaces to underscores and stores the result.
func TestSetConnectionName(t *testing.T) {
	original := connectionName
	t.Cleanup(func() { connectionName = original })

	SetConnectionName("test client")
	if connectionName != "test_client" {
		t.Errorf("SetConnectionName() connectionName = '%v', want = 'test_client'", connectionName)
	}
}

// TestSetMaxConnectionsDefault verifies SetMaxConnectionsDefault updates the global connection limit.
func TestSetMaxConnectionsDefault(t *testing.T) {
	original := dbMaxConn
	t.Cleanup(func() { dbMaxConn = original })

	SetMaxConnectionsDefault(50)
	if dbMaxConn != 50 {
		t.Errorf("SetMaxConnectionsDefault() dbMaxConn = '%v', want = '50'", dbMaxConn)
	}
}

// TestErrInvalidCharacter_Error verifies ErrInvalidCharacter produces a non-empty error message.
func TestErrInvalidCharacter_Error(t *testing.T) {
	errInvalidCharacter := ErrInvalidCharacter{ParamName: "username", Value: "bad!value"}
	if errInvalidCharacter.Error() == "" {
		t.Errorf("ErrInvalidCharacter.Error() = '', want = 'non-empty'")
	}
}

// TestCloseScopeDbs_NoConnections verifies CloseScopeDbs returns nil when no scope database handles are open.
func TestCloseScopeDbs_NoConnections(t *testing.T) {
	errs := CloseScopeDbs()
	if errs != nil {
		t.Errorf("CloseScopeDbs() = '%v', want = 'nil'", errs)
	}
}

// TestGetServerDbHandle_HealthyCachedHandle_ReusesHandle verifies a healthy server handle avoids replacement
func TestGetServerDbHandle_HealthyCachedHandle_ReusesHandle(t *testing.T) {

	// Prepare a healthy cached server handle under an isolated ID
	const dbServerId uint64 = 9001001
	cachedDb, _ := test_openSqliteDb(t)
	test_cacheDbHandle(t, &dbsLock, &dbs, dbServerId, cachedDb)
	testLogger := scanUtils.NewTestLogger()

	// Retrieve the server handle through the production cache path
	dbHandle, errDbHandle := GetServerDbHandle(testLogger, &database.T_db_server{Id: dbServerId})
	if errDbHandle != nil {
		t.Fatalf("GetServerDbHandle() error = '%v', want = 'nil'", errDbHandle)
	}

	// Verify no replacement was opened or cached
	if dbHandle != cachedDb {
		t.Errorf("GetServerDbHandle() handle = '%p', want = '%p'", dbHandle, cachedDb)
	}
}

// TestGetScopeDbHandle_HealthyCachedHandle_ReusesHandle verifies a healthy scope handle avoids replacement
func TestGetScopeDbHandle_HealthyCachedHandle_ReusesHandle(t *testing.T) {

	// Prepare a healthy cached scope handle under an isolated ID
	const scopeId uint64 = 9001002
	cachedDb, _ := test_openSqliteDb(t)
	test_cacheDbHandle(t, &scopeDbsLock, &scopeDbs, scopeId, cachedDb)
	testLogger := scanUtils.NewTestLogger()
	scopeEntry := &database.T_scan_scope{
		Id:       scopeId,
		DbServer: database.T_db_server{Id: 1},
	}

	// Retrieve the scope handle through the production cache path
	dbHandle, errDbHandle := GetScopeDbHandle(testLogger, scopeEntry)
	if errDbHandle != nil {
		t.Fatalf("GetScopeDbHandle() error = '%v', want = 'nil'", errDbHandle)
	}

	// Verify no replacement was opened or cached
	if dbHandle != cachedDb {
		t.Errorf("GetScopeDbHandle() handle = '%p', want = '%p'", dbHandle, cachedDb)
	}
}

// TestGetServerDbHandle_CachedDbError_ReleasesCacheLock verifies DB errors do not leak the server cache lock
func TestGetServerDbHandle_CachedDbError_ReleasesCacheLock(t *testing.T) {

	// Prepare a cached handle whose DB method returns gorm.ErrInvalidDB
	const dbServerId uint64 = 9001003
	invalidDb := &gorm.DB{Config: &gorm.Config{}}
	test_cacheDbHandle(t, &dbsLock, &dbs, dbServerId, invalidDb)
	testLogger := scanUtils.NewTestLogger()

	// Retrieve the invalid handle through the production cache path
	_, errHandle := GetServerDbHandle(testLogger, &database.T_db_server{Id: dbServerId})
	if !errors.Is(errHandle, gorm.ErrInvalidDB) {
		t.Errorf("GetServerDbHandle() error = '%v', want = '%v'", errHandle, gorm.ErrInvalidDB)
	}

	// Verify the early error return released its read lock
	test_assertRwLockReleased(t, &dbsLock)
}

// TestGetScopeDbHandle_CachedDbError_ReleasesCacheLock verifies DB errors do not leak the scope cache lock
func TestGetScopeDbHandle_CachedDbError_ReleasesCacheLock(t *testing.T) {

	// Prepare a cached handle whose DB method returns gorm.ErrInvalidDB
	const scopeId uint64 = 9001004
	invalidDb := &gorm.DB{Config: &gorm.Config{}}
	test_cacheDbHandle(t, &scopeDbsLock, &scopeDbs, scopeId, invalidDb)
	testLogger := scanUtils.NewTestLogger()
	scopeEntry := &database.T_scan_scope{
		Id:       scopeId,
		DbServer: database.T_db_server{Id: 1},
	}

	// Retrieve the invalid handle through the production cache path
	_, errHandle := GetScopeDbHandle(testLogger, scopeEntry)
	if !errors.Is(errHandle, gorm.ErrInvalidDB) {
		t.Errorf("GetScopeDbHandle() error = '%v', want = '%v'", errHandle, gorm.ErrInvalidDB)
	}

	// Verify the early error return released its read lock
	test_assertRwLockReleased(t, &scopeDbsLock)
}

// TestGetServerDbHandle_ReplacementFailure_PreservesStaleHandle verifies failed replacement retains cache ownership
func TestGetServerDbHandle_ReplacementFailure_PreservesStaleHandle(t *testing.T) {

	// Prepare an open cached pool whose ping fails until its backing file exists
	const dbServerId uint64 = 9001005
	staleDb, staleSqlDb, dbPath := test_openRecoverableStaleDb(t)
	test_cacheDbHandle(t, &dbsLock, &dbs, dbServerId, staleDb)
	testLogger := scanUtils.NewTestLogger()
	dbServer := &database.T_db_server{
		Id:      dbServerId,
		Name:    "test-server",
		Dialect: "unsupported",
	}

	// Trigger a deterministic replacement failure after both cached ping checks fail
	_, errHandle := GetServerDbHandle(testLogger, dbServer)
	if errHandle == nil {
		t.Fatal("GetServerDbHandle() error = 'nil', want = 'unsupported dialect error'")
	}

	// Verify the stale handle remains active after the failed replacement
	dbsLock.RLock()
	activeDb := dbs[dbServerId]
	dbsLock.RUnlock()
	if activeDb != staleDb {
		t.Errorf("GetServerDbHandle() cache state = '%p', want = '%p'", activeDb, staleDb)
	}

	// Make the stale pool healthy and prove the failed replacement did not close it
	if errCreate := os.WriteFile(dbPath, nil, 0o600); errCreate != nil {
		t.Fatalf("os.WriteFile() error = '%v', want = 'nil'", errCreate)
	}
	if errPing := staleSqlDb.Ping(); errPing != nil {
		t.Errorf("GetServerDbHandle() cached pool Ping() error = '%v', want = 'nil'", errPing)
	}
}

// TestGetScopeDbHandle_ReplacementFailure_PreservesStaleHandle verifies failed replacement retains cache ownership
func TestGetScopeDbHandle_ReplacementFailure_PreservesStaleHandle(t *testing.T) {

	// Prepare an open cached pool whose ping fails until its backing file exists
	const scopeId uint64 = 9001006
	staleDb, staleSqlDb, dbPath := test_openRecoverableStaleDb(t)
	test_cacheDbHandle(t, &scopeDbsLock, &scopeDbs, scopeId, staleDb)
	testLogger := scanUtils.NewTestLogger()
	scopeEntry := &database.T_scan_scope{
		Id:     scopeId,
		Name:   "test-scope",
		DbName: "test_scope",
		DbServer: database.T_db_server{
			Id:      1,
			Dialect: "unsupported",
		},
	}

	// Trigger a deterministic replacement failure after both cached ping checks fail
	_, errHandle := GetScopeDbHandle(testLogger, scopeEntry)
	if errHandle == nil {
		t.Fatal("GetScopeDbHandle() error = 'nil', want = 'unsupported dialect error'")
	}

	// Verify the stale handle remains active after the failed replacement
	scopeDbsLock.RLock()
	activeDb := scopeDbs[scopeId]
	scopeDbsLock.RUnlock()
	if activeDb != staleDb {
		t.Errorf("GetScopeDbHandle() cache state = '%p', want = '%p'", activeDb, staleDb)
	}

	// Make the stale pool healthy and prove the failed replacement did not close it
	if errCreate := os.WriteFile(dbPath, nil, 0o600); errCreate != nil {
		t.Fatalf("os.WriteFile() error = '%v', want = 'nil'", errCreate)
	}
	if errPing := staleSqlDb.Ping(); errPing != nil {
		t.Errorf("GetScopeDbHandle() cached pool Ping() error = '%v', want = 'nil'", errPing)
	}
}

// TestGetServerDbHandle_Nil verifies GetServerDbHandle returns an error when called with a nil server.
func TestGetServerDbHandle_Nil(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetServerDbHandle(testLogger, nil)
	if errHandle == nil {
		t.Errorf("GetServerDbHandle(nil) expected error, got = 'nil'")
	}
}

// TestGetServerDbHandle_UnsupportedDialect verifies GetServerDbHandle returns an error for unknown dialects.
func TestGetServerDbHandle_UnsupportedDialect(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetServerDbHandle(testLogger, &database.T_db_server{Dialect: "mysql"})
	if errHandle == nil {
		t.Errorf("GetServerDbHandle(mysql) expected error, got = 'nil'")
	}
}

// TestGetScopeDbHandle_Nil verifies GetScopeDbHandle returns an error when called with a nil scope.
func TestGetScopeDbHandle_Nil(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetScopeDbHandle(testLogger, nil)
	if errHandle == nil {
		t.Errorf("GetScopeDbHandle(nil) expected error, got = 'nil'")
	}
}

// TestGetScopeDbHandle_NoDbServer verifies GetScopeDbHandle returns an error when the scope has no DB server ID.
func TestGetScopeDbHandle_NoDbServer(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetScopeDbHandle(testLogger, &database.T_scan_scope{})
	if errHandle == nil {
		t.Errorf("GetScopeDbHandle(empty-scope) expected error, got = 'nil'")
	}
}

// pgTestServer returns a database server from _test/settings.go, skipping if PgHost is not configured.
func pgTestServer(t *testing.T) *database.T_db_server {
	t.Helper()

	settings := _test.GetSettings()
	if settings.PgHost == "" {
		t.Skip("Integration test skipped: PgHost not configured in _test/settings.go")
		return nil
	}

	return &database.T_db_server{
		Name:     "test-pg-server",
		Dialect:  "postgres",
		Host:     settings.PgHost,
		Port:     settings.PgPort,
		Admin:    settings.PgUser,
		Password: settings.PgPassword,
	}
}

// TestPg_GetServerDbHandle verifies GetServerDbHandle successfully connects to a real PostgreSQL server.
func TestPg_GetServerDbHandle(t *testing.T) {
	srv := pgTestServer(t)
	testLogger := scanUtils.NewTestLogger()

	serverDb, errServerDb := GetServerDbHandle(testLogger, srv)
	if errServerDb != nil {
		t.Fatalf("GetServerDbHandle() error = '%v'", errServerDb)
	}
	if serverDb == nil {
		t.Fatal("GetServerDbHandle() returned nil")
	}

	sqlDb, errSqlDb := serverDb.DB()
	if errSqlDb != nil {
		t.Fatalf("GetServerDbHandle() sql DB error = '%v'", errSqlDb)
	}
	if errPing := sqlDb.Ping(); errPing != nil {
		t.Fatalf("GetServerDbHandle() ping error = '%v'", errPing)
	}
}

// TestPg_InstallTrigramIndices verifies InstallTrigramIndices succeeds when applied to a connected PG server DB.
func TestPg_InstallTrigramIndices(t *testing.T) {
	srv := pgTestServer(t)
	testLogger := scanUtils.NewTestLogger()

	serverDb, errServerDb := GetServerDbHandle(testLogger, srv)
	if errServerDb != nil {
		t.Fatalf("GetServerDbHandle() error = '%v'", errServerDb)
	}

	if errInstall := InstallTrigramIndices(serverDb); errInstall != nil {
		t.Errorf("InstallTrigramIndices() error = '%v'", errInstall)
	}
}

// TestAutomigrateScanScopes_NoScopes verifies AutomigrateScanScopes returns nil when no scan scopes are registered.
func TestAutomigrateScanScopes_NoScopes(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	if errMigrate := AutomigrateScanScopes(testLogger); errMigrate != nil {
		t.Errorf("AutomigrateScanScopes() with empty scope list error = '%v'", errMigrate)
	}
}
