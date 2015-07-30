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
	"bytes"
	"crypto/x509/pkix"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/jlmucb/cloudproxy/go/tao"
)

var password = make([]byte, 32)
var network = "tcp"
var routerAddr = "localhost:7007"
var dstAddr = "localhost:7009"

var id = pkix.Name{
	Organization: []string{"Mixnet tester"},
}

func makeTrivialDomain(configDir string) (*tao.Domain, error) {
	var policyDomainConfig tao.DomainConfig
	policyDomainConfig.SetDefaults()
	policyDomainConfig.DomainInfo.GuardType = proto.String("AllowAll")
	configPath := path.Join(configDir, "tao.config")
	return tao.CreateDomain(policyDomainConfig, configPath, password)
}

func makeContext(batchSize int) (*RouterContext, *ProxyContext, error) {
	configDir := "/tmp/mixnet_test_domain"
	configPath := path.Join(configDir, "tao.config")

	// Create a domain with a LiberalGuard.
	_, err := makeTrivialDomain(configDir)
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(configDir)

	// Create a SoftTao from the domain.
	st, err := tao.NewSoftTao(configDir, password)
	if err != nil {
		return nil, nil, err
	}

	// Create router context. This loads the domain and binds a
	// socket and an anddress.
	router, err := NewRouterContext(configPath, network, routerAddr, batchSize, &id, st)
	if err != nil {
		return nil, nil, err
	}

	// Create a proxy context. This just loads the domain.
	proxy, err := NewProxyContext(configPath)
	if err != nil {
		router.Close()
		return nil, nil, err
	}

	return router, proxy, nil
}

type testResult struct {
	err error
	msg []byte
}

// Router accepts a connection from a proxy and reads a cell.
func runRouterReadCell(router *RouterContext, ch chan<- testResult) {
	c, err := router.AcceptProxy()
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}

	cell := make([]byte, CellBytes)
	if _, err := c.Read(cell); err != nil {
		ch <- testResult{err, cell}
	} else {
		ch <- testResult{nil, cell}
	}
}

// Proxy dials a router and sends a cell.
func runProxyWriteCell(proxy *ProxyContext, msg []byte) error {
	c, err := proxy.DialRouter(network, routerAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	if _, err := c.Write(msg); err != nil {
		return err
	}

	return nil
}

// Router accepts a connection from a proxy and handles a number of
// requests.
func runRouterHandleProxy(router *RouterContext, requestCount int, ch chan<- testResult) {
	c, err := router.AcceptProxy()
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}
	defer c.Close()

	for i := 0; i < requestCount; i++ {
		if err = router.HandleProxy(c); err != nil {
			ch <- testResult{err, []byte{}}
			return
		}
	}

	ch <- testResult{nil, []byte{}}
}

