package gorethink

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestSelectGet(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("Table1").Get(6)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 6, "g1": 1, "g2": 1, "num": 15})

	res.Close()
}

func (s *RethinkSuite) TestSelectGetAll(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)
	DB("test").Table("Table1").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table1").GetAll(6).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 15, "id": 6, "g2": 1, "g1": 1},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectGetAllMultiple(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)
	DB("test").Table("Table1").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table1").GetAll(1, 2, 3).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 0, "id": 1, "g2": 1, "g1": 1},
		map[string]interface{}{"num": 5, "id": 2, "g2": 2, "g1": 2},
		map[string]interface{}{"num": 10, "id": 3, "g2": 2, "g1": 3},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectGetAllByIndex(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)
	DB("test").Table("Table1").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("Table1").GetAllByIndex("num", 15).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 6, "g1": 1, "g2": 1, "num": 15})

	res.Close()
}

func (s *RethinkSuite) TestSelectGetAllMultipleByIndex(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table2").Exec(session)
	DB("test").Table("Table2").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table2").Insert(objList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("Table2").GetAllByIndex("num", 15).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 6, "g1": 1, "g2": 1, "num": 15})

	res.Close()
}

func (s *RethinkSuite) TestSelectGetAllCompoundIndex(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableDrop("TableCompound").Exec(session)
	DB("test").TableCreate("TableCompound").Exec(session)
	write, err := DB("test").Table("TableCompound").IndexCreateFunc("full_name", func(row Term) interface{} {
		return []interface{}{row.Field("first_name"), row.Field("last_name")}
	}).RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(write.Created, test.Equals, 1)

	// Insert rows
	DB("test").Table("TableCompound").Insert(nameList).Exec(session)

	// Test query
	var response interface{}
	query := DB("test").Table("TableCompound").GetAllByIndex("full_name", []interface{}{"John", "Smith"})
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, map[string]interface{}{"id": 1, "first_name": "John", "last_name": "Smith", "gender": "M"})

	res.Close()
}

