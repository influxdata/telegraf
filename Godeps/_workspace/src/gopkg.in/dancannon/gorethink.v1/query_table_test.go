package gorethink

import (
	"sync"
	"time"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestTableCreate(c *test.C) {
	DB("test").TableDrop("test").Exec(session)

	// Test database creation
	query := DB("test").TableCreate("test")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.TablesCreated, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableCreatePrimaryKey(c *test.C) {
	DB("test").TableDrop("testOpts").Exec(session)

	// Test database creation
	query := DB("test").TableCreate("testOpts", TableCreateOpts{
		PrimaryKey: "it",
	})

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.TablesCreated, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableCreateSoftDurability(c *test.C) {
	DB("test").TableDrop("testOpts").Exec(session)

	// Test database creation
	query := DB("test").TableCreate("testOpts", TableCreateOpts{
		Durability: "soft",
	})

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.TablesCreated, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableCreateSoftMultipleOpts(c *test.C) {
	DB("test").TableDrop("testOpts").Exec(session)

	// Test database creation
	query := DB("test").TableCreate("testOpts", TableCreateOpts{
		PrimaryKey: "it",
		Durability: "soft",
	})

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.TablesCreated, jsonEquals, 1)

	DB("test").TableDrop("test").Exec(session)
}

func (s *RethinkSuite) TestTableList(c *test.C) {
	var response []interface{}

	DB("test").TableCreate("test").Exec(session)

	// Try and find it in the list
	success := false
	res, err := DB("test").TableList().Run(session)
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

func (s *RethinkSuite) TestTableDelete(c *test.C) {
	DB("test").TableCreate("test").Exec(session)

	// Test database creation
	query := DB("test").TableDrop("test")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.TablesDropped, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableIndexCreate(c *test.C) {
	DB("test").TableCreate("test").Exec(session)
	DB("test").Table("test").IndexDrop("test").Exec(session)

	// Test database creation
	query := DB("test").Table("test").IndexCreate("test", IndexCreateOpts{
		Multi: true,
	})

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.Created, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableCompoundIndexCreate(c *test.C) {
	DBCreate("test").Exec(session)
	DB("test").TableDrop("TableCompound").Exec(session)
	DB("test").TableCreate("TableCompound").Exec(session)
	response, err := DB("test").Table("TableCompound").IndexCreateFunc("full_name", func(row Term) interface{} {
		return []interface{}{row.Field("first_name"), row.Field("last_name")}
	}).RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.Created, test.Equals, 1)
}

func (s *RethinkSuite) TestTableIndexList(c *test.C) {
	var response []interface{}

	DB("test").TableCreate("test").Exec(session)
	DB("test").Table("test").IndexCreate("test").Exec(session)

	// Try and find it in the list
	success := false
	res, err := DB("test").Table("test").IndexList().Run(session)
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

func (s *RethinkSuite) TestTableIndexDelete(c *test.C) {
	DB("test").TableCreate("test").Exec(session)
	DB("test").Table("test").IndexCreate("test").Exec(session)

	// Test database creation
	query := DB("test").Table("test").IndexDrop("test")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.Dropped, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableIndexRename(c *test.C) {
	DB("test").TableDrop("test").Exec(session)
	DB("test").TableCreate("test").Exec(session)
	DB("test").Table("test").IndexCreate("test").Exec(session)

	// Test index rename
	query := DB("test").Table("test").IndexRename("test", "test2")

	response, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(response.Renamed, jsonEquals, 1)
}

func (s *RethinkSuite) TestTableChanges(c *test.C) {
	DB("test").TableDrop("changes").Exec(session)
	DB("test").TableCreate("changes").Exec(session)

	var n int

	res, err := DB("test").Table("changes").Changes().Run(session)
	if err != nil {
		c.Fatal(err.Error())
	}
	c.Assert(res.Type(), test.Equals, "Feed")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Use goroutine to wait for changes. Prints the first 10 results
	go func() {
		var response interface{}
		for n < 10 && res.Next(&response) {
			n++
		}

		if res.Err() != nil {
			c.Fatal(res.Err())
		}

		wg.Done()
	}()

	DB("test").Table("changes").Insert(map[string]interface{}{"n": 1}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 2}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 3}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 4}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 5}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 6}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 7}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 8}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 9}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 10}).Exec(session)

	wg.Wait()

	c.Assert(n, test.Equals, 10)
}

func (s *RethinkSuite) TestTableChangesExit(c *test.C) {
	DB("test").TableDrop("changes").Exec(session)
	DB("test").TableCreate("changes").Exec(session)

	var n int

	res, err := DB("test").Table("changes").Changes().Run(session)
	if err != nil {
		c.Fatal(err.Error())
	}
	c.Assert(res.Type(), test.Equals, "Feed")

	change := make(chan ChangeResponse)

	// Close cursor after one second
	go func() {
		<-time.After(time.Second)
		res.Close()
	}()

	// Insert 5 docs
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 1}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 2}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 3}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 4}).Exec(session)
	DB("test").Table("changes").Insert(map[string]interface{}{"n": 5}).Exec(session)

	// Listen for changes
	res.Listen(change)
	for _ = range change {
		n++
	}
	if res.Err() != nil {
		c.Fatal(res.Err())
	}

	c.Assert(n, test.Equals, 5)
}
