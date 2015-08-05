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
	"testing"
	"time"

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
func runSocksClient(proxyAddr string, msg []byte) testResult {
	dialer, err := netproxy.SOCKS5(network, proxyAddr, nil, netproxy.Direct)
	if err != nil {
		return testResult{err, nil}
	}

	c, err := dialer.Dial(network, dstAddr)
	if err != nil {
		return testResult{err, nil}
	}
	defer c.Close()

	c.SetDeadline(time.Now().Add(timeout))
	if _, err = c.Write(msg); err != nil {
		return testResult{err, nil}
	}

	reply := make([]byte, MaxMsgBytes)
	c.SetDeadline(time.Now().Add(timeout))
	bytes, err := c.Read(reply)
	if err != nil {
		return testResult{err, nil}
	}

	return testResult{nil, reply[:bytes]}
}

// Test mixnet end-to-end with many clients. Proxy a protocol through mixnet.
// The client sends the server a message and the server echoes it back.
func TestMixnet(t *testing.T) {

	//TODO(cjpatton) rewrite!!
}
