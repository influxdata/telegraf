assert(math.fmod(13.5, 2) == 1.5)
assert(math.pow(7, 2) == 49)

local ok, msg = pcall(function()
  math.max()
end)
assert(not ok and string.find(msg, "wrong number of arguments"))

local ok, msg = pcall(function()
  math.min()
end)
assert(not ok and string.find(msg, "wrong number of arguments"))
