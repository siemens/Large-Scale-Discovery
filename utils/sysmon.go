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
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"time"
)

type SystemData struct {
	Platform        string  `json:"platform"`
	PlatformFamily  string  `json:"platform_family"`
	PlatformVersion string  `json:"platform_version"`
	CpuRate         float64 `json:"cpu_rate"`
	MemoryRate      float64 `json:"memory_rate"`
}

type SystemMonitor struct {
	sysMonData SystemData
	ctx        context.Context
	ctxCancel  context.CancelFunc
	chNew      chan SystemData
	chRead     chan SystemData
}

// NewSystemMonitor initializes a new system resources monitor, regularly measuring the system utilization.
func NewSystemMonitor(ctxParent context.Context) *SystemMonitor {

	// Prepare a new context, deriving from the parent one
	ctx, ctxCancelFunc := context.WithCancel(ctxParent)

	// Construct and return the system monitor
	return &SystemMonitor{
		sysMonData: SystemData{},
		ctx:        ctx,
		ctxCancel:  ctxCancelFunc,
		chNew:      make(chan SystemData),
		chRead:     make(chan SystemData),
	}
}

// Run launches the system utilization monitor
func (sm *SystemMonitor) Run(interval time.Duration) {

	// Define measurement procedure
	fnUpdate := func() {

		// Get system information
		platform, family, version, _ := host.PlatformInformation()

		// Prepare memory for new measurement
		measurement := SystemData{
			Platform:        platform,
			PlatformFamily:  family,
			PlatformVersion: version,
			CpuRate:         0,
			MemoryRate:      0,
		}

		// Measure for x amount of time (blocking)
		cRate, errCpu := cpu.Percent(interval, false)
		if errCpu != nil {
			measurement.CpuRate = -1
		} else {
			measurement.CpuRate = cRate[0]
		}

		// Get system memory usage
		m, errMem := mem.VirtualMemory()
		if errMem != nil {
			measurement.MemoryRate = -1
		} else {
			measurement.MemoryRate = m.UsedPercent
		}

		// Update latest measurement data
		sm.chNew <- measurement
	}

	// Launch initial measurement
	go fnUpdate()

	// Handle measurement updates or data requests
	for {

		// Get latest measurements
		cached := sm.sysMonData

		// Read new measurement or answer data requests
		select {
		case newData := <-sm.chNew:

			// Update cached system data
			sm.sysMonData = newData

			// Launch next measurement
			go fnUpdate()
		case sm.chRead <- cached:

			// Latest known measurement values were returned
		case <-sm.ctx.Done():

			// Terminate system monitor
			return
		}
	}
}

// Shutdown terminates the system usage monitor
func (sm *SystemMonitor) Shutdown() {
	sm.ctxCancel()
}

// Get retrieves the last known system usage data
func (sm *SystemMonitor) Get() SystemData {
	return <-sm.chRead
}
