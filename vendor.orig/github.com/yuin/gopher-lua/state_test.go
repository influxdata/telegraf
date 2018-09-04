package lua

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCallStackOverflow(t *testing.T) {
	L := NewState(Options{
		CallStackSize: 3,
	})
	defer L.Close()
	errorIfScriptNotFail(t, L, `
    local function a()
    end
    local function b()
      a()
    end
    local function c()
      print(_printregs())
      b()
    end
    c()
    `, "stack overflow")
}

func TestSkipOpenLibs(t *testing.T) {
	L := NewState(Options{SkipOpenLibs: true})
	defer L.Close()
	errorIfScriptNotFail(t, L, `print("")`,
		"attempt to call a non-function object")
	L2 := NewState()
	defer L2.Close()
	errorIfScriptFail(t, L2, `print("")`)
}

func TestGetAndReplace(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LString("a"))
	L.Replace(1, LString("b"))
	L.Replace(0, LString("c"))
	errorIfNotEqual(t, LNil, L.Get(0))
	errorIfNotEqual(t, LNil, L.Get(-10))
	errorIfNotEqual(t, L.Env, L.Get(EnvironIndex))
	errorIfNotEqual(t, LString("b"), L.Get(1))
	L.Push(LString("c"))
	L.Push(LString("d"))
	L.Replace(-2, LString("e"))
	errorIfNotEqual(t, LString("e"), L.Get(-2))
	registry := L.NewTable()
	L.Replace(RegistryIndex, registry)
	L.G.Registry = registry
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Replace(RegistryIndex, LNil)
		return 0
	}, "registry must be a table")
	errorIfGFuncFail(t, L, func(L *LState) int {
		env := L.NewTable()
		L.Replace(EnvironIndex, env)
		errorIfNotEqual(t, env, L.Get(EnvironIndex))
		return 0
	})
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Replace(EnvironIndex, LNil)
		return 0
	}, "environment must be a table")
	errorIfGFuncFail(t, L, func(L *LState) int {
		gbl := L.NewTable()
		L.Replace(GlobalsIndex, gbl)
		errorIfNotEqual(t, gbl, L.G.Global)
		return 0
	})
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Replace(GlobalsIndex, LNil)
		return 0
	}, "_G must be a table")

	L2 := NewState()
	defer L2.Close()
	clo := L2.NewClosure(func(L2 *LState) int {
		L2.Replace(UpvalueIndex(1), LNumber(3))
		errorIfNotEqual(t, LNumber(3), L2.Get(UpvalueIndex(1)))
		return 0
	}, LNumber(1), LNumber(2))
	L2.SetGlobal("clo", clo)
	errorIfScriptFail(t, L2, `clo()`)
}

func TestRemove(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LString("a"))
	L.Push(LString("b"))
	L.Push(LString("c"))

	L.Remove(4)
	errorIfNotEqual(t, LString("a"), L.Get(1))
	errorIfNotEqual(t, LString("b"), L.Get(2))
	errorIfNotEqual(t, LString("c"), L.Get(3))
	errorIfNotEqual(t, 3, L.GetTop())

	L.Remove(3)
	errorIfNotEqual(t, LString("a"), L.Get(1))
	errorIfNotEqual(t, LString("b"), L.Get(2))
	errorIfNotEqual(t, LNil, L.Get(3))
	errorIfNotEqual(t, 2, L.GetTop())
	L.Push(LString("c"))

	L.Remove(-10)
	errorIfNotEqual(t, LString("a"), L.Get(1))
	errorIfNotEqual(t, LString("b"), L.Get(2))
	errorIfNotEqual(t, LString("c"), L.Get(3))
	errorIfNotEqual(t, 3, L.GetTop())

	L.Remove(2)
	errorIfNotEqual(t, LString("a"), L.Get(1))
	errorIfNotEqual(t, LString("c"), L.Get(2))
	errorIfNotEqual(t, LNil, L.Get(3))
	errorIfNotEqual(t, 2, L.GetTop())
}

func TestToInt(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewTable())
	errorIfNotEqual(t, 10, L.ToInt(1))
	errorIfNotEqual(t, 99, L.ToInt(2))
	errorIfNotEqual(t, 0, L.ToInt(3))
}

func TestToInt64(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewTable())
	errorIfNotEqual(t, int64(10), L.ToInt64(1))
	errorIfNotEqual(t, int64(99), L.ToInt64(2))
	errorIfNotEqual(t, int64(0), L.ToInt64(3))
}

