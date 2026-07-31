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
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

// TestWithTransaction_CallbackError_RollsBackChanges verifies opaque transactions preserve rollback behavior
func TestWithTransaction_CallbackError_RollsBackChanges(t *testing.T) {

	// Prepare the error that should abort the transaction
	errExpected := errors.New("abort transaction")
	databaseName := "rollback-test"

	// Create an entry and deliberately abort before commit
	errTransaction := WithTransaction(func(transaction *Transaction) error {
		_, errDatabaseEntry := transaction.CreateDatabaseEntry(
			databaseName,
			"postgres",
			"localhost",
			5432,
			"postgres",
			"password",
			"localhost",
			"sslmode=disable",
		)
		if errDatabaseEntry != nil {
			return errDatabaseEntry
		}

		// Force the transaction to roll back
		return errExpected
	})
	if !errors.Is(errTransaction, errExpected) {
		t.Fatalf("WithTransaction() error = '%v', want = '%v'", errTransaction, errExpected)
	}

	// Verify the aborted entry was not persisted
	databaseEntry, errDatabaseEntry := GetDatabaseEntryByName(databaseName)
	if errDatabaseEntry != nil {
		t.Fatalf("GetDatabaseEntryByName() error = '%v', want = 'nil'", errDatabaseEntry)
	}
	if databaseEntry != nil {
		t.Errorf("GetDatabaseEntryByName() entry = '%v', want = 'nil'", databaseEntry)
	}
}

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
func insertTestScope(t *testing.T, srvEntry *T_db_server, name, dbName, secret string) *T_scan_scope {
	t.Helper()
	scopeEntry, errScopeEntry := CreateScopeEntry(
		srvEntry,
		name,
		dbName,
		42,
		"testuser",
		secret,
		"custom",
		false,
		-1,
		utils.NetworkSizeSplit,
		utils.JsonMap{},
		0,
		defaultScanSettings(),
	)
	if errScopeEntry != nil {
		t.Fatalf("insertTestScope: %v", errScopeEntry)
	}
	t.Cleanup(func() { _ = managerDb.Delete(scopeEntry) })
	return scopeEntry
}

// insertTestViewEntry creates a view entry for the given scope and schedules cleanup via t.Cleanup.
func insertTestViewEntry(t *testing.T, scopeEntry *T_scan_scope, name string) *T_scope_view {
	t.Helper()
	view, errView := CreateViewEntry(scopeEntry, name, "testuser", map[string][]string{}, []string{"hosts_" + name})
	if errView != nil {
		t.Fatalf("insertTestViewEntry: %v", errView)
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

	got, errGot := GetDatabaseEntry(entry.Id)
	if errGot != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGot)
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

	got2, errGot2 := GetDatabaseEntry(entry.Id)
	if errGot2 != nil {
		t.Fatalf("GetDatabaseEntry() after update error = '%v'", errGot2)
	}
	if got2 == nil || got2.Name != "updated-server" {
		t.Errorf("GetDatabaseEntry() name after update = '%v', want = 'updated-server'", got2.Name)
	}
}

// TestGetDatabaseEntry_NotFound verifies GetDatabaseEntry returns nil for an unknown ID.
func TestGetDatabaseEntry_NotFound(t *testing.T) {
	got, errGot := GetDatabaseEntry(999999)
	if errGot != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGot)
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

	got, errGot := GetDatabaseEntry(entry.Id)
	if errGot != nil {
		t.Fatalf("GetDatabaseEntry() after remove error = '%v'", errGot)
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

	got, errGot := GetDatabaseEntryByName("named-server")
	if errGot != nil {
		t.Fatalf("getDatabaseEntryByName() error = '%v'", errGot)
	}
	if got == nil {
		t.Fatal("getDatabaseEntryByName() returned nil")
	}
	if got.Name != "named-server" {
		t.Errorf("getDatabaseEntryByName() name = '%v', want = 'named-server'", got.Name)
	}

	missing, errMissing := GetDatabaseEntryByName("does-not-exist")
	if errMissing != nil {
		t.Fatalf("getDatabaseEntryByName() missing error = '%v'", errMissing)
	}
	if missing != nil {
		t.Errorf("getDatabaseEntryByName() = '%v', want = 'nil' for unknown name", missing)
	}
}

