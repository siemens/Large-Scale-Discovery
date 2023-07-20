/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"encoding/json"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"gorm.io/gorm"
	"large-scale-discovery/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

// T_db_server contains connection details and credentials to a database hosting one or multiple scan scopes
type T_db_server struct {
	Id         uint64 `gorm:"column:id;primaryKey"`         // Id autoincrement
	Name       string `gorm:"column:name;type:text"`        // Name of the database as a note for administrators
	Dialect    string `gorm:"column:dialect;type:text"`     // DB connection details...
	Host       string `gorm:"column:host;type:text"`        // ...
	Port       int    `gorm:"column:port"`                  // ...
	Admin      string `gorm:"column:admin;type:text"`       // ...
	Password   string `gorm:"column:password;type:text"`    // ...
	HostPublic string `gorm:"column:host_public;type:text"` // Public endpoint for user access (might be different to the internally used one, e.g. load balancer)
	Args       string `gorm:"column:args;type:text"`        // Additional connection arguments

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

	dbServer.Password = b.Sanitize(dbServer.Password)
	tx.Statement.SetColumn("password", dbServer.Password)

	dbServer.HostPublic = b.Sanitize(dbServer.HostPublic)
	tx.Statement.SetColumn("host_public", dbServer.HostPublic)

	dbServer.Args = b.Sanitize(dbServer.Args)
	tx.Statement.SetColumn("args", dbServer.Args)

	// Return nil as everything went fine
	return nil
}

