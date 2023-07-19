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
	"github.com/siemens/GoScans/filecrawler"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"large-scale-discovery/agent/config"
	broker "large-scale-discovery/broker/core"
	"large-scale-discovery/log"
	"strings"
	"time"
)

func launchSmb(
	chResults chan broker.ArgsSaveScanResult,
	scanTask *broker.ScanTask,
) {

	label := smb.Label

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
	scanTimeout := time.Minute * time.Duration(scanTask.ScanSettings.SmbScanTimeoutMinutes)
	excludedShares := strings.Split(scanTask.ScanSettings.SmbExcludeShares, ",")
	excludedFolders := strings.Split(scanTask.ScanSettings.SmbExcludeFolders, ",")
	excludedExtensions := strings.Split(scanTask.ScanSettings.SmbExcludeExtensions, ",")

	// Initiate scanner
	scan, errScan := smb.NewScanner(
		logger,
		scanTask.Target,
		scanTask.ScanSettings.SmbDepth,
		scanTask.ScanSettings.SmbThreads,
		excludedShares,
		excludedFolders,
		excludedExtensions,
		scanTask.ScanSettings.SmbExcludeLastModifiedBelow,
		scanTask.ScanSettings.SmbExcludeFileSizeBelow,
		scanTask.ScanSettings.SmbAccessibleOnly,
		conf.Authentication.Smb.Domain,
		conf.Authentication.Smb.User,
		conf.Authentication.Smb.Password,
	)
	if errScan != nil {
		logger.Warningf("%s scan initialization failed: %s", label, errScan)
		rpcArgs.Result = &smb.Result{
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

func newSslScanner(
	logger scanUtils.Logger,
	sslyzeAdditionalTruststore string, // Sslyze always applies default CAs, but you can add additional ones via custom trust store
	target string,
	port int,
	vhosts []string,
	conf *config.AgentConfig,
) (*ssl.Scanner, error) {
	return ssl.NewScanner(
		logger,
		conf.Paths.Sslyze,
		sslyzeAdditionalTruststore,
		target,
		port,
		vhosts,
	)
}
