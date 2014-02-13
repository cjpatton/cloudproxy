//  File: linux_tao.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: LinuxTao implements the Tao for the Linux
//  operating system
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
#ifndef TAO_LINUX_TAO_H_
#define TAO_LINUX_TAO_H_

#include <list>
#include <mutex>
#include <set>
#include <string>

#include <keyczar/base/basictypes.h>  // DISALLOW_COPY_AND_ASSIGN
#include <keyczar/base/scoped_ptr.h>
#include <keyczar/base/stl_util-inl.h>  // ScopedSafeString

#include "tao/attestation.pb.h"
#include "tao/tao.h"
#include "tao/tao_domain.h"

using std::list;
using std::mutex;
using std::set;
using std::string;

namespace tao {

class HostedProgramFactory;
class TaoChannel;
class TaoChildChannel;
class TaoDomain;

/// An implementation of the Tao for Linux. This implementation can take
/// different HostedProgramFactory implementations, different TaoChannel
/// implementations for communicating with its hosted programs, and different
/// TaoChildChannel implementations for communicating with its parent Tao. The
/// only assumptions LinuxTao makes are basic: it has the normal filesystem API
/// to which it can write files, and it is an intermediate Tao rather than the
/// root Tao.
class LinuxTao : public Tao {
 public:
  /// The LinuxTao is constructed with paths to keys and implementations of
  /// factories and channels. Ownership is taken for all relevant parameters.
  /// @param keys_path The directory storing keys and secrets.
  /// @param host_channel A channel to communicate with the parent Tao.
  /// @param child_channel A channel to communicate with hosted programs it
  /// starts.
  /// @param program_factory A factory that creates hosted programs in the OS.
  /// @param admin The configuration for this administrative domain.
  LinuxTao(const string &keys_path, TaoChildChannel *host_channel,
           TaoChannel *child_channel, HostedProgramFactory *program_factory,
           TaoDomain *admin)
      : admin_(admin),
        keys_path_(keys_path),
        host_channel_(host_channel),
        child_channel_(child_channel),
        program_factory_(program_factory) {}
  virtual ~LinuxTao() {}

  /// Start listening for Tao messages on channels.
  bool Listen();

  /// LinuxTao follows the normal semantics of the Tao for these methods.
  /// @{
  virtual bool Init();
  virtual bool Destroy() { return true; }
  virtual bool StartHostedProgram(const string &program,
                                  const list<string> &args, string *identifier);
  virtual bool RemoveHostedProgram(const string &child_hash);
  virtual bool GetRandomBytes(size_t size, string *bytes) const;
  virtual bool Seal(const string &child_hash, const string &data,
                    string *sealed) const;
  virtual bool Unseal(const string &child_hash, const string &sealed,
                      string *data) const;
  virtual bool Attest(const string &child_hash, const string &data,
                      string *attestation) const;
  /// @}

 private:
  /// Configuration for this administrative domain.
  scoped_ptr<TaoDomain> admin_;

  // Size of secret for protecting our crypting and signing keys.
  static const int SecretSize = 128;

  /// Directory to store our keys.
  string keys_path_;

  /// A crypting key for sealing, protected by a sealed secret.
  scoped_ptr<keyczar::Crypter> crypter_;

  /// A signing key for signing, protected by crypter_.
  scoped_ptr<keyczar::Signer> signer_;

  /// An attestation for our signing public key.
  string attestation_;

  /// The channel to use for host communication.
  scoped_ptr<TaoChildChannel> host_channel_;

  /// A channel that handles all child connections.
  scoped_ptr<TaoChannel> child_channel_;

  /// A factory that can be used to start hosted programs.
  scoped_ptr<HostedProgramFactory> program_factory_;

  /// The set of hosted programs that the LinuxTao has started.
  set<string> running_children_;

  /// The address:port for the TCCA. If empty, then this code won't call the CA.
  /// @{
  string ca_host_;
  string ca_port_;
  /// {@

  /// A mutex for accessing the auth manager.
  mutable mutex auth_m_;

  /// A mutex for accessing and modifying running_children_.
  mutable mutex data_m_;

  /// Unseal a secret using the TaoChildChannel.
  /// @param secret[out] The unsealed secret.
  bool GetSecret(keyczar::base::ScopedSafeString *secret);

  /// Create and seal a secret using the TaoChildChannel.
  /// @param secret[out] The new secret.
  bool MakeSecret(keyczar::base::ScopedSafeString *secret);

  /// Get an attestation from the host Tao on our key. Note that this
  /// will get an attestation on serialized_key for this Tao host; it
  /// is for use by this Tao and its subclasses.
  /// @param attest[out] The attestation.
  bool GetTaoAttestation(Attestation *attest);

  /// Send the attestation to the TCCA on a port/address supplied at
  /// initialization and uses the returned attestation as its attestation.
  /// @param attest[in,out] Existing attestation to be replaced by a new
  bool GetTaoCAAttestation(Attestation *attest);

  DISALLOW_COPY_AND_ASSIGN(LinuxTao);
};
}  // namespace tao

#endif  // TAO_LINUX_TAO_H_