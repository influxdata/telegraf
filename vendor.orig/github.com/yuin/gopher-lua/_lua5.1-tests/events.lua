print('testing metatables')

X = 20; B = 30

setfenv(1, setmetatable({}, {__index=_G}))

collectgarbage()

X = X+10
assert(X == 30 and _G.X == 20)
B = false
assert(B == false)
B = nil
assert(B == 30)

assert(getmetatable{} == nil)
assert(getmetatable(4) == nil)
assert(getmetatable(nil) == nil)
a={}; setmetatable(a, {__metatable = "xuxu",
                    __tostring=function(x) return x.name end})
assert(getmetatable(a) == "xuxu")
assert(tostring(a) == nil)
-- cannot change a protected metatable
assert(pcall(setmetatable, a, {}) == false)
a.name = "gororoba"
assert(tostring(a) == "gororoba")

local a, t = {10,20,30; x="10", y="20"}, {}
assert(setmetatable(a,t) == a)
assert(getmetatable(a) == t)
assert(setmetatable(a,nil) == a)
assert(getmetatable(a) == nil)
assert(setmetatable(a,t) == a)


function f (t, i, e)
  assert(not e)
  local p = rawget(t, "parent")
  return (p and p[i]+3), "dummy return"
end

t.__index = f

a.parent = {z=25, x=12, [4] = 24}
assert(a[1] == 10 and a.z == 28 and a[4] == 27 and a.x == "10")

collectgarbage()

a = setmetatable({}, t)
function f(t, i, v) rawset(t, i, v-3) end
t.__newindex = f
a[1] = 30; a.x = "101"; a[5] = 200
assert(a[1] == 27 and a.x == 98 and a[5] == 197)


local c = {}
a = setmetatable({}, t)
t.__newindex = c
a[1] = 10; a[2] = 20; a[3] = 90
assert(c[1] == 10 and c[2] == 20 and c[3] == 90)


do
  local a;
  a = setmetatable({}, {__index = setmetatable({},
                     {__index = setmetatable({},
                     {__index = function (_,n) return a[n-3]+4, "lixo" end})})})
  a[0] = 20
  for i=0,10 do
    assert(a[i*3] == 20 + i*4)
  end
end


do  -- newindex
  local foi
  local a = {}
  for i=1,10 do a[i] = 0; a['a'..i] = 0; end
  setmetatable(a, {__newindex = function (t,k,v) foi=true; rawset(t,k,v) end})
  foi = false; a[1]=0; assert(not foi)
  foi = false; a['a1']=0; assert(not foi)
  foi = false; a['a11']=0; assert(foi)
  foi = false; a[11]=0; assert(foi)
  foi = false; a[1]=nil; assert(not foi)
  foi = false; a[1]=nil; assert(foi)
end


function f (t, ...) return t, {...} end
t.__call = f

do
  local x,y = a(unpack{'a', 1})
  assert(x==a and y[1]=='a' and y[2]==1 and y[3]==nil)
  x,y = a()
  assert(x==a and y[1]==nil)
end


local b = setmetatable({}, t)
setmetatable(b,t)

function f(op)
  return function (...) cap = {[0] = op, ...} ; return (...) end
end
t.__add = f("add")
t.__sub = f("sub")
t.__mul = f("mul")
t.__div = f("div")
t.__mod = f("mod")
t.__unm = f("unm")
t.__pow = f("pow")

