
if T==nil then
  (Message or print)('\a\n >>> testC not active: skipping API tests <<<\n\a')
  return
end



function tcheck (t1, t2)
  table.remove(t1, 1)  -- remove code
  assert(table.getn(t1) == table.getn(t2))
  for i=1,table.getn(t1) do assert(t1[i] == t2[i]) end
end

function pack(...) return arg end


print('testing C API')

-- testing allignment
a = T.d2s(12458954321123)
assert(string.len(a) == 8)   -- sizeof(double)
assert(T.s2d(a) == 12458954321123)

a,b,c = T.testC("pushnum 1; pushnum 2; pushnum 3; return 2")
assert(a == 2 and b == 3 and not c)

-- test that all trues are equal
a,b,c = T.testC("pushbool 1; pushbool 2; pushbool 0; return 3")
assert(a == b and a == true and c == false)
a,b,c = T.testC"pushbool 0; pushbool 10; pushnil;\
                      tobool -3; tobool -3; tobool -3; return 3"
assert(a==0 and b==1 and c==0)


a,b,c = T.testC("gettop; return 2", 10, 20, 30, 40)
assert(a == 40 and b == 5 and not c)

t = pack(T.testC("settop 5; gettop; return .", 2, 3))
tcheck(t, {n=4,2,3})

t = pack(T.testC("settop 0; settop 15; return 10", 3, 1, 23))
assert(t.n == 10 and t[1] == nil and t[10] == nil)

t = pack(T.testC("remove -2; gettop; return .", 2, 3, 4))
tcheck(t, {n=2,2,4})

t = pack(T.testC("insert -1; gettop; return .", 2, 3))
tcheck(t, {n=2,2,3})

t = pack(T.testC("insert 3; gettop; return .", 2, 3, 4, 5))
tcheck(t, {n=4,2,5,3,4})

t = pack(T.testC("replace 2; gettop; return .", 2, 3, 4, 5))
tcheck(t, {n=3,5,3,4})

t = pack(T.testC("replace -2; gettop; return .", 2, 3, 4, 5))
tcheck(t, {n=3,2,3,5})

t = pack(T.testC("remove 3; gettop; return .", 2, 3, 4, 5))
tcheck(t, {n=3,2,4,5})

t = pack(T.testC("insert 3; pushvalue 3; remove 3; pushvalue 2; remove 2; \
                  insert 2; pushvalue 1; remove 1; insert 1; \
      insert -2; pushvalue -2; remove -3; gettop; return .",
      2, 3, 4, 5, 10, 40, 90))
tcheck(t, {n=7,2,3,4,5,10,40,90})

t = pack(T.testC("concat 5; gettop; return .", "alo", 2, 3, "joao", 12))
tcheck(t, {n=1,"alo23joao12"})

-- testing MULTRET
t = pack(T.testC("rawcall 2,-1; gettop; return .",
     function (a,b) return 1,2,3,4,a,b end, "alo", "joao"))
tcheck(t, {n=6,1,2,3,4,"alo", "joao"})

do  -- test returning more results than fit in the caller stack
  local a = {}
  for i=1,1000 do a[i] = true end; a[999] = 10
  local b = T.testC([[call 1 -1; pop 1; tostring -1; return 1]], unpack, a)
  assert(b == "10")
end


-- testing lessthan
assert(T.testC("lessthan 2 5, return 1", 3, 2, 2, 4, 2, 2))
assert(T.testC("lessthan 5 2, return 1", 4, 2, 2, 3, 2, 2))
assert(not T.testC("lessthan 2 -3, return 1", "4", "2", "2", "3", "2", "2"))
assert(not T.testC("lessthan -3 2, return 1", "3", "2", "2", "4", "2", "2"))

local b = {__lt = function (a,b) return a[1] < b[1] end}
local a1,a3,a4 = setmetatable({1}, b),
                 setmetatable({3}, b),
                 setmetatable({4}, b)
