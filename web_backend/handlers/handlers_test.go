/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_test"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// idReq is a minimal request body carrying a single numeric ID, shared across handler tests.
type idReq struct {
	Id uint64 `json:"id"`
}

// TestMain initialises a shared in-memory SQLite DB and an unconnected RPC client
// (returns ErrRpcConnectivity on every call) before any test runs.
func TestMain(m *testing.M) {

	// GetSettings sets the working directory to _test/ so all test-created files are isolated there.
	_test.GetSettings()

	gin.SetMode(gin.TestMode)

	if errOpenForTesting := database.OpenForTesting(); errOpenForTesting != nil {
		panic("test DB open failed: " + errOpenForTesting.Error())
	}
	if errAutoMigrate := database.AutoMigrate(); errAutoMigrate != nil {
		panic("test DB migrate failed: " + errAutoMigrate.Error())
	}

	// Initialise a minimal web config so handlers that read it (e.g. ViewGrantToken
	// validates ExpiryTokenMax before making an RPC call) don't short-circuit.
	config.Set(&config.Web{
		ExpiryTokenMax: 365 * 24 * time.Hour,
	})

	// Port 1 is reserved and won't be reachable; the RPC client will return
	// ErrRpcConnectivity on every call, exercising the connectivity-error paths.
	core.SetRpcClientForTest(utils.NewRpcClient("127.0.0.1:1", false, "", ""))

	os.Exit(m.Run())
}

// newCtx builds a Gin test context with the given JSON body and authenticated user.
func newCtx(body interface{}, userEntry *database.T_user) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, _ := http.NewRequest(http.MethodPost, "/test", bodyReader)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	core.SetContextStorage(ctx, &core.ContextStorage{
		Logger:      scanUtils.NewTestLogger(),
		CurrentUser: userEntry,
	})
	return ctx, w
}

// badBodyCtx creates a context whose request body is intentionally malformed JSON.
func badBodyCtx(userEntry *database.T_user) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	core.SetContextStorage(ctx, &core.ContextStorage{
		Logger:      scanUtils.NewTestLogger(),
		CurrentUser: userEntry,
	})
	return ctx, w
}

// adminUser returns a minimal admin user for test contexts.
func adminUser() *database.T_user {
	return &database.T_user{
		Id:        999,
		Email:     "admin@test.local",
		Admin:     true,
		Active:    true,
		LastLogin: time.Now(),
	}
}

// regularUser returns a minimal non-admin user for test contexts.
func regularUser() *database.T_user {
	return &database.T_user{
		Id:        998,
		Email:     "user@test.local",
		Admin:     false,
		Active:    true,
		LastLogin: time.Now(),
	}
}

// TestBackendSettings_Returns200 verifies that the settings endpoint returns 200 with no authentication required.
func TestBackendSettings_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, nil)
	BackendSettings()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("BackendSettings() status = '%v', want = '200'", w.Code)
	}
}

// TestBackendAuthenticator_EmptyEmail_Returns200Error verifies that an empty email returns 200 with an error flag.
func TestBackendAuthenticator_EmptyEmail_Returns200Error(t *testing.T) {
	type req struct {
		Email string `json:"email"`
	}
	ctx, w := newCtx(req{Email: ""}, nil)
	BackendAuthenticator()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("BackendAuthenticator() status = '%v', want = '200'", w.Code)
	}
}

// TestBackendAuthenticator_InvalidEmail_Returns401 verifies that a non-email string returns 401.
func TestBackendAuthenticator_InvalidEmail_Returns401(t *testing.T) {
	type req struct {
		Email string `json:"email"`
	}
	ctx, w := newCtx(req{Email: "not-an-email"}, nil)
	BackendAuthenticator()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("BackendAuthenticator() status = '%v', want = '401'", w.Code)
	}
}

// TestBackendAuthenticator_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestBackendAuthenticator_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(nil)
	BackendAuthenticator()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("BackendAuthenticator() status = '%v', want = '400'", w.Code)
	}
}

// TestUserResetDbPassword_DemoUser_Returns401 verifies that demo accounts cannot create database credentials.
func TestUserResetDbPassword_DemoUser_Returns401(t *testing.T) {

	// Prepare unit test data
	user := regularUser()
	user.Demo = true
	ctx, recorder := newCtx(nil, user)

	// Invoke the protected handler
	UserResetDbPassword(nil)(ctx)

	// Verify demo access is rejected before password side effects begin
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("UserResetDbPassword() status = '%v', want = '401'", recorder.Code)
	}
}
