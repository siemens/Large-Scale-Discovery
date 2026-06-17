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
	"os"
	"testing"
	"time"

	"github.com/siemens/Large-Scale-Discovery/utils"
)

// insertTestDbServer registers a test database server and schedules cleanup via t.Cleanup.
func insertTestDbServer(t *testing.T, host string, port int) *T_db_server {
	t.Helper()
	s := &T_db_server{
		Name:    "test-db-" + host,
		Dialect: "postgres",
		Host:    host,
		Port:    port,
		Admin:   "admin",
	}
	if errUpdate := UpdateDatabaseEntry(s); errUpdate != nil {
		t.Fatalf("insertTestDbServer: %v", errUpdate)
	}
	t.Cleanup(func() { _ = RemoveDatabaseEntry(s.Id) })
	return s
}

// defaultScanSettings returns a T_scan_setting with all max-instance fields set to 1.
func defaultScanSettings() T_scan_setting {
	return T_scan_setting{
		MaxInstancesDiscovery:  1,
		MaxInstancesBanner:     1,
		MaxInstancesNfs:        1,
		MaxInstancesNuclei:     1,
		MaxInstancesSmb:        1,
		MaxInstancesSsh:        1,
		MaxInstancesSsl:        1,
		MaxInstancesWebcrawler: 1,
		MaxInstancesWebenum:    1,
	}
}

// insertTestScope creates a scope associated with the given server and schedules cleanup via t.Cleanup.
func insertTestScope(t *testing.T, srv *T_db_server, name, dbName, secret string) *T_scan_scope {
	t.Helper()
	scope, errCreate := createScopeEntry(
		managerDb,
		srv,
		name,
		dbName,
		42,
		"testuser",
		secret,
		"custom",
		false,
		-1,
		utils.JsonMap{},
		0,
		defaultScanSettings(),
	)
	if errCreate != nil {
		t.Fatalf("insertTestScope: %v", errCreate)
	}
	t.Cleanup(func() { _ = managerDb.Delete(scope) })
	return scope
}

// insertTestViewEntry creates a view entry for the given scope and schedules cleanup via t.Cleanup.
func insertTestViewEntry(t *testing.T, scope *T_scan_scope, name string) *T_scope_view {
	t.Helper()
	view, errCreate := createViewEntry(managerDb, scope, name, "testuser", map[string][]string{}, []string{"hosts_" + name})
	if errCreate != nil {
		t.Fatalf("insertTestViewEntry: %v", errCreate)
	}
	t.Cleanup(func() { _ = managerDb.Delete(view) })
	return view
}

// TestGetDatabaseEntries_Empty verifies GetDatabaseEntries returns an empty slice when no servers are registered.
func TestGetDatabaseEntries_Empty(t *testing.T) {
	entries, errEntries := GetDatabaseEntries()
	if errEntries != nil {
		t.Fatalf("GetDatabaseEntries() error = '%v'", errEntries)
	}
	if len(entries) != 0 {
		t.Errorf("GetDatabaseEntries() len = '%v', want = '0'", len(entries))
	}
}

// TestUpdateAndGetDatabaseEntry verifies creating and querying a DB server entry round-trips correctly.
func TestUpdateAndGetDatabaseEntry(t *testing.T) {
	entry := &T_db_server{
		Name:    "test-server",
		Dialect: "postgres",
		Host:    "127.0.0.1",
		Port:    5432,
		Admin:   "admin",
	}
	if errUpdate := UpdateDatabaseEntry(entry); errUpdate != nil {
		t.Fatalf("UpdateDatabaseEntry() error = '%v'", errUpdate)
	}
	if entry.Id == 0 {
		t.Errorf("UpdateDatabaseEntry() Id = '0', want = 'non-zero after create'")
	}
	defer func() { _ = RemoveDatabaseEntry(entry.Id) }()

	got, errGet := GetDatabaseEntry(entry.Id)
	if errGet != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGet)
	}
	if got == nil {
		t.Fatal("GetDatabaseEntry() returned nil")
	}
	if got.Name != "test-server" {
		t.Errorf("GetDatabaseEntry() name = '%v', want = 'test-server'", got.Name)
	}

	entry.Name = "updated-server"
	if errUpdate2 := UpdateDatabaseEntry(entry); errUpdate2 != nil {
		t.Fatalf("UpdateDatabaseEntry() update error = '%v'", errUpdate2)
	}

	got2, errGet2 := GetDatabaseEntry(entry.Id)
	if errGet2 != nil {
		t.Fatalf("GetDatabaseEntry() after update error = '%v'", errGet2)
	}
	if got2 == nil || got2.Name != "updated-server" {
		t.Errorf("GetDatabaseEntry() name after update = '%v', want = 'updated-server'", got2.Name)
	}
}

