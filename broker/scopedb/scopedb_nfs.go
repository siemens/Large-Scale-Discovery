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
	"strings"
	"sync"
	"time"

	"github.com/siemens/GoScans/nfs"
	scanUtils "github.com/siemens/GoScans/utils"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"gorm.io/gorm"
)

// PrepareNfsResult creates an info entry in the database to indicate an active process.
func PrepareNfsResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {
	prepareInfoEntry(logger, scopeEntry, idsTDiscoveryService, scanStarted, scanIp, scanHostname, wg,
		"NFS", scopedb.MaxBatchSizeNfs,
		func(id uint64, cols scopedb.ColumnsScan) scopedb.T_nfs {
			return scopedb.T_nfs{IdTDiscoveryService: id, ColumnsScan: cols}
		},
	)
}

// SaveNfsResult parses a result and adds its data into the database. Furthermore, it sets the "scan_finished"
// timestamp to indicate a completed process.
func SaveNfsResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *nfs.Result,
) error {

	// Log action
	logger.Debugf("Saving NFS scan result.")

	// Open scope's database
	scopeDb, errScopeDb := scopedb.GetScopeDbHandle(logger, scopeEntry)
	if errScopeDb != nil {
		return fmt.Errorf("could not open scope database: %s", errScopeDb)
	}

	// Start transaction on the scope database. The new Transaction function will commit if the provided function returns nil
	// and rollback if an error is returned.
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
		var nfsEntry = scopedb.T_nfs{}
		var nfsEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&nfsEntry).
			Count(&nfsEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope database: %s", errDb)
		}

		// Check if info entry was found
		if nfsEntryCount != 1 {
			logger.Infof("Dropping NFS result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if !nfsEntry.ScanFinished.Valid {

			// Add file system crawler information
			errAdd := addNfsResults(txScopeDb, nfsEntry.Id, idTDiscoveryService, result)
			if errAdd != nil {
				return fmt.Errorf("could not insert result into scope database: %s", errAdd)
			}

			// Add scan status to info entry and additional info to the info column
			nfsEntry.ScanStatus = result.Status
			nfsEntry.FoldersReadable = result.FoldersReadable
			nfsEntry.FilesReadable = result.FilesReadable
			nfsEntry.FilesWritable = result.FilesWritable
		}

		// Update finished timestamp. Set or update it to the current time.
		nfsEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&nfsEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(nfsEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope database: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}

func addNfsResults(txScopeDb *gorm.DB, tNfsInfoId uint64, idTDiscoveryService uint64, result *nfs.Result) error {

	// Append scan results
	nfsFileEntries := make([]scopedb.T_nfs_file, 0, len(result.Data))
	for _, item := range result.Data {

		// Prepare data entry
		nfsFileEntries = append(nfsFileEntries, scopedb.T_nfs_file{
			IdTDiscoveryService: idTDiscoveryService,
			IdTNfs:              tNfsInfoId,
			Share:               item.Share,
			Path:                item.Path,
			Name:                item.Name,
			Extension:           item.Extension,
			Mime:                item.Mime,
			Readable:            item.Readable,
			Writable:            item.Writable,
			Flags:               item.Flags,
			SizeKb:              item.SizeKb,
			LastModified:        sql.NullTime{Time: item.LastModified, Valid: item.LastModified != time.Time{}},
			Depth:               item.Depth,
			Properties:          strings.Join(item.Properties, DbValueSeparator),
			IsSymlink:           item.IsSymlink,
			Restrictions:        strings.Join(item.NfsRestrictions, DbValueSeparator),
		})
	}

	// Insert data into db. Empty slices would result in an error and we don't consider an empty slice an error, as
	// there might simply be no results
	if len(nfsFileEntries) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: scopedb.MaxBatchSizeNfsFile}).
			Create(&nfsFileEntries).Error
		if errDb != nil {
			return errDb
		}
	}
	// Return nil as everything went fine
	return nil
}