assert(T.testC("lessthan 2 5, return 1", a3, 2, 2, a4, 2, 2))
assert(T.testC("lessthan 5 -6, return 1", a4, 2, 2, a3, 2, 2))
a,b = T.testC("lessthan 5 -6, return 2", a1, 2, 2, a3, 2, 20)
assert(a == 20 and b == false)


-- testing lua_is

function count (x, n)
  n = n or 2
  local prog = [[
    isnumber %d;
    isstring %d;
    isfunction %d;
    iscfunction %d;
    istable %d;
    isuserdata %d;
    isnil %d;
    isnull %d;
    return 8
  ]]
  prog = string.format(prog, n, n, n, n, n, n, n, n)
  local a,b,c,d,e,f,g,h = T.testC(prog, x)
  return a+b+c+d+e+f+g+(100*h)
end

assert(count(3) == 2)
assert(count('alo') == 1)
assert(count('32') == 2)
assert(count({}) == 1)
assert(count(print) == 2)
assert(count(function () end) == 1)
assert(count(nil) == 1)
assert(count(io.stdin) == 1)
assert(count(nil, 15) == 100)

-- testing lua_to...

function to (s, x, n)
  n = n or 2
  return T.testC(string.format("%s %d; return 1", s, n), x)
end

assert(to("tostring", {}) == nil)
assert(to("tostring", "alo") == "alo")
assert(to("tostring", 12) == "12")
assert(to("tostring", 12, 3) == nil)
assert(to("objsize", {}) == 0)
assert(to("objsize", "alo\0\0a") == 6)
assert(to("objsize", T.newuserdata(0)) == 0)
assert(to("objsize", T.newuserdata(101)) == 101)
assert(to("objsize", 12) == 2)
assert(to("objsize", 12, 3) == 0)
assert(to("tonumber", {}) == 0)
assert(to("tonumber", "12") == 12)
assert(to("tonumber", "s2") == 0)
assert(to("tonumber", 1, 20) == 0)
a = to("tocfunction", math.deg)
assert(a(3) == math.deg(3) and a ~= math.deg)


-- testing errors

a = T.testC([[
  loadstring 2; call 0,1;
  pushvalue 3; insert -2; call 1, 1;
  call 0, 0;
  return 1
]], "x=150", function (a) assert(a==nil); return 3 end)

assert(type(a) == 'string' and x == 150)

function check3(p, ...)
  assert(arg.n == 3)
  assert(string.find(arg[3], p))
end
check3(":1:", T.testC("loadstring 2; gettop; return .", "x="))
check3("cannot read", T.testC("loadfile 2; gettop; return .", "."))
check3("cannot open xxxx", T.testC("loadfile 2; gettop; return .", "xxxx"))

-- testing table access

a = {x=0, y=12}
x, y = T.testC("gettable 2; pushvalue 4; gettable 2; return 2",
                a, 3, "y", 4, "x")
assert(x == 0 and y == 12)
T.testC("settable -5", a, 3, 4, "x", 15)
assert(a.x == 15)
a[a] = print
x = T.testC("gettable 2; return 1", a)  -- table and key are the same object!
assert(x == print)
T.testC("settable 2", a, "x")    -- table and key are the same object!
assert(a[a] == "x")

b = setmetatable({p = a}, {})
getmetatable(b).__index = function (t, i) return t.p[i] end
k, x = T.testC("gettable 3, return 2", 4, b, 20, 35, "x")
assert(x == 15 and k == 35)
getmetatable(b).__index = function (t, i) return a[i] end
getmetatable(b).__newindex = function (t, i,v ) a[i] = v end
y = T.testC("insert 2; gettable -5; return 1", 2, 3, 4, "y", b)
assert(y == 12)
k = T.testC("settable -5, return 1", b, 3, 4, "x", 16)
assert(a.x == 16 and k == 4)
a[b] = 'xuxu'
y = T.testC("gettable 2, return 1", b)
assert(y == 'xuxu')
T.testC("settable 2", b, 19)
assert(a[b] == 19)

