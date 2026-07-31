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
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// TestEvents_Admin_LimitReturnsNewestEntries verifies that the optional limit retains only the newest events
func TestEvents_Admin_LimitReturnsNewestEntries(t *testing.T) {

	// Prepare an isolated event owner and two ordered events
	eventUser := database.NewUser("events-limit@domain.tld", "example-org", "example", "", "Event", "Owner")
	eventUser.Active = true
	if errUserCreate := eventUser.Create(); errUserCreate != nil {
		t.Fatalf("T_user.Create() error = '%v', want = 'nil'", errUserCreate)
	}
	t.Cleanup(func() {
		_ = eventUser.Delete()
	})
	if errEventFirst := database.LogEvent(eventUser, database.EventApiToken, "first"); errEventFirst != nil {
		t.Fatalf("LogEvent() first error = '%v', want = 'nil'", errEventFirst)
	}
	if errEventSecond := database.LogEvent(eventUser, database.EventApiToken, "second"); errEventSecond != nil {
		t.Fatalf("LogEvent() second error = '%v', want = 'nil'", errEventSecond)
	}

	// Request only the newest matching event
	limit := 1
	type req struct {
		Event database.Event `json:"event"`
		Limit *int           `json:"limit"`
	}
	ctx, recorder := newCtx(req{Event: database.EventApiToken, Limit: &limit}, adminUser())
	Events()(ctx)

	// Decode the response envelope for the returned event list
	var response struct {
		Body struct {
			Result []database.T_event `json:"result"`
		} `json:"body"`
	}
	if errResponse := json.Unmarshal(recorder.Body.Bytes(), &response); errResponse != nil {
		t.Fatalf("json.Unmarshal() error = '%v', want = 'nil'", errResponse)
	}

	// Verify the limit retained only the latest event
	if len(response.Body.Result) != 1 {
		t.Fatalf("Events() result count = '%v', want = '1'", len(response.Body.Result))
	}
	if response.Body.Result[0].EventDetail != "second" {
		t.Errorf("Events() EventDetail = '%v', want = 'second'", response.Body.Result[0].EventDetail)
	}
}

// TestEvents_NonAdmin_Returns401 verifies that a non-admin user is rejected with 401.
func TestEvents_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Events()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Events() status = '%v', want = '401'", w.Code)
	}
}

// TestEvents_Admin_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestEvents_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	Events()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Events() status = '%v', want = '400'", w.Code)
	}
}

// TestEvents_Admin_NoSince_Returns200 verifies that querying events without a time filter returns 200 (empty DB).
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

// TestEvents_Admin_WithSince_Returns200 verifies that querying events with a time filter returns 200 (empty DB).
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
