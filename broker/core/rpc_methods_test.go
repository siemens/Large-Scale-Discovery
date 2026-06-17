/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"testing"

	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/broker/memory"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

// TestAgentCompatible verifies agentCompatible accepts versions >= the broker API version and rejects older ones.
func TestAgentCompatible(t *testing.T) {

	// Prepare and run test cases
	tests := []struct {
		name       string
		version    scanUtils.Version
		wantCompat bool
	}{
		{
			name:       "same-version",
			version:    BrokerApiVersion,
			wantCompat: true,
		},
		{
			name:       "higher-major",
			version:    scanUtils.Version{Major: BrokerApiVersion.Major + 1, Minor: 0, Patch: 0},
			wantCompat: true,
		},
		{
			name:       "higher-minor",
			version:    scanUtils.Version{Major: BrokerApiVersion.Major, Minor: BrokerApiVersion.Minor + 1, Patch: 0},
			wantCompat: true,
		},
		{
			name:       "higher-patch",
			version:    scanUtils.Version{Major: BrokerApiVersion.Major, Minor: BrokerApiVersion.Minor, Patch: BrokerApiVersion.Patch + 1},
			wantCompat: true,
		},
		{
			name:       "lower-minor",
			version:    scanUtils.Version{Major: BrokerApiVersion.Major, Minor: BrokerApiVersion.Minor - 1, Patch: 0},
			wantCompat: false,
		},
		{
			name:       "lower-major",
			version:    scanUtils.Version{Major: 0, Minor: 9, Patch: 9},
			wantCompat: false,
		},
		{
			name:       "zero-version",
			version:    scanUtils.Version{},
			wantCompat: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agentCompatible(tt.version)
			if got != tt.wantCompat {
				t.Errorf("agentCompatible() = '%v', want = '%v'", got, tt.wantCompat)
			}
		})
	}
}

// TestLockUnlock verifies lock/unlock complete without deadlock or panic and release the slot correctly.
func TestLockUnlock(t *testing.T) {

	// Prepare unit test data
	scopeId := uint64(99)
	module := "test-module"

	// Verify lock and unlock complete without blocking
	lock(scopeId, module)
	unlock(scopeId, module)

	// A second round confirms the lock was actually released
	lock(scopeId, module)
	unlock(scopeId, module)
}

// TestRegisterGobs verifies RegisterGobs is idempotent and does not panic.
func TestRegisterGobs(t *testing.T) {

	// Verify repeated calls do not panic
	RegisterGobs()
	RegisterGobs()
}

// TestBrokerApiVersion verifies BrokerApiVersion is not the zero value.
func TestBrokerApiVersion(t *testing.T) {

	// Verify version is non-zero
	if BrokerApiVersion.Major == 0 && BrokerApiVersion.Minor == 0 && BrokerApiVersion.Patch == 0 {
		t.Errorf("BrokerApiVersion = '%v', want non-zero version", BrokerApiVersion)
	}
}

// TestRequestScanTasks_IncompatibleVersion verifies incompatible agents receive ErrRpcCompatibility and no tasks.
func TestRequestScanTasks_IncompatibleVersion(t *testing.T) {

	// Prepare unit test data
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: scanUtils.Version{Major: 0, Minor: 0, Patch: 1},
			Name:       "old-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: "any-secret",
	}
	reply := &ReplyGetScanTask{}

	// Run test case
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != utils.ErrRpcCompatibility {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, utils.ErrRpcCompatibility)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}

// TestRequestScanTasks_PausedScope verifies no tasks are returned for a paused scan scope.
func TestRequestScanTasks_PausedScope(t *testing.T) {

	// Prepare unit test data
	const secret = "paused-scope-secret-thea14"
	memory.AddScope(secret, managerdb.T_scan_scope{
		Id:      101,
		Name:    "paused-scope",
		Enabled: false,
		Cycles:  false,
	})

	// Prepare cleanup
	defer memory.RemoveScope(secret)

	// Run test case
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "test-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: secret,
		ModuleData:  []ModuleData{},
	}
	reply := &ReplyGetScanTask{}
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != nil {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, nil)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}

// TestRequestScanTasks_ScopeInstancesFull verifies no tasks are dispatched when all scope-level slots are occupied.
func TestRequestScanTasks_ScopeInstancesFull(t *testing.T) {

	// Prepare unit test data
	const secret = "scope-full-secret-thea14"
	const maxInstances = uint32(3)
	memory.AddScope(secret, managerdb.T_scan_scope{
		Id:      102,
		Name:    "scope-full",
		Enabled: true,
		Cycles:  false,
		ScanSettings: managerdb.T_scan_setting{
			MaxInstancesDiscovery: maxInstances,
		},
	})

	// Prepare cleanup
	defer memory.RemoveScope(secret)

	// Run test case
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "test-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: secret,
		ModuleData: []ModuleData{
			{
				Label:          discovery.Label,
				MaxInstances:   -1, // use scope-level limit
				TotalInstances: 0,
				ScopeInstances: int(maxInstances), // all slots occupied
			},
		},
	}
	reply := &ReplyGetScanTask{}
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != nil {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, nil)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}

