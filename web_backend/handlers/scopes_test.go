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
	"net/http"
	"testing"

	"github.com/siemens/Large-Scale-Discovery/utils"
)

// TestScopes_Admin_RpcUnavailable_Returns503 verifies that an admin user triggers an RPC call and receives 503 when RPC is down.
func TestScopes_Admin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, adminUser())
	Scopes()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Scopes() status = '%v', want = '503'", w.Code)
	}
}

// TestScopes_NonAdmin_RpcUnavailable_Returns503 verifies that a non-admin user also triggers an RPC call and receives 503 when RPC is down.
func TestScopes_NonAdmin_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(nil, regularUser())
	Scopes()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Scopes() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeDelete_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeDelete_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeDelete() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeDelete_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeDelete_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeDelete() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeDelete_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeDelete_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeDelete()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeDelete() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeTargets_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeTargets_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeTargets() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeTargets_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeTargets_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeTargets() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeTargets_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeTargets_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeTargets()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeTargets() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeResetFailed_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeResetFailed_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetFailed() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetFailed_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeResetFailed_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeResetFailed() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetFailed_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeResetFailed_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeResetFailed()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeResetFailed() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeNewCycle_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeNewCycle_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeNewCycle() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeNewCycle_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeNewCycle_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeNewCycle() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeNewCycle_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeNewCycle_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeNewCycle()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeNewCycle() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeTogglePause_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeTogglePause_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeTogglePause() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeTogglePause_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeTogglePause_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeTogglePause() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeTogglePause_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeTogglePause_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeTogglePause()(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeTogglePause() status = '%v', want = '503'", w.Code)
	}
}

// TestScopeUpdateSettings_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeUpdateSettings_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeUpdateSettings()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeUpdateSettings() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeUpdateSettings_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeUpdateSettings_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeUpdateSettings()(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeUpdateSettings() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeCreateUpdateCustom_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeCreateUpdateCustom_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateCustom()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateCustom() status = '%v', want = '400'", w.Code)
	}
}

// TestNormalizeNetworkSize_ValidOrOmittedValue_ReturnsSafeSize verifies legacy defaults and valid choices are preserved.
func TestNormalizeNetworkSize_ValidOrOmittedValue_ReturnsSafeSize(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name      string
		splitSize uint32
		want      uint32
	}{
		{
			name:      "omitted-value",
			splitSize: 0,
			want:      utils.NetworkSizeSplit,
		},
		{
			name:      "smaller-valid-value",
			splitSize: 512,
			want:      512,
		},
		{
			name:      "maximum-value",
			splitSize: utils.NetworkSizeSplit,
			want:      utils.NetworkSizeSplit,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			splitSize, errSplitSize := normalizeNetworkSize(test.splitSize)
			if errSplitSize != nil {
				t.Errorf("normalizeNetworkSize() error = '%v', want = 'nil'", errSplitSize)
			}
			if splitSize != test.want {
				t.Errorf("normalizeNetworkSize() result = '%v', want = '%v'", splitSize, test.want)
			}
		})
	}
}

// TestNormalizeNetworkSize_InvalidValue_ReturnsError verifies unsafe split choices cannot enter scope storage.
func TestNormalizeNetworkSize_InvalidValue_ReturnsError(t *testing.T) {

	// Prepare and run test cases
	for _, splitSize := range []uint32{3, 1_000, utils.NetworkSizeSplit + 1} {
		_, errSplitSize := normalizeNetworkSize(splitSize)
		if errSplitSize == nil {
			t.Errorf("normalizeNetworkSize() error = 'nil', want = 'non-nil'")
		}
	}
}

// TestScopeCreateUpdateCustom_BothIdsSet_Returns200Error verifies that providing both scope ID and group ID returns a 200 with an error body.
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

// TestScopeCreateUpdateCustom_EmptyName_Returns200Error verifies that an empty scope name returns a 200 with an error body.
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

// TestScopeCreateUpdateNetworks_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeCreateUpdateNetworks_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateNetworks()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateNetworks() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeCreateUpdateNetworks_EmptyName_Returns200Error verifies that an empty scope name returns a 200 with an error body.
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

// TestScopeCreateUpdateAssets_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeCreateUpdateAssets_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeCreateUpdateAssets()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeCreateUpdateAssets() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeCreateUpdateAssets_EmptyName_Returns200Error verifies that an empty scope name returns a 200 with an error body.
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

// TestScopeResetInput_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeResetInput_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetInput()(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetInput() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetInput_ZeroScopeId_Returns200Error verifies that a scope ID of zero returns a 200 with an error body.
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

// TestScopeResetInput_ValidScopeId_RpcUnavailable_Returns503 verifies that a valid scope ID triggers an RPC call and returns 503 when RPC is down.
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

// TestScopeResetSecret_BadBody_Returns400 verifies that a malformed request body returns 400.
func TestScopeResetSecret_BadBody_Returns400(t *testing.T) {
	ctx, w := badBodyCtx(adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusBadRequest {
		t.Errorf("ScopeResetSecret() status = '%v', want = '400'", w.Code)
	}
}

// TestScopeResetSecret_ZeroId_Returns200Error verifies that an ID of zero returns a 200 with an error body.
func TestScopeResetSecret_ZeroId_Returns200Error(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 0}, adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("ScopeResetSecret() status = '%v', want = '200'", w.Code)
	}
}

// TestScopeResetSecret_ValidId_RpcUnavailable_Returns503 verifies that a valid ID triggers an RPC call and returns 503 when RPC is down.
func TestScopeResetSecret_ValidId_RpcUnavailable_Returns503(t *testing.T) {
	ctx, w := newCtx(idReq{Id: 1}, adminUser())
	ScopeResetSecret(nil)(ctx)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ScopeResetSecret() status = '%v', want = '503'", w.Code)
	}
}
