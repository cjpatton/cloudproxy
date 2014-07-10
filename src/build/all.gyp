# Copyright (c) 2013, Google Inc. All rights reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{
  'target_defaults': {
    'cflags': [
      '-Wall',
      '-Werror',
      '-std=c++0x',
    ],
    'configurations': {
      'Release': {
        'cflags': [
          '-O2',
        ],
      },
      'Debug': {
        'cflags': [
          '-g',
        ],
      },
    },
  },
  'targets': [
    {
      'target_name': 'All',
      'type': 'none',
      'variables': {
        'src': 'cloudproxy',
      },
      'dependencies': [
        '../apps/apps.gyp:*',
       '../cloudproxy/cloudproxy.gyp:*',
        '../tao/tao.gyp:*',
        '../third_party/keyczar/keyczar.gyp:keyczart',
        '../third_party/libb64/libb64.gyp:*',
      ],
    },
  ]
}
