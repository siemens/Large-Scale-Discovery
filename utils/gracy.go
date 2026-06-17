/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2025.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Gracy handles our graceful shutdown procedure. Fortunately, Gracy is more capable in doing so then Golang itself.
// Unfortunately, Golang does not execute registered deferred statements after an interrupt. Instead, it aborts affected
// Goroutines right away and jumps into the signal handling function. So, Gracy will keep track of everything you want
// to clean up instead.
type Gracy struct {
	shutdownFns   []func()
	registerOnce  sync.Once
	shutdownOnce  sync.Once
	shutdownMutex sync.Mutex
}

// NewGracy yields a fresh Gracy keeping track of registered shutdown functions.
func NewGracy() *Gracy {
	return &Gracy{}
}

// Promote assigns Gracy as our default interrupt handler. In her duty, she will listen for interrupts and run all
// previously registered shutdown functions.
func (g *Gracy) Promote() {
	g.registerOnce.Do(func() {

		// Prepare exit signal channel and specify the signals it should receive notifications from.
		chSignals := make(chan os.Signal, 1)
		signal.Notify(chSignals, syscall.SIGINT, syscall.SIGTERM) // Keyboard interrupt + Linux termination signal

		// Run observer goroutine
		go func() {

			// Wait for interrupt
			<-chSignals

			// Execute shutdown
			g.Shutdown()
		}()
	})
}

// Register registers a function to be executed on shutdown. Comparable to the defer statement.
func (g *Gracy) Register(shutdownFn func()) {

	// Avoid manipulation of shutdown functions while they might be executed
	g.shutdownMutex.Lock()
	defer g.shutdownMutex.Unlock()

	// Register shutdown function
	g.shutdownFns = append(g.shutdownFns, shutdownFn)
}

// Shutdown executes all registered shutdown functions sequentially in reverse order. Comparable to the execution of
// deferred statements.
func (g *Gracy) Shutdown() {

	// Avoid manipulation of shutdown functions while they might be executed
	g.shutdownMutex.Lock()
	defer g.shutdownMutex.Unlock()

	// Execute shutdown functions, but only once if called multiple times
	g.shutdownOnce.Do(func() {

		for i := len(g.shutdownFns) - 1; i >= 0; i-- {
			g.shutdownFns[i]()
		}
	})
}
