package gorethink

import (
	"time"

	test "gopkg.in/check.v1"
)

type object struct {
	ID    int64  `gorethink:"id,omitempty"`
	Name  string `gorethink:"name"`
	Attrs []attr
}

type attr struct {
	Name  string
	Value interface{}
}

func (s *RethinkSuite) TestCursorLiteral(c *test.C) {
	res, err := Expr(5).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response interface{}
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, 5)
}

func (s *RethinkSuite) TestCursorSlice(c *test.C) {
	res, err := Expr([]interface{}{1, 2, 3, 4, 5}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response []interface{}
	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{1, 2, 3, 4, 5})
}

func (s *RethinkSuite) TestCursorPartiallyNilSlice(c *test.C) {
	res, err := Expr(map[string]interface{}{
		"item": []interface{}{
			map[string]interface{}{"num": 1},
			nil,
		},
	}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response map[string]interface{}
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"item": []interface{}{
			map[string]interface{}{"num": 1},
			nil,
		},
	})
}

func (s *RethinkSuite) TestCursorMap(c *test.C) {
	res, err := Expr(map[string]interface{}{
		"id":   2,
		"name": "Object 1",
	}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response map[string]interface{}
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"id":   2,
		"name": "Object 1",
	})
}

func (s *RethinkSuite) TestCursorMapIntoInterface(c *test.C) {
	res, err := Expr(map[string]interface{}{
		"id":   2,
		"name": "Object 1",
	}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response interface{}
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"id":   2,
		"name": "Object 1",
	})
}

func (s *RethinkSuite) TestCursorMapNested(c *test.C) {
	res, err := Expr(map[string]interface{}{
		"id":   2,
		"name": "Object 1",
		"attr": []interface{}{map[string]interface{}{
			"name":  "attr 1",
			"value": "value 1",
		}},
	}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response interface{}
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{
		"id":   2,
		"name": "Object 1",
		"attr": []interface{}{map[string]interface{}{
			"name":  "attr 1",
			"value": "value 1",
		}},
	})
}

func (s *RethinkSuite) TestCursorStruct(c *test.C) {
	res, err := Expr(map[string]interface{}{
		"id":   2,
		"name": "Object 1",
		"Attrs": []interface{}{map[string]interface{}{
			"Name":  "attr 1",
			"Value": "value 1",
		}},
	}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response object
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, object{
		ID:   2,
		Name: "Object 1",
		Attrs: []attr{attr{
			Name:  "attr 1",
			Value: "value 1",
		}},
	})
}

func (s *RethinkSuite) TestCursorStructPseudoTypes(c *test.C) {
	var zeroTime time.Time
	t := time.Now()

	res, err := Expr(map[string]interface{}{
		"T": time.Unix(t.Unix(), 0).In(time.UTC),
		"Z": zeroTime,
		"B": []byte("hello"),
	}).Run(session)
	c.Assert(err, test.IsNil)

	var response PseudoTypes
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	c.Assert(response.T.Equal(time.Unix(t.Unix(), 0)), test.Equals, true)
	c.Assert(response.Z.Equal(zeroTime), test.Equals, true)
	c.Assert(response.B, jsonEquals, []byte("hello"))
}

func (s *RethinkSuite) TestCursorAtomString(c *test.C) {
	res, err := Expr("a").Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response string
	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "a")
}

func (s *RethinkSuite) TestCursorAtomArray(c *test.C) {
	res, err := Expr([]interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Type(), test.Equals, "Cursor")

	var response []int
	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0})
}

func (s *RethinkSuite) TestEmptyResults(c *test.C) {
	DBCreate("test").Exec(session)
	DB("test").TableCreate("test").Exec(session)
	res, err := DB("test").Table("test").Get("missing value").Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.IsNil(), test.Equals, true)

	res, err = DB("test").Table("test").Get("missing value").Run(session)
	c.Assert(err, test.IsNil)
	var response interface{}
	err = res.One(&response)
	c.Assert(err, test.Equals, ErrEmptyResult)
	c.Assert(res.IsNil(), test.Equals, true)

	res, err = Expr(nil).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.IsNil(), test.Equals, true)

	res, err = DB("test").Table("test").Get("missing value").Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.IsNil(), test.Equals, true)

	res, err = DB("test").Table("test").GetAll("missing value", "another missing value").Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Next(&response), test.Equals, false)

	var obj object
	obj.Name = "missing value"
	res, err = DB("test").Table("test").Filter(obj).Run(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.IsNil(), test.Equals, true)
}

