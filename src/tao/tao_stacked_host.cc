//  File: tao_stacked_host.cc
//  Author: Kevin Walsh <kwalsh@holycross.edu>
//
//  Description: Tao host implemented on top of a host Tao.
//
//  Copyright (c) 2014, Kevin Walsh.  All rights reserved.
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
#include "tao/tao_stacked_host.h"

#include <glog/logging.h>

#include "tao/attestation.h"
#include "tao/attestation.pb.h"
#include "tao/keys.h"
#include "tao/tao.h"

namespace tao {

bool TaoStackedHost::Init() {
  // Get our name early and cache it.
  if (!host_tao_->GetTaoName(&tao_host_name_)) {
    LOG(ERROR) << "Could not obtain Tao host name";
    return false;
  }
  // When using a signing key, we require a delegation to accompany it.
  if (keys_ == nullptr || keys_->Signer() == nullptr ||
      keys_->GetHostDelegation() == "") {
    LOG(ERROR) << "Could not load delegation for attestation key";
    return false;
  }

  VLOG(1) << "TaoStackedHost: Initialization finished successfully";
  VLOG(1) << "TaoStackedHost: " << elideString(tao_host_name_);
  return true;
}

bool TaoStackedHost::GetRandomBytes(const string &child_subprin, size_t size,
                                    string *bytes) const {
  return host_tao_->GetRandomBytes(size, bytes);
}

bool TaoStackedHost::GetSharedSecret(const string &tag, size_t size,
                                     string *bytes) const {
  if (keys_ == nullptr || keys_->Deriver() == nullptr) {
    // TODO(kwalsh) - Generate new deriver key using shared secret material from
    // our host, e.g.:
    // int key_size = 256;
    // string key_bytes;
    // if (!host_tao_->GetSharedSecret(key_size, &key_bytes,
    // SharedSecretPolicyDefault)) {
    //   LOG(ERROR) << "Stacked tao host could not get shared master secret";
    //   return false;
    // }
    // Keys keys(Keys::Deriving, ...key_bytes...);
    // if (!keys.DeriveKey(tag, size, bytes)) {
    //   LOG(ERROR) << "Could not derive shared secret";
    //   return false;
    // }
    // return true;
    LOG(ERROR) << "Not yet implemented";
    return false;
  } else {
    if (!keys_->Deriver()->Derive(size, tag, bytes)) {
      LOG(ERROR) << "Could not derive shared secret";
      return false;
    }
    return true;
  }
}

bool TaoStackedHost::Attest(const string &child_subprin, Statement *stmt,
                            string *attestation) const {
  // Make sure issuer is identical to (or a subprincipal of) the hosted
  // program's principal name.
  if (!stmt->has_issuer()) {
    stmt->set_issuer(tao_host_name_ + "::" + child_subprin);
  } else if (!IsSubprincipalOrIdentical(
                 stmt->issuer(), tao_host_name_ + "::" + child_subprin)) {
    LOG(ERROR) << "Invalid issuer in statement";
    return false;
  }
  // Sign it.
  if (keys_ == nullptr || keys_->Signer() == nullptr) {
    return host_tao_->Attest(*stmt, attestation);
  } else {
    return GenerateAttestation(*keys_->Signer(), keys_->GetHostDelegation(),
                               *stmt, attestation);
  }
}

bool TaoStackedHost::Encrypt(const google::protobuf::Message &data,
                             string *encrypted) const {
  string serialized_data;
  if (!data.SerializeToString(&serialized_data)) {
    LOG(ERROR) << "Could not serialize data to be sealed";
    return false;
  }
  if (keys_ == nullptr || keys_->Crypter() == nullptr) {
    // TODO(kwalsh) Should policy here come from elsewhere?
    return host_tao_->Seal(serialized_data, Tao::SealPolicyDefault, encrypted);
  } else {
    return keys_->Crypter()->Encrypt(serialized_data, encrypted);
  }
}

bool TaoStackedHost::Decrypt(const string &encrypted,
                             google::protobuf::Message *data) const {
  string serialized_data;
  if (keys_ == nullptr || keys_->Crypter() == nullptr) {
    string policy;
    if (!host_tao_->Unseal(encrypted, &serialized_data, &policy)) {
      LOG(ERROR) << "Could not unseal sealed data";
      return false;
    }
    if (policy != Tao::SealPolicyDefault) {
      LOG(ERROR) << "Unsealed data with uncertain provenance";
      return false;
    }
  } else {
    if (!keys_->Crypter()->Decrypt(encrypted, &serialized_data)) {
      LOG(ERROR) << "Could not decrypt sealed data";
      return false;
    }
  }
  if (!data->ParseFromString(serialized_data)) {
    LOG(ERROR) << "Could not deserialize sealed data";
    return false;
  }
  return true;
}

}  // namespace tao
