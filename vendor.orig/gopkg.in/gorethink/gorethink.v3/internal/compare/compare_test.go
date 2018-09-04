package compare

import "testing"

func TestCompareString(t *testing.T) {

	// simple
	Assert(t, "a", "a")
	Assert(t, "á", "á")
	Assert(t, "something longer\nwith two lines", "something longer\nwith two lines")

	AssertFalse(t, "a", "b")
	AssertFalse(t, "a", 1)
	AssertFalse(t, "a", []interface{}{})
	AssertFalse(t, "a", nil)
	AssertFalse(t, "a", []interface{}{"a"})
	AssertFalse(t, "a", map[string]interface{}{"a": 1})
}
func TestCompareArray(t *testing.T) {

	// simple pass
	Assert(t, []interface{}{1, 2, 3}, []interface{}{1, 2, 3})

	// out of order
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{1, 3, 2})

	// totally mistmatched lists
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{3, 4, 5})

	// missing items
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{1, 2})
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{1, 3})

	// extra items
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{1, 2, 3, 4})

	// empty array
	Assert(t, []interface{}{}, []interface{}{})
	AssertFalse(t, []interface{}{1}, []interface{}{})
	AssertFalse(t, []interface{}{}, []interface{}{1})
	AssertFalse(t, []interface{}{}, nil)

	// strings
	Assert(t, []interface{}{"a", "b"}, []interface{}{"a", "b"})
	AssertFalse(t, []interface{}{"a", "c"}, []interface{}{"a", "b"})

	// multiple of a single value
	Assert(t, []interface{}{1, 2, 2, 3, 3, 3}, []interface{}{1, 2, 2, 3, 3, 3})
	AssertFalse(t, []interface{}{1, 2, 2, 3, 3, 3}, []interface{}{1, 2, 3})
	AssertFalse(t, []interface{}{1, 2, 3}, []interface{}{1, 2, 2, 3, 3, 3})
}
func TestCompareArray_PartialMatch(t *testing.T) {
	// note that these are all in-order

	// simple
	Assert(t, PartialMatch([]interface{}{1}), []interface{}{1, 2, 3})
	Assert(t, PartialMatch([]interface{}{2}), []interface{}{1, 2, 3})
	Assert(t, PartialMatch([]interface{}{3}), []interface{}{1, 2, 3})

	Assert(t, PartialMatch([]interface{}{1, 2}), []interface{}{1, 2, 3})
	Assert(t, PartialMatch([]interface{}{1, 3}), []interface{}{1, 2, 3})

	Assert(t, PartialMatch([]interface{}{1, 2, 3}), []interface{}{1, 2, 3})

	AssertFalse(t, PartialMatch([]interface{}{4}), []interface{}{1, 2, 3})

	// ordered
	AssertFalse(t, PartialMatch([]interface{}{3, 2, 1}).SetOrdered(true), []interface{}{1, 2, 3})
	AssertFalse(t, PartialMatch([]interface{}{1, 3, 2}).SetOrdered(true), []interface{}{1, 2, 3})

	// empty array
	Assert(t, PartialMatch([]interface{}{}), []interface{}{1, 2, 3})

	// multiple of a single items
	Assert(t, PartialMatch([]interface{}{1, 2, 2}), []interface{}{1, 2, 2, 3, 3, 3})
	AssertFalse(t, PartialMatch([]interface{}{1, 2, 2, 2}), []interface{}{1, 2, 2, 3, 3, 3})
}
func TestCompareArray_unordered(t *testing.T) {

	// simple
	Assert(t, UnorderedMatch([]interface{}{1, 2}), []interface{}{1, 2})
	Assert(t, UnorderedMatch([]interface{}{2, 1}), []interface{}{1, 2})

	AssertFalse(t, UnorderedMatch([]interface{}{1, 2}), []interface{}{1, 2, 3})
	AssertFalse(t, UnorderedMatch([]interface{}{1, 3}), []interface{}{1, 2, 3})
	AssertFalse(t, UnorderedMatch([]interface{}{3, 1}), []interface{}{1, 2, 3})

	// empty array
	Assert(t, UnorderedMatch([]interface{}{}), []interface{}{})
}
func TestCompareMap(t *testing.T) {

	// simple
	Assert(t, map[string]interface{}{"a": 1, "b": 2, "c": 3}, map[string]interface{}{"a": 1, "b": 2, "c": 3})
	Assert(t, map[string]interface{}{"a": 1, "b": 2, "c": 3}, map[string]interface{}{"c": 3, "a": 1, "b": 2})

	AssertFalse(t, map[string]interface{}{"a": 1, "b": 2, "c": 3}, map[string]interface{}{"a": 1})
	AssertFalse(t, map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1, "b": 2, "c": 3})

	// empty
	Assert(t, map[string]interface{}{}, map[string]interface{}{})
	AssertFalse(t, map[string]interface{}{}, map[string]interface{}{"a": 1})
	AssertFalse(t, map[string]interface{}{"a": 1}, map[string]interface{}{})

	Assert(t, map[interface{}]interface{}{1: 1225, 2: 1250, 3: 1275, 0: 1200}, map[string]interface{}{"2": 1250, "3": 1275, "0": 1200, "1": 1225})
	Assert(t, map[interface{}]interface{}{0: 22, 20: 22, 30: 23}, map[string]interface{}{"30": 23, "0": 22, "20": 22})
}
func TestCompareMap_PartialMatch(t *testing.T) {

	// simple
	Assert(t, PartialMatch(map[string]interface{}{"a": 1}), map[string]interface{}{"a": 1})
	Assert(t, PartialMatch(map[string]interface{}{"a": 1}), map[string]interface{}{"a": 1, "b": 2})

	AssertFalse(t, PartialMatch(map[string]interface{}{"a": 2}), map[string]interface{}{"a": 1, "b": 2})
	AssertFalse(t, PartialMatch(map[string]interface{}{"c": 1}), map[string]interface{}{"a": 1, "b": 2})
	AssertFalse(t, PartialMatch(map[string]interface{}{"a": 1, "b": 2}), map[string]interface{}{"b": 2})

	// empty
	Assert(t, PartialMatch(map[string]interface{}{}), map[string]interface{}{})
	Assert(t, PartialMatch(map[string]interface{}{}), map[string]interface{}{"a": 1})
	AssertFalse(t, PartialMatch(map[string]interface{}{"a": 1}), map[string]interface{}{})
}
func TestCompareMap_inSlice(t *testing.T) {

	// simple
	Assert(t, []interface{}{map[string]interface{}{"a": 1}}, []interface{}{map[string]interface{}{"a": 1}})
	Assert(t, []interface{}{map[string]interface{}{"a": 1, "b": 2}}, []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	Assert(t, []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}, []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}})

	AssertFalse(t, []interface{}{map[string]interface{}{"a": 1}}, []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	AssertFalse(t, []interface{}{map[string]interface{}{"a": 2, "b": 2}}, []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	AssertFalse(t, []interface{}{map[string]interface{}{"a": 2, "c": 3}}, []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	AssertFalse(t, []interface{}{map[string]interface{}{"a": 2, "c": 3}}, []interface{}{map[string]interface{}{"a": 1}})
	AssertFalse(t, []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}, []interface{}{map[string]interface{}{"a": 1, "b": 2}})

	// order
	AssertFalse(t, []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}, []interface{}{map[string]interface{}{"b": 2}, map[string]interface{}{"a": 1}})

	// partial
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{}}), []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{}}), []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1}}), []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1, "b": 2}}), []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}), []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}, map[string]interface{}{"c": 3}})

	AssertFalse(t, PartialMatch([]interface{}{map[string]interface{}{"a": 2}}), []interface{}{map[string]interface{}{"a": 1, "b": 2}})
	AssertFalse(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1, "b": 2}}), []interface{}{map[string]interface{}{"a": 1}})

	// partial order
	AssertFalse(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}).SetOrdered(true), []interface{}{map[string]interface{}{"b": 2}, map[string]interface{}{"a": 1}})

	// partial unordered
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}), []interface{}{map[string]interface{}{"b": 2}, map[string]interface{}{"a": 1}})
	Assert(t, PartialMatch([]interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}).SetOrdered(false), []interface{}{map[string]interface{}{"b": 2}, map[string]interface{}{"a": 1}})

	Assert(t, []interface{}{map[string]interface{}{"a": 1, "b": 1}, PartialMatch(map[string]interface{}{"a": 2})}, []interface{}{map[string]interface{}{"a": 1, "b": 1}, map[string]interface{}{"a": 2, "b": 2}})
}

func TestCompareUUID(t *testing.T) {

	// simple
	Assert(t, IsUUID(), "4e9e5bc2-9b11-4143-9aa1-75c10e7a193a")
	AssertFalse(t, IsUUID(), "4")
	AssertFalse(t, IsUUID(), "*")
	AssertFalse(t, IsUUID(), nil)
}

func TestCompareNumbers(t *testing.T) {

	// simple
	Assert(t, 1, 1)
	Assert(t, 1, 1.0)
	Assert(t, 1.0, 1)
	Assert(t, 1.0, 1.0)

	AssertFalse(t, 1, 2)
	AssertFalse(t, 1, 2.0)
	AssertFalse(t, 1.0, 2)
	AssertFalse(t, 1.0, 2.0)

	// precision
	AssertPrecision(t, 1, 1.4, 0.5)
	AssertPrecision(t, 1.0, 1.4, 0.5)

	AssertPrecisionFalse(t, 1, 2, 0.5)
	AssertPrecisionFalse(t, 1, 1.6, 0.5)
	AssertPrecisionFalse(t, 1.0, 2, 0.5)
	AssertPrecisionFalse(t, 1.0, 1.6, 0.5)
}
