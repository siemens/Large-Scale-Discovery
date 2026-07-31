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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// databaseShutdownTimeout limits how long database shutdown may take
const databaseShutdownTimeout = 10 * time.Second

// backendDb holds the backend database handle
var backendDb *gorm.DB

type Result = map[string]interface{}             // One dynamically shaped database result
type ResultSeries = []Result                     // Ordered set of database results
type ResultSeriesStats = map[string]ResultSeries // Result series grouped by a handler-defined statistic

// Open opens the backend database from disk
func Open() error {

	// Prepare sqlite connection
	// Gorm might open additional connections, so it's important that
	// attributes are set in the DSN, rather than a subsequent PRAGMA query.
	// Attributes definition according to https://github.com/glebarez/go-sqlite.
	dsn := "file:backend.sqlite?" +
		"_pragma=journal_mode(WAL)&" + // WAL journal mode for better concurrency
		"_pragma=busy_timeout(60000)&" + // busy timeout to avoid instant database locked errors
		"_pragma=foreign_keys(1)" // with enforced foreign key constraints, otherwise the SQLITE database might ignore them

	// Open sqlite database
	db, errDb := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errDb != nil {
		return fmt.Errorf("could not open backend database: %w", errDb)
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
		return fmt.Errorf("could not apply PRAGMA settings to backend database: %w", errPragmas)
	}

	// Assign db to global variable
	backendDb = db

	// Return nil as everything went fine
	return nil
}

// Close closes the backend database and attempts all maintenance and cleanup steps
func Close() error {

	// Detach the handle first so repeated shutdown calls remain harmless
	database := backendDb
	backendDb = nil
	if database == nil {
		return nil
	}

	// Bound maintenance so shutdown cannot wait indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), databaseShutdownTimeout)
	defer cancel()

	// Collect maintenance errors without skipping the mandatory pool close
	var closeErrors []error
	if errOptimize := database.WithContext(ctx).Exec("PRAGMA optimize").Error; errOptimize != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not optimize backend database: %w", errOptimize))
	}
	if errVacuum := database.WithContext(ctx).Exec("VACUUM").Error; errVacuum != nil {
		closeErrors = append(closeErrors, fmt.Errorf("could not vacuum backend database: %w", errVacuum))
	}

	// Close the pool even when either maintenance operation failed
	if errClose := utils.CloseDatabasePool(database); errClose != nil {
		closeErrors = append(closeErrors, errClose)
	}

	// Return every shutdown error after attempting all cleanup steps
	return errors.Join(closeErrors...)
}

// AutoMigrate migrates the backend database tables to the latest structure
func AutoMigrate() error {
	return backendDb.AutoMigrate(
		&T_group{},
		&T_event{},
		&T_user{},
		&T_ownership{},
	)
}

// Create inserts a value into the backend database
func Create(value interface{}) (tx *gorm.DB) {
	return backendDb.Create(value)
}

