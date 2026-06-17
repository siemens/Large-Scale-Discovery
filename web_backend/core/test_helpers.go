/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import "github.com/siemens/Large-Scale-Discovery/utils"

// SetRpcClientForTest injects an RPC client for use in tests. Call this from TestMain
// before running handlers that invoke manager RPC functions.
func SetRpcClientForTest(client *utils.Client) {
	rpcClient = client
}