-- testing next
a = {}
t = pack(T.testC("next; gettop; return .", a, nil))
tcheck(t, {n=1,a})
a = {a=3}
t = pack(T.testC("next; gettop; return .", a, nil))
tcheck(t, {n=3,a,'a',3})
t = pack(T.testC("next; pop 1; next; gettop; return .", a, nil))
tcheck(t, {n=1,a})



-- testing upvalues

do
  local A = T.testC[[ pushnum 10; pushnum 20; pushcclosure 2; return 1]]
  t, b, c = A([[pushvalue U0; pushvalue U1; pushvalue U2; return 3]])
  assert(b == 10 and c == 20 and type(t) == 'table')
  a, b = A([[tostring U3; tonumber U4; return 2]])
  assert(a == nil and b == 0)
  A([[pushnum 100; pushnum 200; replace U2; replace U1]])
  b, c = A([[pushvalue U1; pushvalue U2; return 2]])
  assert(b == 100 and c == 200)
  A([[replace U2; replace U1]], {x=1}, {x=2})
  b, c = A([[pushvalue U1; pushvalue U2; return 2]])
  assert(b.x == 1 and c.x == 2)
  T.checkmemory()
end

local f = T.testC[[ pushnum 10; pushnum 20; pushcclosure 2; return 1]]
assert(T.upvalue(f, 1) == 10 and
       T.upvalue(f, 2) == 20 and
       T.upvalue(f, 3) == nil)
T.upvalue(f, 2, "xuxu")
assert(T.upvalue(f, 2) == "xuxu")


-- testing environments

assert(T.testC"pushvalue G; return 1" == _G)
assert(T.testC"pushvalue E; return 1" == _G)
local a = {}
T.testC("replace E; return 1", a)
assert(T.testC"pushvalue G; return 1" == _G)
assert(T.testC"pushvalue E; return 1" == a)
assert(debug.getfenv(T.testC) == a)
assert(debug.getfenv(T.upvalue) == _G)
-- userdata inherit environment
local u = T.testC"newuserdata 0; return 1"
assert(debug.getfenv(u) == a)
-- functions inherit environment
u = T.testC"pushcclosure 0; return 1"
assert(debug.getfenv(u) == a)
debug.setfenv(T.testC, _G)
assert(T.testC"pushvalue E; return 1" == _G)

local b = newproxy()
assert(debug.getfenv(b) == _G)
assert(debug.setfenv(b, a))
assert(debug.getfenv(b) == a)



-- testing locks (refs)

-- reuse of references
local i = T.ref{}
T.unref(i)
assert(T.ref{} == i)

Arr = {}
Lim = 100
for i=1,Lim do   -- lock many objects
  Arr[i] = T.ref({})
end

assert(T.ref(nil) == -1 and T.getref(-1) == nil)
T.unref(-1); T.unref(-1)

for i=1,Lim do   -- unlock all them
  T.unref(Arr[i])
end

function printlocks ()
  local n = T.testC("gettable R; return 1", "n")
  print("n", n)
  for i=0,n do
    print(i, T.testC("gettable R; return 1", i))
  end
end


for i=1,Lim do   -- lock many objects
  Arr[i] = T.ref({})
end

for i=1,Lim,2 do   -- unlock half of them
  T.unref(Arr[i])
end

assert(type(T.getref(Arr[2])) == 'table')


assert(T.getref(-1) == nil)


a = T.ref({})

collectgarbage()

assert(type(T.getref(a)) == 'table')


