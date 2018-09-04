package kazoo

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	// By default, assume we're using Sarama's vagrant cluster when running tests
	zookeeperPeers []string = []string{"192.168.100.67:2181", "192.168.100.67:2182", "192.168.100.67:2183", "192.168.100.67:2184", "192.168.100.67:2185"}
)

func init() {
	if zookeeperPeersEnv := os.Getenv("ZOOKEEPER_PEERS"); zookeeperPeersEnv != "" {
		zookeeperPeers = strings.Split(zookeeperPeersEnv, ",")
	}

	fmt.Printf("Using Zookeeper cluster at %v\n", zookeeperPeers)
}

func TestBrokers(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	brokers, err := kz.Brokers()
	if err != nil {
		t.Fatal(err)
	}

	if len(brokers) == 0 {
		t.Error("Expected at least one broker")
	}

	for id, addr := range brokers {
		if conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond); err != nil {
			t.Errorf("Failed to connect to Kafka broker %d at %s", id, addr)
		} else {
			_ = conn.Close()
		}
	}

	assertSuccessfulClose(t, kz)
}

func TestBrokerList(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	brokers, err := kz.BrokerList()
	if err != nil {
		t.Fatal(err)
	}

	if len(brokers) == 0 {
		t.Error("Expected at least one broker")
	}

	for _, addr := range brokers {
		if conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond); err != nil {
			t.Errorf("Failed to connect to Kafka broker at %s", addr)
		} else {
			_ = conn.Close()
		}
	}

	assertSuccessfulClose(t, kz)
}

func TestController(t *testing.T) {
	kz, err := NewKazoo(zookeeperPeers, nil)
	if err != nil {
		t.Fatal(err)
	}

	brokers, err := kz.Brokers()
	if err != nil {
		t.Fatal(err)
	}

	controller, err := kz.Controller()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := brokers[controller]; !ok {
		t.Error("Expected the controller's BrokerID to be an existing one")
	}

	assertSuccessfulClose(t, kz)
}
