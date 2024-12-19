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
	"context"
	"errors"
	"fmt"
	"github.com/sanyokbig/pqinterval"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"time"
)

/*
 * The RPC wrapper function defined in this file matches a respective manager RPC method. It can be used to call
 * the respective RPC method. These functions actually belong to the foreign components making use of them, calling
 * the respective manager RPC methods. They are put here to make them reusable for any foreign component. Multiple
 * foreign components might want to make use of same RPC functions to call the respective RPC methods. Hence,
 * these functions will be compiled into the other foreign components. The foreign component must pass its own RPC
 * client to execute the request.
 */

// RpcSubscribeNotification initializes a goroutine continuously listening for scope changes on the manager via RPC.
// This function can be used by other components to subscribe to notifications about scan scope changes. It returns
// two channels:
//   - A channel for notifications about specific scan scopes that changed
//   - A channel for re-connect notifications. In this case scan scope changes might have gone unobserved and a
//     complete integrity check is necessary!
func RpcSubscribeNotification(
	logger scanUtils.Logger,
	rpc *utils.Client,
	ctx context.Context, // The context by which cascaded goroutines shall terminate
) (chan ReplyNotification, chan struct{}) {

	// Prepare channels for notifying caller
	chNotification := make(chan ReplyNotification)
	chNotificationReconnect := make(chan struct{})

	// Only launch if shutdown is not already in progress
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}

	// Initialize background look querying listening for manager notifications
	go func() {

		// Log potential panics before letting them move on
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
				panic(r)
			}
		}()

		// Loop until shutdown
		for {

			// Log action
			logger.Debugf("Listening for manager notifications.")

			// Prepare RPC request
			rpcEndpoint := "Manager.SubscribeNotification"
			rpcReply := ReplyNotification{}
			rpcArgs := struct{}{} // Does not require arguments

			// Send RPC request. It will not return until scan scope changes become known.
			errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
			if errors.Is(errRpc, utils.ErrRpcConnectivity) {

				// Abort if termination is requested
				select {
				case <-ctx.Done(): // Cancellation signal

					// Close channels to inform about termination
					close(chNotification)
					close(chNotificationReconnect)

					// Terminate goroutine
					return
				default:

					// Log situation
					logger.Debugf("Waiting for re-connection.")

					// Wait until RPC connection or agent shutdown
					select {
					case <-ctx.Done(): // Cancellation signal

						// Close channels to inform about termination
						close(chNotification)
						close(chNotificationReconnect)

						// Terminate goroutine
						return
					case <-rpc.Established():

						// Monitoring subscriptions can continue again
						logger.Debugf("Manger re-connected.")

						// Pass on notification
						chNotificationReconnect <- struct{}{}

						// Subscribe again for next notification
						break
					}
				}
			} else if errRpc != nil && errRpc.Error() == utils.ErrNotifierShuttingDown.Error() { // Errors received from RPC lose their original type!!
				time.Sleep(time.Second / 2) // Prevent useless retries, seems like the manager is shutting down
			} else if errRpc != nil {
				logger.Warningf("Listening for manager notifications failed: %s", errRpc)
			} else {

				// Pass on notification
				chNotification <- rpcReply
			}
		}
	}()

	// Return active channels
	return chNotification, chNotificationReconnect
}

// RpcGetDatabases requests all database servers from the scope manager via RPC
func RpcGetDatabases(logger scanUtils.Logger, rpc *utils.Client) ([]database.T_db_server, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetDatabases"
	rpcReply := ReplyDatabases{}
	rpcArgs := struct{}{} // Does not require arguments

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return scan scopes
	return rpcReply.Databases, nil
}