// TestGetDatabaseEntry_NotFound verifies GetDatabaseEntry returns nil for an unknown ID.
func TestGetDatabaseEntry_NotFound(t *testing.T) {
	got, errGet := GetDatabaseEntry(999999)
	if errGet != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGet)
	}
	if got != nil {
		t.Errorf("GetDatabaseEntry() = '%v', want = 'nil' for non-existent ID", got)
	}
}

// TestRemoveDatabaseEntry verifies a registered server is absent after removal.
func TestRemoveDatabaseEntry(t *testing.T) {
	entry := &T_db_server{
		Name:    "remove-me",
		Dialect: "postgres",
		Host:    "127.0.0.1",
		Port:    5433,
		Admin:   "admin",
	}
	if errUpdate := UpdateDatabaseEntry(entry); errUpdate != nil {
		t.Fatalf("UpdateDatabaseEntry() error = '%v'", errUpdate)
	}

	if errRemove := RemoveDatabaseEntry(entry.Id); errRemove != nil {
		t.Fatalf("RemoveDatabaseEntry() error = '%v'", errRemove)
	}

	got, errGet := GetDatabaseEntry(entry.Id)
	if errGet != nil {
		t.Fatalf("GetDatabaseEntry() after remove error = '%v'", errGet)
	}
	if got != nil {
		t.Errorf("GetDatabaseEntry() after remove = '%v', want = 'nil'", got)
	}
}

// TestGetDatabaseEntries verifies GetDatabaseEntries returns all registered servers.
func TestGetDatabaseEntries(t *testing.T) {
	e1 := &T_db_server{Name: "srv-a", Dialect: "postgres", Host: "192.0.2.101", Port: 5432, Admin: "a"}
	e2 := &T_db_server{Name: "srv-b", Dialect: "postgres", Host: "192.0.2.102", Port: 5432, Admin: "b"}
	_ = UpdateDatabaseEntry(e1)
	_ = UpdateDatabaseEntry(e2)
	defer func() {
		_ = RemoveDatabaseEntry(e1.Id)
		_ = RemoveDatabaseEntry(e2.Id)
	}()

	entries, errEntries := GetDatabaseEntries()
	if errEntries != nil {
		t.Fatalf("GetDatabaseEntries() error = '%v'", errEntries)
	}
	if len(entries) < 2 {
		t.Errorf("GetDatabaseEntries() len = '%v', want = '>=2'", len(entries))
	}
}

// TestGetDatabaseEntryByName verifies lookup by name returns the correct entry and nil for unknowns.
func TestGetDatabaseEntryByName(t *testing.T) {
	entry := &T_db_server{
		Name:    "named-server",
		Dialect: "postgres",
		Host:    "192.0.2.8",
		Port:    5432,
		Admin:   "admin",
	}
	if errUpdate := UpdateDatabaseEntry(entry); errUpdate != nil {
		t.Fatalf("UpdateDatabaseEntry() error = '%v'", errUpdate)
	}
	defer func() { _ = RemoveDatabaseEntry(entry.Id) }()

	got, errGet := getDatabaseEntryByName("named-server")
	if errGet != nil {
		t.Fatalf("getDatabaseEntryByName() error = '%v'", errGet)
	}
	if got == nil {
		t.Fatal("getDatabaseEntryByName() returned nil")
	}
	if got.Name != "named-server" {
		t.Errorf("getDatabaseEntryByName() name = '%v', want = 'named-server'", got.Name)
	}

	missing, errGetMissing := getDatabaseEntryByName("does-not-exist")
	if errGetMissing != nil {
		t.Fatalf("getDatabaseEntryByName() missing error = '%v'", errGetMissing)
	}
	if missing != nil {
		t.Errorf("getDatabaseEntryByName() = '%v', want = 'nil' for unknown name", missing)
	}
}

