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
	"fmt"
	"net"
	"net/rpc"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

// notificationRpcManager provides deterministic long polling responses for subscription lifecycle tests
type notificationRpcManager struct {
	calls         atomic.Int32
	notification  ReplyNotification
	secondStarted chan struct{}
	secondRelease chan struct{}
}

// SubscribeNotification returns one notification and blocks the following request until the test releases it
func (manager *notificationRpcManager) SubscribeNotification(_ struct{}, reply *ReplyNotification) error {
	if manager.calls.Add(1) == 1 {
		*reply = manager.notification
		return nil
	}

	close(manager.secondStarted)
	<-manager.secondRelease
	return errors.New("test subscription stopped")
}

// startNotificationRpcServer starts an isolated manager RPC service for one client connection
func startNotificationRpcServer(t *testing.T, manager *notificationRpcManager) (string, func()) {
	t.Helper()

	server := rpc.NewServer()
	if errRegister := server.RegisterName("Manager", manager); errRegister != nil {
		t.Fatalf("Could not register notification RPC service: %v", errRegister)
	}

	listener, errListener := net.Listen("tcp", "127.0.0.1:0")
	if errListener != nil {
		t.Fatalf("Could not listen for notification RPC test: %v", errListener)
	}
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		connection, errConnection := listener.Accept()
		if errConnection != nil {
			return
		}
		defer func() { _ = connection.Close() }()
		server.ServeConn(connection)
	}()

	return listener.Addr().String(), func() {
		_ = listener.Close()
		select {
		case <-serverDone:
		case <-time.After(time.Second):
			t.Error("Could not stop notification RPC test server")
		}
	}
}

// requireNotificationOutputsClosed verifies that both subscription outputs close without further values
func requireNotificationOutputsClosed(
	t *testing.T,
	notifications chan ReplyNotification,
	reconnections chan struct{},
) {
	t.Helper()

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()
	for notifications != nil || reconnections != nil {
		select {
		case _, open := <-notifications:
			if open {
				t.Fatal("Notification channel produced an unexpected value during shutdown")
			}
			notifications = nil
		case _, open := <-reconnections:
			if open {
				t.Fatal("Reconnection channel produced an unexpected value during shutdown")
			}
			reconnections = nil
		case <-timeout.C:
			t.Fatal("Subscription output channels did not close during shutdown")
		}
	}
}

// TestSubscribeNotification_NotifierShutdown_ReturnsErrRpcGeneric verifies that SubscribeNotification returns ErrRpcGeneric when the notifier shuts down.
func TestSubscribeNotification_NotifierShutdown_ReturnsErrRpcGeneric(t *testing.T) {

	// Prepare unit test data
	testNotifier := utils.NewNotifier(time.Millisecond, func(a, b interface{}) bool { return a == b })
	original := scopeChangeNotifier
	scopeChangeNotifier = testNotifier
	defer func() { scopeChangeNotifier = original }()

	type callResult struct{ err error }
	ch := make(chan callResult, 1)
	go func() {
		reply := ReplyNotification{}
		ch <- callResult{err: (&Manager{}).SubscribeNotification(struct{}{}, &reply)}
	}()

	time.Sleep(10 * time.Millisecond)
	testNotifier.Shutdown()

	// Verify
	select {
	case result := <-ch:
		if !errors.Is(result.err, utils.ErrRpcGeneric) {
			t.Errorf("SubscribeNotification() err = '%v', want '%v'", result.err, utils.ErrRpcGeneric)
		}
	case <-time.After(time.Second):
		t.Fatal("SubscribeNotification() did not return after notifier shutdown")
	}
}

