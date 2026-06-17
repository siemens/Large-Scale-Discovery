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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/siemens/GoScans/nuclei"

	"github.com/microcosm-cc/bluemonday"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// T_db_server contains connection details and credentials to a database hosting one or multiple scan scopes
type T_db_server struct {
	Id         uint64 `gorm:"column:id;primaryKey" json:"id"`                                                 // Id autoincrement
	Name       string `gorm:"column:name;type:text" json:"name"`                                              // Name of the database as a note for administrators
	Dialect    string `gorm:"column:dialect;type:text;uniqueIndex:uix_db_server_multi_column" json:"dialect"` // DB connection details...
	Host       string `gorm:"column:host;type:text;uniqueIndex:uix_db_server_multi_column" json:"host"`       // ...
	HostPublic string `gorm:"column:host_public;type:text" json:"host_public"`                                // Public endpoint for user access (might be different to the internally used one, e.g. load balancer)
	Port       int    `gorm:"column:port;uniqueIndex:uix_db_server_multi_column" json:"port"`                 // ...
	Admin      string `gorm:"column:admin;type:text" json:"admin"`                                            // ...
	Password   string `gorm:"column:password;type:text" json:"-"`                                             // ...
	Args       string `gorm:"column:args;type:text" json:"args"`                                              // Additional connection arguments

	ScanScopes []T_scan_scope `gorm:"foreignKey:IdTDbServer"`
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (dbServer *T_db_server) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	dbServer.Name = b.Sanitize(dbServer.Name)
	tx.Statement.SetColumn("name", dbServer.Name)

	dbServer.Dialect = b.Sanitize(dbServer.Dialect)
	tx.Statement.SetColumn("dialect", dbServer.Dialect)

	dbServer.Host = b.Sanitize(dbServer.Host)
	tx.Statement.SetColumn("host", dbServer.Host)

	dbServer.Admin = b.Sanitize(dbServer.Admin)
	tx.Statement.SetColumn("admin", dbServer.Admin)

	// Password should never be shown anywhere. Therefore, it doesn't need to be encoded.
	// It should be stored plain original, otherwise quotes symbols in passwords would make it invalid.

	dbServer.HostPublic = b.Sanitize(dbServer.HostPublic)
	tx.Statement.SetColumn("host_public", dbServer.HostPublic)

	dbServer.Args = b.Sanitize(dbServer.Args)
	tx.Statement.SetColumn("args", dbServer.Args)

	// Return nil as everything went fine
	return nil
}

// Delete a user
func (dbServer *T_db_server) Delete() error {

	// Delete user from database
	db := managerDb.Delete(dbServer)
	if db.Error != nil {
		return db.Error
	}

	// Return nil as everything went fine
	return nil
}

// T_scan_scope contains available scan scopes and their configuration
type T_scan_scope struct {
	Id          uint64 `gorm:"column:id;primaryKey" json:"id"`                         // Id autoincrement
	IdTDbServer uint64 `gorm:"column:id_t_db_server;type:int;not null;index" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTGroup    uint64 `gorm:"column:id_t_group;type:int;not null;index" json:"-"`     // Group reference this scan scope belongs to. Table "t_groups" is maintained by the web backend and not known to the manager.

	Name            string        `gorm:"column:name;type:text" json:"name"`                                   // Name of the scope selected by the user
	DbName          string        `gorm:"column:db_name;type:text;uniqueIndex" json:"-"`                       // Database name (UID) to connect to on respective DB server. Unique across all DB servers to avoid future conflicts.
	Created         time.Time     `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`             // Timestamp of creation
	CreatedBy       string        `gorm:"column:created_by;type:text;not null" json:"created_by"`              // User who created this scope
	Secret          string        `gorm:"column:secret;type:text;unique" json:"-"`                             // Random none-guessable scope secret used by agents to authenticate/associate. Value may change.
	Enabled         bool          `gorm:"column:enabled;default:true" json:"enabled"`                          // Whether new target should be fed to scan agents for this scan scope
	Type            string        `gorm:"column:type;type:text" json:"type"`                                   // The kind of scope, there might be different ones initialized via different mechanisms. E.g. custom, remote repository,...
	LastSync        time.Time     `gorm:"column:last_sync" json:"last_sync"`                                   // Timestamp when the scan scope targets were set/updated/synchronized the last time
	Size            uint          `gorm:"column:size;type:int;default:0" json:"size"`                          // Amount of IPs currently within this scan scope. Needs to be calculated/updated during population of the actual scan scope's t_discovery table.
	Cycles          bool          `gorm:"column:cycles" json:"cycles"`                                         // Scan in cycles
	CyclesRetention int           `gorm:"column:cycles_retention;type:int;default:-1" json:"cycles_retention"` // Amount of previous scan cycles to keep. Older ones will be cleaned up.
	Attributes      utils.JsonMap `gorm:"column:attributes;type:json;not null" json:"attributes"`              // Scope arguments that can be arbitrary to your deployment environment, e.g., describing how to populate, import, refresh, synchronize scan inputs...

	Cycle        uint          `gorm:"column:cycle;type:int;default:1" json:"cycle"`                        // The current cycle the scan is in. Relevant, if scanning in cycles is enabled
	CycleStarted time.Time     `gorm:"column:cycle_started;default:CURRENT_TIMESTAMP" json:"cycle_started"` // Timestamp of last cycle start
	CycleDone    float64       `gorm:"column:cycle_done;type:float;default:0" json:"cycle_done"`            // Percentage of completed input scan tasks. Is updated in intervals and not a 100% current.
	CycleActive  float64       `gorm:"column:cycle_active;type:float;default:0" json:"cycle_active"`        // Percentage of active input scan tasks. Is updated in intervals and not a 100% current.
	CycleFailed  float64       `gorm:"column:cycle_failed;type:float;default:0" json:"cycle_failed"`        // Percentage of failed input scan tasks. Is updated in intervals and not a 100% current.
	CycleQueue   utils.JsonMap `gorm:"column:cycle_queue;type:json;not null" json:"cycle_queue"`            // Per-module counts of queued scan tasks. Is updated in intervals and not a 100% current.

	DbServer     T_db_server    `gorm:"foreignKey:IdTDbServer" json:"-"` // Database server connection details to connect to
	ScanSettings T_scan_setting `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	ScanAgents   []T_scan_agent `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // Scan agent data (cached data)
	ScopeViews   []T_scope_view `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (scanScope *T_scan_scope) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	scanScope.Name = b.Sanitize(scanScope.Name)
	tx.Statement.SetColumn("name", scanScope.Name)

	scanScope.DbName = b.Sanitize(scanScope.DbName)
	tx.Statement.SetColumn("db_name", scanScope.DbName)

	scanScope.CreatedBy = b.Sanitize(scanScope.CreatedBy)
	tx.Statement.SetColumn("created_by", scanScope.CreatedBy)

	scanScope.Secret = b.Sanitize(scanScope.Secret)
	tx.Statement.SetColumn("secret", scanScope.Secret)

	scanScope.Type = b.Sanitize(scanScope.Type)
	tx.Statement.SetColumn("type", scanScope.Type)

	// Return nil as everything went fine
	return nil
}

