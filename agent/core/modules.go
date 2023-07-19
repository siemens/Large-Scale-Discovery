/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/filecrawler"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"large-scale-discovery/agent/config"
	broker "large-scale-discovery/broker/core"
	"large-scale-discovery/log"
	"strings"
	"time"
)

const SslOsTruststoreFile = "./data/os_truststore.pem"
const WebenumProbesFile = "./data/webenum_probes.txt"

func launchBanner(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := banner.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)

	// Initiate scanner
	scan, errScan := banner.NewScanner(
		logger,
		scanTask.Target,
		scanTask.Port,
		scanTask.Protocol,
		networkTimeout,
		networkTimeout,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &banner.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs
		return
	}

	// Execute the scan
	result := scan.Run()

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchDiscovery(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := discovery.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	var nmapArgsPreScan []string
	var excludeDomains []string
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	nmapArgs := strings.Split(scanTask.ScanSettings.DiscoveryNmapArgs, " ") // Nmap arguments for the main port scan scan
	if len(scanTask.ScanSettings.DiscoveryNmapArgsPrescan) > 0 {
		nmapArgsPreScan = strings.Split(scanTask.ScanSettings.DiscoveryNmapArgsPrescan, " ") // Nmap arguments for a smaller pre-scan intended to discover some scan results before an IDS might block
	}
	if len(scanTask.ScanSettings.DiscoveryExcludeDomains) > 0 {
		excludeDomains = strings.Split(strings.Replace(scanTask.ScanSettings.DiscoveryExcludeDomains, " ", "", -1), ",")
	}

	// Classify whether target is reachable via IPv4 and/or IPv6
	hasV4Ips, hasV6Ips := classifyInput(scanTask.Target)

	// Skip scan if to scan target could be resolved
	if !hasV4Ips && !hasV6Ips {
		logger.Debugf("Target could not be resolved.")
		rpcArgs.Result = &discovery.Result{
			Data:      nil,
			Status:    utils.StatusNotReachable,
			Exception: false,
		}
		chResults <- rpcArgs
		return
	}

	// Prepare scan input classified by IPv4/v6
	targetV4 := make([]string, 0, 1)
	targetV6 := make([]string, 0, 1)
	if hasV4Ips {
		targetV4 = append(targetV4, scanTask.Target)
	}
	if hasV6Ips {
		targetV6 = append(targetV6, scanTask.Target)
	}

	// Execute pre-scan with minimum settings trying to discover some likely data where bigger scans might fail (e.g.
	// due to IPS systems going on or unexpected timeouts), in order to register some more hosts that might remain
	// undiscovered otherwise.
	var resultPreScan *discovery.Result
	if len(nmapArgsPreScan) > 0 {
		logger.Debugf("Executing pre-scan.")
		var errPreScan error
		resultPreScan, errPreScan = executeDiscoveryScan(logger, label, targetV4, targetV6, conf, networkTimeout, nmapArgsPreScan, excludeDomains)
		if resultPreScan == nil {
			return
		}
		if errPreScan != nil {
			rpcArgs.Result = resultPreScan
			chResults <- rpcArgs
			return
		}
	}

	// Execute actual scan of the target
	logger.Debugf("Executing main scan.")
	result, errScan := executeDiscoveryScan(logger, label, targetV4, targetV6, conf, networkTimeout, nmapArgs, excludeDomains)
	if result == nil {
		return
	}
	if errScan != nil {
		rpcArgs.Result = result
		chResults <- rpcArgs
		return
	}

	// Prepare map of IPs where the actual port scan succeeded (delivered something)
	hostsDiscovered := make(map[string]struct{})
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
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchNfs(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := nfs.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.NfsScanTimeoutMinutes)
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	excludedShares := strings.Split(scanTask.ScanSettings.NfsExcludeShares, ",")
	excludedFolders := strings.Split(scanTask.ScanSettings.NfsExcludeFolders, ",")
	excludedExtensions := strings.Split(scanTask.ScanSettings.NfsExcludeExtensions, ",")

	// Initiate scanner
	scan, errScan := nfs.NewScanner(
		logger,
		scanTask.Target,
		scanTask.ScanSettings.NfsDepth,
		scanTask.ScanSettings.NfsThreads,
		excludedShares,
		excludedFolders,
		excludedExtensions,
		scanTask.ScanSettings.NfsExcludeLastModifiedBelow,
		scanTask.ScanSettings.NfsExcludeFileSizeBelow,
		scanTask.ScanSettings.NfsAccessibleOnly,
		networkTimeout,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &nfs.Result{
			Result: filecrawler.Result{
				Exception: true,
				Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
			},
		}
		chResults <- rpcArgs
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchSsh(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := ssh.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.SshScanTimeoutMinutes)

	// Initiate scanner
	scan, errScan := ssh.NewScanner(
		logger,
		scanTask.Target,
		scanTask.Port,
		networkTimeout,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &ssh.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchSsl(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := ssl.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.SslScanTimeoutMinutes)

	// Decide trust store file to use (OS-generated or custom one)
	var sslyzeAdditionalTruststore string
	if len(conf.Modules.Ssl.CustomTruststoreFile) == 0 {
		sslyzeAdditionalTruststore = SslOsTruststoreFile
	} else {
		sslyzeAdditionalTruststore = conf.Modules.Ssl.CustomTruststoreFile
	}

	// Initiate scanner
	scan, errScan := newSslScanner(
		logger,
		sslyzeAdditionalTruststore,
		scanTask.Target,
		scanTask.Port,
		scanTask.OtherNames,
		conf,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &ssl.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchWebcrawler(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := webcrawler.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.WebcrawlerScanTimeoutMinutes)
	followTypes := strings.Split(strings.Replace(scanTask.ScanSettings.WebcrawlerFollowTypes, " ", "", -1), ",")

	// Determine whether to use http or https scheme.
	https := false
	if strings.Contains(scanTask.Service, "https") || strings.Contains(scanTask.Service, "ssl") {
		https = true
	}

	// Initiate scanner
	scan, errScan := webcrawler.NewScanner(
		logger,
		scanTask.Target,
		scanTask.Port,
		scanTask.OtherNames,
		https,
		scanTask.ScanSettings.WebcrawlerDepth,
		scanTask.ScanSettings.WebcrawlerMaxThreads,
		scanTask.ScanSettings.WebcrawlerFollowQueryStrings,
		scanTask.ScanSettings.WebcrawlerAlwaysStoreRoot,
		conf.Modules.Webcrawler.Download,
		conf.Modules.Webcrawler.DownloadPath,
		conf.Authentication.Webcrawler.Domain,
		conf.Authentication.Webcrawler.User,
		conf.Authentication.Webcrawler.Password,
		scanTask.ScanSettings.HttpUserAgent,
		"",
		networkTimeout,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &webcrawler.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs
		return
	}

	// Set webcrawler follow types
	errFollow := scan.SetFollowContentTypes(followTypes)
	if errFollow != nil {
		logger.Debugf("Could not set which content types to be followed: %s", errFollow)
	}

	// Set webcrawler download types
	errDownload := scan.SetDownloadContentTypes(conf.Modules.Webcrawler.DownloadTypes)
	if errDownload != nil {
		logger.Debugf("Could not set which content types to download: %s", errFollow)
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchWebenum(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := webenum.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer decreaseUsageModule(label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo: broker.AgentInfo{
			Name: instanceName,
			Host: instanceHostname,
			Ip:   instanceIp,
		},
		ScopeSecret: scopeSecret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.WebenumScanTimeoutMinutes)

	// Determine whether to use http or https scheme.
	https := false
	if strings.Contains(scanTask.Service, "https") || strings.Contains(scanTask.Service, "ssl") {
		https = true
	}

	// Initiate scanner
	scan, errScan := webenum.NewScanner(
		logger,
		scanTask.Target,
		scanTask.Port,
		scanTask.OtherNames,
		https,
		conf.Authentication.Webenum.Domain,
		conf.Authentication.Webenum.User,
		conf.Authentication.Webenum.Password,
		WebenumProbesFile,
		scanTask.ScanSettings.WebenumProbeRobots,
		scanTask.ScanSettings.HttpUserAgent,
		"",
		networkTimeout,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &webenum.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}
