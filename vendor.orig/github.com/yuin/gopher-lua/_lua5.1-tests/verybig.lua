if rawget(_G, "_soft") then return 10 end

print "testing large programs (>64k)"

-- template to create a very big test file
prog = [[$

local a,b

b = {$1$
  b30009 = 65534,
  b30010 = 65535,
  b30011 = 65536,
  b30012 = 65537,
  b30013 = 16777214,
  b30014 = 16777215,
  b30015 = 16777216,
  b30016 = 16777217,
  b30017 = 4294967294,
  b30018 = 4294967295,
  b30019 = 4294967296,
  b30020 = 4294967297,
  b30021 = -65534,
  b30022 = -65535,
  b30023 = -65536,
  b30024 = -4294967297,
  b30025 = 15012.5,
  $2$
};

assert(b.a50008 == 25004 and b["a11"] == 5.5)
assert(b.a33007 == 16503.5 and b.a50009 == 25004.5)
assert(b["b"..30024] == -4294967297)

function b:xxx (a,b) return a+b end
assert(b:xxx(10, 12) == 22)   -- pushself with non-constant index
b.xxx = nil

s = 0; n=0
for a,b in pairs(b) do s=s+b; n=n+1 end
assert(s==13977183656.5  and n==70001)

require "checktable"
stat(b)

a = nil; b = nil
print'+'

function f(x) b=x end

a = f{$3$} or 10

assert(a==10)
assert(b[1] == "a10" and b[2] == 5 and b[table.getn(b)-1] == "a50009")


function xxxx (x) return b[x] end

assert(xxxx(3) == "a11")

a = nil; b=nil
xxxx = nil

return 10

]]

-- functions to fill in the $n$
F = {
function ()   -- $1$
  for i=10,50009 do
    io.write('a', i, ' = ', 5+((i-10)/2), ',\n')
  end
end,

function ()   -- $2$
  for i=30026,50009 do
    io.write('b', i, ' = ', 15013+((i-30026)/2), ',\n')
  end
end,

function ()   -- $3$
  for i=10,50009 do
    io.write('"a', i, '", ', 5+((i-10)/2), ',\n')
  end
end,
}

file = os.tmpname()
io.output(file)
for s in string.gmatch(prog, "$([^$]+)") do
  local n = tonumber(s)
  if not n then io.write(s) else F[n]() end
end
io.close()
result = dofile(file)
assert(os.remove(file))
print'OK'
return result

