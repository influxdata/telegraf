This tarball contains the official test scripts for Lua 5.1.
Unlike Lua itself, these tests do not aim portability, small footprint,
or easy of use. (Their main goal is to try to crash Lua.) They are not
intended for general use. You are wellcome to use them, but expect to
have to "dirt your hands".

The tarball should expand in the following contents:
  - several .lua scripts with the tests
  - a main "all.lua" Lua script that invokes all the other scripts
  - a subdirectory "libs" with an empty subdirectory "libs/P1",
    to be used by the scripts
  - a subdirectory "etc" with some extra files

To run the tests, do as follows:

- go to the test directory

- set LUA_PATH to "?;./?.lua" (or, better yet, set LUA_PATH to "./?.lua;;"
  and LUA_INIT to "package.path = '?;'..package.path")

- run "lua all.lua"


--------------------------------------------
Internal tests
--------------------------------------------

Some tests need a special library, "testC", that gives access to
several internal structures in Lua.
This library is only available when Lua is compiled in debug mode.
The scripts automatically detect its absence and skip those tests.

If you want to run these tests, move etc/ltests.c and etc/ltests.h to
the directory with the source Lua files, and recompile Lua with
the option -DLUA_USER_H='"ltests.h"' (or its equivalent to define
LUA_USER_H as the string "ltests.h", including the quotes). This
option not only adds the testC library, but it adds several other
internal tests as well. After the recompilation, run the tests
as before.


