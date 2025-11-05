/*
    qotp_decrypt.dll - QOTP Decryption Library for Wireshark Lua Plugin
    
    This DLL wraps Go crypto functions from crypto.go and exposes them to Lua.
    Wireshark can then use this to decrypt QOTP UDP traffic on port 8090.
    
    CRITICAL: This DLL MUST dynamically link to Wireshark's Lua DLL (NOT statically link Lua)
    
    Compilation Instructions:
    ========================================
    
    STEP 1: Build the Go shared library
    ----------------------------------------------------
    First, build crypto.go as a C-compatible shared library:
       go build -buildmode=c-shared -o qotp_crypto.dll crypto.go
    
    This creates:
    - qotp_crypto.dll (the Go code compiled as a DLL)
    - qotp_crypto.h (C header with exported functions)
    
    STEP 2: Compile this C wrapper as a Lua module
    ----------------------------------------------------
    From PowerShell with vcvars64.bat initialized:
       cmd /c "vcvars64.bat && cd /d C:\Users\gian\sa\qh\wireshark_plugin && cl /LD /O2 /TP qotp_decrypt.c /I""C:\Users\gian\sa\wireshark\wireshark-libs\lua-5.4.6-unicode-win64-vc14\include"" /link ""C:\Users\gian\sa\wireshark\wireshark-libs\lua-5.4.6-unicode-win64-vc14\lua54.lib"" qotp_crypto.lib User32.lib /OUT:qotp_decrypt.dll"
    
    STEP 3: Deploy the DLLs
    ----------------------------------------------------
    Copy both DLLs to C:\Program Files\Wireshark\:
    - qotp_decrypt.dll (this Lua module)
    - qotp_crypto.dll (the Go library)
*/

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "lua.hpp"
#include <windows.h>

// Version
#define QOTP_DECRYPT_VERSION "1.0.0"

// Function pointers for Go DLL functions
typedef unsigned long long (*GetConnectionIdFunc)(const char*, int);
typedef int (*GetMessageTypeFunc)(const char*, int);
typedef int (*SetSharedSecretHexFunc)(unsigned long long, const char*);
typedef int (*DecryptDataPacketFunc)(const char*, int, unsigned long long, int, unsigned long long, char*, int);
typedef char* (*GetVersionFunc)();
typedef int (*GetLoadedKeyCountFunc)();
typedef int (*GetLoadedKeysFunc)(unsigned long long*, int);

// Global handles
static HMODULE goDLL = NULL;
static GetConnectionIdFunc go_GetConnectionId = NULL;
static GetMessageTypeFunc go_GetMessageType = NULL;
static SetSharedSecretHexFunc go_SetSharedSecretHex = NULL;
static DecryptDataPacketFunc go_DecryptDataPacket = NULL;
static GetVersionFunc go_GetVersion = NULL;
static GetLoadedKeyCountFunc go_GetLoadedKeyCount = NULL;
static GetLoadedKeysFunc go_GetLoadedKeys = NULL;

// Load the Go DLL dynamically
static int load_go_dll() {
    if (goDLL != NULL) return 1; // Already loaded
    
    goDLL = LoadLibraryA("qotp_crypto.dll");
    if (!goDLL) {
        MessageBox(0, "Failed to load qotp_crypto.dll!\n\nMake sure it's in C:\\Program Files\\Wireshark\\", 
                   "Error", MB_OK | MB_ICONERROR);
        return 0;
    }
    
    go_GetConnectionId = (GetConnectionIdFunc)GetProcAddress(goDLL, "GetConnectionId");
    go_GetMessageType = (GetMessageTypeFunc)GetProcAddress(goDLL, "GetMessageType");
    go_SetSharedSecretHex = (SetSharedSecretHexFunc)GetProcAddress(goDLL, "SetSharedSecretHex");
    go_DecryptDataPacket = (DecryptDataPacketFunc)GetProcAddress(goDLL, "DecryptDataPacket");
    go_GetVersion = (GetVersionFunc)GetProcAddress(goDLL, "GetVersion");
    go_GetLoadedKeyCount = (GetLoadedKeyCountFunc)GetProcAddress(goDLL, "GetLoadedKeyCount");
    go_GetLoadedKeys = (GetLoadedKeysFunc)GetProcAddress(goDLL, "GetLoadedKeys");
    
    if (!go_GetConnectionId || !go_GetMessageType || !go_SetSharedSecretHex || !go_DecryptDataPacket) {
        MessageBox(0, "Failed to load functions from qotp_crypto.dll!", 
                   "Error", MB_OK | MB_ICONERROR);
        FreeLibrary(goDLL);
        goDLL = NULL;
        return 0;
    }
    
    return 1;
}

