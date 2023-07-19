/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"errors"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
	"time"
)

type T_group struct {
	// - Set the JSON ignore flag (json:"-") for sensitive columns that may NEVER be leaked by a JSON response.
	// - Make columns "not null" if possible. Otherwise, use null-types (e.g. sql.NullString).
	// - Avoid 'default' constraints or gorm will replace empty values (0, "", false) with set default values on CREATE!
	// - Define a lower-snake-case json name for every attribute.
	Id         uint64    `gorm:"column:id;primaryKey" json:"id"`
	Name       string    `gorm:"column:name;not null" json:"name"`
	Created    time.Time `gorm:"column:created;not null;default:CURRENT_TIMESTAMP" json:"created"`
	CreatedBy  string    `gorm:"column:created_by;not null" json:"created_by"`
	MaxScopes  int       `gorm:"column:max_scopes;not null" json:"max_scopes"`
	MaxViews   int       `gorm:"column:max_views;not null" json:"max_views"`
	MaxTargets int       `gorm:"column:max_targets;not null" json:"max_targets"`
	MaxOwners  int       `gorm:"column:max_owners;not null" json:"max_owners"`

	AllowCustom  bool `gorm:"column:allow_custom;not null;default:true" json:"allow_custom"`
	AllowNetwork bool `gorm:"column:allow_network;not null;default:false" json:"allow_network"`
	AllowAsset   bool `gorm:"column:allow_asset;not null;default:false" json:"allow_asset"`

	Ownerships []T_ownership `gorm:"foreignKey:IdTGroup" json:"ownerships"`
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (group *T_group) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize value
	group.Name = b.Sanitize(group.Name)
	tx.Statement.SetColumn("name", group.Name)

	// Sanitize value
	group.CreatedBy = b.Sanitize(group.CreatedBy)
	tx.Statement.SetColumn("created_by", group.CreatedBy)

	// Return nil as everything went fine
	return nil
}

// Create a group
func (group *T_group) Create() error {

	// Write group to database
	errDb := backendDb.Create(group).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// Save updates defined columns of a group entry in the database. It updates defined columns, to the currently
// set values, even if the values are empty ones, such as 0, false or "".
// ATTENTION: Only update required columns to avoid overwriting changes of parallel processes (with data in memory)
func (group *T_group) Save(columns ...string) (int64, error) {

	// Verify that columns were supplied
	if len(columns) < 1 {
		return 0, nil
	}

	// Prevent the creation of new groups
	if group.Id == 0 {
		return 0, errors.New("invalid entry ID")
	}

	// Prepare arguments to pass to GORM. Cannot pass string types, but interface types.
	// GORM requires some strange set of arguments
	var arg0 interface{} = columns[0]
	var args = make([]interface{}, 0, len(columns)-1)
	for _, column := range columns[1:] {
		args = append(args, column)
	}

	// Update group in database
	// This doesn't affect association values, so turn off auto-update to avoid unnecessary automatic update queries
	db := backendDb.
		Select(arg0, args...). // Select defines the columns to be updated
		Save(group)            // Save will also update empty values (false, 0, "")
	if db.Error != nil {
		return 0, db.Error
	}

	// Return nil as everything went fine
	return db.RowsAffected, nil
}

// UpdateOwners removes all owners and sets them to the given list of new owners. The ownerships set in the group will
// be updated by this function.
func (group *T_group) UpdateOwners(users []T_user) error {

	return backendDb.Transaction(func(txBackendDb *gorm.DB) error {

		// Delete existing owners
		errDb := txBackendDb.Where("id_t_group = ?", group.Id).Delete(&T_ownership{}).Error
		if errDb != nil {
			return fmt.Errorf("could not initiate update of group owners: %s", errDb)
		}

		// Check for empty slices, as this would lead to a error when inserting
		if len(users) < 1 {
			return nil
		}

		// Set the ownerships to an empty slice, as gorm won't delete the old ones from the group. Back them up though
		// in case of an error.
		owners := group.Ownerships
		group.Ownerships = []T_ownership{}

		// Add existing Owners
		relations := make([]T_ownership, 0, len(users))
		for _, user := range users {

			// Prepare join table entry
			relations = append(relations, T_ownership{
				IdTGroup: group.Id,
				IdTUser:  user.Id,
			})
		}

		errDb = txBackendDb.Model(group).Association("Ownerships").Append(&relations)
		if errDb != nil {
			// Restore the ownerships, as the new ownerships are added to the group regardless whether an error
			// occurs
			group.Ownerships = owners

			return errDb
		}

		// Return nil as everything went fine
		return nil
	})
}

// Delete a group
func (group *T_group) Delete() error {

	// Write group to database
	errDb := backendDb.Delete(group).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// AddOwner creates an ownership by adding a user to a group. The ownerships set in the group will be updated by this
// function. However, the existing ownerships must not have the User.Ownerships or Group values set, as this will result
// in an endless SQL query. (The group returned by GetGroup is valid)
func (group *T_group) AddOwner(user *T_user) error {

	// Prepare join table entry
	relation := T_ownership{
		IdTGroup: group.Id,
		IdTUser:  user.Id,
	}

	// Backup ownerships in case of an error
	owners := group.Ownerships

	// Write group to database
	errDb := backendDb.Model(group).Association("Ownerships").Append(&relation)
	if errDb != nil {

		// Restore the ownerships, as the new ownership is added to the group regardless whether an error
		// occurs
		group.Ownerships = owners

		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// GetGroups gets all groups from the db
func GetGroups() ([]T_group, error) {

	// Declare query results
	var entries = make([]T_group, 0, 3) // Initialize empty slice to avoid returning nil to frontend

	// Execute query
	errDb := backendDb.
		Preload("Ownerships").
		Preload("Ownerships.User").
		Find(&entries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return entries
	return entries, nil
}

func GetGroupsOfUser(userId uint64) ([]T_group, error) {

	// Declare query results
	var entries = make([]T_group, 0, 3) // Initialize empty slice to avoid returning nil to frontend

	// Execute query
	errDb := backendDb.
		Preload("Ownerships").
		Preload("Ownerships.User").
		Joins("LEFT JOIN t_ownerships on t_ownerships.id_t_group=t_groups.id").
		Where("t_ownerships.id_t_user = ?", userId).
		Find(&entries).Error
	if errDb != nil {
		return entries, errDb
	}

	// Return entries
	return entries, nil
}

// GetGroup searches a group by ID and returns a pointer to the found group. If no group is found, a nil pointer but no
// error will be returned.
func GetGroup(id uint64) (*T_group, error) {

	// Declare query results
	entries := make([]T_group, 0, 1)

	// Execute query
	errDb := backendDb.
		Preload("Ownerships").
		Preload("Ownerships.User").
		Where("id = ?", id).
		Find(&entries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return nil if no entries were found
	if len(entries) < 1 {
		return nil, nil
	}

	// Return entries
	return &entries[0], nil
}
