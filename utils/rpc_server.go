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
	scanUtils "github.com/siemens/GoScans/utils"
	"net"
	"net/rpc"
	"strings"
)

// ServeRpc loops to accept and process RPC connections until the passed context is terminated
func ServeRpc(
	logger scanUtils.Logger,
	ctx context.Context,
	name string, // Just a name printed in the log message
	certFilePath string,
	keyFilePath string,
	listenAddress string,
) error {

	// Data transfer struct for connection attempts
	type connAttempt struct {
		conn net.Conn
		err  error
	}

	// Initialize connection variables
	var chConnAttempts = make(chan connAttempt, 1)

	// Open listening socket
	socket, errSocket := SslSocket(listenAddress, certFilePath, keyFilePath)
	if errSocket != nil {
		return errSocket
	} else {
		logger.Infof("%s RPC is listening on '%s'.", strings.Title(name), listenAddress)
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

				// Serve connection
				rpc.ServeConn(attempt.conn)
			}()

		case <-ctx.Done():
			return nil // Just return if core is shutting down
		}
	}
}
