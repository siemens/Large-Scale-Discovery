/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/discovery/ot"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
	broker "github.com/siemens/Large-Scale-Discovery/broker/core"
)

// otListenerOnce ensures the OT listener is initialized only once
var otListenerOnce sync.Once

// launchDiscoveryOt performs an OT discovery: the agent self-detects its network segment(s) and runs
// Nmap (with OT-aware settings already injected by the broker) plus the OT discovery protocols
// (PROFINET DCP, EtherCAT, LLDP, NDP, mDNS) on each detected subnet. The passive OT listener is
// started on the first invocation and remains alive for the agent lifetime so that passively captured
// hosts accumulate across scan iterations. The result merges hosts from Nmap, active OT protocols,
// and the passive listener cache. The detected subnets are reported back so the broker can populate
// t_discovery.Input for display.
func launchDiscoveryOt(
	logger scanUtils.Logger,
	label string,
	nmapArgs []string,
	networkTimeout time.Duration,
	excludeDomains []string,
	conf *config.Agent,
	rpcResult broker.ArgsSaveScanResult,
) broker.ArgsSaveScanResult {

	// Prepare memory for scan results
	var subnetsScanned []string
	var result = &discovery.Result{
		Data:      nil,
		Status:    scanUtils.StatusCompleted,
		Exception: false,
	}

	// Determine interfaces to scan: configured list (if non-empty) or all non-loopback IPv4 ones.
	targetInterfaces := getInterfaces(conf.OtInterfaces)

	// Start the passive OT listener on the first OT scan. It is pinned to the agent-level context
	// so it survives across all subsequent scan iterations without ever restarting.
	// Passive hosts are cached by the listener inside GoScans and drained into scan OT scan results automatically.
	otListenerOnce.Do(func() {

		// Prepare list of interface names
		ifaces := make([]string, 0, len(targetInterfaces))
		for _, iface := range targetInterfaces {
			ifaces = append(ifaces, iface.Name)
		}

		// Start passive listening for OT devices which will continue running from now on
		errListener := ot.StartPassiveListener(coreCtx, logger, ifaces...)
		if errListener != nil {
			logger.Warningf("Could not start OT passive listeners: %s", errListener)
		}
	})

	// Run discovery for each detected interface/subnet, merging all hosts into one result.
	for _, targetInterface := range targetInterfaces {

		// Prepare scan input classified by IPv4/v6
		var targetV4 = []string{targetInterface.Cidr} // Currently the OT discovery is only supported for IPv4
		var targetV6 []string                         // Currently the OT discovery is only supported for IPv4

		// Execute OT scan iteration
		logger.Infof("Executing OT scan on interface '%s' with subnet '%s'.", targetInterface.Name, targetInterface.Cidr)
		resultOtScan, errResultOtScan := executeDiscoveryScan(
			logger,
			label,
			targetV4,
			targetV6,
			conf,
			networkTimeout,
			nmapArgs,
			excludeDomains,
			targetInterface.Name, // Triggers OT protocols (PROFINET DCP, EtherCAT, LLDP, NDP, mDNS) on this interface
		)
		if errResultOtScan != nil {

			// Log issue
			logger.Errorf("%s scan initialization failed: %s", label, errResultOtScan)

			// Forward issue to broker
			rpcResult.Result = &discovery.Result{
				Exception: true,
				Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errResultOtScan.Error()),
			}

			// Shutdown agent as this is a critical issue
			defer Shutdown()

			// Return
			return rpcResult
		}

		// Scan did neither return a result struct nor an error, so the execution must have run into context shutdown.
		// A nil result struct is only returned in case of context shutdown.
		if resultOtScan == nil {
			logger.Debugf("Returning without results after OT scan due to agent shutdown.")
			return rpcResult
		}

		// Add scanned subnet to feedback data
		subnetsScanned = append(subnetsScanned, targetInterface.Cidr)

		// Merge scan result of this OT scan iteration to the result data of the other interfaces and subnets
		result.Data = append(result.Data, resultOtScan.Data...)
	}

	// Feed back discovered and used input targets
	rpcResult.ResultSubnets = subnetsScanned

	// Update result data
	rpcResult.Result = result

	// Return updated RPC args
	return rpcResult
}

