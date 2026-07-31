/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"encoding/json"
	"testing"
)

// validScanSettingJSON is a minimal valid scan-settings payload with all required fields populated.
const validScanSettingJSON = `{
	"max_instances_discovery": 5,
	"max_instances_banner": 5,
	"max_instances_nfs": 5,
	"max_instances_nuclei": 5,
	"max_instances_smb": 5,
	"max_instances_ssh": 5,
	"max_instances_ssl": 5,
	"max_instances_webcrawler": 5,
	"max_instances_webenum": 5,
	"network_timeout_seconds": 30,
	"nfs_scan_timeout_minutes": 10,
	"nfs_threads": 4,
	"nuclei_scan_timeout_minutes": 5,
	"smb_scan_timeout_minutes": 10,
	"smb_threads": 4,
	"ssl_scan_timeout_minutes": 5,
	"ssh_scan_timeout_minutes": 5,
	"webcrawler_scan_timeout_minutes": 5,
	"webcrawler_max_threads": 2,
	"webenum_scan_timeout_minutes": 5
}`

// patchJSON appends a key-value pair to a JSON object string, letting the later value override any earlier one.
func patchJSON(base, key, value string) string {
	return base[:len(base)-1] + `, "` + key + `": ` + value + `}`
}

// TestScanSettingMaxInstances verifies MaxInstances returns the correct count for each known module label.
func TestScanSettingMaxInstances(t *testing.T) {
	settingsEntry := T_scan_setting{
		MaxInstancesDiscovery:  10,
		MaxInstancesBanner:     20,
		MaxInstancesNfs:        30,
		MaxInstancesNuclei:     40,
		MaxInstancesSmb:        50,
		MaxInstancesSsh:        60,
		MaxInstancesSsl:        70,
		MaxInstancesWebcrawler: 80,
		MaxInstancesWebenum:    90,
	}

	tests := []struct {
		name  string
		label string
		want  int
	}{
		{
			name:  "discovery",
			label: "Discovery",
			want:  10,
		},
		{
			name:  "banner",
			label: "Banner",
			want:  20,
		},
		{
			name:  "nfs",
			label: "Nfs",
			want:  30,
		},
		{
			name:  "nuclei",
			label: "Nuclei",
			want:  40,
		},
		{
			name:  "smb",
			label: "Smb",
			want:  50,
		},
		{
			name:  "ssh",
			label: "Ssh",
			want:  60,
		},
		{
			name:  "ssl",
			label: "Ssl",
			want:  70,
		},
		{
			name:  "webcrawler",
			label: "Webcrawler",
			want:  80,
		},
		{
			name:  "webenum",
			label: "Webenum",
			want:  90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGot := settingsEntry.MaxInstances(tt.label)
			if errGot != nil {
				t.Errorf("MaxInstances(%q) error = '%v'", tt.label, errGot)
			}
			if got != tt.want {
				t.Errorf("MaxInstances(%q) = '%v', want = '%v'", tt.label, got, tt.want)
			}
		})
	}

	_, errUnknown := settingsEntry.MaxInstances("unknown-module")
	if errUnknown == nil {
		t.Errorf("MaxInstances('unknown-module') expected error, got = 'nil'")
	}
}

// TestScanSettingSaveAll verifies SaveAll persists updated scan settings to the manager database.
func TestScanSettingSaveAll(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.2", 5432)
	scopeEntry := insertTestScope(t, srv, "settings-test", "db_settings_001", "secret-st-001")

	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryLoaded != nil || scopeEntryLoaded == nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errScopeEntryLoaded)
	}

	updated := defaultScanSettings()
	updated.MaxInstancesDiscovery = 99
	rows, errRows := scopeEntryLoaded.ScanSettings.SaveAll(&updated)
	if errRows != nil {
		t.Fatalf("ScanSettings.SaveAll() error = '%v'", errRows)
	}
	if rows == 0 {
		t.Errorf("ScanSettings.SaveAll() rows = '0', want = '>0'")
	}
}