// RpcAddUpdateDatabase creates a new or updates an existing database server in the manager's db
func RpcAddUpdateDatabase(
	logger scanUtils.Logger,
	rpc *utils.Client,
	dbServerId uint64,
	name string,
	dialect string,
	host string,
	hostPublic string,
	port int,
	admin string,
	password string,
	args string,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.AddUpdateDatabase"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsDatabaseDetails{
		DbServerId: dbServerId,
		Name:       name,
		Dialect:    dialect,
		Host:       host,
		HostPublic: hostPublic,
		Port:       port,
		Admin:      admin,
		Password:   password,
		Args:       args,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcRemoveDatabase requests the deletion of a database server from the scope manager via RPC
func RpcRemoveDatabase(logger scanUtils.Logger, rpc *utils.Client, dbServerId uint64) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.RemoveDatabase"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsDbServerId{DbServerId: dbServerId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcGetScopes requests all scan scopes from the scope manager via RPC
func RpcGetScopes(logger scanUtils.Logger, rpc *utils.Client) ([]database.T_scan_scope, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetScopes"
	rpcReply := ReplyScanScopes{}
	rpcArgs := struct{}{} // Does not require arguments

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return scan scopes
	return rpcReply.ScanScopes, nil
}

// RpcGetScopesOf requests the scan scopes of given groups from the scope manager via RPC
func RpcGetScopesOf(
	logger scanUtils.Logger, rpc *utils.Client, groupIds []uint64) ([]database.T_scan_scope, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetScopesOf"
	rpcReply := ReplyScanScopes{}
	rpcArgs := ArgsGroupIds{GroupIds: groupIds}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return scan scopes
	return rpcReply.ScanScopes, nil
}

// RpcGetScope requests a certain scan scopes from the scope manager via RPC
func RpcGetScope(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64) (database.T_scan_scope, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetScope"
	rpcReply := ReplyScanScope{}
	rpcArgs := ArgsScopeId{ScopeId: scopeId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return database.T_scan_scope{}, errRpc
	}

	// Return scan scopes
	return rpcReply.ScanScope, nil
}

// RpcCreateScope creates a new scope and its corresponding database in the manager's db
func RpcCreateScope(
	logger scanUtils.Logger,
	rpc *utils.Client,
	dbServerId uint64,
	name string,
	groupId uint64,
	createdBy string,
	cycles bool,
	cyclesRetention int,
	scopeType string,
	attributes utils.JsonMap,
) (uint64, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.CreateScope"
	rpcReply := ReplyScopeId{}
	rpcArgs := ArgsScopeDetails{
		DbServerId:      dbServerId,
		Name:            name,
		GroupId:         groupId,
		CreatedBy:       createdBy,
		Type:            scopeType,
		Cycles:          cycles,
		CyclesRetention: cyclesRetention,
		Attributes:      attributes,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return 0, errRpc
	}

	// Return nil as everything went fine
	return rpcReply.ScopeId, nil
}

// RpcDeleteScope requests the deletion of a scan scopes from the scope manager via RPC
func RpcDeleteScope(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.DeleteScope"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsScopeId{ScopeId: scopeId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcGetScopeTargets requests a list of current scan scope targets
func RpcGetScopeTargets(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64) (bool, []database.T_discovery, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetScopeTargets"
	rpcReply := ReplyScopeTargets{}
	rpcArgs := ArgsScopeId{ScopeId: scopeId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return false, nil, errRpc
	}

	// Return scan scopes
	return rpcReply.Synchronization, rpcReply.Targets, nil
}

// RpcToggleScope requests to enable/disable a scan scope by the scope manager via RPC
func RpcToggleScope(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.ToggleScope"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsScopeId{ScopeId: scopeId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcUpdateScope updates a certain scope in the manager's db
func RpcUpdateScope(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
	scopeName string,
	cycles bool,
	cyclesRetention int,
	attributes *utils.JsonMap, // Optional value, set to nil to skip update
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateScope"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsScopeUpdate{
		IdTScanScopes:   scopeId,
		Name:            scopeName,
		Cycles:          cycles,
		CyclesRetention: cyclesRetention,
		Attributes:      attributes,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcUpdateScopeTargets updates the input targets of a certain scope in the scopedb. This can only run once in
// parallel per scan scope.
func RpcUpdateScopeTargets(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
	targets []database.T_discovery, // Optional values, if there are values that should be set in the database right away
	blocking bool,
) (ReplyTargetsUpdate, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateScopeTargets"
	rpcReply := ReplyTargetsUpdate{}
	rpcArgs := ArgsTargetsUpdate{
		IdTScanScopes: scopeId,
		Targets:       targets,
		Blocking:      blocking,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return ReplyTargetsUpdate{}, errRpc
	}

	// Check if manager ran into error setting scope targets
	if rpcReply.Error != "" {
		return ReplyTargetsUpdate{}, fmt.Errorf(rpcReply.Error)
	}

	// Return scan scopes
	return rpcReply, nil
}

// RpcUpdateSettings updates a certain scope and its scan settings in the manager's db
func RpcUpdateSettings(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
	scanSettings database.T_scan_setting,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateSettings"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsSettingsUpdate{
		IdTScanScopes: scopeId,
		ScanSettings:  scanSettings,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcResetFailed resets failed scan targets to that they will be rescheduled for scanning within the current scan cycle
func RpcResetFailed(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.ResetFailed"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsScopeId{
		ScopeId: scopeId,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcNewCycle initializes a completely new scan cycle
func RpcNewCycle(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.NewCycle"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsScopeId{
		ScopeId: scopeId,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcResetInput requests the reset of a scan scope's input target from the scope manager via RPC
func RpcResetInput(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64, input string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.ResetInput"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsTargetReset{ScopeId: scopeId, Input: input}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcResetSecret requests the reset of a scan scope's secret from the scope manager via RPC
func RpcResetSecret(logger scanUtils.Logger, rpc *utils.Client, scopeId uint64) (string, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.ResetSecret"
	rpcReply := ReplyScopeSecret{}
	rpcArgs := ArgsScopeId{ScopeId: scopeId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return "", errRpc
	}

	// Return scan scopes
	return rpcReply.ScopeSecret, nil
}

// RpcUpdateAgents updates scan agent stats of a certain scope in the manager's db.
func RpcUpdateAgents(
	logger scanUtils.Logger,
	rpc *utils.Client,
	agentStatsByScope map[uint64][]database.T_scan_agent,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateAgents"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsStatsUpdate{
		ScanAgents: agentStatsByScope,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcGetViews requests all views from the scope manager via RPC
func RpcGetViews(logger scanUtils.Logger, rpc *utils.Client) ([]database.T_scope_view, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetViews"
	rpcReply := ReplyScopeViews{}
	rpcArgs := struct{}{} // Does not require arguments

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return scan scopes
	return rpcReply.ScopeViews, nil
}

// RpcGetViewsOf requests the views belonging to scopes of given groups from the scope manager via RPC
func RpcGetViewsOf(
	logger scanUtils.Logger, rpc *utils.Client, groupIds []uint64) ([]database.T_scope_view, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetViewsOf"
	rpcReply := ReplyScopeViews{}
	rpcArgs := ArgsGroupIds{GroupIds: groupIds}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return views
	return rpcReply.ScopeViews, nil
}

// RpcGetViewsGranted requests views with access rights for a certain user from the scope manager via RPC
func RpcGetViewsGranted(
	logger scanUtils.Logger, rpc *utils.Client, username string) ([]database.T_scope_view, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetViewsGranted"
	rpcReply := ReplyScopeViews{}
	rpcArgs := ArgsUsername{
		Username: username,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return nil as everything went fine
	return rpcReply.ScopeViews, nil
}

// RpcGetView requests the view of a given Id from the scope manager via RPC
func RpcGetView(logger scanUtils.Logger, rpc *utils.Client, viewId uint64) (database.T_scope_view, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetView"
	rpcReply := ReplyScopeView{}
	rpcArgs := ArgsViewId{ViewId: viewId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return database.T_scope_view{}, errRpc
	}

	// Return view
	return rpcReply.ScopeViews, nil
}

// RpcCreateView requests creation of view from the scope manager via RPC
func RpcCreateView(
	logger scanUtils.Logger,
	rpc *utils.Client,
	scopeId uint64,
	viewName string,
	createdBy string,
	filters map[string][]string,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.CreateView"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsViewDetails{
		ScopeId:   scopeId,
		ViewName:  viewName,
		CreatedBy: createdBy,
		Filters:   filters,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!

	// Return possible RPC error
	return errRpc
}

// RpcDeleteView requests deletion of view, including associated access rights, from the scope manager via RPC
func RpcDeleteView(logger scanUtils.Logger, rpc *utils.Client, viewId uint64) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.DeleteView"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsViewId{ViewId: viewId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!

	// Return possible RPC error
	return errRpc
}

// RpcUpdateView updates a certain view in the manager's db
func RpcUpdateView(logger scanUtils.Logger, rpc *utils.Client, viewId uint64, viewName string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateView"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsViewUpdate{
		ViewId: viewId,
		Name:   viewName,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

// RpcGrantToken requests to generate an access token and grant it access for a given view
func RpcGrantToken(
	logger scanUtils.Logger,
	rpc *utils.Client,
	viewId uint64,
	description string,
	cratedBy string,
	expiry time.Duration,
) (username string, password string, err error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GrantToken"
	rpcReply := ReplyCredentials{}
	rpcArgs := ArgsGrantToken{
		ViewId:      viewId,
		Description: description,
		CreatedBy:   cratedBy,
		Expiry:      expiry,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return "", "", errRpc
	}

	// Return scan scopes
	return rpcReply.Username, rpcReply.Password, nil
}

// RpcGrantUsers requests to grant a list of user access to a given view
func RpcGrantUsers(
	logger scanUtils.Logger,
	rpc *utils.Client,
	viewId uint64,
	dbUsers []database.DbCredentials,
	grantedBy string,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.GrantUsers"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsGrantUsers{
		ViewId:        viewId,
		DbCredentials: dbUsers,
		GrantedBy:     grantedBy,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcRevokeGrants requests the revocation of a single user access right on a view from the scope manager via RPC
func RpcRevokeGrants(logger scanUtils.Logger, rpc *utils.Client, viewId uint64, username ...string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.RevokeGrants"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsRevokeGrants{
		ViewId:    viewId,
		Usernames: username,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcUpdateDatabaseCredentials requests the update of the user's password on all database servers via RPC
func RpcUpdateDatabaseCredentials(logger scanUtils.Logger, rpc *utils.Client, username string, password string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.UpdateDatabaseCredentials"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsCredentials{
		Username: username,
		Password: password,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcDisableDbUser requests to disable a user on all database servers via RPC
func RpcDisableDbUser(logger scanUtils.Logger, rpc *utils.Client, username string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.DisableDbCredentials"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsUsername{
		Username: username,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcEnableDbUser requests to disable a user on all database servers via RPC
func RpcEnableDbUser(logger scanUtils.Logger, rpc *utils.Client, username string) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.EnableDbCredentials"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsUsername{
		Username: username,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return nil as everything went fine
	return nil
}

// RpcGetAgents requests all scan agents from the scope manager via RPC
func RpcGetAgents(logger scanUtils.Logger, rpc *utils.Client) ([]database.T_scan_agent, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetAgents"
	rpcReply := ReplyScanAgents{}
	rpcArgs := struct{}{} // Does not require arguments

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return nil, errRpc
	}

	// Return scan scopes
	return rpcReply.ScanAgents, nil
}

// RpcDeleteAgent requests the deletion of scan agent stats from the scope manager via RPC
func RpcDeleteAgent(logger scanUtils.Logger, rpc *utils.Client, agentId uint64) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.DeleteAgent"
	rpcReply := struct{}{} // Does not return values
	rpcArgs := ArgsAgentId{AgentId: agentId}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

func RpcCreateSqlLogs(
	logger scanUtils.Logger,
	rpc *utils.Client,
	dbName string,
	dbUser string,
	dbTable string,
	query string,
	queryResults int,
	queryTimestamp time.Time,
	queryDuration time.Duration,
	totalDuration time.Duration,
	clientName string,
) error {

	// Prepare RPC request
	rpcEndpoint := "Manager.CreateSqlLog"
	rpcReply := struct{}{}
	rpcArgs := ArgsSqlLogCreate{
		DbName:         dbName,
		DbUser:         dbUser,
		DbTable:        dbTable,
		Query:          query,
		QueryResults:   queryResults,
		QueryTimestamp: queryTimestamp,
		QueryDuration:  pqinterval.Duration(queryDuration),
		TotalDuration:  pqinterval.Duration(totalDuration),
		ClientName:     clientName,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return errRpc
	}

	// Return scan scopes
	return nil
}

func RpcGetSqlLogs(
	logger scanUtils.Logger,
	rpc *utils.Client,
	dbName string,
	since time.Time,
) ([]database.T_sql_log, error) {

	// Prepare RPC request
	rpcEndpoint := "Manager.GetSqlLogs"
	rpcReply := ReplySqlLogs{}
	rpcArgs := ArgsSqlLogsFilter{
		DbName: dbName,
		Since:  since,
	}

	// Send RPC request
	errRpc := rpc.Call(logger, rpcEndpoint, rpcArgs, &rpcReply) // rpcReply must be pointer to receive result!
	if errRpc != nil {
		return []database.T_sql_log{}, errRpc
	}

	// Return scan scopes
	return rpcReply.Logs, nil
}
