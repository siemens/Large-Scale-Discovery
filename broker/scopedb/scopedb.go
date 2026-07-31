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
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"gorm.io/gorm"
)

// DbValueSeparator separates multiple values stored in a database field
const DbValueSeparator = "\n"

// CleanExceededDiscovery sets discovery entries that exceeded their calculated timeout to failed
func CleanExceededDiscovery(
	scopeDb *gorm.DB,
	nmapHostgroup int,
	nmapHostgroupTimeoutMinutes int,
) (int64, error) {

	// Define the failed state written for exceeded discovery work
	values := map[string]interface{}{
		"scan_finished": time.Now(),
		"scan_status":   scanUtils.StatusFailed,
		"priority":      false,
	}

	// Apply the timeout relative to each discovery input size
	db := scopeDb.Model(&scopedb.T_discovery{}).
		Where("scan_finished IS NULL").
		Where(
			"scan_started < NOW() - ((((CEIL(CAST (input_size AS FLOAT) / ?)) * ?)+30)::text || ' minutes')::interval",
			nmapHostgroup,
			nmapHostgroupTimeoutMinutes,
		).
		Updates(values)
	if db.Error != nil {
		return 0, db.Error
	}

	// Return the amount of discovery entries marked failed
	return db.RowsAffected, nil
}

// CleanExceeded sets scan-module entries that exceeded their timeout to failed
func CleanExceeded(scopeDb *gorm.DB, label string, startedBefore time.Time) (int64, error) {

	// Select the database model associated with the scan module
	var moduleType interface{}
	switch label {
	case banner.Label:
		moduleType = scopedb.T_banner{}
	case nfs.Label:
		moduleType = scopedb.T_nfs{}
	case smb.Label:
		moduleType = scopedb.T_smb{}
	case ssl.Label:
		moduleType = scopedb.T_ssl{}
	case ssh.Label:
		moduleType = scopedb.T_ssh{}
	case webcrawler.Label:
		moduleType = scopedb.T_webcrawler{}
	case webenum.Label:
		moduleType = scopedb.T_webenum{}
	default:
		return 0, fmt.Errorf("invalid module label '%s'", label)
	}

	// Convert the timeout threshold into the database column type
	threshold := sql.NullTime{Time: startedBefore, Valid: true}

	// Define the failed state written for exceeded scan-module work
	values := map[string]interface{}{
		"scan_finished": time.Now(),
		"scan_status":   scanUtils.StatusFailed,
	}

	// Mark unfinished entries older than the timeout threshold as failed
	db := scopeDb.Model(moduleType).
		Where("scan_finished IS NULL").
		Where("scan_started < ?", threshold).
		Updates(values)
	if db.Error != nil {
		return 0, db.Error
	}

	// Return the amount of scan-module entries marked failed
	return db.RowsAffected, nil
}

// prepareInfoEntry is a generic helper that creates info entries in the scope database to indicate active scan processes.
// It handles panic recovery, WaitGroup notification, batch creation with Postgres unique constraint fallback.
func prepareInfoEntry[T any](
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
	moduleName string,
	batchSize int,
	entryFactory func(id uint64, cols scopedb.ColumnsScan) T,
) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
			panic(r)
		}
	}()

	// Notify the WaitGroup that we're done
	defer wg.Done()

	// Log action
	logger.Debugf("Preparing %s info entry.", moduleName)

	// Return if there were no results to be created
	if len(idsTDiscoveryService) == 0 {
		logger.Debugf("No discovery services for %s info entries provided.", moduleName)
		return
	}

	// Open scope's database
	scopeDb, errScopeDb := scopedb.GetScopeDbHandle(logger, scopeEntry)
	if errScopeDb != nil {
		logger.Errorf("Could not open scope database: %s", errScopeDb)
		return
	}

	// Prepare common scan columns
	cols := scopedb.ColumnsScan{
		ScanStarted:  scanStarted,
		ScanFinished: sql.NullTime{Valid: false},
		ScanStatus:   scanUtils.StatusRunning,
		ScanIp:       scanIp,
		ScanHostname: scanHostname,
	}

	// Iterate service IDs to create database info entries for
	infoEntries := make([]T, 0, len(idsTDiscoveryService))
	for _, idTDiscoveryService := range idsTDiscoveryService {
		infoEntries = append(infoEntries, entryFactory(idTDiscoveryService, cols))
	}

	// Create info entries, if not yet existing. It might exist from a previous scan attempt (if the scan crashed).
	// This check which is applied here is specific to postgres.
	// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
	// exceed the database's limit of 65535 parameters
	errDb := scopeDb.
		Session(&gorm.Session{CreateBatchSize: batchSize}).
		Create(&infoEntries).Error

	// Extract Postgres error if existing
	var errPg *pgconn.PgError
	errors.As(errDb, &errPg)

	// Handle database errors
	moduleNameTitle := strings.ToUpper(moduleName[:1]) + moduleName[1:]
	if errPg != nil && errPg.Code == "23505" { // Code for Postgres' unique constraint violation

		// Fall back to inserting the entries one by one to ensure as many entries as possible being added to the db
		for i, entry := range infoEntries {
			errDb2 := scopeDb.Create(&entry).Error

			// Extract Postgres error if existing
			var errPg2 *pgconn.PgError
			errors.As(errDb2, &errPg2)

			// Handle database errors
			if errPg2 != nil && errPg2.Code == "23505" { // Code for Postgres' unique constraint violation
				logger.Debugf("%s info entry %d already existing.", moduleNameTitle, idsTDiscoveryService[i])
			} else if errDb2 != nil {
				logger.Errorf("%s info entry %d could not be created: %s", moduleNameTitle, idsTDiscoveryService[i], errDb2)
			}
		}
	} else if errDb != nil {
		logger.Errorf("%s info entries could not be created: %s", moduleNameTitle, errDb)
	}
}
