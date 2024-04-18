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
	"errors"
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/manager/config"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"gorm.io/gorm"
	"time"
)

var ErrInvalidPrivilege = fmt.Errorf("invalid privilege secret")

//
// ATTENTION: This RPC endpoint does return scan scope data INCLUDING SENSITIVE attributes, such as, scope secret and
// 			  database credentials. This RPC endpoint may only be called with a valid privilege token.
//

// GetScopeFull returns a scan scope, identified by its scope secret, with FULL details (secrets, credentials,...)
// to an RPC client. The RPC client has to provide a valid privilege secret, shared between the client and the manager.
// ATTENTION: If the supplied scope secret is invalid (unknown), an empty scan scope is returned. The client has to
//
//	check whether the returned scan scope's ID is != 0. No error is returned, because it would trigger a
//	critical log. End user configuration errors or scan agents left behind, should not flood with critical
//	log messages.
func (s *Manager) GetScopeFull(rpcArgs *ArgsScopeFull, rpcReply *ReplyScanScope) error {

	// Generate UUID for context
	uuid := shortuuid.New()[0:10] // Shorten uuid, doesn't need to be that long

	// Get tagged logger
	logger := log.GetLogger().Tagged(fmt.Sprintf("%s-GetScopeFull", uuid))

	// Log RPC call benchmark to be able to identify bottlenecks later on
	start := time.Now()
	defer func() {
		logger.Debugf("RPC call took %s.", time.Since(start))
	}()

	// Get config
	conf := config.GetConfig()

	// Check whether client is allowed to request full scope details
	if !scanUtils.StrContained(rpcArgs.PrivilegeSecret, conf.PrivilegeSecrets) {
		logger.Warningf("Client requested sensitive scope details with invalid privilege token!")
		return ErrInvalidPrivilege // Error message returned to agent
	}

	// Get scan scope for given secret
	scopeEntry, errScopeEntry := database.GetScopeEntryBySecret(rpcArgs.ScopeSecret)
	if errors.Is(errScopeEntry, gorm.ErrRecordNotFound) { // Check if entry didn't exist
		logger.Infof("Could not query scan scope with secret '%s...'.", rpcArgs.ScopeSecret[0:5])
		return nil // No error, the client just didn't have a valid scope secret
	}
	if errScopeEntry != nil {
		return fmt.Errorf("could not query scan scope: %s", errScopeEntry)
	}

	// Log action
	logger.Infof(
		"Client requesting *sensitive* scope details for scan scope '%s' (ID %d).", scopeEntry.Name, scopeEntry.Id)

	// Copy scan scope data into RPC response
	rpcReply.ScanScope = *scopeEntry

	// Log completion
	logger.Debugf("Scope returned with full details.")

	// Return nil to indicate successful RPC call
	return nil
}