// TestScanSettingUnmarshalJSON_Valid verifies valid JSON unmarshals without error.
func TestScanSettingUnmarshalJSON_Valid(t *testing.T) {
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(validScanSettingJSON), &settingsEntry); errUnmarshal != nil {
		t.Fatalf("UnmarshalJSON() error = '%v'", errUnmarshal)
	}
	if settingsEntry.MaxInstancesDiscovery != 5 {
		t.Errorf("UnmarshalJSON() MaxInstancesDiscovery = '%v', want = '5'", settingsEntry.MaxInstancesDiscovery)
	}
}

// TestScanSettingUnmarshalJSON_InvalidTimeout verifies zero network timeout is rejected.
func TestScanSettingUnmarshalJSON_InvalidTimeout(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "network_timeout_seconds", "0")
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal == nil {
		t.Errorf("UnmarshalJSON() with zero network_timeout_seconds expected error, got = 'nil'")
	}
}

// TestScanSettingUnmarshalJSON_InvalidFields verifies zero or negative constraint values are rejected.
func TestScanSettingUnmarshalJSON_InvalidFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
	}{
		{
			name:  "nfs-timeout-zero",
			field: "nfs_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "nfs-threads-zero",
			field: "nfs_threads",
			value: "0",
		},
		{
			name:  "nuclei-timeout-zero",
			field: "nuclei_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "smb-timeout-zero",
			field: "smb_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "smb-threads-zero",
			field: "smb_threads",
			value: "0",
		},
		{
			name:  "ssl-timeout-zero",
			field: "ssl_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "ssh-timeout-zero",
			field: "ssh_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "webcrawler-timeout-zero",
			field: "webcrawler_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "webcrawler-threads-zero",
			field: "webcrawler_max_threads",
			value: "0",
		},
		{
			name:  "webenum-timeout-zero",
			field: "webenum_scan_timeout_minutes",
			value: "0",
		},
		{
			name:  "discovery-negative",
			field: "max_instances_discovery",
			value: "-1",
		},
		{
			name:  "banner-negative",
			field: "max_instances_banner",
			value: "-1",
		},
		{
			name:  "nfs-negative",
			field: "max_instances_nfs",
			value: "-1",
		},
		{
			name:  "nuclei-negative",
			field: "max_instances_nuclei",
			value: "-1",
		},
		{
			name:  "smb-negative",
			field: "max_instances_smb",
			value: "-1",
		},
		{
			name:  "ssh-negative",
			field: "max_instances_ssh",
			value: "-1",
		},
		{
			name:  "ssl-negative",
			field: "max_instances_ssl",
			value: "-1",
		},
		{
			name:  "webcrawler-negative",
			field: "max_instances_webcrawler",
			value: "-1",
		},
		{
			name:  "webenum-negative",
			field: "max_instances_webenum",
			value: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := patchJSON(validScanSettingJSON, tt.field, tt.value)
			var settingsEntry T_scan_setting
			if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal == nil {
				t.Errorf("UnmarshalJSON() with %s=%s expected error, got = 'nil'", tt.field, tt.value)
			}
		})
	}
}

// TestScanSettingUnmarshalJSON_SensitivePorts verifies sensitive_ports slice is populated on unmarshal.
func TestScanSettingUnmarshalJSON_SensitivePorts(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "sensitive_ports", "[22, 443, 8080]")
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal != nil {
		t.Fatalf("UnmarshalJSON() with sensitive_ports error = '%v'", errUnmarshal)
	}
	if len(settingsEntry.SensitivePortsSlice) != 3 {
		t.Errorf("UnmarshalJSON() SensitivePortsSlice len = '%v', want = '3'", len(settingsEntry.SensitivePortsSlice))
	}
}

// TestScanSettingUnmarshalJSON_InvalidPort verifies out-of-range port numbers are rejected.
func TestScanSettingUnmarshalJSON_InvalidPort(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "sensitive_ports", "[99999]")
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal == nil {
		t.Errorf("UnmarshalJSON() with out-of-range port expected error, got = 'nil'")
	}
}