// Proxy dials a router, creates a circuit, and sends a message over
// the circuit.
func runProxySendMessage(proxy *ProxyContext, msg []byte) error {
	c, err := proxy.DialRouter(network, routerAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	if _, err = proxy.CreateCircuit(c, []string{dstAddr}); err != nil {
		return err
	}

	if _, err = c.SendMessage(msg); err != nil {
		return err
	}

	return nil
}

func runProxyReceiveMessage(proxy *ProxyContext) ([]byte, error) {
	c, err := proxy.DialRouter(network, routerAddr)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if _, err = proxy.CreateCircuit(c, []string{dstAddr}); err != nil {
		return nil, err
	}

	if _, err = c.ReceiveDirective(); err != nil {
		return nil, err
	}

	if _, err = c.SendDirective(dirAwaitMsg); err != nil {
	}

	msg, err := c.ReceiveMessage()
	if err != nil {
		return nil, err
	}
	fmt.Println("got here")
	return msg, nil
}

// A dummy server that waits for a message from the a client.
func runDestination(ch chan<- testResult) {
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

	buff := make([]byte, CellBytes*10)
	bytes, err := c.Read(buff)
	if err != nil {
		ch <- testResult{err, []byte{}}
		return
	}

	ch <- testResult{nil, buff[:bytes]}
	return
}

// Test connection set up.
func TestProxyRouterConnect(t *testing.T) {
	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()

	// Wait for a connection from the proxy.
	ch := make(chan bool)
	go func(ch chan<- bool) {
		router.AcceptProxy()
		ch <- true
	}(ch)

	c, err := proxy.DialRouter(network, routerAddr)
	if err != nil {
		router.Close()
		t.Fatal(err)
	}
	defer c.Close()

	<-ch
}

// Test sending a cell.
func TestProxyRouterCell(t *testing.T) {
	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	ch := make(chan testResult)

	msg := make([]byte, CellBytes+1)
	for i := 0; i < len(msg); i++ {
		msg[i] = byte(i)
	}

	// The cell is just right.
	go runRouterReadCell(router, ch)
	if err = runProxyWriteCell(proxy, msg[:CellBytes]); err != nil {
		t.Error(err)
	}
	res := <-ch
	if res.err != nil && res.err != io.EOF {
		t.Error(res.err)
	} else if bytes.Compare(res.msg, msg[:CellBytes]) != 0 {
		t.Errorf("Server got: %s", res.msg)
	}

	// This cell is too big.
	go runRouterReadCell(router, ch)
	if err = runProxyWriteCell(proxy, msg); err != errCellLength {
		t.Error("runProxyWriteCell(): should have returned errCellLength")
	}
	res = <-ch
	if res.err != io.EOF {
		t.Error("runRouterReadCell(): should have returned EOF.")
	}
}

// Test setting up a circuit and relay a message to destination. Try
// messages of various lengths.
func TestProxyRouterRelay(t *testing.T) {
	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	routerCh := make(chan testResult)
	dstCh := make(chan testResult)

	// Create a long message.
	msg := make([]byte, (CellBytes*5)+237)
	for i := 0; i < len(msg); i++ {
		msg[i] = byte(i)
	}
	var res testResult

	trials := []int{
		37,        // A short message
		CellBytes, // A cell
		len(msg),  // A long message
	}

	go runDummyServer(len(trials), 1, dstCh, receiver)

	for _, l := range trials {

		go runRouterHandleProxy(router, 2, routerCh)
		if err = runProxySendMessage(proxy, msg[:l]); err != nil {
			t.Errorf(">relay (length=%d): %s", l, err)
		}

		res = <-routerCh
		if res.err != nil {
			t.Errorf(">relay (length=%d): %s", l, res.err)
		}

		res = <-dstCh
		if res.err != nil {
			t.Error(res.err)
		} else if bytes.Compare(res.msg, msg[:l]) != 0 {
			t.Errorf(">relay (length=%d): Server got: %s", l, res.msg)
		}
	}
}

// Test setting up a circuit and waiting for a message from the
// destination.
func TestRouterProxyRelay(t *testing.T) {
	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	routerCh := make(chan testResult)
	dstCh := make(chan testResult)
	var res testResult

	go runDummyServer(1, 1, dstCh, sender)
	go runRouterHandleProxy(router, 2, routerCh)

	msg, err := runProxyReceiveMessage(proxy)
	if err != nil {
		t.Errorf("<relay: %s", err)
	}

	res = <-routerCh
	if res.err != nil {
		t.Errorf("<relay: %s", res.err)
	}

	res = <-dstCh
	if res.err != nil {
		t.Errorf("<relay: %s", res.err)
	}
	t.Log(string(msg))

	// TODO(cjpatton) test a longer message.
}

// Test sending malformed messages from the proxy to the router.
func TestMaliciousProxyRouterRelay(t *testing.T) {
	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	ch := make(chan testResult)

	cell := make([]byte, CellBytes)

	// Unrecognized cell type.
	cell[0] = dirCell ^ msgCell ^ relayCell
	go runRouterHandleProxy(router, 1, ch)
	c, err := proxy.DialRouter(network, routerAddr)
	if err != nil {
		t.Error(err)
	}
	if _, err = c.Write(cell); err != nil {
		t.Error(err)
	}
	_, err = c.ReceiveDirective()
	if err == nil || err.Error() != "router error: "+errBadCellType.Error() {
		t.Error("Bad cell, got incorrect error: ", err)
	}
	c.Close()
	<-ch

	// Message too long.
	cell[0] = msgCell
	binary.PutUvarint(cell[1:], uint64(MaxMsgBytes+1))
	go runRouterHandleProxy(router, 1, ch)
	c, err = proxy.DialRouter(network, routerAddr)
	if err != nil {
		t.Error(err)
	}
	if _, err = c.Write(cell); err != nil {
		t.Error(err)
	}
	_, err = c.ReceiveDirective()
	if err == nil || err.Error() != "router error: "+errMsgLength.Error()+" (connection closed)" {
		t.Error("Long message, got incorrect error: ", err)
	}
	c.Close()
	<-ch
}
