//  File: tao_auth.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: An interface for hosted-program authorization mechanisms
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

#ifndef TAO_TAO_AUTH_H_
#define TAO_TAO_AUTH_H_

#include <string>

using std::string;

namespace tao {
/// An interface that is used to answer authorization questions for hosts and
/// hosted programs under the Tao.
class TaoAuth {
 public:
  virtual ~TaoAuth() {}
  virtual bool Init() = 0;

  /// Check that a given program hash is authorized.
  /// @param program_hash The hash to check.
  virtual bool IsAuthorized(const string &program_hash) const = 0;
  
  /// Check that a given name/hash pair is authorized.
  /// @param program_name The name to check.
  /// @param program_hash The hash to check.
  virtual bool IsAuthorized(const string &program_name,
                            const string &program_hash) const = 0;

  /// Check an attestation produced by the Tao method Attest
  /// for a given data string.
  /// @param attestation An Attestation produced by Tao::Attest()
  /// @param[out] data The extracted data from the Statement in the Attestation
  /// @return true if the attestation passes verification
  virtual bool VerifyAttestation(const string &attestation,
                                 string *data) const = 0;
};
}  // namespace tao

#endif  // TAO_TAO_AUTH_H_
