/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package memory

import (
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"time"
)

var agents = cmap.New[managerdb.T_scan_agent]() // Map of recently seen scan agents and associated metadata and stats

// UpdateAgent updates cached agent statistics with latest values. Host, IP and instance are used to build a
// unique agent identifier (A machine can run multiple scan agents for different scopes).
func UpdateAgent(
	agentName string,
	agentHost string,
	agentIp string,
	scopeId uint64, // Associating the scan agent to a scan scope
	moduleData map[string]int,
	systemData utils.SystemData,
) {

	// Generate agent identifier from agent information
	identifier := buildIdentifier(agentName, agentHost, "") // Don't use IP for identification as it might be a dynamic one

	// Convert tasks map to json data
	tasks := make(map[string]interface{}, len(moduleData))
	for k, v := range moduleData {
		tasks[k] = v
	}

	// Updated cached data
	agents.Set(identifier, managerdb.T_scan_agent{
		IdTScanScope:    scopeId,
		Name:            agentName,
		Host:            agentHost,
		Ip:              agentIp,
		LastSeen:        time.Now(),
		Tasks:           tasks,
		CpuRate:         systemData.CpuRate,
		MemoryRate:      systemData.MemoryRate,
		Platform:        systemData.Platform,
		PlatformFamily:  systemData.PlatformFamily,
		PlatformVersion: systemData.PlatformVersion,
	})
}

// RemoveAgent removes stats of a certain scan agent instance from memory. The scan agent identifier is a generated
// string comprising the agent's name, hostname and ip (to be unique across all existing scan agents).
func RemoveAgent(agentName string, agentHost string, agentIp string) {

	// Generate agent identifier from agent information
	identifier := buildIdentifier(agentName, agentHost, "") // Don't use IP for identification as it might be a dynamic one

	// Remove entry
	agents.Remove(identifier)
}

// ClearAgents removes all entries from memory. This is necessary, e.g., after transferring all entries to the
// manager, otherwise cleaned up entries on the manager side might pop up again.
func ClearAgents() {
	agents.Clear()
}

// GetAgents returns a copied list of all stored agent stats
func GetAgents() map[string]managerdb.T_scan_agent {

	// Grab cached scope items
	cachedAgentItems := agents.Items()

	// Prepare memory for copy
	cachedAgents := make(
		map[string]managerdb.T_scan_agent,
		len(cachedAgentItems),
	)

	// Iterate cached scope items and copy data over
	for identifier, item := range cachedAgentItems {
		statItem := item
		cachedAgents[identifier] = statItem
	}

	// Return copied map of scan scopes
	return cachedAgents
}

func buildIdentifier(agentName string, agentHost string, agentIp string) string {
	return fmt.Sprintf("%s-%s-%s", agentName, agentHost, agentIp)
}
