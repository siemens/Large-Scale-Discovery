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
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

/*
 * The RPC wrapper function defined in this file matches a respective manager RPC method. It can be used to call
 * the respective RPC method. These functions actually belong to the foreign components making use of them, calling
 * the respective manager RPC methods. They are put here to make them reusable for any foreign component. Multiple
 * foreign components might want to make use of same RPC functions to call the respective RPC methods. Hence,
 * these functions will be compiled into the other foreign components. The foreign component must pass its own RPC
 * client to execute the request.
 */

func RpcGetScopeFull(
	logger scanUtils.Logger,
	rpc *utils.Client,
	managerPrivilegeSecret string,
	scopeSecret string,
) (database.T_scan_scope, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetScopeFull"
	rpcReply := ReplyScanScope{}
	rpcArgs := ArgsScopeFull{
		PrivilegeSecret: managerPrivilegeSecret,
		ScopeSecret:     scopeSecret,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return database.T_scan_scope{}, errRpc
	}

	// Return scan scopes
	return rpcReply.ScanScope, nil
}
