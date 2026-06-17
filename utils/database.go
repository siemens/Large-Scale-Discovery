/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"fmt"

	"gorm.io/gorm"
)

// CheckPragmas queries the database to verify that required PRAGMAs are correctly set.
func CheckPragmas(db *gorm.DB, pragmas map[string]interface{}) error {
	for name, exp := range pragmas {
		switch v := exp.(type) {
		case int:
			var value int
			if err := db.Raw(fmt.Sprintf("PRAGMA %s", name)).Scan(&value).Error; err != nil {
				return fmt.Errorf("could not read PRAGMA '%s': %w", name, err)
			}
			if value != v {
				return fmt.Errorf("PRAGMA '%s' is set to %d instead of %d", name, value, v)
			}

		case string:
			var value string
			if err := db.Raw(fmt.Sprintf("PRAGMA %s", name)).Scan(&value).Error; err != nil {
				return fmt.Errorf("could not read PRAGMA '%s': %w", name, err)
			}
			if value != v {
				return fmt.Errorf("PRAGMA '%s' is set to '%s' instead of '%s'", name, value, v)
			}

		default:
			return fmt.Errorf("unexpected type for PRAGMA '%s'", name)
		}
	}

	// Return nil as everything went fine
	return nil
}
