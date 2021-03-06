/*
 * Copyright (c) 2013 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef _VMX_TIMER_H_
#define _VMX_TIMER_H_

// Function : vmx_timer_hw_setup
// Purpose  : Checks if VMX timer is supported by hardware and if so,
//          : calculates its rate relative to TSC.
// Arguments: void
// Return   : TRUE id supported
// Note     : Must be call 1st on the given core.
BOOLEAN vmx_timer_hw_setup(void);
BOOLEAN vmx_timer_create(GUEST_CPU_HANDLE gcpu);
BOOLEAN vmx_timer_start(GUEST_CPU_HANDLE gcpu);
BOOLEAN vmx_timer_stop(GUEST_CPU_HANDLE gcpu);
BOOLEAN vmx_timer_set_period(GUEST_CPU_HANDLE gcpu, UINT64 period);
BOOLEAN vmx_timer_launch(GUEST_CPU_HANDLE gcpu, UINT64 time_to_expiration, BOOLEAN periodic);
BOOLEAN vmx_timer_set_mode(GUEST_CPU_HANDLE gcpu, BOOLEAN save_value_mode);

#endif // _VMX_TIMER_H_

