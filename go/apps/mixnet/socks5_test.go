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
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"testing"
	"time"

	netproxy "golang.org/x/net/proxy"
)

var _ = fmt.Println

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

func TestSocksProxy(t *testing.T) {
	buf := make([]byte, 1000)

	// Start SOCKS server.
	ch := make(chan testResult)
	proxy, err := makeProxyContext(proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	go runSocksServer(proxy, ch)

	// Start server.
	ncServe := exec.Command("nc", "-l", "8080")
	out, err := ncServe.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := ncServe.Start(); err != nil {
		t.Fatal(err)
	}

	// Sart client.
	ncClient := exec.Command("nc", "localhost", "8080", "-X", "5", "-x", "localhost:1080")
	in, err := ncClient.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	cout, err := ncClient.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	_, err = in.Write([]byte("Hello"))
	if err != nil {
		t.Fatal(err)
	}

	if err := ncClient.Start(); err != nil {
		t.Error(err)
	}

	bytes, err := cout.Read(buf)
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	t.Log("client got:", string(buf[:bytes]))

	if err = ncClient.Wait(); err != nil {
		t.Fatal(err)
	}

	bytes, err = out.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	t.Log("server got:", string(buf[:bytes]))

	if err = ncServe.Wait(); err != nil {
		t.Fatal(err)
	}

	<-ch
}

// Test mixnet end-to-end many clients. Proxy a protocol through mixnet. The
// client sends the server a message and the server echoes it back.
func TestMixnet(t *testing.T) {
	clientCt := 20

	router, proxy, err := makeContext(clientCt)
	if err != nil {
		t.Fatal(err)
	}
	proxy.Close()
	defer router.Close()

	routerCh := make(chan error)
	dstCh := make(chan testResult)
	var res testResult

	// Run RouterContext.HanleProxy four times (CreateCircuit(), SendMessage(),
	// SendMessage(), and DestroyCircuit()) for each client.
	go runRouterHandleProxy(router, clientCt, 2, routerCh)

	// Echo two messages per client.
	go runDummyServer(clientCt, 2, dstCh)

	// Spawn a client/proxy for each test.
	ch := make(chan error)
	for i := 0; i < clientCt; i++ {
		go func(i int, ch chan<- error) {

			// Create a ProxyContext and bind a proxy to a port for each client.
			proxyCh := make(chan testResult)
			proxyAddr := "127.0.0.1:" + strconv.Itoa(1080+i)
			proxy, err = makeProxyContext(proxyAddr)
			if err != nil {
				ch <- err
				return
			}
			defer proxy.Close()

			// Run SOCKS5 proxy.
			go runSocksServer(proxy, proxyCh)

			// Run client.
			msg := []byte(fmt.Sprintf("Who am I? I am %d.", i))
			if res := runSocksClient(proxyAddr, msg); res.err != nil {
				ch <- res.err
				return
			} else if bytes.Compare(msg, res.msg) != 0 {
				ch <- errors.New("received message different from sent")
				return
			}

			// Wait for proxy to finish.
			res = <-proxyCh
			if res.err != nil {
				ch <- res.err
				return
			}
			ch <- nil
		}(i, ch)
	}

	// Wait for all client/proxy routines to finish.
	for i := 0; i < clientCt; i++ {
		select {
		case err = <-ch:
			if err != nil {
				t.Error(err)
			}
		}
	}
}
