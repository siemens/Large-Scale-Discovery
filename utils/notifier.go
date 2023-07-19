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
	"context"
	"fmt"
	"sync"
	"time"
)

var ErrNotifierShuttingDown = fmt.Errorf("notifier shutting down")

// Notifier allows to send a notification message to all current subscribers. Subscribers must subscribe *again*, if
// after each notification they received. In order to give them some time to subscribe again, the Notifier is
// initialized withe a certain broadcast interval. It will collect messages until the interval is triggered. Messages
// can be of any type.
type Notifier struct {
	listener  chan interface{}
	receivers chan chan []interface{}
	ticker    *time.Ticker
	close     chan struct{}
	ctx       context.Context
	ctxCancel context.CancelFunc
	fnSkip    func(interface{}, interface{}) bool // Function to compare if two values are equal
	wgDone    sync.WaitGroup
}

// NewNotifier initializes a new notifier that can be subscribed to. 'fnSkip' is a function that can be used to
// decide whether a new value shall be added to the list of messages to be broadcast. E.g. if you want to filter
// duplicates.
func NewNotifier(broadcastInterval time.Duration, fnSkip func(interface{}, interface{}) bool) *Notifier {

	// Initialize notifier context
	ctx, ctxCancel := context.WithCancel(context.Background())

	// Initialize notifier
	bc := &Notifier{
		listener:  make(chan interface{}),
		receivers: make(chan chan []interface{}),
		ticker:    time.NewTicker(broadcastInterval),
		ctx:       ctx,
		ctxCancel: ctxCancel,
		fnSkip:    fnSkip,
		wgDone:    sync.WaitGroup{},
	}

	// Launch notifier
	go bc.listenMessages()

	// Return notifier
	return bc
}

// Shutdown shuts down the notifier and waits until all final notifications are fully processed.
func (b *Notifier) Shutdown() {
	b.ctxCancel()
	b.ticker.Stop()
	b.wgDone.Wait()
}

func (b *Notifier) Send(msg interface{}) {
	select {
	case <-b.ctx.Done():
		// Close signal or closed channel
		return
	case b.listener <- msg:
		// Message forwarded
		return
	}
}

func (b *Notifier) Receive() ([]interface{}, error) {
	receiver := make(chan []interface{})
	select {
	case <-b.ctx.Done():
		// Close signal or closed channel
		return nil, ErrNotifierShuttingDown
	case b.receivers <- receiver:
		return <-receiver, nil
	}
}

func (b *Notifier) listenMessages() {

	// Prepare message cache
	var messages []interface{}

	// Set wait group
	b.wgDone.Add(1)

	// Make sure wait group gets released after completion
	defer b.wgDone.Done()

	// Listen for messages, termination signal or send signal
	for {
		select {
		case <-b.ctx.Done():

			// Send remaining messages
			if len(messages) > 0 {

				// Broadcast collected messages
				b.sendMessages(messages)

				// Clean cache
				messages = []interface{}{}
			}

			// Close channels to indicate shutdown
			b.closeReceivers()

			// Return
			return

		case msg := <-b.listener:

			// Check if message should be added
			yes := true
			for _, m := range messages {
				if b.fnSkip(m, msg) {
					yes = false
					break
				}
			}

			// Cache message
			if yes {
				messages = append(messages, msg)
			}

		case <-b.ticker.C:

			// Send messages
			if len(messages) > 0 {

				// Broadcast collected messages
				b.sendMessages(messages)

				// Clean cache
				messages = []interface{}{}
			}
		}
	}
}

func (b *Notifier) sendMessages(messages []interface{}) {
	cnt := 0
	for {
		select {
		case receiver := <-b.receivers:

			// Count receiver
			cnt += 1

			// Send message to subscribed receiver
			receiver <- messages

			// Close receiver channel as it will not be reused
			close(receiver)

		default:

			// Message sent to all receivers, listen for next message
			return
		}
	}
}

func (b *Notifier) closeReceivers() {
	for {
		select {
		case receiver := <-b.receivers:
			close(receiver)
		default:
			// All receiver got the message
			return
		}
	}
}
