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
	"os"
	"testing"
	"time"

	"github.com/siemens/Large-Scale-Discovery/_test"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"go.uber.org/zap/zapcore"
)

// TestMain initializes a minimal silent logger so that functions calling log.GetLogger() do not panic.
func TestMain(m *testing.M) {

	// GetSettings sets cwd to _test/ so all test-created files are isolated there
	_ = _test.GetSettings()

	// Initialize a minimal silent logger so functions calling log.GetLogger() do not panic
	_, _ = log.InitGlobalLogger(log.Settings{
		Console: &log.ConsoleHandler{
			Enabled: false,
			Level:   zapcore.InfoLevel,
		},
	})

	// Run all tests
	os.Exit(m.Run())
}

// TestGetScanScope_SecondaryShutdownReturnsErrRpcGeneric verifies that when the broker context
// is cancelled while a secondary goroutine is waiting on the scope lock, the secondary returns
// ErrRpcGeneric and not errInvalidScopeSecret. This is the regression for the shutdown scenario:
// primary aborts due to coreCtx.Done(), secondary must not mis-classify the empty cache as an
// invalid secret.
func TestGetScanScope_SecondaryShutdownReturnsErrRpcGeneric(t *testing.T) {
	// Save and restore package-level context so cancellation does not leak into other tests.
	origCtx, origCancel := coreCtx, coreCtxCancelFunc
	coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background())
	defer func() {
		coreCtxCancelFunc()
		coreCtx, coreCtxCancelFunc = origCtx, origCancel
	}()

	const secret = "shutdown-secondary-thea483"
	logger := log.GetLogger()

	// Hold the named mutex directly, simulating a primary goroutine that is blocking on the manager.
	rpcScopeLock.Lock(secret)

	// Start a secondary caller — will wait on the named mutex and enter the secondary path.
	ch := make(chan error, 1)
	go func() {
		_, errGetScanScope := getScanScope(logger, secret)
		ch <- errGetScanScope
	}()

	// Give the goroutine time to reach rpcScopeLock.Lock inside the secondary path.
	time.Sleep(10 * time.Millisecond)

	// Simulate broker shutdown: cancel the context, then release the scope lock.
	coreCtxCancelFunc()
	rpcScopeLock.Unlock(secret)

	got := <-ch
	if got != utils.ErrRpcGeneric {
		t.Errorf("getScanScope() err = %v, want ErrRpcGeneric", got)
	}
}

// TestGetScanScope_SecondaryAfterRealInvalidSecretReturnsInvalidSecret verifies that when the
// primary goroutine completes normally with a cache-miss (secret truly unknown), the secondary
// still returns errInvalidScopeSecret. This confirms the shutdown fix did not over-correct the
// non-shutdown path.
func TestGetScanScope_SecondaryAfterRealInvalidSecretReturnsInvalidSecret(t *testing.T) {
	// Use a fresh context that is not cancelled.
	origCtx, origCancel := coreCtx, coreCtxCancelFunc
	coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background())
	defer func() {
		coreCtxCancelFunc()
		coreCtx, coreCtxCancelFunc = origCtx, origCancel
	}()

	const secret = "real-invalid-secret-thea483"
	logger := log.GetLogger()

	// Hold the named mutex, simulating a primary that is processing the scope lookup.
	rpcScopeLock.Lock(secret)

	// Start a secondary caller — will wait on the named mutex and enter the secondary path.
	ch := make(chan error, 1)
	go func() {
		_, errGetScanScope := getScanScope(logger, secret)
		ch <- errGetScanScope
	}()

	// Give the goroutine time to reach rpcScopeLock.Lock inside the secondary path.
	time.Sleep(10 * time.Millisecond)

	// Release the lock without cancelling the context, simulating a primary that determined
	// the secret is invalid and left the cache empty.
	rpcScopeLock.Unlock(secret)

	got := <-ch
	if got != errInvalidScopeSecret {
		t.Errorf("getScanScope() err = %v, want errInvalidScopeSecret", got)
	}
}
