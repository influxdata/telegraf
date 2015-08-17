package gorethink

import (
	"time"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestTimeTime(c *test.C) {
	var response time.Time
	res, err := Time(1986, 11, 3, 12, 30, 15, "Z").Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response.Equal(time.Date(1986, 11, 3, 12, 30, 15, 0, time.UTC)), test.Equals, true)
}

func (s *RethinkSuite) TestTimeTimeMillisecond(c *test.C) {
	var response time.Time
	res, err := Time(1986, 11, 3, 12, 30, 15.679, "Z").Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response.Equal(time.Date(1986, 11, 3, 12, 30, 15, 679.00002*1000*1000, time.UTC)), test.Equals, true)
}

func (s *RethinkSuite) TestTimeEpochTime(c *test.C) {
	var response time.Time
	res, err := EpochTime(531360000).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response.Equal(time.Date(1986, 11, 3, 0, 0, 0, 0, time.UTC)), test.Equals, true)
}

func (s *RethinkSuite) TestTimeExpr(c *test.C) {
	var response time.Time
	t := time.Unix(531360000, 0)
	res, err := Expr(Expr(t)).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
}

func (s *RethinkSuite) TestTimeExprMillisecond(c *test.C) {
	var response time.Time
	t := time.Unix(531360000, 679000000)
	res, err := Expr(t).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)
	c.Assert(err, test.IsNil)
	c.Assert(float64(response.UnixNano()), test.Equals, float64(t.UnixNano()))
}

func (s *RethinkSuite) TestTimeISO8601(c *test.C) {
	var t1, t2 time.Time
	t2, _ = time.Parse("2006-01-02T15:04:05-07:00", "1986-11-03T08:30:00-07:00")
	res, err := ISO8601("1986-11-03T08:30:00-07:00").Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&t1)
	c.Assert(err, test.IsNil)
	c.Assert(t1.Equal(t2), test.Equals, true)
}

func (s *RethinkSuite) TestTimeInTimezone(c *test.C) {
	loc, err := time.LoadLocation("MST")
	c.Assert(err, test.IsNil)
	var response []time.Time
	res, err2 := Expr([]interface{}{Now(), Now().InTimezone("-07:00")}).Run(session)
	c.Assert(err2, test.IsNil)

	err = res.All(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response[1].Equal(response[0].In(loc)), test.Equals, true)
}

func (s *RethinkSuite) TestTimeBetween(c *test.C) {
	var response interface{}

	times := Expr([]interface{}{
		Time(1986, 9, 3, 12, 30, 15, "Z"),
		Time(1986, 10, 3, 12, 30, 15, "Z"),
		Time(1986, 11, 3, 12, 30, 15, "Z"),
		Time(1986, 12, 3, 12, 30, 15, "Z"),
	})
	res, err := times.Filter(func(row Term) Term {
		return row.During(Time(1986, 9, 3, 12, 30, 15, "Z"), Time(1986, 11, 3, 12, 30, 15, "Z"))
	}).Count().Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(int(response.(float64)), test.Equals, 2)
}

func (s *RethinkSuite) TestTimeYear(c *test.C) {
	var response interface{}

	res, err := Time(1986, 12, 3, 12, 30, 15, "Z").Year().Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(int(response.(float64)), test.Equals, 1986)
}

func (s *RethinkSuite) TestTimeMonth(c *test.C) {
	var response interface{}

	res, err := Time(1986, 12, 3, 12, 30, 15, "Z").Month().Eq(December).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response.(bool), test.Equals, true)
}

func (s *RethinkSuite) TestTimeDay(c *test.C) {
	var response interface{}

	res, err := Time(1986, 12, 3, 12, 30, 15, "Z").Day().Eq(Wednesday).Run(session)
	c.Assert(err, test.IsNil)

	err = res.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response.(bool), test.Equals, true)
}
