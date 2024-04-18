/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/orcaman/concurrent-map/v2"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/agent/config"
	broker "github.com/siemens/Large-Scale-Discovery/broker/core"
	"github.com/siemens/Large-Scale-Discovery/log"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const taskRequestInterval = time.Second * 5 // Interval in which to ask the broker for new scan tasks
const saveRequestsMax = 10                  // Maximum number scan results sent to the broker in parallel
const instanceFile = ".instance"            // File storing a generated agent instance name to distinguish multiple scan agents on the same machine

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Agent context and cancellation function. Agent should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var instanceName string                                                   // Agent instance name, randomly generated to be (most likely) unique
var instanceIp string                                                     // Agent IP at the default gateway used for scanning
var instanceHostname string                                               // Agent hostname used for scanning
var scopeSecret string                                                    // Scope secret to authenticate/associate the agent with a certain scan scope during RPC requests
var moduleInstances = cmap.New[int]()                                     // Concurrent map holding the total number of running scans for each module
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests
var sysMon *utils.SystemMonitor                                           // Monitoring service collecting information about system utilization, e.g. CPU, memory,...)

// Init initializes the agent and all of its parameters
func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Load agent instance name. Generate new one if not existing.
	errName := loadInstanceName()
	if errName != nil {
		return errName
	}

	// Lookup agent IP
	instanceIp = utils.GetOutboundIP()
	if instanceIp == "" {
		var errIp error
		instanceIp, errIp = utils.GetLocalIp()
		if errIp != nil {
			return fmt.Errorf("could not read local ip: %s", errIp)
		}
	}

	// Lookup agent hostname
	var errHostname error
	instanceHostname, errHostname = os.Hostname()
	if errHostname != nil {
		return fmt.Errorf("could not read local hostname: %s", errHostname)
	}

	// Initialize attributes
	scopeSecret = conf.ScopeSecret

	// Log agent attributes
	logger.Infof("Agent Name  : %s", instanceName)
	logger.Infof("Agent Host  : %s", instanceHostname)
	logger.Infof("Agent IP    : %s", instanceIp)
	logger.Infof("Scope Secret: %s...", scopeSecret[0:5])

	// Prepare RPC certificate path
	rpcRemoteCrt := filepath.Join("keys", "broker.crt")
	if _build.DevMode {
		rpcRemoteCrt = filepath.Join("keys", "broker_dev.crt")
	}
	errCrt := scanUtils.IsValidFile(rpcRemoteCrt)
	if errCrt != nil {
		return errCrt
	}

	// Initialize system monitor
	sysMon = utils.NewSystemMonitor(coreCtx)
	go sysMon.Run(taskRequestInterval)

	// Initialize module counters for all available scan modules
	osInitModules()

	// Loads the ciphers
	ssl.LoadCiphers(logger)

	// Register gob structures that will be sent via interface{}
	broker.RegisterGobs()

	// Initialize RPC client broker facing
	rpcClient = utils.NewRpcClient(conf.BrokerAddress, rpcRemoteCrt)

	// Connect to broker and wait for successful connection. Abort on shutdown request.
	success := rpcClient.Connect(logger, true)
	if !success {
		select {
		case <-coreCtx.Done():
			return nil
		case <-rpcClient.Established():
		}
	}

	// Return nil as everything went fine
	return nil
}

// Run launches background jobs requesting scan tasks, launching executors and returning scan results until core
// context is terminated
func Run() {

	// Only launch if shutdown is not already in progress
	select {
	case <-coreCtx.Done():
		return
	default:
	}

	// Creates a wait group for goroutines
	var wg sync.WaitGroup
	var wgSave sync.WaitGroup

	// Prepare channel to transfer data across goroutines
	chTargets := make(chan broker.ScanTask)
	chResults := make(chan broker.ArgsSaveScanResult)

	// Start goroutine continuously asking the broker for new scan tasks, if necessary. This goroutine will stop
	// requesting new tasks, once the agent termination signal is set and will cause scanTaskLauncher to terminate
	// subsequently.
	wg.Add(1)
	go scanTaskLoader(&wg, chTargets)

	// Start goroutine that starts a module job when new targets are received after asking for targets. This goroutine
	// will run until its input channel is terminated.
	wg.Add(1)
	go scanTaskLauncher(&wg, chTargets, chResults)

	// Start goroutine listening for scan results and sending them to the broker. This goroutine will run until its
	// input channel is terminated.
	wg.Add(1)
	go scanTaskSaver(&wg, &wgSave, chResults)

	// Wait and process until shutdown request
	<-coreCtx.Done()

	// Wait for known results to be saved, before terminating. While known results are saved, new results might
	// become known and waited for.
	wgSave.Wait()

	// Close results channel to cause scanTaskSaver to terminate
	close(chResults)

	// Wait for all goroutines to have terminated.
	wg.Wait()
}

