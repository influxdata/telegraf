print("testing functions and calls")

-- get the opportunity to test 'type' too ;)

assert(type(1<2) == 'boolean')
assert(type(true) == 'boolean' and type(false) == 'boolean')
assert(type(nil) == 'nil' and type(-3) == 'number' and type'x' == 'string' and
       type{} == 'table' and type(type) == 'function')

assert(type(assert) == type(print))
f = nil
function f (x) return a:x (x) end
assert(type(f) == 'function')


-- testing local-function recursion
fact = false
do
  local res = 1
  local function fact (n)
    if n==0 then return res
    else return n*fact(n-1)
    end
  end
  assert(fact(5) == 120)
end
assert(fact == false)

-- testing declarations
a = {i = 10}
self = 20
function a:x (x) return x+self.i end
function a.y (x) return x+self end

assert(a:x(1)+10 == a.y(1))

a.t = {i=-100}
a["t"].x = function (self, a,b) return self.i+a+b end

assert(a.t:x(2,3) == -95)

do
  local a = {x=0}
  function a:add (x) self.x, a.y = self.x+x, 20; return self end
  assert(a:add(10):add(20):add(30).x == 60 and a.y == 20)
end

local a = {b={c={}}}

function a.b.c.f1 (x) return x+1 end
function a.b.c:f2 (x,y) self[x] = y end
assert(a.b.c.f1(4) == 5)
a.b.c:f2('k', 12); assert(a.b.c.k == 12)

print('+')

t = nil   -- 'declare' t
function f(a,b,c) local d = 'a'; t={a,b,c,d} end

f(      -- this line change must be valid
  1,2)
assert(t[1] == 1 and t[2] == 2 and t[3] == nil and t[4] == 'a')
f(1,2,   -- this one too
      3,4)
assert(t[1] == 1 and t[2] == 2 and t[3] == 3 and t[4] == 'a')

function fat(x)
  if x <= 1 then return 1
  else return x*loadstring("return fat(" .. x-1 .. ")")()
  end
end

assert(loadstring "loadstring 'assert(fat(6)==720)' () ")()
a = loadstring('return fat(5), 3')
a,b = a()
assert(a == 120 and b == 3)
print('+')

function err_on_n (n)
  if n==0 then error(); exit(1);
  else err_on_n (n-1); exit(1);
  end
end

do
  function dummy (n)
    if n > 0 then
      assert(not pcall(err_on_n, n))
      dummy(n-1)
    end
  end
end

dummy(10)

function deep (n)
  if n>0 then deep(n-1) end
end
deep(10)
deep(200)

-- testing tail call
function deep (n) if n>0 then return deep(n-1) else return 101 end end
assert(deep(30000) == 101)
a = {}
function a:deep (n) if n>0 then return self:deep(n-1) else return 101 end end
assert(a:deep(30000) == 101)

print('+')


a = nil
(function (x) a=x end)(23)
assert(a == 23 and (function (x) return x*2 end)(20) == 40)


local x,y,z,a
a = {}; lim = 2000
for i=1, lim do a[i]=i end
assert(select(lim, unpack(a)) == lim and select('#', unpack(a)) == lim)
x = unpack(a)
assert(x == 1)
x = {unpack(a)}
assert(table.getn(x) == lim and x[1] == 1 and x[lim] == lim)
x = {unpack(a, lim-2)}
assert(table.getn(x) == 3 and x[1] == lim-2 and x[3] == lim)
x = {unpack(a, 10, 6)}
assert(next(x) == nil)   -- no elements
x = {unpack(a, 11, 10)}
assert(next(x) == nil)   -- no elements
x,y = unpack(a, 10, 10)
assert(x == 10 and y == nil)
x,y,z = unpack(a, 10, 11)
assert(x == 10 and y == 11 and z == nil)
a,x = unpack{1}
assert(a==1 and x==nil)
a,x = unpack({1,2}, 1, 1)
assert(a==1 and x==nil)


-- testing closures

-- fixed-point operator
Y = function (le)
      local function a (f)
        return le(function (x) return f(f)(x) end)
      end
      return a(a)
    end


-- non-recursive factorial

