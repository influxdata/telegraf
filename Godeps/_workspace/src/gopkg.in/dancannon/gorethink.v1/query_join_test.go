package gorethink

import (
	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestJoinInnerJoin(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Join1").Exec(session)
	DB("test").TableCreate("Join2").Exec(session)

	// Insert rows
	DB("test").Table("Join1").Insert(joinTable1).Exec(session)
	DB("test").Table("Join2").Insert(joinTable2).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Join1").InnerJoin(DB("test").Table("Join2"), func(a, b Term) Term {
		return a.Field("id").Eq(b.Field("id"))
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)
	err = res.All(&response)

	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{
			"right": map[string]interface{}{"title": "goof", "id": 0},
			"left":  map[string]interface{}{"name": "bob", "id": 0},
		},
		map[string]interface{}{
			"right": map[string]interface{}{"title": "lmoe", "id": 2},
			"left":  map[string]interface{}{"name": "joe", "id": 2},
		},
	})
}

func (s *RethinkSuite) TestJoinInnerJoinZip(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Join1").Exec(session)
	DB("test").TableCreate("Join2").Exec(session)

	// Insert rows
	DB("test").Table("Join1").Insert(joinTable1).Exec(session)
	DB("test").Table("Join2").Insert(joinTable2).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Join1").InnerJoin(DB("test").Table("Join2"), func(a, b Term) Term {
		return a.Field("id").Eq(b.Field("id"))
	}).Zip()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"title": "goof", "name": "bob", "id": 0},
		map[string]interface{}{"title": "lmoe", "name": "joe", "id": 2},
	})
}

func (s *RethinkSuite) TestJoinOuterJoinZip(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Join1").Exec(session)
	DB("test").TableCreate("Join2").Exec(session)

	// Insert rows
	DB("test").Table("Join1").Insert(joinTable1).Exec(session)
	DB("test").Table("Join2").Insert(joinTable2).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Join1").OuterJoin(DB("test").Table("Join2"), func(a, b Term) Term {
		return a.Field("id").Eq(b.Field("id"))
	}).Zip().OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"title": "goof", "name": "bob", "id": 0},
		map[string]interface{}{"name": "tom", "id": 1},
		map[string]interface{}{"title": "lmoe", "name": "joe", "id": 2},
	})
}

func (s *RethinkSuite) TestJoinEqJoinZip(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Join1").Exec(session)
	DB("test").TableCreate("Join2").Exec(session)

	// Insert rows
	DB("test").Table("Join1").Insert(joinTable1).Exec(session)
	DB("test").Table("Join2").Insert(joinTable2).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Join1").EqJoin("id", DB("test").Table("Join2")).Zip()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"title": "goof", "name": "bob", "id": 0},
		map[string]interface{}{"title": "lmoe", "name": "joe", "id": 2},
	})
}

func (s *RethinkSuite) TestJoinEqJoinDiffIdsZip(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Join1").Exec(session)
	err := DB("test").TableCreate("Join3", TableCreateOpts{
		PrimaryKey: "it",
	}).Exec(session)
	c.Assert(err, test.IsNil)
	DB("test").Table("Join3").IndexCreate("it").Exec(session)

	// Insert rows
	DB("test").Table("Join1").Delete().Exec(session)
	DB("test").Table("Join3").Delete().Exec(session)
	DB("test").Table("Join1").Insert(joinTable1).Exec(session)
	DB("test").Table("Join3").Insert(joinTable3).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Join1").EqJoin("id", DB("test").Table("Join3"), EqJoinOpts{
		Index: "it",
	}).Zip()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"title": "goof", "name": "bob", "id": 0, "it": 0},
		map[string]interface{}{"title": "lmoe", "name": "joe", "id": 2, "it": 2},
	})
}

func (s *RethinkSuite) TestOrderByJoinEq(c *test.C) {
	type Map map[string]interface{}
	var err error

	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("test").Exec(session)
	DB("test").TableCreate("test2").Exec(session)
	tab := DB("test").Table("test")
	tab2 := DB("test").Table("test2")

	// insert rows
	err = tab.Insert(Map{"S": "s1", "T": 2}).Exec(session)
	err = tab.Insert(Map{"S": "s1", "T": 1}).Exec(session)
	err = tab.Insert(Map{"S": "s1", "T": 3}).Exec(session)
	err = tab.Insert(Map{"S": "s2", "T": 3}).Exec(session)
	c.Assert(err, test.IsNil)

	err = tab2.Insert(Map{"id": "s1", "N": "Rob"}).Exec(session)
	err = tab2.Insert(Map{"id": "s2", "N": "Zar"}).Exec(session)
	c.Assert(err, test.IsNil)

	// Test query
	var response []Map
	res, err := tab.OrderBy("T").EqJoin("S", tab2).Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Check(len(response), test.Equals, 4, test.Commentf("%v", response))
}