// Shutdown terminates the application context, which causes associated components to gracefully shut down.
func Shutdown() {
	shutdownOnce.Do(func() {

		// Log termination request
		logger := log.GetLogger()
		logger.Infof("Shutting down.")

		// Close agent context. Waiting goroutines will abort if it is closed.
		coreCtxCancelFunc()

		// Disconnect from broker
		// This connection is still required by the agent until the scanTaskSaver goroutine finished!
		if rpcClient != nil {
			rpcClient.Disconnect()
		}
	})
}

func loadInstanceName() error {

	// Check whether agent instance name exists
	if _, err := os.Stat(instanceFile); err == nil {

		// Load agent instance name
		name, errRead := os.ReadFile(instanceFile)
		if errRead != nil {
			return fmt.Errorf("could not load agent idientifier: %s", errRead)
		}

		// Set agent instance name
		instanceName = string(name)

	} else if os.IsNotExist(err) {

		// Prepare name file
		outputFile, errOpen := os.OpenFile(instanceFile, os.O_CREATE|os.O_WRONLY, 0660)
		if errOpen != nil {
			return fmt.Errorf("could not create agent idientifier: %s", errOpen)
		}

		// Decide random gender based on cryptographic random number generator to not run into diversity issues
		n, _ := rand.Int(rand.Reader, big.NewInt(2))
		g := int(n.Int64())

		// Generate agent instance name
		instanceName = fmt.Sprintf("%s %s", randomdata.SillyName(), randomdata.FirstName(g))

		// Store agent instance name
		_, errWrite := outputFile.WriteString(instanceName)
		if errWrite != nil {
			return fmt.Errorf("could not save agent idientifier: %s", errWrite)
		}

	} else {
		return err
	}

	// Return nil as everything went fine
	return nil
}

// increaseUsageModule - increase the usage counter
func increaseUsageModule(moduleName string) error {

	// Get interface to map value
	count, ok := moduleInstances.Get(moduleName)

	// Abort if module name is not existing
	if !ok {
		return fmt.Errorf("unknwon module")
	}

	// Cast and increase the counter
	moduleInstances.Set(moduleName, count+1)

	// Return nil as everything went fine
	return nil
}

// decreaseUsageModule - decrease the usage counter
func decreaseUsageModule(moduleName string) {

	// Get interface to map value
	count, ok := moduleInstances.Get(moduleName)

	// Abort if module name is not existing
	if !ok {
		return
	}

	// Cast and decrease the counter
	if count > 0 {
		moduleInstances.Set(moduleName, count-1)
	}
}