-- colect in cl the `val' of all collected userdata
tt = {}
cl = {n=0}
A = nil; B = nil
local F
F = function (x)
  local udval = T.udataval(x)
  table.insert(cl, udval)
  local d = T.newuserdata(100)   -- cria lixo
  d = nil
  assert(debug.getmetatable(x).__gc == F)
  loadstring("table.insert({}, {})")()   -- cria mais lixo
  collectgarbage()   -- forca coleta de lixo durante coleta!
  assert(debug.getmetatable(x).__gc == F)   -- coleta anterior nao melou isso?
  local dummy = {}    -- cria lixo durante coleta
  if A ~= nil then
    assert(type(A) == "userdata")
    assert(T.udataval(A) == B)
    debug.getmetatable(A)    -- just acess it
  end
  A = x   -- ressucita userdata
  B = udval
  return 1,2,3
end
tt.__gc = F

-- test whether udate collection frees memory in the right time
do
  collectgarbage();
  collectgarbage();
  local x = collectgarbage("count");
  local a = T.newuserdata(5001)
  assert(T.testC("objsize 2; return 1", a) == 5001)
  assert(collectgarbage("count") >= x+4) 
  a = nil
  collectgarbage();
  assert(collectgarbage("count") <= x+1)
  -- udata without finalizer
  x = collectgarbage("count")
  collectgarbage("stop")
  for i=1,1000 do newproxy(false) end
  assert(collectgarbage("count") > x+10)
  collectgarbage()
  assert(collectgarbage("count") <= x+1)
  -- udata with finalizer
  x = collectgarbage("count")
  collectgarbage()
  collectgarbage("stop")
  a = newproxy(true)
  getmetatable(a).__gc = function () end
  for i=1,1000 do newproxy(a) end
  assert(collectgarbage("count") >= x+10)
  collectgarbage()  -- this collection only calls TM, without freeing memory
  assert(collectgarbage("count") >= x+10)
  collectgarbage()  -- now frees memory
  assert(collectgarbage("count") <= x+1)
end


collectgarbage("stop")

-- create 3 userdatas with tag `tt'
a = T.newuserdata(0); debug.setmetatable(a, tt); na = T.udataval(a)
b = T.newuserdata(0); debug.setmetatable(b, tt); nb = T.udataval(b)
c = T.newuserdata(0); debug.setmetatable(c, tt); nc = T.udataval(c)

-- create userdata without meta table
x = T.newuserdata(4)
y = T.newuserdata(0)

assert(debug.getmetatable(x) == nil and debug.getmetatable(y) == nil)

d=T.ref(a);
e=T.ref(b);
f=T.ref(c);
t = {T.getref(d), T.getref(e), T.getref(f)}
assert(t[1] == a and t[2] == b and t[3] == c)

t=nil; a=nil; c=nil;
T.unref(e); T.unref(f)

collectgarbage()

-- check that unref objects have been collected
assert(table.getn(cl) == 1 and cl[1] == nc)

x = T.getref(d)
assert(type(x) == 'userdata' and debug.getmetatable(x) == tt)
x =nil
tt.b = b  -- create cycle
tt=nil    -- frees tt for GC
A = nil
b = nil
T.unref(d);
n5 = T.newuserdata(0)
debug.setmetatable(n5, {__gc=F})
n5 = T.udataval(n5)
collectgarbage()
assert(table.getn(cl) == 4)
-- check order of collection
assert(cl[2] == n5 and cl[3] == nb and cl[4] == na)


a, na = {}, {}
for i=30,1,-1 do
  a[i] = T.newuserdata(0)
  debug.setmetatable(a[i], {__gc=F})
  na[i] = T.udataval(a[i])
end
cl = {}
a = nil; collectgarbage()
assert(table.getn(cl) == 30)
for i=1,30 do assert(cl[i] == na[i]) end
na = nil


for i=2,Lim,2 do   -- unlock the other half
  T.unref(Arr[i])
end