// TestGetScopeEntry_NotFound verifies GetScopeEntry returns nil for an unknown scope ID.
func TestGetScopeEntry_NotFound(t *testing.T) {
	got, errGet := GetScopeEntry(999999)
	if errGet != nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errGet)
	}
	if got != nil {
		t.Errorf("GetScopeEntry() = '%v', want = 'nil' for non-existent ID", got)
	}
}

// TestGetScopeEntryBySecret_NotFound verifies GetScopeEntryBySecret returns nil for an unknown secret.
func TestGetScopeEntryBySecret_NotFound(t *testing.T) {
	got, errGet := GetScopeEntryBySecret("non-existent-secret")
	if errGet != nil {
		t.Fatalf("GetScopeEntryBySecret() error = '%v'", errGet)
	}
	if got != nil {
		t.Errorf("GetScopeEntryBySecret() = '%v', want = 'nil' for unknown secret", got)
	}
}

// TestGetScopeEntryByName_NotFound verifies GetScopeEntryByName returns nil for an unknown DB name.
func TestGetScopeEntryByName_NotFound(t *testing.T) {
	got, errGet := GetScopeEntryByName("non-existent-db-name")
	if errGet != nil {
		t.Fatalf("GetScopeEntryByName() error = '%v'", errGet)
	}
	if got != nil {
		t.Errorf("GetScopeEntryByName() = '%v', want = 'nil' for unknown DB name", got)
	}
}

// TestScopeCRUD verifies create, get, lookup by secret/name, list, ID query, and Save on scopes.
func TestScopeCRUD(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.1", 5432)
	scope := insertTestScope(t, srv, "My Scope", "db_myscope_001", "secret-abc-001")

	got, errGet := GetScopeEntry(scope.Id)
	if errGet != nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errGet)
	}
	if got == nil {
		t.Fatal("GetScopeEntry() returned nil")
	}
	if got.Name != "My Scope" {
		t.Errorf("GetScopeEntry() name = '%v', want = 'My Scope'", got.Name)
	}

	gotSecret, errSecret := GetScopeEntryBySecret("secret-abc-001")
	if errSecret != nil {
		t.Fatalf("GetScopeEntryBySecret() error = '%v'", errSecret)
	}
	if gotSecret == nil || gotSecret.Id != scope.Id {
		t.Errorf("GetScopeEntryBySecret() Id = '%v', want = '%v'", gotSecret.Id, scope.Id)
	}

	gotName, errName := GetScopeEntryByName("db_myscope_001")
	if errName != nil {
		t.Fatalf("GetScopeEntryByName() error = '%v'", errName)
	}
	if gotName == nil || gotName.Id != scope.Id {
		t.Errorf("GetScopeEntryByName() Id = '%v', want = '%v'", gotName.Id, scope.Id)
	}

	ids, errIds := GetScopeEntryIds()
	if errIds != nil {
		t.Fatalf("GetScopeEntryIds() error = '%v'", errIds)
	}
	found := false
	for _, id := range ids {
		if id == scope.Id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetScopeEntryIds() did not include scope Id = '%v'", scope.Id)
	}

	scopesOf, errOf := GetScopeEntriesOf([]uint64{42})
	if errOf != nil {
		t.Fatalf("GetScopeEntriesOf() error = '%v'", errOf)
	}
	if len(scopesOf) == 0 {
		t.Errorf("GetScopeEntriesOf() len = '0', want = '>0' for matching group")
	}

	scopesEmpty, errEmpty := GetScopeEntriesOf([]uint64{})
	if errEmpty != nil {
		t.Fatalf("GetScopeEntriesOf([]) error = '%v'", errEmpty)
	}
	if len(scopesEmpty) != 0 {
		t.Errorf("GetScopeEntriesOf([]) len = '%v', want = '0'", len(scopesEmpty))
	}

	all, errAll := GetScopeEntries()
	if errAll != nil {
		t.Fatalf("GetScopeEntries() error = '%v'", errAll)
	}
	if len(all) == 0 {
		t.Errorf("GetScopeEntries() len = '0', want = '>0' after insert")
	}

	scope.Name = "Updated Scope"
	rows, errSave := scope.Save("name")
	if errSave != nil {
		t.Fatalf("scope.Save() error = '%v'", errSave)
	}
	if rows == 0 {
		t.Errorf("scope.Save() rows = '0', want = '>0'")
	}

	got2, errGet2 := GetScopeEntry(scope.Id)
	if errGet2 != nil {
		t.Fatalf("GetScopeEntry() after Save error = '%v'", errGet2)
	}
	if got2 == nil || got2.Name != "Updated Scope" {
		t.Errorf("GetScopeEntry() name after Save = '%v', want = 'Updated Scope'", got2.Name)
	}
}

