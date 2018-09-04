# testing special comment on first line

print ("testing lua.c options")

assert(os.execute() ~= 0)   -- machine has a system command

prog = os.tmpname()
otherprog = os.tmpname()
out = os.tmpname()

do
  local i = 0
  while arg[i] do i=i-1 end
  progname = '"'..arg[i+1]..'"'
end
print(progname)

local prepfile = function (s, p)
  p = p or prog
  io.output(p)
  io.write(s)
  assert(io.close())
end

function checkout (s)
  io.input(out)
  local t = io.read("*a")
  io.input():close()
  assert(os.remove(out))
  if s ~= t then print(string.format("'%s' - '%s'\n", s, t)) end
  assert(s == t)
  return t
end

function auxrun (...)
  local s = string.format(...)
  s = string.gsub(s, "lua", progname, 1)
  return os.execute(s)
end

function RUN (...)
  assert(auxrun(...) == 0)
end

function NoRun (...)
  print("\n(the next error is expected by the test)")
  assert(auxrun(...) ~= 0)
end

-- test 2 files
prepfile("print(1); a=2")
prepfile("print(a)", otherprog)
RUN("lua -l %s -l%s -lstring -l io %s > %s", prog, otherprog, otherprog, out)
checkout("1\n2\n2\n")

local a = [[
  assert(table.getn(arg) == 3 and arg[1] == 'a' and
         arg[2] == 'b' and arg[3] == 'c')
  assert(arg[-1] == '--' and arg[-2] == "-e " and arg[-3] == %s)
  assert(arg[4] == nil and arg[-4] == nil)
  local a, b, c = ...
  assert(... == 'a' and a == 'a' and b == 'b' and c == 'c')
]]
a = string.format(a, progname)
prepfile(a)
RUN('lua "-e " -- %s a b c', prog)

prepfile"assert(arg==nil)"
prepfile("assert(arg)", otherprog)
RUN("lua -l%s - < %s", prog, otherprog)

prepfile""
RUN("lua - < %s > %s", prog, out)
checkout("")

-- test many arguments
prepfile[[print(({...})[30])]]
RUN("lua %s %s > %s", prog, string.rep(" a", 30), out)
checkout("a\n")

RUN([[lua "-eprint(1)" -ea=3 -e "print(a)" > %s]], out)
checkout("1\n3\n")

prepfile[[
  print(
1, a
)
]]
RUN("lua - < %s > %s", prog, out)
checkout("1\tnil\n")

prepfile[[
= (6*2-6) -- ===
a 
= 10
print(a)
= a]]
RUN([[lua -e"_PROMPT='' _PROMPT2=''" -i < %s > %s]], prog, out)
checkout("6\n10\n10\n\n")

prepfile("a = [[b\nc\nd\ne]]\n=a")
print(prog)
RUN([[lua -e"_PROMPT='' _PROMPT2=''" -i < %s > %s]], prog, out)
checkout("b\nc\nd\ne\n\n")

prompt = "alo"
prepfile[[ --
a = 2
]]
RUN([[lua "-e_PROMPT='%s'" -i < %s > %s]], prompt, prog, out)
checkout(string.rep(prompt, 3).."\n")

s = [=[ -- 
function f ( x ) 
  local a = [[
xuxu
]]
  local b = "\
xuxu\n"
  if x == 11 then return 1 , 2 end  --[[ test multiple returns ]]
  return x + 1 
  --\\
end
=( f( 10 ) )
assert( a == b )
=f( 11 )  ]=]
s = string.gsub(s, ' ', '\n\n')
prepfile(s)
RUN([[lua -e"_PROMPT='' _PROMPT2=''" -i < %s > %s]], prog, out)
checkout("11\n1\t2\n\n")
  
prepfile[[#comment in 1st line without \n at the end]]
RUN("lua %s", prog)

prepfile("#comment with a binary file\n"..string.dump(loadstring("print(1)")))
RUN("lua %s > %s", prog, out)
checkout("1\n")

prepfile("#comment with a binary file\r\n"..string.dump(loadstring("print(1)")))
RUN("lua %s > %s", prog, out)
checkout("1\n")

-- close Lua with an open file
prepfile(string.format([[io.output(%q); io.write('alo')]], out))
RUN("lua %s", prog)
checkout('alo')

assert(os.remove(prog))
assert(os.remove(otherprog))
assert(not os.remove(out))

RUN("lua -v")

NoRun("lua -h")
NoRun("lua -e")
NoRun("lua -e a")
NoRun("lua -f")

print("OK")
