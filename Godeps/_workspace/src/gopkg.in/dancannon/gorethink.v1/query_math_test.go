package gorethink

import (
	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestMathAdd(c *test.C) {
	query := Expr(1).Add(2)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 3)
}

func (s *RethinkSuite) TestMathSub(c *test.C) {
	query := Expr(2).Sub(1)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 1)
}

func (s *RethinkSuite) TestMathSubNegative(c *test.C) {
	query := Expr(1).Sub(2)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, -1)
}

func (s *RethinkSuite) TestMathMul(c *test.C) {
	query := Expr(5).Mul(4)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 20)
}

func (s *RethinkSuite) TestMathDiv(c *test.C) {
	query := Expr(8).Div(4)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 2)
}

func (s *RethinkSuite) TestMathMod(c *test.C) {
	query := Expr(7).Mod(2)

	var response int
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 1)
}

func (s *RethinkSuite) TestMathEqTrue(c *test.C) {
	query := Expr(1).Eq(1)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestMathEqFalse(c *test.C) {
	query := Expr(1).Eq(2)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestMathEqStringTrue(c *test.C) {
	query := Expr("test").Eq("test")

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestCompareLt(c *test.C) {
	query := Expr(2).Lt(1)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestCompareLe(c *test.C) {
	query := Expr(2).Le(1)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestCompareLeEqual(c *test.C) {
	query := Expr(2).Le(2)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestCompareGt(c *test.C) {
	query := Expr(2).Gt(1)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestCompareGe(c *test.C) {
	query := Expr(2).Ge(1)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestCompareGeEqual(c *test.C) {
	query := Expr(2).Le(2)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestBoolNotTrue(c *test.C) {
	query := Expr(true).Not()

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, false)
}

func (s *RethinkSuite) TestBoolAnd(c *test.C) {
	query := Expr(true).And(true)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestBoolOr(c *test.C) {
	query := Expr(true).Or(false)

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}

func (s *RethinkSuite) TestBoolDeMorgan(c *test.C) {
	query := Expr(true).And(false).Eq(Expr(true).Not().Or(Expr(false).Not()).Not())

	var response bool
	r, err := query.Run(session)
	c.Assert(err, test.IsNil)

	err = r.One(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}
