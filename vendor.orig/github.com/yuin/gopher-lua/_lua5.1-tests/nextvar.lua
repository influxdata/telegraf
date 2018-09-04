print('testing tables, next, and for')

local a = {}

-- make sure table has lots of space in hash part
for i=1,100 do a[i.."+"] = true end
for i=1,100 do a[i.."+"] = nil end
-- fill hash part with numeric indices testing size operator
for i=1,100 do
  a[i] = true
  assert(#a == i)
end


if T then
-- testing table sizes

local l2 = math.log(2)
local function log2 (x) return math.log(x)/l2 end

local function mp2 (n)   -- minimum power of 2 >= n
  local mp = 2^math.ceil(log2(n))
  assert(n == 0 or (mp/2 < n and n <= mp))
  return mp
end

local function fb (n)
  local r, nn = T.int2fb(n)
  assert(r < 256)
  return nn
end

-- test fb function
local a = 1
local lim = 2^30
while a < lim do
  local n = fb(a)
  assert(a <= n and n <= a*1.125)
  a = math.ceil(a*1.3)
end

 
local function check (t, na, nh)
  local a, h = T.querytab(t)
  if a ~= na or h ~= nh then
    print(na, nh, a, h)
    assert(nil)
  end
end

-- testing constructor sizes
local lim = 40
local s = 'return {'
for i=1,lim do
  s = s..i..','
  local s = s
  for k=0,lim do 
    local t = loadstring(s..'}')()
    assert(#t == i)
    check(t, fb(i), mp2(k))
    s = string.format('%sa%d=%d,', s, k, k)
  end
end


-- tests with unknown number of elements
local a = {}
for i=1,lim do a[i] = i end   -- build auxiliary table
for k=0,lim do
  local a = {unpack(a,1,k)}
  assert(#a == k)
  check(a, k, 0)
  a = {1,2,3,unpack(a,1,k)}
  check(a, k+3, 0)
  assert(#a == k + 3)
end


print'+'

-- testing tables dynamically built
local lim = 130
local a = {}; a[2] = 1; check(a, 0, 1)
a = {}; a[0] = 1; check(a, 0, 1); a[2] = 1; check(a, 0, 2)
a = {}; a[0] = 1; a[1] = 1; check(a, 1, 1)
a = {}
for i = 1,lim do
  a[i] = 1
  assert(#a == i)
  check(a, mp2(i), 0)
end

a = {}
for i = 1,lim do
  a['a'..i] = 1
  assert(#a == 0)
  check(a, 0, mp2(i))
end

a = {}
for i=1,16 do a[i] = i end
check(a, 16, 0)
for i=1,11 do a[i] = nil end
for i=30,40 do a[i] = nil end   -- force a rehash (?)
check(a, 0, 8)
a[10] = 1
for i=30,40 do a[i] = nil end   -- force a rehash (?)
check(a, 0, 8)
for i=1,14 do a[i] = nil end
for i=30,50 do a[i] = nil end   -- force a rehash (?)
check(a, 0, 4)

-- reverse filling
for i=1,lim do
  local a = {}
  for i=i,1,-1 do a[i] = i end   -- fill in reverse
  check(a, mp2(i), 0)
end

-- size tests for vararg
lim = 35
function foo (n, ...)
  local arg = {...}
  check(arg, n, 0)
  assert(select('#', ...) == n)
  arg[n+1] = true
  check(arg, mp2(n+1), 0)
  arg.x = true
  check(arg, mp2(n+1), 1)
end
local a = {}
for i=1,lim do a[i] = true; foo(i, unpack(a)) end

end


-- test size operation on empty tables
assert(#{} == 0)
assert(#{nil} == 0)
assert(#{nil, nil} == 0)
assert(#{nil, nil, nil} == 0)
assert(#{nil, nil, nil, nil} == 0)
print'+'


local nofind = {}

a,b,c = 1,2,3
a,b,c = nil

local function find (name)
  local n,v
  while 1 do
    n,v = next(_G, n)
    if not n then return nofind end
    assert(v ~= nil)
    if n == name then return v end
  end
end

local function find1 (name)
  for n,v in pairs(_G) do
    if n==name then return v end
  end
  return nil  -- not found
end

do   -- create 10000 new global variables
  for i=1,10000 do _G[i] = i end
end


a = {x=90, y=8, z=23}
assert(table.foreach(a, function(i,v) if i=='x' then return v end end) == 90)
assert(table.foreach(a, function(i,v) if i=='a' then return v end end) == nil)
table.foreach({}, error)

table.foreachi({x=10, y=20}, error)
local a = {n = 1}
table.foreachi({n=3}, function (i, v)
  assert(a.n == i and not v)
  a.n=a.n+1
end)
a = {10,20,30,nil,50}
table.foreachi(a, function (i,v) assert(a[i] == v) end)
assert(table.foreachi({'a', 'b', 'c'}, function (i,v)
         if i==2 then return v end
       end) == 'b')


assert(print==find("print") and print == find1("print"))
assert(_G["print"]==find("print"))
assert(assert==find1("assert"))
assert(nofind==find("return"))
assert(not find1("return"))
_G["ret" .. "urn"] = nil
assert(nofind==find("return"))
_G["xxx"] = 1
assert(xxx==find("xxx"))
print('+')

a = {}
for i=0,10000 do
  if math.mod(i,10) ~= 0 then
    a['x'..i] = i
  end
end

n = {n=0}
for i,v in pairs(a) do
  n.n = n.n+1
  assert(i and v and a[i] == v)
end
assert(n.n == 9000)
a = nil

-- remove those 10000 new global variables
for i=1,10000 do _G[i] = nil end

do   -- clear global table
  local a = {}
  local preserve = {io = 1, string = 1, debug = 1, os = 1,
                    coroutine = 1, table = 1, math = 1}
  for n,v in pairs(_G) do a[n]=v end
  for n,v in pairs(a) do
    if not preserve[n] and type(v) ~= "function" and
       not string.find(n, "^[%u_]") then
     _G[n] = nil
    end
    collectgarbage()
  end
end

local function foo ()
  local getfenv, setfenv, assert, next =
        getfenv, setfenv, assert, next
  local n = {gl1=3}
  setfenv(foo, n)
  assert(getfenv(foo) == getfenv(1))
  assert(getfenv(foo) == n)
  assert(print == nil and gl1 == 3)
  gl1 = nil
  gl = 1
  assert(n.gl == 1 and next(n, 'gl') == nil)
end
foo()

print'+'

local function checknext (a)
  local b = {}
  table.foreach(a, function (k,v) b[k] = v end)
  for k,v in pairs(b) do assert(a[k] == v) end
  for k,v in pairs(a) do assert(b[k] == v) end
  b = {}
  do local k,v = next(a); while k do b[k] = v; k,v = next(a,k) end end
  for k,v in pairs(b) do assert(a[k] == v) end
  for k,v in pairs(a) do assert(b[k] == v) end
end

checknext{1,x=1,y=2,z=3}
checknext{1,2,x=1,y=2,z=3}
checknext{1,2,3,x=1,y=2,z=3}
checknext{1,2,3,4,x=1,y=2,z=3}
checknext{1,2,3,4,5,x=1,y=2,z=3}

assert(table.getn{} == 0)
assert(table.getn{[-1] = 2} == 0)
assert(table.getn{1,2,3,nil,nil} == 3)
for i=0,40 do
  local a = {}
  for j=1,i do a[j]=j end
  assert(table.getn(a) == i)
end


assert(table.maxn{} == 0)
assert(table.maxn{["1000"] = true} == 0)
assert(table.maxn{["1000"] = true, [24.5] = 3} == 24.5)
assert(table.maxn{[1000] = true} == 1000)
assert(table.maxn{[10] = true, [100*math.pi] = print} == 100*math.pi)


-- int overflow
a = {}
for i=0,50 do a[math.pow(2,i)] = true end
assert(a[table.getn(a)])

print("+")


-- erasing values
local t = {[{1}] = 1, [{2}] = 2, [string.rep("x ", 4)] = 3,
           [100.3] = 4, [4] = 5}

local n = 0
for k, v in pairs( t ) do
  n = n+1
  assert(t[k] == v)
  t[k] = nil
  collectgarbage()
  assert(t[k] == nil)
end
assert(n == 5)


local function test (a)
  table.insert(a, 10); table.insert(a, 2, 20);
  table.insert(a, 1, -1); table.insert(a, 40);
  table.insert(a, table.getn(a)+1, 50)
  table.insert(a, 2, -2)
  assert(table.remove(a,1) == -1)
  assert(table.remove(a,1) == -2)
  assert(table.remove(a,1) == 10)
  assert(table.remove(a,1) == 20)
  assert(table.remove(a,1) == 40)
  assert(table.remove(a,1) == 50)
  assert(table.remove(a,1) == nil)
end

a = {n=0, [-7] = "ban"}
test(a)
assert(a.n == 0 and a[-7] == "ban")

a = {[-7] = "ban"};
test(a)
assert(a.n == nil and table.getn(a) == 0 and a[-7] == "ban")


table.insert(a, 1, 10); table.insert(a, 1, 20); table.insert(a, 1, -1)
assert(table.remove(a) == 10)
assert(table.remove(a) == 20)
assert(table.remove(a) == -1)

a = {'c', 'd'}
table.insert(a, 3, 'a')
table.insert(a, 'b')
assert(table.remove(a, 1) == 'c')
assert(table.remove(a, 1) == 'd')
assert(table.remove(a, 1) == 'a')
assert(table.remove(a, 1) == 'b')
assert(table.getn(a) == 0 and a.n == nil)
print("+")

a = {}
for i=1,1000 do
  a[i] = i; a[i-1] = nil
end
assert(next(a,nil) == 1000 and next(a,1000) == nil)

assert(next({}) == nil)
assert(next({}, nil) == nil)

for a,b in pairs{} do error"not here" end
for i=1,0 do error'not here' end
for i=0,1,-1 do error'not here' end
a = nil; for i=1,1 do assert(not a); a=1 end; assert(a)
a = nil; for i=1,1,-1 do assert(not a); a=1 end; assert(a)

a = 0; for i=0, 1, 0.1 do a=a+1 end; assert(a==11)
-- precision problems
--a = 0; for i=1, 0, -0.01 do a=a+1 end; assert(a==101)
a = 0; for i=0, 0.999999999, 0.1 do a=a+1 end; assert(a==10)
a = 0; for i=1, 1, 1 do a=a+1 end; assert(a==1)
a = 0; for i=1e10, 1e10, -1 do a=a+1 end; assert(a==1)
a = 0; for i=1, 0.99999, 1 do a=a+1 end; assert(a==0)
a = 0; for i=99999, 1e5, -1 do a=a+1 end; assert(a==0)
a = 0; for i=1, 0.99999, -1 do a=a+1 end; assert(a==1)

-- conversion
a = 0; for i="10","1","-2" do a=a+1 end; assert(a==5)


collectgarbage()


-- testing generic 'for'

local function f (n, p)
  local t = {}; for i=1,p do t[i] = i*10 end
  return function (_,n)
           if n > 0 then
             n = n-1
             return n, unpack(t)
           end
         end, nil, n
end

local x = 0
for n,a,b,c,d in f(5,3) do
  x = x+1
  assert(a == 10 and b == 20 and c == 30 and d == nil)
end
assert(x == 5)

print"OK"
