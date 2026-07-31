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
	"sync"
	"testing"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
)

// EchoArgs drives a minimal net/rpc round trip in the integration test
type EchoArgs struct{ Message string }

// EchoReply contains the echoed RPC message
type EchoReply struct{ Message string }

// EchoService is a minimal net/rpc receiver echoing its input to exercise a real connection
type EchoService struct{}

// Echo returns the supplied message
func (EchoService) Echo(args *EchoArgs, reply *EchoReply) error {

	// Copy the request message into the response
	reply.Message = args.Message
	return nil
}

// BlockingRpcService exposes a method that remains in flight until its test releases it
type BlockingRpcService struct {
	lock      sync.Mutex
	chStarted chan struct{}
	chRelease chan struct{}
}

// Block waits until the test releases the in-flight RPC method
func (service *BlockingRpcService) Block(_ *struct{}, _ *struct{}) error {

	// Snapshot the channels configured for this test invocation
	service.lock.Lock()
	chStarted := service.chStarted
	chRelease := service.chRelease
	service.lock.Unlock()
	close(chStarted)
	<-chRelease
	return nil
}

var blockingRpcService = &BlockingRpcService{}
var registerRpcTestServicesOnce sync.Once
var errRegisterRpcTestServices error

// registerRpcTestServices registers test services once on the process-global default RPC server
func registerRpcTestServices(t *testing.T) {

	// Register shared services only once because go test may repeat tests in the same process
	t.Helper()
	registerRpcTestServicesOnce.Do(func() {
		errRegisterRpcTestServices = rpc.Register(EchoService{})
		if errRegisterRpcTestServices == nil {
			errRegisterRpcTestServices = rpc.RegisterName("BlockingRpcService", blockingRpcService)
		}
	})
	if errRegisterRpcTestServices != nil {
		t.Fatalf("Could not register RPC test services: '%v'", errRegisterRpcTestServices)
	}
}

// freeRpcAddr returns a currently free localhost TCP address for the test server to bind to.
func freeRpcAddr(t *testing.T) string {
	listener, errListener := net.Listen("tcp", "127.0.0.1:0")
	if errListener != nil {
		t.Fatalf("could not allocate free port: %v", errListener)
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

	// Register the services on the default RPC server
	registerRpcTestServices(t)

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

// TestServeRpcShutdownPreservesActiveConnection verifies listener shutdown does not interrupt established clients
func TestServeRpcShutdownPreservesActiveConnection(t *testing.T) {

	// Prepare unit test data
	registerRpcTestServices(t)
	logger := scanUtils.NewTestLogger()
	addr := freeRpcAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	chServerDone := make(chan error, 1)
	go func() { chServerDone <- ServeRpc(logger, ctx, "Test", "", "", addr, nil) }()

	// Connect and complete a call to prove the connection was accepted before shutdown
	var rpcClient *rpc.Client
	deadline := time.Now().Add(2 * time.Second)
	for rpcClient == nil && time.Now().Before(deadline) {
		candidate, errCandidate := rpc.Dial("tcp", addr)
		if errCandidate == nil {
			rpcClient = candidate
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if rpcClient == nil {
		cancel()
		t.Fatal("could not connect to RPC server")
	}
	defer func() { _ = rpcClient.Close() }()
	var reply EchoReply
	if errCall := rpcClient.Call("EchoService.Echo", &EchoArgs{Message: "before"}, &reply); errCall != nil {
		cancel()
		t.Fatalf("RPC call before listener shutdown failed: %v", errCall)
	}

	// Canceling the server must stop the listener without closing the established connection
	cancel()
	select {
	case errServe := <-chServerDone:
		if errServe != nil {
			t.Fatalf("ServeRpc() error = %v, want nil", errServe)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ServeRpc() did not return after cancellation")
	}

	reply = EchoReply{}
	if errCall := rpcClient.Call("EchoService.Echo", &EchoArgs{Message: "after"}, &reply); errCall != nil {
		t.Fatalf("RPC call on established connection after listener shutdown failed: %v", errCall)
	}
	if reply.Message != "after" {
		t.Fatalf("RPC echo after listener shutdown = %q, want %q", reply.Message, "after")
	}

	newConnection, errNewConnection := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if errNewConnection == nil {
		_ = newConnection.Close()
		t.Fatal("new connection succeeded after listener shutdown")
	}
}

// TestServeRpcShutdownDoesNotWaitForInFlightMethod verifies a blocked method cannot hold server shutdown open
func TestServeRpcShutdownDoesNotWaitForInFlightMethod(t *testing.T) {

	// Prepare a blocking service invocation on the default RPC server
	registerRpcTestServices(t)
	chStarted := make(chan struct{})
	chRelease := make(chan struct{})
	blockingRpcService.lock.Lock()
	blockingRpcService.chStarted = chStarted
	blockingRpcService.chRelease = chRelease
	blockingRpcService.lock.Unlock()
	logger := scanUtils.NewTestLogger()
	address := freeRpcAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	chServerDone := make(chan error, 1)
	go func() { chServerDone <- ServeRpc(logger, ctx, "Test", "", "", address, nil) }()

	// Connect after the listener becomes available
	var rpcClient *rpc.Client
	deadline := time.Now().Add(2 * time.Second)
	for rpcClient == nil && time.Now().Before(deadline) {
		client, errClient := rpc.Dial("tcp", address)
		if errClient == nil {
			rpcClient = client
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if rpcClient == nil {
		cancel()
		t.Fatal("Could not connect to RPC server")
	}
	defer func() { _ = rpcClient.Close() }()

	// Start a method that stays blocked while listener shutdown completes
	call := rpcClient.Go("BlockingRpcService.Block", &struct{}{}, &struct{}{}, nil)
	select {
	case <-chStarted:
	case <-time.After(2 * time.Second):
		cancel()
		close(chRelease)
		t.Fatal("Blocking RPC method did not start")
	}

	// Verify server shutdown returns before the RPC method is released
	cancel()
	select {
	case errServe := <-chServerDone:
		if errServe != nil {
			t.Fatalf("ServeRpc() error = '%v', want nil", errServe)
		}
	case <-time.After(2 * time.Second):
		close(chRelease)
		t.Fatal("ServeRpc() waited for the in-flight RPC method")
	}

	// Release and join the method so the test leaves no goroutine behind
	close(chRelease)
	select {
	case <-call.Done:
	case <-time.After(2 * time.Second):
		t.Fatal("In-flight RPC method did not exit after release")
	}
}
