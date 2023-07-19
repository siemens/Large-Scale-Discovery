/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"golang.org/x/sync/semaphore"
	"net/rpc"
	"sync"
	"time"
)

const ReconnectInterval = time.Second * 5

var ErrRpcConnectivity = fmt.Errorf("RPC connectivity error")

// IsRpcConnectionError validates a given error and checks whether it is a kind of error indicating connectivity issues
func IsRpcConnectionError(err error) bool {
	if err == rpc.ErrShutdown { // Check if RPC connectivity error
		return true
	}
	if IsConnectionError(err) { // Check if socket connectivity error
		return true
	}
	return false // Return false if not connectivity error
}

type Client struct {
	chNotify       chan chan struct{}  // A channel of channels waiting for a signal after a connection was established
	server         string              // RPC server to connect to
	serverCertPath string              // Public key of server to check certificate hash
	ctx            context.Context     // Context of this RPC client
	ctxCancelFunc  context.CancelFunc  // Function to cancel this RPC client's context in order to shut it down
	sem            *semaphore.Weighted // RPC semaphore making sure only one goroutine is initiating a connection
	con            *tls.Conn           // RPC connection
	client         *rpc.Client         // RPC client connected
	wg             sync.WaitGroup      // Wait group indicating whether everything is shut down
}

// NewRpcClient prepares an RPC client struct providing connectivity to an RPC server.
func NewRpcClient(
	server string, // Remote RPC host:port to connect to
	serverCertPath string, // Remote RPC certificate to validate connection
) *Client {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	return &Client{
		chNotify:       make(chan chan struct{}),
		server:         server,
		serverCertPath: serverCertPath,
		ctx:            ctx,
		ctxCancelFunc:  ctxCancelFunc,
		sem:            semaphore.NewWeighted(int64(1)),
	}
}

// Connect tries to establish a connection (if no other goroutine does yet). Returns boolean indicating whether a
// connection could be established. Optionally, after the first try, the connection attempt can be continued in the
// background sending a notification to current subscribers on success. Subscribe via "Established()".
func (c *Client) Connect(logger scanUtils.Logger, continueBackground bool) bool {

	// Make shutdown wait (just needed to improve order of remaining log messages)
	c.wg.Add(1)
	defer c.wg.Done()

	// Make sure only one goroutine is trying to initiate a (re)connection at once
	if !c.sem.TryAcquire(1) {
		logger.Debugf("Connecting RPC already in progress.")
		return false
	}

	// Log step
	logger.Debugf("Connecting RPC.")

	// Try to connect
	err := c.connect(logger)
	if err != nil {

		// Continue re-connection in background if desired
		if continueBackground {

			// Keep retrying in background until success or shutdown
			c.wg.Add(1) // Make shutdown wait (just needed to improve order of remaining log messages)
			go func() {
				defer c.wg.Done()
				ticker := time.NewTicker(ReconnectInterval)
				for {
					select {
					case <-ticker.C: // Wait for next attempt
						errConnect := c.connect(logger)
						if errConnect == nil {
							c.sem.Release(1) // Release semaphore, function does not continue in the background
							return           // Return if connection succeeded, otherwise retry.
						}
					case <-c.ctx.Done(): // Cancellation signal
						logger.Infof("Connecting RPC aborted.")
						c.sem.Release(1) // Release semaphore, function does not continue in the background
						return
					}
				}
			}()

			// Return false to indicate that currently no connection was available
			// WITHOUT releasing the semaphore, as the function continues in the background
			return false
		} else {
			// Return false to indicate that currently no connection was available
			c.sem.Release(1) // Release semaphore, function does not continue in the background
			return false
		}
	}

	// Return true to indicate successful connection
	c.sem.Release(1) // Release semaphore, function does not continue in the background
	return true
}

// Established returns a notification channel triggering when an RPC connection was (re)established. This can be
// used to block after connections issues until it re-connection. The channel will not trigger, if a connection
// is currently established and working!
func (c *Client) Established() <-chan struct{} {
	chNotify := make(chan struct{})
	go func() {
		c.chNotify <- chNotify
	}()
	return chNotify
}

