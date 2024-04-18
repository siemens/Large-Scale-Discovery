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
	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/broker/brokerdb"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	"strings"
	"time"
)

// GetBlockDiscoveryTargets queries the database for new discovery scan targets and blocks them
func GetBlockDiscoveryTargets(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	amount int,
	timezonesRanges [][]int, // List of timezone ranges currently relevant (within working hours)
	scanIp string,
	scanHostname string,
) ([]managerdb.T_discovery, error) {

	// Return if no timezones suite the current configuration
	if len(timezonesRanges) < 1 {
		logger.Infof("No timezones within the configured working hours.")
		return []managerdb.T_discovery{}, nil
	}

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		return nil, fmt.Errorf("could not open scope database: %s", errHandle)
	}

	// Prepare query result
	var targets []managerdb.T_discovery

	// Start transaction on the scoped db. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// Search for new targets
		query := txScopeDb.Model(&managerdb.T_discovery{}).
			Where("scan_started IS NULL"). // Not active elements only
			Where("enabled IS TRUE")       // Enabled elements only

		// Attach timezone condition
		if len(timezonesRanges) > 0 {
			clause := make([]string, 0, len(timezonesRanges))
			clauseValues := make([]interface{}, 0, len(timezonesRanges)*2)
			for _, r := range timezonesRanges {
				clause = append(clause, "timezone BETWEEN ? AND ?")
				clauseValues = append(clauseValues, r[0], r[1])
			}
			query = query.Where(
				strings.Join(clause, " OR "),
				clauseValues...,
			)
		}

		// Attach constraints
		query = query.
			Order("priority DESC"). // Priority elements first
			Order("random()").      // Randomize, despite rowing priority elements first
			Limit(amount)

		// Execute query
		errDb := query.Find(&targets).Error
		if errDb != nil {
			logger.Errorf("Could not query targets from scope db: %s", errDb)
			return errDb
		}

		// Block targets
		for _, target := range targets {
			errUpdate := blockDiscoveryTarget(txScopeDb, target.Id, scanIp, scanHostname)
			if errUpdate != nil {
				logger.Errorf("Could not update discovery target: %s", errUpdate)
				return errUpdate
			}
		}

		// Return nil as everything went fine
		return nil
	})

	// Return transaction error if transaction failed
	if errTx != nil {
		return nil, errTx
	}

	// Return selected targets
	return targets, nil
}

