/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/manager/scopedb"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

var devDbServerName = "Local development DB"  // Development database server name
var devDbServerAdmin = "postgres"             // Development database administrator
var devDbServerPassword = "test123!$LSD"      // Development database password
var devScopeMaxInstances = map[string]uint32{ // Development scan-module instance limits
	discovery.Label:  15,
	banner.Label:     10,
	nfs.Label:        10,
	nuclei.Label:     5,
	smb.Label:        10,
	ssh.Label:        10,
	ssl.Label:        10,
	webcrawler.Label: 10,
	webenum.Label:    10,
}

// XCreateScope creates a new scope entry in the manager database and the associated scan database in the scope database.
// Unfortunately, this step cannot happen in a transactional manner.
func XCreateScope(
	serverDb *gorm.DB, // ATTENTION: Cannot be a transaction!! Must be a plain connection, database cannot be created within a transaction.
	dbServerEntry *database.T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	splitSize uint32,
	attributes utils.JsonMap,
	settingsEntry database.T_scan_setting,
) (*database.T_scan_scope, error) {

	// Prepare scan scope memory
	var scopeEntry *database.T_scan_scope

	// Create scope entry in manager database to keep track
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		var errScopeEntry error
		scopeEntry, errScopeEntry = transaction.CreateScopeEntry(
			dbServerEntry,
			name,
			dbName,
			groupId,
			createdBy,
			secret,
			scopeType,
			cycles,
			cyclesRetention,
			splitSize,
			attributes,
			0,
			settingsEntry,
		)
		return errScopeEntry
	})
	if errTx != nil {
		return nil, errTx
	}

	// Create database in scope database
	errCreate := scopedb.CreateScopeDb(serverDb, dbName, name)
	if errCreate != nil {
		return scopeEntry, errCreate // Return scan scope entry, nevertheless, because it got created despite this error.
	}

	// Return newly created scan scope entry
	return scopeEntry, nil
}

