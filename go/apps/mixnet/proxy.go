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
	"strconv"
	"time"

	"github.com/jlmucb/cloudproxy/go/tao"
)

const SocksVersion = 0x05

// ProxyContext stores the runtime environment for a mixnet proxy. A mixnet
// proxy connects to a mixnet router on behalf of a client's application.
type ProxyContext struct {
	domain   *tao.Domain  // Policy guard and public key.
	listener net.Listener // SOCKS5 server for listening to clients.

	id uint64 // Next serial identifier that will assigned to a new connection.

	network string        // Network protocol, e.g. "tcp".
	timeout time.Duration // Timeout on read.
}

// NewProxyContext loads a domain from a local configuration.
func NewProxyContext(path, network, addr string, timeout time.Duration) (p *ProxyContext, err error) {
	p = new(ProxyContext)
	p.network = network
	p.timeout = timeout

	// Load domain from a local configuration.
	if p.domain, err = tao.LoadDomain(path, nil); err != nil {
		return nil, err
	}

	if p.listener, err = net.Listen(network, addr); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *ProxyContext) Close() {
	if p.listener != nil {
		p.listener.Close()
	}
}

// DialRouter connects anonymously to a remote Tao-delegated mixnet router.
func (p *ProxyContext) DialRouter(network, addr string) (*Conn, error) {
	c, err := tao.Dial(network, addr, p.domain.Guard, p.domain.Keys.VerifyingKey, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{c, p.nextID()}, nil
}

// SendDirective serializes and pads a directive to the length of a cell and
// sends it to the peer. A directive is signaled to the receiver by the first
// byte of the cell. The next few bytes encode the length of of the serialized
// protocol buffer. If the buffer doesn't fit in a cell, then throw an error.
func (p *ProxyContext) SendDirective(c *Conn, d *Directive) (int, error) {
	cell, err := marshalDirective(d)
	if err != nil {
		return 0, err
	}
	return c.Write(cell)
}

// ReceiveDirective awaits a reply from the peer and returns the directive
// received, e.g. in response to RouterContext.HandleProxy(). If the directive
// type is ERROR, return an error.
func (p *ProxyContext) ReceiveDirective(c *Conn, d *Directive) (int, error) {
	cell := make([]byte, CellBytes)
	bytes, err := c.Read(cell)
	if err != nil && err != io.EOF {
		return 0, err
	}

	err = unmarshalDirective(cell, d)
	if err != nil {
		return 0, err
	}

	if *d.Type == DirectiveType_ERROR {
		return bytes, errors.New("router error: " + (*d.Error))
	}
	return bytes, nil
}

// CreateCircuit connects anonymously to a remote Tao-delegated mixnet router
// specified by addrs[0]. It directs the router to construct a circuit to a
// particular destination over the mixnet specified by addrs[len(addrs)-1].
func (p *ProxyContext) CreateCircuit(addrs ...string) (*Conn, error) {
	c, err := p.DialRouter(p.network, addrs[0])
	if err != nil {
		return nil, err
	}

	d := &Directive{
		Type:  DirectiveType_CREATE.Enum(),
		Addrs: addrs[1:],
	}

	// Send CREATE directive to router.
	if _, err := p.SendDirective(c, d); err != nil {
		return c, err
	}

	// Wait for CREATED directive from router.
	c.SetReadDeadline(time.Now().Add(p.timeout))
	if _, err := p.ReceiveDirective(c, d); err != nil {
		return c, err
	} else if *d.Type != DirectiveType_CREATED {
		return c, errors.New("could not create circuit")
	}

	return c, nil
}

// DestroyCircuit directs the router to close the connection to the destination
// and destroy the circuit then closes the connection. TODO(cjpatton) in order
// to support multi-hop circuits, this code will need to wait for a DESTROYED
// directive from the first hop.
func (p *ProxyContext) DestroyCircuit(c *Conn) error {
	// Send DESTROY directive to router.
	if _, err := p.SendDirective(c, dirDestroy); err != nil {
		return err
	}
	c.Close()
	return nil
}

// SendMessage divides a message into cells and sends each cell over the network
// connection. A message is signaled to the receiver by the first byte of the
// first cell. The next few bytes encode the total number of bytes in the
// message.
func (p *ProxyContext) SendMessage(c *Conn, msg []byte) error {
	msgBytes := len(msg)
	cell := make([]byte, CellBytes)
	cell[0] = msgCell
	n := binary.PutUvarint(cell[1:], uint64(msgBytes))

	bytes := copy(cell[1+n:], msg)
	if _, err := c.Write(cell); err != nil {
		return err
	}

	for bytes < msgBytes {
		zeroCell(cell)
		cell[0] = msgCell
		bytes += copy(cell[1:], msg[bytes:])
		if _, err := c.Write(cell); err != nil {
			return err
		}
	}
	return nil
}

// ReceiveMessage reads message cells from the router and assembles them into
// a messsage.
func (p *ProxyContext) ReceiveMessage(c *Conn) ([]byte, error) {
	var err error

	// Receive cells from router.
	cell := make([]byte, CellBytes)
	c.SetReadDeadline(time.Now().Add(p.timeout))
	if _, err = c.Read(cell); err != nil && err != io.EOF {
		return nil, err
	}

	if cell[0] == dirCell {
		var d Directive
		if err = unmarshalDirective(cell, &d); err != nil {
			return nil, err
		}
		if *d.Type == DirectiveType_ERROR {
			return nil, errors.New("router error: " + (*d.Error))
		}
		return nil, errCellType
	} else if cell[0] != msgCell {
		return nil, errCellType
	}

	msgBytes, n := binary.Uvarint(cell[1:])
	if msgBytes > MaxMsgBytes {
		return nil, errMsgLength
	}

	msg := make([]byte, msgBytes)
	bytes := copy(msg, cell[1+n:])

	for err != io.EOF && uint64(bytes) < msgBytes {
		c.SetReadDeadline(time.Now().Add(p.timeout))
		if _, err = c.Read(cell); err != nil && err != io.EOF {
			return nil, err
		}
		if cell[0] != msgCell {
			return nil, errCellType
		}
		bytes += copy(msg[bytes:], cell[1:])
	}

	return msg, nil
}

// Return the next serial identifier.
func (p *ProxyContext) nextID() (id uint64) {
	id = p.id
	p.id++
	return id
}

// Accept partially implements the server role in version 5 of the SOCKS
// protocol specified in RFC 1928. In particular, it only supports TCP clients
// with no authentication who request CONNECT to IPv4 addresses; neither BIND
// nor UDP ASSOCIATE are supported. Return the connection to the client and the
// requested destination address.
func (p *ProxyContext) Accept() (net.Conn, string, error) {
	c, err := p.listener.Accept()
	if err != nil {
		return nil, "", err
	}

	// First, wait for greeting from client containing the SOCKS version and
	// requested methods.
	buf := make([]byte, CellBytes*4)
	if _, err = c.Read(buf); err != nil {
		c.Close()
		return nil, "", err
	}

	// Only the NO AUTHENTICATION REQUIRED method is allowed. Note that this
	// makes the server non-complient since GSSAPI is not allowed.
	ver := int(buf[0])
	nmethods := int(buf[1])
	ok := false
	for _, method := range buf[2 : 2+nmethods] {
		if method == 0x00 {
			ok = true
		}
	}

	// Second, reply with selected method.
	if ver == SocksVersion && ok {
		buf[1] = 0x00 // NO AUTHENTICATION REQUIRED
	} else {
		buf[1] = 0xff // NO ACCEPTABLE METHODS
	}

	if _, err = c.Write(buf[:2]); err != nil {
		c.Close()
		return nil, "", err
	}

	// If NO ACCEPTBALE METHOD, the client closes the connection.
	if buf[1] != 0x00 {
		c.Close()
		return nil, "", errors.New("socks: client did not provide acceptable method")
	}

	// Third, wait for command from client.
	bytes, err := c.Read(buf)
	if err != nil {
		c.Close()
		return nil, "", err
	}
	ver = int(buf[0])
	cmd := buf[1]
	atyp := buf[3]

	// Only CONNECT to IPv4 addresses is allowed. Since traffic will be proxied over
	// the mixnet, don't connect to the intended host just yet; call CreateCircuit().
	if ver == SocksVersion && cmd == 0x01 /* CONNECT */ && atyp == 0x01 /* IPv4 */ {
		buf[1] = 0x00 // SUCCEEDED.
	} else {
		buf[2] = 0x07 // COMMAND NOT SUPPORTED
	}
	if _, err = c.Write(buf[:bytes]); err != nil {
		c.Close()
		return nil, "", err
	}

	// dstAddr specifies the destination of the client. At this point the
	// proxy is ready to construct a circuit and relay a message on behalf of
	// the client.
	port := strconv.Itoa((int(buf[bytes-2]) << 8) + int(buf[bytes-1]))
	dstAddr := strconv.Itoa(int(buf[4])) + "." +
		strconv.Itoa(int(buf[5])) + "." +
		strconv.Itoa(int(buf[6])) + "." +
		strconv.Itoa(int(buf[7])) + ":" + port

	return c, dstAddr, nil
}
