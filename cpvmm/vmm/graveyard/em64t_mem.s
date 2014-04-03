#
# Copyright (c) 2013 Intel Corporation
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
#

/*
externdef vmm_memset:NEAR
externdef vmm_memcpy:NEAR
externdef vmm_strlen:NEAR


PUBLIC vmm_lock_xchg_qword
PUBLIC vmm_lock_xchg_byte
 */

.intel_syntax
.text

#
#
# Calling conventions
#
# Floating : First 4 parameters � XMM0 through XMM3. Others passed on stack.
#
# Integer  : First 4 parameters � RCX, RDX, R8, R9. Others passed on stack.
#
# Aggregates (8, 16, 32, or 64 bits) and __m64:
#       First 4 parameters � RCX, RDX, R8, R9. Others passed on stack.
#
# Aggregates (other):
#       By pointer. First 4 parameters passed as pointers in RCX, RDX, R8, and R9
#
# __m128   : By pointer. First 4 parameters passed as pointers in RCX, RDX, R8, and R9
#
#
# Return values that can fit into 64-bits are returned through RAX 
# (including __m64 types), except for __m128, __m128i, __m128d, floats, 
# and doubles, which are returned in XMM0.  If the return value does not 
# fit within 64 bits, then the caller assumes the responsibility
# of allocating and passing a pointer for the return value as the 
# first argument. Subsequent arguments are then shifted one argument 
# to the right. That same pointer must be returned by the callee in RAX. 
# User defined types to be returned must be 1, 2, 4, 8, 16, 32, or 64
# bits in length.
#
# Register usage
#
# Caller-saved and scratch:
#      RAX, RCX, RDX, R8, R9, R10, R11
#
# Callee-saved
#      RBX, RBP, RDI, RSI, R12, R13, R14, and R15
#

#------------------------------------------------------------------------------
#  force compiler intrinsics to use our code
#------------------------------------------------------------------------------
/*
memset PROC
    jmp vmm_memset
memset ENDP

memcpy PROC
    jmp vmm_memcpy
memcpy ENDP

strlen PROC
    jmp vmm_strlen
strlen ENDP
 */

#
#
# Lock exchange qword
# VOID
# vmm_lock_xchg_qword (
#                      UINT64 *dst, # rcx
#                      UINT64 *src  # rdx
#                     )

.globl vmm_lock_xchg_qword
vmm_lock_xchg_qword:
    push r8
    mov r8, [rdx] # copy src to r8
    lock xchg [rcx], r8
    pop r8
    ret

#
#
# Lock exchange byte
# VOID
# vmm_lock_xchg_byte (
#                     UINT8 *dst, # rcx
#                     UINT8 *src  # rdx
#                    )
#
.globl  vmm_lock_xchg_byte
vmm_lock_xchg_byte:
    push rbx
    mov bl, byte ptr [rdx] # copy src to bl
    lock xchg byte ptr [rcx], bl
    pop rbx
    ret
