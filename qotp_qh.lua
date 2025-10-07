-- QOTP and QH Protocol Dissector for Wireshark
-- QOTP and QH Protocol Dissector for Wireshark with Decryption

--[[
This script decodes the QOTP (Quite Ok Transport Protocol) and the nested
QH (Quite Ok HTTP) protocol.

To use:
1. Save this file as `qotp_qh.lua`.
2. Place it in your personal Wireshark plugins directory.
   (Find this via Help > About Wireshark > Folders > Personal Lua Plugins).
3. In Wireshark, go to Edit > Preferences > Protocols > QOTP.
4. Set the "Key Log File" path to your `qotp_keylog.log` file.
5. Restart Wireshark and capture traffic on UDP port 8090.
--]]

-- =============================================================================
-- QOTP Dissector
-- =============================================================================

-- Attempt to load the crypto library. This is required for decryption.
-- If this fails, the script will still dissect unencrypted parts.
local has_openssl, openssl = pcall(require, "openssl")
if not has_openssl then
    print("QOTP Dissector: 'openssl' library not found. Decryption will be disabled.")
end

-- Create a new protocol
local qotp_protocol = Proto("QOTP", "Quite Ok Transport Protocol")

-- Preferences for the protocol
local keylog_filename_pref = Pref.string("Key Log File", "", "Path to the QOTP key log file")
qotp_protocol.prefs.keylog_filename = keylog_filename_pref

-- QOTP Header Fields
local f_qotp_version = ProtoField.uint8("qotp.version", "Version", base.DEC)
local f_qotp_type = ProtoField.uint8("qotp.type", "Type", base.DEC, {
    [0] = "InitSnd",
    [1] = "InitRcv",
    [2] = "InitCryptoSnd",
    [3] = "InitCryptoRcv",
    [4] = "Data"
})
local f_qotp_conn_id = ProtoField.uint64("qotp.conn_id", "Connection ID", base.HEX)
local f_qotp_pubkey_id_snd = ProtoField.bytes("qotp.pubkey_id_snd", "Public Key ID (Sender)", base.NONE)
local f_qotp_pubkey_ep_snd = ProtoField.bytes("qotp.pubkey_ep_snd", "Public Key Ephemeral (Sender)", base.NONE)
local f_qotp_pubkey_id_rcv = ProtoField.bytes("qotp.pubkey_id_rcv", "Public Key ID (Receiver)", base.NONE)
local f_qotp_pubkey_ep_rcv = ProtoField.bytes("qotp.pubkey_ep_rcv", "Public Key Ephemeral (Receiver)", base.NONE)
local f_qotp_sn_crypto = ProtoField.bytes("qotp.sn_crypto", "Crypto Sequence Number", base.NONE)
local f_qotp_filler = ProtoField.bytes("qotp.filler", "Filler", base.NONE)
local f_qotp_payload = ProtoField.bytes("qotp.payload", "Encrypted Payload", base.NONE)
local f_qotp_decryption_error = ProtoField.string("qotp.decryption_error", "Decryption Error", base.ASCII)

qotp_protocol.fields = {
    f_qotp_version,
    f_qotp_type,
    f_qotp_conn_id,
    f_qotp_pubkey_id_snd,
    f_qotp_pubkey_ep_snd,
    f_qotp_pubkey_id_rcv,
    f_qotp_pubkey_ep_rcv,
    f_qotp_sn_crypto,
    f_qotp_filler,
    f_qotp_payload,
    f_qotp_decryption_error
}

-- Table to store session secrets
local session_secrets = {}

