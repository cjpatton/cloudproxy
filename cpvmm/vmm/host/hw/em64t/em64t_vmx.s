#
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
#

.intel_syntax
.text

.set    ARG1_U8 , %cl
.set    ARG1_U16, %cx
.set    ARG1_U32, %ecx
.set    ARG1_U64, %rcx
.set    ARG2_U8 , %dl
.set    ARG2_U16, %dx
.set    ARG2_U32, %edx
.set    ARG2_U64, %rdx
.set    ARG3_U32, %r8l
.set    ARG3_U64, %r8

#
# Register usage
#
# Caller-saved and scratch:
#        RAX, RCX, RDX, R8, R9, R10, R11
#
# Callee-saved
#        RBX, RBP, RDI, RSI, R12, R13, R14, and R15
#
#
/*
extrn gcpu_save_registers       :NEAR 
extrn gcpu_restore_registers    :NEAR
extrn vmexit_common_handler     :NEAR
extrn vmentry_failure_function  :NEAR

ifdef ENABLE_TMSL_PROFILING
extern profiling_vmexit        :NEAR
extern g_tmsl_profiling_vmexit :NEAR
; Assumption - hw_cpu_id() uses RAX only and returns host cpu id in ax
extrn hw_cpu_id:NEAR

#RNB: TODO have to create an equivalent c struct
; Define PROF_VMEXIT_TIME structure
PROF_VMEXIT_TIME struc
    last_vmexit     qword  00h;
    last_vmentry    qword  00h;
    last_reason    qword  00h;
    last_cpu_id    qword  00h;
    this_vmexit     qword  00h;
    this_vmentry    qword  00h;
    this_reason    qword  00h;
    this_cpu_id    qword  00h;
PROF_VMEXIT_TIME ends;
endif
*/
#
#  Function:    Restore space on the stack for calling C-function
#
#  Arguments:   RCX - contains the number of arguments, passed to C-function
#
.macro RESTORE_C_STACK
        cmp     %rcx, $4
        ja      @F                      #; goto parameters are normalized
        mov     %rcx, $4                  #; at least 4 arguments must be allocated
@@:                                     #; parameters are normalized
        shl     %rcx, $3
        add     %rsp, %rcx
.endm

#
#  Function:    Allocates space on the stack for calling C-function
#
#  Arguments:   RCX - contains the number of arguments, passed to C-function
#
.macro ALLOCATE_C_STACK
        cmp     %rcx, $4
        ja      @F                      ;; goto parameters are normalized
        mov     %rcx, $4                  ;; at least 4 arguments must be allocated
@@:                                     ;; parameters are normalized
        shl     %rcx, $3
        sub     %rsp, %rcx
.endm

#ifdef ENABLE_TMSL_PROFILING

.globl  profiling_serialize 
profiling_serialize:
    # serialize
    mov     %rax, %cr0
    mov     %cr0, %rax

    ret


.globl  profiling_save_vmexit_time 
profiling_save_vmexit_time:
    # save registers. rax must be saved, else will hang.
    push %rax
    push %rbx
    push %r8

    # calculate host cpu id and put it into the rax (ax)
    call hw_cpu_id
    mov %r8, %rax

    # put pointer to the array of GUEST_CPU_SAVE_AREA_PREFIX* to RBX
    mov  %rbx, g_tmsl_profiling_vmexit
    shl  %rax, $6 #size of PROF_VMEXIT_TIME = 64 * rax
    add  %rbx, %rax

    # save last vmexit/vmentry time
    mov  %rax, (PROF_VMEXIT_TIME ptr [rbx]).this_vmexit
    mov  (PROF_VMEXIT_TIME ptr [rbx]).last_vmexit,  %rax
    mov  %rax, (PROF_VMEXIT_TIME ptr [rbx]).this_vmentry
    mov  (PROF_VMEXIT_TIME ptr [rbx]).last_vmentry,  %rax
    mov  %rax, (PROF_VMEXIT_TIME ptr [rbx]).this_reason
    mov  (PROF_VMEXIT_TIME ptr [rbx]).last_reason,  %rax
    mov  %rax, (PROF_VMEXIT_TIME ptr [rbx]).this_cpu_id
    mov  (PROF_VMEXIT_TIME ptr [rbx]).last_cpu_id,  %rax

    # serialize
    call profiling_serialize

    # rdtsc
    push    %rdx
    push    %rcx
    rdtsc
    shl     %rdx, $32
    add     %rax, %rdx

    # save this vmexit time
    mov  (PROF_VMEXIT_TIME ptr [rbx]).this_vmexit, %rax
    mov  (PROF_VMEXIT_TIME ptr [rbx]).this_cpu_id, %r8

    pop    %rcx
    pop    %rdx

    # restore registers
    pop %r8
    pop %rbx
    pop %rax

    ret


