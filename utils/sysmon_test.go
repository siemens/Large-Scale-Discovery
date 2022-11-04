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
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewSysMon(t *testing.T) {

	// Define test interval
	interval := time.Second / 3

	// Define check frequency
	frequency := 3

	// Initialize sys mon
	sm := NewSystemMonitor(context.Background())

	// Launch sys mon
	go func() {

		// Delay measurement a bit to avoid race conditions with subsequent checs
		time.Sleep(interval / time.Duration(frequency*5))

		// Run sys mon in intervals
		sm.Run(interval)
	}()

	// Terminate sys mon
	defer sm.Shutdown()

	// Prepare some cache to compare values later
	cache := SystemData{}

	// Read some sys mon values
	for i := 0; i < frequency*5; i++ {

		// Newline every time a new measurement should be available
		if i >= frequency && i%frequency == 0 {
			fmt.Println()
		}

		// Measure after some delay to mitigate race conditions
		time.Sleep(interval / time.Duration(frequency))

		// Grab latest values
		values := sm.Get()

		// Print loop count
		fmt.Println(fmt.Sprintf("Loop °%2d: CPU=%f, Mem=%f", i+1, values.CpuRate, values.MemoryRate))

		// Check for errors
		if values.CpuRate == -1 {
			t.Errorf("\tERROR: CPU rate is negative indicating an error.")
		}
		if values.MemoryRate == -1 {
			t.Errorf("\tERROR: Memory rate is negative indicating an error.")
		}

		// First values should be empty until first interval passed
		if i < frequency && values.CpuRate != 0 {
			t.Errorf("\tERROR: CPU rate should yet be 0.")
		}
		if i < frequency && values.MemoryRate != 0 {
			t.Errorf("\tERROR: Memory rate should yet be 0.")
		}

		// First interval has passed and values should be greater 0
		if i >= frequency && values.CpuRate == 0 {
			t.Errorf("\tERROR: CPU rate is still 0.")
		}
		if i >= frequency && values.MemoryRate == 0 {
			t.Errorf("\tERROR: Memory rate is still 0.")
		}

		// Check whether values of subsequent measurement differ
		if i >= frequency && i%frequency == 0 {
			if values.MemoryRate == cache.MemoryRate && values.CpuRate == cache.CpuRate {
				t.Errorf("\tERROR: Measurements did not update.")
			}
		}

		// Cache values to compare them in one of the next cycles
		cache = values
	}
}