-- Function to load keys from the key log file
function load_keylog_file()
    local filename = qotp_protocol.prefs.keylog_filename
    if filename == "" then return end

    -- Clear existing keys
    session_secrets = {}
    local key_count = 0

    local file, err = io.open(filename, "r")
    if not file then
        print("QOTP Dissector: Could not open keylog file: " .. filename .. " (" .. tostring(err) .. ")")
        return
    end

    for line in file:lines() do
        -- Expected format: QOTP_SHARED_SECRET <connId_hex> <secret_hex>
        local _, _, conn_id_hex, secret_hex = string.find(line, "^QOTP_SHARED_SECRET%s+([0-9a-fA-F]+)%s+([0-9a-fA-F]+)")
        if conn_id_hex and secret_hex then
            session_secrets[conn_id_hex] = secret_hex
            key_count = key_count + 1
        end
    end
    file:close()
    print("QOTP Dissector: Loaded " .. tostring(key_count) .. " keys from " .. filename)
end

function qotp_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qotp_protocol.name

    local subtree = tree:add(qotp_protocol, buffer(), "QOTP Protocol Data")
    local offset = 0

    -- Decode 1-byte header (Version and Type)
    local header_byte = buffer(offset, 1):uint()
    local version = bit.band(header_byte, 0x1F) -- Lower 5 bits (as per codec.go)
    local msg_type = bit.rshift(header_byte, 5) -- Upper 3 bits
    local type_str = f_qotp_type.valuestring[msg_type] or "Unknown"

    subtree:add(f_qotp_version, buffer(offset, 1), version)
    subtree:add(f_qotp_type, buffer(offset, 1), msg_type)
    offset = offset + 1

    pinfo.cols.info:set(string.format("Type: %s", type_str))

    local conn_id_hex = ""
    -- Most messages have a Connection ID
    if msg_type ~= 0 then -- InitSnd derives ConnID from the key
        local conn_id_tvb = buffer(offset, 8)
        subtree:add(f_qotp_conn_id, conn_id_tvb)
        conn_id_hex = conn_id_tvb:string() -- Use :string() directly on the TvbRange
    end

    -- Dissect based on message type
    if msg_type == 0 then -- InitSnd
        local pubkey_ep_snd_tvb = buffer(offset, 32)
        subtree:add(f_qotp_pubkey_ep_snd, pubkey_ep_snd_tvb)
        offset = offset + 32
        subtree:add(f_qotp_pubkey_id_snd, buffer(offset, 32))
        offset = offset + 32
        -- For InitSnd, the first 8 bytes of the ephemeral key are the ConnID
        conn_id_hex = pubkey_ep_snd_tvb(0, 8):string()
        if buffer:len() > offset then
            subtree:add(f_qotp_filler, buffer(offset))
        end
    elseif msg_type == 1 then -- InitRcv
        offset = offset + 8 -- Skip ConnID already added
        subtree:add(f_qotp_pubkey_id_rcv, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_pubkey_ep_rcv, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_sn_crypto, buffer(offset, 6))
        offset = offset + 6
    elseif msg_type == 2 then -- InitCryptoSnd
        subtree:add(f_qotp_pubkey_ep_snd, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_pubkey_id_snd, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_sn_crypto, buffer(offset, 6))
        offset = offset + 6
    elseif msg_type == 3 then -- InitCryptoRcv
        offset = offset + 8 -- Skip ConnID
        subtree:add(f_qotp_pubkey_ep_rcv, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_sn_crypto, buffer(offset, 6))
        offset = offset + 6
    elseif msg_type == 4 then -- Data
        offset = offset + 8 -- Skip ConnID
        subtree:add(f_qotp_sn_crypto, buffer(offset, 6))
        offset = offset + 6
    end

    -- After dissecting the unencrypted header, try to decrypt the payload
    if buffer:len() > offset then
        local encrypted_payload_tvb = buffer(offset)
        local secret_hex = session_secrets[conn_id_hex]

        if secret_hex and has_openssl then
            -- We have a key, let's try to decrypt
            local secret_bytes = ByteArray.new(secret_hex, true)
            local aad_tvb = buffer(0, offset) -- The header is the AAD

            -- Extract the encrypted sequence number and the main payload
            local sn_encrypted_tvb = encrypted_payload_tvb(0, 6)
            local first_layer_ciphertext_tvb = encrypted_payload_tvb(6)

            -- Step 1: Decrypt the sequence number (XChaCha20)
            -- The nonce for this is the first 24 bytes of the main payload.
            if first_layer_ciphertext_tvb:len() < 24 then
                subtree:add(f_qotp_decryption_error, "Payload too short for SN decryption nonce")
                return
            end
            local sn_nonce_tvb = first_layer_ciphertext_tvb(0, 24)
            local sn_decrypted_bytes, err = pcall(openssl.aead.decrypt, 'xchacha20-poly1305', secret_bytes:raw(), sn_nonce_tvb:raw(), "", sn_encrypted_tvb:raw())
            if not sn_decrypted_bytes then
                subtree:add(f_qotp_decryption_error, "Failed to decrypt sequence number: " .. tostring(err))
                return
            end

            -- Step 2: Decrypt the main payload (ChaCha20)
            -- The nonce is derived from the decrypted sequence number.
            local payload_nonce = ByteArray.new()
            payload_nonce:append(ByteArray.new(sn_decrypted_bytes))
            payload_nonce:set_len(12) -- Pad with zeros to 12 bytes

            local plaintext_bytes, err = pcall(openssl.aead.decrypt, 'chacha20-poly1305', secret_bytes:raw(), payload_nonce:raw(), aad_tvb:raw(), first_layer_ciphertext_tvb:raw())

            if plaintext_bytes then
                -- Decryption successful! Call the QH sub-dissector.
                local decrypted_tvb = ByteArray.new(plaintext_bytes):tvb("Decrypted QH Payload")
                Dissector.get("qh"):call(decrypted_tvb, pinfo, tree)
            else
                -- Decryption failed
                subtree:add(f_qotp_payload, encrypted_payload_tvb)
                subtree:add(f_qotp_decryption_error, "Payload decryption failed: " .. tostring(err))
            end

            -- If decryption were successful, you would call the QH dissector:
            -- Dissector.get("qh"):call(decrypted_tvb, pinfo, tree)

        else
            -- No key found, just show the encrypted payload
            subtree:add(f_qotp_payload, encrypted_payload_tvb)
            if not has_openssl then
                subtree:add_expert_info(PI_COMMENT, PI_NOTE, "Decryption skipped: 'openssl' Lua library not found.")
            end
        end
    end
