local http = require "http"
local metrics = require "metrics"

local counter = metrics:gauge("request", "request", { key = "value"})

function run(id)
    http:get("http://localhost:10080")()
    counter:record(1)
end

