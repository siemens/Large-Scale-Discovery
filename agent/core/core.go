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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/juju/fslock"
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
	"reflect"
	"strings"
	"sync"
	"time"
)

const taskRequestInterval = time.Second * 5 // Interval in which to ask the broker for new scan tasks
const saveRequestsMax = 10                  // Maximum number scan results sent to the broker in parallel
const instanceFile = ".instance"            // File storing a generated agent instance name to distinguish multiple scan agents on the same machine

var coreCtx, coreCtxCancelFunc = context.WithCancel(context.Background()) // Agent context and cancellation function. Agent should terminate when context is closed
var shutdownOnce sync.Once                                                // Helper variable to prevent shutdown from doing its work multiple times.
var instanceInfo broker.AgentInfo                                         // Agent instance info holding static information like ip, hostname, compatibility version,...
var rpcClient *utils.Client                                               // RPC client struct handling RPC connections and requests
var sysMon *utils.SystemMonitor                                           // Monitoring service collecting information about system utilization, e.g. CPU, memory,...)

var instanceIp string
var instanceHostname string

// Init initializes the agent and all of its parameters
func Init() error {

	// Get global logger
	logger := log.GetLogger()

	// Get config
	conf := config.GetConfig()

	// Lookup agent IP
	instanceIp = utils.GetOutboundIP()
	if instanceIp == "" {
		var errIp error
		instanceIp, errIp = utils.GetLocalIp()
		if errIp != nil {
			return fmt.Errorf("could not read local ip: %s", errIp)
		}
	}

	// Lookup agent
	var errHostname error
	instanceHostname, errHostname = os.Hostname()
	if errHostname != nil {
		return fmt.Errorf("could not read local hostname: %s", errHostname)
	}

	// Load agent instance name. Generate new one if not existing.
	instanceName, errName := loadInstanceName()
	if errName != nil {
		return errName
	}

	// Prepare anonymized scope secret strings
	var secrets []string
	for _, secret := range conf.ScopeSecrets {
		secrets = append(secrets, secret[0:5]+"...")
	}

	// Check whether scan agent is shared across multiple scan scopes
	shared := false
	if len(conf.ScopeSecrets) > 1 {
		shared = true
	}

	// Check whether scan agent has dedicated limits configured
	limits := false
	v := reflect.ValueOf(conf.Modules)
	for i := 0; i < v.NumField(); i++ {
		moduleInstances := v.Field(i).FieldByName("MaxInstances").Interface().(int)
		if moduleInstances > 0 {
			limits = true
		}
	}

	// Log agent attributes
	logger.Infof("Agent Name   : %s", instanceName)
	logger.Infof("Agent Host   : %s", instanceHostname)
	logger.Infof("Agent IP     : %s", instanceIp)
	logger.Infof("Scope Secrets: %s", strings.Join(secrets, ", "))

	// Prepare agent info data to be sent along with RPC requests to the broker
	instanceInfo = broker.AgentInfo{
		CompatibilityLevel: broker.CompatibilityLevel,
		Name:               instanceName,
		Host:               instanceHostname,
		Ip:                 instanceIp,
		Shared:             shared,
		Limits:             limits,
	}

	// Prepare RPC certificate path
	rpcRemoteCrt := filepath.Join("keys", "broker.crt")
	if _build.DevMode {
		rpcRemoteCrt = filepath.Join("keys", "broker_dev.crt")
	}

	// Initialize system monitor
	sysMon = utils.NewSystemMonitor(coreCtx)
	go sysMon.Run(taskRequestInterval)

	// Initialize module counters for all available scan modules
	initModules(conf.ScopeSecrets)

	// Loads the ciphers
	ssl.LoadCiphers(logger)

	// Register gob structures that will be sent via interface{}
	broker.RegisterGobs()

	// Initialize RPC client broker facing
	rpcClient = utils.NewRpcClient(conf.BrokerAddress, true, rpcRemoteCrt)

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

func loadInstanceName() (string, error) {

	// Prepare memory for instance name and origin hash
	instanceName := ""
	instanceOrigin := ""

	// Prepare file lock on instance file so that only one process can use it
	l := fslock.New(instanceFile)

	// Test whether file is accessible
	// This will create the file on the first time if not existing
	if l.TryLock() != nil {
		return "", fmt.Errorf("same agent already running")
	}
	_ = l.Unlock() // Unlock to allow reading agent identifier

	// Load agent instance name
	content, errRead := os.ReadFile(instanceFile)
	if errRead != nil {
		return "", fmt.Errorf("could not load agent idientifier: %s", errRead)
	}

	// Get current location
	instanceFilePath, errInstanceFilePath := filepath.Abs(instanceFile)
	if errInstanceFilePath != nil {
		return "", fmt.Errorf("could retrieve instance file path: %s", errInstanceFilePath)
	}

	// Hash current location
	h := sha256.New()
	h.Write([]byte(instanceIp + instanceHostname + instanceFilePath)) // Distinguishable hashes across hosts and file paths
	currentLocation := hex.EncodeToString(h.Sum(nil))

	// Extract instance name and origin hash
	splits := strings.Split(string(content), "|")
	if len(splits) == 2 {
		instanceName = strings.Trim(splits[0], " ")
		instanceOrigin = splits[1]

		// Generate new instance name if location hash does not match current location
		if instanceOrigin != currentLocation {
			instanceName = "" // Agent might have been moved, give it a new name
			instanceOrigin = ""
		}
	} else if len(splits) == 1 {
		instanceName = strings.Trim(splits[0], " ") // Allow existing instance name but store with location hash
	}

	// Generate new instance name if necessary
	if instanceName == "" {
		n, _ := rand.Int(rand.Reader, big.NewInt(2))
		g := int(n.Int64())
		instanceName = fmt.Sprintf("%s %s", randomdata.SillyName(), randomdata.FirstName(g))
	}

	// Persist new values if instance origin was invalid
	if instanceOrigin == "" {

		// Prepare output in bytes
		output := []byte(instanceName + "|" + currentLocation)

		// Persist new name
		err := os.WriteFile(instanceFile, output, 0660)
		if err != nil {
			return "", fmt.Errorf("could not create agent idientifier: %s", err)
		}
	}

	// Lock instance file until process terminates to prevent other instances using the same identifier
	if l.TryLock() != nil {
		return "", fmt.Errorf("same agent already running")
	}

	// Return nil as everything went fine
	return instanceName, nil
}

// scanTaskLoader keeps asking the broker for new scan tasks
func scanTaskLoader(wg *sync.WaitGroup, chOut chan broker.ScanTask) {

	// Decrement wait group on return
	defer wg.Done()

	// Close the channel on exit
	defer close(chOut)

	// Get tagged logger
	logger := log.GetLogger().Tagged("scanTaskLoader")

	// Catch potential panics to gracefully log issue
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic: %s", r)
			Shutdown() // Shutdown the agent for safety reasons. It should neither end in a stuck state, nor suck all
			// scan targets from the broker transforming them into crashed tasks. This panic might be a severe issue!
		}
	}()

	// Get config
	conf := config.GetConfig()

	// Closure for reusable request code
	requestFunc := func() {

		// Get current system load. Use same system load for all scope task requests,
		// otherwise it might be confusing to the frontend user.
		systemData := sysMon.Get()

		// Get Overall scan agent module instances counts
		totalInstances := GetTotalInstanceCounts()

		// Randomize scope secret order to give each scan scope a similar chance or processing,
		// if overall scan agent limits are configured in the agent configuration.
		scopeSecrets := scanUtils.Shuffle(conf.ScopeSecrets)

		// Execute request for each scan scope
		for _, scopeSecret := range scopeSecrets {

			// Log attempt
			logger.Debugf("Asking for targets for scan scope '%s...'.", scopeSecret[:5])

			// Read module instances counts by agent and by scan scope
			scopeInstances := GetScopeInstanceCounts(scopeSecret)

			// Get initialized module labels
			labels := GetModuleLabels()

			// Prepare module data to be filled with active instances and module settings
			moduleData := make([]broker.ModuleData, 0, len(labels))

			// Get list of initialized module labels to iterate
			for _, label := range labels {

				// Check if module has dedicated agent-side limit configured in agent config file
				labelMax := conf.Modules.ReadMaxInstances(label)

				// Prepare and append module data
				moduleData = append(moduleData, broker.ModuleData{
					Label:          label,
					MaxInstances:   labelMax,              // Maximum instances of this scan module the scan agents want to handle
					TotalInstances: totalInstances[label], // Current amount of instances of this module on the scan agent across all scan scopes
					ScopeInstances: scopeInstances[label], // Current amount of instances of this module on the scan agent in the current scan scope
				})
			}

			// Prepare the request to the RPC server
			rpcArgs := broker.ArgsGetScanTask{
				AgentInfo:   instanceInfo,
				ScopeSecret: scopeSecret,
				ModuleData:  moduleData,
				SystemData:  systemData,
			}

			// Send RPC request for scan targets
			scanTasks, errScanTasks := broker.RpcRequestScanTasks(logger, rpcClient, coreCtx, &rpcArgs)
			if errors.Is(errScanTasks, utils.ErrRpcCompatibility) { // Just address compatibility errors. Ignore other server-side errors to retry again later.
				Shutdown()
				return
			}

			// Log response information
			logger.Debugf("Received %d scan tasks.", len(scanTasks))

			// Send received scan tasks to launcher
			for _, scanTask := range scanTasks {
				chOut <- scanTask
			}
		}
	}

	// Execute initial scope task requests for all initialized scan scopes (immediately, without ticker delay)
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
			requestFunc() // Execute scope task requests for all initialized scan scopes
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

	// Catch potential panics to gracefully log issue
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
		errIncrease := IncrementModuleCount(scanTask.Secret, scanTask.Label)
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

	// Catch potential panics to gracefully log issue
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
