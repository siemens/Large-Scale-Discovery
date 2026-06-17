/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JsonMap is a generic helper type to allow reading writing dynamic map structures. The map's value types can be
// arbitrary and arbitrarily deep. However, to work with the data after unmarshalling, the values must be known and
// cast into their real data types.
// https://www.alexedwards.net/blog/using-postgresql-jsonb
type JsonMap map[string]interface{}

func (a JsonMap) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *JsonMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		s, ok := value.(string) // A different SQLITE editor might have written a JSON string instead of binary JSON
		if !ok {
			return fmt.Errorf("type assertion to []byte failed")
		}
		b = []byte(s)
	}
	return json.Unmarshal(b, &a)
}
