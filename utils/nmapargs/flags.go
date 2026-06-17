/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package nmapargs

import "math"

// argKind describes what a flag expects after it.
type argKind int

const (
	argNone     argKind = iota // flag takes no argument (e.g. -sS)
	argRequired                // flag must be followed by an argument (e.g. -p <ports>)
	argOptional                // argument may be glued directly or absent (e.g. -PS[portlist])
)

// flagDef defines one Nmap flag and how its argument should be validated.
type flagDef struct {
	kind      argKind
	validator func(string) error // nil means any non-empty value is accepted
}

// longFlags maps every supported long flag (--flag) to its definition.
var longFlags = map[string]flagDef{

	// Target specification.
	"--exclude":       {argRequired, validateAny},
	"--excludefile":   {argRequired, validateAny},
	"--exclude-ports": {argRequired, validatePortList},

	// Host discovery.
	"--traceroute":           {argNone, nil},
	"--system-dns":           {argNone, nil},
	"--dns-servers":          {argRequired, validateAny},
	"--resolve-all":          {argNone, nil},
	"--unique":               {argNone, nil},
	"--disable-arp-ping":     {argNone, nil},
	"--discovery-ignore-rst": {argNone, nil},

	// Scan technique.
	"--scanflags": {argRequired, validateScanFlags},

	// Port specification and scan order.
	"--top-ports":       {argRequired, validateIntRange(1, 65535)},
	"--port-ratio":      {argRequired, validatePortRatio},
	"--randomize-hosts": {argNone, nil},
	"--rH":              {argNone, nil}, // Alias for --randomize-hosts.

	// Version detection.
	"--version-intensity": {argRequired, validateIntRange(0, 9)},
	"--version-light":     {argNone, nil},
	"--version-all":       {argNone, nil},
	"--version-trace":     {argNone, nil},
	"--allports":          {argNone, nil},

	// Script scanning.
	"--script":           {argRequired, validateAny},
	"--script-args":      {argRequired, validateAny},
	"--script-args-file": {argRequired, validateAny},
	"--script-trace":     {argNone, nil},
	"--script-updatedb":  {argNone, nil},
	"--script-help":      {argRequired, validateAny},
	"--script-timeout":   {argRequired, validateTime},

	// OS detection.
	"--osscan-limit": {argNone, nil},
	"--osscan-guess": {argNone, nil},
	"--fuzzy":        {argNone, nil},
	"--max-os-tries": {argRequired, validateIntRange(1, 50)},

	// Timing and performance.
	"--nogcc":                 {argNone, nil},
	"--min-hostgroup":         {argRequired, validateIntRange(0, math.MaxInt)},
	"--max-hostgroup":         {argRequired, validateIntRange(1, math.MaxInt)},
	"--min-parallelism":       {argRequired, validateIntRange(0, math.MaxInt)},
	"--max-parallelism":       {argRequired, validateIntRange(0, math.MaxInt)},
	"--min-rtt-timeout":       {argRequired, validateTime},
	"--max-rtt-timeout":       {argRequired, validateTimePositive},
	"--initial-rtt-timeout":   {argRequired, validateTimePositive},
	"--max-retries":           {argRequired, validateIntRange(0, math.MaxInt)},
	"--host-timeout":          {argRequired, validateTime},
	"--scan-delay":            {argRequired, validateTime},
	"--max-scan-delay":        {argRequired, validateTime},
	"--min-rate":              {argRequired, validateFloatRange(0.001, math.MaxFloat64)},
	"--max-rate":              {argRequired, validateFloatRange(0.001, math.MaxFloat64)},
	"--defeat-rst-ratelimit":  {argNone, nil},
	"--defeat-icmp-ratelimit": {argNone, nil},
	"--nsock-engine":          {argRequired, validateAny},

	// Evasion and spoofing.
	"--mtu":         {argRequired, validateMtu},
	"--data-length": {argRequired, validateIntRange(0, 65435)},
	"--ip-options":  {argRequired, validateAny},
	"--ttl":         {argRequired, validateIntRange(0, 255)},
	"--spoof-mac":   {argRequired, validateMac},
	"--proxies":     {argRequired, validateAny},
	"--proxy":       {argRequired, validateAny}, // Singular alias for --proxies.
	"--data":        {argRequired, validateAny},
	"--data-string": {argRequired, validateAny},
	"--source-port": {argRequired, validateSinglePort},
	"--badsum":      {argNone, nil},
	"--adler32":     {argNone, nil},

	// Output.
	"--webxml":                 {argNone, nil},
	"--no-stylesheet":          {argNone, nil},
	"--stylesheet":             {argRequired, validateAny},
	"--append-output":          {argNone, nil},
	"--open":                   {argNone, nil},
	"--packet-trace":           {argNone, nil},
	"--iflist":                 {argNone, nil},
	"--log-errors":             {argNone, nil},
	"--stats-every":            {argRequired, validateTime},
	"--reason":                 {argNone, nil},
	"--vv":                     {argNone, nil}, // Alias for double verbosity.
	"--ff":                     {argNone, nil}, // Alias for double fragmentation.
	"--deprecated-xml-osclass": {argNone, nil},

	// Miscellaneous.
	"--resume":         {argRequired, validateAny},
	"--noninteractive": {argNone, nil},
	"--datadir":        {argRequired, validateAny},
	"--servicedb":      {argRequired, validateAny},
	"--versiondb":      {argRequired, validateAny},
	"--send-eth":       {argNone, nil},
	"--send-ip":        {argNone, nil},
	"--privileged":     {argNone, nil},
	"--unprivileged":   {argNone, nil},
	"--release-memory": {argNone, nil},
	"--route-dst":      {argRequired, validateAny},
}

