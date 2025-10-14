-- QOTP and QH Protocol Dissector for Wireshark

--[[
This script decodes the QOTP (Quite Ok Transport Protocol) and the nested
QH (Quite Ok HTTP) protocol. This version assumes the traffic has been
pre-decrypted using an external tool.

To use:
1. Save this file as `qotp_qh.lua`.
2. Place it in your personal Wireshark plugins directory.
   (Find this via Help > About Wireshark > Folders > Personal Lua Plugins).
3. Restart Wireshark.
4. Use the `qotp-decrypt` tool to create a decrypted pcap file.
5. Open the decrypted pcap file in Wireshark.
--]]

-- =============================================================================
-- QOTP Dissector
-- =============================================================================

-- Create a new protocol
local qotp_protocol = Proto("QOTP", "Quite Ok Transport Protocol")

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

    -- Most messages have a Connection ID
    if msg_type ~= 0 then -- InitSnd derives ConnID from the key
        local conn_id_tvb = buffer(offset, 8)
        subtree:add(f_qotp_conn_id, conn_id_tvb)
    end

    -- Dissect based on message type
    if msg_type == 0 then -- InitSnd
        local pubkey_ep_snd_tvb = buffer(offset, 32)
        subtree:add(f_qotp_pubkey_ep_snd, pubkey_ep_snd_tvb)
        offset = offset + 32
        subtree:add(f_qotp_pubkey_id_snd, buffer(offset, 32))
        offset = offset + 32
        -- For InitSnd, the first 8 bytes of the ephemeral key are the ConnID
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
        -- If the packet was decrypted, the remaining buffer is the plaintext QH payload.
        local payload_tvbr = buffer(offset)
        if msg_type == 4 then -- Data
            -- Create a new Tvb from the TvbRange before calling the sub-dissector
            Dissector.get("qh"):call(payload_tvbr:tvb("QH Payload"), pinfo, tree)
        else
            subtree:add(f_qotp_payload, payload_tvbr)
        end
    end
end

-- =============================================================================
-- QH Dissector
-- =============================================================================
local qh_protocol = Proto("QH", "Quite Ok HTTP Protocol")

-- QH Fields
local f_qh_version = ProtoField.uint8("qh.version", "Version", base.DEC)
-- Request fields
local f_qh_method = ProtoField.uint8("qh.method", "Method", base.DEC, { [0] = "GET", [1] = "POST" })
local f_qh_host = ProtoField.string("qh.host", "Host", base.ASCII)
local f_qh_path = ProtoField.string("qh.path", "Path", base.ASCII)
-- Response fields
local f_qh_status_compact = ProtoField.uint8("qh.status_code_compact", "Status Code (Compact)", base.DEC)
local f_qh_status = ProtoField.uint16("qh.status_code", "Status Code", base.DEC)
-- Common fields
local f_qh_headers = ProtoField.string("qh.headers", "Headers", base.ASCII)
local f_qh_body = ProtoField.bytes("qh.body", "Body", base.NONE)

qh_protocol.fields = {
    f_qh_version, f_qh_method, f_qh_host, f_qh_path, f_qh_status_compact, f_qh_status, f_qh_headers, f_qh_body
}

-- This is the dissector for the PLAINTEXT QH data after decryption.
function qh_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qh_protocol.name
    local subtree = tree:add(qh_protocol, buffer, "QH Application Data")
    local offset = 0

    if buffer:len() < 1 then
        subtree:add_expert_info(PI_MALFORMED, PI_ERROR, "QH Payload is empty")
        return
    end

    -- Heuristic to determine if it's a request or response.
    -- If the source port is the well-known port, it's a response.
    local is_response = pinfo.src_port == 8090

    -- Decode the first byte
    local header_byte = buffer(offset, 1):uint()
    local version = bit.band(bit.rshift(header_byte, 6), 0x03) -- Upper 2 bits
    subtree:add(f_qh_version, buffer(offset, 1), version)
    offset = offset + 1

    -- Find the ETX character (\x03) which separates headers from the body
    local etx_offset = string.find(buffer:raw(), string.char(3), 1, true)
    if not etx_offset then
        subtree:add_expert_info(PI_MALFORMED, PI_ERROR, "No ETX body separator found")
        return
    end

    local header_data_tvbr = buffer(offset, etx_offset - offset)
    local header_parts = {}
    for part in header_data_tvbr:string():gmatch("([^\0]*)") do
        table.insert(header_parts, part)
    end

    if is_response then
        -- Dissect as a Response
        local compact_status = bit.band(header_byte, 0x3F) -- Lower 6 bits (0-63)
        subtree:add(f_qh_status_compact, buffer(0, 1), compact_status)
        -- We can't decode to the full HTTP status without a large mapping table in Lua.
        -- The compact code is sufficient for debugging.
        pinfo.cols.info:set(string.format("Response, Status (Compact): %d", compact_status))

        local headers_tree = subtree:add(header_data_tvbr, "Headers")
        for i, header_val in ipairs(header_parts) do
            headers_tree:add(f_qh_headers, header_val)
        end
    else
        -- Dissect as a Request
        local method = bit.band(bit.rshift(header_byte, 3), 0x07) -- Middle 3 bits
        local method_str = f_qh_method.valuestring[method] or "Unknown"
        subtree:add(f_qh_method, buffer(0, 1), method)

        if #header_parts < 2 then
            subtree:add_expert_info(PI_MALFORMED, PI_ERROR, "Request is missing Host or Path")
            return
        end

        local host = header_parts[1]
        local path = header_parts[2]
        if path == "" then path = "/" end

        pinfo.cols.info:set(string.format("Request: %s %s", method_str, path))

        subtree:add(f_qh_host, host)
        subtree:add(f_qh_path, path)

        if #header_parts > 2 then
            local headers_tree = subtree:add(header_data_tvbr, "Headers")
            -- Headers start at index 3 of the parts table
            for i = 3, #header_parts do
                headers_tree:add(f_qh_headers, header_parts[i])
            end
        end
    end

    -- The body is everything after the ETX character
    local body_offset = etx_offset -- The offset returned by string.find is 1-based index
    if body_offset < buffer:len() then
        local body_tvb = buffer(body_offset)
        subtree:add(f_qh_body, body_tvb)
    end
end

-- Register QOTP to handle UDP traffic on port 8090
local udp_dissector_table = DissectorTable.get("udp.port")
udp_dissector_table:add(8090, qotp_protocol)
