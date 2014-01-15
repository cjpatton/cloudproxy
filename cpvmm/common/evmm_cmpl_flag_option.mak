#############################################################################
# Copyright (c) 2013 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#############################################################################

#############################################################################
# INTEL CONFIDENTIAL
# Copyright 2001-2013 Intel Corporation All Rights Reserved.
#
# The source code contained or described herein and all documents related to
# the source code ("Material") are owned by Intel Corporation or its
# suppliers or licensors.  Title to the Material remains with Intel
# Corporation or its suppliers and licensors.  The Material contains trade
# secrets and proprietary and confidential information of Intel or its
# suppliers and licensors.  The Material is protected by worldwide copyright
# and trade secret laws and treaty provisions.  No part of the Material may
# be used, copied, reproduced, modified, published, uploaded, posted,
# transmitted, distributed, or disclosed in any way without Intel's prior
# express written permission.
#
# No license under any patent, copyright, trade secret or other intellectual
# property right is granted to or conferred upon you by disclosure or
# delivery of the Materials, either expressly, by implication, inducement,
# estoppel or otherwise.  Any license under such intellectual property rights
# must be express and approved by Intel in writing.
#############################################################################
#----------------------------------------------------------------------------
# eVMM options
#----------------------------------------------------------------------------

# (1) Remove ENABLE_RELEASE_VMM_LOG for external build.
# (3) To build eVMM for pre-os load, add ENABLE_INT15_VIRTUALIZATION.
# (4) Commenting out any option will not work.  If an option is disabled,
#     move it to the "Disabled options" section.

EVMM_CMPL_OPT_FLAGS = \
 /DENABLE_RELEASE_VMM_LOG \
 /DVMCALL_NOT_ALLOWED_FROM_RING_1_TO_3 \
 /DFAST_VIEW_SWITCH \
 /DENABLE_INT15_VIRTUALIZATION \
 /DENABLE_VMM_EXTENSION \

# Use this function in Makefiles to find whether any option is enabled.
find_opt = \
    $(if $(findstring $(1),$(EVMM_CMPL_OPT_FLAGS)),1)


#----------------------------------------------------------------------------
# Disabled options
#----------------------------------------------------------------------------
#/DENABLE_PREEMPTION_TIMER
#/DUSE_ACPI
#/DPCI_SCAN

#-----------------------------------------------                 
# Enable the following 2 flags to support S3
#-----------------------------------------------
#/DENABLE_PM_S3
#/DUSE_ACPI