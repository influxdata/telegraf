local osname = "linux"
if string.find(os.getenv("OS") or "", "Windows") then
  osname = "windows"
end

if osname == "linux" then
  -- travis ci failed to start date command?
  -- assert(os.execute("date") == 0)
  assert(os.execute("date -a") == 1)
else
  assert(os.execute("date /T") == 0)
  assert(os.execute("md") == 1)
end

assert(os.getenv("PATH") ~= "")
assert(os.getenv("_____GLUATEST______") == nil)
assert(os.setenv("_____GLUATEST______", "1"))
assert(os.getenv("_____GLUATEST______") == "1")
