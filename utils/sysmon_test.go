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
	"context"
	"fmt"
	"testing"
	"time"
)

// TestNewSysMon verifies that the system monitor reports zero values before the first interval elapses,
// non-zero values afterward, and that measurements update across consecutive intervals.
func TestNewSysMon(t *testing.T) {

	// Define test interval
	interval := time.Second / 3

	// Define check frequency
	frequency := 3

	// Initialize sys mon
	sm := NewSystemMonitor(context.Background())

	// Launch sys mon
	go func() {

		// Delay measurement a bit to avoid race conditions with subsequent checks
		time.Sleep(interval / time.Duration(frequency*5))

		// Run sys mon in intervals
		sm.Run(interval)
	}()

	// Terminate sys mon
	defer sm.Shutdown()

	// Boundary cache only updates at interval crossings so the cross-interval comparison always
	// compares readings from distinct measurement windows, not consecutive ticks within the same one.
	var boundaryCache SystemData

	// Record loop start for reliable time-based assertions that do not depend on scheduler precision.
	startTime := time.Now()

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
		elapsed := time.Since(startTime)

		// Print loop count
		fmt.Printf("Loop °%2d: CPU=%f, Mem=%f\n", i+1, values.CpuRate, values.MemoryRate)

		// Check for errors
		if values.CpuRate == -1 {
			t.Errorf("NewSystemMonitor() CPU rate is negative, indicating an error")
		}
		if values.MemoryRate == -1 {
			t.Errorf("NewSystemMonitor() memory rate is negative, indicating an error")
		}

		// Values must be zero while well inside the first measurement window. The first measurement
		// cannot complete before goroutine startup delay (interval/15) + measurement duration (interval)
		// ≈ 1.07*interval. Using interval/2 as a conservative cutoff guarantees no false positive even
		// when sleeps overrun under load, since interval/2 << goroutine_delay + interval.
		if elapsed < interval/2 {
			if values.CpuRate != 0 {
				t.Errorf("NewSystemMonitor() CPU rate = '%v', want = '0' (elapsed '%v' < interval/2)", values.CpuRate, elapsed)
			}
			if values.MemoryRate != 0 {
				t.Errorf("NewSystemMonitor() memory rate = '%v', want = '0' (elapsed '%v' < interval/2)", values.MemoryRate, elapsed)
			}
		}

		// After two full intervals, memory must be non-zero on any live system.
		// CPU rate may legitimately be 0% on an idle system, so it is not asserted here.
		if elapsed > interval*2 && values.MemoryRate == 0 {
			t.Errorf("NewSystemMonitor() memory rate = '0', want > '0' (elapsed '%v' > interval*2)", elapsed)
		}

		// At each interval boundary, compare against the previous boundary reading to verify that the
		// monitor updated. Skip the comparison when either side is still the zero-value struct (first
		// measurement may arrive slightly late under load), so we only compare real measurement pairs.
		if i >= frequency && i%frequency == 0 {
			if values.MemoryRate != 0 && boundaryCache.MemoryRate != 0 {
				if values.MemoryRate == boundaryCache.MemoryRate && values.CpuRate == boundaryCache.CpuRate {
					t.Errorf("NewSystemMonitor() measurements did not update between intervals: CPU='%v', Mem='%v'", values.CpuRate, values.MemoryRate)
				}
			}
			boundaryCache = values
		}
	}
}