// DeploySampleData applies development defaults and sample data to the backend database
func DeploySampleData() error {

	// Prepare sample users
	var mailSample1 = "user1@domain.tld"
	var userEntrySample1 *T_user
	var userEntrySample2 *T_user

	// Create sample user 1 if not existing
	userEntry, _ := GetUserByMail(mailSample1)
	if userEntry == nil {

		// Prepare sample user
		userEntrySample1 = NewUser(
			mailSample1,
			"domain.tld",
			"Dep 1",
			"",
			"Name",
			"Surname",
		)
		userEntrySample1.Admin = true
		userEntrySample1.Gender = "M"

		// Create sample user in DB
		errCreate := userEntrySample1.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create second sample user
		userEntrySample2 = NewUser("user2@domain.tld", "domain.tld", "Dep 1", "", "User", "2")
		userEntrySample2.LastLogin = time.Now().Add(-(time.Hour * 6))
		userEntrySample2.Gender = "M"
		errCreate = userEntrySample2.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create third sample user
		sampleUser3 := NewUser("user3@domain.tld", "domain.tld", "Dep 2", "", "User", "3")
		sampleUser3.LastLogin = time.Now().Add(-(time.Hour * 12))
		sampleUser3.Gender = ""
		_ = sampleUser3.Create()

		// Create fourth sample user
		sampleUser4 := NewUser("name1@own.tld", "own.tld", "Dep A", "", "Name1", "Own")
		sampleUser4.LastLogin = time.Now().Add(-(time.Hour * 16))
		errCreate = sampleUser4.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create fifth sample user (demo mode — can view but not execute modules)
		sampleUser5 := NewUser("name2@own.tld", "own.tld", "Dep B", "", "Name2", "Own")
		sampleUser5.LastLogin = time.Now().Add(-(time.Hour * 60))
		sampleUser5.Demo = true
		_ = sampleUser5.Create()

		// Create sixth sample user
		sampleUser6 := NewUser("name1@new.tld", "new.tld", "", "", "Customer", "New")
		sampleUser6.LastLogin = time.Now().Add(-(time.Hour * 80))
		errCreate = sampleUser6.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create seventh sample user (inactive — login disabled)
		sampleUser7 := NewUser("inactive@domain.tld", "domain.tld", "Dep 1", "", "Inactive", "User")
		sampleUser7.LastLogin = time.Now().Add(-(time.Hour * 24 * 200))
		sampleUser7.Active = false
		_ = sampleUser7.Create()

		// Create some sample events covering all event types across multiple users
		day := time.Hour * 24
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 2), Event: EventLogin})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 10), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 20), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 30), Event: EventScopeCreate, EventDetail: "Scope Alpha"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 40), Event: EventScopeSecret, EventDetail: "Scope Alpha"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 50), Event: EventViewGrant, EventDetail: "user2@domain.tld"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 60), Event: EventViewToken, EventDetail: "Read-only token"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 65), Event: EventDatabaseAdd, EventDetail: "db.internal:5432"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 70), Event: EventApiToken})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 90), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 100), Event: EventScopeCreate, EventDetail: "Scope Beta"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 110), Event: EventViewGrant, EventDetail: "name1@own.tld"})
		backendDb.Create(&T_event{IdTUser: userEntrySample1.Id, Email: userEntrySample1.Email, Timestamp: time.Now().Add(-day * 150), Event: EventDbPassword})

		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 5), Event: EventLogin})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 10), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 30), Event: EventScopeCreate, EventDetail: "Scope Gamma"})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 45), Event: EventViewToken, EventDetail: "CI pipeline token"})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 60), Event: EventScopeSecret, EventDetail: "Scope Gamma"})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 90), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-day * 150), Event: EventApiToken})

		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-day * 58), Event: EventLogin})
		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-day * 59), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-day * 61), Event: EventViewGrant, EventDetail: "name2@own.tld"})
	} else {
		userEntrySample1 = userEntry
	}

	// Ensure sampleUser2 is available for group ownership, even if user1 was already in the database
	if userEntrySample2 == nil {
		userEntrySample2, _ = GetUserByMail("user2@domain.tld")
	}

	// Keep the shared sample identity aligned with the sibling applications
	if userEntrySample2.Gender != "M" {
		userEntrySample2.Gender = "M"
		_, errSaveGender := userEntrySample2.Save("gender")
		if errSaveGender != nil {
			return fmt.Errorf("could not update sample user 'user2@domain.tld': %w", errSaveGender)
		}
	}

	// Create sample group
	group, _ := GetGroupById(1)
	if group == nil {

		// Prepare dev sample group
		groupEntrySample := T_group{
			Name:         "Dev Group",
			Created:      time.Now(),
			CreatedBy:    userEntrySample1.Email,
			DbServerId:   1,
			MaxScopes:    10,
			MaxViews:     10,
			MaxTargets:   20000000,
			MaxOwners:    100,
			AllowCustom:  true,
			AllowNetwork: true,
			AllowAsset:   true,
		}
		errCreate := groupEntrySample.Create()
		if errCreate != nil {
			return errCreate
		}
		errAdd := groupEntrySample.AddOwner(userEntrySample1)
		if errAdd != nil {
			return errAdd
		}
		if userEntrySample2 != nil {
			if errAdd = groupEntrySample.AddOwner(userEntrySample2); errAdd != nil {
				return errAdd
			}
		}

		// Create some more sample groups
		groupEntrySample = T_group{
			Name:         "Dummy Group 2",
			Created:      time.Now(),
			CreatedBy:    userEntrySample1.Email,
			MaxScopes:    11,
			MaxViews:     10,
			MaxTargets:   22,
			MaxOwners:    33,
			AllowCustom:  false,
			AllowNetwork: true,
			AllowAsset:   true,
		}
		_ = groupEntrySample.Create()
		if userEntrySample2 != nil {
			_ = groupEntrySample.AddOwner(userEntrySample2)
		}
		groupEntrySample = T_group{
			Name:       "Dummy Group 3",
			Created:    time.Now(),
			CreatedBy:  userEntrySample1.Email,
			MaxScopes:  11,
			MaxViews:   10,
			MaxTargets: 22,
			MaxOwners:  33,

			AllowCustom:  true,
			AllowNetwork: false,
			AllowAsset:   false,
		}
		_ = groupEntrySample.Create()
	}

	// Return nil as everything went fine
	return nil

}