// TestScopeSave_NoColumns verifies Save returns an error when called without column arguments.
func TestScopeSave_NoColumns(t *testing.T) {
	scope := &T_scan_scope{}
	_, errSave := scope.Save()
	if errSave == nil {
		t.Errorf("scope.Save() with no columns expected error, got = 'nil'")
	}
}

// TestUpdateScanAgents_Empty verifies UpdateScanAgents succeeds with an empty agent slice.
func TestUpdateScanAgents_Empty(t *testing.T) {
	if errUpdate := UpdateScanAgents(1, []T_scan_agent{}); errUpdate != nil {
		t.Fatalf("UpdateScanAgents([]) error = '%v'", errUpdate)
	}
}

// TestUpdateAndDeleteAgent verifies inserting and then deleting a scan agent.
func TestUpdateAndDeleteAgent(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.3", 5432)
	scope := insertTestScope(t, srv, "agent-test", "db_agent_001", "secret-ag-001")

	agents := []T_scan_agent{{
		Name:     "agent-1",
		Host:     "host-1",
		Ip:       "192.0.2.201",
		LastSeen: time.Now(),
		Tasks:    utils.JsonMap{},
	}}
	if errUpdate := UpdateScanAgents(scope.Id, agents); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() error = '%v'", errUpdate)
	}

	all, errAll := GetAgentEntries()
	if errAll != nil {
		t.Fatalf("GetAgentEntries() error = '%v'", errAll)
	}
	var agentId uint64
	for _, a := range all {
		if a.Name == "agent-1" && a.IdTScanScope == scope.Id {
			agentId = a.Id
			break
		}
	}
	if agentId == 0 {
		t.Fatal("GetAgentEntries() did not include inserted agent")
	}

	if errDelete := DeleteAgent(agentId); errDelete != nil {
		t.Fatalf("DeleteAgent() error = '%v'", errDelete)
	}
}

// TestScanAgentSave verifies Save updates a column and returns an error when no columns are specified.
func TestScanAgentSave(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.4", 5432)
	scope := insertTestScope(t, srv, "agent-save-test", "db_agentsave_001", "secret-as-001")

	agents := []T_scan_agent{{
		Name:     "agent-save",
		Host:     "host-save",
		Ip:       "192.0.2.211",
		LastSeen: time.Now(),
		Tasks:    utils.JsonMap{},
	}}
	if errUpdate := UpdateScanAgents(scope.Id, agents); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() error = '%v'", errUpdate)
	}

	all, errAll := GetAgentEntries()
	if errAll != nil {
		t.Fatalf("GetAgentEntries() error = '%v'", errAll)
	}
	var agent *T_scan_agent
	for i := range all {
		if all[i].Name == "agent-save" && all[i].IdTScanScope == scope.Id {
			agent = &all[i]
			break
		}
	}
	if agent == nil {
		t.Fatal("GetAgentEntries() did not include inserted agent")
	}
	defer func() { _ = DeleteAgent(agent.Id) }()

	agent.Ip = "192.0.2.212"
	rows, errSave := agent.Save("ip")
	if errSave != nil {
		t.Fatalf("agent.Save() error = '%v'", errSave)
	}
	if rows == 0 {
		t.Errorf("agent.Save() rows = '0', want = '>0'")
	}

	_, errSaveEmpty := agent.Save()
	if errSaveEmpty == nil {
		t.Errorf("agent.Save() with no columns expected error, got = 'nil'")
	}
}