// SaveDiscoveryResult parses a discovery scan result and writes it into the database. Subsequently, it passes results
// to the cache in order to queue relevant targets for submodules. Furthermore, it sets the "scan_finished" timestamp
// to indicate a completed process.
func SaveDiscoveryResult(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	idTDiscovery uint64, // T_discovery ID (not T_discovery_services!!), this result belongs to
	result *discovery.Result,
) error {

	// Log action
	logger.Debugf("Saving discovery scan result.")

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		return fmt.Errorf("could not open scope database: %s", errHandle)
	}

	// Prepare list of potential submodule scan targets to add later
	var potentialSubTargets []*brokerdb.T_sub_input

	// Note time for timing measurement
	start := time.Now()

	// Prepare services counter for log messages
	services := 0

	// Start transaction on the scoped db. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	errTx := scopeDb.Transaction(func(txScopeDb *gorm.DB) error {

		// Log database request
		logger.Debugf("Querying discovery target data.")

		// Prepare query result
		var inputTarget = managerdb.T_discovery{}

		// Get data from t_discovery which needs to replicated into t_discovery_hosts/t_discovery_services/t_discovery_scripts
		errDb := txScopeDb.Model(&inputTarget).
			Where("id = ?", idTDiscovery).
			First(&inputTarget).Error
		if errors.Is(errDb, gorm.ErrRecordNotFound) { // Check if entry didn't exist
			logger.Infof("Discovery target '%d' does not exist anymore. Dropping result.", idTDiscovery)
			return nil
		} else if errDb != nil {
			return fmt.Errorf("could not query meta data from scope db: %s", errDb)
		}

		// Drop scan result if the scan input entry was reset (e.g. to initiate a fresh scan)
		if inputTarget.ScanStarted.Valid == false {
			logger.Infof("Discovery target '%d' was reset. Dropping result.", idTDiscovery)
			return nil
		}

		// Prepare scan finished timestamp
		finished := sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}

		// Prepare entry slices. We need a slice of pointers for the hostEntries, as we need the same reference
		// in order to set it in the respective service and script entries.
		hostEntries := make([]*managerdb.T_discovery_host, 0, len(result.Data))
		serviceEntries := make([]managerdb.T_discovery_service, 0, 10)
		scriptEntries := make([]managerdb.T_discovery_script, 0, 10)

		// Iterate hosts and prepare hosts/services/script database entries
		for _, hostResult := range result.Data {

			// Log current step
			logger.Debugf("Creating host entry for '%s'.", hostResult.Ip)

			// Prepare address data
			var address string
			if hostResult.DnsName != "" {
				address = hostResult.DnsName
			} else {
				address = hostResult.Ip
			}

			// Decide critical flag
			critical := false
			if hostResult.Ad.CriticalObject || hostResult.Critical {
				critical = true
			}

			// Create structs that are the same for host, service and script entries
			columnsHost := managerdb.ColumnsHost{
				Address:    utils.ValidUtf8String(address),
				Ip:         utils.ValidUtf8String(hostResult.Ip),
				DnsName:    utils.ValidUtf8String(hostResult.DnsName),
				OtherNames: utils.ValidUtf8String(strings.Join(hostResult.OtherNames, DbValueSeparator)),
				OtherIps:   utils.ValidUtf8String(strings.Join(hostResult.OtherIps, DbValueSeparator)),
				Critical:   critical,
				Hops:       utils.ValidUtf8String(strings.Join(hostResult.Hops, DbValueSeparator)),
				ScanCycle:  scanScope.Cycle,
			}
			columnsOs := managerdb.ColumnsOs{
				OsGuess:    utils.ValidUtf8String(strings.Join(hostResult.OsGuesses, DbValueSeparator)),
				OsSmb:      utils.ValidUtf8String(hostResult.OsSmb),
				OsLastBoot: sql.NullTime{Time: hostResult.LastBoot, Valid: true},
				OsUptime: sql.NullInt64{
					Int64: int64(hostResult.Uptime.Seconds()),
					Valid: true,
				},
				OsAdminUsers: utils.ValidUtf8String(strings.Join(hostResult.AdminUsers, DbValueSeparator)),
				OsRdpUsers:   utils.ValidUtf8String(strings.Join(hostResult.RdpUsers, DbValueSeparator)),
			}
			columnsAsset := managerdb.ColumnsAsset{
				AssetCompany:    utils.ValidUtf8String(hostResult.Company),
				AssetDepartment: utils.ValidUtf8String(hostResult.Department),
				AssetOwner:      utils.ValidUtf8String(hostResult.Owner),
			}
			columnsScan := managerdb.ColumnsScan{
				ScanStarted:  inputTarget.ScanStarted,
				ScanFinished: finished,
				ScanStatus:   result.Status,
				ScanIp:       inputTarget.ScanIp,
				ScanHostname: inputTarget.ScanHostname,
			}
			columnsInput := managerdb.ColumnsInput{
				Input:     inputTarget.Input,
				InputSize: inputTarget.InputSize,
			}
			columnsInputDetails := managerdb.ColumnsInputDetails{
				Timezone:           inputTarget.Timezone,
				Lat:                inputTarget.Lat,
				Lng:                inputTarget.Lng,
				PostalAddress:      inputTarget.PostalAddress,
				InputNetwork:       inputTarget.InputNetwork,
				InputCountry:       inputTarget.InputCountry,
				InputLocation:      inputTarget.InputLocation,
				InputRoutingDomain: inputTarget.InputRoutingDomain,
				InputZone:          inputTarget.InputZone,
				InputPurpose:       inputTarget.InputPurpose,
				InputCompany:       inputTarget.InputCompany,
				InputDepartment:    inputTarget.InputDepartment,
				InputManager:       inputTarget.InputManager,
				InputContact:       inputTarget.InputContact,
				InputComment:       inputTarget.InputComment,
			}
			columnsAd := managerdb.ColumnsAd{
				AdName:                 utils.ValidUtf8String(hostResult.Ad.Name),
				AdDistinguishedName:    utils.ValidUtf8String(hostResult.Ad.DistinguishedName),
				AdDnsName:              utils.ValidUtf8String(hostResult.Ad.Name),
				AdCreated:              sql.NullTime{Time: hostResult.Ad.Created, Valid: hostResult.Ad.Created != time.Time{}},
				AdLastLogon:            sql.NullTime{Time: hostResult.Ad.LastLogon, Valid: hostResult.Ad.LastLogon != time.Time{}},
				AdLastPassword:         sql.NullTime{Time: hostResult.Ad.LastPassword, Valid: hostResult.Ad.LastPassword != time.Time{}},
				AdDescription:          utils.ValidUtf8String(strings.Join(hostResult.Ad.Description, DbValueSeparator)),
				AdLocation:             utils.ValidUtf8String(hostResult.Ad.Location),
				AdManagedBy:            utils.ValidUtf8String(hostResult.Ad.ManagedByCn),
				AdManagedByGid:         utils.ValidUtf8String(hostResult.Ad.ManagedByGid),
				AdManagedByDepartment:  utils.ValidUtf8String(hostResult.Ad.ManagedByDepartment),
				AdOs:                   utils.ValidUtf8String(hostResult.Ad.Os),
				AdOsVersion:            utils.ValidUtf8String(hostResult.Ad.OsVersion),
				AdServicePrincipalName: utils.ValidUtf8String(strings.Join(hostResult.Ad.ServicePrincipalName, DbValueSeparator)),
			}

			// Prepare host entry
			// Fields used from this entry that are supposed to be set by the DB and are used by the service or script
			// entries need to be pointers to be set correctly! See the ID reference as an example
			tDiscoveryHost := &managerdb.T_discovery_host{
				IdTDiscovery:        idTDiscovery, // Associate the fk
				ColumnsHost:         columnsHost,
				PortsOpen:           len(hostResult.Services),
				ColumnsOs:           columnsOs,
				ColumnsAsset:        columnsAsset,
				ColumnsScan:         columnsScan,
				ColumnsInput:        columnsInput,
				ColumnsInputDetails: columnsInputDetails,
				ColumnsAd:           columnsAd,
			}
			hostEntries = append(hostEntries, tDiscoveryHost)

			// Increment counter
			services += len(hostResult.Services)

			// Insert related service outputs
			for _, serviceResult := range hostResult.Services {

				// Log current step
				logger.Debugf("Creating service entry for %d/%s.", serviceResult.Port, serviceResult.Protocol)

				// Unify case
				service := strings.ToLower(serviceResult.Name)
				tunnel := strings.ToLower(serviceResult.Tunnel)

				// In some cases Nmap uses the additional tunnel attribute in XML output, instead of setting the
				// usual service name. E.g. instead of "https" it might discover a service as "http" and additionally
				// setting the tunnel attribute to "ssl". In that its necessary to manually reflect the SSL
				// characteristic by building a service name string similar to other Nmap's service names.
				if strings.Contains(service, "https") {
					// all information covered
				} else if strings.Contains(service, tunnel) {
					// all information covered
				} else if tunnel == "ssl" && service == "http" {
					service = "https" // Cleanup, Nmap sometimes describes https as http indicating SSL through the additional tunnel attribute
				} else if tunnel != "" {
					service = tunnel + "/" + service
				}

				// Prepare service entry
				serviceEntries = append(serviceEntries, managerdb.T_discovery_service{
					ColumnsHost:         columnsHost,
					Port:                serviceResult.Port,
					Protocol:            serviceResult.Protocol,
					Service:             service,
					ServiceProduct:      serviceResult.Product,
					ServiceVersion:      serviceResult.Version,
					ServiceDeviceType:   serviceResult.DeviceType,
					ServiceCpes:         strings.Join(serviceResult.Cpes, DbValueSeparator),
					ServiceFlavor:       serviceResult.Flavor,
					ServiceTtl:          serviceResult.Ttl,
					ColumnsOs:           columnsOs,
					ColumnsAsset:        columnsAsset,
					ColumnsScan:         columnsScan,
					ColumnsInput:        columnsInput,
					ColumnsInputDetails: columnsInputDetails,
					ColumnsAd:           columnsAd,

					TDiscoveryHost: tDiscoveryHost,
				})
			}

			// Insert related script outputs
			for _, scriptResult := range hostResult.Scripts {

				// Log current step
				logger.Debugf(
					"Creating script entry for %s (%d/%s).",
					scriptResult.Name,
					scriptResult.Port,
					scriptResult.Protocol,
				)

				// Prepare script entry
				scriptEntries = append(scriptEntries, managerdb.T_discovery_script{
					ColumnsHost:         columnsHost,
					Port:                scriptResult.Port,
					Protocol:            scriptResult.Protocol,
					ScriptType:          scriptResult.Type,
					ScriptName:          scriptResult.Name,
					ScriptOutput:        strings.Trim(scriptResult.Result, "\n"), // Some NSE outputs start with newline
					ColumnsOs:           columnsOs,
					ColumnsAsset:        columnsAsset,
					ColumnsScan:         columnsScan,
					ColumnsInput:        columnsInput,
					ColumnsInputDetails: columnsInputDetails,
					ColumnsAd:           columnsAd,

					TDiscoveryHost: tDiscoveryHost,
				})
			}

		}

		// Insert discovery host entries into the db. Order matters here, as we'd get a unique index violation if we
		// first create the service/script entries. In this case gorm would:
		// - insert service/script, detect that the corresponding hosts are not existing
		// - create the missing hosts (correctly) according to the host struct in the service/script
		// - try to create the hosts from the hostEntries slice again
		// Empty slices would result in an error and we don't consider an empty slice an error, as there might simply
		// be no results
		if len(hostEntries) > 0 {

			// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
			// exceed PostgreSQLs limit of 65535 parameters
			errDb2 := txScopeDb.
				Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeDiscoveryHost}).
				Create(&hostEntries).Error
			if errDb2 != nil {

				// Rollback everything if we cannot insert something
				return fmt.Errorf("could not insert host entry into scope db: %s", errDb2)
			}
		}

		// Create new entries for the service results. Empty slices would result in an error and we don't consider an
		// empty slice an error, as there might simply be no results.
		if len(serviceEntries) > 0 {

			// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
			// exceed PostgreSQLs limit of 65535 parameters
			errDb3 := txScopeDb.
				Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeDiscoveryService}).
				Create(&serviceEntries).Error
			if errDb3 != nil {

				// Rollback everything if we cannot insert something
				return fmt.Errorf("could not insert service entry into scope db: %s", errDb3) // Rollback everything if we cannot insert something
			}
		}

		// Create new entries for the script results. Empty slices would result in an error and we don't consider an
		// empty slice an error, as there might simply be no results.
		if len(scriptEntries) > 0 {

			// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
			// exceed PostgreSQLs limit of 65535 parameters
			errDb4 := txScopeDb.
				Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeDiscoveryScript}).
				Create(&scriptEntries).Error
			if errDb4 != nil {

				// Rollback everything if we cannot insert something
				return fmt.Errorf("could not insert script entry into scope db: %s", errDb4) // Rollback everything if we cannot insert something
			}
		}

		// Mark discovery target as finished
		logger.Debugf("Completing result.")
		errCompletion := completeDiscoveryResult(logger, txScopeDb, idTDiscovery, finished.Time, result.Status)
		if errCompletion != nil {
			return fmt.Errorf("could not update associated discovery entry in scope db: %s", errCompletion)
		}

		// Assemble the potential sub-targets, as the services now have valid IDs
		// Attention: The sub targets can only be committed/created after the scope db transaction got committed, to
		// avoid race conditions, where submodule targets are started before parent entries are available in scope db.
		for _, service := range serviceEntries {

			// Create generic submodule target, which might be added to the cache for multiple submodules
			potentialSubTargets = append(potentialSubTargets, &brokerdb.T_sub_input{
				IdTDiscoveryService: service.Id,
				Address:             service.Address,
				Ip:                  service.Ip,
				DnsName:             service.DnsName,
				OtherNames:          service.OtherNames,
				Protocol:            service.Protocol,
				Port:                service.Port,
				Service:             service.Service,
				ServiceProduct:      service.ServiceProduct,
			})
		}

		// Return nil as everything went fine
		return nil
	})

	// Return with error if transaction failed
	if errTx != nil {
		return errTx
	}

	// Add targets to brokerdb for submodule scans. brokerdb.AddScopeTargets() will only add necessary ones.
	errAdd := brokerdb.AddScopeTargets(logger, scanScope, potentialSubTargets)
	if errAdd != nil {
		return fmt.Errorf("could not add submodule targets to broker db: %s", errAdd)
	}

	// Log benchmark, warn about longer inserts. If too many too long inserts happen, sqlite might not suffice anymore!
	logger.Debugf("Saving discovery result for %d open ports took %s.", services, time.Since(start).String())

	// Return nil as everything went fine
	return nil
}

