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

package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/jlmucb/cloudproxy/go/apps/mixnet"
)

// Command line arguments.
var proxyAddr = flag.String("proxy_addr", "localhost:8123", "Address and port for the Tao-delegated mixnet router.")
var routerAddr = flag.String("router_addr", "localhost:8123", "Address and port for the Tao-delegated mixnet router.")
var network = flag.String("network", "tcp", "Network protocol for the mixnet proxy and router.")
var configPath = flag.String("config", "tao.config", "Path to domain configuration file.")
var timeoutDuration = flag.String("timeout", "10s", "Timeout on TCP connections, e.g. \"10s\".")

func main() {
	flag.Parse()
	timeout, err := time.ParseDuration(*timeoutDuration)
	if err != nil {
		glog.Fatalf("failed to parse timeout duration: %s", err)
	}

	proxy, err := mixnet.NewProxyContext(*configPath, *network, timeout)
	if err != nil {
		glog.Fatalf("failed to configure proxy: %s", err)
	}

	c, err := proxy.CreateCircuit(*routerAddr, "localhost:8080")
	if err != nil {
		glog.Fatalf("CreateCircuit(): %s", err)
	}
	defer c.Close()

	if err = proxy.SendMessage(c, []byte("Hello!!!")); err != nil {
		glog.Errorf("SendMessage(): %s", err)
	}

	reply, err := proxy.ReceiveMessage(c)
	if err != nil {
		glog.Errorf("ReceiveMessage(): %s", err)
	} else {
		glog.Info("Got: ", string(reply))
	}

	if err = proxy.DestroyCircuit(c); err != nil {
		glog.Error("DestroyCircuit(): %s", err)
	}

	glog.Flush()
}
