/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
)

// nmapArgsAddSensitivePorts merges a comma separated list of sensitive ports into an existing Nmap argument slice.
// Merges into an existing exclude-ports flag or appends a new one if not yet specified.
func nmapArgsAddSensitivePorts(args []string, sensitivePorts string) []string {

	// Return if no sensitive ports to add
	if sensitivePorts == "" || len(args) == 0 {
		return args
	}

	// Remove potential whitespaces from comma separated ports string
	sensitivePorts = strings.ReplaceAll(sensitivePorts, " ", "")

	// Inject sensitive ports
	for i, arg := range args {

		// Look for exclude-ports flag
		if arg == "--exclude-ports" && i+1 < len(args) {

			// Merge existing ports with sensitive ports
			args[i+1] = args[i+1] + "," + sensitivePorts

			// Return updated Nmap Arguments
			return args
		}
	}

	// Append sensitive ports, no existing exclude-ports flag was found
	args = append(args, "--exclude-ports", sensitivePorts)

	// Return updated Nmap Arguments
	return args
}

// interfaceSubnet describes a single network interface together with the IPv4 CIDR it belongs to.
type interfaceSubnet struct {
	Name string // Interface name (e.g. "eth0")
	Cidr string // IPv4 CIDR derived from the interface's primary IPv4 address (e.g. "192.168.1.0/24")
}

// getInterfaces returns all non-loopback interfaces that are UP and have at least one IPv4
// address, together with the CIDR of that address. If filterNames is non-empty, only interfaces whose
// name is contained in the list are returned (if they qualify).
func getInterfaces(filterNames []string) []interfaceSubnet {
	var result []interfaceSubnet
	ifaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	filterSet := make(map[string]struct{}, len(filterNames))
	for _, n := range filterNames {
		if n != "" {
			filterSet[n] = struct{}{}
		}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(filterSet) > 0 {
			if _, ok := filterSet[iface.Name]; !ok {
				continue
			}
		}
		addrs, errAddr := iface.Addrs()
		if errAddr != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue // Skip IPv6 for now - OT discovery targets IPv4 networks
			}

			// Derive the network CIDR (zeroing the host bits)
			network := ip4.Mask(ipNet.Mask)
			ones, _ := ipNet.Mask.Size()
			cidr := fmt.Sprintf("%s/%d", network.String(), ones)
			result = append(result, interfaceSubnet{Name: iface.Name, Cidr: cidr})
			break // One CIDR per interface is enough
		}
	}
	return result
}

// classifyInput checks whether a scan input can be scanned via IPv4 and/or IPv6. The function does not handle errors.
func classifyInput(target string) (hasV4Ips bool, hasV6Ips bool) {

	// Check if input is an IP address and classify it as IPv4 or IPv6
	addr := net.ParseIP(target)
	if addr != nil {
		if strings.Count(target, ":") >= 2 {
			return false, true
		} else {
			return true, false
		}
	}

	// Check if input is a network range and classify it as IPv4 or IPv6
	_, _, errCIDR := net.ParseCIDR(target)
	if errCIDR == nil {
		if strings.Count(target, ":") >= 2 {
			return false, true
		} else {
			return true, false
		}
	}

	// Check if input is a hostname and resolve it to IPv4 or IPv6 addresses
	ips, errHostname := net.LookupIP(target)
	if errHostname == nil {

		// Iterate resolved IPs to check whether there are IPv4 and IPv6 ones
		for _, ip := range ips {
			if strings.Count(ip.String(), ":") >= 2 {
				hasV6Ips = true
			} else {
				hasV4Ips = true
			}
		}

		// Return resolved IP addresses grouped by v4/v6
		return hasV4Ips, hasV6Ips
	}

	// Input could not be parsed to any kind of address
	return false, false
}