// TestRpcSubscribeNotification_CancellationAndClientShutdownClosesOutputs verifies the full shutdown sequence
func TestRpcSubscribeNotification_CancellationAndClientShutdownClosesOutputs(t *testing.T) {

	// Start a server that succeeds once and then blocks a second subscription request
	expectedNotification := ReplyNotification{
		UpdateScopeIds:    []uint64{11},
		RemainingScopeIds: []uint64{11, 12},
	}
	testManager := &notificationRpcManager{
		notification:  expectedNotification,
		secondStarted: make(chan struct{}),
		secondRelease: make(chan struct{}),
	}
	address, stopServer := startNotificationRpcServer(t, testManager)
	logger := scanUtils.NewTestLogger()
	client := utils.NewRpcClient(address, false, "", "")
	var releaseSecond sync.Once
	releaseBlockedCall := func() {
		releaseSecond.Do(func() { close(testManager.secondRelease) })
	}
	if !client.Connect(logger, false) {
		client.Disconnect()
		releaseBlockedCall()
		stopServer()
		t.Fatal("Could not connect notification RPC test client")
	}
	defer func() {
		releaseBlockedCall()
		client.Disconnect()
		stopServer()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	notifications, reconnections := RpcSubscribeNotification(logger, client, ctx)
	select {
	case notification := <-notifications:
		if !reflect.DeepEqual(notification, expectedNotification) {
			t.Fatalf("RpcSubscribeNotification() notification = %#v, want %#v", notification, expectedNotification)
		}
	case <-time.After(time.Second):
		t.Fatal("RpcSubscribeNotification() did not deliver the notification")
	}

	// Cancel while the next long polling request is inside the RPC server
	select {
	case <-testManager.secondStarted:
	case <-time.After(time.Second):
		t.Fatal("RpcSubscribeNotification() did not start the next request")
	}
	cancel()
	client.Disconnect()
	requireNotificationOutputsClosed(t, notifications, reconnections)
	releaseBlockedCall()
}

// TestRpcSubscribeNotification_AlreadyCanceledDoesNotStart verifies no goroutine or channels are created after shutdown
func TestRpcSubscribeNotification_AlreadyCanceledDoesNotStart(t *testing.T) {
	logger := scanUtils.NewTestLogger()
	client := utils.NewRpcClient("127.0.0.1:1", false, "", "")
	defer client.Disconnect()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	notifications, reconnections := RpcSubscribeNotification(logger, client, ctx)
	if notifications != nil || reconnections != nil {
		t.Fatal("RpcSubscribeNotification() created output channels for an already canceled context")
	}
}

// TestRpcSubscribeNotification_DisconnectedClientClosesOutputs verifies permanent client shutdown releases subscribers
func TestRpcSubscribeNotification_DisconnectedClientClosesOutputs(t *testing.T) {
	logger := scanUtils.NewTestLogger()
	client := utils.NewRpcClient("127.0.0.1:1", false, "", "")
	client.Disconnect()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifications, reconnections := RpcSubscribeNotification(logger, client, ctx)
	requireNotificationOutputsClosed(t, notifications, reconnections)
}

// TestRestoreRpcServerError verifies manager sentinel errors regain their identity after net/rpc transport.
func TestRestoreRpcServerError(t *testing.T) {
	unknownServerError := rpc.ServerError("unknown manager RPC error")
	localError := errors.New(ErrDatabaseInUse.Error())
	wrappedKnownError := fmt.Errorf("wrapped manager RPC error: %w", rpc.ServerError(ErrDatabaseInUse.Error()))
	wrappedUnknownError := fmt.Errorf("wrapped manager RPC error: %w", unknownServerError)
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", err: nil, want: nil},
		{name: "local error", err: localError, want: localError},
		{name: "database in use", err: rpc.ServerError(ErrDatabaseInUse.Error()), want: ErrDatabaseInUse},
		{name: "wrapped database in use", err: wrappedKnownError, want: ErrDatabaseInUse},
		{name: "database duplicate", err: rpc.ServerError(ErrDatabaseDuplicate.Error()), want: ErrDatabaseDuplicate},
		{name: "scope update ongoing", err: rpc.ServerError(ErrScopeUpdateOngoing.Error()), want: ErrScopeUpdateOngoing},
		{name: "view name existing", err: rpc.ServerError(ErrViewNameExisting.Error()), want: ErrViewNameExisting},
		{name: "invalid privilege", err: rpc.ServerError(ErrInvalidPrivilege.Error()), want: ErrInvalidPrivilege},
		{name: "unknown server error", err: unknownServerError, want: unknownServerError},
		{name: "wrapped unknown server error", err: wrappedUnknownError, want: wrappedUnknownError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := restoreRpcServerError(test.err); got != test.want {
				t.Errorf("restoreRpcServerError() = %v (%T), want %v (%T)", got, got, test.want, test.want)
			}
		})
	}
}
