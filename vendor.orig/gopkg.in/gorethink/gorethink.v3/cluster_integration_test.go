// +build cluster
// +build integration

package gorethink

import (
	"time"

	test "gopkg.in/check.v1"
)

func (s *RethinkSuite) TestClusterDetectNewNode(c *test.C) {
	session, err := Connect(ConnectOpts{
		Addresses:           []string{url, url2},
		DiscoverHosts:       true,
		NodeRefreshInterval: time.Second,
	})
	c.Assert(err, test.IsNil)

	t := time.NewTimer(time.Second * 30)
	for {
		select {
		// Fail if deadline has passed
		case <-t.C:
			c.Fatal("No node was added to the cluster")
		default:
			// Pass if another node was added
			if len(session.cluster.GetNodes()) >= 3 {
				return
			}
		}
	}
}

func (s *RethinkSuite) TestClusterRecoverAfterNoNodes(c *test.C) {
	session, err := Connect(ConnectOpts{
		Addresses:           []string{url, url2},
		DiscoverHosts:       true,
		NodeRefreshInterval: time.Second,
	})
	c.Assert(err, test.IsNil)

	t := time.NewTimer(time.Second * 30)
	hasHadZeroNodes := false
	for {
		select {
		// Fail if deadline has passed
		case <-t.C:
			c.Fatal("No node was added to the cluster")
		default:
			// Check if there are no nodes
			if len(session.cluster.GetNodes()) == 0 {
				hasHadZeroNodes = true
			}

			// Pass if another node was added
			if len(session.cluster.GetNodes()) >= 1 && hasHadZeroNodes {
				return
			}
		}
	}
}

func (s *RethinkSuite) TestClusterNodeHealth(c *test.C) {
	session, err := Connect(ConnectOpts{
		Addresses:           []string{url1, url2, url3},
		DiscoverHosts:       true,
		NodeRefreshInterval: time.Second,
		InitialCap:          50,
		MaxOpen:             200,
	})
	c.Assert(err, test.IsNil)

	attempts := 0
	failed := 0
	seconds := 0

	t := time.NewTimer(time.Second * 30)
	tick := time.NewTicker(time.Second)
	for {
		select {
		// Fail if deadline has passed
		case <-tick.C:
			seconds++
			c.Logf("%ds elapsed", seconds)
		case <-t.C:
			// Execute queries for 10s and check that at most 5% of the queries fail
			c.Logf("%d of the %d(%d%%) queries failed", failed, attempts, (failed / attempts))
			c.Assert(failed <= 100, test.Equals, true)
			return
		default:
			attempts++
			if err := Expr(1).Exec(session); err != nil {
				c.Logf("Query failed, %s", err)
				failed++
			}
		}
	}
}
