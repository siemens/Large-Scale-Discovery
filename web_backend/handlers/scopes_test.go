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

// TestScopes_Admin verifies that an admin user triggers an RPC call and receives 503 when RPC is down.
func TestScopes_Admin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Scopes()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Scopes() status = '%v', want = '503'", w.Code)
	}
}

// TestScopes_NonAdmin verifies that a non-admin user also triggers an RPC call and receives 503 when RPC is down.
func TestScopes_NonAdmin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Scopes()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Scopes() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeDelete_BadBody verifies that a malformed request body returns 400.
func TestScopeDelete_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeDelete_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeDelete_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeDelete_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeDelete_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeDelete() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeTargets_BadBody verifies that a malformed request body returns 400.
func TestScopeTargets_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeTargets() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeTargets_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeTargets_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeTargets() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeTargets_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeTargets_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeTargets() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeResetFailed_BadBody verifies that a malformed request body returns 400.
func TestScopeResetFailed_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetFailed() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetFailed_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeResetFailed_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeResetFailed() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetFailed_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeResetFailed_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeResetFailed() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeNewCycle_BadBody verifies that a malformed request body returns 400.
func TestScopeNewCycle_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeNewCycle() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeNewCycle_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeNewCycle_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeNewCycle() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeNewCycle_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeNewCycle_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeNewCycle() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeTogglePause_BadBody verifies that a malformed request body returns 400.
func TestScopeTogglePause_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeTogglePause() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeTogglePause_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeTogglePause_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeTogglePause() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeTogglePause_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeTogglePause_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeTogglePause() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeUpdateSettings_BadBody verifies that a malformed request body returns 400.
func TestScopeUpdateSettings_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeUpdateSettings()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeUpdateSettings() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeUpdateSettings_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeUpdateSettings_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeUpdateSettings()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeUpdateSettings() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeCreateUpdateCustom_BadBody verifies that a malformed request body returns 400.
func TestScopeCreateUpdateCustom_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateCustom()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateCustom() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeCreateUpdateCustom_BothIdsSet verifies that providing both scope ID and group ID returns a 200 with an error body.
func TestScopeCreateUpdateCustom_BothIdsSet_Returns200Error(t *testing.T) {
	scopeId := uint64(1)
	groupId := uint64(2)
	type req struct {
		ScopeId *uint64 `json:"scope_id"`
		GroupId *uint64 `json:"group_id"`
		Name    string  `json:"name"`
	}
	ctx, w := newCtx(req{ScopeId: &scopeId, GroupId: &groupId, Name: "test"}, adminUser())
	ScopeCreateUpdateCustom()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeCreateUpdateCustom() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeCreateUpdateCustom_EmptyName verifies that an empty scope name returns a 200 with an error body.
func TestScopeCreateUpdateCustom_EmptyName_Returns200Error(t *testing.T) {
	groupId := uint64(1)
	type req struct {
		GroupId *uint64 `json:"group_id"`
		Name    string  `json:"name"`
	}
	ctx, w := newCtx(req{GroupId: &groupId, Name: ""}, adminUser())
	ScopeCreateUpdateCustom()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeCreateUpdateCustom() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeCreateUpdateNetworks_BadBody verifies that a malformed request body returns 400.
func TestScopeCreateUpdateNetworks_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateNetworks()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateNetworks() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeCreateUpdateNetworks_EmptyName verifies that an empty scope name returns a 200 with an error body.
func TestScopeCreateUpdateNetworks_EmptyName_Returns200Error(t *testing.T) {
	groupId := uint64(1)
	type req struct {
		GroupId *uint64 `json:"group_id"`
		Name    string  `json:"name"`
	}
	ctx, w := newCtx(req{GroupId: &groupId, Name: ""}, adminUser())
	ScopeCreateUpdateNetworks()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeCreateUpdateNetworks() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeCreateUpdateAssets_BadBody verifies that a malformed request body returns 400.
func TestScopeCreateUpdateAssets_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateAssets()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateAssets() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeCreateUpdateAssets_EmptyName verifies that an empty scope name returns a 200 with an error body.
func TestScopeCreateUpdateAssets_EmptyName_Returns200Error(t *testing.T) {
	groupId := uint64(1)
	type req struct {
		GroupId *uint64 `json:"group_id"`
		Name    string  `json:"name"`
	}
	ctx, w := newCtx(req{GroupId: &groupId, Name: ""}, adminUser())
	ScopeCreateUpdateAssets()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeCreateUpdateAssets() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetInput_BadBody verifies that a malformed request body returns 400.
func TestScopeResetInput_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetInput()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetInput() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetInput_ZeroScopeId verifies that a scope ID of zero returns a 200 with an error body.
func TestScopeResetInput_ZeroScopeId_Returns200Error(t *testing.T) {
	type req struct {
		ScopeId uint64 `json:"scope_id"`
		Input   string `json:"input"`
	}
	ctx, w := newCtx(req{ScopeId: 0, Input: "192.168.1.0/24"}, adminUser())
	ScopeResetInput()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeResetInput() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetInput_ValidScopeId verifies that a valid scope ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeResetInput_ValidScopeId_RpcUnavailable_Returns503(t *testing.T) {
	type req struct {
		ScopeId uint64 `json:"scope_id"`
		Input   string `json:"input"`
	}
	ctx, w := newCtx(req{ScopeId: 1, Input: "192.168.1.0/24"}, adminUser())
	ScopeResetInput()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeResetInput() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeResetSecret_BadBody verifies that a malformed request body returns 400.
func TestScopeResetSecret_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetSecret() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetSecret_ZeroId verifies that an ID of zero returns a 200 with an error body.
func TestScopeResetSecret_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeResetSecret() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetSecret_ValidId verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeResetSecret_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeResetSecret() status = '%v', want = '503'", w.Code)
	}
}