// executeDiscoveryScan starts Nmap scans based on the input target. Each family of IPs (v4 or v6) gets scanned.
// It may ONLY return nil/nil to indicate context shutdown.
// In all other circumstances it MUST return a result struct, event if it is just an empty one!
func executeDiscoveryScan(
	logger scanUtils.Logger,
	label string,
	scanTargetsV4 []string,
	scanTargetsV6 []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
	nmapArgs []string,
	excludeDomains []string,
	otInterface string, // Network interface name to execute OT discovery on. Empty interface disables the OT discovery.
) (result *discovery.Result, errScan error) {

	// Execute necessary scans, execute two scans if IPv4 and IPv6 addresses are mixed
	if len(scanTargetsV4) > 0 && len(scanTargetsV6) == 0 {
		result, errScan = executeNmapScan(logger, label, scanTargetsV4, nmapArgs, conf, networkTimeout, excludeDomains, otInterface)
	} else if len(scanTargetsV6) > 0 && len(scanTargetsV4) == 0 {
		nmapArgs = append(nmapArgs, "-6")
		result, errScan = executeNmapScan(logger, label, scanTargetsV6, nmapArgs, conf, networkTimeout, excludeDomains, otInterface)
	} else if len(scanTargetsV4) > 0 && len(scanTargetsV6) > 0 { // Execute two scans if hostname has IPv4 and IPv6 IPs

		// Execute IPv4 scan
		result, errScan = executeNmapScan(logger, label, scanTargetsV4, nmapArgs, conf, networkTimeout, excludeDomains, otInterface)
		if errScan != nil {
			return result, errScan
		}

		// Execute IPv6 scan
		nmapArgs = append(nmapArgs, "-6")
		resultIpV6, _ := executeNmapScan(logger, label, scanTargetsV6, nmapArgs, conf, networkTimeout, excludeDomains, "")

		// Save results of the two scans
		if result != nil && resultIpV6 != nil {
			result.Data = append(result.Data, resultIpV6.Data...)
		}
	}

	// Return scan results or error
	return result, errScan
}

// executeNmapScan scan initializes the discovery scanner and runs it.
// It may ONLY return nil/nil to indicate context shutdown.
// In all other circumstances it MUST return a result struct, event if it is just an empty one!
func executeNmapScan(
	logger scanUtils.Logger,
	label string,
	scanTargets []string,
	nmapArgs []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
	excludeDomains []string,
	otInterface string, // Network interface name to execute OT discovery on. Empty interface disables the OT discovery.
) (*discovery.Result, error) {

	// Initialize scan
	scan, errScan := discovery.NewScanner(
		logger,
		scanTargets,
		conf.Paths.Nmap,
		nmapArgs,
		false,
		[]string{instanceInfo.Ip}, // Exclude local IP from scans, scan would have extended privileges discovering content that isn't visible from the outside
		conf.Modules.Discovery.BlacklistFile,
		conf.Modules.Discovery.DomainOrder,
		conf.Modules.Discovery.LdapServer,
		conf.Authentication.Ldap.Domain,
		conf.Authentication.Ldap.User,
		conf.Authentication.Ldap.Password,
		conf.Authentication.Ldap.DisableGssapi,
		excludeDomains,
		networkTimeout,
	)
	if errScan != nil {
		return &discovery.Result{}, errScan // ATTENTION: Must return result struct!
	}

	// Enable OT discovery (via PROFINET DCP, EtherCAT, LLDP, NDP, mDNS) and set required interface
	if otInterface != "" {
		errOt := scan.EnableOtScanner(otInterface)
		if errOt != nil {
			return &discovery.Result{}, errOt
		}
	}

	// Only launch if shutdown is not already in progress
	select {
	case <-coreCtx.Done():
		logger.Infof("Omitting discovery scan due to agent shutdown.")
		return nil, nil
	default:
	}

	// Execute the scan.
	scanResult := scan.Run(0)

	// Drop Nmap exception caused in association with termination signal (broken XML output), which might occur
	// due to a race condition. Nmap sub processes might terminate before the termination signal is fulfilled by
	// the scan agent itself. The scan agent might be able to process the broken Nmap output and return it as an
	// exception.
	if scanResult.Exception {
		select {
		case <-coreCtx.Done():
			logger.Debugf("Dropping broken discovery result, caused by agent shutdown.")
			return nil, nil
		default:
		}
	}

	// Return discovery scan result
	return scanResult, nil
}