// Disconnect disconnects an RPC connection
func (c *Client) Disconnect() {

	// Close RPC client context to shut it down
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}

	// Wait for everything to be shut down
	c.wg.Wait()

	// Close RPC client
	if c.client != nil {
		_ = c.client.Close()
	}

	// Close RPC connection
	if c.con != nil {
		_ = c.con.Close()
	}
}

// Call executes an RPC call and writes back the response into rpcReply. In case of a connection issues, a
// re-connection loop is started in the background and an error returned. Users can subscribe to and wait for a
// re-connection notification (via Established()), or just retry later.
//
// This function returns:
// 		- nil, if everything went fine
//		- ErrRpcConnectivity, if something was wrong with the RPC connection
//		- rpc.ServerError, if the RPC request failed. Unfortunately, net/rpc converts all errors into this type.
//
func (c *Client) Call(logger scanUtils.Logger, rpcEndpoint string, rpcArgs interface{}, rpcReply interface{}) error {

	// Send RPC request
	logger.Debugf("Sending RPC request '%s'.", rpcEndpoint)
	var errCall error
	if c.client == nil { // Seems RPC client wasn't connected before
		errCall = rpc.ErrShutdown
	} else {
		errCall = c.client.Call(rpcEndpoint, rpcArgs, rpcReply)
	}

	// Retry with quick re-connection (RPC service might just have restarted)
	if IsRpcConnectionError(errCall) {

		select {
		case <-c.ctx.Done():
			logger.Infof("RPC call aborted due to shutdown.")
			return ErrRpcConnectivity
		default:

			// Try to re-connect and launch background re-connection loop if re-connection failed
			// An active and continuous background routine trying to re-connect will speed up all other queries, as
			// those do not need to attempt to connect and can return (~a second!) faster.
			success := c.Connect(logger, true)
			if success {
				errCall = c.client.Call(rpcEndpoint, rpcArgs, rpcReply) // Retry RPC call
			} else {
				return ErrRpcConnectivity
			}
		}
	}

	// Handle ultimate RPC call success or error
	if IsRpcConnectionError(errCall) {
		logger.Infof("RPC connection lost.")
		return ErrRpcConnectivity
	} else if errCall != nil {
		logger.Warningf("RPC request '%s' failed: %s", rpcEndpoint, errCall)
		return errCall // ATTENTION: this error will have lost it's original type and be rpc.ServerError!
	} else {
		logger.Debugf("RPC request '%s' succeeded.", rpcEndpoint)
		return nil
	}
}

// connect executes a single connection attempt and returns an error if something went wrong.
// ATTENTION: This may only be called after successfully acquiring the semaphore c.sem
func (c *Client) connect(logger scanUtils.Logger) error {

	// Close previous RPC connection (relevant for re-connects)
	if c.client != nil {
		_ = c.client.Close()
	}
	if c.con != nil {
		_ = c.con.Close()
	}

	// Get fingerprint-verifying tls config
	tlsConfig, errConfig := PinnedTlsConfigFactory(c.serverCertPath)
	if errConfig != nil {
		logger.Warningf("Connecting RPC failed due to TLS issues: %s", errConfig)
		return errConfig
	}

	// Try to connect to RPC server
	conn, errDial := tls.Dial("tcp", c.server, tlsConfig)
	if errDial != nil {
		logger.Debugf("Connecting RPC failed: %s", errDial)
		return errDial
	}

	// Assign connection to agent
	c.con = conn

	// Create RPC client
	c.client = rpc.NewClient(conn)

	// Notify subscribers about successful (re)connection
	c.notify()

	// Return nil to indicate connection success
	logger.Infof("Connecting RPC successful.")
	return nil
}

// notify sends notification to registered subscribers waiting for a successful re-connection.
func (c *Client) notify() {
	for {
		select {
		case chSignal := <-c.chNotify:
			chSignal <- struct{}{}
			close(chSignal)
		default:
			return
		}
	}
}