F = function (f)
      return function (n)
               if n == 0 then return 1
               else return n*f(n-1) end
             end
    end

fat = Y(F)

assert(fat(0) == 1 and fat(4) == 24 and Y(F)(5)==5*Y(F)(4))

local function g (z)
  local function f (a,b,c,d)
    return function (x,y) return a+b+c+d+a+x+y+z end
  end
  return f(z,z+1,z+2,z+3)
end

f = g(10)
assert(f(9, 16) == 10+11+12+13+10+9+16+10)

Y, F, f = nil
print('+')

-- testing multiple returns

function unlpack (t, i)
  i = i or 1
  if (i <= table.getn(t)) then
    return t[i], unlpack(t, i+1)
  end
end

function equaltab (t1, t2)
  assert(table.getn(t1) == table.getn(t2))
  for i,v1 in ipairs(t1) do
    assert(v1 == t2[i])
  end
end

local function pack (...)
  local x = {...}
  x.n = select('#', ...)
  return x
end

function f() return 1,2,30,4 end
function ret2 (a,b) return a,b end

local a,b,c,d = unlpack{1,2,3}
assert(a==1 and b==2 and c==3 and d==nil)
a = {1,2,3,4,false,10,'alo',false,assert}
equaltab(pack(unlpack(a)), a)
equaltab(pack(unlpack(a), -1), {1,-1})
a,b,c,d = ret2(f()), ret2(f())
assert(a==1 and b==1 and c==2 and d==nil)
a,b,c,d = unlpack(pack(ret2(f()), ret2(f())))
assert(a==1 and b==1 and c==2 and d==nil)
a,b,c,d = unlpack(pack(ret2(f()), (ret2(f()))))
assert(a==1 and b==1 and c==nil and d==nil)

a = ret2{ unlpack{1,2,3}, unlpack{3,2,1}, unlpack{"a", "b"}}
assert(a[1] == 1 and a[2] == 3 and a[3] == "a" and a[4] == "b")


-- testing calls with 'incorrect' arguments
rawget({}, "x", 1)
rawset({}, "x", 1, 2)
assert(math.sin(1,2) == math.sin(1))
table.sort({10,9,8,4,19,23,0,0}, function (a,b) return a<b end, "extra arg")


-- test for generic load
x = "-- a comment\n  x = 10 + \n23; \
     local a = function () x = 'hi' end; \
     return ''"
local i = 0
function read1 (x)
  return function ()
    collectgarbage()
    i=i+1
    return string.sub(x, i, i)
  end
end

a = assert(load(read1(x), "modname"))
assert(a() == "" and _G.x == 33)
assert(debug.getinfo(a).source == "modname")

-- x = string.dump(loadstring("x = 1; return x"))
-- i = 0
-- a = assert(load(read1(x)))
-- assert(a() == 1 and _G.x == 1)

-- i = 0
-- local a, b = load(read1("*a = 123"))
-- assert(not a and type(b) == "string" and i == 2)
-- 
-- a, b = load(function () error("hhi") end)
-- assert(not a and string.find(b, "hhi"))

-- test generic load with nested functions
i = 0
x = [[
  return function (x)
    return function (y)
     return function (z)
       return x+y+z
     end
   end
  end
]]

a = assert(load(read1(x)))
assert(a()(2)(3)(10) == 15)


-- test for dump/undump with upvalues
-- local a, b = 20, 30
-- x = loadstring(string.dump(function (x)
--   if x == "set" then a = 10+b; b = b+1 else
--   return a
--   end
-- end))
-- assert(x() == nil)
-- assert(debug.setupvalue(x, 1, "hi") == "a")
-- assert(x() == "hi")
-- assert(debug.setupvalue(x, 2, 13) == "b")
-- assert(not debug.setupvalue(x, 3, 10))   -- only 2 upvalues
-- x("set")
-- assert(x() == 23)
-- x("set")
-- assert(x() == 24)


-- test for bug in parameter adjustment
assert((function () return nil end)(4) == nil)
assert((function () local a; return a end)(4) == nil)
assert((function (a) return a end)() == nil)

print('OK')
return deep
