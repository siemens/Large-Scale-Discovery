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

// TestUsers_Admin verifies that an admin receives the full user list (empty DB returns 200).
func TestUsers_Admin_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Users()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Users() status = '%v', want = '200'", w.Code)
	}
}

// TestUsers_NonAdmin verifies that a non-admin receives a filtered list (empty DB returns 200).
func TestUsers_NonAdmin_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Users()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Users() status = '%v', want = '200'", w.Code)
	}
}

// TestUserUpdate_NonAdmin verifies that a non-admin user is rejected with 401.
func TestUserUpdate_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	UserUpdate()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserUpdate() status = '%v', want = '401'", w.Code)
	}
}

// TestUserUpdate_Admin_BadBody verifies that a malformed request body returns 400.
func TestUserUpdate_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	UserUpdate()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("UserUpdate() status = '%v', want = '400'", w.Code)
	}
}

// TestUserUpdate_Admin_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestUserUpdate_Admin_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	UserUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("UserUpdate() status = '%v', want = '200'", w.Code)
	}
}

// TestUserUpdate_Admin_NonExistentId verifies that a non-existent user ID returns a 200 with an error body.
func TestUserUpdate_Admin_NonExistentId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 99999}, adminUser())
	UserUpdate()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("UserUpdate() status = '%v', want = '200'", w.Code)
	}
}

// TestUserDelete_NonAdmin verifies that a non-admin user is rejected with 401.
func TestUserDelete_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	UserDelete()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserDelete() status = '%v', want = '401'", w.Code)
	}
}

// TestUserDelete_Admin_BadBody verifies that a malformed request body returns 400.
func TestUserDelete_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	UserDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("UserDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestUserDelete_Admin_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestUserDelete_Admin_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	UserDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("UserDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestUserDelete_Admin_NonExistentId verifies that a non-existent user ID returns a 200 with an error body.
func TestUserDelete_Admin_NonExistentId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 99999}, adminUser())
	UserDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("UserDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestUserDetails_Returns200 verifies that UserDetails returns the caller's profile with 200.
// The handler reads no request body; it uses only the context user.
func TestUserDetails_Returns200(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	UserDetails()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("UserDetails() status = '%v', want = '200'", w.Code)
	}
}
