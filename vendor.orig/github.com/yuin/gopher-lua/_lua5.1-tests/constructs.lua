print "testing syntax"

-- testing priorities

assert(2^3^2 == 2^(3^2));
assert(2^3*4 == (2^3)*4);
assert(2^-2 == 1/4 and -2^- -2 == - - -4);
assert(not nil and 2 and not(2>3 or 3<2));
assert(-3-1-5 == 0+0-9);
assert(-2^2 == -4 and (-2)^2 == 4 and 2*2-3-1 == 0);
assert(2*1+3/3 == 3 and 1+2 .. 3*1 == "33");
assert(not(2+1 > 3*1) and "a".."b" > "a");

assert(not ((true or false) and nil))
assert(      true or false  and nil)

local a,b = 1,nil;
assert(-(1 or 2) == -1 and (1 and 2)+(-1.25 or -4) == 0.75);
x = ((b or a)+1 == 2 and (10 or a)+1 == 11); assert(x);
x = (((2<3) or 1) == true and (2<3 and 4) == 4); assert(x);

x,y=1,2;
assert((x>y) and x or y == 2);
x,y=2,1;
assert((x>y) and x or y == 2);

assert(1234567890 == tonumber('1234567890') and 1234567890+1 == 1234567891)


-- silly loops
repeat until 1; repeat until true;
while false do end; while nil do end;

