local proto = require "proto"
local buffer = require "buffer"
local net = require "net"

local messages, services = proto:load("echo.proto", "message/message.proto")
function run(id)
    local err, conn = net:tcp("localhost:8080")()
    assert(err == nil)
    local bytes = messages["message.Message"]:encode({ Body = "hello world" })
    local buf = buffer:new()
    buf:write_uint32_be(bytes:size())
    buf:write(bytes)
    local err = conn:write(buf:bytes())()
    assert(err == nil)

    local err, bytes = conn:read(4)()
    assert(err == nil)
    local buf = buffer:new(bytes)
    local size = buf:read_uint32_be()

    local err, bytes = conn:read(size)()
    assert(err == nil)
    local msg = messages["message.Message"]:decode(bytes)
    print(msg.Body)
    conn:close()()
end

