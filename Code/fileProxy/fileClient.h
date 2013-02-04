//
//  File: fileClient.h
//      John Manferdelli
//
//  Description: Symbol and class definitions for fileClient
//
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


//------------------------------------------------------------------------------------


#ifndef _FILECLIENT__H
#define _FILECLIENT__H


#include "channel.h"
#include "safeChannel.h"
#include "session.h"
#include "objectManager.h"
#include "resource.h"
#include "secPrincipal.h"
#include "tao.h"
#include "vault.h"


class fileClient {
public:
    int                 m_clientState;
    bool                m_fChannelAuthenticated;

    taoHostServices     m_host;
    taoEnvironment      m_tcHome;

    bool                m_fEncryptFiles;
    char*               m_szSealedKeyFile;
    bool                m_fKeysValid;
    u32                 m_uAlg;
    u32                 m_uMode;
    u32                 m_uPad;
    u32                 m_uHmac;
    int                 m_sizeKey;
    byte                m_fileKeys[SMALLKEYSIZE];

    int	                m_fd;
    sessionKeys         m_oKeys;
    char*               m_szPort;
    char*               m_szAddress;

    fileClient();
    ~fileClient();

    bool    initClient(char* configDirectory);
    bool    initPolicy();
    bool    initFileKeys();
    bool    closeClient();
    bool    initSafeChannel(safeChannel& fc);
    bool    protocolNego(int fd, safeChannel& fc);
};


#define SERVICENAME             "fileServer"
#define SERVICEADDRESS          "127.0.0.1"
#define SERVICE_PORT            6000


#endif


//-------------------------------------------------------------------------------


