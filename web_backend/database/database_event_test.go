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
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// TestGetEvents_Limit_ReturnsNewestEntries verifies a filtered query limits rows before returning them
func TestGetEvents_Limit_ReturnsNewestEntries(t *testing.T) {

	// Prepare unit test data
	test_prepareEventEntries(t, EventApiToken)
	limit := 1

	// Query only the newest matching event
	events, errEvents := GetEvents(EventApiToken, time.Time{}, &limit)
	if errEvents != nil {
		t.Fatalf("GetEvents() error = '%v', want = 'nil'", errEvents)
	}

	// Verify the limited result remains chronologically ordered
	if len(events) != 1 {
		t.Fatalf("GetEvents() result count = '%v', want = '1'", len(events))
	}
	if events[0].EventDetail != "newest" {
		t.Errorf("GetEvents() EventDetail = '%v', want = 'newest'", events[0].EventDetail)
	}
}

// TestGetEventsAll_Limit_ReturnsNewestEntries verifies an unfiltered query limits rows before returning them
func TestGetEventsAll_Limit_ReturnsNewestEntries(t *testing.T) {

	// Prepare unit test data
	test_prepareEventEntries(t, EventApiToken)
	limit := 1

	// Query only the newest event
	events, errEvents := GetEventsAll(time.Time{}, &limit)
	if errEvents != nil {
		t.Fatalf("GetEventsAll() error = '%v', want = 'nil'", errEvents)
	}

	// Verify the limited result remains chronologically ordered
	if len(events) != 1 {
		t.Fatalf("GetEventsAll() result count = '%v', want = '1'", len(events))
	}
	if events[0].EventDetail != "newest" {
		t.Errorf("GetEventsAll() EventDetail = '%v', want = 'newest'", events[0].EventDetail)
	}
}

// test_prepareEventEntries initializes an isolated database with two chronologically ordered events
func test_prepareEventEntries(t *testing.T, event Event) {
	t.Helper()

	// Open an isolated in-memory database
	db, errDb := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	})
	if errDb != nil {
		t.Fatalf("gorm.Open() error = '%v', want = 'nil'", errDb)
	}
	sqlDb, errSqlDb := db.DB()
	if errSqlDb != nil {
		t.Fatalf("DB() error = '%v', want = 'nil'", errSqlDb)
	}
	sqlDb.SetMaxOpenConns(1)

	// Replace the package database for this test and restore it afterward
	previousDb := backendDb
	backendDb = db
	t.Cleanup(func() {
		_ = sqlDb.Close()
		backendDb = previousDb
	})

	// Create the event schema
	if errMigrate := backendDb.AutoMigrate(&T_user{}, &T_event{}); errMigrate != nil {
		t.Fatalf("AutoMigrate() error = '%v', want = 'nil'", errMigrate)
	}

	// Persist two deterministically ordered events
	now := time.Now()
	eventEntries := []T_event{
		{
			Email:       "user@domain.tld",
			Timestamp:   now.Add(-time.Hour),
			Event:       event,
			EventDetail: "oldest",
		},
		{
			Email:       "user@domain.tld",
			Timestamp:   now,
			Event:       event,
			EventDetail: "newest",
		},
	}
	if errCreate := backendDb.Create(&eventEntries).Error; errCreate != nil {
		t.Fatalf("Create() error = '%v', want = 'nil'", errCreate)
	}
}
