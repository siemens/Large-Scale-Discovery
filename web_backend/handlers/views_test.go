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

// TestViews_Admin verifies that an admin triggers an RPC call and receives 503 when RPC is down.
func TestViews_Admin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Views()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Views() status = '%v', want = '503'", w.Code)
	}
}

// TestViews_NonAdmin verifies that a non-admin triggers an RPC call and receives 503 when RPC is down.
func TestViews_NonAdmin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Views()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Views() status = '%v', want = '503'", w.Code)
	}
}

// TestViewsGranted_RpcUnavailable verifies that ViewsGranted returns 503 when RPC is down.
func TestViewsGranted_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	ViewsGranted()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewsGranted() status = '%v', want = '503'", w.Code)
	}
}

// TestViewCreate_BadBody verifies that a malformed request body returns 400.
func TestViewCreate_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewCreate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewCreate() status = '%v', want = '400'", w.Code)
	}
}

// TestViewCreate_RpcUnavailable verifies that a valid request returns 503 when RPC is down.
// ViewCreate calls RpcGetScope before validating inputs, so an unavailable RPC surfaces immediately.
func TestViewCreate_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		ScopeId  uint64 `json:"scope_id"`
		ViewName string `json:"view_name"`
	}
	ctx, w := newCtx(req{ScopeId: 1, ViewName: "myview"}, adminUser())
	ViewCreate()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewCreate() status = '%v', want = '503'", w.Code)
	}
}

// TestViewDelete_BadBody verifies that a malformed request body returns 400.
func TestViewDelete_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestViewDelete_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestViewDelete_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ViewDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ViewDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestViewDelete_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestViewDelete_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ViewDelete()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ViewDelete() status = '%v', want = '503'", w.Code)
	}
}

// TestViewUpdate_BadBody verifies that a malformed request body returns 400.
func TestViewUpdate_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ViewUpdate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ViewUpdate() status = '%v', want = '400'", w.Code)
	}
}

// TestViewUpdate_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestViewUpdate_ZeroId_Returns200Error(t *testing.T) {
	type req struct {
		Id   uint64 `json:"id"`
		Name string `json:"name"`
	}
	ctx, w := newCtx(req{Id: 0, Name: "newname"}, adminUser())
	ViewUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ViewUpdate() status = '%v', want = '200'", w.Code)
	}
}
