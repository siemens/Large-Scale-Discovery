/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"database/sql"
)

//
// Data compositions for redundant columns
//

// ColumnsInput contains input data used by the discovery module to discover initial network devices and ports
type ColumnsInput struct {
	Input     string `gorm:"column:input;type:text" json:"input"`
	InputSize uint   `gorm:"column:input_size" json:"-"`
}

// ColumnsHost combines common host columns and will be appended to selected database tables
type ColumnsHost struct {
	Address    string `gorm:"column:address;type:text;not null;index"`
	Ip         string `gorm:"column:ip;type:text;not null;index"`
	DnsName    string `gorm:"column:dns_name;type:text"`
	OtherNames string `gorm:"column:other_names;type:text"`
	Hops       string `gorm:"column:hops;type:text"`
	ScanCycle  uint   `gorm:"column:scan_cycle;index"`
}

// ColumnsOs combines common os columns and will be appended to selected database tables
type ColumnsOs struct {
	OsGuess      string        `gorm:"column:os_guess;type:text"`
	OsSmb        string        `gorm:"column:os_smb;type:text"`
	OsLastBoot   sql.NullTime  `gorm:"column:os_last_boot"`
	OsUptime     sql.NullInt64 `gorm:"column:os_uptime"`
	OsAdminUsers string        `gorm:"column:os_admin_users;type:text"`
	OsRdpUsers   string        `gorm:"column:os_rdp_users;type:text"`
}

// ColumnsScan combines common scan data columns and will be appended to selected database tables
type ColumnsScan struct {
	ScanStarted  sql.NullTime `gorm:"column:scan_started;index" json:"scan_started"`
	ScanFinished sql.NullTime `gorm:"column:scan_finished;index" json:"scan_finished"`
	ScanStatus   string       `gorm:"column:scan_status;type:text;default:'Waiting';index" json:"-"`
	ScanIp       string       `gorm:"column:scan_ip;type:text;not null" json:"-"`
	ScanHostname string       `gorm:"column:scan_hostname;type:text;not null" json:"-"`
}

// ColumnsInputDetails combines common detail related to an input and will be appended to selected database tables.
// Basically some optional/additional fields adding some story to t_discovery_input (a network, hostname or IP).
type ColumnsInputDetails struct {
	Timezone           float32 `gorm:"column:timezone;type:float;default:0" json:"timezone"` // Float because timezones may be 11.5 in certain edge cases
	Lat                string  `gorm:"column:lat;type:text" json:"lat"`
	Lng                string  `gorm:"column:lng;type:text" json:"lng"`
	PostalAddress      string  `gorm:"column:postal_address;type:text" json:"postal_address"`
	InputNetwork       string  `gorm:"column:input_network;type:text" json:"input_network"`               // Network the input address belongs to. Equals input if input is already a network range
	InputCountry       string  `gorm:"column:input_country;type:text" json:"input_country"`               // E.g. "DE"
	InputLocation      string  `gorm:"column:input_location;type:text" json:"input_location"`             // E.g. "Munich"
	InputRoutingDomain string  `gorm:"column:input_routing_domain;type:text" json:"input_routing_domain"` // E.g. "Global", "local", "Intranet", "Internet",...
	InputZone          string  `gorm:"column:input_zone;type:text" json:"input_zone"`                     // E.g. "Office", "Production A", "Printer",...
	InputPurpose       string  `gorm:"column:input_purpose;type:text" json:"input_purpose"`               // E.g. network description, like, "transfer network", "office space",...
	InputCompany       string  `gorm:"column:input_company;type:text" json:"input_company"`               // E.g. company name, useful in a multi company network
	InputDepartment    string  `gorm:"column:input_department;type:text" json:"input_department"`         // E.g. company department, e.g. IT Services
	InputManager       string  `gorm:"column:input_manager;type:text" json:"input_manager"`               // E.g. the responsible manager
	InputContact       string  `gorm:"column:input_contact;type:text" json:"input_contact"`               // E.g. the responsible administrator
	InputComment       string  `gorm:"column:input_comment;type:text" json:"input_comment"`               // E.g. anything that helps
}

