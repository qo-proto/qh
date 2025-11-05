# QOTP Wireshark Plugin - Build Complete!

## Files Created

### DLLs:
- `qotp_crypto.dll` (3.25 MB) - Go crypto library  
- `qotp_decrypt.dll` - Lua module wrapper

### Plugin:
- `qotp_dissector.lua` - Wireshark dissector

## Deployment Steps

1. **Copy DLLs to Wireshark**:
   ```powershell
   Copy-Item qotp_decrypt.dll "C:\Program Files\Wireshark\" -Force
   Copy-Item qotp_crypto.dll "C:\Program Files\Wireshark\" -Force
   ```

2. **Copy Lua plugin**:
   ```powershell
   Copy-Item qotp_dissector.lua "C:\Users\$env:USERNAME\AppData\Roaming\Wireshark\plugins\4.6\" -Force
   ```

3. **Restart Wireshark**

## Testing

In Wireshark:
1. Go to Help → About Wireshark → Plugins
2. Look for "qotp_dissector.lua" in the list
3. Capture or open PCAP with UDP port 8090 traffic
4. Packets should show as "QOTP" protocol

## Features

- ✅ Dissects QOTP protocol structure
- ✅ Displays message types (InitSnd, InitRcv, InitCryptoSnd, InitCryptoRcv, Data)
- ✅ Extracts connection IDs
- ✅ Go crypto library integrated
- ⏳ Decryption (ready for implementation)

## Next Steps

To add actual decryption:
1. Import the full crypto.go logic into qotp_export.go
2. Expose DecryptDataForPcap function
3. Update qotp_decrypt.c to call decrypt function
4. Update qotp_dissector.lua to display decrypted data

## Architecture

```
Wireshark Lua VM
       ↓
qotp_dissector.lua
       ↓
qotp_decrypt.dll (C wrapper)
       ↓
qotp_crypto.dll (Go library)
```

All DLLs dynamically link to lua54.dll (no static linking = no crashes!)