// TestViewEntriesCRUD verifies create, get, list, exists, update, and grant operations on view entries.
func TestViewEntriesCRUD(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.5", 5432)
	scope := insertTestScope(t, srv, "view-test", "db_view_001", "secret-vw-001")
	view := insertTestViewEntry(t, scope, "myview")

	got, errGet := GetViewEntry(view.Id)
	if errGet != nil {
		t.Fatalf("GetViewEntry() error = '%v'", errGet)
	}
	if got == nil {
		t.Fatal("GetViewEntry() returned nil")
	}
	if got.Name != "myview" {
		t.Errorf("GetViewEntry() name = '%v', want = 'myview'", got.Name)
	}

	missing, errMissing := GetViewEntry(999999)
	if errMissing != nil {
		t.Fatalf("GetViewEntry(missing) error = '%v'", errMissing)
	}
	if missing != nil {
		t.Errorf("GetViewEntry(missing) = '%v', want = 'nil'", missing)
	}

	views, errViews := GetViewEntries()
	if errViews != nil {
		t.Fatalf("GetViewEntries() error = '%v'", errViews)
	}
	if len(views) == 0 {
		t.Errorf("GetViewEntries() len = '0', want = '>0' after insert")
	}

	_, errOf := GetViewEntriesOf([]uint64{42})
	if errOf != nil {
		t.Fatalf("GetViewEntriesOf() error = '%v'", errOf)
	}

	emptyOf, errEmptyOf := GetViewEntriesOf([]uint64{})
	if errEmptyOf != nil {
		t.Fatalf("GetViewEntriesOf([]) error = '%v'", errEmptyOf)
	}
	if len(emptyOf) != 0 {
		t.Errorf("GetViewEntriesOf([]) len = '%v', want = '0'", len(emptyOf))
	}

	exists, errExists := ViewExists(scope.Id, "myview")
	if errExists != nil {
		t.Fatalf("ViewExists() error = '%v'", errExists)
	}
	if !exists {
		t.Errorf("ViewExists() = 'false', want = 'true'")
	}

	absent, errAbsent := ViewExists(scope.Id, "non-existing-view")
	if errAbsent != nil {
		t.Fatalf("ViewExists(absent) error = '%v'", errAbsent)
	}
	if absent {
		t.Errorf("ViewExists(absent) = 'true', want = 'false'")
	}

	view.Name = "myview-updated"
	rows, errSave := view.Save("name")
	if errSave != nil {
		t.Fatalf("view.Save() error = '%v'", errSave)
	}
	if rows == 0 {
		t.Errorf("view.Save() rows = '0', want = '>0'")
	}

	_, errSaveEmpty := view.Save()
	if errSaveEmpty == nil {
		t.Errorf("view.Save() with no columns expected error, got = 'nil'")
	}

	granted, errGranted := GetViewsGranted("nobody@domain.tld")
	if errGranted != nil {
		t.Fatalf("GetViewsGranted() error = '%v'", errGranted)
	}
	_ = granted
}

// TestCreateGrantEntry verifies a grant entry is created and reflected in credential checks and view listings.
func TestCreateGrantEntry(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.6", 5432)
	scope := insertTestScope(t, srv, "grant-test", "db_grant_001", "secret-gr-001")
	view := insertTestViewEntry(t, scope, "grantview")

	expiry := time.Now().Add(24 * time.Hour)
	grant, errGrant := createGrantEntry(managerDb, view, true, "user@domain.tld", "testuser", expiry, "test grant")
	if errGrant != nil {
		t.Fatalf("createGrantEntry() error = '%v'", errGrant)
	}
	if grant.Id == 0 {
		t.Errorf("createGrantEntry() Id = '0', want = 'non-zero'")
	}
	defer func() { _ = managerDb.Delete(grant) }()

	required, errRequired := databaseCredentialsRequired(managerDb, "user@domain.tld", srv.Id)
	if errRequired != nil {
		t.Fatalf("databaseCredentialsRequired() error = '%v'", errRequired)
	}
	if !required {
		t.Errorf("databaseCredentialsRequired() = 'false', want = 'true'")
	}

	_, errBadId := databaseCredentialsRequired(managerDb, "user@domain.tld", 0)
	if errBadId == nil {
		t.Errorf("databaseCredentialsRequired(dbServerId=0) expected error, got = 'nil'")
	}

	viewsGranted, errGranted := GetViewsGranted("user@domain.tld")
	if errGranted != nil {
		t.Fatalf("GetViewsGranted() error = '%v'", errGranted)
	}
	if len(viewsGranted) == 0 {
		t.Errorf("GetViewsGranted() len = '0', want = '>0' after grant")
	}
}

