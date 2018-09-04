co = coroutine.wrap(function()
  co()
end)

local ok, msg = pcall(function()
  co()
end)
assert(not ok and string.find(msg, "can not resume a running thread"))

co = coroutine.wrap(function()
  return 1
end)
assert(co() == 1)
local ok, msg = pcall(function()
  co()
end)
assert(not ok and string.find(msg, "can not resume a dead thread"))