x = T.newuserdata(41); debug.setmetatable(x, {__gc=F})
assert(T.testC("objsize 2; return 1", x) == 41)
cl = {}
a = {[x] = 1}
x = T.udataval(x)
collectgarbage()
-- old `x' cannot be collected (`a' still uses it)
assert(table.getn(cl) == 0)
for n in pairs(a) do a[n] = nil end
collectgarbage()
assert(table.getn(cl) == 1 and cl[1] == x)   -- old `x' must be collected

-- testing lua_equal
assert(T.testC("equal 2 4; return 1", print, 1, print, 20))
assert(T.testC("equal 3 2; return 1", 'alo', "alo"))
assert(T.testC("equal 2 3; return 1", nil, nil))
assert(not T.testC("equal 2 3; return 1", {}, {}))
assert(not T.testC("equal 2 3; return 1"))
assert(not T.testC("equal 2 3; return 1", 3))

-- testing lua_equal with fallbacks
do
  local map = {}
  local t = {__eq = function (a,b) return map[a] == map[b] end}
  local function f(x)
    local u = T.newuserdata(0)
    debug.setmetatable(u, t)
    map[u] = x
    return u
  end
  assert(f(10) == f(10))
  assert(f(10) ~= f(11))
  assert(T.testC("equal 2 3; return 1", f(10), f(10)))
  assert(not T.testC("equal 2 3; return 1", f(10), f(20)))
  t.__eq = nil
  assert(f(10) ~= f(10))
end

print'+'



-------------------------------------------------------------------------
do   -- testing errors during GC
  local a = {}
  for i=1,20 do
    a[i] = T.newuserdata(i)   -- creates several udata
  end
  for i=1,20,2 do   -- mark half of them to raise error during GC
    debug.setmetatable(a[i], {__gc = function (x) error("error inside gc") end})
  end
  for i=2,20,2 do   -- mark the other half to count and to create more garbage
    debug.setmetatable(a[i], {__gc = function (x) loadstring("A=A+1")() end})
  end
  _G.A = 0
  a = 0
  while 1 do
  if xpcall(collectgarbage, function (s) a=a+1 end) then
    break   -- stop if no more errors
  end
  end
  assert(a == 10)  -- number of errors
  assert(A == 10)  -- number of normal collections
end
-------------------------------------------------------------------------
-- test for userdata vals
do
  local a = {}; local lim = 30
  for i=0,lim do a[i] = T.pushuserdata(i) end
  for i=0,lim do assert(T.udataval(a[i]) == i) end
  for i=0,lim do assert(T.pushuserdata(i) == a[i]) end
  for i=0,lim do a[a[i]] = i end
  for i=0,lim do a[T.pushuserdata(i)] = i end
  assert(type(tostring(a[1])) == "string")
end


-------------------------------------------------------------------------
-- testing multiple states
T.closestate(T.newstate());
L1 = T.newstate()
assert(L1)
assert(pack(T.doremote(L1, "function f () return 'alo', 3 end; f()")).n == 0)

a, b = T.doremote(L1, "return f()")
assert(a == 'alo' and b == '3')

T.doremote(L1, "_ERRORMESSAGE = nil")
-- error: `sin' is not defined
a, b = T.doremote(L1, "return sin(1)")
assert(a == nil and b == 2)   -- 2 == run-time error

-- error: syntax error
a, b, c = T.doremote(L1, "return a+")
assert(a == nil and b == 3 and type(c) == "string")   -- 3 == syntax error

T.loadlib(L1)
a, b = T.doremote(L1, [[
  a = strlibopen()
  a = packageopen()
  a = baselibopen(); assert(a == _G and require("_G") == a)
  a = iolibopen(); assert(type(a.read) == "function")
  assert(require("io") == a)
  a = tablibopen(); assert(type(a.insert) == "function")
  a = dblibopen(); assert(type(a.getlocal) == "function")
  a = mathlibopen(); assert(type(a.sin) == "function")
  return string.sub('okinama', 1, 2)
]])
assert(a == "ok")

T.closestate(L1);

L1 = T.newstate()
T.loadlib(L1)
T.doremote(L1, "a = {}")
T.testC(L1, [[pushstring a; gettable G; pushstring x; pushnum 1;
             settable -3]])
assert(T.doremote(L1, "return a.x") == "1")

T.closestate(L1)

L1 = nil

print('+')

