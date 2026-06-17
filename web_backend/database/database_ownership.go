/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

// T_ownership is a join-table to establish a many-to-many relationship between users and groups. Each expressed
// relationship contains additional attributes, like whether it is an administrative relationship.
type T_ownership struct {
	// "uniqueIndex" is a workaround to introduce a "unique" mechanism across multiple columns (group id and user id)
	Id       uint64 `gorm:"column:id;primaryKey" json:"id"`
	IdTGroup uint64 `gorm:"column:id_t_group;type:int;not null;uniqueIndex:idx_group_user"` // SQLITE3 does only support FK via type definition https://github.com/go-gorm/gorm/issues/765 https://www.sqlite.org/foreignkeys.html
	IdTUser  uint64 `gorm:"column:id_t_user;type:int;not null;uniqueIndex:idx_group_user"`  // SQLITE3 does only support FK via type definition https://github.com/go-gorm/gorm/issues/765 https://www.sqlite.org/foreignkeys.html

	Group T_group `gorm:"foreignKey:IdTGroup" json:"group"`
	User  T_user  `gorm:"foreignKey:IdTUser" json:"user"`
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
