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
	"fmt"
	"time"

	"github.com/sanyokbig/pqinterval"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
)

// CreateSqlLog creates a new query log entry in the PgProxy database
func CreateSqlLog(
	logger scanUtils.Logger,
	dbName string,
	dbUser string,
	dbTable string,
	query string,
	queryResults int,
	queryTimestamp time.Time,
	queryDuration pqinterval.Duration,
	totalDuration pqinterval.Duration,
	clientName string,
) (*T_sql_log, error) {

	// Get the scope metadata associated with the targeted database
	scopeEntry, errScopeEntry := database.GetScopeEntryByName(dbName)
	if errScopeEntry != nil {
		return nil, fmt.Errorf("could not query database server entry for '%s': %w", dbName, errScopeEntry)
	}
	if scopeEntry == nil {
		return nil, fmt.Errorf("could not find database server entry for '%s'", dbName)
	}

	// Open the associated scope database
	scopeDb, errScopeDb := GetScopeDbHandle(logger, scopeEntry)
	if errScopeDb != nil {
		return nil, fmt.Errorf("could not open scope database: %w", errScopeDb)
	}

	// Prepare the persisted SQL log entry
	logEntry := &T_sql_log{
		DbName:         dbName,
		DbUser:         dbUser,
		DbTable:        dbTable,
		Query:          query,
		QueryResults:   queryResults,
		QueryTimestamp: queryTimestamp,
		QueryDuration:  queryDuration,
		TotalDuration:  totalDuration,
		ClientName:     clientName,
	}

	// Persist the query details in the scope database
	errDb := scopeDb.Create(logEntry).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return the created log entry
	return logEntry, nil
}

// GetSqlLogs returns a list of log entries for given search arguments
func GetSqlLogs(logger scanUtils.Logger, scopeEntry *database.T_scan_scope, since time.Time) ([]T_sql_log, error) {

	// Open the associated scope database
	scopeDb, errScopeDb := GetScopeDbHandle(logger, scopeEntry)
	if errScopeDb != nil {
		return nil, fmt.Errorf("could not open scope database: %w", errScopeDb)
	}

	// Query all entries newer than the requested lower bound
	var logEntries []T_sql_log
	db := scopeDb.Where("query_timestamp > ?", since).Find(&logEntries)
	if db.Error != nil {
		return nil, db.Error
	}

	// Return the matching log entries
	return logEntries, nil
}