func launchDiscoveryIt(
	logger scanUtils.Logger,
	label string,
	scanTarget string,
	nmapArgs []string,
	nmapArgsPreScan []string,
	networkTimeout time.Duration,
	excludeDomains []string,
	conf *config.Agent,
	rpcResult broker.ArgsSaveScanResult,
) broker.ArgsSaveScanResult {

	// Classify whether target is reachable via IPv4 and/or IPv6
	hasV4Ips, hasV6Ips := classifyInput(scanTarget)

	// Skip scan if to scan target could be resolved
	if !hasV4Ips && !hasV6Ips {
		logger.Debugf("Target could not be resolved.")
		rpcResult.Result = &discovery.Result{
			Data:      nil,
			Status:    scanUtils.StatusNotReachable,
			Exception: false,
		}
		return rpcResult
	}

	// Prepare scan input classified by IPv4/v6
	var targetV4 = make([]string, 0, 1)
	var targetV6 = make([]string, 0, 1)
	if hasV4Ips {
		targetV4 = append(targetV4, scanTarget)
	}
	if hasV6Ips {
		targetV6 = append(targetV6, scanTarget)
	}

	// Execute pre-scan with minimum settings trying to discover some likely data where bigger scans might fail (e.g.
	// due to IPS systems going on or unexpected timeouts), in order to register some more hosts that might remain
	// undiscovered otherwise.
	var resultPreScan *discovery.Result
	if len(nmapArgsPreScan) > 0 {
		logger.Debugf("Executing pre-scan.")
		var errResultPreScan error
		resultPreScan, errResultPreScan = executeDiscoveryScan(
			logger,
			label,
			targetV4,
			targetV6,
			conf,
			networkTimeout,
			nmapArgsPreScan,
			excludeDomains,
			"",
		)
		if errResultPreScan != nil {

			// Log issue
			logger.Errorf("%s scan initialization failed: %s", label, errResultPreScan)

			// Forward issue to broker
			rpcResult.Result = &discovery.Result{
				Exception: true,
				Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errResultPreScan.Error()),
			}

			// Shutdown agent as this is a critical issue
			defer Shutdown()

			// Return
			return rpcResult
		}

		// Prescan did neither return a result struct nor an error, so the execution must have run into context shutdown.
		// A nil result struct is only returned in case of context shutdown.
		if resultPreScan == nil {
			logger.Debugf("Returning without results after pre-scan due to agent shutdown.")
			return rpcResult
		}
	}

	// Execute actual scan of the target
	logger.Debugf("Executing main scan.")
	result, errResult := executeDiscoveryScan(
		logger,
		label,
		targetV4,
		targetV6,
		conf,
		networkTimeout,
		nmapArgs,
		excludeDomains,
		"",
	)
	if errResult != nil {

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errResult)

		// Forward issue to broker
		rpcResult.Result = &discovery.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errResult.Error()),
		}

		// Shutdown agent as this is a critical issue
		defer Shutdown()

		// Return
		return rpcResult
	}

	// Scan did neither return a result struct nor an error, so the execution must have run into context shutdown.
	// A nil result struct is only returned in case of context shutdown.
	if result == nil {
		logger.Debugf("Returning without results after main scan due to agent shutdown.")
		return rpcResult
	}

	// Prepare map of IPs where the actual port scan succeeded (delivered something)
	hostsDiscovered := make(map[string]struct{}, len(result.Data))
	for _, host := range result.Data {
		hostsDiscovered[host.Ip] = struct{}{}
	}

	// Find hosts where the actual scan didn't deliver results, but the pre-scan did. Add those pre-scan results to
	// the actual scan results.
	// ATTENTION: If the actual scan ran into its deadline it didn't scan any results at all. Pre-scan results will be
	// 			  injected as fallback.
	if resultPreScan != nil {
		for _, hostPreScan := range resultPreScan.Data {
			if _, hostDiscovered := hostsDiscovered[hostPreScan.Ip]; !hostDiscovered {
				logger.Infof("Injecting pre-scan result of '%s'.", hostPreScan.Ip)
				result.Data = append(result.Data, hostPreScan)
			}
		}
	}

	// Update result template with actual results
	rpcResult.Result = result

	// Return updated RPC args
	return rpcResult
}
