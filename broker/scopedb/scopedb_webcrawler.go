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

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// PrepareWebcrawlerResult creates an info entry in the database to indicate an active process.
func PrepareWebcrawlerResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {
	prepareInfoEntry(logger, scopeEntry, idsTDiscoveryService, scanStarted, scanIp, scanHostname, wg,
		"webcrawler", scopedb.MaxBatchSizeWebcrawler,
		func(id uint64, cols scopedb.ColumnsScan) scopedb.T_webcrawler {
			return scopedb.T_webcrawler{IdTDiscoveryService: id, ColumnsScan: cols}
		},
	)
}

// SaveWebcrawlerResult parses a result and adds its data into the database. Furthermore, it sets the "scan_finished"
// timestamp to indicate a completed process.
func SaveWebcrawlerResult(
	logger scanUtils.Logger,
	scopeEntry *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *webcrawler.Result,
) error {

	// Log action
	logger.Debugf("Saving webcrawler scan result.")

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
		var webcrawlerEntry = scopedb.T_webcrawler{}
		var webcrawlerEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&webcrawlerEntry).
			Count(&webcrawlerEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope database: %s", errDb)
		}

		// Check if info entry was found
		if webcrawlerEntryCount != 1 {
			logger.Infof("Dropping webcrawler result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if !webcrawlerEntry.ScanFinished.Valid {

			// Add scan results
			joinedStates, errJoinedStates := addWebcrawlerResults(logger, txScopeDb, webcrawlerEntry.Id, idTDiscoveryService, result)
			if errJoinedStates != nil {
				return fmt.Errorf("could not insert result into scope database: %s", errJoinedStates)
			}

			// If the outer status is not a success or the concatenated one is empty, we'd rather want to display the
			// outer one.
			if result.Status != scanUtils.StatusCompleted || joinedStates == "" {
				joinedStates = result.Status
			}

			// Add scan status to info entry.
			webcrawlerEntry.ScanStatus = joinedStates
		}

		// Update finished timestamp. Set or update it to the current time.
		webcrawlerEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&webcrawlerEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(webcrawlerEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope database: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}

func addWebcrawlerResults(
	logger scanUtils.Logger,
	txScopeDb *gorm.DB,
	tWebcrawlerInfoId uint64,
	idTDiscoveryService uint64,
	result *webcrawler.Result,
) (string, error) {

	states := make([]string, 0)

	// Prepare entry slices. We need a slice of pointers for the vhosts, as we need the same reference in order to set
	// it in the respective page entries.
	webcrawlerVhostEntries := make([]*scopedb.T_webcrawler_vhost, 0, len(result.Data))
	webcrawlerPageEntries := make([]scopedb.T_webcrawler_page, 0, 20)

	// We need to first create the vhost entries, as gorm will add appropriate IDs during the Create. These IDs are then
	// needed as a foreign key in the page entries.
	for _, vhost := range result.Data {

		// Check whether we have a new status. If so, add it to the previous ones.
		in := false
		for _, s := range states {
			if vhost.Status == s {
				in = true
				break
			}
		}

		if !in {
			states = append(states, vhost.Status)
		}

		// Make sure there are no byte sequences left, the database might have issues with.
		// This might be the case when it was not clear which encoding the web server used in its HTTP response.
		sanitizedVhost := utils.ValidUtf8String(vhost.Vhost)
		sanitizedAuthMethod := utils.ValidUtf8String(vhost.AuthMethod)
		sanitizedDiscoveredVhosts := utils.ValidUtf8String(strings.Join(vhost.DiscoveredVhosts, DbValueSeparator))
		sanitizedDiscoveredDownloads := utils.ValidUtf8String(strings.Join(vhost.DiscoveredDownloads, DbValueSeparator))

		// Prepare data entry
		webcrawlerVhostEntry := &scopedb.T_webcrawler_vhost{
			IdTDiscoveryService: idTDiscoveryService,
			IdTWebcrawler:       tWebcrawlerInfoId,
			Status:              vhost.Status,
			Vhost:               sanitizedVhost,
			FaviconHash:         vhost.FaviconHash,
			AuthMethod:          sanitizedAuthMethod,
			AuthSuccess:         vhost.AuthSuccess,
			RequestsTotal:       vhost.RequestsTotal,
			RequestsRedirect:    vhost.RequestsRedirect,
			RequestsPartial:     vhost.RequestsPartial,
			RequestsComplete:    vhost.RequestsComplete,
			DiscoveredVhosts:    sanitizedDiscoveredVhosts,
			DiscoveredDownloads: sanitizedDiscoveredDownloads,
		}
		webcrawlerVhostEntries = append(webcrawlerVhostEntries, webcrawlerVhostEntry)

		for _, page := range vhost.Pages {

			// Prepare URL as string
			url := page.Url.String()

			// Make sure there are no byte sequences left, the database might have issues with.
			// This might be the case when it was not clear which encoding the web server used in its HTTP response.
			sanitizedAuthMethodPage := utils.ValidUtf8String(page.AuthMethod)
			sanitizedResponseHeaders := utils.ValidUtf8String(page.ResponseHeaders) // e.g. 0x92 Windows-1252 smart quote might be (though should not be) contained
			sanitizedHtmlTitle := utils.ValidUtf8String(page.HtmlTitle)
			sanitizedHtmlContent := utils.ToValidUtf8String(page.HtmlContent)

			// Prepare page entry
			webcrawlerPageEntries = append(webcrawlerPageEntries, scopedb.T_webcrawler_page{
				IdTDiscoveryService: idTDiscoveryService,
				IdTWebcrawler:       tWebcrawlerInfoId,
				Depth:               page.Depth,
				Url:                 url,
				RedirectUrl:         page.RedirectUrl,
				AuthMethod:          sanitizedAuthMethodPage,
				AuthSuccess:         page.AuthSuccess,
				ResponseCode:        page.ResponseCode,
				ResponseMessage:     page.ResponseMessage,
				ResponseContentType: page.ResponseContentType,
				ResponseHeaders:     sanitizedResponseHeaders,
				ResponseEncoding:    page.ResponseEncoding,
				HtmlTitle:           sanitizedHtmlTitle,
				HtmlContent:         sanitizedHtmlContent,
				HtmlContentLength:   len(sanitizedHtmlContent),
				RawLinks:            strings.Join(page.RawLinks, DbValueSeparator),

				TWebcrawlerVhost: webcrawlerVhostEntry, // Need to use the struct, as we don't know the ID yet
			})
		}
	}

	// Insert vhost entries into the db. Order matters here, as we'd get a unique index violation if we first create the
	// page entries. In this case gorm would:
	// - insert pages, detect that the corresponding vhost entries are not existing
	// - create the missing vhosts (correctly) according to the vhost struct in the page
	// - try to create the vhosts from the webcrawlerVhostEntries slice again
	// Of course we could also let gorm create the hosts this way and not call create on the webcrawlerVhostEntries slice, but it
	// might be not as obvious
	// Empty slices would result in an error and we don't consider an empty slice an error, as there might simply be no
	// results
	if len(webcrawlerVhostEntries) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's' limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: scopedb.MaxBatchSizeWebcrawlerVhost}).
			Create(&webcrawlerVhostEntries).Error
		if errDb != nil {
			return "", fmt.Errorf("%w (vhosts)", errDb) // Wrapped error with prefix/suffix for coherent message
		}
	}

	// Insert pages into db. Empty slices would result in an error and we don't consider an empty slice an error, as
	// there might simply be no results
	if len(webcrawlerPageEntries) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: scopedb.MaxBatchSizeWebcrawlerPage}).
			Create(&webcrawlerPageEntries).Error
		if errDb != nil {
			return "", fmt.Errorf("%w (pages)", errDb) // Wrapped error with prefix/suffix for coherent message
		}
	}

	// Join states to one combined status
	joinedStates := strings.Join(states, ", ")

	// Return nil as everything went fine
	return joinedStates, nil
}
