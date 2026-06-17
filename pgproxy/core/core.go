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

import (
	"path/filepath"
	"sync"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/log"
	manager "github.com/siemens/Large-Scale-Discovery/manager/core"
	"github.com/siemens/Large-Scale-Discovery/pgproxy/config"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

var shutdownOnce sync.Once  // Helper variable to prevent shutdown from doing its work multiple times.
var rpcClient *utils.Client // RPC client struct handling RPC connections and requests. This needs to be accessible by handler packages

func ConnectManager() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Prepare RPC certificate path
	rpcRemoteCrt := filepath.Join("keys", "manager.crt")
	if _build.DevMode {
		rpcRemoteCrt = filepath.Join("keys", "manager_dev.crt")
	}
	errRemoteCrt := scanUtils.IsValidFile(rpcRemoteCrt)
	if errRemoteCrt != nil {
		return errRemoteCrt
	}

	// Register gob structures that will be sent via interface{}
	manager.RegisterGobs()

	// Initialize RPC client manager facing.
	// The manager requires a shared secret to authorize the RPC connection.
	rpcClient = utils.NewRpcClient(conf.ManagerAddress, conf.ManagerSsl, rpcRemoteCrt, conf.ManagerSecret)

	// Connect to manager but don't wait to start answering client requests. Connection attempt continues in background.
	_ = rpcClient.Connect(logger, true)

	// Return as everything went fine
	return nil

}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Disconnect from manager
		if rpcClient != nil {
			rpcClient.Disconnect()
		}
	})
}

// RpcClient exposes the RPC client to external packages
func RpcClient() *utils.Client {
	return rpcClient
}
