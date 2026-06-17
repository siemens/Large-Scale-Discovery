/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package handlers

import (
	"net/http"
	"testing"
)

// TestViewGrantToken_BadBody verifies that a malformed request body returns 400.
func TestViewGrantToken_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewGrantToken(nil)(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewGrantToken() status = '%v', want = '400'", w.Code)
	}
}

// TestViewGrantToken_ZeroExpiry verifies that an expiry of zero returns a 200 with an error body.
func TestViewGrantToken_ZeroExpiry_Returns200Error(t *testing.T) {
	type req struct {
		ViewId      uint64 `json:"view_id"`
		Description string `json:"description"`
		ExpiryDays  uint   `json:"expiry_days"`
	}
	ctx, w := newCtx(req{ViewId: 1, Description: "test", ExpiryDays: 0}, adminUser())
	ViewGrantToken(nil)(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ViewGrantToken() status = '%v', want = '200'", w.Code)
	}
}

// TestViewGrantToken_ValidInput verifies that valid input triggers an RPC call and returns 503 when RPC is down.
func TestViewGrantToken_ValidInput_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		ViewId      uint64 `json:"view_id"`
		Description string `json:"description"`
		ExpiryDays  uint   `json:"expiry_days"`
	}
	ctx, w := newCtx(req{ViewId: 1, Description: "test", ExpiryDays: 30}, adminUser())
	ViewGrantToken(nil)(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewGrantToken() status = '%v', want = '503'", w.Code)
	}
}

// TestViewGrantUsers_BadBody verifies that a malformed request body returns 400.
func TestViewGrantUsers_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewGrantUsers(nil)(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewGrantUsers() status = '%v', want = '400'", w.Code)
	}
}

// TestViewGrantUsers_RpcUnavailable verifies that a valid request returns 503 when RPC is down.
// ViewGrantUsers calls RpcGetView before ID validation, so an unavailable RPC surfaces immediately.
func TestViewGrantUsers_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		ViewId uint64   `json:"view_id"`
		Users  []string `json:"users"`
	}
	ctx, w := newCtx(req{ViewId: 1, Users: []string{"user@example.com"}}, adminUser())
	ViewGrantUsers(nil)(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewGrantUsers() status = '%v', want = '503'", w.Code)
	}
}

// TestViewGrantRevoke_BadBody verifies that a malformed request body returns 400.
func TestViewGrantRevoke_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewGrantRevoke()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewGrantRevoke() status = '%v', want = '400'", w.Code)
	}
}

// TestViewGrantRevoke_RpcUnavailable verifies that a valid request returns 503 when RPC is down.
// ViewGrantRevoke calls RpcGetView before ID validation, so an unavailable RPC surfaces immediately.
func TestViewGrantRevoke_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		ViewId   uint64 `json:"view_id"`
		Username string `json:"username"`
	}
	ctx, w := newCtx(req{ViewId: 1, Username: "user@example.com"}, adminUser())
	ViewGrantRevoke()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewGrantRevoke() status = '%v', want = '503'", w.Code)
	}
}
