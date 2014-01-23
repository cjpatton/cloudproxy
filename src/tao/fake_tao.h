//  File: fake_tao.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: A fake implementation of the Tao interface that isn't
//  backed by any trusted hardware.
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

#ifndef TAO_FAKE_TAO_H_
#define TAO_FAKE_TAO_H_

#include <keyczar/base/basictypes.h>  // DISALLOW_COPY_AND_ASSIGN
#include <keyczar/base/scoped_ptr.h>

#include "tao/tao.h"

namespace keyczar {

class Keyczar;

}  // namespace keyczar

namespace tao {
/// A fake Tao implementation that performs crypto operations using
/// in-memory keys, including a fake policy key.
class FakeTao : public Tao {
 public:
  /// Create an in-memory policy key.
  FakeTao();

  /// Use an existing (unencrypted, complete) policy key path.
  /// This version pretends to be the root with the policy key.
  FakeTao(const string &policy_key_path);

  /// Use an existing (unencrypted, complete) key path and add an
  /// attestation from a policy key. This version stands in for hardware
  /// like a TPM.
  /// @param key_path A path to a key to use.
  /// @param attestation An attestation to the key (from the policy key)
  FakeTao(const string &key_path, const string &attestation);
  virtual ~FakeTao() {}

  /// Init initializes the keys and sets everything up.
  virtual bool Init();

  /// Destroy does nothing for this class.
  virtual bool Destroy() { return true; }

  /// The FakeTao pretends to start hosted programs but does nothing.
  virtual bool StartHostedProgram(const string &path, const list<string> &args,
				  string *identifier);

  /// The FakeTao doesn't remove hosted programs, but it accepts the call.
  virtual bool RemoveHostedProgram(const string &child_hash) { return true; }

  // The other Tao methods are implemented using the generated keys just like a
  // normal hardware mechanism would.
  virtual bool GetRandomBytes(size_t size, string *bytes) const;
  virtual bool Seal(const string &child_hash, const string &data,
                    string *sealed) const;
  virtual bool Unseal(const string &child_hash, const string &sealed,
                      string *data) const;
  virtual bool Attest(const string &child_hash, const string &data,
                      string *attestation) const;

 private:
  // The path to the key, if the object was constructed this way.
  string key_path_;

  // A potentially empty attestation to the key.
  string attestation_;

  // An in-memory, temporary symmetric key
  scoped_ptr<keyczar::Keyczar> crypter_;

  // A fake key
  scoped_ptr<keyczar::Keyczar> key_;

  DISALLOW_COPY_AND_ASSIGN(FakeTao);
};
}  // namespace tao

#endif  // TAO_FAKE_TAO_H_