// TestGetScopeEntry_NotFound verifies GetScopeEntry returns nil for an unknown scope ID.
func TestGetScopeEntry_NotFound(t *testing.T) {
	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntry(999999)
	if errScopeEntryLoaded != nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errScopeEntryLoaded)
	}
	if scopeEntryLoaded != nil {
		t.Errorf("GetScopeEntry() = '%v', want = 'nil' for non-existent ID", scopeEntryLoaded)
	}
}

// TestGetScopeEntryBySecret_NotFound verifies GetScopeEntryBySecret returns nil for an unknown secret.
func TestGetScopeEntryBySecret_NotFound(t *testing.T) {
	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntryBySecret("non-existent-secret")
	if errScopeEntryLoaded != nil {
		t.Fatalf("GetScopeEntryBySecret() error = '%v'", errScopeEntryLoaded)
	}
	if scopeEntryLoaded != nil {
		t.Errorf("GetScopeEntryBySecret() = '%v', want = 'nil' for unknown secret", scopeEntryLoaded)
	}
}

// TestGetScopeEntryByName_NotFound verifies GetScopeEntryByName returns nil for an unknown DB name.
func TestGetScopeEntryByName_NotFound(t *testing.T) {
	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntryByName("non-existent-db-name")
	if errScopeEntryLoaded != nil {
		t.Fatalf("GetScopeEntryByName() error = '%v'", errScopeEntryLoaded)
	}
	if scopeEntryLoaded != nil {
		t.Errorf("GetScopeEntryByName() = '%v', want = 'nil' for unknown DB name", scopeEntryLoaded)
	}
}

// TestScopeCRUD verifies create, get, lookup by secret/name, list, ID query, and Save on scopes.
func TestScopeCRUD(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.1", 5432)
	scopeEntry := insertTestScope(t, srv, "My Scope", "db_myscope_001", "secret-abc-001")

	scopeEntryLoaded, errScopeEntryLoaded := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryLoaded != nil {
		t.Fatalf("GetScopeEntry() error = '%v'", errScopeEntryLoaded)
	}
	if scopeEntryLoaded == nil {
		t.Fatal("GetScopeEntry() returned nil")
	}
	if scopeEntryLoaded.Name != "My Scope" {
		t.Errorf("GetScopeEntry() name = '%v', want = 'My Scope'", scopeEntryLoaded.Name)
	}

	scopeEntrySecret, errScopeEntrySecret := GetScopeEntryBySecret("secret-abc-001")
	if errScopeEntrySecret != nil {
		t.Fatalf("GetScopeEntryBySecret() error = '%v'", errScopeEntrySecret)
	}
	if scopeEntrySecret == nil || scopeEntrySecret.Id != scopeEntry.Id {
		t.Errorf("GetScopeEntryBySecret() Id = '%v', want = '%v'", scopeEntrySecret.Id, scopeEntry.Id)
	}

	scopeEntryNamed, errScopeEntryNamed := GetScopeEntryByName("db_myscope_001")
	if errScopeEntryNamed != nil {
		t.Fatalf("GetScopeEntryByName() error = '%v'", errScopeEntryNamed)
	}
	if scopeEntryNamed == nil || scopeEntryNamed.Id != scopeEntry.Id {
		t.Errorf("GetScopeEntryByName() Id = '%v', want = '%v'", scopeEntryNamed.Id, scopeEntry.Id)
	}

	ids, errIds := GetScopeEntryIds()
	if errIds != nil {
		t.Fatalf("GetScopeEntryIds() error = '%v'", errIds)
	}
	found := false
	for _, id := range ids {
		if id == scopeEntry.Id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetScopeEntryIds() did not include scope Id = '%v'", scopeEntry.Id)
	}

	scopeEntriesFiltered, errScopeEntriesFiltered := GetScopeEntriesOf([]uint64{42})
	if errScopeEntriesFiltered != nil {
		t.Fatalf("GetScopeEntriesOf() error = '%v'", errScopeEntriesFiltered)
	}
	if len(scopeEntriesFiltered) == 0 {
		t.Errorf("GetScopeEntriesOf() len = '0', want = '>0' for matching group")
	}

	scopeEntriesEmpty, errScopeEntriesEmpty := GetScopeEntriesOf([]uint64{})
	if errScopeEntriesEmpty != nil {
		t.Fatalf("GetScopeEntriesOf([]) error = '%v'", errScopeEntriesEmpty)
	}
	if len(scopeEntriesEmpty) != 0 {
		t.Errorf("GetScopeEntriesOf([]) len = '%v', want = '0'", len(scopeEntriesEmpty))
	}

	scopeEntriesAll, errScopeEntriesAll := GetScopeEntries()
	if errScopeEntriesAll != nil {
		t.Fatalf("GetScopeEntries() error = '%v'", errScopeEntriesAll)
	}
	if len(scopeEntriesAll) == 0 {
		t.Errorf("GetScopeEntries() len = '0', want = '>0' after insert")
	}

	scopeEntry.Name = "Updated Scope"
	rows, errRows := scopeEntry.Save("name")
	if errRows != nil {
		t.Fatalf("scope.Save() error = '%v'", errRows)
	}
	if rows == 0 {
		t.Errorf("scope.Save() rows = '0', want = '>0'")
	}

	scopeEntryUpdated, errScopeEntryUpdated := GetScopeEntry(scopeEntry.Id)
	if errScopeEntryUpdated != nil {
		t.Fatalf("GetScopeEntry() after Save error = '%v'", errScopeEntryUpdated)
	}
	if scopeEntryUpdated == nil || scopeEntryUpdated.Name != "Updated Scope" {
		t.Errorf("GetScopeEntry() name after Save = '%v', want = 'Updated Scope'", scopeEntryUpdated.Name)
	}
}

