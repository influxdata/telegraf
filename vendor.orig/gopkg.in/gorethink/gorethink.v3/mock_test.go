package gorethink

import (
	"fmt"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestMockExecSuccess(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test").Insert(map[string]string{
		"id": "mocked",
	})).Return(nil, nil)

	err := DB("test").Table("test").Insert(map[string]string{
		"id": "mocked",
	}).Exec(mock)
	c.Assert(err, test.IsNil)
	mock.AssertExpectations(c)
}

func (s *RethinkSuite) TestMockExecFail(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test").Insert(map[string]string{
		"id": "mocked",
	})).Return(nil, fmt.Errorf("Expected error"))

	err := DB("test").Table("test").Insert(map[string]string{
		"id": "mocked",
	}).Exec(mock)
	c.Assert(err, test.NotNil)
	mock.AssertExpectations(c)
}

func (s *RethinkSuite) TestMockRunSuccessSingleResult(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test").Get("mocked")).Return(map[string]interface{}{
		"id": "mocked",
	}, nil)

	res, err := DB("test").Table("test").Get("mocked").Run(mock)
	c.Assert(err, test.IsNil)

	var response interface{}
	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": "mocked"})
	mock.AssertExpectations(c)

	res.Close()
}

func (s *RethinkSuite) TestMockRunSuccessMultipleResults(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]interface{}{
		map[string]interface{}{"id": "mocked"},
	}, nil)

	res, err := DB("test").Table("test").Run(mock)
	c.Assert(err, test.IsNil)

	var response []interface{}
	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{map[string]interface{}{"id": "mocked"}})
	mock.AssertExpectations(c)

	res.Close()
}

func (s *RethinkSuite) TestMockRunSuccessMultipleResults_type(c *test.C) {
	type document struct {
		Id string
	}

	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]document{
		document{"mocked"},
	}, nil)

	res, err := DB("test").Table("test").Run(mock)
	c.Assert(err, test.IsNil)

	var response []interface{}
	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []document{document{"mocked"}})
	mock.AssertExpectations(c)

	res.Close()
}

func (s *RethinkSuite) TestMockRunMissingMock(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]interface{}{
		map[string]interface{}{"id": "mocked"},
	}, nil).Once()

	c.Assert(func() {
		c.Assert(DB("test").Table("test").Exec(mock), test.IsNil)
		c.Assert(DB("test").Table("test").Exec(mock), test.IsNil)
	}, test.PanicMatches, ""+
		"gorethink: mock: This query was unexpected:(?s:.*)")
	mock.AssertExpectations(c)
}

func (s *RethinkSuite) TestMockRunMissingQuery(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]interface{}{
		map[string]interface{}{"id": "mocked"},
	}, nil).Twice()

	c.Assert(DB("test").Table("test").Exec(mock), test.IsNil)

	t := &simpleTestingT{}
	mock.AssertExpectations(t)

	c.Assert(t.Failed(), test.Equals, true)
}

func (s *RethinkSuite) TestMockRunMissingQuerySingle(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]interface{}{
		map[string]interface{}{"id": "mocked"},
	}, nil).Once()

	t := &simpleTestingT{}
	mock.AssertExpectations(t)

	c.Assert(t.Failed(), test.Equals, true)
}

func (s *RethinkSuite) TestMockRunMissingQueryMultiple(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test")).Return([]interface{}{
		map[string]interface{}{"id": "mocked"},
	}, nil).Twice()

	c.Assert(DB("test").Table("test").Exec(mock), test.IsNil)

	t := &simpleTestingT{}
	mock.AssertExpectations(t)

	c.Assert(t.Failed(), test.Equals, true)
}

func (s *RethinkSuite) TestMockRunMutlipleQueries(c *test.C) {
	mock := NewMock()
	mock.On(DB("test").Table("test").Get("mocked1")).Return(map[string]interface{}{
		"id": "mocked1",
	}, nil).Times(2)
	mock.On(DB("test").Table("test").Get("mocked2")).Return(map[string]interface{}{
		"id": "mocked2",
	}, nil).Times(1)

	var response interface{}

	// Query 1
	res, err := DB("test").Table("test").Get("mocked1").Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": "mocked1"})

	// Query 2
	res, err = DB("test").Table("test").Get("mocked1").Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": "mocked1"})

	// Query 3
	res, err = DB("test").Table("test").Get("mocked2").Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": "mocked2"})

	mock.AssertExpectations(c)
}

func (s *RethinkSuite) TestMockQueriesWithFuncs(c *test.C) {
	mock := NewMock()
	mock.On(Expr([]int{2}).Map(func(row Term) interface{} {
		return row.Add(1)
	})).Return([]int{3}, nil).Times(2)
	mock.On(Expr([]int{4}).Map(func(row1, row2 Term) interface{} {
		return row1.Add(1)
	})).Return([]int{5}, nil).Times(1)
	mock.On(Expr([]int{9}).Map(func(row1, row2 Term) interface{} {
		return row2.Add(1)
	})).Return([]int{10}, nil).Times(1)

	var response []int

	// Query 1
	res, err := Expr([]int{2}).Map(func(row Term) interface{} {
		return row.Add(1)
	}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []int{3})

	// Query 2
	res, err = Expr([]int{2}).Map(func(row Term) interface{} {
		return row.Add(1)
	}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []int{3})

	// Query 3
	res, err = Expr([]int{4}).Map(func(row1, row2 Term) interface{} {
		return row1.Add(1)
	}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []int{5})

	// Query 5
	res, err = Expr([]int{9}).Map(func(row1, row2 Term) interface{} {
		return row2.Add(1)
	}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []int{10})

	mock.AssertExpectations(c)
}

func (s *RethinkSuite) TestMockAnything(c *test.C) {
	mock := NewMock()
	mock.On(MockAnything()).Return("okay", nil).Times(1)
	mock.On(Table("test").MockAnything()).Return("okay2", nil).Times(1)
	mock.On(Table("test").Insert(map[string]interface{}{
		"id": MockAnything(),
	})).Return("okay3", nil).Times(1)
	mock.On(Expr([]interface{}{1, 2, MockAnything()})).Return("okay4", nil).Times(1)

	var response string

	// Query 1
	res, err := Expr("test_1").Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, "okay")

	// Query 2
	res, err = Table("test").Get("mocked1").Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, "okay2")

	// Query 3
	res, err = Table("test").Insert(map[string]interface{}{
		"id": "10ECE456-3C4D-4864-A843-879FCB0D133F",
	}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, "okay3")

	// Query 3
	res, err = Expr([]interface{}{1, 2, 3}).Run(mock)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, "okay4")

	mock.AssertExpectations(c)
}

type simpleTestingT struct {
	failed bool
}

func (t *simpleTestingT) Logf(format string, args ...interface{}) {
}
func (t *simpleTestingT) Errorf(format string, args ...interface{}) {
	t.failed = true
}
func (t *simpleTestingT) FailNow() {
	t.failed = true
}
func (t *simpleTestingT) Failed() bool {
	return t.failed
}
