--[[
    QOTP Wireshark Dissector (Lua Plugin)
    
    This plugin dissects and decrypts QOTP (Quick UDP Transport Protocol) traffic on port 8090.
    It uses the qotp_decrypt.dll (C wrapper) which calls Go crypto functions.
    
    Setup:
    1. Build qotp_crypto.dll from crypto.go
    2. Build qotp_decrypt.dll from qotp_decrypt.c
    3. Copy both DLLs to C:\Program Files\Wireshark\
    4. Copy this file to Wireshark plugins directory
    5. Restart Wireshark
--]]

-- Load the QOTP decrypt library
local qotp_decrypt = require("qotp_decrypt")

-- Display version info at startup
local version_info = qotp_decrypt.get_version()
print("========================================")
print("QOTP Wireshark Dissector")
print("========================================")
print("Version Info:")
print(version_info)
print("========================================")

-- Test that the library loaded
qotp_decrypt.test()

-- Create protocol first (before accessing any prefs)
local qotp_proto = Proto("QOTP", "Quick UDP Transport Protocol")

-- Define fields for QOTP protocol (must be defined before using them)
local f_msg_type = ProtoField.string("qotp.msg_type", "Message Type")
local f_version = ProtoField.uint8("qotp.version", "Version", base.DEC)
local f_conn_id = ProtoField.uint64("qotp.conn_id", "Connection ID", base.HEX)
local f_encrypted = ProtoField.bytes("qotp.encrypted", "Encrypted Data")
local f_decrypted = ProtoField.bytes("qotp.decrypted", "Decrypted Data")
local f_header = ProtoField.bytes("qotp.header", "Header")

-- Define fields for QH protocol (inner protocol)
local f_qh_version = ProtoField.uint8("qh.version", "QH Version", base.DEC)
local f_qh_type = ProtoField.string("qh.type", "Type")
local f_qh_method = ProtoField.string("qh.method", "Operation")
local f_qh_status = ProtoField.uint16("qh.status", "Status Code", base.DEC)
local f_qh_host = ProtoField.string("qh.host", "Host")
local f_qh_path = ProtoField.string("qh.path", "Path")
local f_qh_body = ProtoField.bytes("qh.body", "Body")

-- Register all fields with the protocol
qotp_proto.fields = {
    f_msg_type,
    f_version,
    f_conn_id,
    f_encrypted,
    f_decrypted,
    f_header,
    f_qh_version,
    f_qh_type,
    f_qh_method,
    f_qh_status,
    f_qh_host,
    f_qh_path,
    f_qh_body
}

-- Add preferences for keylog file
qotp_proto.prefs.keylog_file = Pref.string("Keylog file", "", "Path to QOTP keylog file")

-- Storage for loaded keys
local shared_secrets = {}
local keys_loaded = false
local last_keylog_path = ""
local last_key_count = 0

-- Helper function to convert UInt64 to Lua number for key lookup
local function uint64_to_number(uint64_obj)
    -- Get the raw bytes and reconstruct the number
    -- UInt64 in Lua can be converted via :tonumber() method if available
    if type(uint64_obj) == "userdata" then
        -- Try the tonumber() method first
        local success, result = pcall(function() return uint64_obj:tonumber() end)
        if success then
            return result
        end
        -- Fallback: convert to string and parse
        local str = tostring(uint64_obj)
        return tonumber(str)
    end
    return uint64_obj
end

-- Helper function to convert buffer bytes to hex string for connection ID
-- Reads the 8-byte connection ID directly from buffer and converts to hex string
local function buffer_to_hex_string(buffer, offset, length)
    local hex = ""
    for i = offset + length - 1, offset, -1 do  -- Read in reverse (little-endian to big-endian)
        hex = hex .. string.format("%02x", buffer(i, 1):uint())
    end
    return hex
end

-- Helper function to format large numbers as hex (for display only)
local function format_hex(num)
    if num == nil then return "nil" end
    -- For large numbers that can't use %x, use the string representation
    local success, result = pcall(function() return string.format("0x%x", num) end)
    if success then
        return result
    else
        -- Fallback: if the number is too large, return as decimal with 0x prefix hint
        return "0x" .. tostring(num)
    end
