local a = {}
assert(table.maxn(a) == 0)
a["key"] = 1
assert(table.maxn(a) == 0)
table.insert(a, 10)
table.insert(a, 3, 10)
assert(table.maxn(a) == 3)

local ok, msg = pcall(function()
  table.insert(a)
end)
assert(not ok and string.find(msg, "wrong number of arguments"))

a = {}
a["key0"] = "0"
a["key1"] = "1"
a[1] = 1
a[2] = 2
a[true] = "true"
a[false] = "false"
for k, v in pairs(a) do
  if k == "key0" then
    assert(v == "0")
  elseif k == "key1" then
    assert(v == "1")
  elseif k == 1 then
    assert(v == 1)
  elseif k == 2 then
    assert(v == 2)
  elseif k == true then
    assert(v == "true")
  elseif k == false then
    assert(v == "false")
  else
    error("unexpected key:" .. tostring(k))
  end
end
