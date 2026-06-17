/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
)

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
		conf.Paths.Python,
		sslyzeAdditionalTruststore,
		target,
		port,
		vhosts,
	)
}