func (s *RethinkSuite) TestCursorAll(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableDrop("Table3").Exec(session)
	DB("test").TableCreate("Table3").Exec(session)
	DB("test").Table("Table3").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table3").Insert([]interface{}{
		map[string]interface{}{
			"id":   2,
			"name": "Object 1",
			"Attrs": []interface{}{map[string]interface{}{
				"Name":  "attr 1",
				"Value": "value 1",
			}},
		},
		map[string]interface{}{
			"id":   3,
			"name": "Object 2",
			"Attrs": []interface{}{map[string]interface{}{
				"Name":  "attr 1",
				"Value": "value 1",
			}},
		},
	}).Exec(session)

	// Test query
	query := DB("test").Table("Table3").OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	var response []object
	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.HasLen, 2)
	c.Assert(response, test.DeepEquals, []object{
		object{
			ID:   2,
			Name: "Object 1",
			Attrs: []attr{attr{
				Name:  "attr 1",
				Value: "value 1",
			}},
		},
		object{
			ID:   3,
			Name: "Object 2",
			Attrs: []attr{attr{
				Name:  "attr 1",
				Value: "value 1",
			}},
		},
	})
}

func (s *RethinkSuite) TestCursorListen(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableDrop("Table3").Exec(session)
	DB("test").TableCreate("Table3").Exec(session)
	DB("test").Table("Table3").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table3").Insert([]interface{}{
		map[string]interface{}{
			"id":   2,
			"name": "Object 1",
			"Attrs": []interface{}{map[string]interface{}{
				"Name":  "attr 1",
				"Value": "value 1",
			}},
		},
		map[string]interface{}{
			"id":   3,
			"name": "Object 2",
			"Attrs": []interface{}{map[string]interface{}{
				"Name":  "attr 1",
				"Value": "value 1",
			}},
		},
	}).Exec(session)

	// Test query
	query := DB("test").Table("Table3").OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	ch := make(chan object)
	res.Listen(ch)
	var response []object
	for v := range ch {
		response = append(response, v)
	}

	c.Assert(response, test.HasLen, 2)
	c.Assert(response, test.DeepEquals, []object{
		object{
			ID:   2,
			Name: "Object 1",
			Attrs: []attr{attr{
				Name:  "attr 1",
				Value: "value 1",
			}},
		},
		object{
			ID:   3,
			Name: "Object 2",
			Attrs: []attr{attr{
				Name:  "attr 1",
				Value: "value 1",
			}},
		},
	})
}

func (s *RethinkSuite) TestCursorReuseResult(c *test.C) {
	// Test query
	query := Expr([]interface{}{
		map[string]interface{}{
			"A": "a",
		},
		map[string]interface{}{
			"B": 1,
		},
		map[string]interface{}{
			"A": "a",
		},
		map[string]interface{}{
			"B": 1,
		},
		map[string]interface{}{
			"A": "a",
			"B": 1,
		},
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	var i int
	var result SimpleT
	for res.Next(&result) {
		switch i {
		case 0:
			c.Assert(result, test.DeepEquals, SimpleT{
				A: "a",
				B: 0,
			})
		case 1:
			c.Assert(result, test.DeepEquals, SimpleT{
				A: "",
				B: 1,
			})
		case 2:
			c.Assert(result, test.DeepEquals, SimpleT{
				A: "a",
				B: 0,
			})
		case 3:
			c.Assert(result, test.DeepEquals, SimpleT{
				A: "",
				B: 1,
			})
		case 4:
			c.Assert(result, test.DeepEquals, SimpleT{
				A: "a",
				B: 1,
			})
		default:
			c.Fatalf("Unexpected number of results")
		}

		i++
	}
	c.Assert(res.Err(), test.IsNil)
}
