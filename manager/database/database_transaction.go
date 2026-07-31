/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package database

import (
	"time"

	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// Transaction provides domain-specific manager database operations without exposing the underlying GORM handle
type Transaction struct {
	handle *gorm.DB
}

// WithTransaction executes manager database operations atomically
func WithTransaction(action func(*Transaction) error) error {

	// Keep the GORM transaction handle private to the database package
	return managerDb.Transaction(func(handle *gorm.DB) error {
		return action(&Transaction{handle: handle})
	})
}

// CreateDatabaseEntry creates a database server entry within the transaction
func (transaction *Transaction) CreateDatabaseEntry(
	dbName string,
	dbDialect string,
	dbServer string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbHostPublic string,
	dbArgs string,
) (*T_db_server, error) {

	// Create the entry through the private transaction handle
	return createDatabaseEntry(
		transaction.handle,
		dbName,
		dbDialect,
		dbServer,
		dbPort,
		dbUser,
		dbPassword,
		dbHostPublic,
		dbArgs,
	)
}

// CreateScopeEntry creates a scan scope entry within the transaction
func (transaction *Transaction) CreateScopeEntry(
	dbServerEntry *T_db_server,
	name string,
	dbName string,
	groupId uint64,
	createdBy string,
	secret string,
	scopeType string,
	cycles bool,
	cyclesRetention int,
	splitSize uint32,
	attributes utils.JsonMap,
	size uint,
	settingsEntry T_scan_setting,
) (*T_scan_scope, error) {

	// Create the entry through the private transaction handle
	return createScopeEntry(
		transaction.handle,
		dbServerEntry,
		name,
		dbName,
		groupId,
		createdBy,
		secret,
		scopeType,
		cycles,
		cyclesRetention,
		splitSize,
		attributes,
		size,
		settingsEntry,
	)
}

// CreateViewEntry creates a scope view entry within the transaction
func (transaction *Transaction) CreateViewEntry(
	scopeEntry *T_scan_scope,
	name string,
	createdBy string,
	filters map[string][]string,
	viewTableNames []string,
) (*T_scope_view, error) {

	// Create the entry through the private transaction handle
	return createViewEntry(transaction.handle, scopeEntry, name, createdBy, filters, viewTableNames)
}

// CreateGrantEntry creates a view grant entry within the transaction
func (transaction *Transaction) CreateGrantEntry(
	viewEntry *T_scope_view,
	isUser bool,
	username string,
	createdBy string,
	expiry time.Time,
	description string,
) (*T_view_grant, error) {

	// Create the entry through the private transaction handle
	return createGrantEntry(transaction.handle, viewEntry, isUser, username, createdBy, expiry, description)
}

// DeleteAgentEntry removes a scan agent entry within the transaction
func (transaction *Transaction) DeleteAgentEntry(agentEntry *T_scan_agent) error {

	// Delete the entry through the private transaction handle
	return transaction.handle.Delete(agentEntry).Error
}

// DeleteScopeSettings removes scan scope settings within the transaction
func (transaction *Transaction) DeleteScopeSettings(settingsEntry *T_scan_setting) error {

	// Delete the entry through the private transaction handle
	return transaction.handle.Delete(settingsEntry).Error
}

// DeleteScopeEntry removes a scan scope entry within the transaction
func (transaction *Transaction) DeleteScopeEntry(scopeEntry *T_scan_scope) error {

	// Delete the entry through the private transaction handle
	return transaction.handle.Delete(scopeEntry).Error
}

// DeleteViewEntry removes a scope view entry within the transaction
func (transaction *Transaction) DeleteViewEntry(viewEntry *T_scope_view) error {

	// Delete the entry through the private transaction handle
	return transaction.handle.Delete(viewEntry).Error
}

// DeleteGrantEntry removes a view grant entry within the transaction
func (transaction *Transaction) DeleteGrantEntry(grantEntry *T_view_grant) error {

	// Delete the entry through the private transaction handle
	return transaction.handle.Delete(grantEntry).Error
}

// UpdateViewEntry persists a scope view update within the transaction
func (transaction *Transaction) UpdateViewEntry(viewEntry *T_scope_view) error {

	// Update the entry through the private transaction handle
	return transaction.handle.Model(viewEntry).Updates(viewEntry).Error
}
