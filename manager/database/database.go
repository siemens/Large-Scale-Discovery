/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"gorm.io/gorm"
	"large-scale-discovery/_build"
	"large-scale-discovery/utils"
	"math/rand"
	"strings"
	"time"
)

// Some development values
var (
	devDbServerName      = "Local development DB" // Name of the db server entry
	devDbServerAdmin     = "postgres"
	devDbServerPassword  = "test123!$LSD"
	devScopeMaxInstances = map[string]uint32{
		discovery.Label:  15,
		banner.Label:     10,
		nfs.Label:        10,
		smb.Label:        10,
		ssh.Label:        10,
		ssl.Label:        10,
		webcrawler.Label: 10,
		webenum.Label:    10,
	}
)

// XDeploySampleData applies a default configuration for development purposes to the manager db and some sample data
// to the scope db
func XDeploySampleData(logger scanUtils.Logger, scanDefaults T_scan_settings) error {

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
		viewUsers             []DbCredentials
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
			viewUsers: []DbCredentials{
				{"user1@domain.tld", ""},
				{"name1@own.tld", ""},
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
			viewUsers:             []DbCredentials{},
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
			viewUsers:  []DbCredentials{},
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
			viewUsers: []DbCredentials{
				{"user2@domain.tld", ""},
				{"user3@domain.tld", ""},
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
			viewUsers: []DbCredentials{
				{"name1@own.tld", ""},
			},
			viewTokenDescriptions: []string{
				"Dummy scope token 1",
			},
		},
	}

	// Get or create development db server entry
	serverEntry, errServerEntry := getServerEntryByName(devDbServerName)
	if errServerEntry != nil && errors.Is(errServerEntry, gorm.ErrRecordNotFound) { // Check if entry didn't exist

		// Log creation
		logger.Debugf("Creating development database server.")

		// Create db server entry
		serverEntry, errServerEntry = createServerEntry(
			managerDb,
			devDbServerName,
			"postgres",
			"localhost",
			5432,
			devDbServerAdmin,
			devDbServerPassword,
			"localhost",
			"sslmode=disable", // Disabled SSL mode ONLY for local development DB!
		)
		if errServerEntry != nil {
			return fmt.Errorf("could not create database server entry: %s", errServerEntry)
		}
	} else if errServerEntry != nil {
		return fmt.Errorf("could not retrieve database server entry: %s", errServerEntry)
	}

	// Open connection to database server
	serverDb, errHandle := GetServerDbHandle(logger, serverEntry)
	if errHandle != nil {
		return fmt.Errorf("could not open database server: %s", errHandle)
	}

	// Prepare databases with sample data
	for i, sample := range samples {

		// Log step
		logger.Infof("Deploying sample scope data %d.", i)

		// Check if sample entry exists
		scopeEntry, errScopeEntry := GetScopeEntryBySecret(sample.scopeSecret)
		if errScopeEntry != nil && !errors.Is(errScopeEntry, gorm.ErrRecordNotFound) { // Check if entry didn't exist
			logger.Warningf("could not get scan scope entry: %s", errScopeEntry)
		}

		// Create sample scope database in scope db and manager db. Recreate in manager db if it was already existing in scope db.
		if scopeEntry == nil {

			// Create scan scope
			logger.Debugf("Creating scan scope.")
			var errScanScope error
			scopeEntry, errScanScope = XCreateScope(
				serverDb,
				serverEntry,
				sample.scopeName,
				sample.scopeDbName,
				sample.scopeGroupId,
				sample.scopeCreatedBy,
				sample.scopeSecret,
				sample.scopeType,
				sample.scopeCycles,
				-1,
				sample.scopeAttributes,
				scanDefaults,
			)
			if scopeEntry == nil {
				return fmt.Errorf("could not create development scan scope: %s", errScanScope)
			}

			// Initialize random number generation
			rand.Seed(time.Now().UnixNano())

			// Set some scan stats
			scopeEntry.CycleDone = float64(rand.Intn(82-10) + 10)
			scopeEntry.CycleActive = float64(rand.Intn(8-2) + 2)
			scopeEntry.CycleFailed = float64(rand.Intn(10-0) + 0)
			_, errSave := scopeEntry.Save("cycle_done", "cycle_active", "cycle_failed")
			if errSave != nil {
				return fmt.Errorf("could not update scan scope stats")
			}

			// Prepare sample scan agent active tasks
			tasks := utils.JsonMap(map[string]interface{}{
				"Discovery": 5, "Banner": 10, "Nfs": 0, "Smb": 1, "Ssh": 2, "Ssl": 2, "Webcrawler": 4, "Webenum": 3,
			})

			// Insert some scan agent data
			scanAgents := []T_scan_agent{
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Koalaarrow Elijah",
					Host:         "md332dg",
					Ip:           "123.45.67.8",
					LastSeen:     time.Now(),
					Tasks:        tasks,
					CpuRate:      80,
					MemoryRate:   63,
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Monkeybitter Lizbeth",
					Host:         "mdf4ht3c",
					Ip:           "123.45.67.9",
					LastSeen:     time.Now().Add(-time.Minute * 10),
					Tasks:        tasks,
					CpuRate:      62,
					MemoryRate:   77,
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Ocelotspring Abigail",
					Host:         "md5dfgss",
					Ip:           "123.45.67.10",
					LastSeen:     time.Now().Add(-time.Hour),
					Tasks:        tasks,
					CpuRate:      66,
					MemoryRate:   54,
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Scowldot Jacob",
					Host:         "md1g3ggg",
					Ip:           "123.45.67.11",
					LastSeen:     time.Now().Add(-time.Hour * 12),
					Tasks:        tasks,
					CpuRate:      81,
					MemoryRate:   61,
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Glasspeat Aubrey",
					Host:         "md5ddsa34",
					Ip:           "123.45.67.12",
					LastSeen:     time.Now().Add(-time.Hour * 24 * 7),
					Tasks:        tasks,
					CpuRate:      98,
					MemoryRate:   95,
				},
				{
					IdTScanScope: scopeEntry.Id,
					Name:         "Foxdestiny David",
					Host:         "md1f54fg",
					Ip:           "123.45.67.12",
					LastSeen:     time.Now().Add(-time.Hour * 24 * 14),
					Tasks:        tasks,
					CpuRate:      12,
					MemoryRate:   12,
				},
			}
			managerDb.Create(&[]T_scan_agent{
				scanAgents[rand.Intn(len(scanAgents))],
				scanAgents[rand.Intn(len(scanAgents))],
				scanAgents[rand.Intn(len(scanAgents))],
			})
		}

		// Update sample scope's max instances
		logger.Debugf("Updating sample scan scope settings.")
		errUpdate := updateScopeInstances(managerDb, scopeEntry.Id, sample.scopeMaxInstances)
		if errUpdate != nil {
			return fmt.Errorf("could not update scan scope settings: %s", errUpdate)
		}

		// Open scope database itself
		logger.Debugf("Opening scope database.")
		scopeDb, errScopeHandle := GetScopeDbHandle(logger, scopeEntry)
		if errScopeHandle != nil {

			// Re-create sample database (might be missing)
			errCreate := createScopeDb(serverDb, sample.scopeDbName, sample.scopeName)
			if errCreate != nil {
				return fmt.Errorf("could not re-create development scope: %s", errCreate)
			}

			// Retry connection
			scopeDb, errScopeHandle = GetScopeDbHandle(logger, scopeEntry)
			if errScopeHandle != nil {
				return fmt.Errorf("could not open scope database: %s", errScopeHandle)
			}
		}

		// Create sample scope database tables by running automigrate
		logger.Debugf("Automigrating scope database.")
		errMigrate := AutoMigrateScopeDb(scopeDb)
		if errMigrate != nil {
			logger.Warningf("Could not automigrate scope databases: %s", errMigrate)
		}

		// Install trigram indexes
		logger.Debugf("Installing trigram full-text search indices in scope database.")
		errExtension := InstallTrigramIndices(scopeDb)
		if errExtension != nil {
			logger.Errorf(
				"Could not install database extensions for scan scope '%s' (ID %d)",
				scopeEntry.Name,
				scopeEntry.Id,
			)
			return errExtension
		}

		// Create sample set of scope views
		exists, errExists := ViewExists(scopeEntry.Id, sample.viewName)
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

	// Close all scope db handles again, after sample data was set, most might not be needed anymore
	CloseScopeDbs()

	// Return nil as everything went fine
	return nil
}