// TestRequestScanTasks_AgentSpecificLimitNoSlots verifies an agent-specific limit of zero produces no tasks.
func TestRequestScanTasks_AgentSpecificLimitNoSlots(t *testing.T) {

	// Prepare unit test data
	const secret = "agent-limit-secret-thea14"
	memory.AddScope(secret, managerdb.T_scan_scope{
		Id:      103,
		Name:    "agent-limit-scope",
		Enabled: true,
		Cycles:  false,
		ScanSettings: managerdb.T_scan_setting{
			MaxInstancesBanner: 10,
		},
	})

	// Prepare cleanup
	defer memory.RemoveScope(secret)

	// Run test case
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "limited-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: secret,
		ModuleData: []ModuleData{
			{
				Label:          banner.Label,
				MaxInstances:   0, // agent-specific override: 0 max
				TotalInstances: 0,
				ScopeInstances: 0,
			},
		},
	}
	reply := &ReplyGetScanTask{}
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != nil {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, nil)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}

// TestRequestScanTasks_MultipleModulesAllFull verifies no tasks when all slots are full across multiple modules.
func TestRequestScanTasks_MultipleModulesAllFull(t *testing.T) {

	// Prepare unit test data
	const secret = "multi-module-full-secret-thea14"
	memory.AddScope(secret, managerdb.T_scan_scope{
		Id:      104,
		Name:    "multi-full-scope",
		Enabled: true,
		Cycles:  false,
		ScanSettings: managerdb.T_scan_setting{
			MaxInstancesDiscovery: 2,
			MaxInstancesBanner:    2,
			MaxInstancesNfs:       2,
			MaxInstancesSmb:       2,
			MaxInstancesSsl:       2,
		},
	})

	// Prepare cleanup
	defer memory.RemoveScope(secret)

	// Run test case
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "busy-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: secret,
		ModuleData: []ModuleData{
			{Label: discovery.Label, MaxInstances: -1, ScopeInstances: 2},
			{Label: banner.Label, MaxInstances: -1, ScopeInstances: 2},
			{Label: nfs.Label, MaxInstances: -1, ScopeInstances: 2},
			{Label: smb.Label, MaxInstances: -1, ScopeInstances: 2},
			{Label: ssl.Label, MaxInstances: -1, ScopeInstances: 2},
		},
	}
	reply := &ReplyGetScanTask{}
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != nil {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, nil)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}

// TestSubmitScanResult_IncompatibleVersion verifies outdated agents receive ErrRpcCompatibility.
func TestSubmitScanResult_IncompatibleVersion(t *testing.T) {

	// Prepare unit test data
	b := &Broker{}
	args := &ArgsSaveScanResult{
		AgentInfo: AgentInfo{
			ApiVersion: scanUtils.Version{Major: 0, Minor: 0, Patch: 1},
			Name:       "old-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: "any-secret",
		Id:          1,
		Result:      nil,
	}

	// Run test case
	errSubmit := b.SubmitScanResult(args, &struct{}{})

	// Verify results
	if errSubmit != utils.ErrRpcCompatibility {
		t.Errorf("SubmitScanResult() err = '%v', want = '%v'", errSubmit, utils.ErrRpcCompatibility)
	}
}

// TestSubmitScanResult_CompatibleVersionDiscoveryResult verifies a compatible agent submitting a discovery result receives nil.
func TestSubmitScanResult_CompatibleVersionDiscoveryResult(t *testing.T) {

	// Prepare unit test data
	b := &Broker{}
	args := &ArgsSaveScanResult{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "test-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: "discovery-result-secret-thea14",
		Id:          1,
		Result:      discovery.Result{},
	}

	// Run test case
	errSubmit := b.SubmitScanResult(args, &struct{}{})

	// Verify results
	if errSubmit != nil {
		t.Errorf("SubmitScanResult() err = '%v', want = '%v'", errSubmit, nil)
	}
}

// TestSubmitScanResult_CompatibleVersionSubResult verifies a compatible agent submitting a non-discovery result receives nil.
func TestSubmitScanResult_CompatibleVersionSubResult(t *testing.T) {

	// Prepare unit test data
	b := &Broker{}
	args := &ArgsSaveScanResult{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "test-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: "sub-result-secret-thea14",
		Id:          2,
		Result:      nil, // nil matches the default branch
	}

	// Run test case
	errSubmit := b.SubmitScanResult(args, &struct{}{})

	// Verify results
	if errSubmit != nil {
		t.Errorf("SubmitScanResult() err = '%v', want = '%v'", errSubmit, nil)
	}
}

// TestRequestScanTasks_UnknownModuleLabelReturnsError verifies an unrecognised module label causes RequestScanTasks to return ErrRpcGeneric.
func TestRequestScanTasks_UnknownModuleLabelReturnsError(t *testing.T) {

	// Prepare unit test data
	const secret = "unknown-module-secret-thea14"
	memory.AddScope(secret, managerdb.T_scan_scope{
		Id:      106,
		Name:    "unknown-module-scope",
		Enabled: true,
		Cycles:  false,
	})

	// Prepare cleanup
	defer memory.RemoveScope(secret)

	// Run test case
	b := &Broker{}
	args := &ArgsGetScanTask{
		AgentInfo: AgentInfo{
			ApiVersion: BrokerApiVersion,
			Name:       "test-agent",
			Host:       "localhost",
			Ip:         "127.0.0.1",
		},
		ScopeSecret: secret,
		ModuleData: []ModuleData{
			{
				Label:          "not-a-real-module",
				MaxInstances:   -1,
				ScopeInstances: 0,
			},
		},
	}
	reply := &ReplyGetScanTask{}
	errRequest := b.RequestScanTasks(args, reply)

	// Verify results
	if errRequest != utils.ErrRpcGeneric {
		t.Errorf("RequestScanTasks() err = '%v', want = '%v'", errRequest, utils.ErrRpcGeneric)
	}
	if len(reply.ScanTasks) != 0 {
		t.Errorf("RequestScanTasks() scan tasks = '%v', want = '%v'", len(reply.ScanTasks), 0)
	}
}
