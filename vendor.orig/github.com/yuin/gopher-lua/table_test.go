package lua

import (
	"testing"
)

func TestTableNewLTable(t *testing.T) {
	tbl := newLTable(-1, -2)
	errorIfNotEqual(t, 0, cap(tbl.array))

	tbl = newLTable(10, 9)
	errorIfNotEqual(t, 10, cap(tbl.array))
}

func TestTableLen(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.RawSetInt(10, LNil)
	tbl.RawSetInt(9, LNumber(10))
	tbl.RawSetInt(8, LNil)
	tbl.RawSetInt(7, LNumber(10))
	errorIfNotEqual(t, 9, tbl.Len())

	tbl = newLTable(0, 0)
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	errorIfNotEqual(t, 3, tbl.Len())
}

func TestTableInsert(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	tbl.Append(LTrue)

	tbl.Insert(5, LFalse)
	errorIfNotEqual(t, LFalse, tbl.RawGetInt(5))
	errorIfNotEqual(t, 5, tbl.Len())

	tbl.Insert(-10, LFalse)
	errorIfNotEqual(t, LFalse, tbl.RawGet(LNumber(-10)))
	errorIfNotEqual(t, 5, tbl.Len())

	tbl = newLTable(0, 0)
	tbl.Append(LNumber(1))
	tbl.Append(LNumber(2))
	tbl.Append(LNumber(3))
	tbl.Insert(1, LNumber(10))
	errorIfNotEqual(t, LNumber(10), tbl.RawGetInt(1))
	errorIfNotEqual(t, LNumber(1), tbl.RawGetInt(2))
	errorIfNotEqual(t, LNumber(2), tbl.RawGetInt(3))
	errorIfNotEqual(t, LNumber(3), tbl.RawGetInt(4))
	errorIfNotEqual(t, 4, tbl.Len())

	tbl = newLTable(0, 0)
	tbl.Insert(5, LNumber(10))
	errorIfNotEqual(t, LNumber(10), tbl.RawGetInt(5))

}

func TestTableMaxN(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	tbl.Append(LTrue)
	errorIfNotEqual(t, 3, tbl.MaxN())

	tbl = newLTable(0, 0)
	errorIfNotEqual(t, 0, tbl.MaxN())

	tbl = newLTable(10, 0)
	errorIfNotEqual(t, 0, tbl.MaxN())
}

func TestTableRemove(t *testing.T) {
	tbl := newLTable(0, 0)
	errorIfNotEqual(t, LNil, tbl.Remove(10))
	tbl.Append(LTrue)
	errorIfNotEqual(t, LNil, tbl.Remove(10))

	tbl.Append(LFalse)
	tbl.Append(LTrue)
	errorIfNotEqual(t, LFalse, tbl.Remove(2))
	errorIfNotEqual(t, 2, tbl.MaxN())
	tbl.Append(LFalse)
	errorIfNotEqual(t, LFalse, tbl.Remove(-1))
	errorIfNotEqual(t, 2, tbl.MaxN())

}

func TestTableRawSetInt(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.RawSetInt(MaxArrayIndex+1, LTrue)
	errorIfNotEqual(t, 0, tbl.MaxN())
	errorIfNotEqual(t, LTrue, tbl.RawGet(LNumber(MaxArrayIndex+1)))

	tbl.RawSetInt(1, LTrue)
	tbl.RawSetInt(3, LTrue)
	errorIfNotEqual(t, 3, tbl.MaxN())
	errorIfNotEqual(t, LTrue, tbl.RawGetInt(1))
	errorIfNotEqual(t, LNil, tbl.RawGetInt(2))
	errorIfNotEqual(t, LTrue, tbl.RawGetInt(3))
	tbl.RawSetInt(2, LTrue)
	errorIfNotEqual(t, LTrue, tbl.RawGetInt(1))
	errorIfNotEqual(t, LTrue, tbl.RawGetInt(2))
	errorIfNotEqual(t, LTrue, tbl.RawGetInt(3))
}

func TestTableRawSetH(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.RawSetH(LString("key"), LTrue)
	tbl.RawSetH(LString("key"), LNil)
	_, found := tbl.dict[LString("key")]
	errorIfNotEqual(t, false, found)

	tbl.RawSetH(LTrue, LTrue)
	tbl.RawSetH(LTrue, LNil)
	_, foundb := tbl.dict[LTrue]
	errorIfNotEqual(t, false, foundb)
}

func TestTableRawGetH(t *testing.T) {
	tbl := newLTable(0, 0)
	errorIfNotEqual(t, LNil, tbl.RawGetH(LNumber(1)))
	errorIfNotEqual(t, LNil, tbl.RawGetH(LString("key0")))
	tbl.RawSetH(LString("key0"), LTrue)
	tbl.RawSetH(LString("key1"), LFalse)
	tbl.RawSetH(LNumber(1), LTrue)
	errorIfNotEqual(t, LTrue, tbl.RawGetH(LString("key0")))
	errorIfNotEqual(t, LTrue, tbl.RawGetH(LNumber(1)))
	errorIfNotEqual(t, LNil, tbl.RawGetH(LString("notexist")))
	errorIfNotEqual(t, LNil, tbl.RawGetH(LTrue))
}

func TestTableForEach(t *testing.T) {
	tbl := newLTable(0, 0)
	tbl.Append(LNumber(1))
	tbl.Append(LNumber(2))
	tbl.Append(LNumber(3))
	tbl.Append(LNil)
	tbl.Append(LNumber(5))

	tbl.RawSetH(LString("a"), LString("a"))
	tbl.RawSetH(LString("b"), LString("b"))
	tbl.RawSetH(LString("c"), LString("c"))

	tbl.RawSetH(LTrue, LString("true"))
	tbl.RawSetH(LFalse, LString("false"))

	tbl.ForEach(func(key, value LValue) {
		switch k := key.(type) {
		case LBool:
			switch bool(k) {
			case true:
				errorIfNotEqual(t, LString("true"), value)
			case false:
				errorIfNotEqual(t, LString("false"), value)
			default:
				t.Fail()
			}
		case LNumber:
			switch int(k) {
			case 1:
				errorIfNotEqual(t, LNumber(1), value)
			case 2:
				errorIfNotEqual(t, LNumber(2), value)
			case 3:
				errorIfNotEqual(t, LNumber(3), value)
			case 4:
				errorIfNotEqual(t, LNumber(5), value)
			default:
				t.Fail()
			}
		case LString:
			switch string(k) {
			case "a":
				errorIfNotEqual(t, LString("a"), value)
			case "b":
				errorIfNotEqual(t, LString("b"), value)
			case "c":
				errorIfNotEqual(t, LString("c"), value)
			default:
				t.Fail()
			}
		}
	})
}