// T_scan_scope contains available scan scopes and their configuration
type T_scan_scope struct {
	Id          uint64 `gorm:"column:id;primaryKey" json:"id"`                         // Id autoincrement
	IdTDbServer uint64 `gorm:"column:id_t_db_server;type:int;not null;index" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTGroup    uint64 `gorm:"column:id_t_group;type:int;not null;index" json:"-"`     // Group reference this scan scope belongs to. Table "t_groups" is maintained by the web backend and not known to the manager.

	Name            string        `gorm:"column:name;type:text" json:"name"`                                   // Name of the scope selected by the user
	DbName          string        `gorm:"column:db_name;type:text" json:"-"`                                   // Database name (UID) to connect to on respective DB server
	Created         time.Time     `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`             // Timestamp of creation
	CreatedBy       string        `gorm:"column:created_by;type:text;not null" json:"created_by"`              // User who created this scope
	Secret          string        `gorm:"column:secret;type:text;unique" json:"-"`                             // Random none-guessable scope secret used by agents to authenticate/associate. Value may change.
	Enabled         bool          `gorm:"column:enabled;default:true" json:"enabled"`                          // Whether new target should be fed to scan agents for this scan scope
	Type            string        `gorm:"column:type;type:string" json:"type"`                                 // The kind of scope, there might be different ones initialized via different mechanisms. E.g. custom, remote repository,...
	LastSync        time.Time     `gorm:"column:last_sync" json:"last_sync"`                                   // Timestamp when the scan scope targets were set/updated/synchronized the last time
	Size            uint          `gorm:"column:size;type:int;default:0" json:"size"`                          // Amount of IPs currently within this scan scope. Needs to be calculated/updated during population of the actual scan scope's t_discovery table.
	Cycles          bool          `gorm:"column:cycles" json:"cycles"`                                         // Scan in cycles
	CyclesRetention int           `gorm:"column:cycles_retention;type:int;default:-1" json:"cycles_retention"` // Amount of previous scan cycles to keep. Older ones will be cleaned up.
	Attributes      utils.JsonMap `gorm:"column:attributes;type:json;not null" json:"attributes"`              // Scope arguments that can be arbitrary to your deployment environment, e.g., describing how to populate, import, refresh, synchronize scan inputs...

	Cycle        uint      `gorm:"column:cycle;type:int;default:1" json:"cycle"`                        // The current cycle the scan is in. Relevant, if scanning in cycles is enabled
	CycleStarted time.Time `gorm:"column:cycle_started;default:CURRENT_TIMESTAMP" json:"cycle_started"` // Timestamp of last cycle start
	CycleDone    float64   `gorm:"column:cycle_done;type:float;default:0" json:"cycle_done"`            // Percentage of completed input scan tasks. Is updated in intervals and not a 100% current.
	CycleActive  float64   `gorm:"column:cycle_active;type:float;default:0" json:"cycle_active"`        // Percentage of active input scan tasks. Is updated in intervals and not a 100% current.
	CycleFailed  float64   `gorm:"column:cycle_failed;type:float;default:0" json:"cycle_failed"`        // Percentage of failed input scan tasks. Is updated in intervals and not a 100% current.

	DbServer     T_db_server     `gorm:"foreignKey:IdTDbServer;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // Database server connection details to connect to
	ScanSettings T_scan_settings `gorm:"foreignKey:IdTScanScope" json:"-"`
	ScanAgents   []T_scan_agent  `gorm:"foreignKey:IdTScanScope" json:"-"` // Scan agent data (cached data)
	ScopeViews   []T_scope_view  `gorm:"foreignKey:IdTScanScope" json:"-"`
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

type T_scan_settings struct {
	Id           uint64 `gorm:"column:id;primaryKey" json:"-"`                                 // Id autoincrement
	IdTScanScope uint64 `gorm:"column:id_t_scan_scope;type:int;not null;uniqueIndex" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions

	MaxInstancesDiscovery        uint32         `gorm:"column:max_instances_discovery;type:int" json:"max_instances_discovery"`   // Maximum parallel instances of discovery scans per agent.
	MaxInstancesBanner           uint32         `gorm:"column:max_instances_banner;type:int" json:"max_instances_banner"`         // Maximum parallel instances of banner scans per agent.
	MaxInstancesNfs              uint32         `gorm:"column:max_instances_nfs;type:int" json:"max_instances_nfs"`               // Maximum parallel instances of nfs scans per agent.
	MaxInstancesSmb              uint32         `gorm:"column:max_instances_smb;type:int" json:"max_instances_smb"`               // Maximum parallel instances of smb scans per agent.
	MaxInstancesSsh              uint32         `gorm:"column:max_instances_ssh;type:int" json:"max_instances_ssh"`               // Maximum parallel instances of ssh scans per agent.
	MaxInstancesSsl              uint32         `gorm:"column:max_instances_ssl;type:int" json:"max_instances_ssl"`               // Maximum parallel instances of ssl scans per agent.
	MaxInstancesWebcrawler       uint32         `gorm:"column:max_instances_webcrawler;type:int" json:"max_instances_webcrawler"` // Maximum parallel instances of webcrawler scans per agent.
	MaxInstancesWebenum          uint32         `gorm:"column:max_instances_webenum;type:int" json:"max_instances_webenum"`       // Maximum parallel instances of webenum scans per agent.
	SensitivePorts               string         `gorm:"column:sensitive_ports;type:text" json:"-"`                                // Comma separated list of sensitive ports that shall not be scanned with submodules
	SensitivePortsSlice          []int          `gorm:"-" json:"sensitive_ports"`                                                 //
	NetworkTimeoutSeconds        int            `gorm:"column:network_timeout_seconds;type:int" json:"network_timeout_seconds"`   //
	HttpUserAgent                string         `gorm:"column:http_user_agent;type:text" json:"http_user_agent"`                  //
	DiscoveryTimeEarliest        string         `gorm:"column:discovery_time_earliest;type:text" json:"discovery_time_earliest"`  //
	DiscoveryTimeLatest          string         `gorm:"column:discovery_time_latest;type:text" json:"discovery_time_latest"`      //
	DiscoverySkipDays            string         `gorm:"column:discovery_skip_days;type:text" json:"-"`                            // Comma separated list of integers (0=Sunday,..., 6=Saturday) where no scanning should take place
	DiscoverySkipDaysSlice       []time.Weekday `gorm:"-" json:"discovery_skip_days"`
	DiscoveryNmapArgs            string         `gorm:"column:discovery_nmap_args;type:text" json:"discovery_nmap_args"`
	DiscoveryNmapArgsPrescan     string         `gorm:"column:discovery_nmap_args_prescan;type:text" json:"discovery_nmap_args_prescan"` // A smaller scan executed before the main scan to at least retrieve some scan results, before a potential IDS kicks in
	DiscoveryExcludeDomains      string         `gorm:"column:discovery_exclude_domains;type:text" json:"discovery_exclude_domains"`
	NfsScanTimeoutMinutes        int            `gorm:"column:nfs_scan_timeout_minutes;type:int" json:"nfs_scan_timeout_minutes"`
	NfsDepth                     int            `gorm:"column:nfs_depth;type:int" json:"nfs_depth"`
	NfsThreads                   int            `gorm:"column:nfs_threads;type:int" json:"nfs_threads"`
	NfsExcludeShares             string         `gorm:"column:nfs_exclude_shares;type:text" json:"nfs_exclude_shares"`
	NfsExcludeFolders            string         `gorm:"column:nfs_exclude_folders;type:text" json:"nfs_exclude_folders"`
	NfsExcludeExtensions         string         `gorm:"column:nfs_exclude_extensions;type:text" json:"nfs_exclude_extensions"`
	NfsExcludeFileSizeBelow      int            `gorm:"column:nfs_exclude_file_size_below;type:int" json:"nfs_exclude_file_size_below"`
	NfsExcludeLastModifiedBelow  time.Time      `gorm:"column:nfs_exclude_last_modified_below;type:datetime" json:"nfs_exclude_last_modified_below"`
	NfsAccessibleOnly            bool           `gorm:"column:nfs_accessible_only;type:bool" json:"nfs_accessible_only"`
	SmbScanTimeoutMinutes        int            `gorm:"column:smb_scan_timeout_minutes;type:int" json:"smb_scan_timeout_minutes"`
	SmbDepth                     int            `gorm:"column:smb_depth;type:int" json:"smb_depth"`
	SmbThreads                   int            `gorm:"column:smb_threads;type:int" json:"smb_threads"`
	SmbExcludeShares             string         `gorm:"column:smb_exclude_shares;type:text" json:"smb_exclude_shares"`
	SmbExcludeFolders            string         `gorm:"column:smb_exclude_folders;type:text" json:"smb_exclude_folders"`
	SmbExcludeExtensions         string         `gorm:"column:smb_exclude_extensions;type:text" json:"smb_exclude_extensions"`
	SmbExcludeFileSizeBelow      int            `gorm:"column:smb_exclude_file_size_below;type:int" json:"smb_exclude_file_size_below"`
	SmbExcludeLastModifiedBelow  time.Time      `gorm:"column:smb_exclude_last_modified_below;type:datetime" json:"smb_exclude_last_modified_below"`
	SmbAccessibleOnly            bool           `gorm:"column:smb_accessible_only;type:bool" json:"smb_accessible_only"`
	SslScanTimeoutMinutes        int            `gorm:"column:ssl_scan_timeout_minutes;type:int" json:"ssl_scan_timeout_minutes"`
	SshScanTimeoutMinutes        int            `gorm:"column:ssh_scan_timeout_minutes;type:int" json:"ssh_scan_timeout_minutes"`
	WebcrawlerScanTimeoutMinutes int            `gorm:"column:webcrawler_scan_timeout_minutes;type:int" json:"webcrawler_scan_timeout_minutes"`
	WebcrawlerDepth              int            `gorm:"column:webcrawler_depth;type:int" json:"webcrawler_depth"`
	WebcrawlerMaxThreads         int            `gorm:"column:webcrawler_max_threads;type:int" json:"webcrawler_max_threads"`
	WebcrawlerFollowQueryStrings bool           `gorm:"column:webcrawler_follow_query_strings;type:bool" json:"webcrawler_follow_query_strings"`
	WebcrawlerAlwaysStoreRoot    bool           `gorm:"column:webcrawler_always_store_root;type:bool" json:"webcrawler_always_store_root"`
	WebcrawlerFollowTypes        string         `gorm:"column:webcrawler_follow_types;type:text" json:"webcrawler_follow_types"`
	WebenumScanTimeoutMinutes    int            `gorm:"column:webenum_scan_timeout_minutes;type:int" json:"webenum_scan_timeout_minutes"`
	WebenumProbeRobots           bool           `gorm:"column:webenum_probe_robots;type:bool" json:"webenum_probe_robots"`

	ScanScope *T_scan_scope `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // Has to be a pointer in order to prevent an invalid recursion. Can be nil if the id has been set
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (scanSettings *T_scan_settings) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize values
	scanSettings.SensitivePorts = b.Sanitize(scanSettings.SensitivePorts)
	tx.Statement.SetColumn("sensitive_ports", scanSettings.SensitivePorts)

	scanSettings.HttpUserAgent = b.Sanitize(scanSettings.HttpUserAgent)
	tx.Statement.SetColumn("http_user_agent", scanSettings.HttpUserAgent)

	scanSettings.DiscoveryTimeEarliest = b.Sanitize(scanSettings.DiscoveryTimeEarliest)
	tx.Statement.SetColumn("discovery_time_earliest", scanSettings.DiscoveryTimeEarliest)

	scanSettings.DiscoveryTimeLatest = b.Sanitize(scanSettings.DiscoveryTimeLatest)
	tx.Statement.SetColumn("discovery_time_latest", scanSettings.DiscoveryTimeLatest)

	scanSettings.DiscoverySkipDays = b.Sanitize(scanSettings.DiscoverySkipDays)
	tx.Statement.SetColumn("discovery_skip_days", scanSettings.DiscoverySkipDays)

	scanSettings.DiscoveryNmapArgs = b.Sanitize(scanSettings.DiscoveryNmapArgs)
	for strings.Contains(scanSettings.DiscoveryNmapArgs, "  ") { // Duplicate spaces in Nmap args may cause scan errors
		scanSettings.DiscoveryNmapArgs = strings.ReplaceAll(scanSettings.DiscoveryNmapArgs, "  ", " ")
	}
	scanSettings.DiscoveryNmapArgs = strings.TrimSpace(scanSettings.DiscoveryNmapArgs)
	tx.Statement.SetColumn("discovery_nmap_args", scanSettings.DiscoveryNmapArgs)

	scanSettings.DiscoveryNmapArgsPrescan = b.Sanitize(scanSettings.DiscoveryNmapArgsPrescan)
	for strings.Contains(scanSettings.DiscoveryNmapArgsPrescan, "  ") { // Duplicate spaces in Nmap args may cause scan errors
		scanSettings.DiscoveryNmapArgsPrescan = strings.ReplaceAll(scanSettings.DiscoveryNmapArgsPrescan, "  ", " ")
	}
	scanSettings.DiscoveryNmapArgsPrescan = strings.TrimSpace(scanSettings.DiscoveryNmapArgsPrescan)
	tx.Statement.SetColumn("discovery_nmap_args_prescan", scanSettings.DiscoveryNmapArgsPrescan)

	scanSettings.DiscoveryExcludeDomains = b.Sanitize(scanSettings.DiscoveryExcludeDomains)
	for strings.Contains(scanSettings.DiscoveryExcludeDomains, " ") { // Remove all spaces
		scanSettings.DiscoveryExcludeDomains = strings.ReplaceAll(scanSettings.DiscoveryExcludeDomains, " ", "")
	}
	for strings.Contains(scanSettings.DiscoveryExcludeDomains, ",,") { // Remove duplicate commas
		scanSettings.DiscoveryExcludeDomains = strings.ReplaceAll(scanSettings.DiscoveryExcludeDomains, ",,", ",")
	}
	scanSettings.DiscoveryExcludeDomains = strings.Trim(scanSettings.DiscoveryExcludeDomains, " ,")
	tx.Statement.SetColumn("discovery_exclude_domains", scanSettings.DiscoveryExcludeDomains)

	scanSettings.SmbExcludeShares = b.Sanitize(scanSettings.SmbExcludeShares)
	tx.Statement.SetColumn("smb_exclude_shares", scanSettings.SmbExcludeShares)

	scanSettings.SmbExcludeFolders = b.Sanitize(scanSettings.SmbExcludeFolders)
	tx.Statement.SetColumn("smb_exclude_folders", scanSettings.SmbExcludeFolders)

	scanSettings.SmbExcludeExtensions = b.Sanitize(scanSettings.SmbExcludeExtensions)
	tx.Statement.SetColumn("smb_exclude_extensions", scanSettings.SmbExcludeExtensions)

	scanSettings.NfsExcludeShares = b.Sanitize(scanSettings.NfsExcludeShares)
	tx.Statement.SetColumn("nfs_exclude_shares", scanSettings.NfsExcludeShares)

	scanSettings.NfsExcludeFolders = b.Sanitize(scanSettings.NfsExcludeFolders)
	tx.Statement.SetColumn("nfs_exclude_folders", scanSettings.NfsExcludeFolders)

	scanSettings.NfsExcludeExtensions = b.Sanitize(scanSettings.NfsExcludeExtensions)
	tx.Statement.SetColumn("nfs_exclude_extensions", scanSettings.NfsExcludeExtensions)

	scanSettings.WebcrawlerFollowTypes = b.Sanitize(scanSettings.WebcrawlerFollowTypes)
	for strings.Contains(scanSettings.WebcrawlerFollowTypes, " ") { // Remove all spaces
		scanSettings.WebcrawlerFollowTypes = strings.ReplaceAll(scanSettings.WebcrawlerFollowTypes, " ", "")
	}
	for strings.Contains(scanSettings.WebcrawlerFollowTypes, ",,") { // Remove duplicate commas
		scanSettings.WebcrawlerFollowTypes = strings.ReplaceAll(scanSettings.WebcrawlerFollowTypes, ",,", ",")
	}
	scanSettings.WebcrawlerFollowTypes = strings.Trim(scanSettings.WebcrawlerFollowTypes, " ,")
	tx.Statement.SetColumn("webcrawler_follow_types", scanSettings.WebcrawlerFollowTypes)

	// Return nil as everything went fine
	return nil
}

// AfterFind updates yet empty struct fields with the derived values. Some setting values are not stored in
// the database but derived from other fields in the database.
func (scanSettings *T_scan_settings) AfterFind(tx *gorm.DB) (err error) {

	// Check if clock values can be parsed
	var errParse error
	timeFormat := "15:04"
	_, errParse = time.Parse(timeFormat, scanSettings.DiscoveryTimeEarliest)
	if errParse != nil {
		return fmt.Errorf("could not deserialize settings from database: %s", errParse)
	}
	_, errParse = time.Parse(timeFormat, scanSettings.DiscoveryTimeLatest)
	if errParse != nil {
		return fmt.Errorf("could not deserialize settings from database: %s", errParse)
	}

	// Parse list of sensitive ports
	if len(scanSettings.SensitivePorts) > 0 {
		for _, port := range strings.Split(scanSettings.SensitivePorts, ",") {
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

	// Parse list of days
	if len(scanSettings.DiscoverySkipDays) > 0 {
		for _, day := range strings.Split(scanSettings.DiscoverySkipDays, ",") {
			dayInt, errDay := strconv.ParseInt(day, 10, 64)
			if errDay != nil {
				return fmt.Errorf("invalid day '%s' (0=Sunday,...,6=Saturday)", day)
			}
			if dayInt < 0 || dayInt > 6 {
				return fmt.Errorf("invalid day '%s' (0=Sunday,...,6=Saturday)", day)
			}
			scanSettings.DiscoverySkipDaysSlice = append(scanSettings.DiscoverySkipDaysSlice, time.Weekday(dayInt))
		}
	}

	// Return as everything went fine
	return
}

func (scanSettings *T_scan_settings) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux T_scan_settings
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check if clock values can be parsed
	timeFormat := "15:04"
	var errParse error
	_, errParse = time.Parse(timeFormat, raw.DiscoveryTimeEarliest)
	if errParse != nil {
		return fmt.Errorf("invalid discovery start time (HH:mm)")
	}
	_, errParse = time.Parse(timeFormat, raw.DiscoveryTimeLatest)
	if errParse != nil {
		return fmt.Errorf("invalid discovery end time (HH:mm)")
	}

	// Do other input validation
	if raw.MaxInstancesDiscovery < 0 {
		return fmt.Errorf("invalid max instances discovery")
	}
	if raw.MaxInstancesBanner < 0 {
		return fmt.Errorf("invalid max instances banner")
	}
	if raw.MaxInstancesNfs < 0 {
		return fmt.Errorf("invalid max instances nfs")
	}
	if raw.MaxInstancesSmb < 0 {
		return fmt.Errorf("invalid max instances smb")
	}
	if raw.MaxInstancesSsh < 0 {
		return fmt.Errorf("invalid max instances ssh")
	}
	if raw.MaxInstancesSsl < 0 {
		return fmt.Errorf("invalid max instances ssl")
	}
	if raw.MaxInstancesWebcrawler < 0 {
		return fmt.Errorf("invalid max instances webcrawler")
	}
	if raw.MaxInstancesWebenum < 0 {
		return fmt.Errorf("invalid max instances webenum")
	}
	if raw.NetworkTimeoutSeconds <= 0 {
		return fmt.Errorf("invalid network timeout")
	}
	if raw.NfsScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid nfs scan timeout")
	}
	if raw.NfsThreads <= 0 {
		return fmt.Errorf("invalid nfs threads")
	}
	if raw.SmbScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid smb scan timeout")
	}
	if raw.SmbThreads <= 0 {
		return fmt.Errorf("invalid smb threads")
	}
	if raw.SslScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid ssl scan timeout")
	}
	if raw.SshScanTimeoutMinutes <= 0 {
		return fmt.Errorf("invalid ssh scan timeout")
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

	// Copy loaded Json values to actual scan settings
	*scanSettings = T_scan_settings(raw)

	// Check if port values can be parsed
	if len(scanSettings.SensitivePortsSlice) > 0 {
		for _, port := range scanSettings.SensitivePortsSlice {
			if port < 0 || port > 65535 {
				return fmt.Errorf("invalid port '%d'", port)
			}
		}
		scanSettings.SensitivePorts = utils.JoinInt(scanSettings.SensitivePortsSlice, ",")
	}

	// JSON unmarshalling is called when settings are read from the config or from a user HTTP request, use this
	// opportunity to unique and sort values, so that they are sanitized before storing
	scanSettings.SensitivePortsSlice = utils.UniqueInts(scanSettings.SensitivePortsSlice)
	sort.Ints(scanSettings.SensitivePortsSlice)

	// Check if day values can be parsed
	if len(scanSettings.DiscoverySkipDaysSlice) > 0 {
		for _, day := range scanSettings.DiscoverySkipDaysSlice {
			if day < 0 || day > 6 {
				return fmt.Errorf("invalid day '%d' (0=Sunday,...,6=Saturday)", day)
			}
		}
		scanSettings.DiscoverySkipDays = utils.JoinWeekdays(scanSettings.DiscoverySkipDaysSlice, ",")
	}

	// JSON unmarshalling is called when settings are read from the config or from a user HTTP request, use this
	// opportunity to unique and sort values, so that they are sanitized before storing
	scanSettings.DiscoverySkipDaysSlice = utils.UniqueWeekdays(scanSettings.DiscoverySkipDaysSlice)
	sort.Slice(scanSettings.SensitivePortsSlice, func(i int, j int) bool {
		if scanSettings.SensitivePortsSlice[i] < scanSettings.SensitivePortsSlice[j] {
			return true
		} else {
			return false
		}
	})

	// Return nil as everything is valid
	return nil
}

