// Copyright (c) 2015, Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License0.

package mixnet

import (
	"fmt"
	"net"
	"testing"
)

type serverType int

const (
	sender serverType = iota
	receiver
)

var sendTemplate = "You are client no. %d, this is msg. no. %d. "

func runDummyServerWriteOne(msg []byte, ch chan<- testResult) {
	l, err := net.Listen(network, dstAddr)
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer c.Close()

	_, err = c.Write(msg)
	ch <- testResult{err, nil}
}

func runDummyServerReadOne(ch chan<- testResult) {
	l, err := net.Listen(network, dstAddr)
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer c.Close()

	buf := make([]byte, CellBytes*10)
	bytes, err := c.Read(buf)
	if err != nil {
		ch <- testResult{err, nil}
		return
	}
	ch <- testResult{nil, buf[:bytes]}
}

// A dummy server that accepts ct connections and waits for a message
// from each client.
func runDummyServer(clientCt, msgCt int, ch chan<- testResult, t serverType) {
	l, err := net.Listen(network, dstAddr)
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer l.Close()

	done := make(chan bool)
	for i := 0; i < clientCt; i++ {
		c, err := l.Accept()
		if err != nil {
			ch <- testResult{err, []byte{}}
			return
		}

		go func(c net.Conn, clientNo int) {
			defer c.Close()
			buff := make([]byte, CellBytes*10)
			for j := 0; j < msgCt; j++ {
				if t == receiver {
					bytes, err := c.Read(buff)
					if err != nil {
						ch <- testResult{err, []byte{}}
					} else {
						ch <- testResult{nil, buff[:bytes]}
					}
				} else if t == sender {
					if _, err := c.Write([]byte(fmt.Sprintf(sendTemplate, clientNo, j))); err != nil {
						ch <- testResult{err, []byte{}}
					} else {
						ch <- testResult{nil, []byte{}}
					}
				}
				done <- true
			}
		}(c, i)
	}

	for i := 0; i < clientCt*msgCt; i++ {
		<-done
	}
}

// Test Queue by enqueueing a bunch of messages and dequeueing them.
// Test multiple rounds.
func TestQueueSend(t *testing.T) {

	// batchSize must divide clientCt; otherwise the sendQueue will block forever.
	batchSize := 2
	clientCt := 4
	msgCt := 3

	sq := NewQueue(network, batchSize)
	kill := make(chan bool)
	done := make(chan bool)
	dstCh := make(chan testResult)

	go runDummyServer(clientCt, msgCt, dstCh, receiver)

	go func() {
		sq.DoQueue(kill)
		done <- true
	}()

	go func() {
		sq.DoQueueErrorHandler(kill)
		done <- true
	}()

	for round := 0; round < msgCt; round++ {
		// Enqueue some messages.
		for i := 0; i < clientCt; i++ {
			q := new(Queueable)
			q.id = uint64(i)
			q.addr = &dstAddr
			q.msg = []byte(
				fmt.Sprintf("I am anonymous, but my ID is %d.", i))
			sq.Enqueue(q)
		}

		// Read results from destination server.
		for i := 0; i < clientCt; i++ {
			res := <-dstCh
			if res.err != nil {
				t.Error(res.err)
				break
			} else {
				t.Log(string(res.msg))
			}
		}
	}

	kill <- true
	kill <- true

	<-done
	<-done
}

func TestQueueReceive(t *testing.T) {

	msgCt := 1

	sq := NewQueue(network, 1)
	kill := make(chan bool)
	done := make(chan bool)
	dstCh := make(chan testResult)

	go runDummyServer(1, msgCt, dstCh, sender)

	go func() {
		sq.DoQueue(kill)
		done <- true
	}()

	go func() {
		sq.DoQueueErrorHandler(kill)
		done <- true
	}()

	q := new(Queueable)
	q.id = 99
	q.addr = &dstAddr
	q.reply = make(chan []byte)
	sq.Enqueue(q)

	res := <-dstCh
	if res.err != nil {
		t.Error(res.err)
	} else {
		t.Log(string(<-q.reply))
	}

	kill <- true
	kill <- true

	<-done
	<-done
}
