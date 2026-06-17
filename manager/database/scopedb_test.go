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
	"testing"

	scanUtils "github.com/siemens/GoScans/utils"
	_test "github.com/siemens/Large-Scale-Discovery/_test"
)

// makeDiscovery returns a T_discovery with the given input value set.
func makeDiscovery(input string) T_discovery {
	d := T_discovery{}
	d.Input = input
	return d
}

// TestSanitizeViewName verifies view names are lower-cased and non-alphanumeric characters become underscores.
func TestSanitizeViewName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "plain-upper",
			input:   "MyView",
			want:    "myview",
			wantErr: false,
		},
		{
			name:    "space-to-underscore",
			input:   "My View",
			want:    "my_view",
			wantErr: false,
		},
		{
			name:    "hyphen-to-underscore",
			input:   "My-View",
			want:    "my_view",
			wantErr: false,
		},
		{
			name:    "digits-preserved",
			input:   "my-view-2024",
			want:    "my_view_2024",
			wantErr: false,
		},
		{
			name:    "specials-stripped",
			input:   "a!@#$b",
			want:    "ab",
			wantErr: false,
		},
		{
			name:    "underscores-kept",
			input:   "__keep_underscores__",
			want:    "__keep_underscores__",
			wantErr: false,
		},
		{
			name:    "all-upper",
			input:   "UPPER CASE",
			want:    "upper_case",
			wantErr: false,
		},
		{
			name:    "mixed",
			input:   "mixed Case-AND-Spaces",
			want:    "mixed_case_and_spaces",
			wantErr: false,
		},
		{
			name:    "only-spaces",
			input:   "   ",
			want:    "___",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errSanitize := sanitizeViewName(tt.input)
			if (errSanitize != nil) != tt.wantErr {
				t.Errorf("sanitizeViewName(%q) error = '%v', wantErr = '%v'", tt.input, errSanitize, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sanitizeViewName(%q) = '%v', want = '%v'", tt.input, got, tt.want)
			}
		})
	}
}

// TestMergeInputs verifies mergeInputs correctly classifies entries as create, remove, or update.
func TestMergeInputs(t *testing.T) {
	tests := []struct {
		name       string
		existing   map[string]T_discovery
		newInputs  map[string]T_discovery
		wantCreate int
		wantRemove int
		wantUpdate int
	}{
		{
			name:       "all-new",
			existing:   map[string]T_discovery{},
			newInputs:  map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1"), "192.0.2.2": makeDiscovery("192.0.2.2")},
			wantCreate: 2,
			wantRemove: 0,
			wantUpdate: 0,
		},
		{
			name:       "all-removed",
			existing:   map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1"), "192.0.2.2": makeDiscovery("192.0.2.2")},
			newInputs:  map[string]T_discovery{},
			wantCreate: 0,
			wantRemove: 2,
			wantUpdate: 0,
		},
		{
			name:       "all-updated",
			existing:   map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1")},
			newInputs:  map[string]T_discovery{"192.0.2.1": makeDiscovery("192.0.2.1")},
			wantCreate: 0,
			wantRemove: 0,
			wantUpdate: 1,
		},
		{
			name:       "mixed",
			existing:   map[string]T_discovery{"keep": makeDiscovery("keep"), "remove": makeDiscovery("remove")},
			newInputs:  map[string]T_discovery{"keep": makeDiscovery("keep"), "add": makeDiscovery("add")},
			wantCreate: 1,
			wantRemove: 1,
			wantUpdate: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			create, remove, update := mergeInputs(tt.existing, tt.newInputs)
			if len(create) != tt.wantCreate {
				t.Errorf("mergeInputs() create = '%v', want = '%v'", len(create), tt.wantCreate)
			}
			if len(remove) != tt.wantRemove {
				t.Errorf("mergeInputs() remove = '%v', want = '%v'", len(remove), tt.wantRemove)
			}
			if len(update) != tt.wantUpdate {
				t.Errorf("mergeInputs() update = '%v', want = '%v'", len(update), tt.wantUpdate)
			}
		})
	}
}

// TestSetConnectionName verifies SetConnectionName sanitizes spaces to underscores and stores the result.
func TestSetConnectionName(t *testing.T) {
	original := connectionName
	t.Cleanup(func() { connectionName = original })

	SetConnectionName("test client")
	if connectionName != "test_client" {
		t.Errorf("SetConnectionName() connectionName = '%v', want = 'test_client'", connectionName)
	}
}

// TestSetMaxConnectionsDefault verifies SetMaxConnectionsDefault updates the global connection limit.
func TestSetMaxConnectionsDefault(t *testing.T) {
	original := dbMaxConn
	t.Cleanup(func() { dbMaxConn = original })

	SetMaxConnectionsDefault(50)
	if dbMaxConn != 50 {
		t.Errorf("SetMaxConnectionsDefault() dbMaxConn = '%v', want = '50'", dbMaxConn)
	}
}

