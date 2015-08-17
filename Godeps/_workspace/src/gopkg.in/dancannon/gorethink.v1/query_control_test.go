package gorethink

import (
	"bytes"
	"testing"
	"time"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestControlExprNil(c *test.C) {
	var response interface{}
	query := Expr(nil)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.Equals, ErrEmptyResult)
	c.Assert(response, test.Equals, nil)
}

func (s *RethinkSuite) TestControlExprSimple(c *test.C) {
	var response int
	query := Expr(1)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 1)
}

func (s *RethinkSuite) TestControlExprList(c *test.C) {
	var response []interface{}
	query := Expr(narr)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		1, 2, 3, 4, 5, 6, []interface{}{
			7.1, 7.2, 7.3,
		},
	})
}

func (s *RethinkSuite) TestControlExprObj(c *test.C) {
	var response map[string]interface{}
	query := Expr(nobj)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"A": 1,
		"B": 2,
		"C": map[string]interface{}{
			"1": 3,
			"2": 4,
		},
	})
}

func (s *RethinkSuite) TestControlStruct(c *test.C) {
	var response map[string]interface{}
	query := Expr(str)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"id": "A",
		"B":  1,
		"D":  map[string]interface{}{"D2": "2", "D1": 1},
		"E":  []interface{}{"E1", "E2", "E3", 4},
		"F": map[string]interface{}{
			"XA": 2,
			"XB": "B",
			"XC": []interface{}{"XC1", "XC2"},
			"XD": map[string]interface{}{
				"YA": 3,
				"YB": map[string]interface{}{
					"1": "1",
					"2": "2",
					"3": 3,
				},
				"YC": map[string]interface{}{
					"YC1": "YC1",
				},
				"YD": map[string]interface{}{
					"YD1": "YD1",
				},
			},
			"XE": "XE",
			"XF": []interface{}{"XE1", "XE2"},
		},
	})
}

func (s *RethinkSuite) TestControlMapTypeAlias(c *test.C) {
	var response TMap
	query := Expr(TMap{"A": 1, "B": 2})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, TMap{"A": 1, "B": 2})
}

func (s *RethinkSuite) TestControlStringTypeAlias(c *test.C) {
	var response TStr
	query := Expr(TStr("Hello"))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, TStr("Hello"))
}

func (s *RethinkSuite) TestControlExprTypes(c *test.C) {
	var response []interface{}
	query := Expr([]interface{}{int64(1), uint64(1), float64(1.0), int32(1), uint32(1), float32(1), "1", true, false})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{int64(1), uint64(1), float64(1.0), int32(1), uint32(1), float32(1), "1", true, false})
}

func (s *RethinkSuite) TestControlJs(c *test.C) {
	var response int
	query := JS("1;")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 1)
}

func (s *RethinkSuite) TestControlHttp(c *test.C) {
	if testing.Short() {
		c.Skip("-short set")
	}

	var response map[string]interface{}
	query := HTTP("httpbin.org/get?data=1")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response["args"], jsonEquals, map[string]interface{}{
		"data": "1",
	})
}

func (s *RethinkSuite) TestControlJson(c *test.C) {
	var response []int
	query := JSON("[1,2,3]")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{1, 2, 3})
}

func (s *RethinkSuite) TestControlError(c *test.C) {
	query := Error("An error occurred")
	err := query.Exec(session)
	c.Assert(err, test.NotNil)

	c.Assert(err, test.NotNil)
	c.Assert(err, test.FitsTypeOf, RQLRuntimeError{})

	c.Assert(err.Error(), test.Equals, "gorethink: An error occurred in: \nr.Error(\"An error occurred\")")
}

func (s *RethinkSuite) TestControlDoNothing(c *test.C) {
	var response []interface{}
	query := Do([]interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"a": 2}, map[string]interface{}{"a": 3}})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"a": 2}, map[string]interface{}{"a": 3}})
}

