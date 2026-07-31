/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"golang.org/x/sync/semaphore"
)

// ReconnectInterval defines the delay between background RPC connection attempts
const ReconnectInterval = time.Second * 5

// IsRpcConnectionError validates a given error and checks whether it is a kind of error indicating connectivity issues
func IsRpcConnectionError(err error) bool {
	if errors.Is(err, rpc.ErrShutdown) { // Check if RPC connectivity error
		return true
	}
	if IsConnectionError(err) { // Check if socket connectivity error
		return true
	}
	return false // Return false if not connectivity error
}

// Client owns one reconnecting net/rpc connection and its lifecycle
type Client struct {
	serverAddress  string // RPC server address to connect to
	serverSsl      bool   // Whether to establish an SSL connection
	serverCertPath string // Public key of server to check certificate hash
	serverSecret   string // Shared secret sent as a connection handshake to authenticate this client. Empty disables it.

	ctx           context.Context    // Context of this RPC client
	ctxCancelFunc context.CancelFunc // Function to cancel this RPC client's context in order to shut it down

	wg         sync.WaitGroup      // Wait group indicating whether everything is shut down
	lockState  sync.Mutex          // Lock protecting the connection pointers, state, and waiters
	semConnect *semaphore.Weighted // RPC semaphore making sure only one goroutine is initiating a connection
	chNotify   []chan struct{}     // One-shot channels parked by Established() callers until a connection becomes available or the client shuts down

	rpcClient    *rpc.Client        // RPC client connected
	rpcCon       io.ReadWriteCloser // RPC connection
	rpcConnected bool               // Whether the currently published RPC client connected successfully
}

// NewRpcClient prepares an RPC client struct providing connectivity to an RPC server.
func NewRpcClient(
	server string, // Remote RPC host:port to connect to
	serverSsl bool, // Whether to establish an SSL connection
	serverCertPath string, // Remote RPC certificate to validate connection. Fallback, if no publicly valid SSL certificate is deployed.
	serverSecret string, // Shared secret sent once per connection to authorize this client. Pass "" if the server does not require it.
) *Client {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	return &Client{
		serverAddress:  server,
		serverSsl:      serverSsl,
		serverCertPath: serverCertPath,
		serverSecret:   serverSecret,
		ctx:            ctx,
		ctxCancelFunc:  ctxCancelFunc,
		semConnect:     semaphore.NewWeighted(int64(1)),
	}
}

// Connect tries to establish a connection (if no other goroutine does yet). Returns boolean indicating whether a
// connection could be established. Optionally, after the first try, the connection attempt can be continued in the
// background sending a notification to current subscribers on success. Subscribe via "Established()".
func (c *Client) Connect(logger scanUtils.Logger, continueBackground bool) bool {

	// Acquire the lock to prevent Disconnect from racing wg.Wait against this goroutine's wg.Add
	c.lockState.Lock()

	// Register connection work
	if c.ctx.Err() != nil {
		c.lockState.Unlock()
		return false
	}

	// Make shutdown wait (just needed to improve order of remaining log messages)
	c.wg.Add(1)

	// Release lock again
	c.lockState.Unlock()

	// Make sure only one goroutine is trying to initiate a (re)connection at once
	if !c.semConnect.TryAcquire(1) {
		logger.Debugf("Connecting RPC already in progress.")
		c.wg.Done()
		return false // No semaphore acquired
	}

	// Try to connect
	errConnect := c.connect(logger)
	if errConnect != nil {

		// Log failed connection attempt
		logger.Debugf("Connecting RPC failed: %s", errConnect)

		// Return without success if no retry is desired
		if !continueBackground {
			c.semConnect.Release(1) // Release semaphore, function does not continue in the background
			c.wg.Done()
			return false
		}

	} else {

		// Log successful connection attempt
		logger.Debugf("Connecting RPC succeeded.")

		// Return with success
		c.semConnect.Release(1) // Release semaphore again after successful connection
		c.wg.Done()
		return true
	}

	// Keep retrying in background until success or shutdown
	go func() {

		// Release the connection slot and shutdown waiter on every terminal path
		defer c.semConnect.Release(1)
		defer c.wg.Done()

		// Retry at the configured interval until connection or cancellation
		attempts := 2 // First attempt was above
		ticker := time.NewTicker(ReconnectInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C: // Wait for next attempt

				// Retry if connection failed, exit with success otherwise
				errConnect := c.connect(logger)
				if errConnect != nil { // Retry
					logger.Debugf("Connecting RPC failed: %s", errConnect)
				} else { // Return with success
					logger.Debugf("Connecting RPC succeeded after %d attempt(s).", attempts)
					return
				}

			case <-c.ctx.Done(): // Cancellation signal
				logger.Infof("Connecting RPC aborted.")
				return
			}

			// Count attempt
			attempts++
		}
	}()

	// Return false to indicate that currently no connection was available
	// WITHOUT releasing the semaphore, as the function continues in the background
	return false
}