// TestScopeSave_NoColumns verifies Save returns an error when called without column arguments.
func TestScopeSave_NoColumns(t *testing.T) {
	scopeEntry := &T_scan_scope{}
	_, errSave := scopeEntry.Save()
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
	scopeEntry := insertTestScope(t, srv, "agent-test", "db_agent_001", "secret-ag-001")

	agentEntries := []T_scan_agent{{
		Name:     "agent-1",
		Host:     "host-1",
		Ip:       "192.0.2.201",
		LastSeen: time.Now(),
		Tasks:    utils.JsonMap{},
	}}
	if errUpdate := UpdateScanAgents(scopeEntry.Id, agentEntries); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() error = '%v'", errUpdate)
	}

	all, errAll := GetAgentEntries()
	if errAll != nil {
		t.Fatalf("GetAgentEntries() error = '%v'", errAll)
	}
	var agentId uint64
	for _, a := range all {
		if a.Name == "agent-1" && a.IdTScanScope == scopeEntry.Id {
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

// TestUpdateScanAgents_VersionsPersistOnConflict verifies that re-registering an agent (same scopeId/name/host)
// updates the four tool-version columns rather than silently dropping them.
func TestUpdateScanAgents_VersionsPersistOnConflict(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.10", 5432)
	scopeEntry := insertTestScope(t, srv, "agent-versions-test", "db_agentver_001", "secret-av-001")

	agentEntriesFirst := []T_scan_agent{{
		Name:          "agent-ver",
		Host:          "host-ver",
		Ip:            "192.0.2.220",
		LastSeen:      time.Now(),
		Tasks:         utils.JsonMap{},
		VersionNmap:   "7.80",
		VersionNpcap:  "1.60",
		VersionSslyze: "5.0.0",
		VersionPython: "3.10.12",
	}}
	if errUpdate := UpdateScanAgents(scopeEntry.Id, agentEntriesFirst); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() first call error = '%v'", errUpdate)
	}

	agentEntriesSecond := []T_scan_agent{{
		Name:          "agent-ver",
		Host:          "host-ver",
		Ip:            "192.0.2.220",
		LastSeen:      time.Now(),
		Tasks:         utils.JsonMap{},
		VersionNmap:   "7.95",
		VersionNpcap:  "1.79",
		VersionSslyze: "6.1.0",
		VersionPython: "3.12.3",
	}}
	if errUpdate := UpdateScanAgents(scopeEntry.Id, agentEntriesSecond); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() second call error = '%v'", errUpdate)
	}

	all, errAll := GetAgentEntries()
	if errAll != nil {
		t.Fatalf("GetAgentEntries() error = '%v'", errAll)
	}
	var agentEntry *T_scan_agent
	for i := range all {
		if all[i].Name == "agent-ver" && all[i].IdTScanScope == scopeEntry.Id {
			agentEntry = &all[i]
			break
		}
	}
	if agentEntry == nil {
		t.Fatal("GetAgentEntries() did not include the agent")
	}
	defer func() { _ = DeleteAgent(agentEntry.Id) }()

	if agentEntry.VersionNmap != "7.95" {
		t.Errorf("UpdateScanAgents() VersionNmap = '%v', want = '%v'", agentEntry.VersionNmap, "7.95")
	}
	if agentEntry.VersionNpcap != "1.79" {
		t.Errorf("UpdateScanAgents() VersionNpcap = '%v', want = '%v'", agentEntry.VersionNpcap, "1.79")
	}
	if agentEntry.VersionSslyze != "6.1.0" {
		t.Errorf("UpdateScanAgents() VersionSslyze = '%v', want = '%v'", agentEntry.VersionSslyze, "6.1.0")
	}
	if agentEntry.VersionPython != "3.12.3" {
		t.Errorf("UpdateScanAgents() VersionPython = '%v', want = '%v'", agentEntry.VersionPython, "3.12.3")
	}
}