// ColumnsAd combines common AD columns and will be appended to selected database tables
type ColumnsAd struct {
	AdName                 string       `gorm:"column:ad_name;type:text"`
	AdDistinguishedName    string       `gorm:"column:ad_distinguished_name;type:text"`
	AdDnsName              string       `gorm:"column:ad_dns_name;type:text"`
	AdCreated              sql.NullTime `gorm:"column:ad_created"`
	AdLastLogon            sql.NullTime `gorm:"column:ad_last_logon"`
	AdLastPassword         sql.NullTime `gorm:"column:ad_last_password"`
	AdDescription          string       `gorm:"column:ad_description;type:text"`
	AdLocation             string       `gorm:"column:ad_location;type:text"`
	AdManagedBy            string       `gorm:"column:ad_managed_by;type:text"`
	AdManagedByGid         string       `gorm:"column:ad_managed_by_gid;type:text"`
	AdManagedByDepartment  string       `gorm:"column:ad_managed_by_ou;type:text"`
	AdOs                   string       `gorm:"column:ad_os;type:text"`
	AdOsServicePack        string       `gorm:"column:ad_os_service_pack;type:text"`
	AdOsVersion            string       `gorm:"column:ad_os_version;type:text"`
	AdServicePrincipalName string       `gorm:"column:ad_service_principal_name;type:text"`
	AdCriticalObject       bool         `gorm:"column:ad_critical_object"`
}

//
// Database model definitions
//

// MaxBatchSizeDiscovery defines the maximum number of T_discovery instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeDiscovery = 2520 // 65535 / 26

type T_discovery struct {
	Id uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex" json:"-"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsInput
	Enabled             bool `gorm:"column:enabled;index" json:"enabled"`
	Priority            bool `gorm:"column:priority;default:false;index" json:"priority"`
	ScanCount           uint `gorm:"column:scan_count;default:0" json:"-"`
	ColumnsScan              // Insert scan data columns composition
	ColumnsInputDetails      // Insert input detail columns composition
}

func (T_discovery) TableName() string {
	return "t_discovery"
}

// MaxBatchSizeDiscoveryHost defines the maximum number T_discovery_host instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeDiscoveryHost = 1236 // 65535 / 53

type T_discovery_host struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscovery        uint64 `gorm:"column:id_t_discovery;type:bigint;index"`         // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsHost                // Insert host data columns composition
	PortsOpen           int    `gorm:"column:ports_open"`
	ColumnsOs                  // Insert target columns composition
	ColumnsScan                // Insert scan data columns composition
	ColumnsInput               // Insert copy of input (original discovery)  data, because t_discovery contents may change over time
	ColumnsInputDetails        // Insert input detail columns composition
	ColumnsAd                  // Insert AD columns composition

	TDiscovery *T_discovery `gorm:"foreignKey:IdTDiscovery;constraint:OnUpdate:SET NULL,OnDelete:SET NULL"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeDiscoveryService defines the maximum number T_discovery_service instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeDiscoveryService = 1074 // 65535 / 61

type T_discovery_service struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryHost    uint64 `gorm:"column:id_t_discovery_host;type:bigint;index"`    // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsHost                // Insert host data columns composition
	Port                int    `gorm:"column:port"`                          //
	Protocol            string `gorm:"column:protocol;type:text"`            //
	Service             string `gorm:"column:service;type:text"`             //
	ServiceProduct      string `gorm:"column:service_product;type:text"`     //
	ServiceVersion      string `gorm:"column:service_version;type:text"`     //
	ServiceDeviceType   string `gorm:"column:service_device_type;type:text"` //
	ServiceCpes         string `gorm:"column:service_cpes;type:text"`        //
	ServiceFlavor       string `gorm:"column:service_flavor;type:text"`      //
	ServiceTtl          int    `gorm:"column:service_ttl"`                   //
	ColumnsOs                  // Insert target columns composition
	ColumnsScan                // Insert scan data columns composition
	ColumnsInput               // Insert copy of input (original discovery)  data, because t_discovery contents may change over time
	ColumnsInputDetails        // Insert input detail columns composition
	ColumnsAd                  // Insert AD columns composition

	TDiscoveryHost *T_discovery_host `gorm:"foreignKey:IdTDiscoveryHost;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeDiscoveryScript defines the maximum number T_discovery_script instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeDiscoveryScript = 1149 // 65535 / 57

