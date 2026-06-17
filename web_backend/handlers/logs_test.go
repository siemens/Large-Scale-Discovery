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
)

// TestSqlLogs_NonAdmin verifies that a non-admin user is rejected with 401.
func TestSqlLogs_NonAdmin_Returns401(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	SqlLogs()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("SqlLogs() status = '%v', want = '401'", w.Code)
	}
}

// TestSqlLogs_Admin_BadBody verifies that a malformed request body returns 400.
func TestSqlLogs_Admin_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	SqlLogs()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("SqlLogs() status = '%v', want = '400'", w.Code)
	}
}

// TestSqlLogs_Admin_RpcUnavailable verifies that a valid request returns 503 when RPC is down.
func TestSqlLogs_Admin_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		DbName string    `json:"db_name"`
		Since  time.Time `json:"since"`
	}
	ctx, w := newCtx(req{DbName: "test_db", Since: time.Now().Add(-time.Hour)}, adminUser())
	SqlLogs()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("SqlLogs() status = '%v', want = '503'", w.Code)
	}
}