// Established returns a channel that signals once an RPC connection is available
//
// The returned channel signals immediately if the client is already connected, waits for the next
// successful connection otherwise, and closes without a value if the client shuts down
func (c *Client) Established() <-chan struct{} {

	// Prepare memory for establisehd channels
	chEstablished := make(chan struct{}, 1)

	// Acquire the lock that guards connection state and waiters to avoid missing a
	// reconnect between the check and the append
	c.lockState.Lock()

	// Register synchronously
	if c.ctx.Err() != nil {
		close(chEstablished)
	} else if c.rpcConnected {
		chEstablished <- struct{}{}
		close(chEstablished)
	} else {
		c.chNotify = append(c.chNotify, chEstablished)
	}

	// Release lock
	c.lockState.Unlock()

	// Return the one-shot connection notification
	return chEstablished
}

// Disconnect disconnects an RPC connection
func (c *Client) Disconnect() {

	// Close the RPC client context while holding the admission lock. Connect() either registers
	// its work before this cancellation or observes the cancelled context and does not register.
	c.lockState.Lock()
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}
	c.lockState.Unlock()

	// Wait for everything to be shut down
	c.wg.Wait()

	// Acquire the lock to detach all resources atomically so callers see a clean nil state.
	// Closing happens outside the lock to avoid blocking it
	c.lockState.Lock()

	// Keep references to the client, connection, and waiters for closing after the lock is released
	rpcClient := c.rpcClient
	connection := c.rpcCon
	waiters := c.chNotify

	// Detach all connection resources so subsequent callers see a clean state
	c.rpcClient = nil
	c.rpcCon = nil
	c.rpcConnected = false
	c.chNotify = nil

	// Release the lock before performing any blocking close operations
	c.lockState.Unlock()

	// Close all pending Established() waiters without sending a value so callers observe shutdown
	for _, waiter := range waiters {
		close(waiter)
	}

	// Close the detached RPC resources
	if rpcClient != nil {
		_ = rpcClient.Close()
	}
	if connection != nil {
		_ = connection.Close()
	}
}

// Call executes an RPC call and writes back the response into rpcReply. In case of a connection issues, a
// reconnection loop is started in the background and an error returned. Users can subscribe to and wait for a
// reconnection notification (via Established()), or just retry later.
//
// This function returns:
//   - nil, if everything went fine
//   - ErrRpcConnectivity, if something was wrong with the RPC connection
//   - A restored common RPC sentinel error, if the RPC server returned one defined by this package
//   - rpc.ServerError, if the RPC request failed with an error owned by the called component or an unknown error
func (c *Client) Call(logger scanUtils.Logger, rpcEndpoint string, rpcArgs interface{}, rpcReply interface{}) error {

	// Send RPC request
	logger.Debugf("Sending RPC request '%s'.", rpcEndpoint)

	// Acquire the lock to read the current RPC client without racing connect() or Disconnect()
	c.lockState.Lock()
	rpcClient := c.rpcClient
	c.lockState.Unlock()

	// Execute the RPC call, treating a nil client the same as a closed connection
	var errCall error
	if rpcClient == nil { // Seems RPC client wasn't connected before
		errCall = rpc.ErrShutdown
	} else {
		errCall = rpcClient.Call(rpcEndpoint, rpcArgs, rpcReply)
	}

	// Retry with a quick reconnection (RPC service might just have restarted)
	if IsRpcConnectionError(errCall) {
		connectionFailed := c.markConnectionUnavailable(rpcClient)

		// Reuse a connection another caller already published instead of replacing that healthy connection
		if !connectionFailed {

			// Read the current RPC client under the lock to avoid a data race with connect() and Disconnect()
			c.lockState.Lock()
			rpcClient = c.rpcClient
			c.lockState.Unlock()

			// Execute the RPC call, treating a nil client the same as a closed connection
			if rpcClient == nil {
				errCall = rpc.ErrShutdown
			} else {
				errCall = rpcClient.Call(rpcEndpoint, rpcArgs, rpcReply)
			}

		} else {
			select {
			case <-c.ctx.Done():
				logger.Infof("RPC call aborted due to shutdown.")
				return ErrRpcConnectivity
			default:

				// Try to reconnect and launch background reconnection if the immediate attempt fails
				success := c.Connect(logger, true)
				if success {

					// Read the current RPC client under the lock to avoid a data race with connect() and Disconnect()
					c.lockState.Lock()
					rpcClient = c.rpcClient
					c.lockState.Unlock()

					// Execute the RPC call, treating a nil client the same as a closed connection
					if rpcClient == nil {
						errCall = rpc.ErrShutdown
					} else {
						errCall = rpcClient.Call(rpcEndpoint, rpcArgs, rpcReply) // Retry RPC call
					}

				} else {
					return ErrRpcConnectivity
				}
			}
		}
	}

	// Restore sentinel identity lost when net/rpc converted the server-side error to rpc.ServerError
	errCall = restoreRpcServerError(errCall)

	// Handle ultimate RPC call success or error
	if IsRpcConnectionError(errCall) {
		if c.markConnectionUnavailable(rpcClient) {
			_ = c.Connect(logger, true)
		}
		logger.Infof("RPC connection lost.")
		return ErrRpcConnectivity // Return connectivity error
	} else if errors.Is(errCall, ErrRpcCompatibility) {
		logger.Errorf("RPC client incompatible.")
		return errCall
	} else if errCall != nil {
		logger.Warningf("RPC request '%s' failed: %s", rpcEndpoint, errCall)
		return errCall // Component-specific and unknown server errors remain rpc.ServerError values
	} else {
		logger.Debugf("RPC request '%s' succeeded.", rpcEndpoint)
		return nil
	}
}

