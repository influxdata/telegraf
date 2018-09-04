local ok, msg = pcall(function()
  dofile("notexist")
end)
assert(not ok and string.find(msg, ".*notexist.*"))

local ok, msg = pcall(function()
  assert(getfenv(2) == _G)
end)
assert(ok)

local i = 1
local fn = assert(load(function()
  local tbl = {"return ", "1", "+", "1"}
  local v = tbl[i]
  i = i + 1
  return v
end))
assert(fn() == 2)

local fn, msg = load(function()
    return {}
end)
assert(not fn and string.find(msg, "must return a string"))

local i = 1
local fn, msg = load(function()
  if i == 1 then
      i = i + 1
      return "returna"
  end
end)
assert(not fn and string.find(string.lower(msg), "eof"))

local ok, a, b = xpcall(function()
     return "a", "b"
  end, 
  function(err)
     assert(nil)
  end)
assert(ok and a == "a" and b == "b")

local ok, a, b = xpcall(function()
     error("error!")
  end, 
  function(err)
     return err .. "!", "b"
  end)
assert(not ok and string.find(a, "error!!") and b == nil)
