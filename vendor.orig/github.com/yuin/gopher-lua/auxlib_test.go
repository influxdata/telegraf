package lua

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCheckInt(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LNumber(10))
		errorIfNotEqual(t, 10, L.CheckInt(2))
		L.Push(LString("aaa"))
		L.CheckInt(3)
		return 0
	}, "number expected, got string")
}

func TestCheckInt64(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LNumber(10))
		errorIfNotEqual(t, int64(10), L.CheckInt64(2))
		L.Push(LString("aaa"))
		L.CheckInt64(3)
		return 0
	}, "number expected, got string")
}

func TestCheckNumber(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LNumber(10))
		errorIfNotEqual(t, LNumber(10), L.CheckNumber(2))
		L.Push(LString("aaa"))
		L.CheckNumber(3)
		return 0
	}, "number expected, got string")
}

func TestCheckString(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LString("aaa"))
		errorIfNotEqual(t, "aaa", L.CheckString(2))
		L.Push(LNumber(10))
		L.CheckString(3)
		return 0
	}, "string expected, got number")
}

func TestCheckBool(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LTrue)
		errorIfNotEqual(t, true, L.CheckBool(2))
		L.Push(LNumber(10))
		L.CheckBool(3)
		return 0
	}, "boolean expected, got number")
}

func TestCheckTable(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		tbl := L.NewTable()
		L.Push(tbl)
		errorIfNotEqual(t, tbl, L.CheckTable(2))
		L.Push(LNumber(10))
		L.CheckTable(3)
		return 0
	}, "table expected, got number")
}

func TestCheckFunction(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		fn := L.NewFunction(func(l *LState) int { return 0 })
		L.Push(fn)
		errorIfNotEqual(t, fn, L.CheckFunction(2))
		L.Push(LNumber(10))
		L.CheckFunction(3)
		return 0
	}, "function expected, got number")
}

func TestCheckUserData(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		ud := L.NewUserData()
		L.Push(ud)
		errorIfNotEqual(t, ud, L.CheckUserData(2))
		L.Push(LNumber(10))
		L.CheckUserData(3)
		return 0
	}, "userdata expected, got number")
}

func TestCheckThread(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		th, _ := L.NewThread()
		L.Push(th)
		errorIfNotEqual(t, th, L.CheckThread(2))
		L.Push(LNumber(10))
		L.CheckThread(3)
		return 0
	}, "thread expected, got number")
}

func TestCheckChannel(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		ch := make(chan LValue)
		L.Push(LChannel(ch))
		errorIfNotEqual(t, ch, L.CheckChannel(2))
		L.Push(LString("aaa"))
		L.CheckChannel(3)
		return 0
	}, "channel expected, got string")
}

func TestCheckType(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LNumber(10))
		L.CheckType(2, LTNumber)
		L.CheckType(2, LTString)
		return 0
	}, "string expected, got number")
}

func TestCheckTypes(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LNumber(10))
		L.CheckTypes(2, LTString, LTBool, LTNumber)
		L.CheckTypes(2, LTString, LTBool)
		return 0
	}, "string or boolean expected, got number")
}

func TestCheckOption(t *testing.T) {
	opts := []string{
		"opt1",
		"opt2",
		"opt3",
	}
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		L.Push(LString("opt1"))
		errorIfNotEqual(t, 0, L.CheckOption(2, opts))
		L.Push(LString("opt5"))
		L.CheckOption(3, opts)
		return 0
	}, "invalid option: opt5 \\(must be one of opt1,opt2,opt3\\)")
}

func TestOptInt(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		errorIfNotEqual(t, 99, L.OptInt(1, 99))
		L.Push(LNumber(10))
		errorIfNotEqual(t, 10, L.OptInt(2, 99))
		L.Push(LString("aaa"))
		L.OptInt(3, 99)
		return 0
	}, "number expected, got string")
}

func TestOptInt64(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		errorIfNotEqual(t, int64(99), L.OptInt64(1, int64(99)))
		L.Push(LNumber(10))
		errorIfNotEqual(t, int64(10), L.OptInt64(2, int64(99)))
		L.Push(LString("aaa"))
		L.OptInt64(3, int64(99))
		return 0
	}, "number expected, got string")
}

func TestOptNumber(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		errorIfNotEqual(t, LNumber(99), L.OptNumber(1, LNumber(99)))
		L.Push(LNumber(10))
		errorIfNotEqual(t, LNumber(10), L.OptNumber(2, LNumber(99)))
		L.Push(LString("aaa"))
		L.OptNumber(3, LNumber(99))
		return 0
	}, "number expected, got string")
}

func TestOptString(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		errorIfNotEqual(t, "bbb", L.OptString(1, "bbb"))
		L.Push(LString("aaa"))
		errorIfNotEqual(t, "aaa", L.OptString(2, "bbb"))
		L.Push(LNumber(10))
		L.OptString(3, "bbb")
		return 0
	}, "string expected, got number")
}

func TestOptBool(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		errorIfNotEqual(t, true, L.OptBool(1, true))
		L.Push(LTrue)
		errorIfNotEqual(t, true, L.OptBool(2, false))
		L.Push(LNumber(10))
		L.OptBool(3, false)
		return 0
	}, "boolean expected, got number")
}

func TestOptTable(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		deftbl := L.NewTable()
		errorIfNotEqual(t, deftbl, L.OptTable(1, deftbl))
		tbl := L.NewTable()
		L.Push(tbl)
		errorIfNotEqual(t, tbl, L.OptTable(2, deftbl))
		L.Push(LNumber(10))
		L.OptTable(3, deftbl)
		return 0
	}, "table expected, got number")
}

func TestOptFunction(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		deffn := L.NewFunction(func(l *LState) int { return 0 })
		errorIfNotEqual(t, deffn, L.OptFunction(1, deffn))
		fn := L.NewFunction(func(l *LState) int { return 0 })
		L.Push(fn)
		errorIfNotEqual(t, fn, L.OptFunction(2, deffn))
		L.Push(LNumber(10))
		L.OptFunction(3, deffn)
		return 0
	}, "function expected, got number")
}

func TestOptUserData(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		defud := L.NewUserData()
		errorIfNotEqual(t, defud, L.OptUserData(1, defud))
		ud := L.NewUserData()
		L.Push(ud)
		errorIfNotEqual(t, ud, L.OptUserData(2, defud))
		L.Push(LNumber(10))
		L.OptUserData(3, defud)
		return 0
	}, "userdata expected, got number")
}

func TestOptChannel(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfGFuncNotFail(t, L, func(L *LState) int {
		defch := make(chan LValue)
		errorIfNotEqual(t, defch, L.OptChannel(1, defch))
		ch := make(chan LValue)
		L.Push(LChannel(ch))
		errorIfNotEqual(t, ch, L.OptChannel(2, defch))
		L.Push(LString("aaa"))
		L.OptChannel(3, defch)
		return 0
	}, "channel expected, got string")
}

func TestLoadFileForShebang(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	errorIfNotNil(t, err)

	err = ioutil.WriteFile(tmpFile.Name(), []byte(`#!/path/to/lua
print("hello")
`), 0644)
	errorIfNotNil(t, err)

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	L := NewState()
	defer L.Close()

	_, err = L.LoadFile(tmpFile.Name())
	errorIfNotNil(t, err)
}

func TestLoadFileForEmptyFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	errorIfNotNil(t, err)

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	L := NewState()
	defer L.Close()

	_, err = L.LoadFile(tmpFile.Name())
	errorIfNotNil(t, err)
}
