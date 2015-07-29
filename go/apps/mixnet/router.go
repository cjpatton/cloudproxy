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
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/jlmucb/cloudproxy/go/tao"
)

// RouterContext stores the runtime environment for a Tao-delegated router.
type RouterContext struct {
	keys          *tao.Keys    // Signing keys of this hosted program.
	domain        *tao.Domain  // Policy guard and public key.
	proxyListener net.Listener // Socket where server listens for proxies.

	id uint64 // Next serial identifier that will be assigned to a connection.

	// Data structures for queueing and batching messages from sender to
	// recipient and recipient to sender respectively.
	nextQueue *Queue
	prevQueue *Queue

	// The queues and error handlers are instantiated as go routines; these
	// channels are for managing them.
	killQueue             chan bool
	killQueueErrorHandler chan bool

	network string // Network protocol, e.g. "tcp"
}

// NewRouterContext generates new keys, loads a local domain configuration from
// path and binds an anonymous listener socket to addr on network
// network. A delegation is requested from the Tao t which is  nominally
// the parent of this hosted program.
func NewRouterContext(path, network, addr string, batchSize int, x509Identity *pkix.Name, t tao.Tao) (hp *RouterContext, err error) {
	hp = new(RouterContext)

	hp.network = network

	// Generate keys and get attestation from parent.
	if hp.keys, err = tao.NewTemporaryTaoDelegatedKeys(tao.Signing|tao.Crypting, t); err != nil {
		return nil, err
	}

	// Create a certificate.
	if hp.keys.Cert, err = hp.keys.SigningKey.CreateSelfSignedX509(x509Identity); err != nil {
		return nil, err
	}

	// Load domain from local configuration.
	if hp.domain, err = tao.LoadDomain(path, nil); err != nil {
		return nil, err
	}

	// Encode TLS certificate.
	cert, err := tao.EncodeTLSCert(hp.keys)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:            x509.NewCertPool(),
		Certificates:       []tls.Certificate{*cert},
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
	}

	// Bind address to socket.
	if hp.proxyListener, err = tao.ListenAnonymous(network, addr, tlsConfig,
		hp.domain.Guard, hp.domain.Keys.VerifyingKey, hp.keys.Delegation); err != nil {
		return nil, err
	}

	// Instantiate the queues.
	hp.nextQueue = NewQueue(network, batchSize)
	hp.prevQueue = NewQueue(network, batchSize)
	hp.killQueue = make(chan bool)
	hp.killQueueErrorHandler = make(chan bool)
	go hp.nextQueue.DoQueue(hp.killQueue)
	go hp.prevQueue.DoQueue(hp.killQueue)
	go hp.nextQueue.DoQueueErrorHandler(hp.killQueueErrorHandler)
	go hp.prevQueue.DoQueueErrorHandler(hp.killQueueErrorHandler)

	return hp, nil
}

// AcceptProxy Waits for connectons from proxies.
func (hp *RouterContext) AcceptProxy() (*Conn, error) {
	c, err := hp.proxyListener.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{c, hp.nextID()}, nil
}

// Close releases any resources held by the hosted program.
func (hp *RouterContext) Close() {
	hp.killQueue <- true
	hp.killQueue <- true
	hp.killQueueErrorHandler <- true
	hp.killQueueErrorHandler <- true
	if hp.proxyListener != nil {
		hp.proxyListener.Close()
	}
}

// HandleProxy reads a directive or a message from a proxy.
func (hp *RouterContext) HandleProxy(c *Conn) error {
	var err error
	cell := make([]byte, CellBytes)
	if _, err = c.Read(cell); err != nil && err != io.EOF {
		return err
	}

	if cell[0] == msgCell { // Handle a message.
		msgBytes, n := binary.Uvarint(cell[1:])
		if msgBytes > MaxMsgBytes {
			if _, err = hp.SendFatal(c, errMsgLength); err != nil {
				return err
			}
			// TODO(cjpatton) send DESTROY, close connection to next hop.
			c.Close()
			return nil
		}

		msg := make([]byte, msgBytes)
		bytes := copy(msg, cell[1+n:])

		// While the connection is open and the message is incomplete, read
		// the next cell.
		for err != io.EOF && uint64(bytes) < msgBytes {
			if _, err = c.Read(cell); err != nil && err != io.EOF {
				return err
			}
			bytes += copy(msg[bytes:], cell)
		}

		q := new(Queueable)
		q.id = c.id
		q.msg = msg
		hp.nextQueue.Enqueue(q)

	} else if cell[0] == dirCell { // Handle a directive.
		dirBytes, n := binary.Uvarint(cell[1:])
		var d Directive
		if err := proto.Unmarshal(cell[1+n:1+n+int(dirBytes)], &d); err != nil {
			return err
		}

		if *d.Type == DirectiveType_CREATE {
			// Construct a circuit and send a CREATED directive to sender to
			// confirm. For now, only single hop circuits are supported.
			if len(d.Addrs) == 0 {
				return errBadDirective
			}
			if len(d.Addrs) > 1 {
				return errors.New("multi-hop circuits not implemented")
			}

			q := new(Queueable)
			q.id = c.id
			q.addr = &d.Addrs[0]
			hp.nextQueue.Enqueue(q)

			_, err = c.SendDirective(dirCreated)
			return err

		} else if *d.Type == DirectiveType_AWAIT_MSG {
			// Wait for a message from the destination, divide it into cells,
			// and reply to sender.
			q := new(Queueable)
			q.id = c.id
			q.reply = make(chan []byte)
			hp.nextQueue.Enqueue(q)

			msg := <-q.reply
			if msg != nil {
				// TODO(cjpatton) add cells to prevQueue.
			}
		}

	} else { // Unknown cell type, return an error.
		if _, err = hp.SendError(c, errBadCellType); err != nil {
			return err
		}
	}

	return nil
}

// SendError sends an error message to a client.
func (hp *RouterContext) SendError(c *Conn, err error) (int, error) {
	var d Directive
	d.Type = DirectiveType_ERROR.Enum()
	d.Error = proto.String(err.Error())
	return c.SendDirective(&d)
}

// SendFatal sends an error message and signals the client that the connection
// was severed on the server side.
func (hp *RouterContext) SendFatal(c *Conn, err error) (int, error) {
	var d Directive
	d.Type = DirectiveType_FATAL.Enum()
	d.Error = proto.String(err.Error())
	return c.SendDirective(&d)
}

// Get the next Id to assign and increment counter.
// TODO(cjpatton) This will need mutual exclusion when AcceptRouter() is
// implemented.
func (hp *RouterContext) nextID() (id uint64) {
	id = hp.id
	hp.id++
	return id
}
