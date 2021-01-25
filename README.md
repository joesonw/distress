![Workflow](https://github.com/joesonw/lte/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/joesonw/lte?status.svg)](https://godoc.org/github.com/joesonw/lte)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


# Introduction

**Lua Test Environment**

> Currently, only test agent is implemented.

# Usage

`GO111MODULE=on go get github.com/joesonw/lte/cmd/ds-agent `

`ds-agent run --debug -d ./examples/http -c 10 -n 100 main.lua`

> other flags please see `ds-agent run -h`
> 
> see [examples](https://github.com/joesonw/lte/tree/master/examples) for more

# How it works

let's take following as an exmaple

```lua
local http = require "http"
local metrics = require "metrics"

local counter = metrics:counter("request", { --[[ tags --]] })

function run(id)
    local err, body, headers, status = http:get("http://localhost:10080")()
    assert(err == nil)
    assert(status == 200)
    counter:add(1)
end
```

## Note
* For each concurrent running instance (controlled by flag `--concurrency/-c`), it will only run once. And then for each request, only global function `run` is called.

* States are shared between runs for each instance, so please keep test related variables local inside `run`

* Async (network, file, sleep, etc) functions returns a `Deferred`, when you want to get the result, you can simply call the function to block.
    
    example:
    ```lua
    print(time:now())
    local a = sleep(1000000000)
    local b = sleep(1000000000)
    a()
    b()
    print(time:now()) -- only 1 second is passed
    ```


## API
### Global
##### group(name, function)
```lua
group("login related", function ()
    -- do login related test here
end)
```

##### sleep(ns)
```lua
local time = require "time"
print(time:now():string())
local block = sleep(1000000000) -- 1,000,000,000 nanoseconds = 1 second
block()
print(time:now():string())
```

##### import(file)
`mylib.lua`
```lua
local function echo(message)
    return "you said: " .. message
end

return echo
```
```lua
local echo = import "mylib.lua"
print(echo("hello"))
```

### bytes

##### new(string, encoding?)
```lua
local bytes = require "bytes"
local s = bytes:new("hello")
```

##### __add, __concat, __eq
```lua
assert((bytes:new("hello") + bytes:new(" world")):string() == "hello world")
assert((bytes:new("hello") + " world"):string() == "hello world")
assert((bytes:new("hello") .. bytes:new(" world")):string() == "hello world")
assert((bytes:new("hello") .. " world"):string() == "hello world")
assert(bytes:new("hello") == bytes:new("hello"))
```

##### bytes:size()
```lua
assert(bytes:new("hello"):size() == 5)
```

##### bytes:string(encoding?)
```lua
assert(bytes:new("af", "hex"):string("hex") == "af")
```

##### bytes:replace(string)
```lua
assert(bytes:new("test"):replace("hello"):string() == "hello")
```

### crypto

##### md5(content)
```lua
assert(crypto:md5("123456"):string() == "e10adc3949ba59abbe56e057f20f883e")
```
##### sha1(content)
##### sha256(content)
##### sha512(content)
##### sha3_224(content)
##### sha3_256(content)
##### sha3_384(content)
##### sha3_512(content)

### FS

##### open(path)
```lua
local err, file = fs:open("/path")()
```

##### create(path)
```lua
local err, file = fs:create("/path")()
```

##### remove(path)
```lua
local err = fs:remove("/path")()
```

##### remove_all(path)
```lua
local err = fs:remove_all("/dir")()
```

##### list(path)
```lua
local err, stats = fs:list("/dir")()
print(stats[1].name)
print(stats[1].size)
print(stats[1].dir)
print(stats[1].modtime:string())
```

##### stat(path)
```lua
local err, stat = fs:stat("/file")()
```

##### mkdir_all(path)
```lua
local err = fs:mkdir_all("/some/deep/path")()
```

##### exist(path)
```lua
local err, ok = fs:exist("/some_file")()
```

##### read_all(path)
```lua
local err, contents = fs:read_all("/some_file")()
```

##### write(path, content)
```lua
local err = fs:write("/file", "hello world")()
```

#### file (from `fs:open`)
##### file:read(size)
```lua
local err, contents = file:read(5)() -- read 5 bytes
```

##### file:read_all()
```lua
local err, contents = file:read_all()()
```

##### file:write(content)
```lua
local err = file:write("hello world")()
```

##### file:close()
```lua
local err = file:close()()
```

##### file:name()
```lua
print(file:name())
```
