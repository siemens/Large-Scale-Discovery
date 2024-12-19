/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"time"
)

type ArgsDbServerId struct {
	DbServerId uint64
}

type ArgsDatabaseDetails struct {
	DbServerId uint64 // The ID of the database server to edit
	Name       string
	Dialect    string
	Host       string
	HostPublic string
	Port       int
	Admin      string
	Password   string
	Args       string
}

type ArgsScopeFull struct {
	PrivilegeSecret string
	ScopeSecret     string
}

type ArgsScopeByIdFull struct {
	PrivilegeSecret string
	ScopeId         uint64
}

type ArgsScopeId struct {
	ScopeId uint64
}

type ArgsGroupIds struct {
	GroupIds []uint64
}

type ArgsScopeDetails struct {
	DbServerId      uint64 // The ID of the scope database server entry to create the scope on
	Name            string
	GroupId         uint64
	CreatedBy       string
	Type            string        // There might be different scope types in the future, e.g. initialized via remote repositories
	Cycles          bool          // Whether scan scope should run in cycles
	CyclesRetention int           // Amount of previous scan cycles to keep. Older ones will be cleaned up.
	Attributes      utils.JsonMap // Key value pairs of scan scope attributes to update
}

type ArgsScopeUpdate struct {
	// Pointer values are optional and will not be updated if nil!
	IdTScanScopes   uint64 // The ID of the scan scope to update
	Name            string
	Cycles          bool
	CyclesRetention int            // Amount of previous scan cycles to keep. Older ones will be cleaned up.
	Attributes      *utils.JsonMap // Key value pairs of scan scope attributes to update
}

type ArgsTargetsUpdate struct {
	IdTScanScopes uint64                 // The ID of the scan scope to update
	Targets       []database.T_discovery // Scope targets to set in the scopedb. Empty list will remove all.
	Blocking      bool                   // Whether to wait for update result
}

type ArgsTargetReset struct {
	ScopeId uint64 // The ID of the scan scope to update target in
	Input   string
}

type ArgsSettingsUpdate struct {
	IdTScanScopes uint64 // The ID of the scan scope to update
	ScanSettings  database.T_scan_setting
}

type ArgsStatsUpdate struct {
	ScanAgents map[uint64][]database.T_scan_agent
}

type ArgsViewDetails struct {
	ScopeId   uint64 // The ID of the scan scope to create the view for
	ViewName  string // Name of the view to describe it
	CreatedBy string
	Filters   map[string][]string
}

type ArgsViewId struct {
	ViewId uint64
}

type ArgsViewUpdate struct {
	ViewId uint64 // The ID of the scope view to update
	Name   string
}

type ArgsUsername struct {
	Username string
}

type ArgsCredentials struct {
	Username string
	Password string
}

type ArgsGrantToken struct {
	ViewId      uint64
	Description string
	CreatedBy   string
	Expiry      time.Duration
}

type ArgsGrantUsers struct {
	ViewId        uint64
	DbCredentials []database.DbCredentials
	GrantedBy     string
}

type ArgsRevokeGrants struct {
	ViewId    uint64
	Usernames []string
}

type ArgsAgentId struct {
	AgentId uint64
}

type ArgsSqlLogCreate database.T_sql_log

type ArgsSqlLogsFilter struct {
	DbName string
	Since  time.Time
}
