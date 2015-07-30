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
// limitations under the License.

package mixnet

import (
	"container/list"
	"errors"
	"math/rand"
	"net"

	"github.com/golang/glog"
)

type Queueable struct {
	id    uint64
	addr  *string
	msg   []byte
	conn  net.Conn
	reply chan []byte
}

type sendQueueError struct {
	id uint64 // Serial identifier of sender to which error pertains.
	error
}

// The Queue structure maps a serial identifier corresponding to a sender
// (in the router context) to a destination. It also maintains a message buffer
// for each sender. Once there messages ready on enough buffers, a batch of
// messages are transmitted simultaneously.
type Queue struct {
	batchSize int // Number of messages to transmit in a roud.
	ct        int // Current number of buffers with messages ready.

	network string // Network protocol, e.g. "tcp".

	nextAddr   map[uint64]string     // Address of destination.
	nextConn   map[uint64]net.Conn   // Connection to destination.
	sendBuffer map[uint64]*list.List // Message buffer of sender.

	queue chan *Queueable     // Channel for queueing messages/directives.
	err   chan sendQueueError // Channel for handling errors.
}

// NewQueue creates a new Queue structure.
func NewQueue(network string, batchSize int) (sq *Queue) {
	sq = new(Queue)
	sq.batchSize = batchSize
	sq.network = network

	sq.nextAddr = make(map[uint64]string)
	sq.nextConn = make(map[uint64]net.Conn)
	sq.sendBuffer = make(map[uint64]*list.List)

	sq.queue = make(chan *Queueable)
	sq.err = make(chan sendQueueError)
	return sq
}

func (sq *Queue) Enqueue(q *Queueable) {
	sq.queue <- q
}

// EnqueueMsg
func (sq *Queue) EnqueueMsg(id uint64, msg []byte) {
	q := new(Queueable)
	q.id = id
	q.msg = make([]byte, len(msg))
	copy(q.msg, msg)
	sq.queue <- q
}

// EnqueueAddr
func (sq *Queue) EnqueueReply(id uint64, reply chan []byte) {
	q := new(Queueable)
	q.id = id
	q.reply = reply
	sq.queue <- q
}

// EnqueueAddr
func (sq *Queue) SetAddr(id uint64, addr string) {
	q := new(Queueable)
	q.id = id
	q.addr = new(string)
	*q.addr = addr
	sq.queue <- q
}

// EnqueueConn
func (sq *Queue) SetConn(id uint64, c net.Conn) {
	q := new(Queueable)
	q.id = id
	q.conn = c
	sq.queue <- q
}

// DoQueue adds messages to a queue and transmits messages in batches.
// Typically a message is a cell, but when the calling router is an exit point,
// the message length is arbitrary. A batch is transmitted when there are
// messages on batchSize distinct sender channels.
func (sq *Queue) DoQueue(kill <-chan bool) {
	for {
		select {
		case <-kill:
			for _, c := range sq.nextConn {
				// TODO(cjpatton) send DESTROY to next hop.
				c.Close()
			}
			return

		case q := <-sq.queue:
			// Set the next-hop address. We don't allow the destination address
			// or connection to be overwritten in order to avoid accumulating
			// stale connections on routers.
			if _, def := sq.nextAddr[q.id]; !def && q.addr != nil {
				sq.nextAddr[q.id] = *q.addr
			}

			// Set the next-hop connection. This is useful for routing replies
			// from the destination over already created circuits back to the
			// source.
			if _, def := sq.nextConn[q.id]; !def && q.conn != nil {
				sq.nextConn[q.id] = q.conn
			}

			if q.msg != nil || q.reply != nil {

				if _, def := sq.nextAddr[q.id]; !def {
					sq.err <- sendQueueError{q.id,
						errors.New("request to send/receive message without a destination")}
					continue
				}

				// Create a send buffer for the sender ID if it doesn't exist.
				if _, def := sq.sendBuffer[q.id]; !def {
					sq.sendBuffer[q.id] = list.New()
				}
				buf := sq.sendBuffer[q.id]

				// The buffer was empty but now has a message ready; increment
				// the counter.
				if buf.Len() == 0 {
					sq.ct++
				}

				// Add message to send buffer.
				buf.PushBack(q)

				// Transmit the message batch if it is full.
				if sq.ct >= sq.batchSize {
					sq.dequeue()
				}
			}
		}
	}
}

// DoQueueErrorHandler handles errors produced by DoQueue. When this
// is fully fleshed out, it will enqueue into the response queue a Directive
// containing an error message. For now, just print out the error.
func (sq *Queue) DoQueueErrorHandler(kill <-chan bool) {
	for {
		select {
		case <-kill:
			return
		case err := <-sq.err:
			glog.Errorf("send queue (%d): %s\n", err.id, err.Error())
		}
	}
}

// dequeue sends one message from each send buffer for each serial ID in a
// random order. This is called by DoQueue and is not safe to call directly
// elsewhere.
func (sq *Queue) dequeue() {

	// Shuffle the serial IDs.
	order := rand.Perm(int(sq.ct)) // TODO(cjpatton) Use tao.GetRandomBytes().
	ids := make([]uint64, sq.ct)
	i := 0
	for id, buf := range sq.sendBuffer {
		if buf.Len() > 0 {
			ids[order[i]] = id
			i++
		}
	}

	// Issue a sendWorker thread for each message to be sent.
	ch := make(chan senderResult)
	for _, id := range ids[:sq.batchSize] {
		addr := sq.nextAddr[id]
		q := sq.sendBuffer[id].Front().Value.(*Queueable)
		c, def := sq.nextConn[id]
		go senderWorker(sq.network, addr, id, q, c, def, ch, sq.err)
	}

	// Wait for workers to finish.
	for _ = range ids[:sq.batchSize] {
		res := <-ch
		if res.c != nil {
			// Save the connection.
			sq.nextConn[res.id] = res.c

			// Pop the message from the buffer and decrement the counter
			// if the buffer is empty.
			buf := sq.sendBuffer[res.id]
			buf.Remove(buf.Front())
			if buf.Len() == 0 {
				sq.ct--
			}
		}
	}
}

type senderResult struct {
	c  net.Conn
	id uint64
}

func senderWorker(network, addr string, id uint64, q *Queueable, c net.Conn, def bool,
	res chan<- senderResult, err chan<- sendQueueError) {
	var e error

	// Wait to connect until the queue is dequeued in order to prevent
	// an observer from correlating an incoming cell with the handshake
	// with the destination server.
	if !def {
		c, e = net.Dial(network, addr) // TODO(cjpatton) timeout.
		if e != nil {
			err <- sendQueueError{id, e}
			res <- senderResult{nil, id}
			return
		}
	}

	if q.msg != nil { // Send the message.
		if _, e := c.Write(q.msg); e != nil {
			err <- sendQueueError{id, e}
			res <- senderResult{nil, id}
			return
		}

	}

	if q.reply != nil { // Receive a message.
		msg := make([]byte, MaxMsgBytes)
		bytes, e := c.Read(msg)
		if e != nil {
			err <- sendQueueError{id, e}
			res <- senderResult{nil, id}
			q.reply <- []byte{}
			return
		}

		// Pass message to channel.
		q.reply <- msg[:bytes]

	}
	res <- senderResult{c, id}
}
