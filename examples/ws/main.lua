local ws = require "websocket"

function run(id)
    local err, conn = ws:open("ws://localhost:9100")()
    assert(err == nil)

    local err = conn:write("hello world") ()
    assert(err == nil)

    local err, message = conn:read()()
    assert(err == nil)
    print(message:string())
    assert(message:string() == "you said hello world")

    local err = conn:close()()
    assert(err == nil)
end