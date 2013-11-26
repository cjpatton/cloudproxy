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

#include <glog/logging.h>
#include <gflags/gflags.h>
#include "tao/linux_tao.h"
#include "tao/pipe_tao_channel.h"
#include "tao/process_factory.h"
#include "tao/tpm_tao_child_channel.h"

#include <openssl/ssl.h>
#include <openssl/crypto.h>
#include <openssl/err.h>

#include <fstream>
#include <memory>
#include <mutex>
#include <sstream>
#include <string>
#include <vector>

DEFINE_string(secret_path, "linux_tao_service_secret",
              "The path to the TPM-sealed key for this binary");
DEFINE_string(aik_blob, "HW/aikblob", "The AIK blob from the TPM");
DEFINE_string(key_path, "linux_tao_service_files/key",
              "An encrypted keyczar directory for an encryption key");
DEFINE_string(pk_key_path, "linux_tao_service_files/public_key",
              "An encrypted keyczar directory for a signing key");
DEFINE_string(whitelist, "signed_whitelist", "A signed whitelist file");
DEFINE_string(policy_pk_path, "./policy_public_key",
              "The path to the public policy key");
DEFINE_string(program, "server", "The program to start under the Tao");

using std::ifstream;
using std::shared_ptr;
using std::string;
using std::stringstream;
using std::vector;

using tao::LinuxTao;
using tao::PipeTaoChannel;
using tao::ProcessFactory;
using tao::TPMTaoChildChannel;

vector<shared_ptr<mutex> > locks;

void locking_function(int mode, int n, const char *file, int line) {
  if (mode & CRYPTO_LOCK) {
    locks[n]->lock();
  } else {
    locks[n]->unlock();
  }
}

int main(int argc, char **argv) {
  google::ParseCommandLineFlags(&argc, &argv, true);
  google::InstallFailureSignalHandler();

  FLAGS_alsologtostderr = true;
  google::InitGoogleLogging(argv[0]);

  // initialize OpenSSL
  SSL_load_error_strings();
  ERR_load_BIO_strings();
  ERR_load_crypto_strings();
  OpenSSL_add_all_algorithms();
  SSL_library_init();

  // set up locking in OpenSSL
  int lock_count = CRYPTO_num_locks();
  locks.resize(lock_count);
  for (int i = 0; i < lock_count; i++) {
    locks[i].reset(new mutex());
  }
  CRYPTO_set_locking_callback(locking_function);

  ifstream aik_blob_file(FLAGS_aik_blob.c_str(), ifstream::in);
  stringstream aik_blob_stream;
  aik_blob_stream << aik_blob_file.rdbuf();

  // The TPM to use for the parent Tao
  // TODO(tmroeder): add a proper AIK attestation from the public key
  scoped_ptr<TPMTaoChildChannel> tpm(
      new TPMTaoChildChannel(aik_blob_stream.str(), "", list<UINT32>{17, 18}));
  tpm->Init();

  // The Channels to use for hosted programs and the way to create hosted
  // programs.
  scoped_ptr<PipeTaoChannel> pipe_channel(new PipeTaoChannel());
  scoped_ptr<ProcessFactory> process_factory(new ProcessFactory());

  scoped_ptr<LinuxTao> tao(
      new LinuxTao(FLAGS_secret_path, FLAGS_key_path, FLAGS_pk_key_path,
                   FLAGS_whitelist, FLAGS_policy_pk_path, tpm.release(),
                   pipe_channel.release(), process_factory.release()));

  // The remain command-line flags are passed to the program it starts.
  list<string> args;
  for (int i = 1; i < argc; i++) {
    string arg(argv[i]);
    args.push_back(arg);
  }

  CHECK(tao->StartHostedProgram(FLAGS_program, args))
      << "Could not start " << FLAGS_program << " as a hosted program";

  // TODO(tmroeder): remove this while loop once we fix the thread handling in
  // the LinuxTao so it doesn't depend on this thread to hold its memory
  while (true)
    ;

  return 0;
}
