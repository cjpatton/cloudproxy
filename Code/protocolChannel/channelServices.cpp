//
//  File: channelServices.cpp
//  Description: cloudProxy channel services prototype
//  Copyright (c) 2011, Intel Corporation. Some contributions 
//    (c) John Manferdelli.  All rights reserved.
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


// -----------------------------------------------------------------------------


#include "jlmTypes.h"
#include "jlmUtility.h"
#include "keys.h"
#include "session.h"
#include "channel.h"
#include "channelServices.h"
#include "safeChannel.h"
#include "request.h"
#include "timer.h"


// ------------------------------------------------------------------------------


channelServices::channelServices(u32 type)
{
    m_serviceType= 0;
    m_fserviceEnabled= false;
}


channelServices::~channelServices()
{
    m_serviceType= 0;
    m_fserviceEnabled= false;
}


// ------------------------------------------------------------------------------

