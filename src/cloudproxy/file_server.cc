//  File: file_server.cc
//  Author: Tom Roeder <tmroeder@google.com>
//
// Description: Implementation of the FileServer class that manages
// files for FileClient
//
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

#include "cloudproxy/file_server.h"

#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#include <mutex>
#include <string>

#include <glog/logging.h>
#include <keyczar/base/base64w.h>
#include <keyczar/base/file_util.h>
#include <keyczar/keyczar.h>

#include "cloudproxy/cloud_auth.h"
#include "cloudproxy/cloud_user_manager.h"
#include "tao/util.h"

using std::lock_guard;
using std::mutex;
using std::string;

using keyczar::base::Base64WEncode;
using keyczar::base::PathExists;
using keyczar::base::ScopedSafeString;
using tao::GenerateSigningKey;
using tao::LoadSigningKey;
using tao::ScopedFile;

namespace cloudproxy {

FileServer::FileServer(const string &file_path, const string &meta_path,
                       const string &server_config_path, const string &secret,
                       const string &acl_location, const string &host,
                       const string &port, tao::TaoDomain *admin)
    : CloudServer(server_config_path, secret, acl_location, host, port, admin),
      enc_key_(new string()),
      hmac_key_(new string()),
      file_path_(file_path),
      meta_path_(meta_path) {

  ScopedSafeString encoded_secret(new string());
  CHECK(Base64WEncode(secret, encoded_secret.get()))
      << "Could not encode the secret as a Base64W string";

  FilePath fp(server_config_path);
  // TODO(kwalsh) It seems FileServer uses a separate signing key than
  // CloudProxy. For now, put fileserver's keys in a subdirectory to avoid name
  // clashes with CloudServer. Maybe use the same key?
  fp = fp.Append("fileserver");
  FilePath priv_key_path = fp.Append(tao::keys::SignPrivateKeySuffix);
  FilePath pub_key_path = fp.Append(tao::keys::SignPublicKeySuffix);
  if (!PathExists(priv_key_path)) {
    CHECK(GenerateSigningKey(priv_key_path.value(), pub_key_path.value(),
                             "file server key", *encoded_secret, &main_key_))
        << "Could not create new signing key for the file server";
  } else {
    CHECK(LoadSigningKey(priv_key_path.value(), *encoded_secret, &main_key_))
        << "Could not load signing key for the file server";
  }

  // check to see if these paths actually exist
  struct stat st;
  CHECK_EQ(stat(file_path_.c_str(), &st), 0) << "Could not stat the directory "
                                             << file_path_;
  CHECK(S_ISDIR(st.st_mode)) << "The path " << file_path_
                             << " is not a directory";

  CHECK_EQ(stat(meta_path_.c_str(), &st), 0) << "Could not stat the directory "
                                             << meta_path_;
  CHECK(S_ISDIR(st.st_mode)) << "The path " << meta_path_
                             << " is not a directory";

  // generate keys
  CHECK(DeriveKeys(main_key_.get(), &enc_key_, &hmac_key_))
      << "Could not derive enc and hmac keys for authenticated encryption";
}

bool FileServer::HandleCreate(const Action &action, SSL *ssl, string *reason,
                              bool *reply, CloudServerThreadData &cstd) {
  // check to see if the file exists
  if (!action.has_object()) {
    LOG(ERROR) << "The CREATE request did not specify a file";
    reason->assign("No file given for CREATE");
    return false;
  }

  // TODO(tmroeder): make this locking more fine-grained so that locks only
  // apply to individual files. Need a locking data structure for this.
  string path = file_path_ + string("/") + action.object();
  string meta_path = meta_path_ + string("/") + action.object();
  {
    lock_guard<mutex> l(data_m_);
    struct stat st;
    if (stat(path.c_str(), &st) == 0) {
      LOG(ERROR) << "File " << path << " already exists";
      reason->assign("Already exists");
      return false;
    }

    if (stat(meta_path.c_str(), &st) == 0) {
      LOG(ERROR) << "File " << meta_path << " already exists";
      reason->assign("Already exists");
      return false;
    }

    ScopedFile f(fopen(path.c_str(), "w"));
    if (nullptr == f.get()) {
      LOG(ERROR) << "Could not create the file " << path;
      reason->assign("Could not create the file");
      return false;
    }

    ScopedFile mf(fopen(meta_path.c_str(), "w"));
    if (nullptr == mf.get()) {
      LOG(ERROR) << "Could not create the file " << meta_path;
      reason->assign("Could not create the file");
      return false;
    }
  }

  VLOG(2) << "Created the file " << path << " and its metadata " << meta_path;
  return true;
}

bool FileServer::HandleDestroy(const Action &action, SSL *ssl, string *reason,
                               bool *reply, CloudServerThreadData &cstd) {
  if (!action.has_object()) {
    LOG(ERROR) << "The DESTROY request did not specify a file";
    reason->assign("No file given for DESTROY");
    return false;
  }

  string path = file_path_ + string("/") + action.object();
  string meta_path = meta_path_ + string("/") + action.object();
  {
    lock_guard<mutex> l(data_m_);
    struct stat st;
    if (stat(path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    if (stat(meta_path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << meta_path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    // ideally, this should be transactional, since the current instantiation of
    // this code can get files into a state that can't be written and can't be
    // destroyed
    if (unlink(path.c_str()) != 0) {
      LOG(ERROR) << "Could not unlink the file " << path;
      reason->assign("Could not delete the file");
      return false;
    }

    if (unlink(meta_path.c_str()) != 0) {
      LOG(ERROR) << "Could not unlink the file " << meta_path;
      reason->assign("Could not delete the file");
      return false;
    }
  }

  return true;
}

bool FileServer::HandleWrite(const Action &action, SSL *ssl, string *reason,
                             bool *reply, CloudServerThreadData &cstd) {
  string path = file_path_ + string("/") + action.object();
  string meta_path = meta_path_ + string("/") + action.object();
  {
    lock_guard<mutex> l(data_m_);
    struct stat st;
    if (stat(path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    if (stat(meta_path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << meta_path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    // send a reply before receiving the stream data
    // the reply tells the FileClient that it can send the data
    string error;
    if (!SendReply(ssl, true, error)) {
      LOG(ERROR) << "Could not send a message to the client to ask it to write";

      // don't try to send another message, since we just failed to send this
      // one
      *reply = false;
      return false;
    }

    if (!ReceiveAndEncryptStreamData(ssl, path, meta_path, action.object(),
                                     enc_key_, hmac_key_, main_key_.get())) {
      LOG(ERROR) << "Could not receive data from the client and write it"
                    " encrypted to disk";
      reason->assign("Receiving failed");
      return false;
    }
  }

  return true;
}

bool FileServer::HandleRead(const Action &action, SSL *ssl, string *reason,
                            bool *reply, CloudServerThreadData &cstd) {
  string path = file_path_ + string("/") + action.object();
  string meta_path = meta_path_ + string("/") + action.object();
  {
    lock_guard<mutex> l(data_m_);
    struct stat st;
    if (stat(path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    if (stat(meta_path.c_str(), &st) != 0) {
      LOG(ERROR) << "File " << meta_path << " does not exist";
      reason->assign("Does not exist");
      return false;
    }

    // send a reply before sending the stream data
    // the reply tells the FileClient that it should expect the data
    string error;
    if (!SendReply(ssl, true, error)) {
      LOG(ERROR) << "Could not send a message to the client to tell it to read";

      // don't try to send another message, since we just failed to send this
      // one
      *reply = false;
      return false;
    }

    if (!DecryptAndSendStreamData(path, meta_path, action.object(), ssl,
                                  enc_key_, hmac_key_, main_key_.get())) {
      LOG(ERROR) << "Could not stream data from the file to the client";
      reason->assign("Could not stream data to the client");
      return false;
    }
  }

  return true;
}

}  // namespace cloudproxy
