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
	"testing"
	"time"
)

// TestBeginRpcRegistersOnlyBeforeShutdown verifies admission and wait-group registration happen as one operation.
func TestBeginRpcRegistersOnlyBeforeShutdown(t *testing.T) {
	previousCoreCtx := coreCtx
	ctx, cancel := context.WithCancel(context.Background())
	coreCtx = ctx
	t.Cleanup(func() {
		cancel()
		coreCtx = previousCoreCtx
	})

	if !beginRpc() {
		t.Fatal("beginRpc() rejected a request before shutdown")
	}
	rpcInflight.Done()

	// Hold admission while a request arrives, then cancel as shutdown does before releasing the lock
	rpcAdmissionLock.Lock()
	admissionLocked := true
	t.Cleanup(func() {
		if admissionLocked {
			rpcAdmissionLock.Unlock()
		}
	})
	result := make(chan bool, 1)
	go func() { result <- beginRpc() }()
	select {
	case <-result:
		t.Fatal("beginRpc() did not wait for the admission lock")
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	rpcAdmissionLock.Unlock()
	admissionLocked = false
	if <-result {
		rpcInflight.Done()
		t.Fatal("beginRpc() admitted a request after shutdown")
	}
}
