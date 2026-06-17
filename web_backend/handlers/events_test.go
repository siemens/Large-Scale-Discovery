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
	"time"

	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// TestEvents_NonAdmin verifies that a non-admin user is rejected with 401.
func TestEvents_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Events()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Events() status = '%v', want = '401'", w.Code)
	}
}

// TestEvents_Admin_BadBody verifies that a malformed request body returns 400.
func TestEvents_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	Events()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Events() status = '%v', want = '400'", w.Code)
	}
}

// TestEvents_Admin_NoSince verifies that querying events without a time filter returns 200 (empty DB).
func TestEvents_Admin_NoSince_Returns200(t *testing.T) {
	type req struct {
		Event database.Event `json:"event"`
	}
	ctx, w := newCtx(req{Event: database.EventScopeCreate}, adminUser())
	Events()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Events() status = '%v', want = '200'", w.Code)
	}
}

// TestEvents_Admin_WithSince verifies that querying events with a time filter returns 200 (empty DB).
func TestEvents_Admin_WithSince_Returns200(t *testing.T) {
	since := time.Now().Add(-24 * time.Hour)
	type req struct {
		Event database.Event `json:"event"`
		Since *time.Time     `json:"since"`
	}
	ctx, w := newCtx(req{Event: database.EventScopeCreate, Since: &since}, adminUser())
	Events()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Events() status = '%v', want = '200'", w.Code)
	}
}
