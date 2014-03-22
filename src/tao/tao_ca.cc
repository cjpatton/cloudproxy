//  File: tao_ca.cc
//  Author: Kevin Walsh <kwalsh@holycross.edu>
//
//  Description: Implementation of a Tao Certificate Authority client.
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
#include "tao/tao_ca.h"

#include <netinet/in.h>
#include <sys/socket.h>
#include <sys/types.h>

#include <glog/logging.h>
#include <google/protobuf/text_format.h>
#include <keyczar/keyczar.h>

#include "tao/attestation.pb.h"
#include "tao/keys.h"
#include "tao/keys.pb.h"
#include "tao/tao_ca.pb.h"
#include "tao/tao_domain.h"

using google::protobuf::TextFormat;

using tao::ReceiveMessage;
using tao::ScopedFd;
using tao::SendMessage;
using tao::TaoCARequest;
using tao::TaoCAResponse;
using tao::TaoDomain;
using tao::X509Details;

namespace tao {

TaoCA::TaoCA(TaoDomain *admin) : admin_(admin) {}

TaoCA::~TaoCA() {}

bool TaoCA::SendRequest(const TaoCARequest &req, TaoCAResponse *resp) {
  string host = admin_->GetTaoCAHost();
  string port = admin_->GetTaoCAPort();
  ScopedFd sock(new int(-1));
  if (!ConnectToTCPServer(host, port, sock.get())) {
    LOG(ERROR) << "Could not connect to TaoCAServer " << host << ":" << port;
    return false;
  }
  if (!tao::SendMessage(*sock, req)) {
    LOG(ERROR) << "Could not send to TaoCAServer " << host << ":" << port;
    return false;
  }
  if (!tao::ReceiveMessage(*sock, resp)) {
    LOG(ERROR) << "Could not receive from TaoCAServer " << host << ":" << port;
    return false;
  }
  if (resp->type() != TAO_CA_RESPONSE_SUCCESS) {
    LOG(ERROR) << "TCCA returned error: " << resp->reason();
    return false;
  }
  return true;
}

bool TaoCA::GetAttestation(const string &intermediate_attestation,
                           string *root_attestation) {
  return GetX509Chain(intermediate_attestation, "" /* x509 details */,
                      root_attestation, nullptr /* pem_cert */);
}

bool TaoCA::GetX509Chain(const string &intermediate_attestation,
                         const string &details_text, string *root_attestation,
                         string *pem_cert) {
  // Check the existing attestation
  string key_data;
  if (!admin_->VerifyAttestation(intermediate_attestation, &key_data)) {
    LOG(ERROR) << "The original attestation did not pass verification";
    return false;
  }
  TaoCARequest req;
  req.set_type(TAO_CA_REQUEST_ATTESTATION);
  Attestation *attest = req.mutable_attestation();
  if (!attest->ParseFromString(intermediate_attestation)) {
    LOG(ERROR) << "Could not deserialize attestation";
    return false;
  }
  if (pem_cert) {
    X509Details *details = req.mutable_x509details();
    if (!TextFormat::ParseFromString(details_text, details)) {
      LOG(ERROR) << "Could not parse x509 details";
      return false;
    }
  }
  TaoCAResponse resp;
  if (!SendRequest(req, &resp)) {
    LOG(ERROR) << "Could not obtain new attestation";
    return false;
  }
  // Sanity check the response.
  if (!resp.has_attestation()) {
    LOG(ERROR) << "Missing attestation in TaoCA response";
    return false;
  }
  const Attestation &root_attest = resp.attestation();
  // Check the attestation to make sure it passes verification.
  if (root_attest.type() != ROOT) {
    LOG(ERROR) << "Expected a Root attestation from TaoCA";
    return false;
  }
  if (!root_attest.SerializeToString(root_attestation)) {
    LOG(ERROR) << "Could not serialize the new attestation";
    return false;
  }
  string resp_key_data;
  if (!admin_->VerifyAttestation(*root_attestation, &resp_key_data)) {
    LOG(ERROR) << "The attestation did not pass verification";
    return false;
  }
  if (resp_key_data != key_data) {
    LOG(ERROR) << "The key in the new attestation doesn't match original key";
    return false;
  }
  if (pem_cert) {
    if (!resp.has_x509chain()) {
      LOG(ERROR) << "Missing x509 chain in TaoCA response";
      return false;
    }
    // TODO(kwalsh): verify the x509 chain
    pem_cert->assign(resp.x509chain());
  }
  return true;
}

bool TaoCA::Shutdown() {
  LOG(INFO) << "Requesting TaoCA shutdown...";
  TaoCARequest req;
  req.set_type(TAO_CA_REQUEST_SHUTDOWN);
  TaoCAResponse resp;
  return SendRequest(req, &resp);
}

}  // namespace tao