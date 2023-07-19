/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package scopedb

import (
	"database/sql"
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"gorm.io/gorm"
	managerdb "large-scale-discovery/manager/database"
	"time"
)

const DbValueSeparator = "\n"

// CleanExceededDiscovery sets discovery input entries that never returned results to status failed, after the
// maximum scan time was reached (when there should already have been a result)
func CleanExceededDiscovery(
	scopeDb *gorm.DB,
	nmapHostgroup int, // Scan input size the scan timeout is relative to
	nmapHostgroupTimeoutMinutes int, // Timeout in minutes per (e.g.) 64 systems chunk. Must be scaled proportionally
) (int64, error) {

	// Define update values
	values := map[string]interface{}{
		"scan_finished": time.Now(),
		"scan_status":   scanUtils.StatusFailed,
		"priority":      false, // Also reset priority flag on error (just like on completion)
	}

	// For discovery scans, the timeout must be calculated dynamically, based on the actual input size
	db := scopeDb.Model(&managerdb.T_discovery{}).
		Where("scan_finished IS NULL").                                                                                                                                 // where not finished
		Where("scan_started < NOW() - ((((CEIL(CAST (input_size AS FLOAT) /  ?)) * ?)+30)::text || ' minutes')::interval", nmapHostgroup, nmapHostgroupTimeoutMinutes). // where scan older than its timeout proportional to its target size
		Updates(values)
	if db.Error != nil {
		return 0, db.Error
	}

	// Return affected rows count
	return db.RowsAffected, nil
}

// CleanExceeded sets scan info entries that never returned results to status failed, after the scan module's
// maximum scan time was reached (when there should already have been a result)
func CleanExceeded(
	scopeDb *gorm.DB,
	label string,
	startedBefore time.Time,
) (int64, error) {

	// Prepare database model to update and maximum scan time value
	var model interface{}
	switch label {
	case banner.Label:
		model = managerdb.T_banner{}
	case nfs.Label:
		model = managerdb.T_nfs{}
	case smb.Label:
		model = managerdb.T_smb{}
	case ssl.Label:
		model = managerdb.T_ssl{}
	case ssh.Label:
		model = managerdb.T_ssh{}
	case webcrawler.Label:
		model = managerdb.T_webcrawler{}
	case webenum.Label:
		model = managerdb.T_webenum{}
	default:
		return 0, fmt.Errorf("invalid module label '%s'", label)
	}

	// Convert startedBefore threshold into required data type
	threshold := sql.NullTime{Time: startedBefore, Valid: true}

	// Define update values
	values := map[string]interface{}{
		"scan_finished": time.Now(),
		"scan_status":   scanUtils.StatusFailed,
	}

	// Update input entry with failed state information
	db := scopeDb.Model(model).
		Where("scan_finished IS NULL").       // where not finished
		Where("scan_started < ?", threshold). // where scan older than its time limit
		Updates(values)
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}
