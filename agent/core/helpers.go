/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2022.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"large-scale-discovery/agent/config"
	"net"
	"strings"
	"time"
)

// resolveInput checks whether a scan input is an IPv4, IPv6 address/range or a hostname. Hostnames are resolved
// into IPv4 and IPv6 addresses, since there might be multiple of each. The method does not handle errors.
func resolveInput(target string) (ipsV4 []string, ipsV6 []string) {

	// Check if input is an IP address and classify it as IPv4 or IPv6
	addr := net.ParseIP(target)
	if addr != nil {
		if strings.Count(target, ":") >= 2 {
			return nil, []string{target}
		} else {
			return []string{target}, nil
		}
	}

	// Check if input is a network range and classify it as IPv4 or IPv6
	_, _, errCIDR := net.ParseCIDR(target)
	if errCIDR == nil {
		if strings.Count(target, ":") >= 2 {
			return nil, []string{target}
		} else {
			return []string{target}, nil
		}
	}

	// Check if input is a hostname and resolve it to IPv4 or IPv6 addresses
	ips, errHostname := net.LookupIP(target)
	if errHostname == nil {
		for _, ip := range ips {
			if strings.Count(ip.String(), ":") >= 2 {
				ipsV6 = append(ipsV6, ip.String())
			} else {
				ipsV4 = append(ipsV4, ip.String())
			}
		}
		// Return resolved IP addresses grouped by v4/v6
		return ipsV4, ipsV6
	}

	// Input could not be parsed to any kind of address
	return nil, nil
}

// executeDiscoveryScan starts Nmap scans based on the input target. Each family of IPs (v4 or v6) gets scanned.
func executeDiscoveryScan(
	logger scanUtils.Logger,
	label string,
	ipsV4 []string,
	ipsV6 []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
	nmapArgs []string,
) (result *discovery.Result, errScan error) {

	// Execute necessary scans, execute two scans if IPv4 and IPv6 addresses are mixed
	if len(ipsV4) > 0 && len(ipsV6) == 0 {
		result, errScan = executeNmapScan(logger, label, ipsV4, nmapArgs, conf, networkTimeout)
	} else if len(ipsV6) > 0 && len(ipsV4) == 0 {
		nmapArgs = append(nmapArgs, "-6")
		result, errScan = executeNmapScan(logger, label, ipsV6, nmapArgs, conf, networkTimeout)
	} else if len(ipsV4) > 0 && len(ipsV6) > 0 { // Execute two scans if hostname has IPv4 and IPv6 IPs

		// Execute IPv4 scan
		result, errScan = executeNmapScan(logger, label, ipsV4, nmapArgs, conf, networkTimeout)

		// Execute IPv6 scan
		nmapArgs = append(nmapArgs, "-6")
		resultIpV6, _ := executeNmapScan(logger, label, ipsV6, nmapArgs, conf, networkTimeout)

		// Save results of the two scans
		result.Data = append(result.Data, resultIpV6.Data...)

	}

	// Return scan results or error
	return result, errScan
}

func executeNmapScan(
	logger scanUtils.Logger,
	label string,
	scanTarget []string,
	nmapArgs []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
) (*discovery.Result, error) {

	// Initialize scanner
	scanner, scannerErr := discovery.NewScanner(
		logger,
		scanTarget,
		conf.Paths.Nmap,
		nmapArgs,
		false,
		[]string{instanceIp}, // Exclude local IP from scans, scan would have extended privileges discovering content that isn't visible from the outside
		conf.Modules.Discovery.BlacklistFile,
		conf.Modules.Discovery.DomainOrder,
		conf.Modules.Discovery.LdapServer,
		conf.Authentication.Ldap.Domain,
		conf.Authentication.Ldap.User,
		conf.Authentication.Ldap.Password,
		networkTimeout,
	)
	if scannerErr != nil {
		logger.Warningf("%s scanner initialization failed: %s", label, scannerErr)
		result := &discovery.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scanner initialization failed: %s", label, scannerErr.Error()),
		}
		return result, scannerErr
	}

	// Execute the scan.
	scanResult := scanner.Run()

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
