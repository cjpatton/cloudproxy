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

	netproxy "golang.org/x/net/proxy"
)

func runSocksServer(proxy *ProxyContext, ch chan<- testResult) {
	c, addr, err := proxy.Accept()
	if err != nil {
		ch <- testResult{err, nil}
		return
	}
	defer c.Close()
	ch <- testResult{nil, []byte(addr)}
}

func runSocksClient() error {
	dialer, err := netproxy.SOCKS5(network, proxyAddr, nil, netproxy.Direct)
	if err != nil {
		return err
	}

	c, err := dialer.Dial(network, dstAddr)
	if err != nil {
		return err
	}
	c.Close()
	return nil
}

func TestSocksServe(t *testing.T) {

	router, proxy, err := makeContext(1)
	if err != nil {
		t.Fatal(err)
	}
	defer router.Close()
	defer proxy.Close()

	ch := make(chan testResult)
	go runSocksServer(proxy, ch)

	if err = runSocksClient(); err != nil {
		router.Close()
		t.Fatal(err)
	}

	t.Log("destination:", string((<-ch).msg))
}
