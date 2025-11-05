-- QOTP and QH Protocol Dissector for Wireshark

-- ChaCha20 implementation
local ChaCha20 = {}

function ChaCha20.quarter_round(state, a, b, c, d)
    state[a] = (state[a] + state[b]) & 0xffffffff
    state[d] = bit.rol((state[d] ~ state[a]), 16)
    state[c] = (state[c] + state[d]) & 0xffffffff
    state[b] = bit.rol((state[b] ~ state[c]), 12)
    state[a] = (state[a] + state[b]) & 0xffffffff
    state[d] = bit.rol((state[d] ~ state[a]), 8)
    state[c] = (state[c] + state[d]) & 0xffffffff
    state[b] = bit.rol((state[b] ~ state[c]), 7)
end

function ChaCha20.create_state(key, nonce, counter)
    local state = {
        -- Constants "expand 32-byte k" in little-endian
        0x61707865, 0x3320646e, 0x79622d32, 0x6b206574
    }

    -- Key (8 words)
    for i = 0, 7 do
        local k = i * 4
        state[i + 5] = string.byte(key, k + 1) +
                       (string.byte(key, k + 2) << 8) +
                       (string.byte(key, k + 3) << 16) +
                       (string.byte(key, k + 4) << 24)
    end

    -- Counter
    state[13] = counter

    -- Nonce (3 words)
    for i = 0, 2 do
        local n = i * 4
        local word = 0
        for j = 0, 3 do
            local byte = string.byte(nonce, n + j + 1) or 0
            word = word | (byte << (j * 8))
        end
        state[i + 14] = word
    end

    return state
end

