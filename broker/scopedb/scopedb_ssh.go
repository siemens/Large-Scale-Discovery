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

	"github.com/siemens/GoScans/ssh"
	scanUtils "github.com/siemens/GoScans/utils"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"gorm.io/gorm"
)

// PrepareSshResult creates an info entry in the database to indicate an active process.
func PrepareSshResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {
	prepareInfoEntry(logger, scopeEntry, idsTDiscoveryService, scanStarted, scanIp, scanHostname, wg,
		"SSH", scopedb.MaxBatchSizeSsh,
		func(id uint64, cols scopedb.ColumnsScan) scopedb.T_ssh {
			return scopedb.T_ssh{IdTDiscoveryService: id, ColumnsScan: cols}
		},
	)
}

// SaveSshResult parses a result and adds its data into the database. Furthermore, it sets the "scan_finished"
// timestamp to indicate a completed process.
func SaveSshResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *ssh.Result,
) error {

	// Log action
	logger.Debugf("Saving SSH scan result.")

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
		var sshEntry = scopedb.T_ssh{}
		var sshEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&sshEntry).
			Count(&sshEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope database: %s", errDb)
		}

		// Check if info entry was found
		if sshEntryCount != 1 {
			logger.Infof("Dropping SSH result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if !sshEntry.ScanFinished.Valid {

			// Add scan result to db entry
			sshEntry.AuthenticationMechanisms = strings.Join(result.Data.AuthenticationMechanisms, DbValueSeparator)
			sshEntry.KeyExchangeAlgorithms = strings.Join(result.Data.KeyExchangeAlgorithms, DbValueSeparator)
			sshEntry.ServerKeyAlgorithms = strings.Join(result.Data.ServerKeyAlgorithms, DbValueSeparator)
			sshEntry.ServerEncryptionAlgorithms = strings.Join(result.Data.ServerEncryptionAlgorithms, DbValueSeparator)
			sshEntry.ServerMacAlgorithms = strings.Join(result.Data.ServerMacAlgorithms, DbValueSeparator)
			sshEntry.ServerCompressAlgorithms = strings.Join(result.Data.ServerCompressAlgorithms, DbValueSeparator)
			sshEntry.UsesGuessedKeyExchange = result.Data.UsesGuessedKeyExchange
			sshEntry.ProtocolVersion = result.Data.ProtocolVersion

			// Add scan status to info entry. Some scan modules might have even more data stored in the info column
			sshEntry.ScanStatus = result.Status
		}

		// Update finished timestamp. Set or update it to the current time.
		sshEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&sshEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(sshEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope database: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}
