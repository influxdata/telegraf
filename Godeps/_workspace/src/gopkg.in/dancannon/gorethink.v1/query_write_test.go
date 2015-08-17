package gorethink

import (
	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestWriteInsert(c *test.C) {
	query := DB("test").Table("test").Insert(map[string]interface{}{"num": 1})
	_, err := query.Run(session)
	c.Assert(err, test.IsNil)
}

func (s *RethinkSuite) TestWriteInsertChanges(c *test.C) {
	query := DB("test").Table("test").Insert([]interface{}{
		map[string]interface{}{"num": 1},
		map[string]interface{}{"num": 2},
	}, InsertOpts{ReturnChanges: true})
	res, err := query.RunWrite(session)
	c.Assert(err, test.IsNil)
	c.Assert(res.Inserted, test.Equals, 2)
	c.Assert(len(res.Changes), test.Equals, 2)
}

func (s *RethinkSuite) TestWriteInsertStruct(c *test.C) {
	var response map[string]interface{}
	o := object{
		Name: "map[string]interface{}ect 3",
		Attrs: []attr{
			attr{
				Name:  "Attr 2",
				Value: "Value",
			},
		},
	}

	query := DB("test").Table("test").Insert(o)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response["inserted"], test.Equals, float64(1))
}

func (s *RethinkSuite) TestWriteInsertStructPointer(c *test.C) {
	var response map[string]interface{}
	o := object{
		Name: "map[string]interface{}ect 3",
		Attrs: []attr{
			attr{
				Name:  "Attr 2",
				Value: "Value",
			},
		},
	}

	query := DB("test").Table("test").Insert(&o)
	res, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response["inserted"], test.Equals, float64(1))
}

func (s *RethinkSuite) TestWriteUpdate(c *test.C) {
	query := DB("test").Table("test").Insert(map[string]interface{}{"num": 1})
	_, err := query.Run(session)
	c.Assert(err, test.IsNil)

	// Update the first row in the table
	query = DB("test").Table("test").Sample(1).Update(map[string]interface{}{"num": 2})
	_, err = query.Run(session)
	c.Assert(err, test.IsNil)
}

func (s *RethinkSuite) TestWriteReplace(c *test.C) {
	query := DB("test").Table("test").Insert(map[string]interface{}{"num": 1})
	_, err := query.Run(session)
	c.Assert(err, test.IsNil)

	// Replace the first row in the table
	query = DB("test").Table("test").Sample(1).Update(map[string]interface{}{"num": 2})
	_, err = query.Run(session)
	c.Assert(err, test.IsNil)
}

func (s *RethinkSuite) TestWriteDelete(c *test.C) {
	query := DB("test").Table("test").Insert(map[string]interface{}{"num": 1})
	_, err := query.Run(session)
	c.Assert(err, test.IsNil)

	// Delete the first row in the table
	query = DB("test").Table("test").Sample(1).Delete()
	_, err = query.Run(session)
	c.Assert(err, test.IsNil)
}
