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
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

var backendDb *gorm.DB // If desired public, code is most likely in the wrong package!

// Open opens the backendDb from disk
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
	var errOpen error
	backendDb, errOpen = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Warn),
	})
	if errOpen != nil {
		return errOpen
	}

	// Set DB log mode when development mode is enabled
	if _build.DevMode {
		backendDb.Logger = backendDb.Logger.LogMode(gormlog.Info) // Apply log mode to database
	}

	// Check if pragmas were set correctly
	errPragmas := utils.CheckPragmas(backendDb, map[string]interface{}{
		"journal_mode": "wal", // string
		"busy_timeout": 60000, // int (ms)
		"foreign_keys": 1,     // int
	})
	if errPragmas != nil {
		return fmt.Errorf("could not set PRAGMAs in manager db: %s", errPragmas)
	}

	// Return nil as everything went fine
	return nil
}

// Close closes an open backendDb
func Close() error {
	if backendDb != nil {

		// Check for potential query optimizations and install them (to be done before closing connection)
		backendDb.Exec("PRAGMA optimize") // https://www.sqlite.org/pragma.html#pragma_module_list

		// Cleanup database to shrink file size
		errVacuum := backendDb.Exec("VACUUM").Error
		if errVacuum != nil {
			return fmt.Errorf("could not vacuum DB: %s", errVacuum)
		}

		// Retrieve and close sql DB connection
		sqlDb, errDb := backendDb.DB()
		if errDb != nil {
			return fmt.Errorf("could not retrieve underlying DB connection: %s", errDb)
		}
		errClose := sqlDb.Close()
		if errClose != nil {
			return fmt.Errorf("could not close DB connection: %s", errClose)
		}
	}

	return nil
}

// AutoMigrate migrates the cache database's tables to the latest structure
func AutoMigrate() error {
	return backendDb.AutoMigrate(
		&T_group{},
		&T_event{},
		&T_user{},
		&T_ownership{},
	)
}

func Create(value interface{}) (tx *gorm.DB) {
	return backendDb.Create(value)
}

// DeploySampleData applies a default configuration for development purposes and some sample data to the db
func DeploySampleData() error {

	// Prepare sample users
	var sampleMail1 = "user1@domain.tld"
	var sampleUser1 *T_user
	var sampleUser2 *T_user

	// Create sample user 1 if not existing
	userEntry, _ := GetUserByMail(sampleMail1)
	if userEntry == nil {

		// Prepare sample user
		sampleUser1 = NewUser(
			sampleMail1,
			"domain.tld",
			"Dep 1",
			"Name",
			"Surname",
		)
		sampleUser1.Admin = true

		// Create sample user in DB
		errCreate := sampleUser1.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create second sample user
		sampleUser2 = NewUser("user2@domain.tld", "domain.tld", "Dep 1", "User", "2")
		sampleUser2.LastLogin = time.Now().Add(-(time.Hour * 6))
		errCreate = sampleUser2.Create()
		if errCreate != nil {
			return errCreate
		}

		// Crate third sample users
		sampleUser3 := NewUser("user3@domain.tld", "domain.tld", "Dep 2", "User", "3")
		sampleUser3.LastLogin = time.Now().Add(-(time.Hour * 12))
		_ = sampleUser3.Create()

		// Crate fourth sample users
		sampleUser4 := NewUser("name1@own.tld", "own.tld", "Dep A", "Name1", "Own")
		sampleUser4.LastLogin = time.Now().Add(-(time.Hour * 16))
		errCreate = sampleUser4.Create()
		if errCreate != nil {
			return errCreate
		}

		// Crate fifth sample users
		sampleUser5 := NewUser("name2@own.tld", "own.tld", "Dep B", "Name2", "Own")
		sampleUser5.LastLogin = time.Now().Add(-(time.Hour * 60))
		_ = sampleUser5.Create()

		// Crate sixth sample users
		sampleUser6 := NewUser("name1@new.tld", "new.tld", "", "Customer", "New")
		sampleUser6.LastLogin = time.Now().Add(-(time.Hour * 80))
		errCreate = sampleUser6.Create()
		if errCreate != nil {
			return errCreate
		}

		// Create some sample events
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 10), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 20), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 30), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 40), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 60), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 65), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 70), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 90), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 100), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 101), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 102), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 103), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser1.Id, Email: sampleUser1.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 150), Event: EventDbPassword})

		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 10), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 30), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 60), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 65), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 90), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 101), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser4.Id, Email: sampleUser4.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 150), Event: EventDbPassword})

		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 60), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 60), Event: EventDbPassword})
		backendDb.Create(&T_event{IdTUser: sampleUser6.Id, Email: sampleUser6.Email, Timestamp: time.Now().Add(-time.Hour * 24 * 61), Event: EventDbPassword})
	} else {
		sampleUser1 = userEntry
	}

	// Create sample group
	group, _ := GetGroup(1)
	if group == nil {

		// Prepare dev sample group
		sampleGroup := T_group{
			Name:         "Dev Group",
			Created:      time.Now(),
			CreatedBy:    sampleUser1.Email,
			DbServerId:   1,
			MaxScopes:    10,
			MaxViews:     10,
			MaxTargets:   20000000,
			MaxOwners:    100,
			AllowCustom:  true,
			AllowNetwork: true,
			AllowAsset:   true,
		}
		errCreate := sampleGroup.Create()
		if errCreate != nil {
			return errCreate
		}
		errAdd := sampleGroup.AddOwner(sampleUser1)
		if errAdd != nil {
			return errAdd
		}
		errAdd = sampleGroup.AddOwner(sampleUser2)
		if errAdd != nil {
			return errAdd
		}

		// Create some more sample groups
		sampleGroup = T_group{
			Name:         "Dummy Group 2",
			Created:      time.Now(),
			CreatedBy:    sampleUser1.Email,
			MaxScopes:    11,
			MaxViews:     10,
			MaxTargets:   22,
			MaxOwners:    33,
			AllowCustom:  false,
			AllowNetwork: true,
			AllowAsset:   true,
		}
		_ = sampleGroup.Create()
		_ = sampleGroup.AddOwner(sampleUser2)
		sampleGroup = T_group{
			Name:       "Dummy Group 3",
			Created:    time.Now(),
			CreatedBy:  sampleUser1.Email,
			MaxScopes:  11,
			MaxViews:   10,
			MaxTargets: 22,
			MaxOwners:  33,

			AllowCustom:  true,
			AllowNetwork: false,
			AllowAsset:   false,
		}
		_ = sampleGroup.Create()
	}

	// Return nil as everything went fine
	return nil

}