type T_discovery_script struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryHost    uint64 `gorm:"column:id_t_discovery_host;type:bigint;index"`    // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsHost                // Insert host data columns composition
	Port                int    `gorm:"column:port"`                    //
	Protocol            string `gorm:"column:protocol;type:text"`      //
	ScriptType          string `gorm:"column:script_type;type:text"`   //
	ScriptName          string `gorm:"column:script_name;type:text"`   //
	ScriptOutput        string `gorm:"column:script_output;type:text"` //
	ColumnsOs                  // Insert target columns composition
	ColumnsScan                // Insert scan data columns composition
	ColumnsInput               // Insert copy of input (original discovery)  data, because t_discovery contents may change over time
	ColumnsInputDetails        // Insert input detail columns composition
	ColumnsAd                  // Insert AD columns composition

	TDiscoveryHost *T_discovery_host `gorm:"foreignKey:IdTDiscoveryHost;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeBanner defines the maximum number T_banner instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeBanner = 5461 // 65535 / 12

type T_banner struct {
	// data table not necessary for banner module, due to simple result structure
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition
	TriggerPlain        string `gorm:"column:trigger_plain;type:text"`
	TriggerSsl          string `gorm:"column:trigger_ssl;type:text"`
	TriggerTelnet       string `gorm:"column:trigger_telnet;type:text"`
	TriggerHttp         string `gorm:"column:trigger_http;type:text"`
	TriggerHttps        string `gorm:"column:trigger_https;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeNfs defines the maximum number T_nfs instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeNfs = 6553 // 65535 / 10

type T_nfs struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition
	FoldersReadable     int    `gorm:"column:folders_readable"`
	FilesReadable       int    `gorm:"column:files_readable"`
	FilesWritable       int    `gorm:"column:files_writable"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_nfs) TableName() string {
	return "t_nfs"
}

// MaxBatchSizeNfsFile defines the maximum number T_nfs_file instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeNfsFile = 3855 // 65535 / 17

type T_nfs_file struct {
	Id                  uint64       `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64       `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTNfs              uint64       `gorm:"column:id_t_nfs;type:bigint;index"`               // Index recommended on foreign keys for efficient update/delete cascaded actions
	Share               string       `gorm:"column:share;type:text"`
	Path                string       `gorm:"column:path;type:text"`
	Name                string       `gorm:"column:name;type:text"`
	Extension           string       `gorm:"column:extension;type:text"`
	Mime                string       `gorm:"column:mime;type:text"`
	Readable            bool         `gorm:"column:readable"`
	Writable            bool         `gorm:"column:writable"`
	Flags               string       `gorm:"column:flags;type:text"`
	SizeKb              int64        `gorm:"column:size_kb"`
	LastModified        sql.NullTime `gorm:"column:last_modified"`
	Depth               int          `gorm:"column:depth"`
	Properties          string       `gorm:"column:properties;type:text"`
	IsSymlink           bool         `gorm:"column:is_symlink"`
	Restrictions        string       `gorm:"column:restrictions;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TNfs              *T_nfs               `gorm:"foreignKey:IdTNfs;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`              // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeSmb defines the maximum number T_smb instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSmb = 6553 // 65535 / 10

type T_smb struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition
	FoldersReadable     int    `gorm:"column:folders_readable"`
	FilesReadable       int    `gorm:"column:files_readable"`
	FilesWritable       int    `gorm:"column:files_writable"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_smb) TableName() string {
	return "t_smb"
}

// MaxBatchSizeSmbFile defines the maximum number T_smb_file instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSmbFile = 4095 // 65535 / 16

