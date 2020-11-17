local http = require "http"
local metrics = require "metrics"

local counter = metrics:counter("request", { key = "value"})

function run(id)
    http:get("http://localhost:10080")()
    counter:add(1)
end

