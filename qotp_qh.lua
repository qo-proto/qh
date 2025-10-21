-- QOTP and QH Protocol Dissector for Wireshark
-- QH Protocol Dissector for Wireshark
--[[
To use:
1. Save this file as `qotp_qh.lua`.
2. Place it in your personal Wireshark plugins directory.
   (Find this via Help > About Wireshark > Folders > Personal Lua Plugins).
3. Restart Wireshark.
--]]

-- =============================================================================
-- QOTP Transport Dissector
-- This is the main dissector that gets called for UDP port 8090.
-- =============================================================================
local qotp_transport_protocol = Proto("QOTP_Transport", "QOTP Transport Layer (Decrypted)")

-- QOTP Transport Header Fields
local f_qotp_transport_header = ProtoField.bytes("qotp.transport.header", "Transport Header", base.NONE)

qotp_transport_protocol.fields = {
    f_qotp_transport_header
}

-- =============================================================================
-- QH Dissector
-- =============================================================================
local qh_protocol = Proto("QH", "Quite Ok HTTP Protocol")

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
local f_qh_body = ProtoField.bytes("qh.body", "Body", base.NONE)
local f_qh_body_str = ProtoField.string("qh.body_str", "Body (as String)", base.ASCII)

qh_protocol.fields = {
    f_qh_version, f_qh_method, f_qh_host, f_qh_path, f_qh_status_compact, f_qh_status, f_qh_headers, f_qh_body, f_qh_body_str
}

-- Request header names mapping (from protocol/types.go)
local req_header_names = {
    [0] = "Accept",
    [1] = "Accept-Encoding",
    [2] = "Content-Type",
    [3] = "Content-Length",
}

-- Response header names mapping (from protocol/types.go)
local resp_header_names = {
    [0] = "Content-Type",
    [1] = "Content-Length",
    [2] = "Cache-Control",
    [3] = "Content-Encoding",
    [4] = "Authorization",
    [5] = "CORS",
    [6] = "ETag",
    [7] = "Date",
    [8] = "CSP",
    [9] = "X-Content-Type-Options",
    [10] = "X-Frame-Options",
}

local function get_qh_dissector()
    -- Try to get existing dissector first
    local dissector = DissectorTable.get("qh")
    if not dissector then
        -- If it doesn't exist, create and register it
        dissector = qh_protocol
    end
    return dissector
end

function qotp_transport_protocol.dissector(buffer, pinfo, tree)
    pinfo.cols.protocol = qotp_transport_protocol.name

    local subtree = tree:add(qotp_transport_protocol, buffer(), "QOTP Transport Layer")
    -- If a display filter is active, the subtree might not be created.
    if not subtree then
        return
    end

    -- The decrypted payload starts with the QOTP transport header. Its size is dynamic.
    -- We parse the first byte to determine the full header length.
    if buffer:len() < 1 then
        return -- Not enough data to dissect
    end

    local transport_header_byte = buffer(0, 1):uint()
    local is_ack = bit.band(bit.rshift(transport_header_byte, 7), 0x01) == 1
    local is_extend = bit.band(bit.rshift(transport_header_byte, 6), 0x01) == 1

    -- Calculate header length based on flags, as per the Go implementation.
    local transport_header_len = 1 + 4 + 3 -- Base: header byte + 32bit StreamID + 24bit StreamOffset
    if is_extend then
        transport_header_len = transport_header_len + 3 -- StreamOffset is now 48bit
    end
    if is_ack then
        transport_header_len = transport_header_len + 4 + 3 + 2 + 1 -- ACK fields: StreamID, StreamOffset, Len, RcvWnd
        if is_extend then
            transport_header_len = transport_header_len + 3 -- ACK StreamOffset is also 48bit
        end
    end

    if buffer:len() < transport_header_len then
        return -- Not enough data for the full header
    end

    local offset = 0
    subtree:add(f_qotp_transport_header, buffer(offset, transport_header_len))
    offset = offset + transport_header_len

    -- The rest of the buffer is the QH application payload
    qh_protocol.dissector(buffer(offset):tvb(), pinfo, tree)
end

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

    -- Calculate length of header parts (after first byte, before ETX)
    local header_len = math.max(0, etx_offset - 1 - offset) -- etx_offset is 1-based, offset is 0-based
    local header_data_tvbr = buffer(offset, header_len)
    local header_parts = {}
    for part in header_data_tvbr:string():gmatch("([^\0]*)") do
        table.insert(header_parts, part)
    end

    if is_response then
        -- Dissect as a Response
        local compact_status = bit.band(header_byte, 0x3F) -- Lower 6 bits (0-63)
        -- version_field can be nil if a display filter is active
        if version_field then
            version_field:add_le(f_qh_status_compact, buffer(offset - 1, 1), compact_status)
        end
        -- We can't decode to the full HTTP status without a large mapping table in Lua.
        -- The compact code is sufficient for debugging.
        pinfo.cols.info:set(string.format("Response, Status (Compact): %d", compact_status))

        local headers_tree = subtree:add(header_data_tvbr, "Headers")
        for i, header_val in ipairs(header_parts) do -- header_parts is 1-indexed
            local header_idx = i - 1 -- Map to 0-indexed protocol header index
            local header_name = resp_header_names[header_idx] or "Unknown Header (" .. header_idx .. ")"
            headers_tree:add(f_qh_headers, header_name .. ": " .. header_val)
        end
    else
        -- Dissect as a Request
        local method = bit.band(bit.rshift(header_byte, 3), 0x07) -- Middle 3 bits for method
        local method_str = f_qh_method.valuestring[method] or "Unknown"
        -- version_field can be nil if a display filter is active
        if version_field then
            version_field:add_le(f_qh_method, buffer(offset - 1, 1), method)
        end

        if #header_parts < 2 then
            return -- Not enough header parts for a valid request
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
            for i = 3, #header_parts do -- header_parts is 1-indexed
                local header_idx = i - 3 -- Map to 0-indexed protocol header index
                local header_name = req_header_names[header_idx] or "Unknown Header (" .. header_idx .. ")"
                headers_tree:add(f_qh_headers, header_name .. ": " .. header_parts[i])
            end
        end
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

-- Register QH to handle UDP traffic on port 8090
local udp_dissector_table = DissectorTable.get("udp.port")
udp_dissector_table:add(8090, qotp_transport_protocol)
