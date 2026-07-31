/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// TestOpenDoesNotPublishDatabaseOnFailure verifies that callers cannot observe an incomplete database handle
func TestOpenDoesNotPublishDatabaseOnFailure(t *testing.T) {

	// Make the configured database path invalid without changing production configuration
	tempDir := t.TempDir()
	errMkdir := os.Mkdir(filepath.Join(tempDir, "backend.sqlite"), 0750)
	if errMkdir != nil {
		t.Fatalf("Could not create invalid database path: %v", errMkdir)
	}
	t.Chdir(tempDir)
	backendDb = nil

	// Attempt initialization against the invalid path
	errOpen := Open()
	if errOpen == nil {
		t.Fatal("Open() returned nil for an invalid database path")
	}
	if backendDb != nil {
		t.Fatal("Open() published a database handle after initialization failed")
	}
}

// TestCloseClosesPoolAfterVacuumError verifies that maintenance failures cannot bypass mandatory cleanup
func TestCloseClosesPoolAfterVacuumError(t *testing.T) {

	// Open an isolated database and retain its underlying pool for the post-close assertion
	databasePath := filepath.Join(t.TempDir(), "backend.sqlite")
	database, errDatabase := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	})
	if errDatabase != nil {
		t.Fatalf("Could not open test database: %v", errDatabase)
	}
	sqlDb, errSqlDb := database.DB()
	if errSqlDb != nil {
		t.Fatalf("Could not retrieve test database pool: %v", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(1)
	backendDb = database
	t.Cleanup(func() {
		_ = Close()
	})

	// Force VACUUM to fail while leaving the pool open and closable
	errQueryOnly := database.Exec("PRAGMA query_only = ON").Error
	if errQueryOnly != nil {
		t.Fatalf("Could not make test database read only: %v", errQueryOnly)
	}

	// Close must report the maintenance error after still closing and detaching the pool
	errClose := Close()
	if errClose == nil {
		t.Fatal("Close() returned nil after VACUUM failed")
	}
	if !strings.Contains(errClose.Error(), "could not vacuum backend database") {
		t.Fatalf("Close() returned an unexpected error: %v", errClose)
	}
	if errPing := sqlDb.Ping(); errPing == nil {
		t.Fatal("Close() left the underlying database pool open")
	}
	if backendDb != nil {
		t.Fatal("Close() left the global database handle attached")
	}

	// Repeated shutdown is harmless after the first call detached the handle
	errCloseAgain := Close()
	if errCloseAgain != nil {
		t.Fatalf("Second Close() returned an error: %v", errCloseAgain)
	}
}

// TestCloseAggregatesMaintenanceErrors verifies that one maintenance failure cannot hide another
func TestCloseAggregatesMaintenanceErrors(t *testing.T) {

	// Open and then invalidate an isolated pool so both maintenance statements fail
	database, errDatabase := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	})
	if errDatabase != nil {
		t.Fatalf("Could not open test database: %v", errDatabase)
	}
	sqlDb, errSqlDb := database.DB()
	if errSqlDb != nil {
		t.Fatalf("Could not retrieve test database pool: %v", errSqlDb)
	}
	errPoolClose := sqlDb.Close()
	if errPoolClose != nil {
		t.Fatalf("Could not close test database pool: %v", errPoolClose)
	}
	backendDb = database
	t.Cleanup(func() {
		_ = Close()
	})

	// Close must preserve the errors from both attempted maintenance operations
	errClose := Close()
	if errClose == nil {
		t.Fatal("Close() returned nil for an invalid database pool")
	}
	for _, expectedError := range []string{
		"could not optimize backend database",
		"could not vacuum backend database",
	} {
		if !strings.Contains(errClose.Error(), expectedError) {
			t.Fatalf("Close() error does not contain %q: %v", expectedError, errClose)
		}
	}
	if backendDb != nil {
		t.Fatal("Close() left the global database handle attached")
	}
}
