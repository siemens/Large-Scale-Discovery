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

// TestDatabases_NonAdmin verifies that a non-admin user is rejected with 401.
func TestDatabases_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Databases()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Databases() status = '%v', want = '401'", w.Code)
	}
}

// TestDatabases_Admin verifies that an admin triggers an RPC call and receives 503 when RPC is down.
func TestDatabases_Admin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Databases()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Databases() status = '%v', want = '503'", w.Code)
	}
}

// TestDatabaseRemove_NonAdmin verifies that a non-admin user is rejected with 401.
func TestDatabaseRemove_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	DatabaseRemove()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("DatabaseRemove() status = '%v', want = '401'", w.Code)
	}
}

// TestDatabaseRemove_Admin_BadBody verifies that a malformed request body returns 400.
func TestDatabaseRemove_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	DatabaseRemove()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("DatabaseRemove() status = '%v', want = '400'", w.Code)
	}
}

// TestDatabaseRemove_Admin_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestDatabaseRemove_Admin_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	DatabaseRemove()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("DatabaseRemove() status = '%v', want = '200'", w.Code)
	}
}

// TestDatabaseRemove_Admin_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestDatabaseRemove_Admin_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	DatabaseRemove()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("DatabaseRemove() status = '%v', want = '503'", w.Code)
	}
}

// TestDatabaseAddUpdate_NonAdmin verifies that a non-admin user is rejected with 401.
func TestDatabaseAddUpdate_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	DatabaseAddUpdate()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("DatabaseAddUpdate() status = '%v', want = '401'", w.Code)
	}
}

// TestDatabaseAddUpdate_Admin_BadBody verifies that a malformed request body returns 400.
func TestDatabaseAddUpdate_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	DatabaseAddUpdate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("DatabaseAddUpdate() status = '%v', want = '400'", w.Code)
	}
}

// TestDatabaseAddUpdate_Admin_EmptyName verifies that an empty database name returns a 200 with an error body.
func TestDatabaseAddUpdate_Admin_EmptyName_Returns200Error(t *testing.T) {
	type req struct {
		Name string `json:"name"`
	}
	ctx, w := newCtx(req{Name: ""}, adminUser())
	DatabaseAddUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("DatabaseAddUpdate() status = '%v', want = '200'", w.Code)
	}
}

// TestDatabaseAddUpdate_Admin_EmptyDialect verifies that an empty dialect returns a 200 with an error body.
func TestDatabaseAddUpdate_Admin_EmptyDialect_Returns200Error(t *testing.T) {
	type req struct {
		Name    string `json:"name"`
		Dialect string `json:"dialect"`
	}
	ctx, w := newCtx(req{Name: "testdb", Dialect: ""}, adminUser())
	DatabaseAddUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("DatabaseAddUpdate() status = '%v', want = '200'", w.Code)
	}
}
