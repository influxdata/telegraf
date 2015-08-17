package gorethink

import (
	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestAggregationReduce(c *test.C) {
	var response int
	query := Expr(arr).Reduce(func(acc, val Term) Term {
		return acc.Add(val)
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 45)
}

func (s *RethinkSuite) TestAggregationExprCount(c *test.C) {
	var response int
	query := Expr(arr).Count()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 9)
}

func (s *RethinkSuite) TestAggregationDistinct(c *test.C) {
	var response []int
	query := Expr(darr).Distinct()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.HasLen, 5)
}

func (s *RethinkSuite) TestAggregationGroupMapReduce(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group(func(row Term) Term {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row Term) Term {
		return row.Field("num")
	}).Reduce(func(acc, num Term) Term {
		return acc.Add(num)
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"reduction": 135, "group": false},
		map[string]interface{}{"reduction": 70, "group": true},
	})
}

func (s *RethinkSuite) TestAggregationGroupMapReduceUngroup(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group(func(row Term) Term {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row Term) Term {
		return row.Field("num")
	}).Reduce(func(acc, num Term) Term {
		return acc.Add(num)
	}).Ungroup().OrderBy("reduction")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"reduction": 70, "group": true},
		map[string]interface{}{"reduction": 135, "group": false},
	})
}

func (s *RethinkSuite) TestAggregationGroupMapReduceTable(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("TestAggregationGroupedMapReduceTable").Exec(session)

	// Insert rows
	err := DB("test").Table("TestAggregationGroupedMapReduceTable").Insert(objList).Exec(session)
	c.Assert(err, test.IsNil)

	var response []interface{}
	query := DB("test").Table("TestAggregationGroupedMapReduceTable").Group(func(row Term) Term {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row Term) Term {
		return row.Field("num")
	}).Reduce(func(acc, num Term) Term {
		return acc.Add(num)
	})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"reduction": 135, "group": false},
		map[string]interface{}{"reduction": 70, "group": true},
	})
}

func (s *RethinkSuite) TestAggregationGroupCount(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": []interface{}{
			map[string]interface{}{"id": 1, "num": 0, "g1": 1, "g2": 1},
			map[string]interface{}{"num": 15, "g1": 1, "g2": 1, "id": 6},
			map[string]interface{}{"id": 7, "num": 0, "g1": 1, "g2": 2},
		}},
		map[string]interface{}{"group": 2, "reduction": []interface{}{
			map[string]interface{}{"g1": 2, "g2": 2, "id": 2, "num": 5},
			map[string]interface{}{"num": 0, "g1": 2, "g2": 3, "id": 4},
			map[string]interface{}{"num": 100, "g1": 2, "g2": 3, "id": 5},
			map[string]interface{}{"g2": 3, "id": 9, "num": 25, "g1": 2},
		}},
		map[string]interface{}{"group": 3, "reduction": []interface{}{
			map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3},
		}},
		map[string]interface{}{"group": 4, "reduction": []interface{}{
			map[string]interface{}{"id": 8, "num": 50, "g1": 4, "g2": 2},
		}},
	})
}

func (s *RethinkSuite) TestAggregationGroupSum(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Sum("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": 15},
		map[string]interface{}{"reduction": 130, "group": 2},
		map[string]interface{}{"reduction": 10, "group": 3},
		map[string]interface{}{"group": 4, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupAvg(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Avg("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": 5},
		map[string]interface{}{"group": 2, "reduction": 32.5},
		map[string]interface{}{"group": 3, "reduction": 10},
		map[string]interface{}{"group": 4, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupMin(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Min("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": map[string]interface{}{"id": 1, "num": 0, "g1": 1, "g2": 1}},
		map[string]interface{}{"reduction": map[string]interface{}{"num": 0, "g1": 2, "g2": 3, "id": 4}, "group": 2},
		map[string]interface{}{"group": 3, "reduction": map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3}},
		map[string]interface{}{"group": 4, "reduction": map[string]interface{}{"g2": 2, "id": 8, "num": 50, "g1": 4}},
	})
}

func (s *RethinkSuite) TestAggregationGroupMax(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Max("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"reduction": map[string]interface{}{"num": 15, "g1": 1, "g2": 1, "id": 6}, "group": 1},
		map[string]interface{}{"group": 2, "reduction": map[string]interface{}{"num": 100, "g1": 2, "g2": 3, "id": 5}},
		map[string]interface{}{"group": 3, "reduction": map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3}},
		map[string]interface{}{"group": 4, "reduction": map[string]interface{}{"g2": 2, "id": 8, "num": 50, "g1": 4}},
	})
}

func (s *RethinkSuite) TestAggregationMin(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table2").Exec(session)
	DB("test").Table("Table2").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table2").Insert(objList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("Table2").MinIndex("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 1, "g1": 1, "g2": 1, "num": 0})
}

func (s *RethinkSuite) TestAggregationMaxIndex(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table2").Exec(session)
	DB("test").Table("Table2").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table2").Insert(objList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("Table2").MaxIndex("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 5, "g1": 2, "g2": 3, "num": 100})
}

func (s *RethinkSuite) TestAggregationMultipleGroupSum(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1", "g2").Sum("num")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": []interface{}{1, 1}, "reduction": 15},
		map[string]interface{}{"reduction": 0, "group": []interface{}{1, 2}},
		map[string]interface{}{"group": []interface{}{2, 2}, "reduction": 5},
		map[string]interface{}{"reduction": 125, "group": []interface{}{2, 3}},
		map[string]interface{}{"group": []interface{}{3, 2}, "reduction": 10},
		map[string]interface{}{"group": []interface{}{4, 2}, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupChained(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Max("num").Field("g2")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": 1},
		map[string]interface{}{"group": 2, "reduction": 3},
		map[string]interface{}{"group": 3, "reduction": 2},
		map[string]interface{}{"group": 4, "reduction": 2},
	})
}

func (s *RethinkSuite) TestAggregationGroupUngroup(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1", "g2").Max("num").Ungroup()
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"group": []interface{}{1, 1}, "reduction": map[string]interface{}{"g1": 1, "g2": 1, "id": 6, "num": 15}},
		map[string]interface{}{"group": []interface{}{1, 2}, "reduction": map[string]interface{}{"g1": 1, "g2": 2, "id": 7, "num": 0}},
		map[string]interface{}{"group": []interface{}{2, 2}, "reduction": map[string]interface{}{"g1": 2, "g2": 2, "id": 2, "num": 5}},
		map[string]interface{}{"group": []interface{}{2, 3}, "reduction": map[string]interface{}{"g1": 2, "g2": 3, "id": 5, "num": 100}},
		map[string]interface{}{"group": []interface{}{3, 2}, "reduction": map[string]interface{}{"g2": 2, "id": 3, "num": 10, "g1": 3}},
		map[string]interface{}{"reduction": map[string]interface{}{"num": 50, "g1": 4, "g2": 2, "id": 8}, "group": []interface{}{4, 2}},
	})
}

func (s *RethinkSuite) TestAggregationContains(c *test.C) {
	var response interface{}
	query := Expr(arr).Contains(2)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}