assert(b+5 == b)
assert(cap[0] == "add" and cap[1] == b and cap[2] == 5 and cap[3]==nil)
assert(b+'5' == b)
assert(cap[0] == "add" and cap[1] == b and cap[2] == '5' and cap[3]==nil)
assert(5+b == 5)
assert(cap[0] == "add" and cap[1] == 5 and cap[2] == b and cap[3]==nil)
assert('5'+b == '5')
assert(cap[0] == "add" and cap[1] == '5' and cap[2] == b and cap[3]==nil)
b=b-3; assert(getmetatable(b) == t)
assert(5-a == 5)
assert(cap[0] == "sub" and cap[1] == 5 and cap[2] == a and cap[3]==nil)
assert('5'-a == '5')
assert(cap[0] == "sub" and cap[1] == '5' and cap[2] == a and cap[3]==nil)
assert(a*a == a)
assert(cap[0] == "mul" and cap[1] == a and cap[2] == a and cap[3]==nil)
assert(a/0 == a)
assert(cap[0] == "div" and cap[1] == a and cap[2] == 0 and cap[3]==nil)
assert(a%2 == a)
assert(cap[0] == "mod" and cap[1] == a and cap[2] == 2 and cap[3]==nil)
assert(-a == a)
assert(cap[0] == "unm" and cap[1] == a)
assert(a^4 == a)
assert(cap[0] == "pow" and cap[1] == a and cap[2] == 4 and cap[3]==nil)
assert(a^'4' == a)
assert(cap[0] == "pow" and cap[1] == a and cap[2] == '4' and cap[3]==nil)
assert(4^a == 4)
assert(cap[0] == "pow" and cap[1] == 4 and cap[2] == a and cap[3]==nil)
assert('4'^a == '4')
assert(cap[0] == "pow" and cap[1] == '4' and cap[2] == a and cap[3]==nil)


t = {}
t.__lt = function (a,b,c)
  collectgarbage()
  assert(c == nil)
  if type(a) == 'table' then a = a.x end
  if type(b) == 'table' then b = b.x end
 return a<b, "dummy"
end

function Op(x) return setmetatable({x=x}, t) end

local function test ()
  assert(not(Op(1)<Op(1)) and (Op(1)<Op(2)) and not(Op(2)<Op(1)))
  assert(not(Op('a')<Op('a')) and (Op('a')<Op('b')) and not(Op('b')<Op('a')))
  assert((Op(1)<=Op(1)) and (Op(1)<=Op(2)) and not(Op(2)<=Op(1)))
  assert((Op('a')<=Op('a')) and (Op('a')<=Op('b')) and not(Op('b')<=Op('a')))
  assert(not(Op(1)>Op(1)) and not(Op(1)>Op(2)) and (Op(2)>Op(1)))
  assert(not(Op('a')>Op('a')) and not(Op('a')>Op('b')) and (Op('b')>Op('a')))
  assert((Op(1)>=Op(1)) and not(Op(1)>=Op(2)) and (Op(2)>=Op(1)))
  assert((Op('a')>=Op('a')) and not(Op('a')>=Op('b')) and (Op('b')>=Op('a')))
end

test()

t.__le = function (a,b,c)
  assert(c == nil)
  if type(a) == 'table' then a = a.x end
  if type(b) == 'table' then b = b.x end
 return a<=b, "dummy"
end

test()  -- retest comparisons, now using both `lt' and `le'


-- test `partial order'

local function Set(x)
  local y = {}
  for _,k in pairs(x) do y[k] = 1 end
  return setmetatable(y, t)
end

t.__lt = function (a,b)
  for k in pairs(a) do
    if not b[k] then return false end
    b[k] = nil
  end
  return next(b) ~= nil
end

t.__le = nil

assert(Set{1,2,3} < Set{1,2,3,4})
assert(not(Set{1,2,3,4} < Set{1,2,3,4}))
assert((Set{1,2,3,4} <= Set{1,2,3,4}))
assert((Set{1,2,3,4} >= Set{1,2,3,4}))
assert((Set{1,3} <= Set{3,5}))   -- wrong!! model needs a `le' method ;-)

t.__le = function (a,b)
  for k in pairs(a) do
    if not b[k] then return false end
  end
  return true
end

assert(not (Set{1,3} <= Set{3,5}))   -- now its OK!
assert(not(Set{1,3} <= Set{3,5}))
assert(not(Set{1,3} >= Set{3,5}))

