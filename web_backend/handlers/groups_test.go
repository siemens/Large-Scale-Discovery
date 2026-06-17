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

// TestGroups_Admin verifies that an admin receives all groups (empty DB returns 200).
func TestGroups_Admin_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Groups()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Groups() status = '%v', want = '200'", w.Code)
	}
}

// TestGroups_NonAdmin verifies that a non-admin receives their own groups (empty DB returns 200).
func TestGroups_NonAdmin_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Groups()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Groups() status = '%v', want = '200'", w.Code)
	}
}

// TestGroupCreate_NonAdmin verifies that a non-admin user is rejected with 401.
func TestGroupCreate_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	GroupCreate()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GroupCreate() status = '%v', want = '401'", w.Code)
	}
}

// TestGroupCreate_Admin_BadBody verifies that a malformed request body returns 400.
func TestGroupCreate_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	GroupCreate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GroupCreate() status = '%v', want = '400'", w.Code)
	}
}

// TestGroupCreate_Admin_EmptyName verifies that an empty group name returns a 200 with an error body.
func TestGroupCreate_Admin_EmptyName_Returns200Error(t *testing.T) {
	type req struct {
		Name string `json:"name"`
	}
	ctx, w := newCtx(req{Name: ""}, adminUser())
	GroupCreate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("GroupCreate() status = '%v', want = '200'", w.Code)
	}
}

// TestGroupUpdate_NonAdmin verifies that a non-admin user is rejected with 401.
func TestGroupUpdate_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	GroupUpdate()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GroupUpdate() status = '%v', want = '401'", w.Code)
	}
}

// TestGroupUpdate_Admin_BadBody verifies that a malformed request body returns 400.
func TestGroupUpdate_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	GroupUpdate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GroupUpdate() status = '%v', want = '400'", w.Code)
	}
}

// TestGroupUpdate_Admin_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestGroupUpdate_Admin_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	GroupUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("GroupUpdate() status = '%v', want = '200'", w.Code)
	}
}

// TestGroupDelete_NonAdmin verifies that a non-admin user is rejected with 401.
func TestGroupDelete_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	GroupDelete()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GroupDelete() status = '%v', want = '401'", w.Code)
	}
}

// TestGroupDelete_Admin_BadBody verifies that a malformed request body returns 400.
func TestGroupDelete_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	GroupDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GroupDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestGroupDelete_Admin_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestGroupDelete_Admin_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	GroupDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("GroupDelete() status = '%v', want = '200'", w.Code)
	}
}