-------------------------------------------------------------------------
-- testing memory limits
-------------------------------------------------------------------------
collectgarbage()
T.totalmem(T.totalmem()+5000)   -- set low memory limit (+5k)
assert(not pcall(loadstring"local a={}; for i=1,100000 do a[i]=i end"))
T.totalmem(1000000000)          -- restore high limit


local function stack(x) if x>0 then stack(x-1) end end

-- test memory errors; increase memory limit in small steps, so that
-- we get memory errors in different parts of a given task, up to there
-- is enough memory to complete the task without errors
function testamem (s, f)
  collectgarbage()
  stack(10)    -- ensure minimum stack size
  local M = T.totalmem()
  local oldM = M
  local a,b = nil
  while 1 do
    M = M+3   -- increase memory limit in small steps
    T.totalmem(M)
    a, b = pcall(f)
    if a and b then break end       -- stop when no more errors
    collectgarbage()
    if not a and not string.find(b, "memory") then   -- `real' error?
      T.totalmem(1000000000)  -- restore high limit
      error(b, 0)
    end
  end
  T.totalmem(1000000000)  -- restore high limit
  print("\nlimit for " .. s .. ": " .. M-oldM)
  return b
end


-- testing memory errors when creating a new state

b = testamem("state creation", T.newstate)
T.closestate(b);  -- close new state


-- testing threads

function expand (n,s)
  if n==0 then return "" end
  local e = string.rep("=", n)
  return string.format("T.doonnewstack([%s[ %s;\n collectgarbage(); %s]%s])\n",
                              e, s, expand(n-1,s), e)
end

G=0; collectgarbage(); a =collectgarbage("count")
loadstring(expand(20,"G=G+1"))()
assert(G==20); collectgarbage();  -- assert(gcinfo() <= a+1)

testamem("thread creation", function ()
  return T.doonnewstack("x=1") == 0  -- try to create thread
end)


-- testing memory x compiler

testamem("loadstring", function ()
  return loadstring("x=1")  -- try to do a loadstring
end)


local testprog = [[
local function foo () return end
local t = {"x"}
a = "aaa"
for _, v in ipairs(t) do a=a..v end
return true
]]

-- testing memory x dofile
_G.a = nil
local t =os.tmpname()
local f = assert(io.open(t, "w"))
f:write(testprog)
f:close()
testamem("dofile", function ()
  local a = loadfile(t)
  return a and a()
end)
assert(os.remove(t))
assert(_G.a == "aaax")


-- other generic tests

testamem("string creation", function ()
  local a, b = string.gsub("alo alo", "(a)", function (x) return x..'b' end)
  return (a == 'ablo ablo')
end)

testamem("dump/undump", function ()
  local a = loadstring(testprog)
  local b = a and string.dump(a)
  a = b and loadstring(b)
  return a and a()
end)

local t = os.tmpname()
testamem("file creation", function ()
  local f = assert(io.open(t, 'w'))
  assert (not io.open"nomenaoexistente")
  io.close(f);
  return not loadfile'nomenaoexistente'
end)
assert(os.remove(t))

testamem("table creation", function ()
  local a, lim = {}, 10
  for i=1,lim do a[i] = i; a[i..'a'] = {} end
  return (type(a[lim..'a']) == 'table' and a[lim] == lim)
end)

local a = 1
close = nil
testamem("closure creation", function ()
  function close (b,c)
   return function (x) return a+b+c+x end
  end
  return (close(2,3)(4) == 10)
end)

testamem("coroutines", function ()
  local a = coroutine.wrap(function ()
              coroutine.yield(string.rep("a", 10))
              return {}
            end)
  assert(string.len(a()) == 10)
  return a()
end)

print'+'

-- testing some auxlib functions
assert(T.gsub("alo.alo.uhuh.", ".", "//") == "alo//alo//uhuh//")
assert(T.gsub("alo.alo.uhuh.", "alo", "//") == "//.//.uhuh.")
assert(T.gsub("", "alo", "//") == "")
assert(T.gsub("...", ".", "/.") == "/././.")
assert(T.gsub("...", "...", "") == "")


print'OK'