// XCreateScope creates a new scope entry in the manager db and the associated scan database in the scope db.
// Unfortunately, this step cannot happen in a transactional manner.
func XCreateScope(
	serverDb *gorm.DB, // ATTENTION: Cannot be a transaction!! Must be a plain connection, database cannot be created within a transaction.
	dbServer *T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	attributes utils.JsonMap,
	scanSettings T_scan_settings,
) (*T_scan_scope, error) {

	// Prepare scan scope memory
	var scopeEntry *T_scan_scope

	// Create scope entry in manager db to keep track
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		var errEntry error
		scopeEntry, errEntry = createScopeEntry(
			txManagerDb,
			dbServer,
			name,
			dbName,
			groupId,
			createdBy,
			secret,
			scopeType,
			cycles,
			cyclesRetention,
			attributes,
			0,
			scanSettings,
		)
		return errEntry
	})
	if errTx != nil {
		return nil, errTx
	}

	// Create database in scope db
	errCreate := createScopeDb(serverDb, dbName, name)
	if errCreate != nil {
		return scopeEntry, errCreate // Return scan scope entry, nevertheless, because it got created despite this error.
	}

	// Return newly created scan scope entry
	return scopeEntry, nil
}

// XDeleteScope deletes scope with given scopeId on scope db and manager db in a transactional manner. Views and access
// rights are removed along.
func XDeleteScope(serverDb *gorm.DB, scopeDb *gorm.DB, scanScope *T_scan_scope) error {

	// Takes care of:
	// 		- removing associated scope views and access rights
	//		- removing scope database from scope db
	//		- removing scope entry from the manager db
	//		- removing associated agent entries from the manager db
	//		- removing associated settings entry from the manager db

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTxManagerDb := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		errTxScopeDb := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Cleanup views
			for _, scopeView := range scanScope.ScopeViews {
				errRevoke := xDeleteView(txScopeDb, txManagerDb, scanScope.DbName, &scopeView)
				if errRevoke != nil {
					return errRevoke
				}
			}

			// Remove agent entries from manager db
			for _, scanAgent := range scanScope.ScanAgents {
				errDelete := txManagerDb.Delete(&scanAgent).Error
				if errDelete != nil {
					return fmt.Errorf(
						"could not remove agent entry '%s' (ID %d) of scan scope '%s' (ID %d): %s",
						scanAgent.Name,
						scanAgent.Id,
						scanScope.Name,
						scanScope.Id,
						errDelete,
					)
				}
			}

			// Remove settings entry from manager db
			errDelete2 := txManagerDb.Delete(&scanScope.ScanSettings).Error
			if errDelete2 != nil {
				return fmt.Errorf(
					"could not remove settings entry of scan scope '%s' (ID %d): %s",
					scanScope.Name,
					scanScope.Id,
					errDelete2,
				)
			}

			// Remove scope entry from manager db
			errDelete3 := txManagerDb.Delete(scanScope).Error
			if errDelete3 != nil {
				return fmt.Errorf(
					"could not remove scope entry '%s' (ID %d): %s",
					scanScope.Name,
					scanScope.Id,
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

		// Delete scope's database (can only be done outside of a transaction)
		// ATTENTION: Scope database connections will be killed and have to be committed already!
		// ATTENTION: Last step, it cannot be rolled back!
		errDelete4 := deleteScopeDb(serverDb, scanScope.DbName)
		if errDb, ok := errDelete4.(*pq.Error); ok && errDb.Code.Name() == "invalid_catalog_name" {
			// Scope database not existing anymore
		} else if errDelete4 != nil {
			return fmt.Errorf(
				"could not remove scope database '%s' (ID %d): %s",
				scanScope.Name,
				scanScope.Id,
				errDelete4,
			)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return transaction result, which may be nil or an error
	return errTxManagerDb
}

// XCycleScope initializes a new scan cycle by resetting the scan input states in the scope db and updating the
// scan progress attributes in the manager db. If a scan cycle retention is configured, this will also delete
// outdated scan results.
func XCycleScope(logger scanUtils.Logger, scopeDb *gorm.DB, scanScope *T_scan_scope) error {

	// Start transaction on the scoped db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// Execute cycle reset by resetting all scan input states
		errScopeDb := txScopeDb.Model(&T_discovery{}).
			Where("1=1"). // Gorm doesn't allow global update without where clause
			Updates(map[string]interface{}{
				"scan_started":  sql.NullTime{},
				"scan_finished": sql.NullTime{},
				"scan_status":   scanUtils.StatusWaiting,
			}).Error
		if errScopeDb != nil {
			return fmt.Errorf("could not initialize new scan cycle in scope db: %s", errScopeDb)
		}

		// Update scope attributes in manager db
		scanScope.Cycle += 1
		scanScope.CycleStarted = time.Now()
		scanScope.CycleDone = 0
		scanScope.CycleActive = 0
		scanScope.CycleFailed = 0

		// Save updated attributes
		_, errManagerDb := scanScope.Save(
			"cycle",
			"cycle_started",
			"cycle_done",
			"cycle_active",
			"cycle_failed",
		)
		if errManagerDb != nil {
			return fmt.Errorf("could not initialize new scan cycle in manager db: %s", errManagerDb)
		}

		// Return nil as everything went fine
		return nil
	})

	// Return error if transaction failed
	if errTx != nil {
		return errTx
	}

	// Cleanup old scan results if desired
	if scanScope.CyclesRetention >= 1 {

		// Calculate old scan cycles to be cleaned up. +1 because cycle count just got incremented above.
		minCycleKeep := int(scanScope.Cycle) - scanScope.CyclesRetention

		// Run cleanup asynchronously in the background, so that the other components can continue with
		// the new scan cycle without larger delays. The cleanup query may take a very long time if there
		// are a lot of scan results that need to be cascade-deleted. If cleanup fails, it can be retried
		// next time.
		go func() {

			// Log background activity
			logger.Infof("Cleaning up outdated scan cycles of scan scope '%s' (ID %d).",
				scanScope.Name,
				scanScope.Id,
			)

			// Remember start time
			start := time.Now()

			// Prepare list of tables and order, subsequent queries should be applied against
			subModules := []interface{}{
				T_banner{},
				T_nfs_file{}, // files before info entry
				T_nfs{},
				T_smb_file{}, // files before info entry
				T_smb{},
				T_ssh{},
				T_ssl_issue{},       // issues before info entry
				T_ssl_cipher{},      // ciphers before info entry
				T_ssl_certificate{}, // certificates before info entry
				T_ssl{},
				T_webcrawler_page{},  // pages before vhost and info entry
				T_webcrawler_vhost{}, // vhost before info entry
				T_webcrawler{},
				T_webenum_results{}, // results before info entry
				T_webenum{},
			}
			discoveryModules := []interface{}{
				T_discovery_script{},
				T_discovery_service{},
				T_discovery_host{},
			}

			// Get t_discovery_service IDs of outdated scan results
			subQuery := scopeDb.
				Model(&T_discovery_service{}).
				Select("id").
				Where("scan_cycle < ?", minCycleKeep)

			// Prepare transaction for cleanup execution
			// To speed up cleanup, large data tables are cleaned using a list of IDs, before relying on
			// cascade deletes. Cascade deletes on large sets of data takes really long.
			errTx2 := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

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
				scanScope.Name,
				scanScope.Id,
				time.Since(start),
			)

			// Return error if transaction failed
			if errTx2 != nil {
				logger.Errorf(
					"Could not cleanup outdated scan cycles from scan scope '%s' (ID %d): %s",
					scanScope.Name,
					scanScope.Id,
					errTx2,
				)
			}
		}()
	}

	// Return nil as everything went fine
	return nil
}

