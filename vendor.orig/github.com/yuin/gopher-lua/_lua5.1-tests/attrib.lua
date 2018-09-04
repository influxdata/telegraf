do --[

print "testing require"

assert(require"string" == string)
assert(require"math" == math)
assert(require"table" == table)
assert(require"io" == io)
assert(require"os" == os)
assert(require"debug" == debug)
assert(require"coroutine" == coroutine)

assert(type(package.path) == "string")
assert(type(package.cpath) == "string")
assert(type(package.loaded) == "table")
assert(type(package.preload) == "table")


local DIR = "libs/"

local function createfiles (files, preextras, posextras)
  for n,c in pairs(files) do
    io.output(DIR..n)
    io.write(string.format(preextras, n))
    io.write(c)
    io.write(string.format(posextras, n))
    io.close(io.output())
  end
end

function removefiles (files)
  for n in pairs(files) do
    os.remove(DIR..n)
  end
end

local files = {
  ["A.lua"] = "",
  ["B.lua"] = "assert(...=='B');require 'A'",
  ["A.lc"] = "",
  ["A"] = "",
  ["L"] = "",
  ["XXxX"] = "",
  ["C.lua"] = "package.loaded[...] = 25; require'C'"
}

AA = nil
local extras = [[
NAME = '%s'
REQUIRED = ...
return AA]]

createfiles(files, "", extras)


local oldpath = package.path

package.path = string.gsub("D/?.lua;D/?.lc;D/?;D/??x?;D/L", "D/", DIR)

local try = function (p, n, r)
  NAME = nil
  local rr = require(p)
  assert(NAME == n)
  assert(REQUIRED == p)
  assert(rr == r)
end

assert(require"C" == 25)
assert(require"C" == 25)
AA = nil
try('B', 'B.lua', true)
assert(package.loaded.B)
assert(require"B" == true)
assert(package.loaded.A)
package.loaded.A = nil
try('B', nil, true)   -- should not reload package
try('A', 'A.lua', true)
package.loaded.A = nil
os.remove(DIR..'A.lua')
AA = {}
try('A', 'A.lc', AA)  -- now must find second option
assert(require("A") == AA)
AA = false
try('K', 'L', false)     -- default option
try('K', 'L', false)     -- default option (should reload it)
assert(rawget(_G, "_REQUIREDNAME") == nil)

AA = "x"
try("X", "XXxX", AA)


removefiles(files)


-- testing require of sub-packages

package.path = string.gsub("D/?.lua;D/?/init.lua", "D/", DIR)

files = {
  ["P1/init.lua"] = "AA = 10",
  ["P1/xuxu.lua"] = "AA = 20",
}

createfiles(files, "module(..., package.seeall)\n", "")
AA = 0

local m = assert(require"P1")
assert(m == P1 and m._NAME == "P1" and AA == 0 and m.AA == 10)
assert(require"P1" == P1 and P1 == m)
assert(require"P1" == P1)
assert(P1._PACKAGE == "")

local m = assert(require"P1.xuxu")
assert(m == P1.xuxu and m._NAME == "P1.xuxu" and AA == 0 and m.AA == 20)
assert(require"P1.xuxu" == P1.xuxu and P1.xuxu == m)
assert(require"P1.xuxu" == P1.xuxu)
assert(require"P1" == P1)
assert(P1.xuxu._PACKAGE == "P1.")
assert(P1.AA == 10 and P1._PACKAGE == "")
assert(P1._G == _G and P1.xuxu._G == _G)



removefiles(files)


package.path = ""
assert(not pcall(require, "file_does_not_exist"))
package.path = "??\0?"
assert(not pcall(require, "file_does_not_exist1"))

package.path = oldpath

-- check 'require' error message
-- local fname = "file_does_not_exist2"
-- local m, err = pcall(require, fname)
-- for t in string.gmatch(package.path..";"..package.cpath, "[^;]+") do
--   t = string.gsub(t, "?", fname)
--   print(t, err)
--   assert(string.find(err, t, 1, true))
-- end


local function import(...)
  local f = {...}
  return function (m)
    for i=1, #f do m[f[i]] = _G[f[i]] end
  end
end

local assert, module, package = assert, module, package
X = nil; x = 0; assert(_G.x == 0)   -- `x' must be a global variable
module"X"; x = 1; assert(_M.x == 1)
module"X.a.b.c"; x = 2; assert(_M.x == 2)
module("X.a.b", package.seeall); x = 3
assert(X._NAME == "X" and X.a.b.c._NAME == "X.a.b.c" and X.a.b._NAME == "X.a.b")
assert(X._M == X and X.a.b.c._M == X.a.b.c and X.a.b._M == X.a.b)
assert(X.x == 1 and X.a.b.c.x == 2 and X.a.b.x == 3)
assert(X._PACKAGE == "" and X.a.b.c._PACKAGE == "X.a.b." and
       X.a.b._PACKAGE == "X.a.")
assert(_PACKAGE.."c" == "X.a.c")
assert(X.a._NAME == nil and X.a._M == nil)
module("X.a", import("X")) ; x = 4
assert(X.a._NAME == "X.a" and X.a.x == 4 and X.a._M == X.a)
module("X.a.b", package.seeall); assert(x == 3); x = 5
assert(_NAME == "X.a.b" and X.a.b.x == 5)

assert(X._G == nil and X.a._G == nil and X.a.b._G == _G and X.a.b.c._G == nil)

setfenv(1, _G)
assert(x == 0)

assert(not pcall(module, "x"))
assert(not pcall(module, "math.sin"))


-- testing C libraries


local p = ""   -- On Mac OS X, redefine this to "_"

-- assert(loadlib == package.loadlib)   -- only for compatibility
-- local f, err, when = package.loadlib("libs/lib1.so", p.."luaopen_lib1")
local f = nil
if not f then
  (Message or print)('\a\n >>> cannot load dynamic library <<<\n\a')
  print(err, when)
else
  f()   -- open library
  assert(require("lib1") == lib1)
  collectgarbage()
  assert(lib1.id("x") == "x")
  f = assert(package.loadlib("libs/lib1.so", p.."anotherfunc"))
  assert(f(10, 20) == "1020\n")
  f, err, when = package.loadlib("libs/lib1.so", p.."xuxu")
  assert(not f and type(err) == "string" and when == "init")
  package.cpath = "libs/?.so"
  require"lib2"
  assert(lib2.id("x") == "x")
  local fs = require"lib1.sub"
  assert(fs == lib1.sub and next(lib1.sub) == nil)
  module("lib2", package.seeall)
  f = require"-lib2"
  assert(f.id("x") == "x" and _M == f and _NAME == "lib2")
  module("lib1.sub", package.seeall)
  assert(_M == fs)
  setfenv(1, _G)
 
end
-- f, err, when = package.loadlib("donotexist", p.."xuxu")
-- assert(not f and type(err) == "string" and (when == "open" or when == "absent"))


-- testing preload

do
  local p = package
  package = {}
  p.preload.pl = function (...)
    module(...)
    function xuxu (x) return x+20 end
  end

  require"pl"
  assert(require"pl" == pl)
  assert(pl.xuxu(10) == 30)

  package = p
  assert(type(package.path) == "string")
end



end  --]

print('+')

print("testing assignments, logical operators, and constructors")

local res, res2 = 27

a, b = 1, 2+3
assert(a==1 and b==5)
a={}
function f() return 10, 11, 12 end
a.x, b, a[1] = 1, 2, f()
assert(a.x==1 and b==2 and a[1]==10)
a[f()], b, a[f()+3] = f(), a, 'x'
assert(a[10] == 10 and b == a and a[13] == 'x')

do
  local f = function (n) local x = {}; for i=1,n do x[i]=i end;
                         return unpack(x) end;
  local a,b,c
  a,b = 0, f(1)
  assert(a == 0 and b == 1)
  A,b = 0, f(1)
  assert(A == 0 and b == 1)
  a,b,c = 0,5,f(4)
  assert(a==0 and b==5 and c==1)
  a,b,c = 0,5,f(0)
  assert(a==0 and b==5 and c==nil)
end


a, b, c, d = 1 and nil, 1 or nil, (1 and (nil or 1)), 6
assert(not a and b and c and d==6)

d = 20
a, b, c, d = f()
assert(a==10 and b==11 and c==12 and d==nil)
a,b = f(), 1, 2, 3, f()
assert(a==10 and b==1)

assert(a<b == false and a>b == true)
assert((10 and 2) == 2)
assert((10 or 2) == 10)
assert((10 or assert(nil)) == 10)
assert(not (nil and assert(nil)))
assert((nil or "alo") == "alo")
assert((nil and 10) == nil)
assert((false and 10) == false)
assert((true or 10) == true)
assert((false or 10) == 10)
assert(false ~= nil)
assert(nil ~= false)
assert(not nil == true)
assert(not not nil == false)
assert(not not 1 == true)
assert(not not a == true)
assert(not not (6 or nil) == true)
assert(not not (nil and 56) == false)
assert(not not (nil and true) == false)
print('+')

a = {}
a[true] = 20
a[false] = 10
assert(a[1<2] == 20 and a[1>2] == 10)

function f(a) return a end

local a = {}
for i=3000,-3000,-1 do a[i] = i; end
a[10e30] = "alo"; a[true] = 10; a[false] = 20
assert(a[10e30] == 'alo' and a[not 1] == 20 and a[10<20] == 10)
for i=3000,-3000,-1 do assert(a[i] == i); end
a[print] = assert
a[f] = print
a[a] = a
assert(a[a][a][a][a][print] == assert)
a[print](a[a[f]] == a[print])
a = nil

a = {10,9,8,7,6,5,4,3,2; [-3]='a', [f]=print, a='a', b='ab'}
a, a.x, a.y = a, a[-3]
assert(a[1]==10 and a[-3]==a.a and a[f]==print and a.x=='a' and not a.y)
a[1], f(a)[2], b, c = {['alo']=assert}, 10, a[1], a[f], 6, 10, 23, f(a), 2
a[1].alo(a[2]==10 and b==10 and c==print)

a[2^31] = 10; a[2^31+1] = 11; a[-2^31] = 12;
a[2^32] = 13; a[-2^32] = 14; a[2^32+1] = 15; a[10^33] = 16;

assert(a[2^31] == 10 and a[2^31+1] == 11 and a[-2^31] == 12 and
       a[2^32] == 13 and a[-2^32] == 14 and a[2^32+1] == 15 and
       a[10^33] == 16)

a = nil


-- do
--   local a,i,j,b
--   a = {'a', 'b'}; i=1; j=2; b=a
--   i, a[i], a, j, a[j], a[i+j] = j, i, i, b, j, i
--   assert(i == 2 and b[1] == 1 and a == 1 and j == b and b[2] == 2 and
--          b[3] == 1)
-- end

print('OK')

return res
