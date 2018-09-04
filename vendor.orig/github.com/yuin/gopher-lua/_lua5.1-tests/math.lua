print("testing numbers and math lib")

do
  local a,b,c = "2", " 3e0 ", " 10  "
  assert(a+b == 5 and -b == -3 and b+"2" == 5 and "10"-c == 0)
  assert(type(a) == 'string' and type(b) == 'string' and type(c) == 'string')
  assert(a == "2" and b == " 3e0 " and c == " 10  " and -c == -"  10 ")
  assert(c%a == 0 and a^b == 8)
end


do
  local a,b = math.modf(3.5)
  assert(a == 3 and b == 0.5)
  assert(math.huge > 10e30)
  assert(-math.huge < -10e30)
end

function f(...)
  if select('#', ...) == 1 then
    return (...)
  else
    return "***"
  end
end

assert(tonumber{} == nil)
assert(tonumber'+0.01' == 1/100 and tonumber'+.01' == 0.01 and
       tonumber'.01' == 0.01    and tonumber'-1.' == -1 and
       tonumber'+1.' == 1)
assert(tonumber'+ 0.01' == nil and tonumber'+.e1' == nil and
       tonumber'1e' == nil     and tonumber'1.0e+' == nil and
       tonumber'.' == nil)
assert(tonumber('-12') == -10-2)
assert(tonumber('-1.2e2') == - - -120)
assert(f(tonumber('1  a')) == nil)
assert(f(tonumber('e1')) == nil)
assert(f(tonumber('e  1')) == nil)
assert(f(tonumber(' 3.4.5 ')) == nil)
assert(f(tonumber('')) == nil)
assert(f(tonumber('', 8)) == nil)
assert(f(tonumber('  ')) == nil)
assert(f(tonumber('  ', 9)) == nil)
assert(f(tonumber('99', 8)) == nil)
assert(tonumber('  1010  ', 2) == 10)
assert(tonumber('10', 36) == 36)
assert(tonumber('\n  -10  \n', 36) == -36)
assert(tonumber('-fFfa', 16) == -(10+(16*(15+(16*(15+(16*15)))))))
assert(tonumber('fFfa', 15) == nil)
assert(tonumber(string.rep('1', 42), 2) + 1 == 2^42)
assert(tonumber(string.rep('1', 32), 2) + 1 == 2^32)
assert(tonumber('-fffffFFFFF', 16)-1 == -2^40)
assert(tonumber('ffffFFFF', 16)+1 == 2^32)
assert(tonumber('0xF') == 15)

assert(1.1 == 1.+.1)
print(100.0, 1E2, .01)
assert(100.0 == 1E2 and .01 == 1e-2)
assert(1111111111111111-1111111111111110== 1000.00e-03)
--     1234567890123456
assert(1.1 == '1.'+'.1')
assert('1111111111111111'-'1111111111111110' == tonumber"  +0.001e+3 \n\t")

function eq (a,b,limit)
  if not limit then limit = 10E-10 end
  return math.abs(a-b) <= limit
end

assert(0.1e-30 > 0.9E-31 and 0.9E30 < 0.1e31)

assert(0.123456 > 0.123455)

assert(tonumber('+1.23E30') == 1.23*10^30)

-- testing order operators
assert(not(1<1) and (1<2) and not(2<1))
assert(not('a'<'a') and ('a'<'b') and not('b'<'a'))
assert((1<=1) and (1<=2) and not(2<=1))
assert(('a'<='a') and ('a'<='b') and not('b'<='a'))
assert(not(1>1) and not(1>2) and (2>1))
assert(not('a'>'a') and not('a'>'b') and ('b'>'a'))
assert((1>=1) and not(1>=2) and (2>=1))
assert(('a'>='a') and not('a'>='b') and ('b'>='a'))

-- testing mod operator
assert(-4%3 == 2)
assert(4%-3 == -2)
assert(math.pi - math.pi % 1 == 3)
assert(math.pi - math.pi % 0.001 == 3.141)