type T_smb_file struct {
	Id                  uint64       `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64       `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTSmb              uint64       `gorm:"column:id_t_smb;type:bigint;index"`               // Index recommended on foreign keys for efficient update/delete cascaded actions
	Share               string       `gorm:"column:share;type:text"`
	Path                string       `gorm:"column:path;type:text"`
	Name                string       `gorm:"column:name;type:text"`
	Extension           string       `gorm:"column:extension;type:text"`
	Mime                string       `gorm:"column:mime;type:text"`
	Readable            bool         `gorm:"column:readable"`
	Writable            bool         `gorm:"column:writable"`
	SizeKb              int64        `gorm:"column:size_kb"`
	LastModified        sql.NullTime `gorm:"column:last_modified"`
	Depth               int          `gorm:"column:depth"`
	Properties          string       `gorm:"column:properties;type:text"`
	IsSymlink           bool         `gorm:"column:is_symlink"`
	IsDfs               bool         `gorm:"column:is_dfs"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TSmb              *T_smb               `gorm:"foreignKey:IdTSmb;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`              // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeSsh defines the maximum number T_ssh instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSsh = 4369 // 65535 / 15

type T_ssh struct {
	// data table not necessary for banner module, due to simple result structure
	Id                         uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService        uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                       // Insert scan data columns composition
	AuthenticationMechanisms   string `gorm:"column:authentication_mechanisms;type:text"`
	KeyExchangeAlgorithms      string `gorm:"column:key_exchange_algos;type:text"`
	ServerKeyAlgorithms        string `gorm:"column:server_key_algos;type:text"`
	ServerEncryptionAlgorithms string `gorm:"column:server_encrypt_algos;type:text"`
	ServerMacAlgorithms        string `gorm:"column:server_mac_algos;type:text"`
	ServerCompressAlgorithms   string `gorm:"column:server_compress_algos;type:text"`
	UsesGuessedKeyExchange     bool   `gorm:"column:uses_guessed_key_exchange"`
	ProtocolVersion            string `gorm:"column:protocol_version;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_ssh) TableName() string {
	return "t_ssh"
}

// MaxBatchSizeSsl defines the maximum number T_ssl instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSsl = 9362 // 65535 / 7

type T_ssl struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_ssl) TableName() string {
	return "t_ssl"
}

// MaxBatchSizeSslCertificate defines the maximum number T_ssl_certificate instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSslCertificate = 2047 // 65535 / 32

type T_ssl_certificate struct {
	Id                     uint64       `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService    uint64       `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTSsl                 uint64       `gorm:"column:id_t_ssl;type:bigint;index"`               // Index recommended on foreign keys for efficient update/delete cascaded actions
	Vhost                  string       `gorm:"column:vhost;type:text"`
	DeploymentId           uint64       `gorm:"column:deployment_id"` // Not a primary/foreign key nor an unique identifier!
	Type                   string       `gorm:"column:type;type:text"`
	Version                int          `gorm:"column:version"`
	Serial                 string       `gorm:"column:serial_number;type:text"`
	ValidChain             bool         `gorm:"column:valid_chain"`
	ChainValidatedBy       string       `gorm:"column:chain_validated_by;type:text"`
	ValidChainOrder        bool         `gorm:"column:valid_chain_order"`
	Subject                string       `gorm:"column:subject;type:text"`
	SubjectCN              string       `gorm:"column:subject_cn;type:text"`
	Issuer                 string       `gorm:"column:issuer;type:text"`
	IssuerCN               string       `gorm:"column:issuer_cn;type:text"`
	AlternativeNames       string       `gorm:"column:alternative_names;type:text"`
	ValidFrom              sql.NullTime `gorm:"column:valid_from"`
	ValidTo                sql.NullTime `gorm:"column:valid_to"`
	PublicKeyAlgorithm     string       `gorm:"column:public_key_algorithm;type:text"`
	PublicKeyInfo          string       `gorm:"column:public_key_info;type:text"`
	PublicKeyBits          uint64       `gorm:"column:public_key_bits"`
	PublicKeyStrength      int          `gorm:"column:public_key_strength"`
	SignatureAlgorithm     string       `gorm:"column:signature_algorithm;type:text"`
	SignatureHashAlgorithm string       `gorm:"column:signature_hash_algorithm;type:text"`
	CrlUrls                string       `gorm:"column:crl_urls;type:text"`
	OcspUrls               string       `gorm:"column:ocsp_urls;type:text"`
	KeyUsage               string       `gorm:"column:key_usage;type:text"`
	ExtendedKeyUsage       string       `gorm:"column:extended_key_usage;type:text"`
	BasicConstraintsValid  bool         `gorm:"column:basic_constraints_valid"`
	Ca                     bool         `gorm:"column:ca"`
	MaxPathLength          int          `gorm:"column:max_path_length"`
	Sha1Fingerprint        string       `gorm:"column:sha1_fingerprint;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TSsl              *T_ssl               `gorm:"foreignKey:IdTSsl;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`              // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeSslCipher defines the maximum number T_ssl_cipher instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSslCipher = 1986 // 65535 / 33

type T_ssl_cipher struct {
	Id                      uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService     uint64 `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTSsl                  uint64 `gorm:"column:id_t_ssl;type:bigint;index"`               // Index recommended on foreign keys for efficient update/delete cascaded actions
	Vhost                   string `gorm:"column:vhost;type:text"`
	ProtocolVersion         string `gorm:"column:protocol_version;type:text"`
	CipherId                string `gorm:"column:cipher_id;type:text"`
	IanaName                string `gorm:"column:iana_name;type:text"`
	OpensslName             string `gorm:"column:open_ssl_name;type:text"`
	SupportsECDHKEyExchange bool   `gorm:"column:support_ecdh_key_exchange"`
	SupportedEllipticCurves string `gorm:"column:supported_elliptic_curves;type:text"`
	RejectedEllipticCurves  string `gorm:"column:rejected_elliptic_curves;type:text"`
	KeyExchange             string `gorm:"column:key_exchange;type:text"`
	KeyExchangeBits         int    `gorm:"column:key_exchange_bits"`
	KeyExchangeStrength     int    `gorm:"column:key_exchange_strength"`
	KeyExchangeInfo         string `gorm:"column:key_exchange_info"`
	ForwardSecrecy          bool   `gorm:"column:forward_secrecy"`
	Authentication          string `gorm:"column:authentication;type:text"`
	Encryption              string `gorm:"column:encryption;type:text"`
	EncryptionMode          string `gorm:"column:encryption_mode;type:text"`
	EncryptionBits          int    `gorm:"column:encryption_bits"`
	EncryptionStrength      int    `gorm:"column:encryption_strength"`
	BlockCipher             bool   `gorm:"column:block_cipher"`
	BlockSize               int    `gorm:"column:block_size"`
	StreamCipher            bool   `gorm:"column:stream_cipher"`
	Mac                     string `gorm:"column:mac;type:text"`
	MacBits                 int    `gorm:"column:mac_bits"`
	MacStrength             int    `gorm:"column:mac_strength"`
	Prf                     string `gorm:"column:prf;type:text"`
	PrfBits                 int    `gorm:"column:prf_bits"`
	PrfStrength             int    `gorm:"column:prf_strength"`
	Export                  bool   `gorm:"column:export"`
	Draft                   bool   `gorm:"column:draft"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TSsl              *T_ssl               `gorm:"foreignKey:IdTSsl;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`              // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeSslIssue defines the maximum number T_ssl_issue instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeSslIssue = 1873 // 65535 / 35

type T_ssl_issue struct {
	Id                           uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService          uint64 `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTSsl                       uint64 `gorm:"column:id_t_ssl;type:bigint;index"`               // Index recommended on foreign keys for efficient update/delete cascaded actions
	Vhost                        string `gorm:"column:vhost;type:text"`
	AnyChainInvalid              bool   `gorm:"column:any_chain_invalid"`
	AnyChainInvalidOrder         bool   `gorm:"column:any_chain_invalid_order"`
	LowestProtocol               string `gorm:"column:lowest_protocol;type:text"`
	MinStrength                  int    `gorm:"column:min_strength"`
	InsecureRenegotiation        bool   `gorm:"column:insecure_renegotiation"`
	AcceptsClientRenegotiation   bool   `gorm:"column:accepts_client_renegotiation"`
	InsecureClientRenegotiation  bool   `gorm:"column:insecure_client_renegotiation"`
	SessionResumptionWithId      bool   `gorm:"column:session_resumption_with_id"`
	SessionResumptionWithTickets bool   `gorm:"column:session_resumption_with_tickets"`
	NoPerfectForwardSecrecy      bool   `gorm:"column:no_perfect_forward_secrecy"`
	Compression                  bool   `gorm:"column:compression"`
	ExportSuite                  bool   `gorm:"column:export_suite"`
	DraftSuite                   bool   `gorm:"column:draft_suite"`
	Sslv2Enabled                 bool   `gorm:"column:sslv2_enabled"`
	Sslv3Enabled                 bool   `gorm:"column:sslv3_enabled"`
	Rc4Enabled                   bool   `gorm:"column:rc4_enabled"`
	Md2Enabled                   bool   `gorm:"column:md2_enabled"`
	Md5Enabled                   bool   `gorm:"column:md5_enabled"`
	Sha1Enabled                  bool   `gorm:"column:sha1_enabled"`
	EarlyDataSupported           bool   `gorm:"column:early_data_supported"`
	CcsInjection                 bool   `gorm:"column:ccs_injection"`
	Beast                        bool   `gorm:"column:beast"`
	Heartbleed                   bool   `gorm:"column:heartbleed"`
	Lucky13                      bool   `gorm:"column:lucky_13"`
	Poodle                       bool   `gorm:"column:poodle"`
	Freak                        bool   `gorm:"column:freak"`
	Logjam                       bool   `gorm:"column:logjam"`
	Sweet32                      bool   `gorm:"column:sweet_32"`
	Drown                        bool   `gorm:"column:drown"`
	IsCompliantToMozillaConfig   bool   `gorm:"column:is_compliant_to_mozilla_config"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TSsl              *T_ssl               `gorm:"foreignKey:IdTSsl;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`              // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeWebcrawler defines the maximum number T_webcrawler instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeWebcrawler = 9362 // 65535 / 7

type T_webcrawler struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_webcrawler) TableName() string {
	return "t_webcrawler"
}

// MaxBatchSizeWebcrawlerVhost defines the maximum number T_webcrawler_vhost instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeWebcrawlerVhost = 4681 // 65535 / 14

type T_webcrawler_vhost struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTWebcrawler       uint64 `gorm:"column:id_t_webcrawler;type:bigint;index"`        // Index recommended on foreign keys for efficient update/delete cascaded actions
	Status              string `gorm:"column:status;type:text"`
	Vhost               string `gorm:"column:vhost;type:text"`
	FaviconHash         string `gorm:"column:favicon_hash;type:text"`
	AuthMethod          string `gorm:"column:auth_method;type:text"`
	AuthSuccess         bool   `gorm:"column:auth_success"`
	RequestsTotal       int    `gorm:"column:requests_total"`
	RequestsRedirect    int    `gorm:"column:requests_redirect"`
	RequestsPartial     int    `gorm:"column:requests_partial"`
	RequestsComplete    int    `gorm:"column:requests_complete"`
	DiscoveredVhosts    string `gorm:"column:discovered_vhosts;type:text"`
	DiscoveredDownloads string `gorm:"column:discovered_downloads;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TWebcrawler       *T_webcrawler        `gorm:"foreignKey:IdTWebcrawler;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`       // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeWebcrawlerPage defines the maximum number T_webcrawler_page instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeWebcrawlerPage = 3449 // 65535 / 19

type T_webcrawler_page struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTWebcrawler       uint64 `gorm:"column:id_t_webcrawler;type:bigint;index"`        // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTWebcrawlerVhost  uint64 `gorm:"column:id_t_webcrawler_vhost;type:bigint;index"`  // Index recommended on foreign keys for efficient update/delete cascaded actions
	Depth               int    `gorm:"column:depth"`
	Url                 string `gorm:"column:url;type:text"`
	RedirectUrl         string `gorm:"column:redirect_url;type:text"`
	RedirectCount       int    `gorm:"column:redirect_count"`
	AuthMethod          string `gorm:"column:auth_method;type:text"`
	AuthSuccess         bool   `gorm:"column:auth_success"`
	ResponseCode        int    `gorm:"column:response_code"`
	ResponseMessage     string `gorm:"column:response_message;type:text"`
	ResponseContentType string `gorm:"column:response_content_type;type:text"`
	ResponseHeaders     string `gorm:"column:response_headers;type:text"`
	ResponseEncoding    string `gorm:"column:response_encoding;type:text;default:''"`
	HtmlTitle           string `gorm:"column:html_title;type:text"`
	HtmlContent         string `gorm:"column:html_content;type:text"`
	HtmlContentLength   int    `gorm:"column:html_content_length;"`
	RawLinks            string `gorm:"column:raw_links;type:text"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TWebcrawler       *T_webcrawler        `gorm:"foreignKey:IdTWebcrawler;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`       // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TWebcrawlerVhost  *T_webcrawler_vhost  `gorm:"foreignKey:IdTWebcrawlerVhost;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`  // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

// MaxBatchSizeWebenum defines the maximum number T_webenum instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeWebenum = 9362 // 65535 / 7

type T_webenum struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"`       // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	ColumnsScan                // Insert scan data columns composition

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}

func (T_webenum) TableName() string {
	return "t_webenum"
}

// MaxBatchSizeWebenumResult defines the maximum number T_webenum_results instances that can be batched together
// during an insert. This is calculated dividing 65535 by the number of fields (that are actually written to the db).
const MaxBatchSizeWebenumResult = 3640 // 3449 / 19

type T_webenum_results struct {
	Id                  uint64 `gorm:"column:id;type:bigserial;primaryKey;uniqueIndex"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTDiscoveryService uint64 `gorm:"column:id_t_discovery_service;type:bigint;index"` // Index recommended on foreign keys for efficient update/delete cascaded actions
	IdTWebenum          uint64 `gorm:"column:id_t_webenum;type:bigint;index"`           // Index recommended on foreign keys for efficient update/delete cascaded actions
	Name                string `gorm:"column:name;type:text"`
	Vhost               string `gorm:"column:vhost;type:text"`
	Url                 string `gorm:"column:url;type:text"`
	RedirectUrl         string `gorm:"column:redirect_url;type:text"`
	RedirectCount       int    `gorm:"column:redirect_count"`
	RedirectOut         bool   `gorm:"column:redirect_out"`
	AuthMethod          string `gorm:"column:auth_method;type:text"`
	AuthSuccess         bool   `gorm:"column:auth_success"`
	ResponseCode        int    `gorm:"column:response_code"`
	ResponseMessage     string `gorm:"column:response_message;type:text"`
	ResponseContentType string `gorm:"column:response_content_type;type:text"`
	ResponseHeaders     string `gorm:"column:response_headers;type:text"`
	ResponseEncoding    string `gorm:"column:response_encoding;type:text;default:''"`
	HtmlTitle           string `gorm:"column:html_title;type:text"`
	HtmlContent         string `gorm:"column:html_content;type:text"`
	HtmlContentLength   int    `gorm:"column:html_content_length"`

	TDiscoveryService *T_discovery_service `gorm:"foreignKey:IdTDiscoveryService;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
	TWebenum          *T_webenum           `gorm:"foreignKey:IdTWebenum;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`          // Relation struct used for gorm configuration and batch inserts (where it can be used to keep track of the IDs) and to enforce constraints. Can be nil if the ID is set in turn
}
