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

// TestAgents_Admin verifies that an admin triggers an RPC call and receives 503 when RPC is down.
func TestAgents_Admin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Agents()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Agents() status = '%v', want = '503'", w.Code)
	}
}

// TestAgents_NonAdmin_NoOwnerships verifies that a non-admin with no ownerships or accessible views receives 401.
// RpcGetViewsGranted is called first; with RPC down it returns an empty list. Both ownerships and accessible
// views are empty, so the handler falls through to the auth-error branch.
func TestAgents_NonAdmin_NoOwnerships_Returns401(t *testing.T) {
	u := regularUser()
	u.Ownerships = nil
	ctx, w := newCtx(nil, u)
	Agents()(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Agents() status = '%v', want = '401'", w.Code)
	}
}

// TestAgentDelete_BadBody verifies that a malformed request body returns 400.
func TestAgentDelete_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	AgentDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("AgentDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestAgentDelete_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestAgentDelete_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	AgentDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("AgentDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestAgentDelete_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestAgentDelete_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	AgentDelete()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("AgentDelete() status = '%v', want = '503'", w.Code)
	}
}
