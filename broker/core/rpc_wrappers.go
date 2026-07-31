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
	"context"
	"errors"
	"net/rpc"
	"sync"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

/*
 * The RPC wrapper function defined in this file matches a respective broker RPC method. It can be used to call
 * the respective RPC method. These functions actually belong to the foreign components making use of them, calling
 * the respective broker RPC methods. They are put here to make them reusable for any foreign component. Multiple
 * foreign components might want to make use of same RPC functions to call the respective RPC methods. Hence,
 * these functions will be compiled into the other foreign components. The foreign component must pass its own RPC
 * client to execute the request.
 */

// restoreRpcServerError restores broker sentinel errors whose concrete type was lost by net/rpc. Every RPC call in
// this RPC wrapper file must pass its result through this function. Add new server-returned broker sentinels here
// so all wrapper functions restore them consistently.
func restoreRpcServerError(err error) error {

	// Return nil errors directly
	if err == nil {
		return nil
	}

	// Return errors that are already restored
	var serverError rpc.ServerError
	if !errors.As(err, &serverError) {
		return err
	}

	// Try to restore error
	switch serverError.Error() {
	case errInvalidScopeSecret.Error():
		return errInvalidScopeSecret
	default:

		// Return unknown error
		return err
	}
}

// RpcRequestScanTasks queries the broker for new scan tasks via RPC
func RpcRequestScanTasks(
	logger scanUtils.Logger,
	rpc *utils.Client,
	ctx context.Context,
	rpcArgs *ArgsGetScanTask,
) ([]ScanTask, error) {

	// Prepare RPC request
	rpcEndpoint := "Broker.RequestScanTasks"
	rpcReply := ReplyGetScanTask{}

	// Loop in case of a currently broken RPC connection
	for {

		// Do not start another request after cancellation
		select {
		case <-ctx.Done():
			return []ScanTask{}, nil
		default:
		}

		// Log action
		logger.Debugf("Requesting scan tasks with scope secret '%s...'.", rpcArgs.ScopeSecret[:5])

		// Clear data from a potentially interrupted previous response
		rpcReply = ReplyGetScanTask{}

		// Send RPC request and restore the sentinel identity lost by net/rpc.
		errRpc := restoreRpcServerError(rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply)) // rpcReply must be a pointer to receive a result!
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {

			// Log situation
			logger.Debugf("Waiting for reconnection.")

			// Wait until RPC connection or client shutdown
			select {
			case <-ctx.Done(): // Cancellation signal
				return []ScanTask{}, nil
			case _, ok := <-rpc.Established():

				// Stop when the RPC client terminates
				if !ok {
					return []ScanTask{}, nil
				}

				// Log reconnection
				logger.Debugf("Broker reconnected.")

				// Break case
				break
			}

		} else if errRpc != nil {
			return []ScanTask{}, errRpc
		} else {
			logger.Debugf("Scan tasks requested.")
			return rpcReply.ScanTasks, nil
		}
	}
}

// RpcSubmitScanResult sends scan results to the broker via RPC. Retries until success, only giving up on
// permanently invalid scope secrets or agent shutdown.
func RpcSubmitScanResult(
	logger scanUtils.Logger,
	rpc *utils.Client,
	ctx context.Context,
	wg *sync.WaitGroup,
	chThrottle chan struct{},
	rpcArgs interface{},
) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the client for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrement wait group
	defer wg.Done()

	// Release throttle slot once done
	defer func() { <-chThrottle }()

	// Prepare RPC request
	rpcEndpoint := "Broker.SubmitScanResult"
	rpcReply := struct{}{}

	// Try to send until success
	logMessage := "Sending scan result."
	for {

		// Do not start another request after cancellation
		select {
		case <-ctx.Done():
			logger.Infof("Scan result discarded, RPC client shutting down.")
			return
		default:
		}

		// Log transmission attempt
		logger.Debugf("%s", logMessage)

		// Attempt to send scan result (immediately, without ticker delay)
		errRpc := restoreRpcServerError(rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply)) // rpcReply must be pointer to receive result!
		if errors.Is(errRpc, utils.ErrRpcConnectivity) {

			// Log situation
			logger.Debugf("Waiting for reconnection.")

			// Wait until RPC connection or client shutdown
			select {
			case <-ctx.Done(): // Cancellation signal

				// Discard results if agents shuts down and broker can't receive results
				logger.Infof("Scan result discarded, RPC client shutting down.")
				return

			case _, ok := <-rpc.Established():

				// Stop when the RPC client terminates
				if !ok {
					logger.Infof("Scan result discarded, RPC client shutting down.")
					return
				}

				// Log reconnection
				logger.Debugf("Broker reconnected.")
				logMessage = "Retrying to send scan result."

				// Break case
				break
			}

		} else if errors.Is(errRpc, errInvalidScopeSecret) { // Scope secret is permanently invalid, retrying would not help
			logger.Warningf("Scan result discarded, scope secret is invalid.")
			return

		} else if errRpc != nil {

			// Check if client is shutting down
			select {
			case <-ctx.Done():
				logger.Infof("Scan result discarded, RPC client shutting down.")
				return
			default:
			}

			// Log situation
			logger.Debugf("Waiting for retry.")

			// Broker could not process the request, so delay the next attempt without blocking cancellation
			retryTimer := time.NewTimer(utils.ReconnectInterval)
			select {
			case <-retryTimer.C:
			case <-ctx.Done():
				_ = retryTimer.Stop()
				logger.Infof("Scan result discarded, RPC client shutting down.")
				return
			}

			// Update log message
			logMessage = "Retrying to send scan result."

		} else {
			logger.Debugf("Scan result sent.")
			return
		}
	}
}