local function testbit(a, n)
  return a/2^n % 2 >= 1
end

assert(eq(math.sin(-9.8)^2 + math.cos(-9.8)^2, 1))
assert(eq(math.tan(math.pi/4), 1))
assert(eq(math.sin(math.pi/2), 1) and eq(math.cos(math.pi/2), 0))
assert(eq(math.atan(1), math.pi/4) and eq(math.acos(0), math.pi/2) and
       eq(math.asin(1), math.pi/2))
assert(eq(math.deg(math.pi/2), 90) and eq(math.rad(90), math.pi/2))
assert(math.abs(-10) == 10)
assert(eq(math.atan2(1,0), math.pi/2))
assert(math.ceil(4.5) == 5.0)
assert(math.floor(4.5) == 4.0)
assert(math.mod(10,3) == 1)
assert(eq(math.sqrt(10)^2, 10))
assert(eq(math.log10(2), math.log(2)/math.log(10)))
assert(eq(math.exp(0), 1))
assert(eq(math.sin(10), math.sin(10%(2*math.pi))))
local v,e = math.frexp(math.pi)
assert(eq(math.ldexp(v,e), math.pi))

assert(eq(math.tanh(3.5), math.sinh(3.5)/math.cosh(3.5)))

assert(tonumber(' 1.3e-2 ') == 1.3e-2)
assert(tonumber(' -1.00000000000001 ') == -1.00000000000001)

-- testing constant limits
-- 2^23 = 8388608
assert(8388609 + -8388609 == 0)
assert(8388608 + -8388608 == 0)
assert(8388607 + -8388607 == 0)

if rawget(_G, "_soft") then return end

f = io.tmpfile()
assert(f)
f:write("a = {")
i = 1
repeat
  f:write("{", math.sin(i), ", ", math.cos(i), ", ", i/3, "},\n")
  i=i+1
until i > 1000
f:write("}")
f:seek("set", 0)
assert(loadstring(f:read('*a')))()
assert(f:close())

assert(eq(a[300][1], math.sin(300)))
assert(eq(a[600][1], math.sin(600)))
assert(eq(a[500][2], math.cos(500)))
assert(eq(a[800][2], math.cos(800)))
assert(eq(a[200][3], 200/3))
assert(eq(a[1000][3], 1000/3, 0.001))
print('+')

do   -- testing NaN
  local NaN = 10e500 - 10e400
  assert(NaN ~= NaN)
  assert(not (NaN < NaN))
  assert(not (NaN <= NaN))
  assert(not (NaN > NaN))
  assert(not (NaN >= NaN))
  assert(not (0 < NaN))
  assert(not (NaN < 0))
  local a = {}
  assert(not pcall(function () a[NaN] = 1 end))
  assert(a[NaN] == nil)
  a[1] = 1
  assert(not pcall(function () a[NaN] = 1 end))
  assert(a[NaN] == nil)
end

require "checktable"
stat(a)

a = nil

-- testing implicit convertions

local a,b = '10', '20'
assert(a*b == 200 and a+b == 30 and a-b == -10 and a/b == 0.5 and -b == -20)
assert(a == '10' and b == '20')


math.randomseed(0)

local i = 0
local Max = 0
local Min = 2
repeat
  local t = math.random()
  Max = math.max(Max, t)
  Min = math.min(Min, t)
  i=i+1
  flag = eq(Max, 1, 0.001) and eq(Min, 0, 0.001)
until flag or i>10000
assert(0 <= Min and Max<1)
assert(flag);

for i=1,10 do
  local t = math.random(5)
  assert(1 <= t and t <= 5)
end

i = 0
Max = -200
Min = 200
repeat
  local t = math.random(-10,0)
  Max = math.max(Max, t)
  Min = math.min(Min, t)
  i=i+1
  flag = (Max == 0 and Min == -10)
until flag or i>10000
assert(-10 <= Min and Max<=0)
assert(flag);


print('OK')
