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
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// OpenForTesting opens a shared in-memory SQLite database for use in tests.
// Call AutoMigrate after this to create the schema.
func OpenForTesting() error {
	var errOpen error
	backendDb, errOpen = gorm.Open(
		sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"),
		&gorm.Config{
			Logger: gormlog.Default.LogMode(gormlog.Silent),
		},
	)
	return errOpen
}
