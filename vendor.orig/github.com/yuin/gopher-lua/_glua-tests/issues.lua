
-- issue #10
local function inspect(options)
    options = options or {}
    return type(options)
end
assert(inspect(nil) == "table")

local function inspect(options)
    options = options or setmetatable({}, {__mode = "test"})
    return type(options)
end
assert(inspect(nil) == "table")

-- issue #16
local ok, msg = pcall(function()
  local a = {}
  a[nil] = 1
end)
assert(not ok and string.find(msg, "table index is nil", 1, true))

-- issue #19
local tbl = {1,2,3,4,5}
assert(#tbl == 5)
assert(table.remove(tbl) == 5)
assert(#tbl == 4)
assert(table.remove(tbl, 3) == 3)
assert(#tbl == 3)

-- issue #24
local tbl = {string.find('hello.world', '.', 0)}
assert(tbl[1] == 1 and tbl[2] == 1)
assert(string.sub('hello.world', 0, 2) == "he")

-- issue 33
local a,b
a = function ()
  pcall(function()
  end)
  coroutine.yield("a")
  return b()
end

b = function ()
  return "b"
end

local co = coroutine.create(a)
assert(select(2, coroutine.resume(co)) == "a")
assert(select(2, coroutine.resume(co)) == "b")
assert(coroutine.status(co) == "dead")

-- issue 37
function test(a, b, c)
    b = b or string.format("b%s", a)
    c = c or string.format("c%s", a)
    assert(a == "test")
    assert(b == "btest")
    assert(c == "ctest")
end
test("test")

-- issue 39
assert(string.match("あいうえお", ".*あ.*") == "あいうえお")
assert(string.match("あいうえお", "あいうえお") == "あいうえお")

-- issue 47
assert(string.gsub("A\nA", ".", "A") == "AAA")

-- issue 62
local function level4() error("error!") end
local function level3() level4() end
local function level2() level3() end
local function level1() level2() end
local ok, result = xpcall(level1, function(err)
  return debug.traceback("msg", 10)
end)
assert(result == [[msg
stack traceback:]])
ok, result = xpcall(level1, function(err)
  return debug.traceback("msg", 9)
end)
assert(result == string.gsub([[msg
stack traceback:
@TAB@[G]: ?]], "@TAB@", "\t"))
local ok, result = xpcall(level1, function(err)
  return debug.traceback("msg", 0)
end)

assert(result == string.gsub([[msg
stack traceback:
@TAB@[G]: in function 'traceback'
@TAB@issues.lua:87: in function <issues.lua:86>
@TAB@[G]: in function 'error'
@TAB@issues.lua:71: in function 'level4'
@TAB@issues.lua:72: in function 'level3'
@TAB@issues.lua:73: in function 'level2'
@TAB@issues.lua:74: in function <issues.lua:74>
@TAB@[G]: in function 'xpcall'
@TAB@issues.lua:86: in main chunk
@TAB@[G]: ?]], "@TAB@", "\t"))

local ok, result = xpcall(level1, function(err)
  return debug.traceback("msg", 3)
end)

assert(result == string.gsub([[msg
stack traceback:
@TAB@issues.lua:71: in function 'level4'
@TAB@issues.lua:72: in function 'level3'
@TAB@issues.lua:73: in function 'level2'
@TAB@issues.lua:74: in function <issues.lua:74>
@TAB@[G]: in function 'xpcall'
@TAB@issues.lua:103: in main chunk
@TAB@[G]: ?]], "@TAB@", "\t"))

-- issue 81
local tbl = {
        [-1] = "a",
        [0] = "b",
        [1] = "c",
}
local a, b = next(tbl, nil)
assert( a == -1 and b == "a" or a == 0 and b == "b" or a == 1 and b == "c")
local a, b = next(tbl, a)
assert( a == -1 and b == "a" or a == 0 and b == "b" or a == 1 and b == "c")
local a, b = next(tbl, a)
assert( a == -1 and b == "a" or a == 0 and b == "b" or a == 1 and b == "c")
local a, b = next(tbl, a)
assert( a == nil and b == nil)

local tbl = {'a', 'b'}
local a, b = next(tbl, nil)
assert(a == 1 and b == "a")
local a, b = next(tbl, a)
assert(a == 2 and b == "b")
local a, b = next(tbl, a)
assert(a == nil and b == nil)

-- issue 82
local cr = function()
        return coroutine.wrap(function()
                coroutine.yield(1, "a")
                coroutine.yield(2, "b")
        end)
end

local f = cr()
local a, b = f()
assert(a == 1 and b == "a")
local a, b = f()
assert(a == 2 and b == "b")

-- issue 91, 92
local url = "www.aaa.bbb_abc123-321-cba_abc123"
assert(string.match(url, ".-([%w-]*)[.]*") == "www")

local s = "hello.world"
assert(s:match("([^.]+).world") == "hello")

local s = "hello-world"
assert(s:match("([^-]+)-world") == "hello")

-- issue 93
local t = {}
local ok, msg = pcall(function() t.notfound() end)
assert(not ok and string.find(msg, "attempt to call a non-function object", 1, true))

-- issue 150
local util = {
  fn = function() end
}
local b
local x = util.fn(
  1,
  (b or {}).x)

local s = [=[["a"]['b'][9] - ["a"]['b'][8] > ]=]
local result = {}
for i in s:gmatch([=[[[][^%s,]*[]]]=]) do 
  table.insert(result, i)
end
assert(result[1] == [=[["a"]['b'][9]]=])
assert(result[2] == [=[["a"]['b'][8]]=])

-- issue 168
local expected = 1

local result = math.random(1)

assert(result == expected)