// connect executes a single connection attempt and returns an error if something went wrong.
// ATTENTION: This may only be called after successfully acquiring the semaphore c.semConnect
func (c *Client) connect(logger scanUtils.Logger) error {

	// Acquire the lock to detach the previous resources atomically so callers see a clean nil state.
	// Closing happens outside the lock to avoid blocking it
	c.lockState.Lock()

	// Keep references to the previous client and connection for closing after the lock is released
	previousClient := c.rpcClient
	previousConnection := c.rpcCon

	// Detach the previous connection resources so subsequent callers see a clean state
	c.rpcClient = nil
	c.rpcCon = nil
	c.rpcConnected = false

	// Release the lock before performing any blocking close operations
	c.lockState.Unlock()

	// Close the detached RPC resources
	if previousClient != nil {
		_ = previousClient.Close()
	}
	if previousConnection != nil {
		_ = previousConnection.Close()
	}

	// Log step
	logger.Debugf("Connecting RPC.")

	// Prepare memory for connection
	var conn io.ReadWriteCloser
	var connErr error

	// Create SSL or plaintext connection, depending on setting
	if c.serverSsl {

		// Try to connect to RPC server with full certificate verification
		// Maybe a fully valid public SSL certificate was deployed
		tlsDialer := tls.Dialer{Config: TlsConfigFactory()}
		conn, connErr = tlsDialer.DialContext(c.ctx, "tcp", c.serverAddress)
		if connErr != nil {

			// Log step
			logger.Debugf("Connecting RPC with pinned certificate.")

			// Get fingerprint-verifying tls config
			tlsConfig, errTlsConfig := TlsConfigFactoryPinned(c.serverCertPath)
			if errTlsConfig != nil {
				logger.Errorf("Could not prepare pinned certificate: %s", errTlsConfig)
				return errTlsConfig
			}

			// Try with pinned fingerprint alternatively, there might be a self-signed certificate deployed
			tlsDialer.Config = tlsConfig
			conn, connErr = tlsDialer.DialContext(c.ctx, "tcp", c.serverAddress)
			if connErr != nil {
				return connErr
			}
		}
	} else {

		// Try to connect to RPC server with a cancelable plaintext connection
		dialer := net.Dialer{}
		conn, connErr = dialer.DialContext(c.ctx, "tcp", c.serverAddress)
		if connErr != nil {
			return connErr
		}
	}

	// Authenticate the connection by sending the shared secret (as a single gob value) before net/rpc takes over, if
	// one is configured. Clients without a secret (e.g. scan agents authenticating per scope) send no handshake.
	if c.serverSecret != "" {
		if errAuth := gob.NewEncoder(conn).Encode(c.serverSecret); errAuth != nil {
			_ = conn.Close()
			return errAuth
		}
	}

	rpcClient := rpc.NewClient(conn)

	// Acquire the lock to publish new resources and detach waiters atomically so no Established()
	// caller can miss the transition; also guards the context check to abort cleanly if shutdown
	// was requested while dialing
	c.lockState.Lock()

	// Collect pending waiters before potentially aborting, so they can be notified outside the lock
	waiters := c.chNotify

	// Abort if termination was requested while the dial was in progress
	if c.ctx.Err() != nil {
		c.lockState.Unlock()
		_ = rpcClient.Close()
		_ = conn.Close()
		return c.ctx.Err()
	}

	// Attach the new connection resources and clear the waiters slice so connect() owns notification
	c.rpcClient = rpcClient
	c.rpcCon = conn
	c.rpcConnected = true
	c.chNotify = nil
	c.lockState.Unlock()

	// Notify every caller that was waiting for a usable connection
	for _, waiter := range waiters {
		waiter <- struct{}{}
		close(waiter)
	}

	// Return nil to indicate connection success
	return nil
}

// markConnectionUnavailable marks a failed published RPC client as disconnected and reports whether it was current
func (c *Client) markConnectionUnavailable(failedClient *rpc.Client) bool {

	// Acquire the lock to check and update the published client atomically, and make
	// sure the lock is released again
	c.lockState.Lock()
	defer c.lockState.Unlock()

	// Ignore late failures from a connection that another goroutine already replaced
	if c.rpcClient == failedClient {
		c.rpcConnected = false
		return true
	}
	return false
}