function ChaCha20.create_keystream(key, nonce, counter)
    local state = ChaCha20.create_state(key, nonce, counter)
    local working_state = {}
    local keystream = {}
    
    -- Copy state
    for j = 1, 16 do
        working_state[j] = state[j]
    end
    
    -- Apply 20 rounds
    for _ = 1, 10 do
        -- Column rounds
        ChaCha20.quarter_round(working_state, 1, 5, 9, 13)
        ChaCha20.quarter_round(working_state, 2, 6, 10, 14)
        ChaCha20.quarter_round(working_state, 3, 7, 11, 15)
        ChaCha20.quarter_round(working_state, 4, 8, 12, 16)
        -- Diagonal rounds
        ChaCha20.quarter_round(working_state, 1, 6, 11, 16)
        ChaCha20.quarter_round(working_state, 2, 7, 12, 13)
        ChaCha20.quarter_round(working_state, 3, 8, 9, 14)
        ChaCha20.quarter_round(working_state, 4, 5, 10, 15)
    end
    
    -- Add working state to original
    for j = 1, 16 do
        local v = (working_state[j] + state[j]) & 0xffffffff
        -- Convert word to bytes
        keystream[#keystream + 1] = string.char(v & 0xff)
        keystream[#keystream + 1] = string.char((v >> 8) & 0xff)
        keystream[#keystream + 1] = string.char((v >> 16) & 0xff)
        keystream[#keystream + 1] = string.char((v >> 24) & 0xff)
    end
    
    return table.concat(keystream)
end

function ChaCha20.decrypt(key, nonce, ciphertext, counter)
    local keystream = ChaCha20.create_keystream(key, nonce, counter)
    local plaintext = {}
    
    -- XOR ciphertext with keystream
    for i = 1, #ciphertext do
        table.insert(plaintext, string.char(
            bit.bxor(string.byte(ciphertext, i), string.byte(keystream, i))
        ))
    end
    
    return table.concat(plaintext)
end

-- ... rest of your existing QOTP/QH dissector code ...

-- Add to QOTP protocol preferences
qotp_protocol.prefs.keylog_path = Pref.string("Keylog file", "", "Path to the QOTP keylog file")

-- Add decryption function to QOTP dissector
function try_decrypt_payload(buffer, pinfo, conn_id, is_from_server)
    -- Load secrets from keylog file
    if not secrets then
        load_secrets(qotp_protocol.prefs.keylog_path)
    end

    local secret = secrets[conn_id]
    if not secret then return nil end

    -- Extract encrypted portions
    local encrypted_sn = buffer(9, 6)  -- After header and conn_id
    local ciphertext = buffer(15)      -- Rest is payload

    -- Decrypt sequence number first
    local sn_nonce = string.rep("\0", 12)  -- 12 bytes of zeros
    local sn_decrypted = ChaCha20.decrypt(secret, sn_nonce, encrypted_sn:raw(), 1)
    
    -- Convert to number
    local sn = 0
    for i = 1, 6 do
        sn = sn + (string.byte(sn_decrypted, i) << ((i-1) * 8))
    end

    -- Try different epochs
    for epoch in {0, -1, 1} do
        -- Construct nonce
        local nonce = string.rep("\0", 12)
        -- Add epoch (first 6 bytes)
        for i = 0, 5 do
            nonce = string.sub(nonce, 1, i) .. 
                   string.char((epoch >> (i * 8)) & 0xFF) ..
                   string.sub(nonce, i + 2)
        end
        -- Add sequence number (next 6 bytes)
        for i = 0, 5 do
            nonce = string.sub(nonce, 1, i + 6) ..
                   string.char((sn >> (i * 8)) & 0xFF) ..
                   string.sub(nonce, i + 8)
        end

        -- Set sender bit
        if not is_from_server then
            nonce = string.char(string.byte(nonce, 1) | 0x80) .. string.sub(nonce, 2)
        end

        -- Try decryption
        local plaintext = ChaCha20.decrypt(secret, nonce, ciphertext:raw(), 0)
        if plaintext then
            return ByteArray.tvb(ByteArray.new(plaintext), "Decrypted Payload")
        end
    end

    return nil
end

function ChaCha20.create_keystream(key, nonce, counter)
    -- Create initial state
    local state = ChaCha20.create_state(key, nonce, counter)
    local working_state = {}
    local keystream = {}
    
    -- Copy state to working state (16 words = 64 bytes)
    for j = 1, 16 do
        working_state[j] = state[j]
    end
    
    -- Apply 20 rounds (10 iterations of the double round function)
    -- This matches the ChaCha20 specification
    for _ = 1, 10 do
        ChaCha20.inner_block(working_state)
    end
    
    -- Add working state to original state to create keystream
    -- Each word is 4 bytes in little-endian format
    for j = 1, 16 do
        local v = (working_state[j] + state[j]) & 0xffffffff
        -- Convert word to bytes in little-endian order
        keystream[#keystream + 1] = string.char(v & 0xff)
        keystream[#keystream + 1] = string.char((v >> 8) & 0xff)
        keystream[#keystream + 1] = string.char((v >> 16) & 0xff)
        keystream[#keystream + 1] = string.char((v >> 24) & 0xff)
    end
    
    return table.concat(keystream)
end

function ChaCha20.encrypt(key, nonce, counter, plaintext)
    local keystream = ChaCha20.create_keystream(key, nonce, counter)
    local ciphertext = {}
    
    for i = 1, #plaintext, 64 do
        -- Copy state to working state
        for j = 1, 16 do
            working_state[j] = state[j]
        end
        
        -- Apply 20 rounds (10 iterations of the double round function)
        for _ = 1, 10 do
            ChaCha20.inner_block(working_state)
        end
        
        -- Add working state to original state
        for j = 1, 16 do
            local v = (working_state[j] + state[j]) & 0xffffffff
            for k = 1, 4 do
                table.insert(keystream, (v >> ((k-1) * 8)) & 0xff)
            end
        end
        
        state[13] = (state[13] + 1) & 0xffffffff
    end
    
    -- XOR plaintext with keystream
    for i = 1, #plaintext do
        table.insert(ciphertext, string.char(bit.bxor(string.byte(plaintext, i), keystream[i])))
    end
    
    return table.concat(ciphertext)
end

--[[
To use:
1. Save this file as `qotp_qh.lua`.
2. Place it in your personal Wireshark plugins directory.
   (Find this via Help > About Wireshark > Folders > Personal Lua Plugins).
3. Restart Wireshark.
4. Go to Edit > Preferences > Protocols > QOTP and set the path to your keylog file.
   The keylog file should be in the format:
   QOTP_SHARED_SECRET <connID_hex> <secret_hex>
--]]

-- Check for LuaJIT and bit library support
if not bit then
    -- For Wireshark versions that use plain Lua, load the bitop library.
    -- Most modern Wireshark versions include LuaJIT, so this is a fallback.
    bit = require("bit32") or require("bit")
end

-- Add bitwise rotation helpers
bit.rol = bit.rol or function(a, b)
    b = b & 31
    return ((a << b) & 0xffffffff) | (a >> (32 - b))
end

bit.ror = bit.ror or function(a, b)
    b = b & 31
    return (a >> b) | ((a << (32 - b)) & 0xffffffff)
end

-- =============================================================================
-- QH Dissector
-- =============================================================================
local qh_protocol = Proto("QH", "Quite Ok HTTP Protocol")

-- Debug fields
local f_debug_info = ProtoField.string("qotp.debug", "Debug Info")
local f_key_info = ProtoField.string("qotp.key_info", "Key Info")
local f_nonce_info = ProtoField.string("qotp.nonce_info", "Nonce Info")

-- QH Fields
local f_qh_version = ProtoField.uint8("qh.version", "Version", base.DEC)
-- Request fields
local f_qh_method = ProtoField.uint8("qh.method", "Method", base.DEC, {
    [0] = "GET",
    [1] = "POST"
}, 0x07) -- Add mask 0x07 for the 3 bits used for method
local f_qh_host = ProtoField.string("qh.host", "Host", base.ASCII)
local f_qh_path = ProtoField.string("qh.path", "Path", base.ASCII)
-- Response fields
local f_qh_status_compact = ProtoField.uint8("qh.status_code_compact", "Status Code (Compact)", base.DEC)
local f_qh_status = ProtoField.uint16("qh.status_code", "Status Code", base.DEC)
-- Common fields
local f_qh_headers = ProtoField.string("qh.headers", "Headers", base.ASCII)
local f_qh_header_item = ProtoField.string("qh.header", "Header", base.ASCII)
local f_qh_body = ProtoField.bytes("qh.body", "Body", base.NONE)
local f_qh_body_str = ProtoField.string("qh.body_str", "Body (as String)", base.ASCII)

qh_protocol.fields = {
    f_qh_version, f_qh_method, f_qh_host, f_qh_path, f_qh_status_compact, f_qh_status, f_qh_headers, f_qh_body, f_qh_body_str
}

-- Request header names mapping (from types.go)
local req_header_names = {
    [0]  = "Custom",
    [1]  = "Accept",
    [2]  = "Accept-Encoding",
    [4]  = "Accept-Language",
    [5]  = "Content-Type",
    [6]  = "Content-Length",
    [7]  = "Authorization",
    [8]  = "Cookie",
    [9]  = "User-Agent",
    [10] = "Referer",
    [11] = "Origin",
    [12] = "If-None-Match",
    [13] = "If-Modified-Since",
    [14] = "Range",
    [15] = "X-Payment",
}

-- Response header names mapping (from types.go)
local resp_header_names = {
    [0]  = "Custom",
    [1]  = "Content-Type",
    [2]  = "Content-Length",
    [4]  = "Cache-Control",
    [5]  = "Content-Encoding",
    [6]  = "Date",
    [7]  = "ETag",
    [8]  = "Expires",
    [9]  = "Last-Modified",
    [10] = "Access-Control-Allow-Origin",
    [11] = "Access-Control-Allow-Methods",
    [12] = "Access-Control-Allow-Headers",
    [13] = "Set-Cookie",
    [14] = "Location",
    [15] = "Content-Security-Policy",
    [16] = "X-Content-Type-Options",
    [17] = "X-Frame-Options",
    [18] = "Vary",
    [19] = "X-Payment-Response",
}

-- This is the dissector for the PLAINTEXT QH data after decryption.
function qh_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qh_protocol.name
    local subtree = tree:add(qh_protocol, buffer(), "QH Application Data (Decrypted)")
    -- If a display filter is active, the subtree might not be created.
    if not subtree then
        return
    end

    if buffer:len() < 1 then
        return -- Not enough data to dissect
    end

    -- Heuristic to determine if it's a request or response.
    -- If the source port is the well-known port, it's a response.
    local is_response = pinfo.src_port == 8090

    local offset = 0
    -- Decode the first byte
    local header_byte = buffer(offset, 1):uint()
    local version = bit.band(bit.rshift(header_byte, 6), 0x03) -- Upper 2 bits
    local version_field = subtree:add(f_qh_version, buffer(offset, 1), version)
    offset = offset + 1

    -- Find the ETX character (\x03) which separates headers from the body
    local etx_offset = string.find(buffer:raw(), string.char(3), 1, true)
    if not etx_offset then
        -- No body separator found, cannot continue dissection
        return -- Cannot continue dissection
    end

    -- This function parses the binary header format: <id>\0<value>\0...
    local function parse_headers(header_buffer, header_map, headers_tree)
        local header_offset = 0
        while header_offset < header_buffer:len() do
            local header_id = header_buffer(header_offset, 1):uint()
            header_offset = header_offset + 1
            if header_offset >= header_buffer:len() or header_buffer(header_offset, 1):uint() ~= 0 then break end
            header_offset = header_offset + 1 -- Skip null after ID

            local key = header_map[header_id] or "Unknown Header (" .. header_id .. ")"
            if header_id == 0 then -- Custom header
                local key_end = string.find(header_buffer:raw(), string.char(0), header_offset + 1, true)
                if not key_end then break end
                key = header_buffer(header_offset, key_end - 1 - header_offset):string()
                header_offset = key_end -- Move offset to null after key
            end

            local value_end = string.find(header_buffer:raw(), string.char(0), header_offset + 1, true)
            if not value_end then break end
            local value = header_buffer(header_offset, value_end - 1 - header_offset):string()
            header_offset = value_end -- Move offset to null after value

            headers_tree:add(f_qh_header_item, string.format("%s: %s", key, value))
        end
    end

    if is_response then
        -- Dissect as a Response
        local compact_status = bit.band(header_byte, 0x3F) -- Lower 6 bits (0-63)
        -- version_field can be nil if a display filter is active
        if version_field then
            version_field:add_le(f_qh_status_compact, buffer(offset - 1, 1), compact_status)
        end
        pinfo.cols.info:set(string.format("Response, Status (Compact): %d", compact_status))

        local header_data_len = etx_offset - 1 - offset
        local headers_tree = subtree:add(buffer(offset, header_data_len), "Headers")
        parse_headers(buffer(offset, header_data_len), resp_header_names, headers_tree)

    else
        -- Dissect as a Request
        local method = bit.band(bit.rshift(header_byte, 3), 0x07) -- Middle 3 bits for method
        local method_str = f_qh_method.valuestring[method] or "Unknown"
        if version_field then
            version_field:add_le(f_qh_method, buffer(offset - 1, 1), method)
        end

        -- Find null terminators for host and path
        local data = buffer(offset):raw()
        local host_end = string.find(data, string.char(0), 1, true)
        if not host_end then return end -- No host terminator found
        
        local path_end = string.find(data, string.char(0), host_end + 1, true)
        if not path_end then return end -- No path terminator found

        -- Extract host and path (excluding null terminators)
        local host = data:sub(1, host_end - 1)
        local path = data:sub(host_end + 1, path_end - 1)
        if path == "" then path = "/" end

        -- Add fields to tree with proper ranges
        subtree:add(f_qh_host, buffer(offset, host_end - 1), host)
        subtree:add(f_qh_path, buffer(offset + host_end, path_end - host_end - 1), path)

        pinfo.cols.info:set(string.format("Request: %s %s", method_str, path))

        -- Dissect remaining headers
        local headers_start_offset = offset + path_end
        local header_data_len = etx_offset - 1 - headers_start_offset
        local headers_tree = subtree:add(buffer(headers_start_offset, header_data_len), "Headers")
        parse_headers(buffer(headers_start_offset, header_data_len), req_header_names, headers_tree)
    end

    -- The body is everything after the ETX character
    local body_offset = etx_offset -- The offset returned by string.find is 1-based index
    if body_offset < buffer:len() then
        local body_tvb = buffer(body_offset)
        subtree:add(f_qh_body, body_tvb)
        -- Also display the body as a string for better readability
        subtree:add(f_qh_body_str, body_tvb)
    end
end

-- =============================================================================
-- QOTP Dissector (handles decryption)
-- This is the main dissector that gets called for UDP port 8090.
-- =============================================================================
local qotp_protocol = Proto("QOTP", "Quite Ok Transport Protocol")

-- Register preferences
qotp_protocol.prefs.keylog_path = Pref.string("Keylog file", "", "Path to the QOTP keylog file")

-- QOTP Fields
local f_qotp_header_byte = ProtoField.uint8("qotp.header_byte", "Header Byte", base.HEX)
local f_qotp_msg_type = ProtoField.uint8("qotp.msg_type", "Message Type", base.DEC, {
    [0] = "InitSnd",
    [1] = "InitRcv",
    [2] = "InitCryptoSnd",
    [3] = "InitCryptoRcv",
    [4] = "Data"
})
local f_qotp_version = ProtoField.uint8("qotp.version", "Version", base.DEC)
local f_qotp_conn_id = ProtoField.uint64("qotp.conn_id", "Connection ID", base.HEX)
local f_qotp_payload_enc = ProtoField.bytes("qotp.payload_encrypted", "Encrypted Payload", base.NONE)

qotp_protocol.fields = {
    f_qotp_header_byte, f_qotp_msg_type, f_qotp_version, f_qotp_conn_id, f_qotp_payload_enc,
    f_debug_info, f_key_info, f_nonce_info
}

-- Debug helper function
local function add_debug(msg)
    if debug_tree then
        debug_tree:add(f_debug_info, msg)
    end
end

-- Store secrets loaded from the keylog file
local secrets = {
    loaded = false,
    data = {}
}

-- Function to load secrets from the keylog file
local function load_secrets(filepath)
    secrets.loaded = true
    secrets.data = {}
    local file, err = io.open(filepath, "r")
    if not file then
        -- Report an error if the file can't be opened.
        report_failure("QOTP: Could not open keylog file: " .. (err or "unknown error"))
        return
    end

    print("QOTP Debug: Loading secrets from " .. filepath)
    for line in file:lines() do
        -- Expected format: QOTP_SHARED_SECRET <connID_hex> <secret_hex>
        local _, _, label, conn_id_hex, secret_hex = string.find(line, "^(QOTP_SHARED_SECRET)%s+([0-9a-fA-F]+)%s+([0-9a-fA-F]+)$")
        if label and conn_id_hex and secret_hex then
            -- Store the secret as a raw byte string, using hex string as key
            local conn_id_key = string.lower(conn_id_hex)
            local secret_bytes = string.gsub(secret_hex, "(..)", function (c) return string.char(tonumber(c, 16)) end)
            secrets.data[conn_id_key] = secret_bytes
            print("QOTP Debug: Loaded secret for connection ID: " .. conn_id_key)
        else
            print("QOTP Debug: Skipping invalid line: " .. line)
        end
    end
    file:close()
end

-- This function attempts to decrypt a QOTP Data packet payload.
-- It mirrors the logic from qotp.DecryptDataForPcap in Go.
function try_decrypt_payload(buffer, pinfo, conn_id, is_from_server, debug_tree)
    if not secrets.loaded and qotp_protocol.prefs.keylog_path ~= "" then
        load_secrets(qotp_protocol.prefs.keylog_path)
    end

    -- Add debug info function
    local function add_debug(msg)
        if debug_tree then
            debug_tree:add(f_debug_info, msg)
        end
    end

    -- Debug: Print the connection ID we're looking for
    print("QOTP Debug: Looking for secret for connection ID: " .. conn_id)
    if secrets.data then
        local count = 0
        for k,_ in pairs(secrets.data) do
            count = count + 1
        end
        print("QOTP Debug: Number of loaded secrets: " .. count)
    else
        print("QOTP Debug: No secrets loaded")
    end

    local secret = secrets.data and secrets.data[conn_id]
    if not secret then
        print("QOTP Debug: No secret found for connection ID: " .. conn_id)
        return nil -- No secret found for this connection ID
    end
    print("QOTP Debug: Found secret for connection ID: " .. conn_id)

    -- Constants from crypto.go
    local HEADER_SIZE = 1
    local CONN_ID_SIZE = 8
    local SN_SIZE = 6
    local MAC_SIZE = 16
    local NONCE_SIZE = 12

    -- The encrypted portion starts after the QOTP header and ConnID
    local encrypted_portion = buffer(HEADER_SIZE + CONN_ID_SIZE)
    if encrypted_portion:len() < SN_SIZE + MAC_SIZE then
        return nil -- Not enough data to decrypt
    end

    -- Reconstruct the AAD (header) for a Data packet
    local data_msg_type = 4
    local crypto_version = 0
    local header_byte = bit.bor(bit.lshift(data_msg_type, 5), crypto_version)
    
    -- Create AAD as a string of bytes
    local aad = string.char(header_byte)
    -- Add connection ID bytes in little-endian order
    for i = 0, 7 do
        aad = aad .. string.char(tonumber(string.sub(conn_id, i*2+1, i*2+2), 16))
    end

    -- Extract components from the encrypted portion
    local encrypted_sn = encrypted_portion(0, SN_SIZE)
    local ciphertext_with_mac = encrypted_portion(SN_SIZE)

    -- Decrypt the sequence number first (openNoVerify logic from Go implementation)
    -- The nonce for the SN is the first 24 bytes of the main ciphertext
    -- This matches the Go code: snNonce := nonce[:chacha20poly1305.NonceSizeX]
    local sn_nonce = ciphertext_with_mac(0, 24):raw()
    add_debug(string.format("SN nonce length: %d bytes", #sn_nonce))
    
    -- Create ChaCha20 keystream with counter 1 (skip first 32 bytes)
    -- This matches the Go code: s.SetCounter(1)
    local sn_keystream = ChaCha20.create_keystream(secret, sn_nonce, 1)
    local sn_bytes_decrypted = ""
    
    -- XOR the encrypted SN with keystream (6 bytes for sequence number)
    local encrypted_sn_raw = encrypted_sn:raw()
    for i = 1, 6 do
        local keystream_byte = string.byte(sn_keystream, i)
        local encrypted_byte = string.byte(encrypted_sn_raw, i)
        local decrypted_byte = bit.bxor(encrypted_byte, keystream_byte)
        sn_bytes_decrypted = sn_bytes_decrypted .. string.char(decrypted_byte)
        add_debug(string.format("SN byte %d: encrypted=0x%02x keystream=0x%02x decrypted=0x%02x", 
            i, encrypted_byte, keystream_byte, decrypted_byte))
    end
    
    -- Convert the decrypted bytes to a number (48-bit sequence number)
    local sn_conn = 0
    for i = 1, 6 do  -- Read 6 bytes (48-bits) in little-endian order
        sn_conn = sn_conn + (string.byte(sn_bytes_decrypted, i) << ((i-1) * 8))
    end
    
    add_debug(string.format("Decrypted sequence number: %d", sn_conn))

    -- Now decrypt the main payload (chainedDecrypt logic)
    -- The nonce is constructed from epoch and sequence number.
    -- We try epoch 0, -1, and +1. For simplicity, we'll start with epoch 0.
    local decrypted_payload = nil
    local epochs_to_try = {0, 1, -1} -- Try current, next, and previous epoch

    -- Collect debug info
    local debug_messages = {}
    table.insert(debug_messages, string.format("Attempting decryption with sequence number: %d", sn_conn))
    table.insert(debug_messages, string.format("Secret key length: %d bytes", #secret))

    for _, epoch_try in ipairs(epochs_to_try) do
        if epoch_try < 0 then epoch_try = 0 end -- Clamp negative
        table.insert(debug_messages, string.format("Trying epoch: %d", epoch_try))

        -- Construct the 12-byte nonce (96-bit as required by ChaCha20)
        local nonce_parts = {}
        -- Pad with zeros first
        for i = 1, 12 do
            table.insert(nonce_parts, string.char(0))
        end
        -- Write epoch (little-endian, first 6 bytes)
        for i = 0, 5 do
            nonce_parts[i + 1] = string.char((epoch_try >> (i * 8)) & 0xFF)
        end
        -- Write sequence number (little-endian, next 6 bytes)
        for i = 0, 5 do
            nonce_parts[i + 7] = string.char((sn_conn >> (i * 8)) & 0xFF)
        end
        local nonce_det = table.concat(nonce_parts)

        -- The sender bit logic is inverted for offline/pcap decryption.
        -- is_from_server corresponds to !isSenderOnInit in the Go code.
        -- So, if it's from the server, we are the receiver, and the original sender was the client.
        -- The Go code's `isSenderOnInit` would be `not is_from_server`.
        -- The offline decryption logic in `chainedDecrypt` is:
        -- if isSender { nonce[0] |= 0x80 } else { nonce[0] &= ~0x80 }
        -- Here, `isSender` corresponds to `isSenderOnInit`.
        local first_byte_val = string.byte(nonce_det, 1)
        if not is_from_server then -- Packet from client (original sender)
            first_byte_val = bit.bor(first_byte_val, 0x80)
        else -- Packet from server
            first_byte_val = bit.band(first_byte_val, bit.bnot(0x80))
        end

        -- Reset debug info
        ChaCha20.debug_info = {}

        -- Attempt decryption
        add_debug(string.format("Attempting decryption with nonce[0]: 0x%02x", string.byte(nonce_det, 1)))
        
        local success, payload = pcall(function()
            -- Create keystream for payload decryption
            local keystream = ChaCha20.create_keystream(secret, nonce_det, 0)
            local cipher_raw = ciphertext_with_mac:raw()
            local plain = {}
            
            -- XOR ciphertext with keystream
            for i = 1, #cipher_raw do
                table.insert(plain, string.char(
                    bit.bxor(string.byte(cipher_raw, i), string.byte(keystream, i))
                ))
            end
            
            return table.concat(plain)
        end)

        -- ChaCha20 debug info to UI
        if ChaCha20.debug_info then
            for _, info in ipairs(ChaCha20.debug_info) do
                add_debug(info.state_info)
                add_debug(info.key_info)
            end
        end

        if success and payload then
            decrypted_payload = payload
            break
        end
    end

    if decrypted_payload then
        -- Convert the decrypted payload to ByteArray first
        local ba = ByteArray.new(decrypted_payload)
        -- Create a new TVB from the ByteArray
        return ByteArray.tvb(ba, "Decrypted QOTP Payload")
    end

    return nil
end

-- This function is called when the user changes the protocol preferences.
function qotp_protocol.prefs_changed()
    secrets.loaded = false -- Invalidate the cache
    secrets.data = {}
end

function qotp_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qotp_protocol.name

    local subtree = tree:add(qotp_protocol, buffer(), "QOTP Packet")
    if not subtree then
        return
    end

    local min_data_size = 1 + 8 -- Header + ConnID
    if buffer:len() < min_data_size then
        return
    end

    local offset = 0
    local header_byte = buffer(offset, 1):uint()
    local msg_type = bit.rshift(header_byte, 5)
    local version = bit.band(header_byte, 0x1F)

    -- Create debug subtree
    debug_tree = subtree:add(qotp_protocol, buffer(), "QOTP Debug Information")

    local header_tree = subtree:add(f_qotp_header_byte, buffer(offset, 1))
    header_tree:add(f_qotp_msg_type, buffer(offset, 1), msg_type)
    header_tree:add(f_qotp_version, buffer(offset, 1), version)
    offset = offset + 1

    -- Heuristic: Only "Data" packets are currently supported for decryption.
    -- Other types (Init*, etc.) have different structures.
    if msg_type == 4 then -- Data packet
        -- Read connection ID bytes directly (little-endian)
        local conn_id_bytes = buffer(offset, 8)
        subtree:add(f_qotp_conn_id, conn_id_bytes)
        
        -- Convert to hex string for lookup (little-endian order)
        -- Read bytes in the same order as Go's binary.LittleEndian.Uint64
        local conn_id = string.format(
            "%02x%02x%02x%02x%02x%02x%02x%02x",
            conn_id_bytes(0,1):uint(),
            conn_id_bytes(1,1):uint(),
            conn_id_bytes(2,1):uint(),
            conn_id_bytes(3,1):uint(),
            conn_id_bytes(4,1):uint(),
            conn_id_bytes(5,1):uint(),
            conn_id_bytes(6,1):uint(),
            conn_id_bytes(7,1):uint()
        )
        
        add_debug(string.format("Connection ID (hex): %s", conn_id))

        -- Try to decrypt the payload
        local is_from_server = pinfo.src_port == 8090
        local debug_tree = subtree:add(qotp_protocol, buffer(), "QOTP Debug Information")
        local decrypted_tvb = try_decrypt_payload(buffer, pinfo, conn_id, is_from_server, debug_tree)

        if decrypted_tvb then
            -- Decryption successful, dissect the inner QH protocol
            pinfo.cols.info:append(" (Decrypted)")
            qh_protocol.dissector(decrypted_tvb, pinfo, tree)
        else
            -- Decryption failed or no key, show as encrypted
            subtree:add(f_qotp_payload_enc, buffer(offset + 8))
        end
    else
        -- For non-Data packets, just show the rest as an encrypted payload for now.
        subtree:add(f_qotp_payload_enc, buffer(offset))
    end
end

-- Register protocols in Wireshark
register_qh_proto = qh_protocol
register_qotp_proto = qotp_protocol

-- Register QH to handle UDP traffic on port 8090
local udp_dissector_table = DissectorTable.get("udp.port")
udp_dissector_table:add(8090, qotp_protocol)
