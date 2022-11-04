/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	managerdb "large-scale-discovery/manager/database"
)

// ScanTask contains data of a single scan task used by the agent to start a scan. This struct is held generic to fit
// all scan modules and to simplify broker-side code. Required information is picked from this struct by the agent.
// This struct will contain copied data from a cached scan target struct.
type ScanTask struct {
	Label          string // Name of the respective module to use this arguments with
	Id             uint64 // PK from the source table (might be a t_discovery ID (scope db) or a sub scan target ID (broker db)
	Target         string
	Protocol       string
	Port           int
	OtherNames     []string
	Service        string
	ServiceProduct string
	ScanSettings   managerdb.T_scan_settings // Current scan settings taken from the scan scope
}

// ReplyGetScanTask contains a list of scan tasks to be returned to a scan agent after requesting
type ReplyGetScanTask struct {
	ScanTasks []ScanTask
}
