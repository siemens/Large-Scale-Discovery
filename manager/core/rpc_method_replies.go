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
	"github.com/siemens/Large-Scale-Discovery/manager/database"
)

type ReplyNotification struct {
	UpdateScopeIds    []uint64 // The scope IDs that were updated on the manager (those need to be re-loaded)
	RemainingScopeIds []uint64 // The scope IDs that are still existing on the manager (all other can be cleaned)
}

type ReplyDatabases struct {
	Databases []database.T_db_server
}

type ReplyScopeId struct {
	ScopeId uint64
}

type ReplyScopeIds struct {
	ScopeIds []uint64
}

type ReplyScopeTargets struct {
	Synchronization bool // Flag indicating whether previous synchronization is still ongoing (no targets in that case)
	Targets         []database.T_discovery
}

type ReplyScopeSecret struct {
	ScopeSecret string
}

type ReplyScanScopes struct {
	ScanScopes []database.T_scan_scope
}

type ReplyScanScope struct {
	ScanScope database.T_scan_scope
}

type ReplyScopeViews struct {
	ScopeViews []database.T_scope_view
}

type ReplyScopeView struct {
	ScopeViews database.T_scope_view
}

type ReplyTargetsUpdate struct {
	Error   string
	Created uint64
	Removed uint64
	Updated uint64
}

type ReplyCredentials struct {
	Username string
	Password string
}

type ReplyScanAgents struct {
	ScanAgents []database.T_scan_agent
}

type ReplySqlLogs struct {
	Logs []database.T_sql_log
}