// Save updates defined columns of a user entry in the database. It updates defined columns, to the currently
// set values, even if the values are empty ones, such as 0, false or "".
// ATTENTION: Only update required columns to avoid overwriting changes of parallel processes (with data in memory)
func (scanScope *T_scan_scope) Save(columns ...string) (int64, error) {

	// Verify that columns were supplied
	if len(columns) < 1 {
		return 0, fmt.Errorf("no update columns specified")
	}

	// Prepare arguments to pass to GORM. Cannot pass string types, but interface types.
	// GORM requires some strange set of arguments
	var arg0 interface{} = columns[0]
	var args = make([]interface{}, 0, len(columns)-1)
	for _, column := range columns[1:] {
		args = append(args, column)
	}

	// Update user in database
	db := managerDb.
		Select(arg0, args...). // Select defines the columns to be updated
		Save(scanScope)        // Save will also update empty values (false, 0, "")
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

type T_scan_setting struct {
	Id           uint64 `gorm:"column:id;primaryKey" json:"-"`                                 // Id autoincrement
	IdTScanScope uint64 `gorm:"column:id_t_scan_scope;type:int;not null;uniqueIndex" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions

	Ot bool `gorm:"column:ot;default:false" json:"ot"` // Enable OT discovery mode (PROFINET DCP, EtherCAT, LLDP, NDP, mDNS) with OT-optimized Nmap timing. Agent auto-detects scan targets based on local configuration.

	MaxInstancesDiscovery        uint32          `gorm:"column:max_instances_discovery;type:int" json:"max_instances_discovery"`   // Maximum parallel instances of discovery scans per agent.
	MaxInstancesBanner           uint32          `gorm:"column:max_instances_banner;type:int" json:"max_instances_banner"`         // Maximum parallel instances of banner scans per agent.
	MaxInstancesNfs              uint32          `gorm:"column:max_instances_nfs;type:int" json:"max_instances_nfs"`               // Maximum parallel instances of nfs scans per agent.
	MaxInstancesNuclei           uint32          `gorm:"column:max_instances_nuclei;type:int" json:"max_instances_nuclei"`         // Maximum parallel instances of nuclei scans per agent.
	MaxInstancesSmb              uint32          `gorm:"column:max_instances_smb;type:int" json:"max_instances_smb"`               // Maximum parallel instances of smb scans per agent.
	MaxInstancesSsh              uint32          `gorm:"column:max_instances_ssh;type:int" json:"max_instances_ssh"`               // Maximum parallel instances of ssh scans per agent.
	MaxInstancesSsl              uint32          `gorm:"column:max_instances_ssl;type:int" json:"max_instances_ssl"`               // Maximum parallel instances of ssl scans per agent.
	MaxInstancesWebcrawler       uint32          `gorm:"column:max_instances_webcrawler;type:int" json:"max_instances_webcrawler"` // Maximum parallel instances of webcrawler scans per agent.
	MaxInstancesWebenum          uint32          `gorm:"column:max_instances_webenum;type:int" json:"max_instances_webenum"`       // Maximum parallel instances of webenum scans per agent.
	SensitivePorts               string          `gorm:"column:sensitive_ports;type:text" json:"-"`                                // Comma separated list of sensitive ports that shall not be scanned with submodules
	SensitivePortsSlice          []int           `gorm:"-" json:"sensitive_ports"`                                                 //
	NetworkTimeoutSeconds        int             `gorm:"column:network_timeout_seconds;type:int" json:"network_timeout_seconds"`   //
	HttpUserAgent                string          `gorm:"column:http_user_agent;type:text" json:"http_user_agent"`                  //
	DiscoveryTimespans           string          `gorm:"column:discovery_timespans;type:text" json:"-"`
	DiscoveryTimespansSlice      utils.Timespans `gorm:"-" json:"discovery_timespans"`
	DiscoveryNmapArgs            string          `gorm:"column:discovery_nmap_args;type:text" json:"discovery_nmap_args"`
	DiscoveryNmapArgsPrescan     string          `gorm:"column:discovery_nmap_args_prescan;type:text" json:"discovery_nmap_args_prescan"` // A smaller scan executed before the main scan to at least retrieve some scan results, before a potential IDS kicks in
	DiscoveryNmapArgsOt          string          `gorm:"-" json:"discovery_nmap_args_ot"`                                                 // A safer scan executed in OT discovery scans. Field is just required as a JSON field to load default settings from the manager.conf. Not required in the managerdb, because it will be copied into the normal nmap args field.
	DiscoveryExcludeDomains      string          `gorm:"column:discovery_exclude_domains;type:text" json:"discovery_exclude_domains"`
	NfsScanTimeoutMinutes        int             `gorm:"column:nfs_scan_timeout_minutes;type:int" json:"nfs_scan_timeout_minutes"`
	NfsDepth                     int             `gorm:"column:nfs_depth;type:int" json:"nfs_depth"`
	NfsThreads                   int             `gorm:"column:nfs_threads;type:int" json:"nfs_threads"`
	NfsExcludeShares             string          `gorm:"column:nfs_exclude_shares;type:text" json:"nfs_exclude_shares"`
	NfsExcludeFolders            string          `gorm:"column:nfs_exclude_folders;type:text" json:"nfs_exclude_folders"`
	NfsExcludeExtensions         string          `gorm:"column:nfs_exclude_extensions;type:text" json:"nfs_exclude_extensions"`
	NfsExcludeFileSizeBelow      int             `gorm:"column:nfs_exclude_file_size_below;type:int" json:"nfs_exclude_file_size_below"`
	NfsExcludeLastModifiedBelow  time.Time       `gorm:"column:nfs_exclude_last_modified_below;type:datetime" json:"nfs_exclude_last_modified_below"`
	NfsAccessibleOnly            bool            `gorm:"column:nfs_accessible_only;type:bool" json:"nfs_accessible_only"`
	NucleiScanTimeoutMinutes     int             `gorm:"column:nuclei_scan_timeout_minutes;type:int" json:"nuclei_scan_timeout_minutes"`
	NucleiIncludeSeverities      string          `gorm:"column:nuclei_include_severities;type:text" json:"nuclei_include_severities"`
	NucleiExcludeSeverities      string          `gorm:"column:nuclei_exclude_severities;type:text" json:"nuclei_exclude_severities"`
	NucleiIncludeTags            string          `gorm:"column:nuclei_include_tags;type:text" json:"nuclei_include_tags"`
	NucleiExcludeTags            string          `gorm:"column:nuclei_exclude_tags;type:text" json:"nuclei_exclude_tags"`
	NucleiIncludeIds             string          `gorm:"column:nuclei_include_ids;type:text" json:"nuclei_include_ids"`
	NucleiExcludeIds             string          `gorm:"column:nuclei_exclude_ids;type:text" json:"nuclei_exclude_ids"`
	NucleiIncludeProtocols       string          `gorm:"column:nuclei_include_protocols;type:text" json:"nuclei_include_protocols"`
	NucleiExcludeProtocols       string          `gorm:"column:nuclei_exclude_protocols;type:text" json:"nuclei_exclude_protocols"`
	SmbScanTimeoutMinutes        int             `gorm:"column:smb_scan_timeout_minutes;type:int" json:"smb_scan_timeout_minutes"`
	SmbDepth                     int             `gorm:"column:smb_depth;type:int" json:"smb_depth"`
	SmbThreads                   int             `gorm:"column:smb_threads;type:int" json:"smb_threads"`
	SmbForcedShares              string          `gorm:"column:smb_forced_shares;type:text" json:"smb_forced_shares"`
	SmbExcludeShares             string          `gorm:"column:smb_exclude_shares;type:text" json:"smb_exclude_shares"`
	SmbExcludeFolders            string          `gorm:"column:smb_exclude_folders;type:text" json:"smb_exclude_folders"`
	SmbExcludeExtensions         string          `gorm:"column:smb_exclude_extensions;type:text" json:"smb_exclude_extensions"`
	SmbExcludeFileSizeBelow      int             `gorm:"column:smb_exclude_file_size_below;type:int" json:"smb_exclude_file_size_below"`
	SmbExcludeLastModifiedBelow  time.Time       `gorm:"column:smb_exclude_last_modified_below;type:datetime" json:"smb_exclude_last_modified_below"`
	SmbAccessibleOnly            bool            `gorm:"column:smb_accessible_only;type:bool" json:"smb_accessible_only"`
	SslScanTimeoutMinutes        int             `gorm:"column:ssl_scan_timeout_minutes;type:int" json:"ssl_scan_timeout_minutes"`
	SshScanTimeoutMinutes        int             `gorm:"column:ssh_scan_timeout_minutes;type:int" json:"ssh_scan_timeout_minutes"`
	WebcrawlerScanTimeoutMinutes int             `gorm:"column:webcrawler_scan_timeout_minutes;type:int" json:"webcrawler_scan_timeout_minutes"`
	WebcrawlerDepth              int             `gorm:"column:webcrawler_depth;type:int" json:"webcrawler_depth"`
	WebcrawlerMaxThreads         int             `gorm:"column:webcrawler_max_threads;type:int" json:"webcrawler_max_threads"`
	WebcrawlerFollowQueryStrings bool            `gorm:"column:webcrawler_follow_query_strings;type:bool" json:"webcrawler_follow_query_strings"`
	WebcrawlerAlwaysStoreRoot    bool            `gorm:"column:webcrawler_always_store_root;type:bool" json:"webcrawler_always_store_root"`
	WebcrawlerFollowTypes        string          `gorm:"column:webcrawler_follow_types;type:text" json:"webcrawler_follow_types"`
	WebenumScanTimeoutMinutes    int             `gorm:"column:webenum_scan_timeout_minutes;type:int" json:"webenum_scan_timeout_minutes"`
	WebenumProbeRobots           bool            `gorm:"column:webenum_probe_robots;type:bool" json:"webenum_probe_robots"`

	ScanScope *T_scan_scope `gorm:"foreignKey:IdTScanScope" json:"-"` // Has to be a pointer in order to prevent an invalid recursion. Can be nil if the id has been set
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (scanSettings *T_scan_setting) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	scanSettings.SensitivePorts = b.Sanitize(scanSettings.SensitivePorts)
	scanSettings.SensitivePorts = utils.SanitizeCommaSeparated(scanSettings.SensitivePorts)
	tx.Statement.SetColumn("sensitive_ports", scanSettings.SensitivePorts)

	scanSettings.HttpUserAgent = b.Sanitize(scanSettings.HttpUserAgent)
	tx.Statement.SetColumn("http_user_agent", scanSettings.HttpUserAgent)

	for strings.Contains(scanSettings.DiscoveryNmapArgs, "  ") { // Duplicate spaces in Nmap args may cause scan errors
		scanSettings.DiscoveryNmapArgs = strings.ReplaceAll(scanSettings.DiscoveryNmapArgs, "  ", " ")
	}
	scanSettings.DiscoveryNmapArgs = strings.TrimSpace(scanSettings.DiscoveryNmapArgs)
	tx.Statement.SetColumn("discovery_nmap_args", scanSettings.DiscoveryNmapArgs)

	for strings.Contains(scanSettings.DiscoveryNmapArgsPrescan, "  ") { // Duplicate spaces in Nmap args may cause scan errors
		scanSettings.DiscoveryNmapArgsPrescan = strings.ReplaceAll(scanSettings.DiscoveryNmapArgsPrescan, "  ", " ")
	}
	scanSettings.DiscoveryNmapArgsPrescan = strings.TrimSpace(scanSettings.DiscoveryNmapArgsPrescan)
	tx.Statement.SetColumn("discovery_nmap_args_prescan", scanSettings.DiscoveryNmapArgsPrescan)

	scanSettings.DiscoveryExcludeDomains = b.Sanitize(scanSettings.DiscoveryExcludeDomains)
	scanSettings.DiscoveryExcludeDomains = utils.SanitizeCommaSeparated(scanSettings.DiscoveryExcludeDomains)
	tx.Statement.SetColumn("discovery_exclude_domains", scanSettings.DiscoveryExcludeDomains)

	scanSettings.SmbForcedShares = b.Sanitize(scanSettings.SmbForcedShares)
	scanSettings.SmbForcedShares = utils.SanitizeCommaSeparated(scanSettings.SmbForcedShares)
	tx.Statement.SetColumn("smb_forced_shares", scanSettings.SmbForcedShares)

	scanSettings.SmbExcludeShares = b.Sanitize(scanSettings.SmbExcludeShares)
	scanSettings.SmbExcludeShares = utils.SanitizeCommaSeparated(scanSettings.SmbExcludeShares)
	tx.Statement.SetColumn("smb_exclude_shares", scanSettings.SmbExcludeShares)

	scanSettings.SmbExcludeFolders = b.Sanitize(scanSettings.SmbExcludeFolders)
	scanSettings.SmbExcludeFolders = utils.SanitizeCommaSeparated(scanSettings.SmbExcludeFolders)
	tx.Statement.SetColumn("smb_exclude_folders", scanSettings.SmbExcludeFolders)

	scanSettings.SmbExcludeExtensions = b.Sanitize(scanSettings.SmbExcludeExtensions)
	scanSettings.SmbExcludeExtensions = utils.SanitizeCommaSeparated(scanSettings.SmbExcludeExtensions)
	tx.Statement.SetColumn("smb_exclude_extensions", scanSettings.SmbExcludeExtensions)

	scanSettings.NfsExcludeShares = b.Sanitize(scanSettings.NfsExcludeShares)
	scanSettings.NfsExcludeShares = utils.SanitizeCommaSeparated(scanSettings.NfsExcludeShares)
	tx.Statement.SetColumn("nfs_exclude_shares", scanSettings.NfsExcludeShares)

	scanSettings.NfsExcludeFolders = b.Sanitize(scanSettings.NfsExcludeFolders)
	scanSettings.NfsExcludeFolders = utils.SanitizeCommaSeparated(scanSettings.NfsExcludeFolders)
	tx.Statement.SetColumn("nfs_exclude_folders", scanSettings.NfsExcludeFolders)

	scanSettings.NfsExcludeExtensions = b.Sanitize(scanSettings.NfsExcludeExtensions)
	scanSettings.NfsExcludeExtensions = utils.SanitizeCommaSeparated(scanSettings.NfsExcludeExtensions)
	tx.Statement.SetColumn("nfs_exclude_extensions", scanSettings.NfsExcludeExtensions)

	scanSettings.NucleiIncludeSeverities = b.Sanitize(scanSettings.NucleiIncludeSeverities)
	scanSettings.NucleiIncludeSeverities = utils.SanitizeCommaSeparated(scanSettings.NucleiIncludeSeverities)
	tx.Statement.SetColumn("nuclei_include_severities", scanSettings.NucleiIncludeSeverities)

	scanSettings.NucleiExcludeSeverities = b.Sanitize(scanSettings.NucleiExcludeSeverities)
	scanSettings.NucleiExcludeSeverities = utils.SanitizeCommaSeparated(scanSettings.NucleiExcludeSeverities)
	tx.Statement.SetColumn("nuclei_exclude_severities", scanSettings.NucleiExcludeSeverities)

	scanSettings.NucleiIncludeTags = b.Sanitize(scanSettings.NucleiIncludeTags)
	scanSettings.NucleiIncludeTags = utils.SanitizeCommaSeparated(scanSettings.NucleiIncludeTags)
	tx.Statement.SetColumn("nuclei_include_tags", scanSettings.NucleiIncludeTags)

	scanSettings.NucleiExcludeTags = b.Sanitize(scanSettings.NucleiExcludeTags)
	scanSettings.NucleiExcludeTags = utils.SanitizeCommaSeparated(scanSettings.NucleiExcludeTags)
	tx.Statement.SetColumn("nuclei_exclude_tags", scanSettings.NucleiExcludeTags)

	scanSettings.NucleiIncludeIds = b.Sanitize(scanSettings.NucleiIncludeIds)
	scanSettings.NucleiIncludeIds = utils.SanitizeCommaSeparated(scanSettings.NucleiIncludeIds)
	tx.Statement.SetColumn("nuclei_include_ids", scanSettings.NucleiIncludeIds)

	scanSettings.NucleiExcludeIds = b.Sanitize(scanSettings.NucleiExcludeIds)
	scanSettings.NucleiExcludeIds = utils.SanitizeCommaSeparated(scanSettings.NucleiExcludeIds)
	tx.Statement.SetColumn("nuclei_exclude_ids", scanSettings.NucleiExcludeIds)

	scanSettings.NucleiIncludeProtocols = b.Sanitize(scanSettings.NucleiIncludeProtocols)
	scanSettings.NucleiIncludeProtocols = utils.SanitizeCommaSeparated(scanSettings.NucleiIncludeProtocols)
	tx.Statement.SetColumn("nuclei_include_protocols", scanSettings.NucleiIncludeProtocols)

	scanSettings.NucleiExcludeProtocols = b.Sanitize(scanSettings.NucleiExcludeProtocols)
	scanSettings.NucleiExcludeProtocols = utils.SanitizeCommaSeparated(scanSettings.NucleiExcludeProtocols)
	tx.Statement.SetColumn("nuclei_exclude_protocols", scanSettings.NucleiExcludeProtocols)

	scanSettings.WebcrawlerFollowTypes = b.Sanitize(scanSettings.WebcrawlerFollowTypes)
	scanSettings.WebcrawlerFollowTypes = utils.SanitizeCommaSeparated(scanSettings.WebcrawlerFollowTypes)
	tx.Statement.SetColumn("webcrawler_follow_types", scanSettings.WebcrawlerFollowTypes)

	// Return nil as everything went fine
	return nil
}

// AfterFind updates yet empty struct fields with the derived values. Some setting values are not stored in
// the database but derived from other fields in the database.
func (scanSettings *T_scan_setting) AfterFind(tx *gorm.DB) (err error) {

	// Parse list of sensitive ports
	if len(scanSettings.SensitivePorts) > 0 {
		for _, port := range utils.SanitizeToSlice(scanSettings.SensitivePorts, ",") {
			portInt, errPort := strconv.ParseInt(port, 10, 64)
			if errPort != nil {
				return fmt.Errorf("invalid port '%s'", port)
			}
			if portInt < 0 || portInt > 65535 {
				return fmt.Errorf("invalid port '%s'", port)
			}
			scanSettings.SensitivePortsSlice = append(scanSettings.SensitivePortsSlice, int(portInt))
		}
	}

	// Parse list of timespans
	if len(scanSettings.DiscoveryTimespans) > 0 {
		_ = json.Unmarshal([]byte(scanSettings.DiscoveryTimespans), &scanSettings.DiscoveryTimespansSlice)
	}

	// Return as everything went fine
	return
}

func (scanSettings *T_scan_setting) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw JSON data
	type aux T_scan_setting
	var raw aux

	// Unmarshal serialized JSON into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Do other input validation
	if raw.NetworkTimeoutSeconds <= 0 {
		return fmt.Errorf("invalid network timeout")
	}
	if raw.NfsScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid NFS scan timeout")
	}
	if raw.NfsThreads <= 0 {
		return fmt.Errorf("invalid NFS threads")
	}
	if raw.NucleiScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid Nuclei scan timeout")
	}
	if raw.SmbScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid SMB scan timeout")
	}
	if raw.SmbThreads <= 0 {
		return fmt.Errorf("invalid SMB threads")
	}
	if raw.SslScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid SSL scan timeout")
	}
	if raw.SshScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid SSH scan timeout")
	}
	if raw.WebcrawlerScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid webcrawler scan timeout")
	}
	if raw.WebcrawlerMaxThreads <= 0 {
		return fmt.Errorf("invalid webcrawler max threads")
	}
	if raw.WebenumScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid webenum scan timeout")
	}

	// Copy loaded JSON values to actual scan settings
	*scanSettings = T_scan_setting(raw)

	// Check if port values can be parsed
	if len(scanSettings.SensitivePortsSlice) > 0 {
		for _, port := range scanSettings.SensitivePortsSlice {
			if port < 0 || port > 65535 {
				return fmt.Errorf("invalid port '%d'", port)
			}
		}
		scanSettings.SensitivePorts = utils.SanitizeCommaSeparated(utils.JoinInt(scanSettings.SensitivePortsSlice, ","))
	}

	// JSON unmarshalling is called when settings are read from the config or from a user HTTP request, use this
	// opportunity to unique and sort values, so that they are sanitized before storing
	scanSettings.SensitivePortsSlice = utils.UniqueInts(scanSettings.SensitivePortsSlice)
	sort.Ints(scanSettings.SensitivePortsSlice)

	// Check if timespan values can be parsed
	if len(scanSettings.DiscoveryTimespansSlice) > 0 {
		for _, timespan := range scanSettings.DiscoveryTimespansSlice {
			startDay, errStartDay := strconv.ParseInt(timespan.StartDay, 10, 64)
			endDay, errEndDay := strconv.ParseInt(timespan.EndDay, 10, 64)
			if errStartDay != nil || startDay < 0 || startDay > 6 {
				return fmt.Errorf("invalid start day in timespan")
			}
			if errEndDay != nil || endDay < 0 || endDay > 6 {
				return fmt.Errorf("invalid end day in timespan")
			}
			_, errStartTime := time.Parse(utils.TimeFormat, timespan.StartTime)
			if errStartTime != nil {
				return fmt.Errorf("invalid start time in timespan")
			}
			_, errEndTime := time.Parse(utils.TimeFormat, timespan.EndTime)
			if errEndTime != nil {
				return fmt.Errorf("invalid end time in timespan")
			}

		}
		scanSettings.DiscoveryTimespans = scanSettings.DiscoveryTimespansSlice.String()
	}

	// Return nil as everything is valid
	return nil
}

// SaveAll updates *all* columns of a user entry in the database. It overwrites existing values with the ones passed in
func (scanSettings *T_scan_setting) SaveAll(values *T_scan_setting) (int64, error) {

	// Inject entry ID to update
	values.Id = scanSettings.Id
	values.IdTScanScope = scanSettings.IdTScanScope

	// Update user in database
	db := managerDb.
		Save(values) // Save will update all attributes, including empty values (false, 0, "")
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

// MaxInstances extracts the set amount of maximum instances from a scan scope for a given scan module label
func (scanSettings *T_scan_setting) MaxInstances(label string) (int, error) {

	// Get instances for module
	switch label {
	case discovery.Label:
		return int(scanSettings.MaxInstancesDiscovery), nil
	case banner.Label:
		return int(scanSettings.MaxInstancesBanner), nil
	case nfs.Label:
		return int(scanSettings.MaxInstancesNfs), nil
	case nuclei.Label:
		return int(scanSettings.MaxInstancesNuclei), nil
	case smb.Label:
		return int(scanSettings.MaxInstancesSmb), nil
	case ssh.Label:
		return int(scanSettings.MaxInstancesSsh), nil
	case ssl.Label:
		return int(scanSettings.MaxInstancesSsl), nil
	case webcrawler.Label:
		return int(scanSettings.MaxInstancesWebcrawler), nil
	case webenum.Label:
		return int(scanSettings.MaxInstancesWebenum), nil
	}

	// Return reference to queried targets
	return -1, fmt.Errorf("unknown module '%s'", label)
}

type T_scan_agent struct {
	Id           uint64 `gorm:"column:id;primary_key" json:"id"`                                                    // Id autoincrement
	IdTScanScope uint64 `gorm:"column:id_t_scan_scope;type:int;not null;uniqueIndex:idx_agent_identifier" json:"-"` // Unique index (across columns) to make sure the same view name only exists once per scan scope!

	Name string `gorm:"column:name;type:text;not null;uniqueIndex:idx_agent_identifier" json:"name"`
	Host string `gorm:"column:host;type:text;not null;uniqueIndex:idx_agent_identifier" json:"host"`
	Ip   string `gorm:"column:ip;type:text;not null" json:"ip"`

	BuildCommit    string `gorm:"column:build_commit;type:text;" json:"build_commit"`
	BuildTimestamp string `gorm:"column:build_timestamp;type:text;" json:"build_timestamp"`
	ApiVersion     string `gorm:"column:api_version;type:text;" json:"api_version"`

	Shared   bool          `gorm:"column:shared;type:bool;default:false" json:"shared"`
	Limits   bool          `gorm:"column:limits;type:bool;default:false" json:"limits"`
	LastSeen time.Time     `gorm:"column:last_seen;default:CURRENT_TIMESTAMP" json:"last_seen"`
	Tasks    utils.JsonMap `gorm:"column:tasks;default:'{}'" json:"tasks"`

	Platform        string  `gorm:"column:platform;type:text;" json:"platform"`
	PlatformFamily  string  `gorm:"column:platform_family;type:text;" json:"platform_family"`
	PlatformVersion string  `gorm:"column:platform_version;type:text;" json:"platform_version"`
	CpuCores        int     `gorm:"column:cpu_cores;type:int;default:0" json:"cpu_cores"`
	CpuMhz          float64 `gorm:"column:cpu_mhz;type:float;default:0" json:"cpu_mhz"`
	CpuRate         float64 `gorm:"column:cpu_rate;type:float;default:0" json:"cpu_rate"` // Usage in %
	MemoryBytes     uint64  `gorm:"column:memory_bytes;type:bigint;default:0" json:"memory_bytes"`
	MemoryRate      float64 `gorm:"column:memory_rate;type:float;default:0" json:"memory_rate"` // Usage in %

	VersionNmap   string `gorm:"column:version_nmap;type:text;" json:"version_nmap"`
	VersionNpcap  string `gorm:"column:version_npcap;type:text;" json:"version_npcap"`
	VersionSslyze string `gorm:"column:version_sslyze;type:text;" json:"version_sslyze"`

	ScanScope T_scan_scope `gorm:"foreignKey:IdTScanScope" json:"-"` // Can be empty if the ID is set in turn
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (scanAgent *T_scan_agent) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize value
	scanAgent.Name = b.Sanitize(scanAgent.Name)
	tx.Statement.SetColumn("name", scanAgent.Name)

	scanAgent.Host = b.Sanitize(scanAgent.Host)
	tx.Statement.SetColumn("host", scanAgent.Host)

	scanAgent.Ip = b.Sanitize(scanAgent.Ip)
	tx.Statement.SetColumn("ip", scanAgent.Ip)

	scanAgent.BuildCommit = b.Sanitize(scanAgent.BuildCommit)
	tx.Statement.SetColumn("build_commit", scanAgent.BuildCommit)

	scanAgent.BuildTimestamp = b.Sanitize(scanAgent.BuildTimestamp)
	tx.Statement.SetColumn("build_timestamp", scanAgent.BuildTimestamp)

	scanAgent.ApiVersion = b.Sanitize(scanAgent.ApiVersion)
	tx.Statement.SetColumn("api_version", scanAgent.ApiVersion)

	scanAgent.Platform = b.Sanitize(scanAgent.Platform)
	tx.Statement.SetColumn("platform", scanAgent.Platform)

	scanAgent.PlatformFamily = b.Sanitize(scanAgent.PlatformFamily)
	tx.Statement.SetColumn("platform_family", scanAgent.PlatformFamily)

	scanAgent.PlatformVersion = b.Sanitize(scanAgent.PlatformVersion)
	tx.Statement.SetColumn("platform_version", scanAgent.PlatformVersion)

	scanAgent.VersionNmap = b.Sanitize(scanAgent.VersionNmap)
	tx.Statement.SetColumn("version_nmap", scanAgent.VersionNmap)

	scanAgent.VersionNpcap = b.Sanitize(scanAgent.VersionNpcap)
	tx.Statement.SetColumn("version_npcap", scanAgent.VersionNpcap)

	scanAgent.VersionSslyze = b.Sanitize(scanAgent.VersionSslyze)
	tx.Statement.SetColumn("version_sslyze", scanAgent.VersionSslyze)

	// Return nil as everything went fine
	return nil
}

// Save updates defined columns of a user entry in the database. It updates defined columns, to the currently
// set values, even if the values are empty ones, such as 0, false or "".
// ATTENTION: Only update required columns to avoid overwriting changes of parallel processes (with data in memory)
func (scanAgent *T_scan_agent) Save(columns ...string) (int64, error) {

	// Verify that columns were supplied
	if len(columns) < 1 {
		return 0, fmt.Errorf("no update columns specified")
	}

	// Prepare arguments to pass to GORM. Cannot pass string types, but interface types.
	// GORM requires some strange set of arguments
	var arg0 interface{} = columns[0]
	var args = make([]interface{}, 0, len(columns)-1)
	for _, column := range columns[1:] {
		args = append(args, column)
	}

	// Update user in database
	db := managerDb.
		Select(arg0, args...). // Select defines the columns to be updated
		Save(scanAgent)        // Save will also update empty values (false, 0, "")
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

// Delete a user
func (scanAgent *T_scan_agent) Delete() error {

	// Delete user from database
	db := managerDb.Delete(scanAgent)
	if db.Error != nil {
		return db.Error
	}

	// Return nil as everything went fine
	return nil
}

type T_scope_view struct {
	// Unique index on IdTScanScope and Name to make sure the same view name only exists once per scan scope!
	// Otherwise, they couldn't be created on the database!
	Id           uint64 `gorm:"column:id;primaryKey" json:"id"`                                               // Id autoincrement
	IdTScanScope uint64 `gorm:"column:id_t_scan_scope;type:int;not null;uniqueIndex:idx_scope_name" json:"-"` // Unique index (across columns) to make sure the same view name only exists once per scan scope!

	Name      string        `gorm:"column:name;type:text;not null;uniqueIndex:idx_scope_name" json:"name"` // Name of the view to show to the user
	Created   time.Time     `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`               // Timestamp of creation
	CreatedBy string        `gorm:"column:created_by;type:text;not null" json:"created_by"`                // User who created this view
	Filters   utils.JsonMap `gorm:"column:filters;type:json;default:'{}'" json:"filters"`                  // Applied filters restricting view on original data
	ViewNames string        `gorm:"column:view_names;type:text" json:"view_names"`                         // Comma separated list of view table names as created in the scope db

	ScanScope T_scan_scope   `gorm:"foreignKey:IdTScanScope" json:"-"`
	Grants    []T_view_grant `gorm:"foreignKey:IdTScopeView;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (scopeView *T_scope_view) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	scopeView.Name = b.Sanitize(scopeView.Name)
	tx.Statement.SetColumn("name", scopeView.Name)

	scopeView.CreatedBy = b.Sanitize(scopeView.CreatedBy)
	tx.Statement.SetColumn("created_by", scopeView.CreatedBy)

	scopeView.ViewNames = b.Sanitize(scopeView.ViewNames)
	tx.Statement.SetColumn("view_names", scopeView.ViewNames)

	// Return nil as everything went fine
	return nil
}

// Save updates defined columns of a user entry in the database. It updates defined columns, to the currently
// set values, even if the values are empty ones, such as 0, false or "".
// ATTENTION: Only update required columns to avoid overwriting changes of parallel processes (with data in memory)
func (scopeView *T_scope_view) Save(columns ...string) (int64, error) {

	// Verify that columns were supplied
	if len(columns) < 1 {
		return 0, fmt.Errorf("no update columns specified")
	}

	// Prepare arguments to pass to GORM. Cannot pass string types, but interface types.
	// GORM requires some strange set of arguments
	var arg0 interface{} = columns[0]
	var args = make([]interface{}, 0, len(columns)-1)
	for _, column := range columns[1:] {
		args = append(args, column)
	}

	// Update user in database
	db := managerDb.
		Select(arg0, args...). // Select defines the columns to be updated
		Save(scopeView)        // Save will also update empty values (false, 0, "")
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

type T_view_grant struct {
	// A view grant is representing an access right to a certain scope view. It may represent a user specific
	// access right, or a generic not user bound access token access right.
	Id           uint64 `gorm:"column:id;primaryKey;uniqueIndex" json:"-"`               // Id autoincrement
	IdTScopeView uint64 `gorm:"column:id_t_scope_view;type:int;not null;index" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions

	IsUser      bool      `gorm:"column:is_user;type:bool" json:"is_user"`                 // Flag indicating whether this entry is for a dedicated user, rather than a not user bound access token
	Username    string    `gorm:"column:username;type:text" json:"username"`               // The database user name this grant references. May be an e-mail address to reference certain user or random string for access tokens
	Created     time.Time `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"` // Timestamp of grant creation
	CreatedBy   string    `gorm:"column:created_by;type:text;not null" json:"created_by"`  // User who created this grant
	Expiry      time.Time `gorm:"column:expiry;not null" json:"expiry"`                    // Timestamp when this access grant is scheduled to expire, should equal the values set on the database servers
	Description string    `gorm:"column:description;type:text" json:"description"`         // Set by the creator, only necessary for access tokens

	ScopeView T_scope_view `gorm:"foreignKey:IdTScopeView" json:"-"`
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (grantEntry *T_view_grant) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	grantEntry.Username = b.Sanitize(grantEntry.Username)
	tx.Statement.SetColumn("username", grantEntry.Username)

	grantEntry.CreatedBy = b.Sanitize(grantEntry.CreatedBy)
	tx.Statement.SetColumn("created_by", grantEntry.CreatedBy)

	grantEntry.Description = b.Sanitize(grantEntry.Description)
	tx.Statement.SetColumn("description", grantEntry.Description)

	// Return nil as everything went fine
	return nil
}