end

-- =============================================================================
-- QH Dissector
-- =============================================================================
local qh_protocol = Proto("QH", "Quite Ok HTTP Protocol")

-- QH Fields
local f_qh_payload_version = ProtoField.uint8("qh.payload.version", "Version", base.DEC)
local f_qh_payload_type = ProtoField.uint8("qh.payload.type", "Type", base.DEC, { [0] = "DATA", [1] = "PING", [2] = "CLOSE" })
local f_qh_payload_offset_size = ProtoField.string("qh.payload.offset_size", "Offset Size", base.ASCII)
local f_qh_payload_has_ack = ProtoField.bool("qh.payload.has_ack", "Has ACK", base.NONE)

local f_qh_ack_stream_id = ProtoField.uint32("qh.ack.stream_id", "ACK Stream ID", base.DEC)
local f_qh_ack_offset = ProtoField.uint64("qh.ack.offset", "ACK Offset", base.DEC)
local f_qh_ack_len = ProtoField.uint16("qh.ack.len", "ACK Length", base.DEC)
local f_qh_ack_rcv_wnd = ProtoField.uint64("qh.ack.rcv_wnd", "ACK Receive Window", base.DEC)

local f_qh_data_stream_id = ProtoField.uint32("qh.data.stream_id", "Data Stream ID", base.DEC)
local f_qh_data_offset = ProtoField.uint64("qh.data.offset", "Data Offset", base.DEC)
local f_qh_data_payload = ProtoField.bytes("qh.data.payload", "Data Payload", base.NONE)

