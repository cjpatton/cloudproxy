//  File: start_hosted_program.cc
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: Invoke linux_tao admin interface to start a program.
//
//  Copyright (c) 2013, Google Inc.  All rights reserved.
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
#include <cstdio>
#include <list>
#include <string>

#include <gflags/gflags.h>
#include <glog/logging.h>

#include "tao/linux_admin_rpc.h"
#include "tao/linux_host.h"
#include "tao/util.h"

using std::list;
using std::string;

using tao::InitializeApp;
using tao::LinuxHost;
using tao::LinuxAdminRPC;

DEFINE_string(host_path, "linux_tao_host", "Location of linux host configuration");

int main(int argc, char **argv) {
  InitializeApp(&argc, &argv, true);

  if (argc == 1) {
    printf("usage: %s [options] -- program args...\n", argv[0]);
    return 1;
  }

  string prog = argv[1];
  list<string> args;
  for (int i = 2; i < argc; i++) {
    args.push_back(string(argv[i]));
  }

  scoped_ptr<LinuxAdminRPC> host(LinuxHost::Connect(FLAGS_host_path));
  CHECK(host.get() != nullptr);
  
  string child_name;
  CHECK(host->StartHostedProgram(prog, args, &child_name));

  printf("Started: %s\n", child_name.c_str());

  return 0;
}
