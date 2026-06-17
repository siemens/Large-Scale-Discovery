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

// BlockFlagsInput contains flags that read targets from the filesystem, allowing an
// attacker to feed arbitrary hosts into the scan or probe files on the host.
var BlockFlagsInput = []string{"-iL", "-iR", "-i", "--resume"}

// BlockFlagsDatabase contains flags that override Nmap's built-in fingerprint databases,
// allowing tampered service or version data to be loaded from attacker-controlled paths.
var BlockFlagsDatabase = []string{"--datadir", "--servicedb", "--versiondb"}

// BlockFlagsOutput contains flags that write scan results to arbitrary filesystem paths,
// potentially overwriting files or leaking data outside the expected output channel.
var BlockFlagsOutput = []string{"-oN", "-oX", "-oG", "-oA", "-oS", "-m"}

// BlockFlagsScript contains flags that invoke or configure the Nmap Scripting Engine (NSE),
// which can execute arbitrary Lua code or load external script files.
var BlockFlagsScript = []string{"--script", "--script-args", "--script-args-file", "--script-updatedb"}

// BlockFlagsSpoof contains flags that forge source addresses, inject decoy traffic,
// bind to arbitrary interfaces, or spoof MAC addresses.
var BlockFlagsSpoof = []string{"-S", "-D", "-e", "--spoof-mac"}

// BlockFlags is a consolidated list of all the grouped ones from above.
var BlockFlags = func(blockLists ...[]string) []string {
	var blockFlags []string
	for _, blockFlag := range blockLists {
		blockFlags = append(blockFlags, blockFlag...)
	}
	return blockFlags
}(
	BlockFlagsInput,
	BlockFlagsDatabase,
	BlockFlagsOutput,
	BlockFlagsScript,
	BlockFlagsSpoof,
)

// options holds the resolved settings for a single Validate call.
type options struct {
	blockedFlags []string
}

// OptionFunc is a functional option that configures a single aspect of the validation.
type OptionFunc func(*options)

// WithBlockFlags returns an optionFunc that adds the given flags to the blocked set.
// May be called multiple times to combine several flag lists. Duplicates are ignored.
// Use DefaultBlockedFlags for a curated set suitable for managed scanning environments.
func WithBlockFlags(flags []string) OptionFunc {
	return func(c *options) {

		// Build a lookup of already-blocked flags to skip duplicates.
		seen := make(map[string]struct{}, len(c.blockedFlags))
		for _, f := range c.blockedFlags {
			seen[f] = struct{}{}
		}

		// Append only flags that are not already in the list.
		for _, f := range flags {
			if _, exists := seen[f]; !exists {
				c.blockedFlags = append(c.blockedFlags, f)
				seen[f] = struct{}{}
			}
		}
	}
}