qh_protocol.fields = {
    f_qh_payload_version, f_qh_payload_type, f_qh_payload_offset_size, f_qh_payload_has_ack,
    f_qh_ack_stream_id, f_qh_ack_offset, f_qh_ack_len, f_qh_ack_rcv_wnd,
    f_qh_data_stream_id, f_qh_data_offset, f_qh_data_payload
}

-- This is the dissector for the PLAINTEXT QH data after decryption.
function qh_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol:append("/QH")
    local subtree = tree:add(qh_protocol, buffer, "QH Payload (Decrypted)")
    local offset = 0

    if buffer:len() < 1 then
        subtree:add_expert_info(PI_MALFORMED, PI_ERROR, "Payload is empty")
        return
    end

    -- Decode payload header byte
    local header_byte = buffer(offset, 1):uint()
    local version = bit.band(header_byte, 0x0F) -- bits 0-3
    local type_flag = bit.band(bit.rshift(header_byte, 4), 0x03) -- bits 4-5
    local is_extend = bit.band(bit.rshift(header_byte, 6), 0x01) == 1 -- bit 6
    local has_ack = bit.band(bit.rshift(header_byte, 7), 0x01) == 1 -- bit 7

    local header_tree = subtree:add(buffer(offset, 1), string.format("QH Payload Header: Type=%s, %s, %s",
        f_qh_payload_type.valuestring[type_flag],
        is_extend and "48-bit" or "24-bit",
        has_ack and "ACK" or "No ACK"))
    header_tree:add(f_qh_payload_version, buffer(offset, 1), version)
    header_tree:add(f_qh_payload_type, buffer(offset, 1), type_flag)
    header_tree:add(f_qh_payload_offset_size, buffer(offset, 1), is_extend and "48-bit" or "24-bit")
    header_tree:add(f_qh_payload_has_ack, buffer(offset, 1), has_ack)
    offset = offset + 1

    local offset_len = is_extend and 6 or 3

    -- Decode ACK section
    if has_ack then
        local ack_tree = subtree:add(buffer(offset), "ACK Section")
        ack_tree:add(f_qh_ack_stream_id, buffer(offset, 4))
        offset = offset + 4
        ack_tree:add(f_qh_ack_offset, buffer(offset, offset_len))
        offset = offset + offset_len
        ack_tree:add(f_qh_ack_len, buffer(offset, 2))
        offset = offset + 2
        -- The rcv_wnd is not fully implemented in the Go code, so we'll just show the byte
        ack_tree:add(f_qh_ack_rcv_wnd, buffer(offset, 1))
        offset = offset + 1
    end

    -- Decode Data section
    local data_tree = subtree:add(buffer(offset), "Data Section")
    data_tree:add(f_qh_data_stream_id, buffer(offset, 4))
    offset = offset + 4
    data_tree:add(f_qh_data_offset, buffer(offset, offset_len))
    offset = offset + offset_len

    if buffer:len() > offset then
        local data_payload_tvb = buffer(offset)
        data_tree:add(f_qh_data_payload, data_payload_tvb)
        -- Here you would call the final application layer dissector (e.g., HTTP-like)
        -- Dissector.get("http"):call(data_payload_tvb, pinfo, tree)
    end
end

-- Reload keys when preferences change
function qotp_protocol.prefs_changed()
    load_keylog_file()
end

-- Load keys on init
function qotp_protocol.init()
    load_keylog_file()
end

-- Register QOTP to handle UDP traffic on port 8090
local udp_dissector_table = DissectorTable.get("udp.port")
udp_dissector_table:add(8090, qotp_protocol)

-- This is a placeholder. To actually decrypt and dissect, you would
-- register the QH dissector to be called from the QOTP dissector.
-- For example, by creating a new dissector table:
-- local qotp_payload_table = DissectorTable.new("qotp.payload", "QOTP Payload")
-- qotp_payload_table:add(4, qh_protocol) -- Assuming '4' is the type for Data