// scanTaskLoader keeps asking the broker for new scan tasks
func scanTaskLoader(wg *sync.WaitGroup, chOut chan broker.ScanTask) {

	// Decrement wait group on return
	defer wg.Done()

	// Close the channel on exit
	defer close(chOut)

	// Get tagged logger
	logger := log.GetLogger().Tagged("scanTaskLoader")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Initialize stats data to send along. Stats data is collected after new scan tasks were launched and reported
	// with the subsequent request for new tasks. This is to avoid under-reporting actual load, because e.g. banner
	// tasks might be completed until next request interval.
	systemData := sysMon.Get()
	moduleData := make([]broker.ModuleData, 0, len(moduleInstances.Items()))

	// Closure for reusable request code
	requestFunc := func() {

		// Log attempt
		logger.Debugf("Asking for targets.")

		// Prepare the request to the RPC server
		rpcArgs := broker.ArgsGetScanTask{
			AgentInfo: broker.AgentInfo{
				Name: instanceName,
				Host: instanceHostname,
				Ip:   instanceIp,
			},
			ScopeSecret: scopeSecret,
			ModuleData:  moduleData,
			SystemData:  systemData,
		}

		// Send RPC request for scan targets
		scanTasks := broker.RpcRequestScanTasks(logger, rpcClient, coreCtx, &rpcArgs)

		// Log response information
		logger.Debugf("Received %d scan tasks.", len(scanTasks))

		// Send received scan tasks to launcher
		for _, scanTask := range scanTasks {
			chOut <- scanTask
		}

		// Capture updated system load, after new scan tasks got launched
		systemData = sysMon.Get()

		// Capture updated active tasks, after new scan tasks got launched
		moduleData = make([]broker.ModuleData, 0, len(moduleInstances.Items()))
		for module := range moduleInstances.IterBuffered() {
			moduleData = append(moduleData, broker.ModuleData{
				Label:       module.Key,
				ActiveTasks: module.Val,
			})
		}
	}

	// Request scan tasks (immediately, without ticker delay)
	requestFunc()

	// Initialize retry interval ticker
	ticker := time.NewTicker(taskRequestInterval)

	// Loop until agent termination signal
	for {
		select {
		case <-coreCtx.Done(): // Cancellation signal
			logger.Infof("Scan task loader terminated.")
			return
		case <-ticker.C: // Wait for next attempt
			requestFunc()
		}
	}
}

// scanTaskLauncher starts a new goroutine of a scan module for a scan task received from the broker. This will run
// until its input channel is terminated
func scanTaskLauncher(wg *sync.WaitGroup, chIn chan broker.ScanTask, chOut chan broker.ArgsSaveScanResult) {

	// Decrement wait group on return
	defer wg.Done()

	// Get tagged logger
	logger := log.GetLogger().Tagged("scanTaskLauncher")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Loop until input channel is closed
	for {

		// Receive scan task to launch
		scanTask, ok := <-chIn

		// Input channel got closed, termination desired.
		if !ok {
			logger.Infof("Scan task launcher terminated.")
			return
		}

		// Log action
		logger.Debugf("Received a scan task for module '%s' (%s).", scanTask.Label, scanTask.Target)

		// Increase module counter
		errIncrease := increaseUsageModule(scanTask.Label)
		if errIncrease != nil {
			logger.Warningf("Scan task '%s' not handled by this agent.", scanTask.Label)
			continue
		}

		// Launch tasks depending on OS and implementations
		launch(logger, chOut, &scanTask)
	}
}

// scanTaskSaver keeps listening for completed scan tasks and sends them to the broker. This will run until its
// input channel is terminated
func scanTaskSaver(wg *sync.WaitGroup, wgSave *sync.WaitGroup, chIn chan broker.ArgsSaveScanResult) {

	// Decrement wait group on return
	defer wg.Done()

	// Get tagged logger
	logger := log.GetLogger().Tagged("scanTaskSaver")

	// Catch potential panics to gracefully log issue with stacktrace
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Prepare throttle to
	//   1) Avoid overloading the broker (scan scopes * agents * scan module instances == boom)
	//   2) Avoid charging up goroutines trying to connect to a broker that might currently be down
	chThrottle := make(chan struct{}, saveRequestsMax)

	// Loop until input channel is closed
	for {

		// Receive scan result to submit
		rpcArgs, ok := <-chIn

		// Input channel got closed, termination desired.
		if !ok {
			logger.Infof("Scan result saver terminated.")
			return
		}

		// Send scan result to broker
		logger.Debugf("Scan result ready to send.")

		// Proceed when there is a throttle slot available
		chThrottle <- struct{}{}

		// Start routine that will send the scan result to broker
		wgSave.Add(1)
		go broker.RpcSubmitScanResult(logger, rpcClient, coreCtx, wgSave, chThrottle, rpcArgs)
	}
}
