//  File: file_client.h
//  Author: Tom Roeder <tmroeder@google.com>
//
// Description: The FileClient class interacts with FileServer
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

#ifndef CLOUDPROXY_FILE_CLIENT_H_
#define CLOUDPROXY_FILE_CLIENT_H_

#include <string>

#include "cloudproxy/cloud_client.h"

using std::string;

namespace cloudproxy {
/// An implementation of CloudClient that sends files to be stored remotely by
/// FileServer. Most of the work of this method is done in the CloudClient
/// implementation directly.
class FileClient : public CloudClient {
 public:
  /// Create a new client for communicating with FileServer. All the parameters
  /// other than file_path have the same semantics as in CloudClient.
  /// @param file_path The path to the files managed by this client.
  /// @param client_config_path A directory to use for keys and TLS files.
  /// @param secret A string to use for a encrypting private keys.
  /// @param admin The configuration for this administrative domain.
  FileClient(const string &file_path, const string &client_config_path,
             const string &secret, tao::TaoDomain *admin);
  virtual ~FileClient() {}

  /// Create a file on a FileServer.
  /// @param ssl The server connection to use.
  /// @param owner The name of a user that has permission to create this file.
  /// @param object_name The filename for the file to create.
  virtual bool Create(SSL *ssl, const string &owner, const string &object_name);

  /// Delete a file on a FileServer.
  /// @param ssl The server connection to use.
  /// @param owner The name of a user that has permission to delete this file.
  /// @param object_name The filename for the file to delete.
  virtual bool Destroy(SSL *ssl, const string &owner,
                       const string &object_name);

  /// Read a file stored on a FileServer.
  /// @param ssl The server connection to use.
  /// @param requestor The name of the user requesting the file.
  /// @param object_name The path of the file to read.
  /// @param output_name A file that will receive the bytes from the server.
  virtual bool Read(SSL *ssl, const string &requestor,
                    const string &object_name, const string &output_name);

  /// Write to a file stored on a FileServer.
  /// @param ssl The server connection to use.
  /// @param requestor The name of the user writing to the file.
  /// @param input_name The name of the local file to read.
  /// @param object_name The name of the remote file that receives the data.
  virtual bool Write(SSL *ssl, const string &requestor,
                     const string &input_name, const string &object_name);

 private:
  /// The base path for files that are read from and written to the server.
  string file_path_;

  DISALLOW_COPY_AND_ASSIGN(FileClient);
};
}  // namespace cloudproxy

#endif  // CLOUDPROXY_FILE_CLIENT_H_