// TestScanSettingUnmarshalJSON_Timespans verifies discovery_timespans slice is populated on unmarshal.
func TestScanSettingUnmarshalJSON_Timespans(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "discovery_timespans",
		`[{"startDay":"1","endDay":"5","startTime":"08:00","endTime":"18:00"}]`)
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal != nil {
		t.Fatalf("UnmarshalJSON() with valid timespans error = '%v'", errUnmarshal)
	}
	if len(settingsEntry.DiscoveryTimespansSlice) != 1 {
		t.Errorf("UnmarshalJSON() DiscoveryTimespansSlice len = '%v', want = '1'", len(settingsEntry.DiscoveryTimespansSlice))
	}
}

// TestScanSettingUnmarshalJSON_InvalidTimespanDay verifies out-of-range day values are rejected.
func TestScanSettingUnmarshalJSON_InvalidTimespanDay(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "discovery_timespans",
		`[{"startDay":"9","endDay":"5","startTime":"08:00","endTime":"18:00"}]`)
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal == nil {
		t.Errorf("UnmarshalJSON() with startDay=9 expected error, got = 'nil'")
	}
}

// TestScanSettingUnmarshalJSON_InvalidTimespanTime verifies malformed time strings are rejected.
func TestScanSettingUnmarshalJSON_InvalidTimespanTime(t *testing.T) {
	raw := patchJSON(validScanSettingJSON, "discovery_timespans",
		`[{"startDay":"1","endDay":"5","startTime":"bad","endTime":"18:00"}]`)
	var settingsEntry T_scan_setting
	if errUnmarshal := json.Unmarshal([]byte(raw), &settingsEntry); errUnmarshal == nil {
		t.Errorf("UnmarshalJSON() with invalid startTime expected error, got = 'nil'")
	}
}

// TestScanSettingAfterFind_SensitivePorts verifies AfterFind populates SensitivePortsSlice from the stored column.
func TestScanSettingAfterFind_SensitivePorts(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.91", 5432)
	scopeEntry := insertTestScope(t, srv, "afterfind-test", "db_afterfind_001", "secret-af-001")

	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryLoaded != nil || scopeEntryLoaded == nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errScopeEntryLoaded)
	}

	scopeEntryLoaded.ScanSettings.SensitivePorts = "22, 443, 8080"
	if _, errSave := scopeEntryLoaded.ScanSettings.SaveAll(&scopeEntryLoaded.ScanSettings); errSave != nil {
		t.Fatalf("ScanSettings.SaveAll() error = '%v'", errSave)
	}

	scopeEntryReloaded, errScopeEntryReloaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryReloaded != nil || scopeEntryReloaded == nil {
		t.Fatalf("GetScopeEntry() reload error = '%v'", errScopeEntryReloaded)
	}
	if len(scopeEntryReloaded.ScanSettings.SensitivePortsSlice) == 0 {
		t.Errorf("AfterFind() SensitivePortsSlice len = '0', want = '>0'")
	}
}

// TestScanSettingAfterFind_DiscoveryTimespans verifies AfterFind populates DiscoveryTimespansSlice from the stored column.
func TestScanSettingAfterFind_DiscoveryTimespans(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.92", 5432)
	scopeEntry := insertTestScope(t, srv, "aftertimespans-test", "db_aftertimespans_001", "secret-at-001")

	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryLoaded != nil || scopeEntryLoaded == nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errScopeEntryLoaded)
	}

	scopeEntryLoaded.ScanSettings.DiscoveryTimespans = `[{"startDay":"1","endDay":"5","startTime":"08:00","endTime":"18:00"}]`
	if _, errSave := scopeEntryLoaded.ScanSettings.SaveAll(&scopeEntryLoaded.ScanSettings); errSave != nil {
		t.Fatalf("ScanSettings.SaveAll() error = '%v'", errSave)
	}

	scopeEntryReloaded, errScopeEntryReloaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryReloaded != nil || scopeEntryReloaded == nil {
		t.Fatalf("GetScopeEntry() reload error = '%v'", errScopeEntryReloaded)
	}
	if len(scopeEntryReloaded.ScanSettings.DiscoveryTimespansSlice) == 0 {
		t.Errorf("AfterFind() DiscoveryTimespansSlice len = '0', want = '>0'")
	}
}
