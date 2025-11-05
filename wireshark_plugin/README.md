# QOTP Wireshark Plugin

This plugin enables Wireshark to dissect and decrypt QOTP (Quick UDP Transport Protocol) traffic on port 8090.

## Architecture

The plugin consists of three components:

1. **qotp_export.go** - Go code that exports crypto functions as a C-compatible DLL
2. **qotp_decrypt.c** - C wrapper that bridges Go functions to Lua
3. **qotp_dissector.lua** - Wireshark Lua dissector plugin

## Building

### Prerequisites

- Go compiler (with CGO enabled)
- Visual Studio 2022 with C++ compiler
- Lua 5.4 headers and libraries
- Wireshark installed

### Build Steps

1. Initialize Visual Studio environment:
   ```powershell
   vcvars64.bat
   ```

2. Run the build script:
   ```powershell
   .\build.ps1
   ```

Or build manually:

```powershell
# Build Go DLL
$env:CGO_ENABLED = "1"
go build -buildmode=c-shared -o qotp_crypto.dll qotp_export.go

# Build C Lua module  
cmd /c "vcvars64.bat && cl /LD /O2 /TP qotp_decrypt.c /I""path\to\lua\include"" /link ""path\to\lua54.lib"" qotp_crypto.lib User32.lib /OUT:qotp_decrypt.dll"
```

## Deployment

1. Copy DLLs to Wireshark directory:
   ```
   qotp_decrypt.dll → C:\Program Files\Wireshark\
   qotp_crypto.dll  → C:\Program Files\Wireshark\
   ```

2. Copy Lua plugin to plugins directory:
   ```
   qotp_dissector.lua → C:\Users\<user>\AppData\Roaming\Wireshark\plugins\4.6\
   ```

3. Restart Wireshark

## Usage

### Basic Dissection

1. Open Wireshark and load a PCAP file with QOTP traffic on UDP port 8090
2. The dissector will automatically decode QOTP packets showing:
   - Message type (InitSnd, InitRcv, Data, etc.)
   - Connection ID
   - Encrypted data

### Decryption with Keylog File

1. Create a keylog file (see `keylog.example` for format):
   ```
   # Format: CONNECTION_ID SHARED_SECRET_HEX
   0x1234567890abcdef 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
   ```

2. In Wireshark, go to: **Edit → Preferences → Protocols → QOTP**

3. Set "Keylog file" to the path of your keylog file

4. Click OK and reload the capture

5. Packets with matching connection IDs will be decrypted

## Important Notes

### Dynamic Linking to Lua

The DLL **MUST** dynamically link to `lua54.dll` to avoid the "multiple Lua VMs detected" crash:

- ✅ Link against: `lua54.lib` (import library)
- ❌ Do NOT link against: static Lua libraries (liblua.a, lua.lib)

### Wireshark Lua Version

Check your Wireshark installation for the correct Lua version:
- Look for `lua52.dll`, `lua53.dll`, or `lua54.dll` in `C:\Program Files\Wireshark\`
- Adjust the build to link against the matching version

## Troubleshooting

### "module 'qotp_decrypt' not found"

- Ensure `qotp_decrypt.dll` is in `C:\Program Files\Wireshark\`
- Ensure `qotp_crypto.dll` is also present (runtime dependency)
- Check that `lua54.dll` exists in the same directory

### Build Errors

- **"stdio.h not found"**: Run from VS Developer Command Prompt or init vcvars64.bat
- **"lua.h not found"**: Check the Lua include path in build command
- **Linker errors**: Ensure qotp_crypto.lib was created by the Go build

### Runtime Crashes

- Verify dynamic linking: `dumpbin /dependents qotp_decrypt.dll` should show `lua54.dll`
- Ensure Go DLL is the correct architecture (x64)
