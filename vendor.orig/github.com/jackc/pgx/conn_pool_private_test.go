package pgx

import (
	"testing"
)

func compareConnSlices(slice1, slice2 []*Conn) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, c := range slice1 {
		if c != slice2[i] {
			return false
		}
	}
	return true
}

func TestConnPoolRemoveFromAllConnections(t *testing.T) {
	t.Parallel()
	pool := ConnPool{}
	conn1 := &Conn{}
	conn2 := &Conn{}
	conn3 := &Conn{}

	// First element
	pool.allConnections = []*Conn{conn1, conn2, conn3}
	pool.removeFromAllConnections(conn1)
	if !compareConnSlices(pool.allConnections, []*Conn{conn2, conn3}) {
		t.Fatal("First element test failed")
	}
	// Element somewhere in the middle
	pool.allConnections = []*Conn{conn1, conn2, conn3}
	pool.removeFromAllConnections(conn2)
	if !compareConnSlices(pool.allConnections, []*Conn{conn1, conn3}) {
		t.Fatal("Middle element test failed")
	}
	// Last element
	pool.allConnections = []*Conn{conn1, conn2, conn3}
	pool.removeFromAllConnections(conn3)
	if !compareConnSlices(pool.allConnections, []*Conn{conn1, conn2}) {
		t.Fatal("Last element test failed")
	}
}