t.__eq = function (a,b)
  for k in pairs(a) do
    if not b[k] then return false end
    b[k] = nil
  end
  return next(b) == nil
end

local s = Set{1,3,5}
assert(s == Set{3,5,1})
assert(not rawequal(s, Set{3,5,1}))
assert(rawequal(s, s))
assert(Set{1,3,5,1} == Set{3,5,1})
assert(Set{1,3,5} ~= Set{3,5,1,6})
t[Set{1,3,5}] = 1
assert(t[Set{1,3,5}] == nil)   -- `__eq' is not valid for table accesses


t.__concat = function (a,b,c)
  assert(c == nil)
  if type(a) == 'table' then a = a.val end
  if type(b) == 'table' then b = b.val end
  if A then return a..b
  else
    return setmetatable({val=a..b}, t)
  end
end

c = {val="c"}; setmetatable(c, t)
d = {val="d"}; setmetatable(d, t)

A = true
assert(c..d == 'cd')
assert(0 .."a".."b"..c..d.."e".."f"..(5+3).."g" == "0abcdef8g")

A = false
x = c..d
assert(getmetatable(x) == t and x.val == 'cd')
x = 0 .."a".."b"..c..d.."e".."f".."g"
assert(x.val == "0abcdefg")


-- test comparison compatibilities
local t1, t2, c, d
t1 = {};  c = {}; setmetatable(c, t1)
d = {}
t1.__eq = function () return true end
t1.__lt = function () return true end
assert(c ~= d and not pcall(function () return c < d end))
setmetatable(d, t1)
assert(c == d and c < d and not(d <= c))
t2 = {}
t2.__eq = t1.__eq
t2.__lt = t1.__lt
setmetatable(d, t2)
assert(c == d and c < d and not(d <= c))



-- test for several levels of calls
local i
local tt = {
  __call = function (t, ...)
    i = i+1
    if t.f then return t.f(...)
    else return {...}
    end
  end
}

local a = setmetatable({}, tt)
local b = setmetatable({f=a}, tt)
local c = setmetatable({f=b}, tt)

i = 0
x = c(3,4,5)
assert(i == 3 and x[1] == 3 and x[3] == 5)


assert(_G.X == 20)
assert(_G == getfenv(0))

print'+'

local _g = _G
setfenv(1, setmetatable({}, {__index=function (_,k) return _g[k] end}))

--[[
-- testing proxies
assert(getmetatable(newproxy()) == nil)
assert(getmetatable(newproxy(false)) == nil)

local u = newproxy(true)

getmetatable(u).__newindex = function (u,k,v)
  getmetatable(u)[k] = v
end

getmetatable(u).__index = function (u,k)
  return getmetatable(u)[k]
end

for i=1,10 do u[i] = i end
for i=1,10 do assert(u[i] == i) end

local k = newproxy(u)
assert(getmetatable(k) == getmetatable(u))


a = {}
rawset(a, "x", 1, 2, 3)
assert(a.x == 1 and rawget(a, "x", 3) == 1)

print '+'
--]]

-- testing metatables for basic types
mt = {}
debug.setmetatable(10, mt)
assert(getmetatable(-2) == mt)
mt.__index = function (a,b) return a+b end
assert((10)[3] == 13)
assert((10)["3"] == 13)
debug.setmetatable(23, nil)
assert(getmetatable(-2) == nil)

debug.setmetatable(true, mt)
assert(getmetatable(false) == mt)
mt.__index = function (a,b) return a or b end
assert((true)[false] == true)
assert((false)[false] == false)
debug.setmetatable(false, nil)
assert(getmetatable(true) == nil)

debug.setmetatable(nil, mt)
assert(getmetatable(nil) == mt)
mt.__add = function (a,b) return (a or 0) + (b or 0) end
assert(10 + nil == 10)
assert(nil + 23 == 23)
assert(nil + nil == 0)
debug.setmetatable(nil, nil)
assert(getmetatable(nil) == nil)

debug.setmetatable(nil, {})


print 'OK'

return 12