// blockDiscoveryTarget blocks a discovery scan target in the database from being taken again by setting a
// "scan_started" timestamp.
func blockDiscoveryTarget(
	txScopeDb *gorm.DB,
	inputId uint64,
	scanIp string,
	scanHostname string,
) (retErr error) {

	// Update input entry with latest scan state information
	db := txScopeDb.Model(&managerdb.T_discovery{}).
		Where("id = ?", inputId).
		Updates(map[string]interface{}{
			"scan_started":  time.Now(),
			"scan_status":   scanUtils.StatusRunning,
			"scan_ip":       scanIp,
			"scan_hostname": scanHostname,
		})
	if db.Error != nil {
		return db.Error
	}

	// No update done ... something went wrong
	if db.RowsAffected != 1 {
		return fmt.Errorf("could not block any discovery target")
	}

	// Return nil as everything went fine
	return nil
}

// completeDiscoveryResult updates the "scan_finished" timestamp to indicate a completed process.
func completeDiscoveryResult(
	logger scanUtils.Logger,
	txScopeDb *gorm.DB,
	idTDiscovery uint64,
	finished time.Time,
	status string,
) (retErr error) {

	// Define update values
	values := map[string]interface{}{
		"priority":      false, // Reset priority flag on completion
		"scan_finished": finished,
		"scan_status":   status,
	}

	// Update input entry with final scan state information
	errDb := txScopeDb.Model(&managerdb.T_discovery{}).
		Where("id = ?", idTDiscovery).
		Updates(values).
		Update("scan_count", gorm.Expr("scan_count + ?", 1)).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}
