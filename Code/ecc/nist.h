//
//  File: nist.h:
//  Description: nist curves
//
//  Copyright (c) 2014, John Manferdelli.  All rights reserved.
//
// Use, duplication and disclosure of this file and derived works of
// this file are subject to and licensed under the Apache License dated
// January, 2004, (the "License").  This License is contained in the
// top level directory originally provided with the CloudProxy Project.
// Your right to use or distribute this file, or derived works thereof,
// is subject to your being bound by those terms and your use indicates
// consent to those terms.
//
// If you distribute this file (or portions derived therefrom), you must
// include License in or with the file and, in the event you do not include
// the entire License in the file, the file must contain a reference
// to the location of the License.


// ----------------------------------------------------------------------------


#ifndef _NIST_H
#define _NIST_H


#include "jlmTypes.h"
#include "bignum.h"
#include "ecc.h"


extern ECurve  nist256curve;
extern ECurve  nist521curve;
extern bool initNist();


#endif    // _NIST_H


// ----------------------------------------------------------------------------
