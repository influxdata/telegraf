package kazoo

import (
	"testing"
)

func TestBuildConnectionString(t *testing.T) {
	nodes := []string{"zk1:2181", "zk2:2181", "zk3:2181"}

	if str := BuildConnectionString(nodes); str != "zk1:2181,zk2:2181,zk3:2181" {
		t.Errorf("The connection string was not built correctly: %s", str)
	}

	if str := BuildConnectionStringWithChroot(nodes, "/chroot"); str != "zk1:2181,zk2:2181,zk3:2181/chroot" {
		t.Errorf("The connection string was not built correctly: %s", str)
	}
}

func TestParseConnectionString(t *testing.T) {
	var (
		nodes  []string
		chroot string
	)

	nodes, chroot = ParseConnectionString("zookeeper/chroot")
	if len(nodes) != 1 || nodes[0] != "zookeeper" {
		t.Error("Parsed nodes incorrectly:", nodes)
	}
	if chroot != "/chroot" {
		t.Error("Parsed chroot incorrectly:", chroot)
	}

	nodes, chroot = ParseConnectionString("zk1:2181,zk2:2181,zk3:2181")
	if len(nodes) != 3 || nodes[0] != "zk1:2181" || nodes[1] != "zk2:2181" || nodes[2] != "zk3:2181" {
		t.Error("Parsed nodes incorrectly:", nodes)
	}
	if chroot != "" {
		t.Error("Parsed chroot incorrectly:", chroot)
	}

	nodes, chroot = ParseConnectionString("zk1:2181,zk2/nested/chroot")
	if len(nodes) != 2 || nodes[0] != "zk1:2181" || nodes[1] != "zk2" {
		t.Error("Parsed nodes incorrectly:", nodes)
	}
	if chroot != "/nested/chroot" {
		t.Error("Parsed chroot incorrectly:", chroot)
	}

	nodes, chroot = ParseConnectionString("")
	if len(nodes) != 1 || nodes[0] != "" {
		t.Error("Parsed nodes incorrectly:", nodes)
	}
	if chroot != "" {
		t.Error("Parsed chroot incorrectly:", chroot)
	}
}