// XCreateView creates a view with given properties on scope db and manager db in a transactional manner
func XCreateView(
	scopeDb *gorm.DB,
	scopeEntry *T_scan_scope,
	viewName string,
	createdBy string,
	filters map[string][]string,
) (*T_scope_view, error) {

	var viewEntry *T_scope_view

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Create actual views in scope db
			viewTableNames, errNewViews := createScopeView(txScopeDb, viewName, filters)
			if errNewViews != nil {
				return errNewViews
			}

			// Create view entry in manager db to keep track
			var errNew error
			viewEntry, errNew = createViewEntry(
				txManagerDb,
				scopeEntry,
				viewName,
				createdBy,
				filters,
				viewTableNames,
			)
			if errNew != nil {
				return errNew
			}

			// Return nil as everything went fine
			return nil
		})
	})

	// Return transaction result, which may be nil or an error
	return viewEntry, errTx
}

// XDeleteView deletes a view on scopeDb and manager db in a transactional manner. Access rights are removed along.
func XDeleteView(scopeDb *gorm.DB, viewEntry *T_scope_view) error {

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Delete view on manager db and scope db, including all associated access rights
			return xDeleteView(txScopeDb, txManagerDb, viewEntry.ScanScope.DbName, viewEntry)
		})
	})

	// Return transaction result, which may be nil or an error
	return errTx
}

