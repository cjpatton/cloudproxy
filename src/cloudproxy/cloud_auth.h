//  File: cloud_auth.h
//  Author: Tom Roeder <tmroeder@google.com>
//
// Description: The CloudAuth class manages authorization of users of
// CloudClient
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

#ifndef CLOUDPROXY_CLOUD_AUTH_H_
#define CLOUDPROXY_CLOUD_AUTH_H_

#include <map>
#include <set>
#include <string>

#include <keyczar/base/basictypes.h>  // DISALLOW_COPY_AND_ASSIGN

#include "cloudproxy/cloudproxy.pb.h"

using std::map;
using std::set;
using std::string;

namespace keyczar {
class Keyczar;
}  // namespace keyczar

namespace cloudproxy {
class CloudAuth {
 public:
  // Instantiates the Auth class with a serialized representation of a
  // cloudproxy::ACL object.
  CloudAuth(const string &acl_path, keyczar::Keyczar *key);

  virtual ~CloudAuth() {}

  // Checks to see if this operation is permitted by the ACL
  virtual bool Permitted(const string &subject, Op op, const string &object);

  // Removes a given entry from the ACL if it exists
  virtual bool Delete(const string &subject, Op op, const string &object);

  // Adds a given entry to the ACL
  virtual bool Insert(const string &subect, Op op, const string &object);

  // serializes the ACL into a given string
  virtual bool Serialize(string *data);

 protected:
  bool findPermissions(const string &subject, const string &object,
                       set<Op> **perms);

 private:
  // a map from subject->(object, permission set)
  map<string, map<string, set<Op> > > permissions_;

  // a list of users with admin privilege (able to perform any action)
  set<string> admins_;

  DISALLOW_COPY_AND_ASSIGN(CloudAuth);
};
}

#endif  // CLOUDPROXY_CLOUD_AUTH_H_
