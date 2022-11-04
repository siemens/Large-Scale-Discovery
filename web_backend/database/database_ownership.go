/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

// T_ownership is a join-table to establish a many-to-many relationship between users and groups. Each expressed
// relation ship contains additional attributes, like whether it is an administrative relationship.
type T_ownership struct {
	// "uniqueIndex" is a workaround to introduce a "unique" mechanism across multiple columns (group id and user id)
	Id       uint64 `gorm:"column:id;primaryKey" json:"id"`
	IdTGroup uint64 `gorm:"column:id_t_group;type:int;not null;uniqueIndex:idx_group_user"` // SQLITE3 does only support FK via type definition https://github.com/go-gorm/gorm/issues/765 https://www.sqlite.org/foreignkeys.html
	IdTUser  uint64 `gorm:"column:id_t_user;type:int;not null;uniqueIndex:idx_group_user"`  // SQLITE3 does only support FK via type definition https://github.com/go-gorm/gorm/issues/765 https://www.sqlite.org/foreignkeys.html

	Group *T_group `gorm:"foreignKey:IdTGroup;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"group"`
	User  *T_user  `gorm:"foreignKey:IdTUser;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
}

// Delete an ownership relation
func (ownership *T_ownership) Delete() error {

	// Delete user from database
	errDb := backendDb.Delete(ownership).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// GetOwnership searches an ownership for given Group ID and User ID and returns the associated entry if existing. This
// function can return one entry at most, as the group id and user id are used for a composite unique index. If the
// entry does not exist nil and no error will be returned.
func GetOwnership(groupId, userId uint64) (*T_ownership, error) {

	// Declare query results
	var entries = make([]T_ownership, 0, 1)

	// Execute query
	errDb := backendDb.
		Where("id_t_group = ? AND id_t_user = ?", groupId, userId).
		Limit(1).
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
