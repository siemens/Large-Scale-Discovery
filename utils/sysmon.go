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
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemData struct {
	Platform        string  `json:"platform"`
	PlatformFamily  string  `json:"platform_family"`
	PlatformVersion string  `json:"platform_version"`
	CpuCores        int     `json:"cpu_cores"`
	CpuMhz          float64 `json:"cpu_mhz"`
	CpuRate         float64 `json:"cpu_rate"` // Usage in %
	MemoryBytes     uint64  `json:"memory_bytes"`
	MemoryRate      float64 `json:"memory_rate"` // Usage in %
	VersionNmap     string  `json:"version_nmap"`
	VersionNpcap    string  `json:"version_npcap"`
	VersionSslyze   string  `json:"version_sslyze"`
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

	// Read static hardware characteristics once; these do not change across measurements
	cpuCores, _ := cpu.Counts(true)
	cpuInfos, errCpuInfo := cpu.Info()
	memStatic, errMemStatic := mem.VirtualMemory()

	// Extract CPU frequency
	var cpuMhz float64
	if errCpuInfo == nil && len(cpuInfos) > 0 {
		cpuMhz = cpuInfos[0].Mhz
	}

	// Extract memory size
	var memoryBytes uint64
	if errMemStatic == nil {
		memoryBytes = memStatic.Total
	}

	// Define measurement procedure
	fnUpdate := func() {

		// Get system information
		platform, family, version, _ := host.PlatformInformation()

		// Prepare memory for new measurement; static hardware fields are included in every snapshot
		measurement := SystemData{
			Platform:        platform,
			PlatformFamily:  family,
			PlatformVersion: version,
			CpuRate:         0,
			MemoryRate:      0,
			CpuCores:        cpuCores,
			CpuMhz:          cpuMhz,
			MemoryBytes:     memoryBytes,
		}

		// Measure for x amount of time (blocking)
		cRate, errCpu := cpu.Percent(interval, false)
		if errCpu != nil {
			measurement.CpuRate = -1
		} else {
			measurement.CpuRate = cRate[0]
		}

		// Get system memory usage
		vmem, errMem := mem.VirtualMemory()
		if errMem != nil {
			measurement.MemoryRate = -1
		} else {
			measurement.MemoryRate = vmem.UsedPercent
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
