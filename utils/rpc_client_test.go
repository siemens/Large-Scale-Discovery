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
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"testing"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
)

// RpcErrorTestService returns configured errors to exercise net/rpc server-error transport.
type RpcErrorTestService struct {
	errors map[string]error
}

// Return returns the error configured for the requested message.
func (s *RpcErrorTestService) Return(message string, _ *struct{}) error {
	return s.errors[message]
}

// startRpcTestServer starts a custom net/rpc server on a free port and returns
// the listening address together with a teardown function. Each call creates an
// independent server instance so tests cannot pollute the global default server.
func startRpcTestServer(t *testing.T, services ...interface{}) (addr string, teardown func()) {
	t.Helper()

	// Register provided services on a fresh, isolated server instance
	srv := rpc.NewServer()
	for _, svc := range services {
		if errRegister := srv.Register(svc); errRegister != nil {
			t.Fatalf("startRpcTestServer: register: '%v'", errRegister)
		}
	}

	// Listen on a free ephemeral port
	ln, errLn := net.Listen("tcp", "127.0.0.1:0")
	if errLn != nil {
		t.Fatalf("startRpcTestServer: listen: '%v'", errLn)
	}

	// Track accepted connections so teardown can force-close them and sever live TCP sockets
	var connMu sync.Mutex
	var accepted []net.Conn

	// Accept connections in the background and register each one for teardown tracking
	go func() {
		for {
			conn, errConn := ln.Accept()
			if errConn != nil {
				return
			}
			connMu.Lock()
			accepted = append(accepted, conn)
			connMu.Unlock()
			go srv.ServeConn(conn)
		}
	}()

	// Return the listener address and a teardown that closes the listener and all accepted connections
	return ln.Addr().String(), func() {
		_ = ln.Close()
		connMu.Lock()
		defer connMu.Unlock()
		for _, conn := range accepted {
			_ = conn.Close()
		}
	}
}

