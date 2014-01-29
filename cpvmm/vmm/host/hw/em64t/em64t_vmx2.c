/*
 * Copyright (c) 2013 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *           
 *     http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License. 
 */
#include "vmm_defs.h"
#define VMM_NATIVE_VMCALL_SIGNATURE 0x024694D40

extern void gcpu_save_registers();
extern void gcpu_restore_registers();

struct VMEXIT_TIME {
    UINT64  last_vmexit;
    UINT64  last_vmentry;
    UINT64  last_reason;
    UINT64  last_cpu_id;
    UINT64  this_vmexit;
    UINT64  this_vmentry;
    UINT64  this_reason;
    UINT64  this_cpu_id;
};


void zero_exit_time(struct VMEXIT_TIME *p)
{
    p->last_vmexit= 0ULL;
    p->last_vmentry= 0ULL;
    p->last_reason= 0ULL;
    p->last_cpu_id= 0ULL;
    p->this_vmexit= 0ULL;
    p->this_vmentry= 0ULL;
    p->this_reason= 0ULL;
    p->this_cpu_id= 0ULL;
}

void vmexit_func()
// Function:    Called upon VMEXIT. Saves GP registers, allocates stack
//              for C-function and calls it.
// Arguments:   none
{
    gcpu_save_registers();
    asm volatile(
        "\txor      %%rcx, %%rcx\n" \
        "\tcmpq     $4,%%rcx\n" \
        "\tja       1f\n" \
        "\tmovq     $4, %%rcx\n" \
        "1:\n" \
        "\tshlq     $3, %%rcx\n" \
        "\tsubq     %%rcx, %%rsp\n" \
        "\tcall    vmexit_common_handler\n" \
        "2:\n" \
        "\tjmp     2b\n"
    : 
    : 
    :"%rcx");
}

void vmentry_func(UINT32 firsttime)
// Function:    Called upon VMENTRY.
// Arguments:   firsttime = 1 if called first time
{
	//RNB: Assumption: rflags_arg is still addressable (by %rsp).
	//RNB: rdx is clobbered after being restore.  Should we restore it even?
	//RNB: The asm file sets rcx to 1 after the call to vmentry_failure_function?
		ADDRESS rflags_arg = 0;
    if(firsttime==0ULL) { //do_launch
			gcpu_restore_registers();
			asm volatile (
        "\tvmlaunch\n"
        "\tpushfq\n" 
        "\tpop %%rdx\n" 
        "\tmovq %%rdx, %[rflags_arg]\n" 
        // JLM FIX:  save arguments for vm_failure function
		    :[rflags_arg] "=m" (rflags_arg) 
				: :
			);
		} else {  //do_resume
			gcpu_restore_registers();
  	  asm volatile(
        // Resume execution of Guest Virtual Machine
        "\tvmresume\n" 
				"\tpushfq \n"
        "\tpop %%rdx\n" 
        "\tmovq %%rdx, %[rflags_arg]\n" 
		    :[rflags_arg] "=m" (rflags_arg) 
				::
			);
		}		
    vmentry_failure_function(rflags_arg);
    vmentry_func(0ULL);
}

UINT64 hw_vmcall(UINT64 vmcall_id, UINT64 arg1, UINT64 arg2, UINT64 arg3)
// Function:    VMCALL
// uVMM expects the following:
//     vmcall_id in RCX
//     arg1      in RDX
//     arg2      in RDI
//     arg3      in RSI
// return value in RAX
{
    UINT64  result;

		//RNB: Original asm file mov r8 -> rsi and r9 ->rdi, not sure why?
    asm volatile(
        "\tmovq %[vmcall_id], %%rcx\n" \
        "\tmovq %[arg1], %%rdx\n" \
        "\tmovq %[arg2], %%rdi\n" \
        "\tmovq %[arg3], %%rsi\n" \
        "\tmovq 0x024694D40, %%rax\n" \
        "\tvmcall\n" \
        "\tjmp  2f\n" \
        "1:\n" \
        "\tjmp  1b\n" \
        "2:\n" \
    : 
    : [vmcall_id] "g" (vmcall_id), [arg1] "g" (arg1), [arg2] "g" (arg2), [arg3] "g" (arg3)
    :"%rax", "%rdi", "%rsi", "%r8", "%rcx", "%rdx");
    return result;
}