// XDeleteScope deletes scope with given scopeId on scope database and manager database in a transactional manner. Views and access
// rights are removed along.
func XDeleteScope(serverDb *gorm.DB, scopeDb *gorm.DB, scopeEntry *database.T_scan_scope) error {

	// Takes care of:
	// 		- removing associated scope views and access rights (manager database and scopedb)
	//		- removing scope database from scope database
	//		- removing scope entry from the manager database
	//		- removing associated agent entries from the manager database
	//		- removing associated settings entry from the manager database

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTxManagerDb := database.WithTransaction(func(transaction *database.Transaction) error {
		errTxScopeDb := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Cleanup views
			for _, scopeView := range scopeEntry.ScopeViews {
				errRevoke := xDeleteView(txScopeDb, transaction, scopeEntry.DbName, &scopeView)
				if errRevoke != nil {
					return errRevoke
				}
			}

			// Remove agent entries from manager database
			for _, scanAgent := range scopeEntry.ScanAgents {
				errDelete := transaction.DeleteAgentEntry(&scanAgent)
				if errDelete != nil {
					return fmt.Errorf(
						"could not remove agent entry '%s' (ID %d) of scan scope '%s' (ID %d): %s",
						scanAgent.Name,
						scanAgent.Id,
						scopeEntry.Name,
						scopeEntry.Id,
						errDelete,
					)
				}
			}

			// Remove settings entry from manager database
			errDelete2 := transaction.DeleteScopeSettings(&scopeEntry.ScanSettings)
			if errDelete2 != nil {
				return fmt.Errorf(
					"could not remove settings entry of scan scope '%s' (ID %d): %s",
					scopeEntry.Name,
					scopeEntry.Id,
					errDelete2,
				)
			}

			// Remove scope entry from manager database
			errDelete3 := transaction.DeleteScopeEntry(scopeEntry)
			if errDelete3 != nil {
				return fmt.Errorf(
					"could not remove scope entry '%s' (ID %d): %s",
					scopeEntry.Name,
					scopeEntry.Id,
					errDelete3,
				)
			}

			// Return nil as everything went fine
			return nil
		})

		// Abort process if inner transaction failed already
		if errTxScopeDb != nil {
			return errTxScopeDb
		}

		// Delete scope's database (can only be done outside a transaction)
		// ATTENTION: Scope database connections will be killed and have to be committed already!
		// ATTENTION: Last step, it cannot be rolled back!
		errDelete4 := scopedb.DeleteScopeDb(serverDb, scopeEntry.DbName)
		var errDb *pq.Error
		if errors.As(errDelete4, &errDb) && errDb.Code.Name() == "invalid_catalog_name" {
			// Scope database not existing anymore
		} else if errDelete4 != nil {
			return fmt.Errorf(
				"could not remove scope database '%s' (ID %d): %s",
				scopeEntry.Name,
				scopeEntry.Id,
				errDelete4,
			)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return transaction result, which may be nil or an error
	return errTxManagerDb
}

// XCycleScope initializes a new scan cycle by resetting the scan input states in the scope database and updating the
// scan progress attributes in the manager database. If a scan cycle retention is configured, this will also delete
// outdated scan results.
func XCycleScope(
	logger scanUtils.Logger,
	scopeDb *gorm.DB,
	scopeEntry *database.T_scan_scope,
	rpcInflight *sync.WaitGroup,
) error {

	// Start transaction on the scope database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// Execute cycle reset by resetting all scan input states
		errScopeDb := txScopeDb.Model(&scopedb.T_discovery{}).
			Where("1=1"). // Gorm doesn't allow global update without where clause
			Updates(map[string]interface{}{
				"scan_started":  sql.NullTime{},
				"scan_finished": sql.NullTime{},
				"scan_status":   scanUtils.StatusWaiting,
			}).Error
		if errScopeDb != nil {
			return fmt.Errorf("could not initialize new scan cycle in scope database: %w", errScopeDb)
		}

		// Update scope attributes in manager database
		scopeEntry.Cycle += 1
		scopeEntry.CycleStarted = time.Now()
		scopeEntry.CycleDone = 0
		scopeEntry.CycleActive = 0
		scopeEntry.CycleFailed = 0

		// Save updated attributes
		_, errManagerDb := scopeEntry.Save(
			"cycle",
			"cycle_started",
			"cycle_done",
			"cycle_active",
			"cycle_failed",
		)
		if errManagerDb != nil {
			return fmt.Errorf("could not initialize new scan cycle in manager database: %s", errManagerDb)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return error if transaction failed
	if errTx != nil {
		return errTx
	}

	// Cleanup old scan results if desired
	if scopeEntry.CyclesRetention >= 1 {

		// Calculate old scan cycles to be cleaned up. +1 because cycle count just got incremented above.
		minCycleKeep := int(scopeEntry.Cycle) - scopeEntry.CyclesRetention

		// Run cleanup asynchronously in the background, so that the other components can continue with
		// the new scan cycle without larger delays. The cleanup query may take a very long time if there
		// are a lot of scan results that need to be cascade-deleted. If cleanup fails, it can be retried
		// next time.
		rpcInflight.Add(1)
		go func() {

			// Release shutdown after this asynchronous database work finishes
			defer rpcInflight.Done()

			// Log potential panics before letting them move on
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t"))
					panic(r)
				}
			}()

			// Log background activity
			logger.Infof("Cleaning up outdated scan cycles of scan scope '%s' (ID %d).",
				scopeEntry.Name,
				scopeEntry.Id,
			)

			// Remember start time
			start := time.Now()

			// Prepare list of tables and order, subsequent queries should be applied against
			subModules := []interface{}{
				scopedb.T_banner{},
				scopedb.T_nfs_file{},
				scopedb.T_nfs{},
				scopedb.T_nuclei_result{},
				scopedb.T_nuclei{},
				scopedb.T_smb_file{},
				scopedb.T_smb{},
				scopedb.T_ssh{},
				scopedb.T_ssl_cipher{},
				scopedb.T_ssl_certificate{},
				scopedb.T_ssl_setting{},
				scopedb.T_ssl_issue{},
				scopedb.T_ssl{},
				scopedb.T_webcrawler_page{},
				scopedb.T_webcrawler_vhost{},
				scopedb.T_webcrawler{},
				scopedb.T_webenum_results{},
				scopedb.T_webenum{},
			}
			discoveryModules := []interface{}{
				scopedb.T_discovery_script{},
				scopedb.T_discovery_service{},
				scopedb.T_discovery_host{},
			}

			// Prepare transaction for cleanup execution
			// To speed up cleanup, large data tables are cleaned using a list of IDs, before relying on
			// cascade deletes. Cascade deletes on large sets of data takes really long.
			errCleanup := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

				// Get t_discovery_service IDs of outdated scan results through the same contextual transaction
				subQuery := txScopeDb.
					Model(&scopedb.T_discovery_service{}).
					Select("id").
					Where("scan_cycle < ?", minCycleKeep)

				// Cleanup old sub module entries manually
				for _, subModule := range subModules {
					errScopeDb2 := txScopeDb.Where("id_t_discovery_service IN (?)", subQuery).Delete(subModule).Error
					if errScopeDb2 != nil {
						return errScopeDb2
					}
				}

				// Final cleanup of old discovery module entries. This will also include a cascade delete and
				// catch everything else that might have been forgotten
				for _, discoveryModule := range discoveryModules {
					errScopeDb3 := txScopeDb.Where("scan_cycle < ?", minCycleKeep).Delete(discoveryModule).Error
					if errScopeDb3 != nil {
						return errScopeDb3
					}
				}

				// Return nil as everything went fine
				return nil
			})

			// Log cleanup time
			logger.Infof("Cleanup of outdated scan cycles of '%s' (ID %d) took %s.",
				scopeEntry.Name,
				scopeEntry.Id,
				time.Since(start),
			)

			// Return error if transaction failed
			if errCleanup != nil {
				logger.Errorf(
					"Could not cleanup outdated scan cycles from scan scope '%s' (ID %d): %s",
					scopeEntry.Name,
					scopeEntry.Id,
					errCleanup,
				)
			}
		}()
	}

	// Return nil as everything went fine
	return nil
}