do  -- test old bug (first name could not be an `upvalue')
 local a; function f(x) x={a=1}; x={x=1}; x={G=1} end
end

function f (i)
  if type(i) ~= 'number' then return i,'jojo'; end;
  if i > 0 then return i, f(i-1); end;
end

x = {f(3), f(5), f(10);};
assert(x[1] == 3 and x[2] == 5 and x[3] == 10 and x[4] == 9 and x[12] == 1);
assert(x[nil] == nil)
x = {f'alo', f'xixi', nil};
assert(x[1] == 'alo' and x[2] == 'xixi' and x[3] == nil);
x = {f'alo'..'xixi'};
assert(x[1] == 'aloxixi')
x = {f{}}
assert(x[2] == 'jojo' and type(x[1]) == 'table')


local f = function (i)
  if i < 10 then return 'a';
  elseif i < 20 then return 'b';
  elseif i < 30 then return 'c';
  end;
end

assert(f(3) == 'a' and f(12) == 'b' and f(26) == 'c' and f(100) == nil)

for i=1,1000 do break; end;
n=100;
i=3;
t = {};
a=nil
while not a do
  a=0; for i=1,n do for i=i,1,-1 do a=a+1; t[i]=1; end; end;
end
assert(a == n*(n+1)/2 and i==3);
assert(t[1] and t[n] and not t[0] and not t[n+1])

function f(b)
  local x = 1;
  repeat
    local a;
    if b==1 then local b=1; x=10; break
    elseif b==2 then x=20; break;
    elseif b==3 then x=30;
    else local a,b,c,d=math.sin(1); x=x+1;
    end
  until x>=12;
  return x;
end;

assert(f(1) == 10 and f(2) == 20 and f(3) == 30 and f(4)==12)


local f = function (i)
  if i < 10 then return 'a'
  elseif i < 20 then return 'b'
  elseif i < 30 then return 'c'
  else return 8
  end
end

assert(f(3) == 'a' and f(12) == 'b' and f(26) == 'c' and f(100) == 8)

local a, b = nil, 23
x = {f(100)*2+3 or a, a or b+2}
assert(x[1] == 19 and x[2] == 25)
x = {f=2+3 or a, a = b+2}
assert(x.f == 5 and x.a == 25)

a={y=1}
x = {a.y}
assert(x[1] == 1)

function f(i)
  while 1 do
    if i>0 then i=i-1;
    else return; end;
  end;
end;

function g(i)
  while 1 do
    if i>0 then i=i-1
    else return end
  end
end

f(10); g(10);

do
  function f () return 1,2,3; end
  local a, b, c = f();
  assert(a==1 and b==2 and c==3)
  a, b, c = (f());
  assert(a==1 and b==nil and c==nil)
end

local a,b = 3 and f();
assert(a==1 and b==nil)

function g() f(); return; end;
assert(g() == nil)
function g() return nil or f() end
a,b = g()
assert(a==1 and b==nil)

print'+';


f = [[
return function ( a , b , c , d , e )
  local x = a >= b or c or ( d and e ) or nil
  return x
end , { a = 1 , b = 2 >= 1 , } or { 1 };
]]
f = string.gsub(f, "%s+", "\n");   -- force a SETLINE between opcodes
f,a = loadstring(f)();
assert(a.a == 1 and a.b)

function g (a,b,c,d,e)
  if not (a>=b or c or d and e or nil) then return 0; else return 1; end;
end

function h (a,b,c,d,e)
  while (a>=b or c or (d and e) or nil) do return 1; end;
  return 0;
end;

assert(f(2,1) == true and g(2,1) == 1 and h(2,1) == 1)
assert(f(1,2,'a') == 'a' and g(1,2,'a') == 1 and h(1,2,'a') == 1)
assert(f(1,2,'a')
~=          -- force SETLINE before nil
nil, "")
assert(f(1,2,'a') == 'a' and g(1,2,'a') == 1 and h(1,2,'a') == 1)
assert(f(1,2,nil,1,'x') == 'x' and g(1,2,nil,1,'x') == 1 and
                                   h(1,2,nil,1,'x') == 1)
assert(f(1,2,nil,nil,'x') == nil and g(1,2,nil,nil,'x') == 0 and
                                     h(1,2,nil,nil,'x') == 0)
assert(f(1,2,nil,1,nil) == nil and g(1,2,nil,1,nil) == 0 and
                                   h(1,2,nil,1,nil) == 0)

assert(1 and 2<3 == true and 2<3 and 'a'<'b' == true)
x = 2<3 and not 3; assert(x==false)
x = 2<1 or (2>1 and 'a'); assert(x=='a')


do
  local a; if nil then a=1; else a=2; end;    -- this nil comes as PUSHNIL 2
  assert(a==2)
end

function F(a)
  assert(debug.getinfo(1, "n").name == 'F')
  return a,2,3
end

a,b = F(1)~=nil; assert(a == true and b == nil);
a,b = F(nil)==nil; assert(a == true and b == nil)

----------------------------------------------------------------
-- creates all combinations of 
-- [not] ([not] arg op [not] (arg op [not] arg ))
-- and tests each one

function ID(x) return x end

function f(t, i)
  local b = t.n
  local res = math.mod(math.floor(i/c), b)+1
  c = c*b
  return t[res]
end

local arg = {" ( 1 < 2 ) ", " ( 1 >= 2 ) ", " F ( ) ", "  nil "; n=4}

local op = {" and ", " or ", " == ", " ~= "; n=4}

local neg = {" ", " not "; n=2}

local i = 0
repeat
  c = 1
  local s = f(neg, i)..'ID('..f(neg, i)..f(arg, i)..f(op, i)..
            f(neg, i)..'ID('..f(arg, i)..f(op, i)..f(neg, i)..f(arg, i)..'))'
  local s1 = string.gsub(s, 'ID', '')
  K,X,NX,WX1,WX2 = nil
  s = string.format([[
      local a = %s
      local b = not %s
      K = b
      local xxx; 
      if %s then X = a  else X = b end
      if %s then NX = b  else NX = a end
      while %s do WX1 = a; break end
      while %s do WX2 = a; break end
      repeat if (%s) then break end; assert(b)  until not(%s)
  ]], s1, s, s1, s, s1, s, s1, s, s)
  assert(loadstring(s))()
  assert(X and not NX and not WX1 == K and not WX2 == K)
  if math.mod(i,4000) == 0 then print('+') end
  i = i+1
until i==c

print'OK'
