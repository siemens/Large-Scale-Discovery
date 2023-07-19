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

// Json implements a scanner valuer interface for json.RawMessage as required by GORM to handle JSON data types.
// The JSON byte stream can be scanned into a given JSON struct for processing.
// https://gorm.io/docs/data_types.html
type Json json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *Json) Scan(value interface{}) error {

	// Cast value to bytes
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("type assertion to []byte failed")
	}

	// Unmarshal bytes into JSON
	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = Json(result)

	// Return nil or error
	return err
}

// Value return json value, implement driver.Valuer interface
func (j Json) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	value, err := json.RawMessage(j).MarshalJSON()
	return value, err
}

// JsonMap is a generic helper type to allow reading writing dynamic map structures. The map's value types can be
// arbitrary and arbitrarily deep. However, to work with the data after unmarshalling, the values must be known and
// casted into their real data types.
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
