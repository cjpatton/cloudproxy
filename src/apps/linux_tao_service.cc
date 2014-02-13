//  File: linux_tao_service.cc
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: The Tao for Linux, implemented over a TPM and creating child
//  processes.
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

#include <fstream>
#include <sstream>
#include <string>

#include <gflags/gflags.h>
#include <glog/logging.h>
#include <keyczar/keyczar.h>

#include "tao/direct_tao_child_channel.h"
#include "tao/fake_tao.h"
#include "tao/linux_tao.h"
#include "tao/pipe_tao_channel.h"
#include "tao/process_factory.h"
#include "tao/tao_domain.h"
#include "tao/tpm_tao_child_channel.h"

using std::ifstream;
using std::string;
using std::stringstream;

using tao::DirectTaoChildChannel;
using tao::FakeTao;
using tao::LinuxTao;
using tao::PipeTaoChannel;
using tao::ProcessFactory;
using tao::TPMTaoChildChannel;
using tao::Tao;
using tao::TaoChildChannel;
using tao::TaoDomain;

DEFINE_string(config_path, "tao.config", "Location of tao configuration");
DEFINE_string(keys_path, "linux_tao_keys", "Location of linux tao keys");
DEFINE_string(program_socket, "/tmp/.linux_tao_socket",
              "File socket for incoming program creation requests");
DEFINE_string(stop_socket, "/tmp/.linux_tao_stop_socket",
              "File socket for stopping the server");

DEFINE_string(aik_blob, "tpm/aikblob", "The AIK blob from the TPM");
DEFINE_string(aik_attestation, "tpm/aik.attest",
              "The attestation to the AIK by the policy key");

// Flags that can be used to switch into a testing mode that doesn't need
// hardware support.
DEFINE_bool(use_tpm, true, "Whether or not to use the TPM Tao");
DEFINE_string(linux_hash, "FAKE_PCRS",
              "The hash of the Linux OS for the DirectTaoChildChannel");
DEFINE_string(fake_keys, "./fake_tpm", "Directory containing signing_key and "
                                       "sealing_key to use with the fake tao");

int main(int argc, char **argv) {
  google::ParseCommandLineFlags(&argc, &argv, true);
  google::InstallFailureSignalHandler();
  google::InitGoogleLogging(argv[0]);
  tao::InitializeOpenSSL();
  tao::LetChildProcsDie();

  scoped_ptr<TaoDomain> admin(TaoDomain::Load(FLAGS_config_path));
  CHECK(admin.get() != nullptr) << "Could not load configuration";

  scoped_ptr<TaoChildChannel> child_channel;
  if (FLAGS_use_tpm) {
    ifstream aik_blob_file(FLAGS_aik_blob.c_str(), ifstream::in);
    if (!aik_blob_file) {
      LOG(ERROR) << "Could not open the file " << FLAGS_aik_blob;
      return 1;
    }

    stringstream aik_blob_stream;
    aik_blob_stream << aik_blob_file.rdbuf();

    ifstream aik_attest_file(FLAGS_aik_attestation.c_str(), ifstream::in);
    if (!aik_attest_file) {
      LOG(ERROR) << "Could not open the file " << FLAGS_aik_attestation;
      return 1;
    }

    stringstream aik_attest_stream;
    aik_attest_stream << aik_attest_file.rdbuf();

    // The TPM to use for the parent Tao
    child_channel.reset(new TPMTaoChildChannel(
        aik_blob_stream.str(), aik_attest_stream.str(), list<UINT32>{17, 18}));
  } else {
    // The FakeTao to use for the parent Tao
    scoped_ptr<Tao> tao(new FakeTao(FLAGS_fake_keys, admin->DeepCopy()));
    CHECK(tao->Init()) << "Could not initialize the FakeTao";
    child_channel.reset(
        new DirectTaoChildChannel(tao.release(), FLAGS_linux_hash));
  }

  CHECK(child_channel->Init()) << "Could not init the TPM";

  // The Channels to use for hosted programs and the way to create hosted
  // programs.
  scoped_ptr<PipeTaoChannel> pipe_channel(
      new PipeTaoChannel(FLAGS_program_socket, FLAGS_stop_socket));
  CHECK(pipe_channel->Init()) << "Could not initialize the pipe channel";
  scoped_ptr<ProcessFactory> process_factory(new ProcessFactory());
  CHECK(process_factory->Init()) << "Could not initialize the process factory";

  scoped_ptr<LinuxTao> tao(new LinuxTao(
      FLAGS_keys_path, child_channel.release(), pipe_channel.release(),
      process_factory.release(), admin.release()));
  CHECK(tao->Init()) << "Could not initialize the LinuxTao";

  LOG(INFO) << "Linux Tao Service started and waiting for requests";

  // Listen for program creation requests and for messages from hosted programs
  // that have been created.
  CHECK(tao->Listen());

  return 0;
}