// TestErrInvalidCharacter_Error verifies ErrInvalidCharacter produces a non-empty error message.
func TestErrInvalidCharacter_Error(t *testing.T) {
	err := ErrInvalidCharacter{ParamName: "username", Value: "bad!value"}
	if err.Error() == "" {
		t.Errorf("ErrInvalidCharacter.Error() = '', want = 'non-empty'")
	}
}

// TestCloseScopeDbs_NoConnections verifies CloseScopeDbs returns nil when no scope DB handles are open.
func TestCloseScopeDbs_NoConnections(t *testing.T) {
	errs := CloseScopeDbs()
	if errs != nil {
		t.Errorf("CloseScopeDbs() = '%v', want = 'nil'", errs)
	}
}

// TestGetServerDbHandle_Nil verifies GetServerDbHandle returns an error when called with a nil server.
func TestGetServerDbHandle_Nil(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetServerDbHandle(testLogger, nil)
	if errHandle == nil {
		t.Errorf("GetServerDbHandle(nil) expected error, got = 'nil'")
	}
}

// TestGetServerDbHandle_UnsupportedDialect verifies GetServerDbHandle returns an error for unknown dialects.
func TestGetServerDbHandle_UnsupportedDialect(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetServerDbHandle(testLogger, &T_db_server{Dialect: "mysql"})
	if errHandle == nil {
		t.Errorf("GetServerDbHandle(mysql) expected error, got = 'nil'")
	}
}

// TestGetScopeDbHandle_Nil verifies GetScopeDbHandle returns an error when called with a nil scope.
func TestGetScopeDbHandle_Nil(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetScopeDbHandle(testLogger, nil)
	if errHandle == nil {
		t.Errorf("GetScopeDbHandle(nil) expected error, got = 'nil'")
	}
}

// TestGetScopeDbHandle_NoDbServer verifies GetScopeDbHandle returns an error when the scope has no DB server ID.
func TestGetScopeDbHandle_NoDbServer(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	_, errHandle := GetScopeDbHandle(testLogger, &T_scan_scope{})
	if errHandle == nil {
		t.Errorf("GetScopeDbHandle(empty-scope) expected error, got = 'nil'")
	}
}

// pgTestServer returns a T_db_server from _test/settings.go, skipping if PgHost is not configured.
func pgTestServer(t *testing.T) *T_db_server {
	t.Helper()

	settings := _test.GetSettings()
	if settings.PgHost == "" {
		t.Skip("Integration test skipped: PgHost not configured in _test/settings.go")
		return nil
	}

	return &T_db_server{
		Name:     "test-pg-server",
		Dialect:  "postgres",
		Host:     settings.PgHost,
		Port:     settings.PgPort,
		Admin:    settings.PgUser,
		Password: settings.PgPassword,
	}
}

// TestPg_GetServerDbHandle verifies GetServerDbHandle successfully connects to a real PostgreSQL server.
func TestPg_GetServerDbHandle(t *testing.T) {
	srv := pgTestServer(t)
	testLogger := scanUtils.NewTestLogger()

	serverDb, errHandle := GetServerDbHandle(testLogger, srv)
	if errHandle != nil {
		t.Fatalf("GetServerDbHandle() error = '%v'", errHandle)
	}
	if serverDb == nil {
		t.Fatal("GetServerDbHandle() returned nil")
	}

	sqlDb, errSql := serverDb.DB()
	if errSql != nil {
		t.Fatalf("GetServerDbHandle() sql DB error = '%v'", errSql)
	}
	if errPing := sqlDb.Ping(); errPing != nil {
		t.Fatalf("GetServerDbHandle() ping error = '%v'", errPing)
	}
}

// TestPg_InstallTrigramIndices verifies InstallTrigramIndices succeeds when applied to a connected PG server DB.
func TestPg_InstallTrigramIndices(t *testing.T) {
	srv := pgTestServer(t)
	testLogger := scanUtils.NewTestLogger()

	serverDb, errHandle := GetServerDbHandle(testLogger, srv)
	if errHandle != nil {
		t.Fatalf("GetServerDbHandle() error = '%v'", errHandle)
	}

	if errInstall := InstallTrigramIndices(serverDb); errInstall != nil {
		t.Errorf("InstallTrigramIndices() error = '%v'", errInstall)
	}
}

// TestAutomigrateScanScopes_NoScopes verifies AutomigrateScanScopes returns nil when no scan scopes are registered.
func TestAutomigrateScanScopes_NoScopes(t *testing.T) {
	testLogger := scanUtils.NewTestLogger()

	if errMigrate := AutomigrateScanScopes(testLogger); errMigrate != nil {
		t.Errorf("AutomigrateScanScopes() with empty scope list error = '%v'", errMigrate)
	}
}