func (s *RethinkSuite) TestSelectBetween(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table1").Between(1, 3).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 0, "id": 1, "g2": 1, "g1": 1},
		map[string]interface{}{"num": 5, "id": 2, "g2": 2, "g1": 2},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectBetweenWithIndex(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table2").Exec(session)
	DB("test").Table("Table2").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table2").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table2").Between(10, 50, BetweenOpts{
		Index: "num",
	}).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 10, "id": 3, "g2": 2, "g1": 3},
		map[string]interface{}{"num": 15, "id": 6, "g2": 1, "g1": 1},
		map[string]interface{}{"num": 25, "id": 9, "g2": 3, "g1": 2},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectBetweenWithOptions(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table2").Exec(session)
	DB("test").Table("Table2").IndexCreate("num").Exec(session)

	// Insert rows
	DB("test").Table("Table2").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table2").Between(10, 50, BetweenOpts{
		Index:      "num",
		RightBound: "closed",
	}).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 10, "id": 3, "g2": 2, "g1": 3},
		map[string]interface{}{"num": 15, "id": 6, "g2": 1, "g1": 1},
		map[string]interface{}{"num": 50, "id": 8, "g2": 2, "g1": 4},
		map[string]interface{}{"num": 25, "id": 9, "g2": 3, "g1": 2},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectFilterImplicit(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table1").Filter(Row.Field("num").Ge(50)).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 100, "id": 5, "g2": 3, "g1": 2},
		map[string]interface{}{"num": 50, "id": 8, "g2": 2, "g1": 4},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectFilterFunc(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").Exec(session)
	DB("test").TableCreate("Table1").Exec(session)

	// Insert rows
	DB("test").Table("Table1").Insert(objList).Exec(session)

	// Test query
	var response []interface{}
	query := DB("test").Table("Table1").Filter(func(row Term) Term {
		return row.Field("num").Ge(50)
	}).OrderBy("id")
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.All(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, jsonEquals, []interface{}{
		map[string]interface{}{"num": 100, "id": 5, "g2": 3, "g1": 2},
		map[string]interface{}{"num": 50, "id": 8, "g2": 2, "g1": 4},
	})

	res.Close()
}

func (s *RethinkSuite) TestSelectManyRows(c *test.C) {
	// Ensure table + database exist
	DBCreate("test").RunWrite(session)
	DB("test").TableCreate("TestMany").RunWrite(session)
	DB("test").Table("TestMany").Delete().RunWrite(session)

	// Insert rows
	for i := 0; i < 100; i++ {
		data := []interface{}{}

		for j := 0; j < 100; j++ {
			data = append(data, map[string]interface{}{
				"i": i,
				"j": j,
			})
		}

		DB("test").Table("TestMany").Insert(data).RunWrite(session)
	}

	// Test query
	res, err := DB("test").Table("TestMany").Run(session, RunOpts{
		MaxBatchRows: 1,
	})
	c.Assert(err, test.IsNil)

	var n int
	var response map[string]interface{}
	for res.Next(&response) {
		n++
	}

	c.Assert(res.Err(), test.IsNil)
	c.Assert(n, test.Equals, 10000)

	res.Close()
}

func (s *RethinkSuite) TestConcurrentSelectManyWorkers(c *test.C) {
	if testing.Short() {
		c.Skip("Skipping long test")
	}

	rand.Seed(time.Now().UnixNano())
	sess, _ := Connect(ConnectOpts{
		Address: url,
		AuthKey: authKey,
		MaxOpen: 200,
		MaxIdle: 200,
	})

	// Ensure table + database exist
	DBCreate("test").RunWrite(sess)
	DB("test").TableDrop("TestConcurrent").RunWrite(sess)
	DB("test").TableCreate("TestConcurrent").RunWrite(sess)
	DB("test").TableDrop("TestConcurrent2").RunWrite(sess)
	DB("test").TableCreate("TestConcurrent2").RunWrite(sess)

	// Insert rows
	for j := 0; j < 200; j++ {
		DB("test").Table("TestConcurrent").Insert(map[string]interface{}{
			"id": j,
			"i":  j,
		}).Exec(sess)
		DB("test").Table("TestConcurrent2").Insert(map[string]interface{}{
			"j": j,
			"k": j * 2,
		}).Exec(sess)
	}

	// Test queries concurrently
	numQueries := 1000
	numWorkers := 10
	queryChan := make(chan int)
	doneChan := make(chan error)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for _ = range queryChan {
				res, err := DB("test").Table("TestConcurrent2").EqJoin("j", DB("test").Table("TestConcurrent")).Zip().Run(sess)
				if err != nil {
					doneChan <- err
					return
				}

				var response []map[string]interface{}
				err = res.All(&response)
				if err != nil {
					doneChan <- err
					return
				}
				if err := res.Close(); err != nil {
					doneChan <- err
					return
				}

				if len(response) != 200 {
					doneChan <- fmt.Errorf("expected response length 200, received %d", len(response))
					return
				}

				res, err = DB("test").Table("TestConcurrent").Get(response[rand.Intn(len(response))]["id"]).Run(sess)
				if err != nil {
					doneChan <- err
					return
				}

				err = res.All(&response)
				if err != nil {
					doneChan <- err
					return
				}
				if err := res.Close(); err != nil {
					doneChan <- err
					return
				}

				if len(response) != 1 {
					doneChan <- fmt.Errorf("expected response length 1, received %d", len(response))
					return
				}

				doneChan <- nil
			}
		}()
	}

	go func() {
		for i := 0; i < numQueries; i++ {
			queryChan <- i
		}
	}()

	for i := 0; i < numQueries; i++ {
		ret := <-doneChan
		if ret != nil {
			c.Fatalf("non-nil error returned (%s)", ret)
		}
	}
}

func (s *RethinkSuite) TestConcurrentSelectManyRows(c *test.C) {
	if testing.Short() {
		c.Skip("Skipping long test")
	}

	// Ensure table + database exist
	DBCreate("test").RunWrite(session)
	DB("test").TableCreate("TestMany").RunWrite(session)
	DB("test").Table("TestMany").Delete().RunWrite(session)

	// Insert rows
	for i := 0; i < 100; i++ {
		DB("test").Table("TestMany").Insert(map[string]interface{}{
			"i": i,
		}).Exec(session)
	}

	// Test queries concurrently
	attempts := 10
	waitChannel := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		go func(i int, c chan error) {
			res, err := DB("test").Table("TestMany").Run(session)
			if err != nil {
				c <- err
				return
			}

			var response []map[string]interface{}
			err = res.All(&response)
			if err != nil {
				c <- err
				return
			}

			if len(response) != 100 {
				c <- fmt.Errorf("expected response length 100, received %d", len(response))
				return
			}

			res.Close()

			c <- nil
		}(i, waitChannel)
	}

	for i := 0; i < attempts; i++ {
		ret := <-waitChannel
		if ret != nil {
			c.Fatalf("non-nil error returned (%s)", ret)
		}
	}
}
