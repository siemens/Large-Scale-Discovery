/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package brokerdb

import (
	"database/sql"
)

// T_sub_input contains details of a queued submodule scan target (scan module, input data,...) and a reference to the
// scan scope it belongs. Scan submodules will be fed from this table.
type T_sub_input struct {
	Id                  uint64 `gorm:"column:id;primaryKey"`                                                 // Id autoincrement
	IdTScanScope        uint64 `gorm:"column:id_t_scan_scope;uniqueIndex:uix_sub_input_multi_column"`        // Scope ID this input target belongs to (two indexes, a multi column combined one and a dedicated one)
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;uniqueIndex:uix_sub_input_multi_column"` // Id of the related t_discovery_services entry (two indexes, a multi column combined one and a dedicated one)

	Module         string       `gorm:"column:module;uniqueIndex:uix_sub_input_multi_column"` // Module that does the scan
	ScanStarted    sql.NullTime `gorm:"column:scan_started"`                                  // Timestamp when scan was started (or nil)
	Address        string       `gorm:"column:address"`                                       // The address of the target (Dns name over IP)
	Ip             string       `gorm:"column:ip"`                                            // The IP address of the target
	DnsName        string       `gorm:"column:dns_name"`                                      // The Dns name of the target (if available)
	OtherNames     string       `gorm:"column:other_names;type:text"`                         // Potential other hostnames of the target (symbol separated string)
	Protocol       string       `gorm:"column:protocol"`                                      // The protocol of the target service. Usually tcp.
	Port           int          `gorm:"column:port"`                                          // The port of the target service
	Service        string       `gorm:"column:service;type:text"`                             // The service type discovered by the version detection
	ServiceProduct string       `gorm:"column:service_product;type:text"`                     // The service product discovered by the version detection
}

// MaxBatchSizeSubInput defines the maximum number T_sub_input instances that can be batched together
// during an insert. This is calculated dividing 999 (SQLITE) by the number of fields (that are actually written to the db).
const MaxBatchSizeSubInput = 76 // 999 (SQLITE) / 13
