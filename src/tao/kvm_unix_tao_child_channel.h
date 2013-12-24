//  File: kvm_unix_tao_child_channel.h
//  Author: Tom Roeder <tmroeder@google.com>
//
//  Description: The hosted program interface for the KvmUnixTaoChannel
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

#ifndef TAO_KVM_UNIX_TAO_CHILD_CHANNEL_H_
#define TAO_KVM_UNIX_TAO_CHILD_CHANNEL_H_

#include "tao/tao_child_channel.h"

namespace tao {
/// The channel a guest VM uses to communicate with the hypervisor Tao. It does
/// this by running an service (like apps/linux_kvm_guest_tao_service.cc) that 
/// takes in the params from the hypervisor and uses them to find the right
/// device to connect to for hypervisor Tao communication. In the current
/// implementation, this is done by passing a Baes64W-encoded string as the last
/// boot parameter. This can be found in /proc/cmdline.
class KvmUnixTaoChildChannel : public TaoChildChannel {
 public:
  /// This constructor stores the parameters but doesn't parse them or try to
  /// connect to the hypervisor Tao.
  KvmUnixTaoChildChannel(const string &params);
  virtual ~KvmUnixTaoChildChannel() {}

  /// Parse the params from the constructor and connect to the file they name.
  virtual bool Init();

  static string ChannelType() { return "KvmUnixTaoChannel"; }

 protected:
  // subclasses implement these methods for the underlying transport.
  virtual bool ReceiveMessage(google::protobuf::Message *m) const;
  virtual bool SendMessage(const google::protobuf::Message &m) const;

 private:
  // The opened file descriptor for the file specified in params_.
  int fd_;

  // A Base64W-serialized TaoChildChannelParams.
  string params_;
};
}  // namespace tao

#endif  // TAO_KVM_UNIX_TAO_CHILD_CHANNEL_H_