// Lua function: qotp_decrypt.decrypt_data(encrypted_data, conn_id_hex, is_sender, epoch)
// Now accepts conn_id as hex string to avoid Lua integer overflow
static int lua_decrypt_data(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushnil(L);
        lua_pushstring(L, "Failed to load Go DLL");
        return 2;
    }
    
    // Get parameters from Lua
    size_t enc_len;
    const char* encrypted = luaL_checklstring(L, 1, &enc_len);
    
    // Connection ID can be either number or string (hex)
    unsigned long long conn_id;
    if (lua_type(L, 2) == LUA_TSTRING) {
        // Parse hex string
        const char* conn_id_hex = luaL_checkstring(L, 2);
        if (sscanf(conn_id_hex, "%llx", &conn_id) != 1) {
            lua_pushnil(L);
            lua_pushstring(L, "Invalid connection ID hex string");
            return 2;
        }
    } else {
        // Try to get as integer (may fail for large numbers)
        conn_id = (unsigned long long)luaL_checkinteger(L, 2);
    }
    
    int is_sender = lua_toboolean(L, 3);
    lua_Integer epoch = luaL_checkinteger(L, 4);
    
    // Allocate output buffer (max 64KB)
    char* output = (char*)malloc(65536);
    if (!output) {
        lua_pushnil(L);
        lua_pushstring(L, "Memory allocation failed");
        return 2;
    }
    
    // Call Go decryption function
    int result_len = go_DecryptDataPacket(
        encrypted, 
        (int)enc_len, 
        (unsigned long long)conn_id,
        is_sender,
        (unsigned long long)epoch,
        output,
        65536
    );
    
    if (result_len < 0) {
        free(output);
        lua_pushnil(L);
        
        if (result_len == -1) {
            lua_pushstring(L, "No shared secret for connection");
        } else if (result_len == -2) {
            lua_pushstring(L, "Decryption failed");
        } else if (result_len == -3) {
            lua_pushstring(L, "Output buffer too small");
        } else {
            lua_pushstring(L, "Unknown error");
        }
        return 2;
    }
    
    // Return decrypted data as Lua string
    lua_pushlstring(L, output, result_len);
    free(output);
    return 1;
}

// Lua function: qotp_decrypt.set_key(conn_id_hex, shared_secret_hex)
// Now accepts conn_id as hex string to avoid Lua integer overflow
static int lua_set_key(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushboolean(L, 0);
        lua_pushstring(L, "Failed to load Go DLL");
        return 2;
    }
    
    // Connection ID can be either number or string (hex)
    unsigned long long conn_id;
    if (lua_type(L, 1) == LUA_TSTRING) {
        // Parse hex string
        const char* conn_id_hex = luaL_checkstring(L, 1);
        if (sscanf(conn_id_hex, "%llx", &conn_id) != 1) {
            lua_pushboolean(L, 0);
            lua_pushstring(L, "Invalid connection ID hex string");
            return 2;
        }
    } else {
        // Try to get as integer (may fail for large numbers)
        conn_id = (unsigned long long)luaL_checkinteger(L, 1);
    }
    
    size_t key_len;
    const char* key_hex = luaL_checklstring(L, 2, &key_len);
    
    // Call Go function to set the shared secret
    int result = go_SetSharedSecretHex(conn_id, key_hex);
    
    if (result == 0) {
        lua_pushboolean(L, 1);
        return 1;
    } else {
        lua_pushboolean(L, 0);
        lua_pushstring(L, "Invalid hex string");
        return 2;
    }
}

