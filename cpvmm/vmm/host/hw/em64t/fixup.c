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
#include "vmm_defs.h"
#define VMM_NATIVE_VMCALL_SIGNATURE 0x024694D40
#ifdef JLMDEBUG
#include "bootstrap_print.h"
#include "jlmdebug.h"

UINT64   t_vmcs_save_area[512];  // never bigger than 4KB
extern void vmm_print_vmcs_region(UINT64* pu);
extern void vmm_vmcs_guest_state_read(UINT64* area);
#endif

extern void vmm_vmcs_guest_state_read(UINT64* area);
extern int vmx_vmread(UINT64 index, UINT64 *value);
extern int vmx_vmwrite(UINT64 index, UINT64 value);


#ifdef JLMDEBUG
typedef unsigned char      uint8_t;
typedef unsigned short     uint16_t;
typedef unsigned int       uint32_t;
typedef long long unsigned uint64_t;

typedef int                 bool;
typedef unsigned char       u8;
typedef unsigned short      u16;
typedef unsigned int        u32;
typedef long long unsigned  u64;
#include "../../../bootstrap/linux_defns.h"
void check_boot_parameters(UINT64 rsi_reg)
{
    bprint("rsi on entry: %p\n", rsi_reg);
    boot_params_t* boot_params= (boot_params_t*) rsi_reg;
    HexDump((UINT8*)rsi_reg, (UINT8*)rsi_reg+32);
    bprint("cmd line ptr: %p\n", boot_params->hdr.cmd_line_ptr);
    bprint("code32_start: %p\n", boot_params->hdr.code32_start);
}
#endif


// fixup control registers and make guest loop forever

asm(
".text\n"
".globl loop_forever\n"
".type loop_forever, @function\n"
"loop_forever:\n"
    "\tjmp   .\n"
    "\tret\n"
);


void fixupvmcs()
{
    UINT64  value;
    UINT64  rsi_reg;
    void loop_forever();
    UINT16* loop= (UINT16*)loop_forever;

#ifdef JLMDEBUG
    bprint("fixupvmcs %04x\n", *loop);
    vmx_vmread(0x681e, &value);  // guest_rip
    // bprint("Code at %p\n", value);
    // HexDump((UINT8*)value, (UINT8*)value+32);
    *((UINT16*) value+8)= *loop;    // feeb
    asm volatile (
        "\t movq   %%rsi, %[rsi_reg]\n"
    : [rsi_reg] "=g" (rsi_reg)
    ::);
    check_boot_parameters(rsi_reg);
#endif

    // was 3e, cruse has 16
    // vmx_vmread(0x4000, &value);  // vmx_pin_controls
    // value= 0x16;
    // vmx_vmwrite(0x4000, value);  // vmx_pin_controls

    // was 96006172, cruse has 401e172
    // vmx_vmread(0x4002, &value);  // vmx_cpu_controls
    // value= 0x80016172;         // can't figure out anything to change here
    // value= 0x96006172;         // can't figure out anything to change here
    // vmx_vmwrite(0x4002, value);  // vmx_cpu_controls

    // vmx_vmread(0x401e, &value);  // vmx_secondary_controls
    // value= 0x8a;                 // no vpid
    // vmx_vmwrite(0x401e, value);  // vmx_secondary_controls

    // was d1ff, cruse has 11ff 
    // vmx_vmread(0x4012, &value);  // vmx_entry_controls
    // value= 0x11ff;
    // vmx_vmwrite(0x4012, value);  // vmx_entry_controls

    // was 3f7fff, cruse has 36fff
    // vmx_vmread(0x4002, &value);  // vmx_exit_controls
    // vmx_vmwrite(0x4002, value);  // vmx_exit_controls

    vmm_vmcs_guest_state_read((UINT64*) t_vmcs_save_area);
    vmm_print_vmcs_region((UINT64*) t_vmcs_save_area);
}


