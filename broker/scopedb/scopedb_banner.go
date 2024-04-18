/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
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
	"github.com/jackc/pgconn"
	"github.com/siemens/GoScans/banner"
	scanUtils "github.com/siemens/GoScans/utils"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	"sync"
	"time"
)

// PrepareBannerResults creates info entries in the database, if not existing, to indicate an active process.
func PrepareBannerResults(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Notify the WaitGroup that we're done
	defer wg.Done()

	// Log action
	logger.Debugf("Preparing banner info entry.")

	// Return if there were no results to be created
	if len(idsTDiscoveryService) == 0 {
		logger.Debugf("No discovery services for banner info entries provided.")
		return
	}

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		logger.Errorf("Could not open scope database: %s", errHandle)
		return
	}

	// Iterate service IDs to create database info entries for
	infoEntries := make([]managerdb.T_banner, 0, len(idsTDiscoveryService))
	for _, idTDiscoveryService := range idsTDiscoveryService {

		// Prepare info entry
		infoEntries = append(infoEntries, managerdb.T_banner{
			IdTDiscoveryService: idTDiscoveryService,
			ColumnsScan: managerdb.ColumnsScan{
				ScanStarted:  scanStarted,
				ScanFinished: sql.NullTime{Valid: false},
				ScanStatus:   scanUtils.StatusRunning,
				ScanIp:       scanIp,
				ScanHostname: scanHostname,
			},
		})
	}

	// Create info entries, if not yet existing. It might exist from a previous scan attempt (if the scan crashed).
	// This check which is applied here is specific to postgres. Use a new gorm session and force a limit on how
	// many Entries can be batched, as we otherwise might exceed PostgreSQLs limit of 65535 parameters
	errDb := scopeDb.
		Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeBanner}).
		Create(&infoEntries).Error
	var errCreate *pgconn.PgError
	if errors.As(errDb, &errCreate) && errCreate.Code == "23505" { // Code for unique constraint violation

		// Fall back to inserting the entries one by one to ensure as many entries as possible being added to the db
		for _, entry := range infoEntries {
			errDb2 := scopeDb.Create(&entry).Error
			var errCreate2 *pgconn.PgError
			if errors.As(errDb2, &errCreate2) && errCreate2.Code == "23505" { // Code for unique constraint violation
				logger.Debugf("Banner info entry '%d' already existing.", entry.IdTDiscoveryService)
			} else if errDb2 != nil {
				logger.Errorf("Banner info entry '%d' could not be created: %s", entry.IdTDiscoveryService, errDb2)
			}
		}
	} else if errDb != nil {
		logger.Errorf("Banner info entries could not be created: %s", errDb)
	}
}

// SaveBannerResult parses a result and adds it to the database. Furthermore, it sets the "scan_finished" timestamp to
// indicate a completed process.
func SaveBannerResult(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *banner.Result,
) error {

	// Log action
	logger.Debugf("Saving banner scan result.")

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		return fmt.Errorf("could not open scope database: %s", errHandle)
	}

	// Start transaction on the scoped db. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {
		// Check the condition of the info entry stored in the database. It got created the first time the target was
		// taken from the cache (info entry should never be removed from the scope database again!). However,
		// 		A) the scan could have crashed before and been restarted by the broker after its timeout
		// 		B) the broker could have been down for a prolonged time, restarting the scan after an assumed timeout.
		// 		   Both scan instances might deliver results in arbitrary order.
		//
		// Case A: The scan_finished timestamp is not set, because this is the first scan instances delivering results
		// 		    => Save results, set scan_finished timestamp
		// Case B: The scan_finished timestamp is already set, because this is a later scan instance delivering results
		// 			=> Discard results (to prevent later manipulation by client), but update scan_finished timestamp (to
		// 			   allow users proper investigations in case of IDS alerts).

		// Prepare query result
		var infoEntry = managerdb.T_banner{}
		var infoEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&infoEntry).
			Count(&infoEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope db: %s", errDb)
		}

		// Check if info entry was found
		if infoEntryCount != 1 {
			logger.Infof("Dropping banner result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if infoEntry.ScanFinished.Valid == false {

			// Add scan result to db entry
			infoEntry.TriggerPlain = utils.ToValidUtf8String(result.Data.Plain)
			infoEntry.TriggerSsl = utils.ToValidUtf8String(result.Data.Ssl)
			infoEntry.TriggerTelnet = utils.ToValidUtf8String(result.Data.Telnet)
			infoEntry.TriggerHttp = utils.ToValidUtf8String(result.Data.Http)
			infoEntry.TriggerHttps = utils.ToValidUtf8String(result.Data.Https)

			// Add scan status to info entry. Some scan modules might have even more data stored in the info column
			infoEntry.ScanStatus = result.Status
		}

		// Update finished timestamp. Set or update it to the current time.
		infoEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&infoEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(infoEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope db: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}