func (s *RethinkSuite) TestControlArgs(c *test.C) {
	var response time.Time
	query := Time(Args(Expr([]interface{}{2014, 7, 12, "Z"})))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response.Unix(), test.Equals, int64(1405123200))
}

func (s *RethinkSuite) TestControlBinaryByteArray(c *test.C) {
	var response []byte

	query := Binary([]byte("Hello World"))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response, []byte("Hello World")), test.Equals, true)
}

type byteArray []byte

func (s *RethinkSuite) TestControlBinaryByteArrayAlias(c *test.C) {
	var response []byte

	query := Binary(byteArray("Hello World"))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response, []byte("Hello World")), test.Equals, true)
}

func (s *RethinkSuite) TestControlBinaryExpr(c *test.C) {
	var response []byte

	query := Expr([]byte("Hello World"))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response, []byte("Hello World")), test.Equals, true)
}

func (s *RethinkSuite) TestControlBinaryExprAlias(c *test.C) {
	var response []byte

	query := Expr(byteArray("Hello World"))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response, []byte("Hello World")), test.Equals, true)
}

func (s *RethinkSuite) TestControlBinaryTerm(c *test.C) {
	var response []byte

	query := Binary(Expr([]byte("Hello World")))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response, []byte("Hello World")), test.Equals, true)
}

func (s *RethinkSuite) TestControlBinaryElemTerm(c *test.C) {
	var response map[string]interface{}

	query := Expr(map[string]interface{}{
		"bytes": []byte("Hello World"),
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(bytes.Equal(response["bytes"].([]byte), []byte("Hello World")), test.Equals, true)
}

func (s *RethinkSuite) TestControlDo(c *test.C) {
	var response []interface{}
	query := Do([]interface{}{
		map[string]interface{}{"a": 1},
		map[string]interface{}{"a": 2},
		map[string]interface{}{"a": 3},
	}, func(row Term) Term {
		return row.Field("a")
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{1, 2, 3})
}

func (s *RethinkSuite) TestControlDoWithExpr(c *test.C) {
	var response []interface{}
	query := Expr([]interface{}{
		map[string]interface{}{"a": 1},
		map[string]interface{}{"a": 2},
		map[string]interface{}{"a": 3},
	}).Do(func(row Term) Term {
		return row.Field("a")
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{1, 2, 3})
}

func (s *RethinkSuite) TestControlBranchSimple(c *test.C) {
	var response int
	query := Branch(
		true,
		1,
		2,
	)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 1)
}

func (s *RethinkSuite) TestControlBranchWithMapExpr(c *test.C) {
	var response []interface{}
	query := Expr([]interface{}{1, 2, 3}).Map(Branch(
		Row.Eq(2),
		Row.Sub(1),
		Row.Add(1),
	))
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{2, 1, 4})
}

func (s *RethinkSuite) TestControlDefault(c *test.C) {
	var response []interface{}
	query := Expr(defaultObjList).Map(func(row Term) Term {
		return row.Field("a").Default(1)
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{1, 1})
}

func (s *RethinkSuite) TestControlCoerceTo(c *test.C) {
	var response string
	query := Expr(1).CoerceTo("STRING")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "1")
}

func (s *RethinkSuite) TestControlTypeOf(c *test.C) {
	var response string
	query := Expr(1).TypeOf()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "NUMBER")
}

func (s *RethinkSuite) TestControlRangeNoArgs(c *test.C) {
	var response []int
	query := Range().Limit(100)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(len(response), test.Equals, 100)
}

func (s *RethinkSuite) TestControlRangeSingleArgs(c *test.C) {
	var response []int
	query := Range(4)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []int{0, 1, 2, 3})
}

func (s *RethinkSuite) TestControlRangeTwoArgs(c *test.C) {
	var response []int
	query := Range(4, 6)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []int{4, 5})
}

func (s *RethinkSuite) TestControlToJSON(c *test.C) {
	var response string
	query := Expr([]int{4, 5}).ToJSON()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "[4,5]")
}