// TestScanAgentSave verifies Save updates a column and returns an error when no columns are specified.
func TestScanAgentSave(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.4", 5432)
	scopeEntry := insertTestScope(t, srv, "agent-save-test", "db_agentsave_001", "secret-as-001")

	agentEntries := []T_scan_agent{{
		Name:     "agent-save",
		Host:     "host-save",
		Ip:       "192.0.2.211",
		LastSeen: time.Now(),
		Tasks:    utils.JsonMap{},
	}}
	if errUpdate := UpdateScanAgents(scopeEntry.Id, agentEntries); errUpdate != nil {
		t.Fatalf("UpdateScanAgents() error = '%v'", errUpdate)
	}

	all, errAll := GetAgentEntries()
	if errAll != nil {
		t.Fatalf("GetAgentEntries() error = '%v'", errAll)
	}
	var agentEntry *T_scan_agent
	for i := range all {
		if all[i].Name == "agent-save" && all[i].IdTScanScope == scopeEntry.Id {
			agentEntry = &all[i]
			break
		}
	}
	if agentEntry == nil {
		t.Fatal("GetAgentEntries() did not include inserted agent")
	}
	defer func() { _ = DeleteAgent(agentEntry.Id) }()

	agentEntry.Ip = "192.0.2.212"
	rows, errRows := agentEntry.Save("ip")
	if errRows != nil {
		t.Fatalf("agent.Save() error = '%v'", errRows)
	}
	if rows == 0 {
		t.Errorf("agent.Save() rows = '0', want = '>0'")
	}

	_, errSaveEmpty := agentEntry.Save()
	if errSaveEmpty == nil {
		t.Errorf("agent.Save() with no columns expected error, got = 'nil'")
	}
}

