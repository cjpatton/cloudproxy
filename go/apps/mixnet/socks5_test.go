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

// This program partially implements the server in version 5 of the SOCKS
// protocol version 5 specified in RFC 1928. In particular, it only supports
// TCP clients with no authentication who request CONNECT; neither BIND nor
// UDP ASSOCIATE are supported.

import (
	"testing"

	"golang.org/x/net/proxy"
)

func TestSocks5Serve(t *testing.T) {
	_, err := proxy.SOCKS5("tcp", "localhos:1080", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}
