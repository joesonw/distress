local proto = require "proto"

local messages, services = proto:load("echo.proto", "message/message.proto")
local err, cc = proto:dial("localhost:10090", { insecure = true })()
local svc = services["grpc_example.Echo"]:new(cc)
function run(id)
    local err, res = svc:Echo({ Body = "nihao" })()
    print(res.Body)
end

