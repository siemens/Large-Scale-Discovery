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
	"context"
	"errors"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"sync"
)

/*
 * The RPC wrapper function defined in this file matches a respective broker RPC method. It can be used to call
 * the respective RPC method. These functions actually belong to the foreign components making use of them, calling
 * the respective broker RPC methods. They are put here to make them reusable for any foreign component. Multiple
 * foreign components might want to make use of same RPC functions to call the respective RPC methods. Hence,
 * these functions will be compiled into the other foreign components. The foreign component must pass its own RPC
 * client to execute the request.
 */

// RpcRequestScanTasks queries the broker for new scan tasks via RPC
func RpcRequestScanTasks(
	logger scanUtils.Logger,
	rpc *utils.Client,
	ctx context.Context,
	rpcArgs *ArgsGetScanTask,
) []ScanTask {

	// Prepare RPC request
	rpcEndpoint := "Broker.RequestScanTasks"
	rpcReply := ReplyGetScanTask{}

	// Loop in case of a currently broken RPC connection
	for {

		// Log action
		logger.Debugf("Requesting scan tasks.")

		// Send RPC request.
		errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {

			// Log situation
			logger.Debugf("Waiting for re-connection.")

			// Wait until RPC connection or agent shutdown
			select {
			case <-ctx.Done(): // Cancellation signal
				return []ScanTask{}
			case <-rpc.Established():
				logger.Debugf("Broker re-connected.")
				break
			}

		} else if errRpc != nil { // In case of error, return empty list, it will be retried later again.
			return []ScanTask{}
		} else {
			logger.Debugf("Scan tasks requested.")
			return rpcReply.ScanTasks
		}
	}
}

// RpcSubmitScanResult sends scan results to the broker via RPC
func RpcSubmitScanResult(
	logger scanUtils.Logger,
	rpc *utils.Client,
	ctx context.Context,
	wg *sync.WaitGroup,
	chThrottle chan struct{},
	rpcArgs interface{},
) {

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the client for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrement wait group
	defer func() { wg.Done() }()

	// Release throttle slot once done
	defer func() { <-chThrottle }()

	// Prepare RPC request
	rpcEndpoint := "Broker.SubmitScanResult"
	rpcReply := struct{}{}

	// Try to send until success
	logMessage := "Sending scan result."
	for {

		// Log transmission attempt
		logger.Debugf(logMessage)

		// Attempt to send scan result (immediately, without ticker delay)
		errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {

			// Log situation
			logger.Debugf("Waiting for re-connection.")

			// Wait until RPC connection or agent shutdown
			select {
			case <-ctx.Done(): // Cancellation signal

				// Discard results if agents shuts down and broker can't receive results
				logger.Infof("Scan result discarded, agent termination already requested.")
				return

			case <-rpc.Established():
				logger.Debugf("Broker re-connected.")
				logMessage = "Retrying to send scan result."
				break
			}

		} else if errRpc != nil {
			logger.Warningf("Sending scan result failed: %s", errRpc)
			return
		} else {
			logger.Debugf("Scan result sent.")
			return
		}
	}
}