func TestToNumber(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewTable())
	errorIfNotEqual(t, LNumber(10), L.ToNumber(1))
	errorIfNotEqual(t, LNumber(99.9), L.ToNumber(2))
	errorIfNotEqual(t, LNumber(0), L.ToNumber(3))
}

func TestToString(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewTable())
	errorIfNotEqual(t, "10", L.ToString(1))
	errorIfNotEqual(t, "99.9", L.ToString(2))
	errorIfNotEqual(t, "", L.ToString(3))
}

func TestToTable(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewTable())
	errorIfFalse(t, L.ToTable(1) == nil, "index 1 must be nil")
	errorIfFalse(t, L.ToTable(2) == nil, "index 2 must be nil")
	errorIfNotEqual(t, L.Get(3), L.ToTable(3))
}

func TestToFunction(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewFunction(func(L *LState) int { return 0 }))
	errorIfFalse(t, L.ToFunction(1) == nil, "index 1 must be nil")
	errorIfFalse(t, L.ToFunction(2) == nil, "index 2 must be nil")
	errorIfNotEqual(t, L.Get(3), L.ToFunction(3))
}

func TestToUserData(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	L.Push(L.NewUserData())
	errorIfFalse(t, L.ToUserData(1) == nil, "index 1 must be nil")
	errorIfFalse(t, L.ToUserData(2) == nil, "index 2 must be nil")
	errorIfNotEqual(t, L.Get(3), L.ToUserData(3))
}

func TestToChannel(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Push(LNumber(10))
	L.Push(LString("99.9"))
	var ch chan LValue
	L.Push(LChannel(ch))
	errorIfFalse(t, L.ToChannel(1) == nil, "index 1 must be nil")
	errorIfFalse(t, L.ToChannel(2) == nil, "index 2 must be nil")
	errorIfNotEqual(t, ch, L.ToChannel(3))
}

func TestObjLen(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfNotEqual(t, 3, L.ObjLen(LString("abc")))
	tbl := L.NewTable()
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	errorIfNotEqual(t, 2, L.ObjLen(tbl))
	mt := L.NewTable()
	L.SetField(mt, "__len", L.NewFunction(func(L *LState) int {
		tbl := L.CheckTable(1)
		L.Push(LNumber(tbl.Len() + 1))
		return 1
	}))
	L.SetMetatable(tbl, mt)
	errorIfNotEqual(t, 3, L.ObjLen(tbl))
	errorIfNotEqual(t, 0, L.ObjLen(LNumber(10)))
}

func TestConcat(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfNotEqual(t, "a1c", L.Concat(LString("a"), LNumber(1), LString("c")))
}

func TestPCall(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.Register("f1", func(L *LState) int {
		panic("panic!")
		return 0
	})
	errorIfScriptNotFail(t, L, `f1()`, "panic!")
	L.Push(L.GetGlobal("f1"))
	err := L.PCall(0, 0, L.NewFunction(func(L *LState) int {
		L.Push(LString("by handler"))
		return 1
	}))
	errorIfFalse(t, strings.Contains(err.Error(), "by handler"), "")

	err = L.PCall(0, 0, L.NewFunction(func(L *LState) int {
		L.RaiseError("error!")
		return 1
	}))
	errorIfFalse(t, strings.Contains(err.Error(), "error!"), "")

	err = L.PCall(0, 0, L.NewFunction(func(L *LState) int {
		panic("panicc!")
		return 1
	}))
	errorIfFalse(t, strings.Contains(err.Error(), "panicc!"), "")
}

