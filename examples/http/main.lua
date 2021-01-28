local http = require "http"

function run(id)
    http:get("http://localhost:8080")()
end