// SaveAll updates *all* columns of a user entry in the database. It overwrites existing values with the ones passed in
func (scanSettings *T_scan_settings) SaveAll(values *T_scan_settings) (int64, error) {

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
func (scanSettings *T_scan_settings) MaxInstances(label string) (int, error) {

	// Get instances for module
	switch label {
	case discovery.Label:
		return int(scanSettings.MaxInstancesDiscovery), nil
	case banner.Label:
		return int(scanSettings.MaxInstancesBanner), nil
	case nfs.Label:
		return int(scanSettings.MaxInstancesNfs), nil
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

	LastSeen   time.Time     `gorm:"column:last_seen;default:CURRENT_TIMESTAMP" json:"last_seen"`
	Tasks      utils.JsonMap `gorm:"column:tasks;default:'{}'" json:"tasks"`
	CpuRate    float64       `gorm:"column:cpu_rate;type:float;default:0" json:"cpu_rate"`
	MemoryRate float64       `gorm:"column:memory_rate;type:float;default:0" json:"memory_rate"`

	Platform        string `gorm:"column:platform;type:text;" json:"platform"`
	PlatformFamily  string `gorm:"column:platform_family;type:text;" json:"platform_family"`
	PlatformVersion string `gorm:"column:platform_version;type:text;" json:"platform_version"`

	ScanScope T_scan_scope `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // Can be empty if the ID is set in turn
}

// MaxBatchSizeScanAgent defines the maximum number T_scan_agent instances that can be batched together
// during an insert. This is calculated dividing 999 (SQLITE) by the number of fields (that are actually written to the db).
const MaxBatchSizeScanAgent = 76 // 999 (SQLITE) / 13

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

	ScanScope T_scan_scope   `gorm:"foreignKey:IdTScanScope;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Grants    []T_view_grant `gorm:"foreignKey:IdTScopeView" json:"-"`
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

	ScopeView T_scope_view `gorm:"foreignKey:IdTScopeView;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
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