// shortFlags maps every supported short flag (-X or -XX) to its definition.
var shortFlags = map[string]flagDef{

	// Target specification.
	"-iL": {argRequired, validateAny},
	"-iR": {argRequired, validateIntRange(0, math.MaxInt)},

	// Host discovery.
	"-sL": {argNone, nil},
	"-sn": {argNone, nil},
	"-sP": {argNone, nil},
	"-Pn": {argNone, nil},
	"-PN": {argNone, nil}, // Legacy alias for -Pn, still accepted by Nmap.
	"-P0": {argNone, nil}, // Legacy alias for -Pn, still accepted by Nmap.
	"-PS": {argOptional, validatePortList},
	"-PA": {argOptional, validatePortList},
	"-PU": {argOptional, validatePortList},
	"-PY": {argOptional, validatePortList},
	"-PE": {argNone, nil},
	"-PP": {argNone, nil},
	"-PM": {argNone, nil},
	"-PO": {argOptional, validateAny},
	"-PR": {argNone, nil},
	"-PB": {argNone, nil},                  // Deprecated combined ping type.
	"-PI": {argNone, nil},                  // Deprecated alias for -PE.
	"-PD": {argNone, nil},                  // Deprecated alias for -Pn.
	"-PT": {argOptional, validatePortList}, // Deprecated alias for -PA.
	"-n":  {argNone, nil},
	"-R":  {argNone, nil},
	"-4":  {argNone, nil},
	"-6":  {argNone, nil},

	// Scan techniques.
	"-sS": {argNone, nil},
	"-sT": {argNone, nil},
	"-sU": {argNone, nil},
	"-sA": {argNone, nil},
	"-sW": {argNone, nil},
	"-sM": {argNone, nil},
	"-sN": {argNone, nil},
	"-sF": {argNone, nil},
	"-sX": {argNone, nil},
	"-sI": {argRequired, validateZombieHost},
	"-sY": {argNone, nil},
	"-sZ": {argNone, nil},
	"-sO": {argNone, nil},
	"-sR": {argNone, nil}, // Deprecated RPC scan, now alias for -sV.
	"-sC": {argNone, nil},
	"-sV": {argNone, nil},
	"-b":  {argRequired, validateAny},

	// Port specification.
	"-p": {argRequired, validatePortList},
	"-F": {argNone, nil},
	"-r": {argNone, nil},

	// OS detection and aggression.
	"-O": {argNone, nil},
	"-A": {argNone, nil},

	// Timing.
	"-T": {argRequired, validateTimingTemplate},

	// Evasion and spoofing.
	"-f": {argOptional, validateFragmentRepeat},
	"-D": {argRequired, validateDecoys},
	"-S": {argRequired, validateAny},
	"-e": {argRequired, validateAny},
	"-g": {argRequired, validateSinglePort},

	// Output.
	"-oN": {argRequired, validateAny},
	"-oX": {argRequired, validateAny},
	"-oG": {argRequired, validateAny},
	"-oA": {argRequired, validateAny},
	"-oS": {argRequired, validateAny},
	"-v":  {argOptional, validateVerbosityLevel},
	"-d":  {argOptional, validateDebugLevel},

	// Miscellaneous.
	"-h": {argNone, nil},
	"-V": {argNone, nil},
	"-I": {argNone, nil},                                  // Deprecated ident scan.
	"-M": {argRequired, validateIntRange(0, math.MaxInt)}, // Alias for --max-parallelism.
	"-m": {argRequired, validateAny},                      // Deprecated alias for -oG.
	"-i": {argRequired, validateAny},                      // Deprecated alias for -iL.
}

// tcpPortScanFlags is the set of mutually exclusive TCP scan-type flags.
var tcpPortScanFlags = map[string]bool{
	"-sS": true, "-sT": true, "-sA": true, "-sW": true,
	"-sM": true, "-sN": true, "-sF": true, "-sX": true,
	"-sI": true,
}

// noPortScanFlags is the set of flags that disable the port-scan phase entirely.
var noPortScanFlags = map[string]bool{
	"-sn": true, "-sP": true, "-sL": true,
}
