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
	"encoding/gob"
	"errors"
	"io"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
)

// rpcAuthTimeout is the time a connection has to complete the authentication handshake before it is dropped.
const rpcAuthTimeout = 5 * time.Second

var ErrRpcGeneric = errors.New("RPC endpoint not available")    // ErrRpcGeneric indicates an unavailable RPC endpoint without exposing sensitive details.
var ErrRpcConnectivity = errors.New("RPC server not available") // ErrRpcConnectivity indicates a temporary or already resolved RPC connectivity problem.
var ErrRpcCompatibility = errors.New("RPC client incompatible") // ErrRpcCompatibility indicates that the RPC client and server versions are incompatible.

// restoreRpcServerError restores common RPC sentinel errors whose concrete type was lost by net/rpc. Client.Call must
// pass every net/rpc result through this function. Keep this switch synchronized with the common errors declared above.
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
	case ErrRpcGeneric.Error():
		return ErrRpcGeneric
	case ErrRpcConnectivity.Error():
		return ErrRpcConnectivity
	case ErrRpcCompatibility.Error():
		return ErrRpcCompatibility
	default:

		// Return unknown error
		return err
	}
}

// byteReader adapts a connection to an io.ByteReader without buffering. gob.NewDecoder wraps any reader that is not
// already an io.ByteReader in a bufio.Reader, which reads ahead and would swallow the start of the following net/rpc
// stream. An unbuffered ByteReader makes gob consume exactly the handshake message.
type byteReader struct{ io.Reader }

func (b byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, errReadFull := io.ReadFull(b, buf[:])
	return buf[0], errReadFull
}

// ServeRpc loops to accept and process RPC connections until the passed context is terminated. If listenSecrets is
// non-empty, every connection must present one of those secrets (connection handshake) before it is served. An empty
// listenSecrets accepts any handshake. Context cancellation stops the listener but leaves established connections
// alive so their in-flight calls can complete; the owning core rejects later calls during shutdown.
func ServeRpc(
	logger scanUtils.Logger,
	ctx context.Context,
	name string, // Just a name printed in the log message
	certFilePath string,
	keyFilePath string,
	listenAddress string,
	listenSecrets []string,
) error {

	// Data transfer struct for connection attempts
	type connAttempt struct {
		conn net.Conn
		err  error
	}

	// Initialize connection variables
	var acceptWg sync.WaitGroup
	var chConnAttempts = make(chan connAttempt, 1)

	// Convert wildcard notation to suitable variant
	if strings.HasPrefix(listenAddress, "*:") {
		listenAddress = strings.Replace(listenAddress, "*:", ":", 1)
	}

	// Open listening socket
	var socket net.Listener
	var errSocket error
	if certFilePath != "" && keyFilePath != "" {
		logger.Infof("Opening SSL socket for %s RPC.", strings.ToUpper(name[:1])+name[1:])
		socket, errSocket = SslSocket(listenAddress, certFilePath, keyFilePath) // Return SSL socket if SSL keys are set
	} else {
		logger.Infof("Opening PLAIN socket for %s RPC.", strings.ToUpper(name[:1])+name[1:])
		socket, errSocket = net.Listen("tcp", listenAddress) // Return plain socket if SSL keys are not set
	}

	// Check result and log
	if errSocket != nil {
		return errSocket
	} else {
		logger.Infof("%s RPC is listening on '%s'.", strings.ToUpper(name[:1])+name[1:], listenAddress)
	}

	// Closing the listener releases a pending Accept without interrupting established RPC connections
	defer func() {

		// Stop accepting and wait until the single accept goroutine has published its result
		_ = socket.Close()
		acceptWg.Wait()

		// Accept may have completed at the same time as cancellation. Close a connection that reached the attempt
		// channel but was not selected and therefore was never served
		select {
		case attempt := <-chConnAttempts:
			if attempt.conn != nil {
				_ = attempt.conn.Close()
			}
		default:
		}
	}()

	// Loop until termination signal
	for {

		// Listen for connection asynchronously
		acceptWg.Add(1)
		go func() {

			// Publish exactly one accept result for the main loop or shutdown cleanup
			defer acceptWg.Done()
			connection, errConnection := socket.Accept()
			chConnAttempts <- connAttempt{connection, errConnection}
		}()

		// Wait for connection or shutdown
		select {
		case attempt := <-chConnAttempts:

			// Reject a connection accepted concurrently with shutdown
			if ctx.Err() != nil {
				if attempt.conn != nil {
					_ = attempt.conn.Close()
				}
				return nil
			}

			// Check for connection error
			if attempt.err != nil {
				logger.Warningf("Could not accept connection: %s", attempt.err)
				continue // Retry if something else went wrong
			}

			// Handle connection asynchronously
			go func(connection net.Conn) {

				// Make sure connection gets closed on exit
				defer func() { _ = connection.Close() }()

				// When no listen secret configured no handshake is read at all, so clients that send none
				// (e.g. scan agents authenticating per scope, or older agents) stay compatible.
				if len(listenSecrets) > 0 {

					// Prepare connection for authorization handshake
					_ = connection.SetReadDeadline(time.Now().Add(rpcAuthTimeout))

					// Read the client's secret, sent as a single gob value, through an unbuffered byteReader so gob
					// consumes exactly the handshake and the following net/rpc stream stays intact.
					var secret string
					var errAuth = gob.NewDecoder(byteReader{connection}).Decode(&secret)

					// Reject the connection if the handshake failed or the secret is not one of the configured listen
					// secrets. Otherwise, all multiplexed calls on it are bound to this authorized peer (no per-method check).
					if errAuth != nil || !scanUtils.StrContained(secret, listenSecrets) {
						return // Reject silently to avoid log spam from heartbeat/health-check connections
					}

					// Clear deadline before net/rpc takes over
					_ = connection.SetReadDeadline(time.Time{})
				}

				// Serve connection
				rpc.ServeConn(connection)
			}(attempt.conn)

		case <-ctx.Done():
			return nil // Just return if core is shutting down
		}
	}
}
