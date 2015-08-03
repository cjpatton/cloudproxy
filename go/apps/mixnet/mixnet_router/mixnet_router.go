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
	"crypto/x509/pkix"
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/jlmucb/cloudproxy/go/apps/mixnet"
	"github.com/jlmucb/cloudproxy/go/tao"
)

// Run mixnet router service for mixnet clients.
// TODO(cjpatton) how to handle interrupts so that defers's are called? Tom:
// signal.Notify allows you to add new signal handlers for a given signal
// (without removing the old ones). So, you could wrap a deferred function in a
// signal handler to make sure it gets called (almost) no matter what.
func serveMixnetClients(hp *mixnet.RouterContext) error {

	for {
		c, err := hp.AcceptProxy()
		if err != nil {
			return err
		}

		go func(c *mixnet.Conn) {
			defer c.Close()
			for {
				if err := hp.HandleProxy(c); err != nil {
					glog.Errorf("error while serving client no. %d: %s", c.GetID(), err)
					break
				}
			}
		}(c)
	}
	return nil
}

// Command line arguments.
var routerAddr = flag.String("addr", "localhost:8123", "Address and port for the Tao-delegated mixnet router.")
var routerNetwork = flag.String("network", "tcp", "Network protocol for the Tao-delegated mixnet router.")
var configPath = flag.String("config", "tao.config", "Path to domain configuration file.")
var batchSize = flag.Int("batch", 1, "Number of senders in a batch.")
var timeoutDuration = flag.String("timeout", "10s", "Timeout on TCP connections, e.g. \"10s\".")

// x509 identity of the mixnet router.
var x509Identity pkix.Name = pkix.Name{
	Organization:       []string{"Google Inc."},
	OrganizationalUnit: []string{"Cloud Security"},
}

func main() {
	flag.Parse()
	timeout, err := time.ParseDuration(*timeoutDuration)
	if err != nil {
		glog.Errorf("failed to parse timeout duration: %s", err)
	}

	hp, err := mixnet.NewRouterContext(*configPath, *routerNetwork, *routerAddr, *batchSize,
		timeout, &x509Identity, tao.Parent())
	if err != nil {
		glog.Errorf("failed to configure server: %s", err)
	}
	defer hp.Close()

	if err = serveMixnetClients(hp); err != nil {
		glog.Errorf("error while serving: %s", err)
	}

	glog.Flush()
}