// XUpdateView updates a view on scopeDb and manager db in a transactional manner. If the view name changed, the prefix
// of scope db view table names has to be updated too.
func XUpdateView(scopeDb *gorm.DB, viewEntry *T_scope_view) error {

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Generate list of view table names from manager db
			viewTableNames := strings.Split(viewEntry.ViewNames, ",")
			newViewTableNames := make([]string, 0, len(viewTableNames))

			// Iterate existing view tables in scope db to update their name
			for _, oldViewTableName := range viewTableNames {
				for view := range viewDefinitions {
					if strings.HasSuffix(oldViewTableName, "_"+view) {

						// Define view name
						newViewTableName, errSanitize := sanitizeViewName(viewEntry.Name + "_" + view)
						if errSanitize != nil {
							return fmt.Errorf("could not sanitize view name: %s", errSanitize)
						}

						// Remember view name
						newViewTableNames = append(newViewTableNames, newViewTableName)

						// Update view name in scope db
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

			// Update view table names
			viewEntry.ViewNames = strings.Join(newViewTableNames, ",")

			// Update entry in manager db
			errDb := txManagerDb.Model(&viewEntry).Updates(viewEntry).Error
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
	viewEntry *T_scope_view,
	grantedBy string,
	expiry time.Duration,
	description string,
	dbConnectionsUser int,
) (string, string, error) {

	// Generate random username
	randomUsername, errUsername := utils.GenerateToken(utils.AlphaNum, 15)
	if errUsername != nil {
		return "", "", errUsername
	}

	// Generate random password
	charSet := strings.Replace(utils.AlphaNumCaseSymbol, "?", "", -1) // Drop question mark from character set, as gorm has issues with them
	randomPassword, errPassword := utils.GenerateToken(charSet, 25)
	if errPassword != nil {
		return "", "", errPassword
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
	credentials := DbCredentials{
		Username: "token-" + randomUsername, // Prefix random access token username portion to cluster them in database user list
		Password: randomPassword,
	}

	// Calculate expiry time
	expiryTime := time.Now().Add(expiry)

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Grant actual rights in scope db
			errGrant := grantScopeView(txScopeDb, viewEntry, credentials, expiryTime, dbConnectionsUser)
			if errGrant != nil {
				return fmt.Errorf(
					"could not grant user '%s' on scope views '%s' (ID %d): %s",
					credentials.Username,
					viewEntry.Name,
					viewEntry.Id,
					errGrant,
				)
			}

			// Create grant entry in manager db to keep track
			_, errNew := createGrantEntry(
				txManagerDb,
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
// a grant entry to manager db in a transactional manner. List of users must be a DbCredentials struct, because the user
//  might need to be created on the database server if not existing yet.
func XGrantUsers(
	scopeDb *gorm.DB,
	viewEntry *T_scope_view,
	grantedBy string,
	dbCredentials []DbCredentials,
	dbConnectionsUser int,
) error {

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Query all users to give access
			for _, credentials := range dbCredentials {

				// Grant actual rights in scope db
				errGrant := grantScopeView(txScopeDb, viewEntry, credentials, time.Now(), dbConnectionsUser)
				if errGrant != nil {
					return fmt.Errorf(
						"could not grant user '%s' on scope views '%s' (ID %d): %s",
						credentials.Username,
						viewEntry.Name,
						viewEntry.Id,
						errGrant,
					)
				}

				// Create grant entry in manager db to keep track
				_, errNew := createGrantEntry(
					txManagerDb,
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
// grant entry from manager db in a transactional manner
func XRevokeGrants(scopeDb *gorm.DB, scopeView *T_scope_view, usernames []string) error {

	// Start transaction on the scoped db and manager db. The 'Transaction' function will commit if the provided
	// function returns nil and rollback if an error is returned.
	errTx := managerDb.Transaction(func(txManagerDb *gorm.DB) error {
		return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

			// Generate list of view table names from manager db
			viewTableNames := strings.Split(scopeView.ViewNames, ",")

			// Iterate view grants to identify removable ones
			for _, viewGrant := range scopeView.Grants {

				// Identify deletable grant
				if scanUtils.StrContained(viewGrant.Username, usernames) {

					// Revoke view grant
					errRevoke := xRevokeGrant(
						txScopeDb, txManagerDb, scopeView.ScanScope.DbName, &viewGrant, viewTableNames)
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

// xDeleteView takes care of
// 		- removing associated access rights
//		- removing view entry from the manager db
//		- removing view tables from the scope db
// ATTENTION: The changes must be committed/rolled back by the calling function!
func xDeleteView(
	txScopeDb *gorm.DB,
	txManagerDb *gorm.DB,
	dbName string, // Required to also revoke DB connect right if necessary
	scopeView *T_scope_view,
) error {

	// Generate list of view table names from manager db
	viewTableNames := strings.Split(scopeView.ViewNames, ",")

	// Cleanup grants
	for _, viewGrant := range scopeView.Grants {
		errRevoke := xRevokeGrant(txScopeDb, txManagerDb, dbName, &viewGrant, viewTableNames)
		if errRevoke != nil {
			return errRevoke
		}
	}

	// Remove view entry from manager db
	errDelete := txManagerDb.Delete(scopeView).Error
	if errDelete != nil {
		return fmt.Errorf(
			"could not remove view entry '%s' (ID %d) from scan scope '%s' (ID %d): %s",
			scopeView.Name,
			scopeView.Id,
			scopeView.ScanScope.Name,
			scopeView.ScanScope.Id,
			errDelete,
		)
	}

	// Remove view tables from scope db
	errDelete2 := deleteScopeView(txScopeDb, viewTableNames)
	if errDelete2 != nil {
		return fmt.Errorf(
			"could not remove view tables from scope database '%s' (ID %d): %s",
			scopeView.ScanScope.Name,
			scopeView.ScanScope.Id,
			errDelete2,
		)
	}

	// Return nil as evrythign went fine
	return nil
}

// xRevokeGrant takes care of
// 		- removing grant entry from the manager db
//		- removing associated access rights from the scope db
//		- cleaning obsolete credentials from associated database server
// ATTENTION: The changes must be committed/rolled back by the calling function!
func xRevokeGrant(
	txScopeDb *gorm.DB,
	txManagerDb *gorm.DB,
	dbName string, // Required to also revoke DB connect right if necessary
	viewGrant *T_view_grant,
	viewTableNames []string,
) error {

	// Remove grant entry from manager db
	errRevoke1 := txManagerDb.Delete(viewGrant).Error
	if errRevoke1 != nil {
		return fmt.Errorf(
			"could not remove grant entry '%s' from scope view '%s' (ID %d): %s",
			viewGrant.Username,
			viewGrant.ScopeView.Name,
			viewGrant.ScopeView.Id,
			errRevoke1,
		)
	}

	// Remove access rights from scope db
	errRevoke2 := revokeScopeView(txScopeDb, dbName, viewGrant.Username, viewTableNames)
	if errRevoke2 != nil {
		return fmt.Errorf(
			"could not revoke grants of '%s' from scope database view '%s' (ID %d): %s",
			viewGrant.Username,
			viewGrant.ScopeView.Name,
			viewGrant.ScopeView.Id,
			errRevoke2,
		)
	}

	// Try to delete user. It will succeed if the user has no other privileges on the same database server
	deleteServerCredentials(txScopeDb, viewGrant.Username)

	// Return nil as everything went fine
	return nil
}
