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
	"fmt"
	"sync"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webenum"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// PrepareWebenumResult creates an info entry in the database to indicate an active process.
func PrepareWebenumResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {
	prepareInfoEntry(logger, scopeEntry, idsTDiscoveryService, scanStarted, scanIp, scanHostname, wg,
		"webenum", scopedb.MaxBatchSizeWebenum,
		func(id uint64, cols scopedb.ColumnsScan) scopedb.T_webenum {
			return scopedb.T_webenum{IdTDiscoveryService: id, ColumnsScan: cols}
		},
	)
}

// SaveWebenumResult parses a result and adds its data into the database. Furthermore, it sets the "scan_finished"
// timestamp to indicate a completed process.
func SaveWebenumResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *webenum.Result,
) error {

	// Log action
	logger.Debugf("Saving webenum scan result.")

	// Open scope's database
	scopeDb, errScopeDb := scopedb.GetScopeDbHandle(logger, scopeEntry)
	if errScopeDb != nil {
		return fmt.Errorf("could not open scope database: %s", errScopeDb)
	}

	// Start transaction on the scope database. The new Transaction function will commit if the provided function
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

		// Prepare memory for result
		var webenumEntry = scopedb.T_webenum{}
		var webenumEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&webenumEntry).
			Count(&webenumEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope database: %s", errDb)
		}

		// Check if info entry was found
		if webenumEntryCount != 1 {
			logger.Infof("Dropping webenum result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if !webenumEntry.ScanFinished.Valid {

			// Add scan results
			errAdd := addWebenumResults(logger, txScopeDb, webenumEntry.Id, idTDiscoveryService, result)
			if errAdd != nil {
				return fmt.Errorf("could not insert result into scope database: %s", errAdd)
			}

			// Add scan status to info entry. Some scan modules might have even more data stored in the info column
			webenumEntry.ScanStatus = result.Status
		}

		// Update finished timestamp. Set or update it to the current time.
		webenumEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&webenumEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(webenumEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope database: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}

func addWebenumResults(
	logger scanUtils.Logger,
	txScopeDb *gorm.DB,
	tWebenumInfoId uint64,
	idTDiscoveryService uint64,
	result *webenum.Result,
) error {

	// Append scan results
	webenumResultEntries := make([]scopedb.T_webenum_results, 0, len(result.Data))
	for _, item := range result.Data {

		// Make sure there are no byte sequences left, the database might have issues with.
		// This might be the case when it was not clear which encoding the web server used in its HTTP response.
		sanitizedVhost := utils.ValidUtf8String(item.Vhost)
		sanitizedAuthMethod := utils.ValidUtf8String(item.AuthMethod)
		sanitizedResponseHeaders := utils.ValidUtf8String(item.ResponseHeaders) // e.g. 0x92 Windows-1252 smart quote might be (though should not be) contained
		sanitizedHtmlTitle := utils.ValidUtf8String(item.HtmlTitle)
		sanitizedHtmlContent := utils.ToValidUtf8String(item.HtmlContent)

		// Prepare data entry
		webenumResultEntries = append(webenumResultEntries, scopedb.T_webenum_results{
			IdTDiscoveryService: idTDiscoveryService,
			IdTWebenum:          tWebenumInfoId,
			Name:                item.Name,
			Vhost:               sanitizedVhost,
			Url:                 item.Url,
			RedirectUrl:         item.RedirectUrl,
			RedirectCount:       item.RedirectCount,
			RedirectOut:         item.RedirectOut,
			AuthMethod:          sanitizedAuthMethod,
			AuthSuccess:         item.AuthSuccess,
			ResponseCode:        item.ResponseCode,
			ResponseMessage:     item.ResponseMessage,
			ResponseContentType: item.ResponseContentType,
			ResponseHeaders:     sanitizedResponseHeaders,
			ResponseEncoding:    item.ResponseEncoding,
			HtmlTitle:           sanitizedHtmlTitle,
			HtmlContent:         sanitizedHtmlContent,
			HtmlContentLength:   len(sanitizedHtmlContent),
		})
	}

	// Insert data into db. Empty slices would result in an error and we don't consider an empty slice an error, as
	// there might simply be no results
	if len(webenumResultEntries) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: scopedb.MaxBatchSizeWebenumResult}).
			Create(&webenumResultEntries).Error
		if errDb != nil {
			return errDb
		}
	}

	// Return nil as everything went fine
	return nil
}