// XCreateView creates a view with given properties on scope database and manager database in a transactional manner
func XCreateView(
	scopeDb *gorm.DB,
	scopeEntry *database.T_scan_scope,
	viewName string,
	createdBy string,
	filters map[string][]string,
) (*database.T_scope_view, error) {

	var viewEntry *database.T_scope_view

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Create actual views in scope database
			viewTableNames, errViewTableNames := scopedb.CreateScopeView(txScopeDb, viewName, filters)
			if errViewTableNames != nil {
				return errViewTableNames
			}

			// Create view entry in manager database to keep track
			var errViewEntry error
			viewEntry, errViewEntry = transaction.CreateViewEntry(
				scopeEntry,
				viewName,
				createdBy,
				filters,
				viewTableNames,
			)
			if errViewEntry != nil {
				return errViewEntry
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return viewEntry, errTx
}

// XDeleteView deletes a view on scopeDb and manager database in a transactional manner. Access rights are removed along.
func XDeleteView(scopeDb *gorm.DB, viewEntry *database.T_scope_view) error {

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Delete view on manager database and scope database, including all associated access rights
			return xDeleteView(txScopeDb, transaction, viewEntry.ScanScope.DbName, viewEntry)
		})
	})

	// Return transaction result, which may be nil or an error
	return errTx
}

// XUpdateView updates a view on scopeDb and manager database in a transactional manner. If the view name changed, the prefix
// of scope database view table names has to be updated too.
func XUpdateView(scopeDb *gorm.DB, viewEntry *database.T_scope_view) error {

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Generate list of view table names from manager database
			viewTableNames := utils.SanitizeToSlice(viewEntry.ViewNames, ",")
			newViewTableNames := make([]string, 0, len(viewTableNames))

			// Iterate existing view tables in scope database to update their name
			for _, oldViewTableName := range viewTableNames {
				for _, view := range scopedb.ViewDefinitionNames() {
					if strings.HasSuffix(oldViewTableName, "_"+view) {

						// Define view name
						newViewTableName, errNewViewTableName := scopedb.SanitizeViewName(viewEntry.Name + "_" + view)
						if errNewViewTableName != nil {
							return fmt.Errorf("could not sanitize view name: %s", errNewViewTableName)
						}

						// Remember view name
						newViewTableNames = append(newViewTableNames, newViewTableName)

						// Update view name in scope database
						if oldViewTableName != newViewTableName {
							errDb := txScopeDb.Exec(`ALTER VIEW "` + oldViewTableName + `" RENAME TO "` + newViewTableName + `"`).Error
							if errDb != nil {
								return errDb
							}
						}

						// Update next view table name
						break
					}
				}
			}

			// Sort views alphabetically
			slices.Sort(newViewTableNames)

			// Update view table names
			viewEntry.ViewNames = strings.Join(newViewTableNames, ",")

			// Update entry in manager database
			errDb := transaction.UpdateViewEntry(viewEntry)
			if errDb != nil {
				return errDb
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return errTx
}

// XGrantToken generates random access credentials and activates them on the defined scope view. Such
// access token are not user bound, valid for a prolonged time frame and only valid on a single scope
// view. A random username and password (representing the random access token) are generated and returned
func XGrantToken(
	scopeDb *gorm.DB,
	viewEntry *database.T_scope_view,
	grantedBy string,
	expiry time.Duration,
	description string,
	dbConnectionsUser int,
) (string, string, error) {

	// Generate random username
	randomUsername, errRandomUsername := utils.GenerateToken(utils.AlphaNum, 15)
	if errRandomUsername != nil {
		return "", "", errRandomUsername
	}

	// Generate random password
	charSet := strings.Replace(utils.AlphaNumCaseSymbol, "?", "", -1) // Drop question mark from character set, as gorm has issues with them
	randomPassword, errRandomPassword := utils.GenerateToken(charSet, 25)
	if errRandomPassword != nil {
		return "", "", errRandomPassword
	}

	// Inject deterministic values during development mode. Otherwise, new access tokens would be generated with
	// each start, jamming the database server
	if _build.DevMode {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(description)))
		hash = strings.ToLower(hash)
		randomUsername = hash[len(hash)-15:] // last characters are without $ separators
		randomPassword = randomUsername
	}

	// Prepare random credentials set to be used
	credentials := scopedb.DbCredentials{
		Username: "token-" + randomUsername, // Prefix random access token username portion to cluster them in database user list
		Password: randomPassword,
	}

	// Calculate expiry time
	expiryTime := time.Now().Add(expiry)

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Grant actual rights in scope database
			errGrant := scopedb.GrantScopeView(txScopeDb, viewEntry, credentials, expiryTime, dbConnectionsUser)
			if errGrant != nil {
				return fmt.Errorf(
					"could not grant user '%s' on scope views '%s' (ID %d): %s",
					credentials.Username,
					viewEntry.Name,
					viewEntry.Id,
					errGrant,
				)
			}

			// Create grant entry in manager database to keep track
			_, errNew := transaction.CreateGrantEntry(
				viewEntry,
				false,
				credentials.Username,
				grantedBy,
				expiryTime,
				description,
			)
			if errNew != nil {
				return fmt.Errorf(
					"could not create grant entry for user '%s' for scope views '%s' (ID %d): %s",
					credentials.Username,
					viewEntry.Name,
					viewEntry.Id,
					errNew,
				)
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return credentials.Username, credentials.Password, errTx
}

