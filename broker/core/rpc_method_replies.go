/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
)

// ScanTask contains data of a single scan task used by the agent to start a scan. This struct is held generic to fit
// all scan modules and to simplify broker-side code. Required information is picked from this struct by the agent.
// This struct will contain copied data from a cached scan target struct.
type ScanTask struct {
	Secret         string // Scope secret identifying the scan scope this scan tasks belongs to
	Label          string // Name of the respective module to use these arguments with
	Id             uint64 // PK from the source table (might be a t_discovery ID (scope db) or a sub scan target ID (broker db)
	Target         string
	Protocol       string
	Port           int
	OtherNames     []string
	Service        string
	ServiceProduct string
	ScanSettings   managerdb.T_scan_setting // Current scan settings taken from the scan scope
}

// ReplyGetScanTask contains a list of scan tasks to be returned to a scan agent after requesting
type ReplyGetScanTask struct {
	ScanTasks []ScanTask
}
