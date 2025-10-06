-- QOTP and QH Protocol Dissector for Wireshark

--[[
This script decodes the QOTP (Quite Ok Transport Protocol) and the nested
QH (Quite Ok HTTP) protocol.

To use:
1. Save this file as `qotp_qh.lua`.
2. Place it in your personal Wireshark plugins directory.
   (Find this via Help > About Wireshark > Folders > Personal Lua Plugins).
3. Restart Wireshark.
4. Capture traffic on UDP port 8090.
--]]

-- =============================================================================
-- QOTP Dissector
-- =============================================================================
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
local f_qotp_sn_crypto = ProtoField.bytes("qotp.sn_crypto", "Encrypted SN", base.NONE)
local f_qotp_filler = ProtoField.bytes("qotp.filler", "Filler", base.NONE)
local f_qotp_decrypted_sn = ProtoField.uint64("qotp.decrypted_sn", "Decrypted SN", base.DEC)
local f_qotp_payload = ProtoField.bytes("qotp.payload", "Encrypted Payload", base.NONE)

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
    f_qotp_decrypted_sn,
    f_qotp_payload
}

-- Field extractor for the sender's ephemeral public key.
local pubkey_ep_snd_field = Field.new("qotp.pubkey_ep_snd")


function qotp_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qotp_protocol.name

    local subtree = tree:add(qotp_protocol, buffer(), "QOTP Protocol Data")
    local offset = 0

    -- Decode 1-byte header (Version and Type)
    local header_byte = buffer(offset, 1):uint()
    local version = bit.band(header_byte, 0x1F) -- Lower 5 bits
    local msg_type = bit.rshift(header_byte, 5) -- Upper 3 bits
    local type_str = f_qotp_type.valuestring[msg_type] or "Unknown"

    subtree:add(f_qotp_version, buffer(offset, 1), version)
    subtree:add(f_qotp_type, buffer(offset, 1), msg_type)
    offset = offset + 1

    pinfo.cols.info:set(string.format("Type: %s", type_str))

    -- Most messages have a Connection ID
    if msg_type ~= 0 then -- InitSnd uses part of a key as the ConnID
        subtree:add(f_qotp_conn_id, buffer(offset, 8))
    end

    -- Dissect based on message type
    if msg_type == 0 then -- InitSnd
        subtree:add(f_qotp_pubkey_ep_snd, buffer(offset, 32))
        offset = offset + 32
        subtree:add(f_qotp_pubkey_id_snd, buffer(offset, 32))
        offset = offset + 32
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

    -- After dissecting the unencrypted header, try to decrypt the rest
    if buffer:len() > offset then
        local payload_tvb = buffer(offset):tvb()
        qh_encrypted_dissector_func(payload_tvb, pinfo, tree)
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
local f_qh_status = ProtoField.uint8("qh.status_code", "Status Code (Compact)", base.DEC)
-- Common fields
local f_qh_headers = ProtoField.string("qh.headers", "Headers", base.ASCII)
local f_qh_body = ProtoField.bytes("qh.body", "Body", base.NONE)

qh_protocol.fields = {
    f_qh_version, f_qh_method, f_qh_host, f_qh_path, f_qh_status, f_qh_headers, f_qh_body
}

-- This is the dissector for the PLAINTEXT QH data after decryption.
function qh_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol:append("/QH")
    local subtree = tree:add(qh_protocol, buffer, "QH Protocol Data (Decrypted)")

    -- Find the ETX character (\x03) which separates headers from the body
    local etx_offset = buffer:find(string.char(3))
    if not etx_offset then
        subtree:add_expert_info(PI_MALFORMED, PI_ERROR, "No ETX body separator found")
        return
    end

    -- The body is everything after the ETX character
    local body_offset = etx_offset + 1
    if body_offset < buffer:len() then
        local body_tvb = buffer(body_offset)
        subtree:add(f_qh_body, body_tvb)
    end

    -- The first byte determines if it's a request or response
    local first_byte = buffer(0, 1):uint()
    local version = bit.rshift(first_byte, 6)

    -- A simple heuristic: if version is 0, we can check the method/status bits.
    -- This part can be expanded to fully parse all QH headers.
    if version == 0 then
        -- Check if it's a request (Method bits are in a known range)
        local method = bit.band(bit.rshift(first_byte, 3), 0x07)
        if method == 0 or method == 1 then -- GET or POST
            pinfo.cols.info:append(" (Request)")
            local header_data = buffer(1, etx_offset - 1):string()
            local parts = {}
            for part in header_data:gmatch("([^\0]*)") do
                table.insert(parts, part)
            end
            subtree:add(f_qh_host, parts[1])
            subtree:add(f_qh_path, parts[2])
        else -- Assume it's a response
            pinfo.cols.info:append(" (Response)")
        end
    end
end

-- This function is called by the QOTP dissector for encrypted payloads.
-- It tells Wireshark to *try* to decrypt the data.
function qh_encrypted_dissector_func(buffer, pinfo, tree)
    -- This is the magic part. To trigger decryption, we must fool Wireshark's TLS dissector.

    -- 1. Check if the client's ephemeral public key is in the CURRENT packet.
    --    This will only be true for the first packet from the client.
    local pubkey_ep_snd_info = pubkey_ep_snd_field()
    if pubkey_ep_snd_info then
        -- This is the initial client packet. We must construct a fake "Client Hello"
        -- to prime the TLS dissector with the session identifier ("Client Random").

        -- Create the "Client Random" by hashing the ephemeral key.
        -- Use ssl_md5_sha1_hash for compatibility with very old Wireshark versions.
        -- Despite its name, it can compute a standalone SHA256 hash.
        -- The first two arguments are nil because we are not computing a TLS-PRF.
        local hash_bytes_str = ssl_md5_sha1_hash(nil, nil, pubkey_ep_snd_info.range, "SHA256")
        local hash_bytes = ByteArray.new(hash_bytes_str)

        -- Create a fake "Client Hello" message containing this "Client Random".
        local client_hello = ByteArray.new("01000026" .. "0303") -- Handshake Type: Client Hello (1), Length: 38, Version: TLS 1.2 (0x0303)
        client_hello:append(hash_bytes)

        -- Create a fake TLS record containing BOTH the fake Client Hello and the real encrypted data.
        local tls_record = ByteArray.new()
        tls_record:append(ByteArray.new("1603030026"))
        tls_record:append(client_hello)
        tls_record:append(ByteArray.new("170303" .. string.format("%04x", buffer:len()))) -- App Data record header
        tls_record:append(buffer:bytes())

        local tvb = tls_record:tvb("QOTP as TLS")
        Dissector.get("tls"):call(tvb, pinfo, tree)
    else
        -- This is a subsequent packet (e.g., a server reply). The TLS session is already
        -- established in Wireshark's memory. We just need to tell the TLS dissector
        -- to decrypt this buffer as application data.
        local tls_dissector = Dissector.get("tls.decrypted_data")
        if tls_dissector then
            tls_dissector:call(buffer(), pinfo, tree)
        else
            -- Optionally log or handle the missing dissector
            print("tls.decrypted_data dissector not available")
        end

    end
end

-- Register QOTP to handle UDP traffic on port 8090
local udp_dissector_table = DissectorTable.get("udp.port")
udp_dissector_table:add(8090, qotp_protocol)

-- Register our QH dissector to be called for decrypted TLS/QOTP application data.
local tls_dissector_table = DissectorTable.get("tls.port")
tls_dissector_table:add(8090, qh_protocol)
