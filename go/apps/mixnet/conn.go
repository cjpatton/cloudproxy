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
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/golang/protobuf/proto"
)

const (

	// CellBytes specifies the length of a cell.
	CellBytes = 1 << 10

	// MaxMsgBytes specifies the maximum length of a message.
	MaxMsgBytes = 1 << 16
)

const (
	msgCell = iota
	dirCell
	relayCell
)

var errCellLength = errors.New("incorrect cell length")
var errCellType = errors.New("incorrect cell type")
var errBadCellType = errors.New("unrecognized cell type")
var errBadDirective = errors.New("received bad directive")
var errMsgLength = errors.New("message too long")

var dirCreated = &Directive{Type: DirectiveType_CREATED.Enum()}
var dirAwaitMsg = &Directive{Type: DirectiveType_AWAIT_MSG.Enum()}

// Conn implements the net.Conn interface. The read and write operations are
// overloaded to check that only cells are sent between agents in the mixnet
// protocol.
type Conn struct {
	net.Conn
	id uint64 // Serial identifier of connection in a given context.
}

// Read a cell from the channel. If len(msg) != CellBytes, return an error.
func (c Conn) Read(msg []byte) (n int, err error) {
	n, err = c.Conn.Read(msg)
	if err != nil {
		return n, err
	}
	if n != CellBytes {
		return n, errCellLength
	}
	return n, nil
}

// Write a cell to the channel. If the len(cell) != CellBytes, return an error.
func (c *Conn) Write(msg []byte) (n int, err error) {
	if len(msg) != CellBytes {
		return 0, errCellLength
	}
	n, err = c.Conn.Write(msg)
	if err != nil {
		return n, err
	}
	return n, nil
}

// SendDirective serializes and pads a directive to the length of a cell and
// sends it to the router. A directive is signaled to the receiver by the first
// byte of the cell. The next few bytes encode the length of of the serialized
// protocol buffer. If the buffer doesn't fit in a cell, then throw an error.
func (c *Conn) SendDirective(d *Directive) (int, error) {
	db, err := proto.Marshal(d)
	if err != nil {
		return 0, err
	}
	dirBytes := len(db)

	// Throw an error if encoded Directive doesn't fit into a cell.
	if dirBytes+9 > CellBytes {
		return 0, errCellLength
	}

	cell := make([]byte, CellBytes)
	cell[0] = dirCell
	n := binary.PutUvarint(cell[1:], uint64(dirBytes))
	copy(cell[1+n:], db)

	return c.Write(cell)
}

// ReceiveDirective awaits a reply from the router and return the type of
// directive received. This is in response to RouterContext.HandleProxy().
// If the directive type is ERROR or FATAL, return an error.
func (c *Conn) ReceiveDirective() (*Directive, error) {
	var err error
	cell := make([]byte, CellBytes)
	if _, err = c.Read(cell); err != nil && err != io.EOF {
		return nil, err
	}

	if cell[0] != dirCell {
		return nil, errBadCellType
	}

	dirBytes, n := binary.Uvarint(cell[1:])
	d := new(Directive)
	if err := proto.Unmarshal(cell[1+n:1+n+int(dirBytes)], d); err != nil {
		return nil, err
	}

	if *d.Type == DirectiveType_ERROR {
		return nil, errors.New("router error: " + (*d.Error))
	} else if *d.Type == DirectiveType_FATAL {
		return nil, errors.New("router error: " + (*d.Error) + " (connection closed)")
	}
	return d, nil
}

// SendMessage divides a message into cells and sends each cell over the network
// connection. directs the router to relay a message over the already constructed
// circuit. A message is signaled to the reecevier by the first byte of the first
// cell. The next few bytes encode the total number of bytes in the message.
func (c *Conn) SendMessage(msg []byte) (int, error) {
	msgBytes := len(msg)
	cell := make([]byte, CellBytes)
	cell[0] = msgCell
	n := binary.PutUvarint(cell[1:], uint64(msgBytes))

	bytes := copy(cell[1+n:], msg)
	if _, err := c.Write(cell); err != nil {
		return 0, err
	}

	for bytes < msgBytes {
		zeroCell(cell)
		bytes += copy(cell, msg[bytes:])
		if _, err := c.Write(cell); err != nil {
			return bytes, err
		}
	}
	return bytes, nil
}

// ReceiveMessage
func (c *Conn) ReceiveMessage() ([]byte, error) {
	var err error
	cell := make([]byte, CellBytes)
	if _, err = c.Read(cell); err != nil && err != io.EOF {
		return nil, err
	}

	if cell[0] != msgCell {
		return nil, errCellType
	}

	msgBytes, n := binary.Uvarint(cell[1:])
	if msgBytes > MaxMsgBytes {
		return nil, errMsgLength
	}

	msg := make([]byte, msgBytes)
	bytes := copy(msg, cell[1+n:])

	for err != io.EOF && uint64(bytes) < msgBytes {
		if _, err = c.Read(cell); err != nil && err != io.EOF {
			return nil, err
		}
		bytes += copy(msg[bytes:], cell)
	}

	return msg, nil
}

// Write zeros to each byte of a cell.
func zeroCell(cell []byte) {
	for i := 0; i < CellBytes; i++ {
		cell[i] = 0
	}
}