func TestCoroutineApi1(t *testing.T) {
	L := NewState()
	defer L.Close()
	co, _ := L.NewThread()
	errorIfScriptFail(t, L, `
      function coro(v)
        assert(v == 10)
        local ret1, ret2 = coroutine.yield(1,2,3)
        assert(ret1 == 11)
        assert(ret2 == 12)
        coroutine.yield(4)
        return 5
      end
    `)
	fn := L.GetGlobal("coro").(*LFunction)
	st, err, values := L.Resume(co, fn, LNumber(10))
	errorIfNotEqual(t, ResumeYield, st)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, 3, len(values))
	errorIfNotEqual(t, LNumber(1), values[0].(LNumber))
	errorIfNotEqual(t, LNumber(2), values[1].(LNumber))
	errorIfNotEqual(t, LNumber(3), values[2].(LNumber))

	st, err, values = L.Resume(co, fn, LNumber(11), LNumber(12))
	errorIfNotEqual(t, ResumeYield, st)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, 1, len(values))
	errorIfNotEqual(t, LNumber(4), values[0].(LNumber))

	st, err, values = L.Resume(co, fn)
	errorIfNotEqual(t, ResumeOK, st)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, 1, len(values))
	errorIfNotEqual(t, LNumber(5), values[0].(LNumber))

	L.Register("myyield", func(L *LState) int {
		return L.Yield(L.ToNumber(1))
	})
	errorIfScriptFail(t, L, `
      function coro_error()
        coroutine.yield(1,2,3)
        myyield(4)
        assert(false, "--failed--")
      end
    `)
	fn = L.GetGlobal("coro_error").(*LFunction)
	co, _ = L.NewThread()
	st, err, values = L.Resume(co, fn)
	errorIfNotEqual(t, ResumeYield, st)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, 3, len(values))
	errorIfNotEqual(t, LNumber(1), values[0].(LNumber))
	errorIfNotEqual(t, LNumber(2), values[1].(LNumber))
	errorIfNotEqual(t, LNumber(3), values[2].(LNumber))

	st, err, values = L.Resume(co, fn)
	errorIfNotEqual(t, ResumeYield, st)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, 1, len(values))
	errorIfNotEqual(t, LNumber(4), values[0].(LNumber))

	st, err, values = L.Resume(co, fn)
	errorIfNotEqual(t, ResumeError, st)
	errorIfNil(t, err)
	errorIfFalse(t, strings.Contains(err.Error(), "--failed--"), "error message must be '--failed--'")
	st, err, values = L.Resume(co, fn)
	errorIfNotEqual(t, ResumeError, st)
	errorIfNil(t, err)
	errorIfFalse(t, strings.Contains(err.Error(), "can not resume a dead thread"), "can not resume a dead thread")

}

func TestContextTimeout(t *testing.T) {
	L := NewState()
	defer L.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	L.SetContext(ctx)
	errorIfNotEqual(t, ctx, L.Context())
	err := L.DoString(`
	  local clock = os.clock
      function sleep(n)  -- seconds
        local t0 = clock()
        while clock() - t0 <= n do end
      end
	  sleep(3)
	`)
	errorIfNil(t, err)
	errorIfFalse(t, strings.Contains(err.Error(), "context deadline exceeded"), "execution must be canceled")

	oldctx := L.RemoveContext()
	errorIfNotEqual(t, ctx, oldctx)
	errorIfNotNil(t, L.ctx)
}

func TestContextCancel(t *testing.T) {
	L := NewState()
	defer L.Close()
	ctx, cancel := context.WithCancel(context.Background())
	errch := make(chan error, 1)
	L.SetContext(ctx)
	go func() {
		errch <- L.DoString(`
	    local clock = os.clock
        function sleep(n)  -- seconds
          local t0 = clock()
          while clock() - t0 <= n do end
        end
	    sleep(3)
	  `)
	}()
	time.Sleep(1 * time.Second)
	cancel()
	err := <-errch
	errorIfNil(t, err)
	errorIfFalse(t, strings.Contains(err.Error(), "context canceled"), "execution must be canceled")
}

func TestContextWithCroutine(t *testing.T) {
	L := NewState()
	defer L.Close()
	ctx, cancel := context.WithCancel(context.Background())
	L.SetContext(ctx)
	defer cancel()
	L.DoString(`
	    function coro()
		  local i = 0
		  while true do
		    coroutine.yield(i)
			i = i+1
		  end
		  return i
	    end
	`)
	co, cocancel := L.NewThread()
	defer cocancel()
	fn := L.GetGlobal("coro").(*LFunction)
	_, err, values := L.Resume(co, fn)
	errorIfNotNil(t, err)
	errorIfNotEqual(t, LNumber(0), values[0])
	// cancel the parent context
	cancel()
	_, err, values = L.Resume(co, fn)
	errorIfNil(t, err)
	errorIfFalse(t, strings.Contains(err.Error(), "context canceled"), "coroutine execution must be canceled when the parent context is canceled")

}

func TestPCallAfterFail(t *testing.T) {
	L := NewState()
	defer L.Close()
	errFn := L.NewFunction(func(L *LState) int {
		L.RaiseError("error!")
		return 0
	})
	changeError := L.NewFunction(func(L *LState) int {
		L.Push(errFn)
		err := L.PCall(0, 0, nil)
		if err != nil {
			L.RaiseError("A New Error")
		}
		return 0
	})
	L.Push(changeError)
	err := L.PCall(0, 0, nil)
	errorIfFalse(t, strings.Contains(err.Error(), "A New Error"), "error not propogated correctly")
}
