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
	"bytes"
	"io"
	"testing"

	netproxy "golang.org/x/net/proxy"
)

// Run proxy server.
func runSocksServer(proxy *ProxyContext, ch chan<- testResult) {
	c, addr, err := proxy.Accept()
	if err != nil {
		ch <- testResult{err, nil}
		return
	}

	defer c.Close()

	err = proxy.ServeClient(c, routerAddr, addr)
	ch <- testResult{err, []byte(addr)}
}

// Connect to a destination through a mixnet proxy, send a message,
// and wait for a response.
func runSocksClient(msg []byte) testResult {
	dialer, err := netproxy.SOCKS5(network, proxyAddr, nil, netproxy.Direct)
	if err != nil {
		return testResult{err, nil}
	}

	c, err := dialer.Dial(network, dstAddr)
	if err != nil {
		return testResult{err, nil}
	}
	defer c.Close()

	if _, err = c.Write(msg); err != nil {
		return testResult{err, nil}
	}

	reply := make([]byte, MaxMsgBytes)
	bytes, err := c.Read(reply)
	if err != nil {
		return testResult{err, nil}
	}

	// Just for fun, try again.
	if _, err = c.Write(msg); err != nil {
		return testResult{err, nil}
	}

	reply = make([]byte, MaxMsgBytes)
	bytes, err = c.Read(reply)
	if err != nil {
		return testResult{err, nil}
	}

	return testResult{nil, reply[:bytes]}
}

// Test mixnet end-to-end with one client. Proxy a protocl through mixnet. The
// client sends the server a message and the server echoes it back; repeat.
func TestMixnetOne(t *testing.T) {

	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	defer proxy.Close()

	proxyCh := make(chan testResult)
	routerCh := make(chan testResult)
	dstCh := make(chan testResult)
	var res testResult

	go runSocksServer(proxy, proxyCh)
	go runRouterHandleOneProxy(router, 4, routerCh)
	go runDummyServer(1, 2, dstCh)

	msg := []byte("Who am I?")
	if res := runSocksClient(msg); res.err != nil {
		t.Error(res.err)
	} else if bytes.Compare(msg, res.msg) != 0 {
		t.Error("received message different from sent")
	}

	res = <-routerCh
	if res.err != nil && res.err != io.EOF {
		t.Error(res.err)
	}

	res = <-proxyCh
	if res.err != nil {
		t.Error(res.err)
	}
}
