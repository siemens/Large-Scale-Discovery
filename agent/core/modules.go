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
	"strings"
	"time"

	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/filecrawler"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/nuclei"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
	broker "github.com/siemens/Large-Scale-Discovery/broker/core"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
)

const pathNucleiTemplatesFolder = "./data/nuclei_templates/"
const pathSslOsTruststoreFile = "./data/os_truststore.pem"
const pathWebenumProbesFile = "./data/webenum_probes.txt"

func launchBanner(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := banner.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
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

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &banner.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
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

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcResult := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
		Id:          scanTask.Id,
	}

	// Prepare pre-scan Nmap arguments
	var nmapArgs = strings.Split(scanTask.ScanSettings.DiscoveryNmapArgs, " ") // Nmap's arguments for the main port scan
	var nmapArgsPreScan []string
	if len(scanTask.ScanSettings.DiscoveryNmapArgsPrescan) > 0 { // Prepare Nmap arguments for a smaller pre-scan intended to discover some scan results before an IDS might block
		nmapArgsPreScan = strings.Split(scanTask.ScanSettings.DiscoveryNmapArgsPrescan, " ")
	}

	// Inject sensitive ports into existing Nmap args
	nmapArgs = nmapArgsAddSensitivePorts(nmapArgs, scanTask.ScanSettings.SensitivePorts)
	nmapArgsPreScan = nmapArgsAddSensitivePorts(nmapArgsPreScan, scanTask.ScanSettings.SensitivePorts)

	// Prepare network timeout
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)

	// Prepare exclude domains
	var excludeDomains []string
	if len(scanTask.ScanSettings.DiscoveryExcludeDomains) > 0 {
		excludeDomains = strings.Split(strings.Replace(scanTask.ScanSettings.DiscoveryExcludeDomains, " ", "", -1), ",")
	}

	// Decide whether to launch an OT scan or a normal nmap scan without special OT discovery
	if scanTask.ScanSettings.Ot {

		// When the scope has OT discovery set to true the agent always auto-detects its own subnet(s) and
		// runs Nmap + OT discovery protocols there, ignoring scanTask.Target. The Target field on later
		// cycles holds the previously detected CIDR (filled by the broker for display purposes), but the
		// agent still auto-detects from its interfaces every cycle.
		rpcResult = launchDiscoveryOt(logger, label, nmapArgs, networkTimeout, excludeDomains, conf, rpcResult)
	} else {
		rpcResult = launchDiscoveryIt(logger, label, scanTask.Target, nmapArgs, nmapArgsPreScan, networkTimeout, excludeDomains, conf, rpcResult)
	}

	// Forward the result
	chResults <- rpcResult
}

func launchNfs(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := nfs.Label

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-%d", label, scanTask.Id))
	logger.Debugf("Initializing scan.")

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.NfsScanTimeoutMinutes)
	networkTimeout := time.Second * time.Duration(scanTask.ScanSettings.NetworkTimeoutSeconds)
	excludedShares := utils.SanitizeToSlice(scanTask.ScanSettings.NfsExcludeShares, ",")
	excludedFolders := utils.SanitizeToSlice(scanTask.ScanSettings.NfsExcludeFolders, ",")
	excludedExtensions := utils.SanitizeToSlice(scanTask.ScanSettings.NfsExcludeExtensions, ",")

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

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &nfs.Result{
			Result: filecrawler.Result{
				Exception: true,
				Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
			},
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}

func launchNuclei(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := nuclei.Label

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
	defer DecrementModuleCount(scanTask.Secret, label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.NucleiScanTimeoutMinutes)

	// Pass nil for host based scans
	scanPort := func(p int) *int {
		if p == -1 {
			return nil
		}
		return &p
	}(scanTask.Port)

	// Append correct scheme for http(s) services
	target := scanTask.Target
	if scanTask.Service == "http" || scanTask.Service == "https" {
		target = scanTask.Service + "://" + target
	}

	// Initiate scanner
	scan, errScan := nuclei.NewScanner(
		logger,
		target,
		scanPort,
		pathNucleiTemplatesFolder,
		scanTask.ScanSettings.NucleiIncludeSeverities,
		scanTask.ScanSettings.NucleiExcludeSeverities,
		utils.SanitizeToSlice(scanTask.ScanSettings.NucleiIncludeTags, ","),
		utils.SanitizeToSlice(scanTask.ScanSettings.NucleiExcludeTags, ","),
		utils.SanitizeToSlice(scanTask.ScanSettings.NucleiIncludeIds, ","),
		utils.SanitizeToSlice(scanTask.ScanSettings.NucleiExcludeIds, ","),
		scanTask.ScanSettings.NucleiIncludeProtocols,
		scanTask.ScanSettings.NucleiExcludeProtocols,
		conf.Authentication.Nuclei.User,
		conf.Authentication.Nuclei.Password,
		"",
	)
	if errScan != nil {

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &nuclei.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
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

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
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

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &ssh.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
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

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
		Id:          scanTask.Id,
	}

	// Prepare variables
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.SslScanTimeoutMinutes)

	// Decide trust store file to use (OS-generated or custom one)
	var sslyzeAdditionalTruststore string
	if len(conf.Modules.Ssl.CustomTruststoreFile) == 0 {
		sslyzeAdditionalTruststore = pathSslOsTruststoreFile
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

		// Return suitable error or pass through original exit status
		switch {
		case strings.Contains(errScan.Error(), "exit status 0xc000013a"): // Exit code for ctrl+c on Windows
			logger.Debugf("%s scan initialization interrupted.", label)
			return
		case strings.Contains(errScan.Error(), "exit status 130"): // Exit code for ctrl+c on Linux
			logger.Debugf("%s scan initialization interrupted.", label)
			return
		// TODO: Add clauses for other known exit codes we might want to define closer.
		default:

			// Log issue
			logger.Errorf("%s scan initialization failed: %s", label, errScan)

			// Forward issue to broker
			rpcArgs.Result = &ssl.Result{
				Exception: true,
				Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
			}
			chResults <- rpcArgs

			// Shutdown agent as this is a critical issue
			Shutdown()

			// Return
			return
		}
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

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
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

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &webcrawler.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
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

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Decrease the module usage counter
	defer DecrementModuleCount(scanTask.Secret, label)

	// Get config
	conf := config.GetConfig()

	// Prepare result template that can be returned to the broker
	rpcArgs := broker.ArgsSaveScanResult{
		AgentInfo:   instanceInfo,
		ScopeSecret: scanTask.Secret,
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
		pathWebenumProbesFile,
		scanTask.ScanSettings.WebenumProbeRobots,
		scanTask.ScanSettings.HttpUserAgent,
		"",
		networkTimeout,
	)
	if errScan != nil {

		// Log issue
		logger.Errorf("%s scan initialization failed: %s", label, errScan)

		// Forward issue to broker
		rpcArgs.Result = &webenum.Result{
			Exception: true,
			Status:    fmt.Sprintf("%s scan initialization failed: %s", label, errScan.Error()),
		}
		chResults <- rpcArgs

		// Shutdown agent as this is a critical issue
		Shutdown()

		// Return
		return
	}

	// Execute the scan
	result := scan.Run(scanTimeout)

	// Update result template with actual results
	rpcArgs.Result = result

	// Forward the result
	chResults <- rpcArgs
}