// TestViewEntriesCRUD verifies create, get, list, exists, update, and grant operations on view entries.
func TestViewEntriesCRUD(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.5", 5432)
	scopeEntry := insertTestScope(t, srv, "view-test", "db_view_001", "secret-vw-001")
	view := insertTestViewEntry(t, scopeEntry, "myview")

	got, errGot := GetViewEntry(view.Id)
	if errGot != nil {
		t.Fatalf("GetViewEntry() error = '%v'", errGot)
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

	exists, errExists := ViewExists(scopeEntry.Id, "myview")
	if errExists != nil {
		t.Fatalf("ViewExists() error = '%v'", errExists)
	}
	if !exists {
		t.Errorf("ViewExists() = 'false', want = 'true'")
	}

	absent, errAbsent := ViewExists(scopeEntry.Id, "non-existing-view")
	if errAbsent != nil {
		t.Fatalf("ViewExists(absent) error = '%v'", errAbsent)
	}
	if absent {
		t.Errorf("ViewExists(absent) = 'true', want = 'false'")
	}

	view.Name = "myview-updated"
	rows, errRows := view.Save("name")
	if errRows != nil {
		t.Fatalf("view.Save() error = '%v'", errRows)
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
	scopeEntry := insertTestScope(t, srv, "grant-test", "db_grant_001", "secret-gr-001")
	view := insertTestViewEntry(t, scopeEntry, "grantview")

	expiry := time.Now().Add(24 * time.Hour)
	grant, errGrant := CreateGrantEntry(view, true, "user@domain.tld", "testuser", expiry, "test grant")
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

	viewsGranted, errViewsGranted := GetViewsGranted("user@domain.tld")
	if errViewsGranted != nil {
		t.Fatalf("GetViewsGranted() error = '%v'", errViewsGranted)
	}
	if len(viewsGranted) == 0 {
		t.Errorf("GetViewsGranted() len = '0', want = '>0' after grant")
	}
}

// TestUpdateScopeInstances verifies updating scan instance counts and expected error cases.
func TestUpdateScopeInstances(t *testing.T) {
	srv := insertTestDbServer(t, "192.0.2.7", 5432)
	scopeEntry := insertTestScope(t, srv, "instances-test", "db_instances_001", "secret-in-001")

	errUpdate := UpdateScopeInstances(scopeEntry.Id, map[string]uint32{
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

	errUnknown := UpdateScopeInstances(scopeEntry.Id, map[string]uint32{"unknown-module": 5})
	if errUnknown == nil {
		t.Errorf("updateScopeInstances(unknown-module) expected error, got = 'nil'")
	}

	errBadScope := UpdateScopeInstances(999999, map[string]uint32{"Discovery": 1})
	if errBadScope == nil {
		t.Errorf("updateScopeInstances(nonexistent-scope) expected error, got = 'nil'")
	}
}

// TestCreateDatabaseEntry verifies createDatabaseEntry inserts a new server entry and assigns a non-zero ID.
func TestCreateDatabaseEntry(t *testing.T) {
	entry, errEntry := CreateDatabaseEntry(
		"entry-via-create",
		"postgres",
		"db.domain.tld",
		5432,
		"admin",
		"",
		"",
		"",
	)
	if errEntry != nil {
		t.Fatalf("createDatabaseEntry() error = '%v'", errEntry)
	}
	if entry == nil || entry.Id == 0 {
		t.Fatal("createDatabaseEntry() returned nil or Id = '0'")
	}
	defer func() { _ = RemoveDatabaseEntry(entry.Id) }()

	got, errGot := GetDatabaseEntry(entry.Id)
	if errGot != nil {
		t.Fatalf("GetDatabaseEntry() error = '%v'", errGot)
	}
	if got == nil || got.Name != "entry-via-create" {
		t.Errorf("GetDatabaseEntry() name = '%v', want = 'entry-via-create'", got.Name)
	}
}

// TestClose verifies Close closes an open manager database without error
func TestClose(t *testing.T) {

	// Preserve the shared package handle around this isolated shutdown check
	savedDb := managerDb
	defer func() { managerDb = savedDb }()

	// Open an independent in-memory database without changing the process working directory
	database, errDatabase := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	})
	if errDatabase != nil {
		t.Fatalf("gorm.Open() error = '%v', want = 'nil'", errDatabase)
	}
	managerDb = database

	// Close the isolated database
	if errClose := Close(); errClose != nil {
		t.Errorf("Close() error = '%v', want = 'nil'", errClose)
	}
}

// TestCloseClosesPoolAfterMaintenanceError verifies maintenance failures cannot bypass mandatory cleanup
func TestCloseClosesPoolAfterMaintenanceError(t *testing.T) {
	savedDb := managerDb
	defer func() { managerDb = savedDb }()

	database, errDatabase := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormlog.Silent),
	})
	if errDatabase != nil {
		t.Fatalf("Could not open test database: %v", errDatabase)
	}
	sqlDb, errSqlDb := database.DB()
	if errSqlDb != nil {
		t.Fatalf("Could not retrieve test database pool: %v", errSqlDb)
	}
	managerDb = database
	t.Cleanup(func() { _ = utils.CloseDatabasePool(database) })

	if errQueryOnly := database.Exec("PRAGMA query_only = ON").Error; errQueryOnly != nil {
		t.Fatalf("Could not make test database read only: %v", errQueryOnly)
	}
	errClose := Close()
	if errClose == nil || !strings.Contains(errClose.Error(), "could not vacuum manager database") {
		t.Fatalf("Close() error = '%v', want = 'manager vacuum error'", errClose)
	}
	if errPing := sqlDb.Ping(); errPing == nil {
		t.Fatal("Close() left the underlying database pool open")
	}
	if managerDb != nil {
		t.Fatal("Close() left the global database handle attached")
	}
	if errCloseAgain := Close(); errCloseAgain != nil {
		t.Fatalf("Second Close() error = '%v', want = 'nil'", errCloseAgain)
	}
}