// connectRpcClient blocks until the client connects or the retry limit is exceeded.
func connectRpcClient(t *testing.T, client *Client, logger scanUtils.Logger) {
	t.Helper()

	// Retry with short pauses until the server is ready or the limit is reached
	for i := 0; i < 200; i++ {
		if client.Connect(logger, false) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("connectRpcClient: could not connect within timeout")
}

// TestRpcClient_Call_Success_DoesNotCloseConnection verifies that a successful RPC
// call leaves the connection open so that subsequent calls on the same client succeed.
func TestRpcClient_Call_Success_DoesNotCloseConnection(t *testing.T) {

	// Prepare unit test data
	addr, teardown := startRpcTestServer(t, EchoService{})
	defer teardown()

	logger := scanUtils.NewTestLogger()
	client := NewRpcClient(addr, false, "", "")
	defer client.Disconnect()
	connectRpcClient(t, client, logger)

	// Prepare and run test cases
	for _, msg := range []string{"ping", "pong"} {
		var reply EchoReply
		errCall := client.Call(logger, "EchoService.Echo", &EchoArgs{Message: msg}, &reply)

		// Verify
		if errCall != nil {
			t.Fatalf("Call() error = '%v', want nil (msg=%q)", errCall, msg)
		}
		if reply.Message != msg {
			t.Errorf("Call() reply.Message = %q, want %q", reply.Message, msg)
		}
	}
}

// TestRpcClient_Call_TcpError_ReturnsErrRpcConnectivity verifies that a TCP-level
// connection loss (server closed) causes Call() to return ErrRpcConnectivity.
func TestRpcClient_Call_TcpError_ReturnsErrRpcConnectivity(t *testing.T) {

	// Prepare unit test data
	addr, teardown := startRpcTestServer(t, EchoService{})

	logger := scanUtils.NewTestLogger()
	client := NewRpcClient(addr, false, "", "")
	defer client.Disconnect()
	connectRpcClient(t, client, logger)

	// Successful baseline call
	var reply EchoReply
	errBaseline := client.Call(logger, "EchoService.Echo", &EchoArgs{Message: "ping"}, &reply)
	if errBaseline != nil {
		t.Fatalf("baseline Call() error = '%v', want nil", errBaseline)
	}

	// Drop the server to simulate TCP-level connectivity loss
	teardown()
	time.Sleep(10 * time.Millisecond) // Let the OS propagate the close

	// Verify the call returns ErrRpcConnectivity, not some other error
	errAfterClose := client.Call(logger, "EchoService.Echo", &EchoArgs{Message: "ping"}, &reply)
	if !errors.Is(errAfterClose, ErrRpcConnectivity) {
		t.Errorf("Call() after TCP loss = '%v', want ErrRpcConnectivity", errAfterClose)
	}
}

// TestRpcClient_Call_RestoresServerErrors verifies common sentinels survive net/rpc transport.
func TestRpcClient_Call_RestoresServerErrors(t *testing.T) {

	// Prepare common and unknown server errors
	unknownError := errors.New("unknown RPC test error")
	serverErrors := []error{
		ErrRpcGeneric,
		ErrRpcConnectivity,
		ErrRpcCompatibility,
		unknownError,
	}
	configuredErrors := make(map[string]error, len(serverErrors))
	for _, serverError := range serverErrors {
		configuredErrors[serverError.Error()] = serverError
	}

	// Start an RPC server returning common and unknown errors
	address, teardown := startRpcTestServer(t, &RpcErrorTestService{errors: configuredErrors})
	defer teardown()
	logger := scanUtils.NewTestLogger()
	client := NewRpcClient(address, false, "", "")
	defer client.Disconnect()
	connectRpcClient(t, client, logger)

	// Common server errors must regain their exact sentinel identity
	for _, expected := range serverErrors[:len(serverErrors)-1] {
		errCall := client.Call(logger, "RpcErrorTestService.Return", expected.Error(), &struct{}{})
		if errCall != expected {
			t.Errorf("Call() error = %v (%T), want restored sentinel %v (%T)", errCall, errCall, expected, expected)
		}
	}

	// Unknown server errors remain rpc.ServerError values
	errCall := client.Call(logger, "RpcErrorTestService.Return", unknownError.Error(), &struct{}{})
	var rpcServerError rpc.ServerError
	if !errors.As(errCall, &rpcServerError) {
		t.Errorf("Call() unknown error = %v (%T), want rpc.ServerError", errCall, errCall)
	}
}

// TestRestoreRpcServerError_Wrapped verifies common sentinels are restored through an error wrapper.
func TestRestoreRpcServerError_Wrapped(t *testing.T) {
	wrappedUnknown := fmt.Errorf("wrapped unknown RPC error: %w", rpc.ServerError("unknown RPC test error"))
	if got := restoreRpcServerError(wrappedUnknown); got != wrappedUnknown {
		t.Errorf("restoreRpcServerError() = %v (%T), want original error %v (%T)", got, got, wrappedUnknown, wrappedUnknown)
	}
}

// TestRpcClient_EstablishedSignalsCurrentConnection verifies late waiters cannot miss an established connection
func TestRpcClient_EstablishedSignalsCurrentConnection(t *testing.T) {

	// Prepare unit test data
	address, teardown := startRpcTestServer(t, EchoService{})
	defer teardown()
	logger := scanUtils.NewTestLogger()
	client := NewRpcClient(address, false, "", "")
	defer client.Disconnect()
	connectRpcClient(t, client, logger)

	// Register only after connection publication to exercise the former missed notification window
	chEstablished := client.Established()

	// Verify the current connection is signaled once and the waiter then closes
	select {
	case _, ok := <-chEstablished:
		if !ok {
			t.Fatal("Established() channel closed without signaling the current connection")
		}
	case <-time.After(time.Second):
		t.Fatal("Established() did not signal the current connection")
	}
	if _, ok := <-chEstablished; ok {
		t.Fatal("Established() channel remained open after signaling")
	}
}

// TestRpcClient_EstablishedSignalsAllWaiters verifies one connection releases every registered waiter
func TestRpcClient_EstablishedSignalsAllWaiters(t *testing.T) {

	// Prepare unit test data
	address, teardown := startRpcTestServer(t, EchoService{})
	defer teardown()
	logger := scanUtils.NewTestLogger()
	client := NewRpcClient(address, false, "", "")
	defer client.Disconnect()
	waiters := []<-chan struct{}{client.Established(), client.Established(), client.Established()}

	// Establish the connection after all waiters are registered
	connectRpcClient(t, client, logger)

	// Verify each waiter receives one signal and closes
	for waiterIndex, chEstablished := range waiters {
		select {
		case _, ok := <-chEstablished:
			if !ok {
				t.Fatalf("Established() waiter %d closed without a signal", waiterIndex)
			}
		case <-time.After(time.Second):
			t.Fatalf("Established() waiter %d was not signaled", waiterIndex)
		}
		if _, ok := <-chEstablished; ok {
			t.Fatalf("Established() waiter %d remained open after signaling", waiterIndex)
		}
	}
}

// TestRpcClient_EstablishedClosesOnDisconnect verifies shutdown releases a waiter without a false signal
func TestRpcClient_EstablishedClosesOnDisconnect(t *testing.T) {

	// Register a waiter without starting a connection attempt, then shut down the client
	client := NewRpcClient("127.0.0.1:1", false, "", "")
	chEstablished := client.Established()
	client.Disconnect()

	// Verify shutdown closes the waiter without reporting a connection
	select {
	case _, ok := <-chEstablished:
		if ok {
			t.Fatal("Established() signaled a connection during shutdown")
		}
	case <-time.After(time.Second):
		t.Fatal("Established() waiter was not released during shutdown")
	}
	if _, ok := <-chEstablished; ok {
		t.Fatal("Established() waiter reopened after shutdown")
	}
}
