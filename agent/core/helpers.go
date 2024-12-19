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
	"github.com/siemens/GoScans/discovery"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
	"net"
	"strings"
	"time"
)

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
func executeDiscoveryScan(
	logger scanUtils.Logger,
	label string,
	scanTargetsV4 []string,
	scanTargetsV6 []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
	nmapArgs []string,
	excludeDomains []string,
) (result *discovery.Result, errScan error) {

	// Execute necessary scans, execute two scans if IPv4 and IPv6 addresses are mixed
	if len(scanTargetsV4) > 0 && len(scanTargetsV6) == 0 {
		result, errScan = executeNmapScan(logger, label, scanTargetsV4, nmapArgs, conf, networkTimeout, excludeDomains)
	} else if len(scanTargetsV6) > 0 && len(scanTargetsV4) == 0 {
		nmapArgs = append(nmapArgs, "-6")
		result, errScan = executeNmapScan(logger, label, scanTargetsV6, nmapArgs, conf, networkTimeout, excludeDomains)
	} else if len(scanTargetsV4) > 0 && len(scanTargetsV6) > 0 { // Execute two scans if hostname has IPv4 and IPv6 IPs

		// Execute IPv4 scan
		result, errScan = executeNmapScan(logger, label, scanTargetsV4, nmapArgs, conf, networkTimeout, excludeDomains)

		// Execute IPv6 scan
		nmapArgs = append(nmapArgs, "-6")
		resultIpV6, _ := executeNmapScan(logger, label, scanTargetsV6, nmapArgs, conf, networkTimeout, excludeDomains)

		// Save results of the two scans
		result.Data = append(result.Data, resultIpV6.Data...)
	}

	// Return scan results or error
	return result, errScan
}

func executeNmapScan(
	logger scanUtils.Logger,
	label string,
	scanTargets []string,
	nmapArgs []string,
	conf *config.AgentConfig,
	networkTimeout time.Duration,
	excludeDomains []string,
) (*discovery.Result, error) {

	// Initialize scanner
	scanner, scannerErr := discovery.NewScanner(
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
		excludeDomains,
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