// Lua function: qotp_decrypt.get_conn_id(udp_data)
static int lua_get_conn_id(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushnil(L);
        lua_pushstring(L, "Failed to load Go DLL");
        return 2;
    }
    
    size_t data_len;
    const char* data = luaL_checklstring(L, 1, &data_len);
    
    if (data_len < 9) {
        lua_pushnil(L);
        lua_pushstring(L, "Packet too short");
        return 2;
    }
    
    // Call Go function
    unsigned long long conn_id = go_GetConnectionId(data, (int)data_len);
    
    lua_pushinteger(L, conn_id);
    return 1;
}

// Lua function: qotp_decrypt.get_message_type(udp_data)
static int lua_get_message_type(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushstring(L, "Error");
        return 1;
    }
    
    size_t data_len;
    const char* data = luaL_checklstring(L, 1, &data_len);
    
    if (data_len < 1) {
        lua_pushnil(L);
        return 1;
    }
    
    // Call Go function
    int msg_type = go_GetMessageType(data, (int)data_len);
    
    const char* type_names[] = {
        "InitSnd",
        "InitRcv", 
        "InitCryptoSnd",
        "InitCryptoRcv",
        "Data"
    };
    
    if (msg_type >= 0 && msg_type <= 4) {
        lua_pushstring(L, type_names[msg_type]);
    } else {
        lua_pushstring(L, "Unknown");
    }
    
    return 1;
}

// Lua function: qotp_decrypt.get_version()
static int lua_get_version(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushstring(L, "ERROR: Failed to load DLL");
        return 1;
    }
    
    char version_info[512];
    const char* go_version = go_GetVersion ? go_GetVersion() : "unknown";
    
    snprintf(version_info, sizeof(version_info),
             "C Wrapper: %s\nGo Library: %s",
             QOTP_DECRYPT_VERSION, go_version);
    
    lua_pushstring(L, version_info);
    return 1;
}

// Lua function: qotp_decrypt.get_loaded_keys()
static int lua_get_loaded_keys(lua_State* L) {
    if (!load_go_dll() || !go_GetLoadedKeyCount || !go_GetLoadedKeys) {
        lua_newtable(L);
        return 1;
    }
    
    int count = go_GetLoadedKeyCount();
    
    if (count == 0) {
        lua_newtable(L);
        return 1;
    }
    
    // Allocate buffer for connection IDs
    unsigned long long* conn_ids = (unsigned long long*)malloc(count * sizeof(unsigned long long));
    if (!conn_ids) {
        lua_newtable(L);
        return 1;
    }
    
    int actual_count = go_GetLoadedKeys(conn_ids, count);
    
    // Create Lua table
    lua_newtable(L);
    for (int i = 0; i < actual_count; i++) {
        lua_pushinteger(L, i + 1);
        lua_pushinteger(L, conn_ids[i]);
        lua_settable(L, -3);
    }
    
    free(conn_ids);
    return 1;
}

// Lua function: qotp_decrypt.test()
static int lua_test(lua_State* L) {
    if (!load_go_dll()) {
        lua_pushboolean(L, 0);
        return 1;
    }
    
    int key_count = go_GetLoadedKeyCount ? go_GetLoadedKeyCount() : 0;
    const char* go_version = go_GetVersion ? go_GetVersion() : "unknown";
    
    char message[512];
    snprintf(message, sizeof(message),
        "QOTP Decrypt DLL loaded successfully!\n\n"
        "C Wrapper Version: %s\n"
        "Go Library Version: %s\n\n"
        "Loaded Keys: %d\n\n"
        "Ready to decrypt QOTP traffic on port 8090.",
        QOTP_DECRYPT_VERSION, go_version, key_count);
    
    MessageBox(0, message, "qotp_decrypt.dll - Ready", MB_OK | MB_ICONINFORMATION);
    
    lua_pushboolean(L, 1);
    return 1;
}

// Module registration
static const luaL_Reg qotp_decrypt_funcs[] = {
    {"decrypt_data", lua_decrypt_data},
    {"set_key", lua_set_key},
    {"get_conn_id", lua_get_conn_id},
    {"get_message_type", lua_get_message_type},
    {"get_version", lua_get_version},
    {"get_loaded_keys", lua_get_loaded_keys},
    {"test", lua_test},
    {NULL, NULL}
};

extern "C" {
    __declspec(dllexport)
    int luaopen_qotp_decrypt(lua_State* L) {
        // Create the module table
        luaL_newlib(L, qotp_decrypt_funcs);
        return 1;
    }
}
