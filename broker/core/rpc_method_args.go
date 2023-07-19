/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import "large-scale-discovery/utils"

// AgentInfo contains agent identifying meta data to describe the origin of the request
// ATTENTION: Do not use this data for security checks, it can be crafted!
type AgentInfo struct {
	Name string // Instance name of the scan agent. There may be multiple scan agents running on the same system (e.g. to target different scan scopes).
	Host string // Host used during scanning. Logged by the broker. Decided by scan agent, because only it knows the IP address of it's scanning interface.
	Ip   string // Ip address used during scanning. Logged by the broker. Decided by scan agent, because only it knows the IP address of it's scanning interface.
}

// ModuleData contains meta data of a scan module on an agent (e.g. how many of its kind are running,...)
type ModuleData struct {
	Label       string // Name of the respective module, as used by the scan module itself
	ActiveTasks int    // Number of instances active
}

// ArgsGetScanTask contains meta data of a scan agent requesting scan targets
type ArgsGetScanTask struct {
	AgentInfo                    // Identifying scan agent information to distinguish scan agent instances for informational purposes
	ScopeSecret string           // Scan scope secret to authenticate/associate this scan result to
	ModuleData  []ModuleData     // List of already running modules on the agent
	SystemData  utils.SystemData // Some system information, like CPU load,...
}

// ArgsSaveScanResult contains meta data about a scan result and the result data itself, sent by a scan agent
type ArgsSaveScanResult struct {
	AgentInfo               // Identifying scan agent information to distinguish scan agent instances for informational purposes
	ScopeSecret string      // Scan scope secret to authenticate/associate this scan result to
	Id          uint64      // Id is passed back to allow the broker associating this result set with the original request. The Id might be either the t_discovery entry ID from the scope db or the t_sub_input entry ID from the brokerdb.
	Result      interface{} // Generic interface that holds structure for different scan results
}
