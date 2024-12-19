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
	"encoding/gob"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"time"
)

// RegisterGobs registers data structs for RPC to make them transferable as interface variables
func RegisterGobs() {
	gob.Register(database.T_db_server{})
	gob.Register(database.T_scan_scope{})
	gob.Register(database.T_scan_setting{})
	gob.Register(database.T_scope_view{})
	gob.Register(database.T_view_grant{})
	gob.Register([]database.T_discovery{}) // database.JsonMap might include this kind of type
	gob.Register(database.T_sql_log{})
	gob.Register(time.Time{})              // Included in some of the above structs
	gob.Register([]time.Weekday{})         // scope settings contain this type
	gob.Register([]interface{}{})          // database.JsonMap might include this kind of type
	gob.Register(map[string]interface{}{}) // database.JsonMap might include this kind of type
	gob.Register(time.Time{})              // Included in some of the above structs
}