end

-- Function to load keys from keylog file (supports reloading)
local function load_keylog_file(filepath, force_reload)
    if filepath == "" then
        if TextWindow then
            local tw = TextWindow.new("QOTP Keylog Error")
            tw:set("No keylog file configured!\n\nGo to:\nEdit -> Preferences -> Protocols -> QOTP\nand set the 'Keylog file' path.")
        end
        return false
    end
    
    -- Skip if already loaded and not forcing reload
    if keys_loaded and not force_reload then
        return false
    end
    
    local file = io.open(filepath, "r")
    if not file then
        print("QOTP: Could not open keylog file: " .. filepath)
        if TextWindow then
            local tw = TextWindow.new("QOTP Keylog Error")
            tw:set(string.format("Could not open keylog file:\n\n%s\n\nPlease check:\n- File exists\n- Path is correct\n- File is readable", filepath))
        end
        return
    end
    
    local count = 0
    local errors = 0
    local line_num = 0
    print("========================================")
    print("Loading QOTP Keylog File: " .. filepath)
    print("========================================")
    
    for line in file:lines() do
        line_num = line_num + 1
        
        -- Skip comments and empty lines
        if line:sub(1,1) == "#" or not line:match("%S") then
            -- Skip silently
        else
            print(string.format("Line %d: %s", line_num, line))
            
            local conn_id_str, secret_hex
            
            -- Try format 1: QOTP_SHARED_SECRET CONNECTION_ID SHARED_SECRET
            local prefix, id1, sec1 = line:match("^(QOTP_SHARED_SECRET)%s+(%S+)%s+(%S+)$")
            if prefix then
                print(string.format("  Detected format: QOTP_SHARED_SECRET (ID=%s, Secret=%s...)", id1, sec1:sub(1, 8)))
                conn_id_str = id1
                secret_hex = sec1
            else
                -- Try format 2: CONNECTION_ID SHARED_SECRET
                conn_id_str, secret_hex = line:match("^(%S+)%s+(%S+)$")
                if conn_id_str then
                    print(string.format("  Detected format: Simple (ID=%s, Secret=%s...)", conn_id_str, secret_hex:sub(1, 8)))
                else
                    print(string.format("  [ERROR] Could not parse line - unrecognized format"))
                    errors = errors + 1
                end
            end
            
            if conn_id_str and secret_hex then
                -- Validate secret length
                if #secret_hex ~= 64 then
                    print(string.format("  [ERROR] Connection %s: Invalid secret length (%d chars, expected 64)", conn_id_str, #secret_hex))
                    errors = errors + 1
                else
                    -- Normalize connection ID to lowercase hex without 0x prefix
                    local normalized_id = conn_id_str:lower():gsub("^0x", "")
                    
                    -- Validate it's valid hex
                    if normalized_id:match("^[0-9a-f]+$") then
                        -- Store using normalized hex string as key (for Lua lookup)
                        shared_secrets[normalized_id] = secret_hex
                        
                        -- Also load into Go DLL using hex string (avoid integer overflow)
                        local success = qotp_decrypt.set_key(normalized_id, secret_hex)
                        if success then
                            print(string.format("  [OK] Loaded key for connection 0x%s", normalized_id))
                            count = count + 1
                        else
                            print(string.format("  [ERROR] Failed to load key for connection 0x%s (invalid hex?)", normalized_id))
                            errors = errors + 1
                        end
                    else
                        print(string.format("  [ERROR] Invalid connection ID: %s", conn_id_str))
                        errors = errors + 1
                    end
                end
            end
        end
    end
    file:close()
    
    keys_loaded = true
    last_keylog_path = filepath
    last_key_count = count
    
    print("========================================")
    print(string.format("Keylog Summary: %d keys loaded, %d errors", count, errors))
    
    -- Display all loaded keys
    local loaded_keys = qotp_decrypt.get_loaded_keys()
    if #loaded_keys > 0 then
        print("Loaded Connection IDs:")
        for i, conn_id in ipairs(loaded_keys) do
            print(string.format("  %d. %s (%s)", i, format_hex(conn_id), tostring(conn_id)))
        end
    end
    print("========================================")
    
    -- Show message box with results (since Windows Wireshark has no visible console)
    return true
end

-- Function to check if keylog needs reloading (new keys added)
local function check_and_reload_keylog(filepath)
    if filepath == "" or filepath ~= last_keylog_path then
        return false
    end
    
    -- Count current non-comment lines in file to see if it changed
    local file = io.open(filepath, "r")
    if not file then
        return false
    end
    
    local line_count = 0
    for line in file:lines() do
        -- Count only non-comment, non-empty lines (actual key entries)
        if line:sub(1,1) ~= "#" and line:match("%S") then
            line_count = line_count + 1
        end
    end
    file:close()
    
    -- If file has more key entries than we've seen, reload
    if line_count > last_key_count then
        print("========================================")
        print(string.format("QOTP: Keylog file has new entries (now %d keys, was %d), reloading...", 
            line_count, last_key_count))
        print("========================================")
        
        -- Reset and reload
        keys_loaded = false
        shared_secrets = {}
        return load_keylog_file(filepath, true)
    end
    
    return false
end

-- QH Method names
local qh_methods = {
    [0] = "GET",
    [1] = "POST",
    [2] = "PUT",
    [3] = "PATCH",
    [4] = "DELETE",
    [5] = "HEAD"
}

-- Compact to HTTP status code mapping (from status.go)
local compact_to_status = {
    [0] = 200,  -- OK
    [1] = 404,  -- Not Found
    [2] = 500,  -- Internal Server Error
    [3] = 302,  -- Found (redirect)
    [4] = 400,  -- Bad Request
    [5] = 403,  -- Forbidden
    [6] = 401,  -- Unauthorized
    [7] = 301,  -- Moved Permanently
    [8] = 304,  -- Not Modified
    [9] = 503,  -- Service Unavailable
    [10] = 201, -- Created
    [11] = 202, -- Accepted
    [12] = 204, -- No Content
    [13] = 206, -- Partial Content
    [14] = 307, -- Temporary Redirect
    [15] = 308, -- Permanent Redirect
    [16] = 409, -- Conflict
    [17] = 410, -- Gone
    [18] = 412, -- Precondition Failed
    [19] = 413, -- Payload Too Large
    [20] = 414, -- URI Too Long
    [21] = 415, -- Unsupported Media Type
    [22] = 422, -- Unprocessable Entity
    [23] = 429, -- Too Many Requests
    [24] = 502, -- Bad Gateway
    [25] = 504, -- Gateway Timeout
    [26] = 505, -- QH Version Not Supported
    [27] = 100, -- Continue
    [28] = 101, -- Switching Protocols
    [29] = 102, -- Processing
    [30] = 103, -- Early Hints
    [31] = 205, -- Reset Content
    [32] = 207, -- Multi-Status
    [33] = 208, -- Already Reported
    [34] = 226, -- IM Used
    [35] = 300, -- Multiple Choices
    [36] = 303, -- See Other
    [37] = 305, -- Use Proxy
    [38] = 402, -- Payment Required
    [39] = 405, -- Method Not Allowed
    [40] = 406, -- Not Acceptable
    [41] = 407, -- Proxy Authentication Required
    [42] = 408, -- Request Timeout
    [43] = 411, -- Length Required
    [44] = 416, -- Range Not Satisfiable
    [45] = 417  -- Expectation Failed
}

-- Function to parse QH protocol from decrypted data string
local function parse_qh_protocol(decrypted_data, tree, pinfo)
    -- Skip very small packets - likely handshake/control data, not QH protocol
    if #decrypted_data < 19 then
        if pinfo.visited == false then
            print(string.format("QH: Skipping QH parsing - data too small (%d bytes)", #decrypted_data))
        end
        return
    end
    
    local qh_tree = tree:add(qotp_proto, "QH Protocol")
    local offset = 1  -- Lua strings are 1-indexed
    
    -- Skip 1 byte general offset (reason unknown - observed in captures)
    offset = offset + 1
    
    -- First byte format:
    -- Request:  Version (2 bits, bits 7-6) | Method (3 bits, bits 5-3) | Reserved (3 bits, bits 2-0)
    -- Response: Version (2 bits, bits 7-6) | Compact Status (6 bits, bits 5-0)
    local first_byte = decrypted_data:byte(offset)
    local qh_version = bit.rshift(first_byte, 6)
    
    qh_tree:add(f_qh_version, qh_version):set_generated()
    
    -- Determine if it's a request or response based on port direction
    -- Request: destination port 8090 (client -> server)
    -- Response: source port 8090 (server -> client)
    local is_request = (pinfo.dst_port == 8090)
    
    if pinfo.visited == false then
        print(string.format("QH: Port detection - src_port=%d, dst_port=%d -> %s", 
            pinfo.src_port, pinfo.dst_port, is_request and "REQUEST" or "RESPONSE"))
    end
    
    local method_bits = bit.band(bit.rshift(first_byte, 3), 0x07)  -- bits 5-3
    
    offset = offset + 1
    
    if is_request then
        -- Parse Request
        qh_tree:add(f_qh_type, "Request"):set_generated()
        local method_name = qh_methods[method_bits] or string.format("Unknown(%d)", method_bits)
        qh_tree:add(f_qh_method, method_name):set_generated()
        pinfo.cols.info:append(string.format(" [%s", method_name))
        
        -- Skip 8 bytes (unknown header data) before host
        -- TODO: Parse these 8 bytes to understand what they contain
        offset = offset + 7
        
        if pinfo.visited == false then
            print(string.format("QH: Skipped 8 header bytes, now at offset %d", offset))
        end
        
        -- Debug: Show all bytes from current offset
        if pinfo.visited == false then
            local debug_len = math.min(50, #decrypted_data - offset + 1)
            local debug_bytes = ""
            local debug_ascii = ""
            for i = 0, debug_len - 1 do
                local b = decrypted_data:byte(offset + i)
                debug_bytes = debug_bytes .. string.format("%02x ", b)
                if b >= 32 and b <= 126 then
                    debug_ascii = debug_ascii .. string.char(b)
                else
                    debug_ascii = debug_ascii .. "."
                end
            end
            print(string.format("QH: Next %d bytes from offset %d:", debug_len, offset))
            print(string.format("  Hex:   %s", debug_bytes))
            print(string.format("  ASCII: %s", debug_ascii))
        end
        
        -- Parse host (null-terminated string)
        if offset <= #decrypted_data then
            local host_end = offset
            while host_end <= #decrypted_data and decrypted_data:byte(host_end) ~= 0 do
                host_end = host_end + 1
            end
            
            if host_end > offset then
                local host = decrypted_data:sub(offset, host_end - 1)
                
                -- Debug output
                if pinfo.visited == false then
                    print(string.format("QH: Host parsing - offset=%d, host_end=%d, length=%d", offset, host_end, host_end - offset))
                    print(string.format("QH: Host string: '%s'", host))
                end
                
                qh_tree:add(f_qh_host, host):set_generated()
                offset = host_end + 1  -- Skip null terminator
            else
                if pinfo.visited == false then
                    print(string.format("QH: No host found at offset %d (immediate null byte)", offset))
                end
            end
        end
        
        -- Parse path (null-terminated string)
        if offset <= #decrypted_data then
            local path_end = offset
            while path_end <= #decrypted_data and decrypted_data:byte(path_end) ~= 0 do
                path_end = path_end + 1
            end
            
            if path_end > offset then
                local path = decrypted_data:sub(offset, path_end - 1)
                qh_tree:add(f_qh_path, path):set_generated()
                pinfo.cols.info:append(string.format(" %s]", path))
                offset = path_end + 1  -- Skip null terminator
            else
                pinfo.cols.info:append("]")
            end
        end
        
    else
        -- Parse Response
        qh_tree:add(f_qh_type, "Response"):set_generated()
        local compact_status = bit.band(first_byte, 0x3F)  -- Lower 6 bits
        local http_status = compact_to_status[compact_status] or 500
        
        if pinfo.visited == false then
            print(string.format("QH: Response - first_byte=0x%02x, compact_status=%d, http_status=%d", 
                first_byte, compact_status, http_status))
        end
        
        qh_tree:add(f_qh_status, http_status):set_generated()
        pinfo.cols.info:append(string.format(" [Status: %d]", http_status))
    end
    
    -- Find body (after ETX marker 0x03)
    while offset <= #decrypted_data do
        if decrypted_data:byte(offset) == 0x03 then
            offset = offset + 1
            break
        end
        offset = offset + 1
    end
    
    if offset <= #decrypted_data then
        local body = decrypted_data:sub(offset)
        if #body > 0 then
            qh_tree:add(f_qh_body, string.format("%d bytes", #body)):set_generated()
        end
    end
end

-- Dissector function
function qotp_proto.dissector(buffer, pinfo, tree)
    local length = buffer:len()
    if length == 0 then return end
    
    local keylog_path = qotp_proto.prefs.keylog_file
    
    -- Load keylog file on first dissection
    if not keys_loaded then
        if keylog_path and keylog_path ~= "" then
            load_keylog_file(keylog_path, false)
        end
    else
        -- Check if keylog file has been updated (new keys added)
        -- Only check on first pass to avoid performance issues
        if pinfo.visited == false and keylog_path and keylog_path ~= "" then
            check_and_reload_keylog(keylog_path)
        end
    end
    
    pinfo.cols.protocol = qotp_proto.name
    
    -- Create subtree
    local subtree = tree:add(qotp_proto, buffer(), "QOTP Protocol Data")
    
    -- Parse header byte
    local header_byte = buffer(0, 1):uint()
    local msg_type = bit.rshift(header_byte, 5)
    local version = bit.band(header_byte, 0x1F)
    
    -- Get message type string
    local msg_type_names = {
        [0] = "InitSnd",
        [1] = "InitRcv",
        [2] = "InitCryptoSnd", 
        [3] = "InitCryptoRcv",
        [4] = "Data"
    }
    
    local msg_type_str = msg_type_names[msg_type] or "Unknown"
    
    subtree:add(f_msg_type, buffer(0, 1), msg_type_str)
    subtree:add(f_version, buffer(0, 1), version)
    
    -- Parse connection ID (if present - not in InitSnd)
    if msg_type ~= 0 and length >= 9 then
        subtree:add_le(f_conn_id, buffer(1, 8))
        
        -- Get connection ID as hex string for display
        local conn_id_hex = buffer_to_hex_string(buffer, 1, 8)
        pinfo.cols.info = string.format("%s (ConnID: 0x%s)", msg_type_str, conn_id_hex)
    else
        pinfo.cols.info = msg_type_str
    end
    
    -- Add encrypted data field and attempt decryption for Data packets
    if msg_type == 4 and length > 9 then -- Data packet
        local conn_id = buffer(1, 8):le_uint64()
        local encrypted_portion = buffer(9, length - 9):bytes()
        
        subtree:add(f_encrypted, buffer(9, length - 9))
        
        -- Get connection ID as hex string for key lookup (avoids floating point issues)
        local conn_id_hex = buffer_to_hex_string(buffer, 1, 8)
        
        -- Debug output on first pass
        if pinfo.visited == false then
            print(string.format("QOTP: Data packet - ConnID hex=%s", conn_id_hex))
            print(string.format("QOTP: Available keys: %d", table.getn(shared_secrets)))
            for k, v in pairs(shared_secrets) do
                print(string.format("  Key: %s", k))
            end
        end

        if shared_secrets[conn_id_hex] then
            -- Try to decrypt with epoch 0 first, then 1, etc.
            local decrypted = nil
            local used_epoch = 0
            local used_sender = false
            local decrypt_error = nil
            
            -- Debug on first pass
            if pinfo.visited == false then
                print(string.format("QOTP: Attempting decryption for conn 0x%s, encrypted len=%d", 
                    conn_id_hex, length - 9))
            end
            
            -- Try both sender flags if needed
            for _, is_sender in ipairs({false, true}) do
                for epoch = 0, 2 do
                    if pinfo.visited == false then
                        print(string.format("  Trying: is_sender=%s, epoch=%d", tostring(is_sender), epoch))
                    end
                    
                    decrypted, decrypt_error = qotp_decrypt.decrypt_data(
                        encrypted_portion:raw(),
                        conn_id_hex,  -- Always pass hex string
                        is_sender,
                        epoch
                    )
                    
                    if decrypted then
                        used_epoch = epoch
                        used_sender = is_sender
                        if pinfo.visited == false then
                            print(string.format("  SUCCESS! Decrypted %d bytes with is_sender=%s, epoch=%d", 
                                #decrypted, tostring(is_sender), epoch))
                        end
                        break
                    else
                        if pinfo.visited == false and decrypt_error then
                            print(string.format("    Failed: %s", decrypt_error))
                        end
                    end
                end
                if decrypted then break end
            end
            
            if decrypted then
                -- Add decrypted data to the tree
                local decrypted_tvb = ByteArray.new(decrypted, true):tvb("Decrypted Data")
                local decrypted_tree = subtree:add(f_decrypted, decrypted_tvb():range())
                decrypted_tree:append_text(string.format(" (Epoch: %d, Sender: %s, Length: %d)", 
                    used_epoch, tostring(used_sender), #decrypted))
                
                -- Debug: print first bytes of decrypted data
                if pinfo.visited == false and #decrypted > 0 then
                    local first_byte = decrypted:byte(1)
                    print(string.format("QOTP: First byte of decrypted data: 0x%02x (binary: %s)", 
                        first_byte, string.format("%08d", tonumber(string.format("%d", first_byte), 2))))
                    print(string.format("  Version bits (7-6): %d", bit.rshift(first_byte, 6)))
                    print(string.format("  Method bits (5-3): %d", bit.band(bit.rshift(first_byte, 3), 0x07)))
                end
                
                -- Parse QH protocol from decrypted data string
                parse_qh_protocol(decrypted, subtree, pinfo)
                
                -- Show decrypted data as hex and ASCII
                local hex_str = ""
                local ascii_str = ""
                for i = 1, math.min(#decrypted, 64) do
                    local byte = decrypted:byte(i)
                    hex_str = hex_str .. string.format("%02x ", byte)
                    if byte >= 32 and byte <= 126 then
                        ascii_str = ascii_str .. string.char(byte)
                    else
                        ascii_str = ascii_str .. "."
                    end
                end
                
                if pinfo.visited == false then
                    print(string.format("QOTP: Decrypted payload (first %d bytes):", math.min(#decrypted, 64)))
                    print(string.format("  Hex: %s", hex_str))
                    print(string.format("  ASCII: %s", ascii_str))
                end
                
                pinfo.cols.info:append(" [Decrypted]")
            else
                local error_msg = decrypt_error or "Decryption failed (wrong key/epoch?)"
                subtree:add_expert_info(PI_DECRYPTION, PI_WARN, error_msg)
                -- Debug output
                if pinfo.visited == false then
                    print(string.format("QOTP: Failed to decrypt packet for conn 0x%s: %s", conn_id_hex, error_msg))
                end
            end
        else
            subtree:add_expert_info(PI_DECRYPTION, PI_NOTE, string.format("No key for connection 0x%s", conn_id_hex))
            -- Debug output on first pass
            if pinfo.visited == false then
                print(string.format("QOTP: No key available for connection 0x%s", conn_id_hex))
            end
        end
    elseif length > 9 then
        subtree:add(f_encrypted, buffer(9, length - 9))
    end
    
    local info_str = string.format("QOTP %s, Length: %d", msg_type_str, length)
    subtree:append_text(", " .. info_str)
end

-- Add init function that gets called when preferences are loaded
function qotp_proto.init()
    keys_loaded = false  -- Reset on init
    shared_secrets = {}  -- Clear keys
    
    print("========================================")
    print("QOTP Init: Loading preferences...")
    local keylog_path = qotp_proto.prefs.keylog_file
    print("Configured Keylog File: '" .. (keylog_path or "(not set)") .. "'")
    
    if keylog_path and keylog_path ~= "" then
        print("Attempting to load keylog...")
        load_keylog_file(keylog_path)
    else
        print("No keylog file configured.")
        print("To configure: Edit -> Preferences -> Protocols -> QOTP")
    end
    print("========================================")
end

-- Register the protocol on UDP port 8090
local udp_port = DissectorTable.get("udp.port")
udp_port:add(8090, qotp_proto)

print("QOTP dissector registered on UDP port 8090")
