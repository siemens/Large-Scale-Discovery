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
	"net"
	"net/rpc"
	"testing"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
)

// EchoArgs and EchoReply drive a minimal net/rpc round trip in the integration test.
type EchoArgs struct{ Message string }
type EchoReply struct{ Message string }

// EchoService is a minimal net/rpc receiver echoing its input, used to exercise a real RPC call over a real connection.
type EchoService struct{}

func (EchoService) Echo(args *EchoArgs, reply *EchoReply) error {
	reply.Message = args.Message
	return nil
}

// freeRpcAddr returns a currently free localhost TCP address for the test server to bind to.
func freeRpcAddr(t *testing.T) string {
	listener, errListen := net.Listen("tcp", "127.0.0.1:0")
	if errListen != nil {
		t.Fatalf("could not allocate free port: %v", errListen)
	}
	addr := listener.Addr().String()
	_ = listener.Close()
	return addr
}

// TestServeRpcConnectionAuth drives a real RPC call over a real TCP connection through the connection handshake. It is
// the end-to-end counterpart to TestConnectionHandshake: a successful call proves the handshake does not corrupt the
// following net/rpc stream on a real net.Conn (the case the in-memory test cannot reach). It also verifies that a wrong
// secret is rejected and that an empty listen secret accepts any client (the broker case).
func TestServeRpcConnectionAuth(t *testing.T) {

	// Register the echo service once on the default RPC server
	_ = rpc.Register(EchoService{})

	// Get a test logger
	logger := scanUtils.NewTestLogger()

	// Prepare and run test cases
	tests := []struct {
		name          string
		listenSecrets []string
		clientSecret  string
		wantCall      bool // Whether the RPC call is expected to succeed
	}{
		{name: "matching-secret", listenSecrets: []string{"top_secret"}, clientSecret: "top_secret", wantCall: true},
		{name: "wrong-secret", listenSecrets: []string{"top_secret"}, clientSecret: "nope", wantCall: false},
		{name: "no-handshake", listenSecrets: nil, clientSecret: "", wantCall: true}, // Broker / old-agent case: no handshake either side
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Start a plain (no TLS) RPC server on a free port
			addr := freeRpcAddr(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() { _ = ServeRpc(logger, ctx, "Test", "", "", addr, tt.listenSecrets) }()

			// Connect a client with the test secret, retrying until the server is listening
			client := NewRpcClient(addr, false, "", tt.clientSecret)
			defer client.Disconnect()
			var connected bool
			for i := 0; i < 100; i++ {
				if client.Connect(logger, false) {
					connected = true
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if !connected {
				t.Fatalf("could not connect RPC client")
			}

			// Execute a real RPC call and check the outcome against expectation
			rpcReply := EchoReply{}
			errCall := client.Call(logger, "EchoService.Echo", &EchoArgs{Message: "ping"}, &rpcReply)
			if tt.wantCall {
				if errCall != nil {
					t.Fatalf("RPC call failed: %v", errCall)
				}
				if rpcReply.Message != "ping" {
					t.Fatalf("RPC echo mismatch: got %q, want %q", rpcReply.Message, "ping")
				}
			} else if errCall == nil {
				t.Fatalf("RPC call succeeded but should have been rejected")
			}
		})
	}
}
