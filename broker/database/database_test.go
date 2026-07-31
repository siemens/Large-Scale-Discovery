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
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// TestCloseClosesPoolAfterMaintenanceError verifies maintenance failures cannot bypass mandatory cleanup.
func TestCloseClosesPoolAfterMaintenanceError(t *testing.T) {
	savedDb := brokerDb
	defer func() { brokerDb = savedDb }()

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
	brokerDb = database
	t.Cleanup(func() { _ = utils.CloseDatabasePool(database) })

	if errQueryOnly := database.Exec("PRAGMA query_only = ON").Error; errQueryOnly != nil {
		t.Fatalf("Could not make test database read only: %v", errQueryOnly)
	}
	errClose := Close()
	if errClose == nil || !strings.Contains(errClose.Error(), "could not vacuum broker database") {
		t.Fatalf("Close() error = %v, want broker vacuum error", errClose)
	}
	if errPing := sqlDb.Ping(); errPing == nil {
		t.Fatal("Close() left the underlying database pool open")
	}
	if brokerDb != nil {
		t.Fatal("Close() left the global database handle attached")
	}
	if errCloseAgain := Close(); errCloseAgain != nil {
		t.Fatalf("Second Close() error = %v, want nil", errCloseAgain)
	}
}
