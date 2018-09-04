print('testing pattern matching')

function f(s, p)
  local i,e = string.find(s, p)
  if i then return string.sub(s, i, e) end
end

function f1(s, p)
  p = string.gsub(p, "%%([0-9])", function (s) return "%" .. (s+1) end)
  p = string.gsub(p, "^(^?)", "%1()", 1)
  p = string.gsub(p, "($?)$", "()%1", 1)
  local t = {string.match(s, p)}
  return string.sub(s, t[1], t[#t] - 1)
end

a,b = string.find('', '')    -- empty patterns are tricky
assert(a == 1 and b == 0);
a,b = string.find('alo', '')
assert(a == 1 and b == 0)
a,b = string.find('a\0o a\0o a\0o', 'a', 1)   -- first position
assert(a == 1 and b == 1)
a,b = string.find('a\0o a\0o a\0o', 'a\0o', 2)   -- starts in the midle
assert(a == 5 and b == 7)
a,b = string.find('a\0o a\0o a\0o', 'a\0o', 9)   -- starts in the midle
assert(a == 9 and b == 11)
a,b = string.find('a\0a\0a\0a\0\0ab', '\0ab', 2);  -- finds at the end
assert(a == 9 and b == 11);
a,b = string.find('a\0a\0a\0a\0\0ab', 'b')    -- last position
assert(a == 11 and b == 11)
assert(string.find('a\0a\0a\0a\0\0ab', 'b\0') == nil)   -- check ending
assert(string.find('', '\0') == nil)
assert(string.find('alo123alo', '12') == 4)
assert(string.find('alo123alo', '^12') == nil)

assert(f('aloALO', '%l*') == 'alo')
assert(f('aLo_ALO', '%a*') == 'aLo')

assert(f('aaab', 'a*') == 'aaa');
assert(f('aaa', '^.*$') == 'aaa');
assert(f('aaa', 'b*') == '');
assert(f('aaa', 'ab*a') == 'aa')
assert(f('aba', 'ab*a') == 'aba')
assert(f('aaab', 'a+') == 'aaa')
assert(f('aaa', '^.+$') == 'aaa')
assert(f('aaa', 'b+') == nil)
assert(f('aaa', 'ab+a') == nil)
assert(f('aba', 'ab+a') == 'aba')
assert(f('a$a', '.$') == 'a')
assert(f('a$a', '.%$') == 'a$')
assert(f('a$a', '.$.') == 'a$a')
assert(f('a$a', '$$') == nil)
assert(f('a$b', 'a$') == nil)
assert(f('a$a', '$') == '')
assert(f('', 'b*') == '')
assert(f('aaa', 'bb*') == nil)
assert(f('aaab', 'a-') == '')
assert(f('aaa', '^.-$') == 'aaa')
assert(f('aabaaabaaabaaaba', 'b.*b') == 'baaabaaabaaab')
assert(f('aabaaabaaabaaaba', 'b.-b') == 'baaab')
assert(f('alo xo', '.o$') == 'xo')
assert(f(' \n isto é assim', '%S%S*') == 'isto')
assert(f(' \n isto é assim', '%S*$') == 'assim')
assert(f(' \n isto é assim', '[a-z]*$') == 'assim')
assert(f('um caracter ? extra', '[^%sa-z]') == '?')
assert(f('', 'a?') == '')
assert(f('á', 'á?') == 'á')
assert(f('ábl', 'á?b?l?') == 'ábl')
assert(f('  ábl', 'á?b?l?') == '')
assert(f('aa', '^aa?a?a') == 'aa')
-- assert(f(']]]áb', '[^]]') == 'á')
assert(f(']]]áb', '[^%]]') == 'á')
assert(f("0alo alo", "%x*") == "0a")
assert(f("alo alo", "%C+") == "alo alo")
print('+')

assert(f1('alo alx 123 b\0o b\0o', '(..*) %1') == "b\0o b\0o")
assert(f1('axz123= 4= 4 34', '(.+)=(.*)=%2 %1') == '3= 4= 4 3')
assert(f1('=======', '^(=*)=%1$') == '=======')
assert(string.match('==========', '^([=]*)=%1$') == nil)

local function range (i, j)
  if i <= j then
    return i, range(i+1, j)
  end
end

local function range (i, j)
  local ret = {}
  for k=i, j do; table.insert(ret, k); end
  return unpack(ret)
end

local abc = string.char(range(0, 255));

assert(string.len(abc) == 256)

function strset (p)
  local res = {s=''}
  string.gsub(abc, p, function (c) res.s = res.s .. c end)
  return res.s
end;

assert(string.len(strset('[\200-\210]')) == 11)

assert(strset('[a-z]') == "abcdefghijklmnopqrstuvwxyz")
assert(strset('[a-z%d]') == strset('[%da-uu-z]'))
-- assert(strset('[a-]') == "-a")
assert(strset('[a%-]') == "-a")
assert(strset('[^%W]') == strset('[%w]'))
-- assert(strset('[]%%]') == '%]')
assert(strset('[%]%%]') == '%]')
assert(strset('[a%-z]') == '-az')
assert(strset('[%^%[%-a%]%-b]') == '-[]^ab')
assert(strset('%Z') == strset('[\1-\255]'))
assert(strset('.') == strset('[\1-\255%z]'))
print('+');

assert(string.match("alo xyzK", "(%w+)K") == "xyz")
assert(string.match("254 K", "(%d*)K") == "")
assert(string.match("alo ", "(%w*)$") == "")
assert(string.match("alo ", "(%w+)$") == nil)
assert(string.find("(álo)", "%(á") == 1)
local a, b, c, d, e = string.match("âlo alo", "^(((.).).* (%w*))$")
assert(a == 'âlo alo' and b == 'âl' and c == 'â' and d == 'alo' and e == nil)
a, b, c, d  = string.match('0123456789', '(.+(.?)())')
assert(a == '0123456789' and b == '' and c == 11 and d == nil)
print('+')

assert(string.gsub('ülo ülo', 'ü', 'x') == 'xlo xlo')
assert(string.gsub('alo úlo  ', ' +$', '') == 'alo úlo')  -- trim
assert(string.gsub('  alo alo  ', '^%s*(.-)%s*$', '%1') == 'alo alo')  -- double trim
assert(string.gsub('alo  alo  \n 123\n ', '%s+', ' ') == 'alo alo 123 ')
t = "abç d"
a, b = string.gsub(t, '(.)', '%1@')
assert('@'..a == string.gsub(t, '', '@') and b == 5)
a, b = string.gsub('abçd', '(.)', '%0@', 2)
assert(a == 'a@b@çd' and b == 2)
assert(string.gsub('alo alo', '()[al]', '%1') == '12o 56o')
assert(string.gsub("abc=xyz", "(%w*)(%p)(%w+)", "%3%2%1-%0") ==
              "xyz=abc-abc=xyz")
assert(string.gsub("abc", "%w", "%1%0") == "aabbcc")
assert(string.gsub("abc", "%w+", "%0%1") == "abcabc")
assert(string.gsub('áéí', '$', '\0óú') == 'áéí\0óú')
assert(string.gsub('', '^', 'r') == 'r')
assert(string.gsub('', '$', 'r') == 'r')
print('+')

assert(string.gsub("um (dois) tres (quatro)", "(%(%w+%))", string.upper) ==
            "um (DOIS) tres (QUATRO)")

do
  local function setglobal (n,v) rawset(_G, n, v) end
  string.gsub("a=roberto,roberto=a", "(%w+)=(%w%w*)", setglobal)
  assert(_G.a=="roberto" and _G.roberto=="a")
end

function f(a,b) return string.gsub(a,'.',b) end
assert(string.gsub("trocar tudo em |teste|b| é |beleza|al|", "|([^|]*)|([^|]*)|", f) ==
            "trocar tudo em bbbbb é alalalalalal")

local function dostring (s) return loadstring(s)() or "" end
assert(string.gsub("alo $a=1$ novamente $return a$", "$([^$]*)%$", dostring) ==
            "alo  novamente 1")

x = string.gsub("$x=string.gsub('alo', '.', string.upper)$ assim vai para $return x$",
         "$([^$]*)%$", dostring)
assert(x == ' assim vai para ALO')

t = {}
s = 'a alo jose  joao'
r = string.gsub(s, '()(%w+)()', function (a,w,b)
      assert(string.len(w) == b-a);
      t[a] = b-a;
    end)
assert(s == r and t[1] == 1 and t[3] == 3 and t[7] == 4 and t[13] == 4)


function isbalanced (s)
  return string.find(string.gsub(s, "%b()", ""), "[()]") == nil
end

assert(isbalanced("(9 ((8))(\0) 7) \0\0 a b ()(c)() a"))
assert(not isbalanced("(9 ((8) 7) a b (\0 c) a"))
assert(string.gsub("alo 'oi' alo", "%b''", '"') == 'alo " alo')


local t = {"apple", "orange", "lime"; n=0}
assert(string.gsub("x and x and x", "x", function () t.n=t.n+1; return t[t.n] end)
        == "apple and orange and lime")

t = {n=0}
string.gsub("first second word", "%w%w*", function (w) t.n=t.n+1; t[t.n] = w end)
assert(t[1] == "first" and t[2] == "second" and t[3] == "word" and t.n == 3)

t = {n=0}
assert(string.gsub("first second word", "%w+",
         function (w) t.n=t.n+1; t[t.n] = w end, 2) == "first second word")
assert(t[1] == "first" and t[2] == "second" and t[3] == nil)

assert(not pcall(string.gsub, "alo", "(.", print))
assert(not pcall(string.gsub, "alo", ".)", print))
assert(not pcall(string.gsub, "alo", "(.", {}))
assert(not pcall(string.gsub, "alo", "(.)", "%2"))
assert(not pcall(string.gsub, "alo", "(%1)", "a"))
assert(not pcall(string.gsub, "alo", "(%0)", "a"))

-- big strings
local a = string.rep('a', 300000)
assert(string.find(a, '^a*.?$'))
assert(not string.find(a, '^a*.?b$'))
assert(string.find(a, '^a-.?$'))

-- deep nest of gsubs
function rev (s)
  return string.gsub(s, "(.)(.+)", function (c,s1) return rev(s1)..c end)
end

local x = string.rep('012345', 10)
assert(rev(rev(x)) == x)


-- gsub with tables
assert(string.gsub("alo alo", ".", {}) == "alo alo")
assert(string.gsub("alo alo", "(.)", {a="AA", l=""}) == "AAo AAo")
assert(string.gsub("alo alo", "(.).", {a="AA", l="K"}) == "AAo AAo")
assert(string.gsub("alo alo", "((.)(.?))", {al="AA", o=false}) == "AAo AAo")

assert(string.gsub("alo alo", "().", {2,5,6}) == "256 alo")

t = {}; setmetatable(t, {__index = function (t,s) return string.upper(s) end})
assert(string.gsub("a alo b hi", "%w%w+", t) == "a ALO b HI")


-- tests for gmatch
assert(string.gfind == string.gmatch)
local a = 0
for i in string.gmatch('abcde', '()') do assert(i == a+1); a=i end
assert(a==6)

t = {n=0}
for w in string.gmatch("first second word", "%w+") do
      t.n=t.n+1; t[t.n] = w
end
assert(t[1] == "first" and t[2] == "second" and t[3] == "word")

t = {3, 6, 9}
for i in string.gmatch ("xuxx uu ppar r", "()(.)%2") do
  assert(i == table.remove(t, 1))
end
assert(table.getn(t) == 0)

t = {}
for i,j in string.gmatch("13 14 10 = 11, 15= 16, 22=23", "(%d+)%s*=%s*(%d+)") do
  t[i] = j
end
a = 0
for k,v in pairs(t) do assert(k+1 == v+0); a=a+1 end
assert(a == 3)


-- tests for `%f' (`frontiers')

-- assert(string.gsub("aaa aa a aaa a", "%f[%w]a", "x") == "xaa xa x xaa x")
-- assert(string.gsub("[[]] [][] [[[[", "%f[[].", "x") == "x[]] x]x] x[[[")
-- assert(string.gsub("01abc45de3", "%f[%d]", ".") == ".01abc.45de.3")
-- assert(string.gsub("01abc45 de3x", "%f[%D]%w", ".") == "01.bc45 de3.")
-- assert(string.gsub("function", "%f[\1-\255]%w", ".") == ".unction")
-- assert(string.gsub("function", "%f[^\1-\255]", ".") == "function.")
-- 
-- local i, e = string.find(" alo aalo allo", "%f[%S].-%f[%s].-%f[%S]")
-- assert(i == 2 and e == 5)
-- local k = string.match(" alo aalo allo", "%f[%S](.-%f[%s].-%f[%S])")
-- assert(k == 'alo ')
-- 
-- local a = {1, 5, 9, 14, 17,}
-- for k in string.gmatch("alo alo th02 is 1hat", "()%f[%w%d]") do
--   assert(table.remove(a, 1) == k)
-- end
-- assert(table.getn(a) == 0)


print('OK')
