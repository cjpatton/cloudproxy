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
	"fmt"
	"io"
	"strconv"
	"testing"

	netproxy "golang.org/x/net/proxy"
)

// Run proxy server.
func runSocksServerOne(proxy *ProxyContext, ch chan<- testResult) {
	c, addr, err := proxy.Accept()
	if err != nil {
		ch <- testResult{err, nil}
		return
	}
	defer c.Close()

	d, err := proxy.CreateCircuit(routerAddr, addr)
	if err != nil {
		ch <- testResult{err, nil}
		return
	}

	if err = proxy.HandleClient(c, d); err != nil {
		ch <- testResult{err, nil}
		return
	}

	if err = proxy.DestroyCircuit(d); err != nil {
		ch <- testResult{err, nil}
		return
	}

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

	if _, err = c.Write(msg); err != nil {
		return testResult{err, nil}
	}

	bytes, err := c.Read(msg)
	if err != nil {
		return testResult{err, nil}
	}

	return testResult{nil, msg[:bytes]}
}

// Test the SOCKS proxy server.
func TestSocks(t *testing.T) {

	proxy, err := makeProxyContext(proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	ch := make(chan testResult)
	go func() {
		c, addr, err := proxy.Accept()
		if err != nil {
			ch <- testResult{err, nil}
			return
		}
		c.Close()
		ch <- testResult{nil, []byte(addr)}
	}()

	dialer, err := netproxy.SOCKS5(network, proxyAddr, nil, netproxy.Direct)
	if err != nil {
		t.Error(err)
	}

	c, err := dialer.Dial(network, dstAddr)
	if err != nil {
		t.Error(err)
	}
	c.Close()

	res := <-ch
	if res.err != nil {
		t.Error(res.err)
	} else {
		t.Log("server got:", string(res.msg))
	}
}

// Test mixnet end-to-end with many clients. Proxy a protocol through mixnet.
// The client sends the server a message and the server echoes it back.
func TestMixnet(t *testing.T) {

	clientCt := 20
	router, proxy, err := makeContext(clientCt)
	if err != nil {
		t.Fatal(err)
	}
	proxy.Close()
	defer router.Close()

	var res testResult
	clientCh := make(chan testResult)
	proxyCh := make(chan testResult)
	routerCh := make(chan testResult)
	dstCh := make(chan testResult)

	go runRouterHandleProxy(router, clientCt, 3, routerCh)
	go runDummyServer(clientCt, 1, dstCh)

	for i := 0; i < clientCt; i++ {
		go func(pid int, ch chan<- testResult) {
			proxyAddr := "127.0.0.1:" + strconv.Itoa(1080+pid)
			proxy, err = makeProxyContext(proxyAddr)
			if err != nil {
				ch <- testResult{err, nil}
				return
			}
			defer proxy.Close()
			go runSocksServerOne(proxy, proxyCh)

			msg := []byte(fmt.Sprintf("Hello, my name is %d", pid))
			ch <- runSocksClient(proxyAddr, msg)

		}(i, clientCh)
	}

	// Wait for clients to finish.
	for i := 0; i < clientCt; i++ {
		res = <-clientCh
		if res.err != nil {
			t.Error(res.err)
		} else {
			t.Log("client got:", string(res.msg))
		}
	}

	// Wait for proxies to finish.
	for i := 0; i < clientCt; i++ {
		res = <-proxyCh
		if res.err != nil {
			t.Error(res.err)
		}
	}

	// Wait for server to finish.
	for i := 0; i < clientCt; i++ {
		res = <-dstCh
		if res.err != nil {
			t.Error(res.err)
		}
	}

	// Wait for router to finish.
	for i := 0; i < clientCt; i++ {
		res = <-routerCh
		if res.err != nil && res.err != io.EOF {
			t.Error(res.err)
		}
	}
}
