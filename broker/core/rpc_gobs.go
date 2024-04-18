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
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
)

// RegisterGobs registers data structs for RPC to make them transferable as interface variables
func RegisterGobs() {
	gob.Register(banner.Result{})
	gob.Register(discovery.Result{})
	gob.Register(nfs.Result{})
	gob.Register(smb.Result{})
	gob.Register(ssh.Result{})
	gob.Register(ssl.Result{})
	gob.Register(webcrawler.Result{})
	gob.Register(webenum.Result{})
	gob.Register(managerdb.T_scan_settings{})
	gob.Register([]interface{}{})          // database.JsonMap might include this kind of type
	gob.Register(map[string]interface{}{}) // database.JsonMap might include this kind of type
}
