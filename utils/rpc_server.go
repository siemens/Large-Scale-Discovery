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
	"fmt"
	"io"
	"net"
	"net/rpc"
	"strings"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
)

// rpcAuthTimeout is the time a connection has to complete the authentication handshake before it is dropped.
const rpcAuthTimeout = 5 * time.Second

var ErrRpcGeneric = fmt.Errorf("RPC endpoint not available")    // Generic error returned by the RPC client, which may not contain sensitive details
var ErrRpcConnectivity = fmt.Errorf("RPC server not available") // Error indicating connectivity error which might be gone already or temporary
var ErrRpcCompatibility = fmt.Errorf("RPC client incompatible") // Error indicating that the RPC client is incompatible and needs to be updated

// byteReader adapts a connection to an io.ByteReader without buffering. gob.NewDecoder wraps any reader that is not
// already an io.ByteReader in a bufio.Reader, which reads ahead and would swallow the start of the following net/rpc
// stream. An unbuffered ByteReader makes gob consume exactly the handshake message.
type byteReader struct{ io.Reader }

func (b byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(b, buf[:])
	return buf[0], err
}

// ServeRpc loops to accept and process RPC connections until the passed context is terminated. If listenSecrets is
// non-empty, every connection must present one of those secrets (connection handshake) before it is served. An empty
// listenSecrets accepts any handshake.
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

	// Make sure socket is closed on exit
	defer func() { _ = socket.Close() }()

	// Loop until termination signal
	for {

		// Listen for connection asynchronously
		go func() {
			conn, err := socket.Accept()
			chConnAttempts <- connAttempt{conn, err}
		}()

		// Wait for connection or shutdown
		select {
		case attempt := <-chConnAttempts:

			// Check for connection error
			if attempt.err != nil {
				logger.Warningf("Could not accept connection: %s", attempt.err)
				continue // Retry if something else went wrong
			}

			// Handle connection asynchronously
			go func() {

				// Make sure connection gets closed on exit
				defer func() { _ = attempt.conn.Close() }()

				// When no listen secret configured no handshake is read at all, so clients that send none
				// (e.g. scan agents authenticating per scope, or older agents) stay compatible.
				if len(listenSecrets) > 0 {

					// Prepare connection for authorization handshake
					_ = attempt.conn.SetReadDeadline(time.Now().Add(rpcAuthTimeout))

					// Read the client's secret, sent as a single gob value, through an unbuffered byteReader so gob
					// consumes exactly the handshake and the following net/rpc stream stays intact.
					var secret string
					var errAuth = gob.NewDecoder(byteReader{attempt.conn}).Decode(&secret)

					// Reject the connection if the handshake failed or the secret is not one of the configured listen
					// secrets. Otherwise all multiplexed calls on it are bound to this authorized peer (no per-method check).
					if errAuth != nil || !scanUtils.StrContained(secret, listenSecrets) {
						return // Reject silently to avoid log spam from heartbeat/health-check connections
					}

					// Clear deadline before net/rpc takes over
					_ = attempt.conn.SetReadDeadline(time.Time{})
				}

				// Serve connection
				rpc.ServeConn(attempt.conn)
			}()

		case <-ctx.Done():
			return nil // Just return if core is shutting down
		}
	}
}
