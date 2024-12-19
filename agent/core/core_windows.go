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
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	broker "github.com/siemens/Large-Scale-Discovery/broker/core"
)

// initModules sets the scan modules that can be run under the current OS and depending on the configuration
func initModules(scanScopeSecrets []string) {
	InitScopeCounters(scanScopeSecrets, []string{
		discovery.Label,
		banner.Label,
		nfs.Label,
		smb.Label,
		ssh.Label,
		ssl.Label,
		webcrawler.Label,
		webenum.Label,
	})
}

// launch launches a task for a given scan module if there is an implementation for this OS
func launch(logger scanUtils.Logger, chOut chan broker.ArgsSaveScanResult, scanTask *broker.ScanTask) {

	// Execute scan based on module type
	switch scanTask.Label {
	case discovery.Label:
		go launchDiscovery(chOut, scanTask)
	case banner.Label:
		go launchBanner(chOut, scanTask)
	case nfs.Label:
		go launchNfs(chOut, scanTask)
	case smb.Label:
		go launchSmb(chOut, scanTask)
	case ssl.Label:
		go launchSsl(chOut, scanTask)
	case ssh.Label:
		go launchSsh(chOut, scanTask)
	case webcrawler.Label:
		go launchWebcrawler(chOut, scanTask)
	case webenum.Label:
		go launchWebenum(chOut, scanTask)
	default:
		logger.Warningf("Invalid scan module '%s'.", scanTask.Label)
		DecrementModuleCount(scanTask.Secret, scanTask.Label)
	} // Switch End
}