.globl  profiling_save_vmentry_time 
profiling_save_vmentry_time:
    # save registers
    push %rax
    push %rbx

    # calculate host cpu id and put it into the rax (ax)
    call hw_cpu_id

    # put pointer to the array of GUEST_CPU_SAVE_AREA_PREFIX* to RBX
    mov  %rbx, g_tmsl_profiling_vmexit
    shl  %rax, $6 #size of PROF_VMEXIT_TIME = 64 * rax
    add  %rbx, %rax

    # serialize
    call profiling_serialize

    # rdtsc
    push    %rdx
    push    %rcx
    rdtsc
    shl     %rdx, 32
    add     %rax, %rdx

    # save this vmexit time
    mov  (PROF_VMEXIT_TIME ptr [%rbx]).this_vmentry, %rax

        pop    %rcx
        pop    %rdx


    # restore registers
    pop %rbx
    pop %rax

    ret

#
#  Function:    Called upon VMEXIT. Saves GP registers, allocates stack
#               for C-function and calls it.
#
#  Arguments:   none
#
.globl  vmexit_func
vmexit_func:
#ifdef ENABLE_TMSL_PROFILING
#        call profiling_save_vmexit_time
#endif
        call    gcpu_save_registers
        xor     %rcx, %rcx
        ALLOCATE_C_STACK
        call    vmexit_common_handler
        jmp     $                       ## should never return


#
#  Function:    Called upon VMENTRY.
#
#  Arguments:   RCX = 1 if called first time
#
vmentry_func    PROC
        push    %rcx
        cmp     %rcx, $0
        jnz     do_launch
do_resume:

#ifdef ENABLE_TMSL_PROFILING
#        call profiling_save_vmentry_time
#        call profiling_vmexit            # profiling the cost before vmentry
#endif
        call    gcpu_restore_registers 

        vmresume                        # Resume execution of Guest Virtual Machine
        jmp     handle_error
do_launch:
        call    gcpu_restore_registers
        vmlaunch                        # Launch execution of Guest Virtual Machine

handle_error:
        pushfq                          # use RFLAGS as argument if VMRESUME failed
        pop     %rdx                     # save RFLAGS in RDX
        mov     %rcx, $1                  # RCX contains number of argments for vmentry_failure_function
        ALLOCATE_C_STACK                # for for vmentry_failure_function
        mov     %rcx, %rdx                # 1st argument (passed via RCX) contains RFLAGS
        call    vmentry_failure_function
        mov     %rcx, $1                  # RCX contains number of argments for vmentry_failure_function
        RESTORE_C_STACK
        pop     %rcx                     # restore RCX. stack is expected to be the same as in entry point
        jmp     vmentry_func            # retry


#
#  Function:    VMCALL
#
#  uVMM expects the following:
#      vmcall_id in RCX
#      arg1      in RDX
#      arg2      in RDI
#      arg3      in RSI
#
#  return value in RAX
#
.set    VMM_NATIVE_VMCALL_SIGNATURE, 0x024694D40
.globl  hw_vmcall
hw_vmcall:
        push    %rdi
        push    %rsi
        mov     %rdi, %r8
        mov     %rsi, %r9
        mov     %rax, VMM_NATIVE_VMCALL_SIGNATURE
        vmcall
        mov     %r9, %rsi
        mov     %r8, %rdi
        pop     %rsi
        pop     %rdi
        ret

ITP_JMP_DEADLOOP PROC
        jmp $
        ret
ITP_JMP_DEADLOOP ENDP