// TestUpdateScopeInstances verifies updating scan instance counts and expected error cases.
func TestUpdateScopeInstances(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.7", 5432)
	scope := insertTestScope(t, srv, "instances-test", "db_instances_001", "secret-in-001")

	errUpdate := updateScopeInstances(managerDb, scope.Id, map[string]uint32{
		"Discovery":  5,
		"Banner":     3,
		"Nfs":        2,
		"Nuclei":     1,
		"Smb":        2,
		"Ssh":        2,
		"Ssl":        2,
		"Webcrawler": 2,
		"Webenum":    2,
	})
	if errUpdate != nil {
		t.Fatalf("updateScopeInstances() error = '%v'", errUpdate)
	}

	errUnknown := updateScopeInstances(managerDb, scope.Id, map[string]uint32{"unknown-module": 5})
	if errUnknown == nil {
		t.Errorf("updateScopeInstances(unknown-module) expected error, got = 'nil'")
	}

	errBadScope := updateScopeInstances(managerDb, 999999, map[string]uint32{"Discovery": 1})
	if errBadScope == nil {
		t.Errorf("updateScopeInstances(nonexistent-scope) expected error, got = 'nil'")
	}
}

// TestCreateDatabaseEntry verifies createDatabaseEntry inserts a new server entry and assigns a non-zero ID.
func TestCreateDatabaseEntry(t *testing.T) {
	entry, errCreate := createDatabaseEntry(
		managerDb,
		"entry-via-create",
		"postgres",
		"db.domain.tld",
		5432,
		"admin",
		"",
		"",
		"",
	)
	if errCreate != nil {
		t.Fatalf("createDatabaseEntry() error = '%v'", errCreate)
	}
	if entry == nil || entry.Id == 0 {
		t.Fatal("createDatabaseEntry() returned nil or Id = '0'")
	}
	defer func() { _ = RemoveDatabaseEntry(entry.Id) }()

	got, errGet := GetDatabaseEntry(entry.Id)
	if errGet != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGet)
	}
	if got == nil || got.Name != "entry-via-create" {
		t.Errorf("GetDatabaseEntry() name = '%v', want = 'entry-via-create'", got.Name)
	}
}

// TestCloseManagerDb verifies CloseManagerDb closes an open manager DB without error.
func TestCloseManagerDb(t *testing.T) {

	// Save global connection so we can restore it after this test
	savedDb := managerDb

	// Set up a temporary directory with its own SQLite file
	tmpDir := t.TempDir()
	prevDir, errWd := os.Getwd()
	if errWd != nil {
		t.Fatalf("TestCloseManagerDb: could not get working directory: '%v'", errWd)
	}
	if errChdir := os.Chdir(tmpDir); errChdir != nil {
		t.Fatalf("TestCloseManagerDb: could not chdir to tmpDir: '%v'", errChdir)
	}
	defer func() {
		managerDb = savedDb
		_ = os.Chdir(prevDir)
	}()

	// Open a fresh connection in the temp directory
	if errOpen := OpenManagerDb(); errOpen != nil {
		t.Fatalf("OpenManagerDb() error = '%v'", errOpen)
	}

	// Close the fresh connection — this is the code under test
	if errClose := CloseManagerDb(); errClose != nil {
		t.Errorf("CloseManagerDb() error = '%v'", errClose)
	}
}