// XGrantUsers grants view access rights for the given list of users by adding the rights to the scopeDb and adding the
// a grant entry to manager database in a transactional manner. List of users must be a DbCredentials struct, because the user
//
//	might need to be created on the database server if not existing yet.
func XGrantUsers(
	scopeDb *gorm.DB,
	viewEntry *database.T_scope_view,
	grantedBy string,
	dbCredentials []scopedb.DbCredentials,
	dbConnectionsUser int,
) error {

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Query all users to give access
			for _, credentials := range dbCredentials {

				// Grant actual rights in scope database
				errGrant := scopedb.GrantScopeView(txScopeDb, viewEntry, credentials, time.Now(), dbConnectionsUser)
				if errGrant != nil {
					return fmt.Errorf(
						"could not grant user '%s' on scope views '%s' (ID %d): %s",
						credentials.Username,
						viewEntry.Name,
						viewEntry.Id,
						errGrant,
					)
				}

				// Create grant entry in manager database to keep track
				_, errNew := transaction.CreateGrantEntry(
					viewEntry,
					true,
					credentials.Username,
					grantedBy,
					time.Time{},
					"",
				)
				if errNew != nil {
					return fmt.Errorf(
						"could not create grant entry for user '%s' for scope views '%s' (ID %d): %s",
						credentials.Username,
						viewEntry.Name,
						viewEntry.Id,
						errNew,
					)
				}
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return errTx
}

// XRevokeGrants revokes view access rights for the given list of users by removing the rights from scopeDb and the
// grant entry from manager database in a transactional manner
func XRevokeGrants(scopeDb *gorm.DB, viewEntry *database.T_scope_view, usernames []string) error {

	// Start transaction on the scope database and manager database. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := database.WithTransaction(func(transaction *database.Transaction) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Generate list of view table names from manager database
			viewTableNames := utils.SanitizeToSlice(viewEntry.ViewNames, ",")

			// Iterate view grants to identify removable ones
			for _, viewGrant := range viewEntry.Grants {

				// Identify deletable grant
				if scanUtils.StrContained(viewGrant.Username, usernames) {

					// Revoke view grant
					errRevoke := xRevokeGrant(
						txScopeDb, transaction, viewEntry.ScanScope.DbName, &viewGrant, viewTableNames)
					if errRevoke != nil {
						return errRevoke
					}
				}
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return errTx
}

// XDeploySampleData applies a default configuration for development purposes to the manager database and some sample data
// to the scope database
func XDeploySampleData(logger scanUtils.Logger, settingsEntryDefault database.T_scan_setting) error {

	// Define sample data struct
	type sampleScopesAndViews struct {
		scopeGroupId          uint64
		scopeName             string
		scopeDbName           string
		scopeSecret           string
		scopeCreatedBy        string
		scopeCycles           bool
		scopeMaxInstances     map[string]uint32
		scopeType             string
		scopeAttributes       utils.JsonMap // Arbitrary JSON data with further scope details that might be needed to decide how to populate, import, refresh, synchronize scan inputs... This may be very specific to the deployed environment.
		viewName              string
		viewFilter            map[string][]string
		viewUsers             []scopedb.DbCredentials
		viewTokenDescriptions []string
	}

	// Initialize sample data
	samples := []sampleScopesAndViews{
		{
			scopeGroupId:      1,
			scopeName:         "Dev scope",
			scopeDbName:       "dev_scope",
			scopeSecret:       "dev_secret",
			scopeCreatedBy:    "user1@domain.tld",
			scopeCycles:       false,
			scopeMaxInstances: devScopeMaxInstances,
			scopeType:         "custom",
			scopeAttributes:   utils.JsonMap{},
			viewName:          "All",
			viewFilter:        map[string][]string{},
			viewUsers: []scopedb.DbCredentials{
				{Username: "user1@domain.tld", Password: ""},
				{Username: "name1@own.tld", Password: ""},
			},
			viewTokenDescriptions: []string{
				"Dev scope token 1",
				"Dev scope token 2",
				"Dev scope token 3",
			},
		},
		{
			scopeGroupId:      2,
			scopeName:         "Sample From Network Inventory",
			scopeDbName:       "sample1",
			scopeSecret:       "irgvolj2vf3yd94jbpyogdptxp2ty9b9009idp8lde2q6zt3neuxtj09r0mxzh16",
			scopeCreatedBy:    "user1@domain.tld",
			scopeCycles:       false,
			scopeMaxInstances: devScopeMaxInstances,
			scopeType:         "networks",
			scopeAttributes: utils.JsonMap{ // Just a sample how a remote repository could be defined
				"sync":                   true, // Remote repositories could be imported once, or kept in sync by another component
				"asset_companies":        []string{},
				"asset_departments":      []string{},
				"asset_zones":            []string{},
				"asset_purposes":         []string{},
				"asset_routing_domains":  []string{},
				"asset_countries":        []string{"DE", "AT"},
				"asset_locations":        []string{"MCH"},
				"asset_contacts":         []string{},
				"asset_exclude_keywords": []string{"disabled", "sensitive"},
			},
			viewName:              "View 101112",
			viewFilter:            map[string][]string{"input_company": {"Company Inc."}},
			viewUsers:             []scopedb.DbCredentials{},
			viewTokenDescriptions: []string{},
		},
		{
			scopeGroupId:      2,
			scopeName:         "Sample From Asset Inventory",
			scopeDbName:       "sample2",
			scopeSecret:       "ry6g6zxkkaxm6vwi7r1tjc7nofbju9nh4c7iav6nhnrscbsbgyoxnab48tetkbl8",
			scopeCreatedBy:    "user1@domain.tld",
			scopeCycles:       false,
			scopeMaxInstances: devScopeMaxInstances,
			scopeType:         "assets",
			scopeAttributes: utils.JsonMap{ // Just a sample how a remote repository could be defined
				"sync":              false, // Remote repositories could be imported once, or kept in sync by another component
				"asset_type":        "Client",
				"asset_countries":   []string{"DE", "AT"},
				"asset_locations":   []string{"MCH"},
				"asset_companies":   []string{"Dev Corp.", "Operations Corp."},
				"asset_departments": []string{"IT", "CYS"},
				"asset_contacts":    []string{"Mike", "Andy"},
				"asset_critical":    "No",
			},
			viewName:   "View 789",
			viewFilter: map[string][]string{"input_country": {"de"}, "input_location": {"munich"}},
			viewUsers:  []scopedb.DbCredentials{},
		},
		{
			scopeGroupId:      1,
			scopeName:         "Another scope",
			scopeDbName:       "sample3",
			scopeSecret:       "t3og5czedblsw2px39qlum1t1k26p4cpczzvs0cl9fbxe1cd9i20vvhtm5x9bm3x",
			scopeCreatedBy:    "user1@domain.tld",
			scopeCycles:       true,
			scopeMaxInstances: devScopeMaxInstances,
			scopeType:         "custom",
			scopeAttributes:   utils.JsonMap{},
			viewName:          "View 456",
			viewFilter:        map[string][]string{"input_country": {"de", "at"}},
			viewUsers: []scopedb.DbCredentials{
				{Username: "user2@domain.tld", Password: ""},
				{Username: "user3@domain.tld", Password: ""},
			},
			viewTokenDescriptions: []string{},
		},
		{
			scopeGroupId:      1,
			scopeName:         "Dummy Scope",
			scopeDbName:       "sample4",
			scopeSecret:       "saornsio9sgy6x4jrsj2p7c719q4y7mvjvrybrmcmmt47z8kk2nu8b0ii0yldrw3",
			scopeCreatedBy:    "user1@domain.tld",
			scopeCycles:       true,
			scopeMaxInstances: devScopeMaxInstances,
			scopeType:         "custom",
			scopeAttributes:   utils.JsonMap{},
			viewName:          "View 123",
			viewFilter:        map[string][]string{"input_country": {"de"}},
			viewUsers: []scopedb.DbCredentials{
				{Username: "name1@own.tld", Password: ""},
			},
			viewTokenDescriptions: []string{
				"Dummy scope token 1",
			},
		},
	}

	// Get or create development database server entry
	dbServerEntry, errDbServerEntry := database.GetDatabaseEntryByName(devDbServerName)
	if errDbServerEntry != nil {
		return fmt.Errorf("could not query database server entry '%s': %s", devDbServerName, errDbServerEntry)
	}

	// Check if database server exists
	if dbServerEntry == nil {

		// Log creation
		logger.Debugf("Creating development database server.")

		// Create database server entry
		dbServerEntry, errDbServerEntry = database.CreateDatabaseEntry(
			devDbServerName,
			"postgres",
			"localhost",
			5432,
			devDbServerAdmin,
			devDbServerPassword,
			"localhost",
			"sslmode=disable", // Disabled SSL mode ONLY for local development DB!
		)
		if errDbServerEntry != nil {
			return fmt.Errorf("could not create database server entry: %s", errDbServerEntry)
		}
	}

	// Open connection to database server
	serverDb, errServerDb := scopedb.GetServerDbHandle(logger, dbServerEntry)
	if errServerDb != nil {
		return fmt.Errorf("could not open database server: %s", errServerDb)
	}

	// Prepare databases with sample data
	for i, sample := range samples {

		// Log step
		logger.Infof("Deploying sample scope data %d.", i)

		// Check if sample entry exists
		scopeEntry, errScopeEntry := database.GetScopeEntryBySecret(sample.scopeSecret)
		if errScopeEntry != nil {
			logger.Warningf("could not query scan scope entry: %s", errScopeEntry)
		}

		// Create sample scope database in scope database and manager database. Recreate in manager database if it was already existing in scope database.
		if scopeEntry == nil {

			// Create scan scope
			logger.Debugf("Creating scan scope.")
			var errScopeEntry error
			scopeEntry, errScopeEntry = XCreateScope(
				serverDb,
				dbServerEntry,
				sample.scopeName,
				sample.scopeDbName,
				sample.scopeGroupId,
				sample.scopeCreatedBy,
				sample.scopeSecret,
				sample.scopeType,
				sample.scopeCycles,
				-1,
				utils.NetworkSizeSplit,
				sample.scopeAttributes,
				settingsEntryDefault,
			)
			if scopeEntry == nil {
				return fmt.Errorf("could not create development scan scope: %s", errScopeEntry)
			}

			// Set some scan stats
			scopeEntry.CycleDone = float64(rand.Intn(82-10) + 10)
			scopeEntry.CycleActive = float64(rand.Intn(8-2) + 2)
			scopeEntry.CycleFailed = float64(rand.Intn(10-0) + 0)
			_, errSave := scopeEntry.Save("cycle_done", "cycle_active", "cycle_failed")
			if errSave != nil {
				return fmt.Errorf("could not update scan scope stats")
			}

			// Prepare sample scan agent active instances
			tasks := utils.JsonMap(map[string]interface{}{
				"Discovery": 5, "Banner": 10, "Nfs": 0, "Nuclei": 0, "Smb": 1, "Ssh": 2, "Ssl": 2, "Webcrawler": 4, "Webenum": 3,
			})

			// Insert some scan agent data
			agentEntries := []database.T_scan_agent{
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Koalaarrow Elijah",
					Host:         "md332dg",
					Ip:           "123.45.67.8",
					LastSeen:     time.Now(),
					Tasks:        tasks,
					CpuCores:     8,
					CpuMhz:       3600,
					CpuRate:      80,
					MemoryBytes:  32 * 1024 * 1024 * 1024,
					MemoryRate:   63,

					Shared: false,
					Limits: false,

					VersionNmap:   "7.95",
					VersionNpcap:  "1.79",
					VersionSslyze: "6.1.0",

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Monkeybitter Lizbeth",
					Host:         "mdf4ht3c",
					Ip:           "123.45.67.9",
					LastSeen:     time.Now().Add(-time.Minute * 10),
					Tasks:        tasks,
					CpuCores:     4,
					CpuMhz:       2800,
					CpuRate:      62,
					MemoryBytes:  16 * 1024 * 1024 * 1024,
					MemoryRate:   77,

					Shared: false,
					Limits: false,

					VersionNmap:   "7.95",
					VersionNpcap:  "1.79",
					VersionSslyze: "6.1.0",
					VersionPython: "3.13.2",

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Ocelotspring Abigail",
					Host:         "md5dfgss",
					Ip:           "123.45.67.10",
					LastSeen:     time.Now().Add(-time.Hour),
					Tasks:        tasks,
					CpuCores:     16,
					CpuMhz:       2400,
					CpuRate:      66,
					MemoryBytes:  64 * 1024 * 1024 * 1024,
					MemoryRate:   54,

					Shared: false,
					Limits: true,

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Scowldot Jacob",
					Host:         "md1g3ggg",
					Ip:           "123.45.67.11",
					LastSeen:     time.Now().Add(-time.Hour * 12),
					Tasks:        tasks,
					CpuCores:     8,
					CpuMhz:       3200,
					CpuRate:      81,
					MemoryBytes:  32 * 1024 * 1024 * 1024,
					MemoryRate:   61,

					Shared: true,
					Limits: true,

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Glasspeat Aubrey",
					Host:         "md5ddsa34",
					Ip:           "123.45.67.12",
					LastSeen:     time.Now().Add(-time.Hour * 24 * 7),
					Tasks:        tasks,
					CpuCores:     2,
					CpuMhz:       2200,
					CpuRate:      98,
					MemoryBytes:  8 * 1024 * 1024 * 1024,
					MemoryRate:   95,

					Shared: true,
					Limits: false,

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Foxdestiny David",
					Host:         "md1f54fg",
					Ip:           "123.45.67.13",
					LastSeen:     time.Now().Add(-time.Hour * 24 * 14),
					Tasks:        tasks,
					CpuCores:     4,
					CpuMhz:       2600,
					CpuRate:      12,
					MemoryBytes:  16 * 1024 * 1024 * 1024,
					MemoryRate:   12,

					Shared: true,
					Limits: false,

					Platform:        "Microsoft Windows 10 Enterprise",
					PlatformFamily:  "Standalone Workstation",
					PlatformVersion: "10.0.19045.5131 Build 19045.5131",

					BuildCommit:    "3f7f7785fc900f7f09eada4cfea25d9b31910cdc",
					BuildTimestamp: "2025-05-26T14:50:00+02:00",
					ApiVersion:     "1.1.0",
				},
			}
			errAgents := database.CreateAgentEntries([]database.T_scan_agent{
				agentEntries[rand.Intn(len(agentEntries))],
				agentEntries[rand.Intn(len(agentEntries))],
				agentEntries[rand.Intn(len(agentEntries))],
			})
			if errAgents != nil {
				return fmt.Errorf("could not create sample scan agents: %s", errAgents)
			}
		}

		// Update sample scope's max instances
		logger.Debugf("Updating sample scan scope settings.")
		errUpdate := database.UpdateScopeInstances(scopeEntry.Id, sample.scopeMaxInstances)
		if errUpdate != nil {
			return fmt.Errorf("could not update scan scope settings: %s", errUpdate)
		}

		// Open scope database itself
		logger.Debugf("Opening scope database.")
		scopeDb, errScopeDb := scopedb.GetScopeDbHandle(logger, scopeEntry)
		var netConnectErr *net.OpError
		var pgConnectErr *pgconn.ConnectError
		if errors.As(errScopeDb, &netConnectErr) {
			return fmt.Errorf("could not connect to scope database '%s'", scopeEntry.DbServer.Name)
		} else if errors.As(errScopeDb, &pgConnectErr) {
			return fmt.Errorf("could not authenticate on scope database '%s'", scopeEntry.DbServer.Name)
		} else if errScopeDb != nil {

			// Re-create sample database (might be missing)
			errCreate := scopedb.CreateScopeDb(serverDb, sample.scopeDbName, sample.scopeName)
			if errCreate != nil {
				return fmt.Errorf("could not re-create development scope: %s", errCreate)
			}

			// Retry connection
			scopeDb, errScopeDb = scopedb.GetScopeDbHandle(logger, scopeEntry)
			if errScopeDb != nil {
				return fmt.Errorf("could not open scope database: %s", errScopeDb)
			}
		}

		// Create sample scope database tables by running automigrate
		logger.Debugf("Automigrating scope database.")
		errMigrate := scopedb.AutoMigrateScopeDb(scopeDb)
		if errMigrate != nil {
			logger.Warningf("Could not automigrate scope databases: %s", errMigrate)
		}

		// Install trigram indexes
		logger.Debugf("Installing trigram full-text search indices in scope database.")
		errExtension := scopedb.InstallTrigramIndices(scopeDb)
		if errExtension != nil {
			logger.Errorf(
				"Could not install database extensions for scan scope '%s' (ID %d)",
				scopeEntry.Name,
				scopeEntry.Id,
			)
			return errExtension
		}

		// Create sample set of scope views
		exists, errExists := database.ViewExists(scopeEntry.Id, sample.viewName)
		if errExists != nil {
			logger.Errorf(
				"Could not check existence of scope view '%s' of scan scope '%s'.",
				sample.viewName,
				scopeEntry.Name,
			)
			return errExists
		}

		if !exists {
			logger.Debugf("Creating scan scope view.")
			scopeView, errScopeView := XCreateView(scopeDb, scopeEntry, sample.viewName, sample.scopeCreatedBy, sample.viewFilter)
			if errScopeView != nil {
				logger.Warningf("Could not create scan scope view: %s", errScopeView)
			} else {

				// Create sample view grants
				logger.Debugf("Creating scan scope view grants.")
				errGrant := XGrantUsers(scopeDb, scopeView, sample.scopeCreatedBy, sample.viewUsers, 20)
				if errGrant != nil {
					logger.Warningf("Could not create sample scan scope view grant: %s", errGrant)
				}

				// Create sample view tokens
				logger.Debugf("Creating view access token.")
				for _, viewTokenDescription := range sample.viewTokenDescriptions {
					daysToExpire := []time.Duration{1, 4, 7, 14, 30, 60, 90, 180, 365}
					_, _, errToken := XGrantToken(
						scopeDb,
						scopeView,
						sample.scopeCreatedBy,
						time.Hour*24*daysToExpire[rand.Int()%len(daysToExpire)],
						viewTokenDescription,
						20,
					)
					if errToken != nil {
						logger.Warningf("Could not create sample view access token: %s", errToken)
					}
				}
			}
		} else {
			logger.Debugf("Scan scope view already existing.")
		}
	}

	// Close all scope database handles again because most are no longer needed after deploying sample data
	for _, errClose := range scopedb.CloseScopeDbs() {
		logger.Warningf("Could not close sample-data scope database handle: %s", errClose)
	}

	// Return nil as everything went fine
	return nil
}

// xDeleteView takes care of
//   - removing associated access rights
//   - removing view entry from the manager database
//   - removing view tables from the scope database
//
// ATTENTION: The changes must be committed/rolled back by the calling function!
func xDeleteView(
	txScopeDb *gorm.DB,
	transaction *database.Transaction,
	dbName string, // Required to also revoke DB connect right if necessary
	viewEntry *database.T_scope_view,
) error {

	// Generate list of view table names from manager database
	viewTableNames := utils.SanitizeToSlice(viewEntry.ViewNames, ",")

	// Cleanup grants
	for _, viewGrant := range viewEntry.Grants {
		errRevoke := xRevokeGrant(txScopeDb, transaction, dbName, &viewGrant, viewTableNames)
		if errRevoke != nil {
			return errRevoke
		}
	}

	// Remove view entry from manager database
	errDelete := transaction.DeleteViewEntry(viewEntry)
	if errDelete != nil {
		return fmt.Errorf(
			"could not remove view entry '%s' (ID %d) from scan scope '%s' (ID %d): %s",
			viewEntry.Name,
			viewEntry.Id,
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errDelete,
		)
	}

	// Remove view tables from scope database
	errDelete2 := scopedb.DeleteScopeView(txScopeDb, viewTableNames...)
	if errDelete2 != nil {
		return fmt.Errorf(
			"could not remove view tables from scope database '%s' (ID %d): %s",
			viewEntry.ScanScope.Name,
			viewEntry.ScanScope.Id,
			errDelete2,
		)
	}

	// Return nil as everything went fine
	return nil
}

// xRevokeGrant takes care of
//   - removing grant entry from the manager database
//   - removing associated access rights from the scope database
//   - cleaning obsolete credentials from associated database server
//
// ATTENTION: The changes must be committed/rolled back by the calling function!
func xRevokeGrant(
	txScopeDb *gorm.DB,
	transaction *database.Transaction,
	dbName string, // Required to also revoke DB connect right if necessary
	grantEntry *database.T_view_grant,
	viewTableNames []string,
) error {

	// Remove grant entry from manager database
	errRevoke1 := transaction.DeleteGrantEntry(grantEntry)
	if errRevoke1 != nil {
		return fmt.Errorf(
			"could not remove grant entry '%s' from scope view '%s' (ID %d): %s",
			grantEntry.Username,
			grantEntry.ScopeView.Name,
			grantEntry.ScopeView.Id,
			errRevoke1,
		)
	}

	// Remove access rights from scope database
	errRevoke2 := scopedb.RevokeScopeView(txScopeDb, dbName, grantEntry.Username, viewTableNames)
	if errRevoke2 != nil {
		return fmt.Errorf(
			"could not revoke grants of '%s' from scope database view '%s' (ID %d): %s",
			grantEntry.Username,
			grantEntry.ScopeView.Name,
			grantEntry.ScopeView.Id,
			errRevoke2,
		)
	}

	// Try to delete user. It will succeed if the user has no other privileges on the same database server
	scopedb.DeleteDatabaseCredentials(txScopeDb, grantEntry.Username)

	// Return nil as everything went fine
	return nil
}
