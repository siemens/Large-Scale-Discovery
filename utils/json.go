/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"encoding/json"
	"fmt"
)

// JsonToStruct parses JSON data (e.g. map[string]interface) and populates a referenced struct.
// ATTENTION: The JSON keys must either match the struct attribute names or the JSON struct tags to fill!
func JsonToStruct(jsonData interface{}, jsonStruct interface{}) error {

	// Convert data blob into JSON bytes, in order to reconstruct the JSON attributes
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("could not parse targets in scan scope configuration")
	}

	// Load JSON bytes into target data struct
	errU := json.Unmarshal(jsonBytes, &jsonStruct)
	if errU != nil {
		return fmt.Errorf("could not load targets in scan scope configuration")
	}

	// Return populated struct
	return nil
}
