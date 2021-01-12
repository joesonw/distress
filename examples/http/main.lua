local http = require "http"
local metrics = require "metrics"

function run(id)
    http:get("http://localhost:8080")()
end

