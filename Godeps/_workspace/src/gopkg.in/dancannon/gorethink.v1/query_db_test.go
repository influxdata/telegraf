package gorethink

import (
	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestDbCreate(c *test.C) {
	// Delete the test2 database if it already exists
	DBDrop("test").Exec(session)

	// Test database creation
	query := DBCreate("test")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.DBsCreated, jsonEquals, 1)
}

func (s *RethinkSuite) TestDbList(c *test.C) {
	var response []interface{}

	// create database
	DBCreate("test").Exec(session)

	// Try and find it in the list
	success := false
	res, err := DBList().Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.FitsTypeOf, []interface{}{})

	for _, db := range response {
		if db == "test" {
			success = true
		}
	}

	c.Assert(success, test.Equals, true)
}

func (s *RethinkSuite) TestDbDelete(c *test.C) {
	// Delete the test2 database if it already exists
	DBCreate("test").Exec(session)

	// Test database creation
	query := DBDrop("test")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.DBsDropped, jsonEquals, 1)

	// Ensure that there is still a test DB after the test has finished
	DBCreate("test").Exec(session)
}
