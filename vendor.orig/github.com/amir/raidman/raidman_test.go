package raidman

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestTCP(t *testing.T) {
	c, err := Dial("tcp", "localhost:5555")
	if err != nil {
		t.Fatal(err.Error())
	}
	var event = &Event{
		State:      "success",
		Host:       "raidman",
		Service:    "tcp",
		Metric:     42,
		Ttl:        1,
		Tags:       []string{"tcp", "test", "raidman"},
		Attributes: map[string]string{"type": "test"},
	}

	err = c.Send(event)
	if err != nil {
		t.Error(err.Error())
	}

	events, err := c.Query("tagged \"test\"")
	if err != nil {
		t.Error(err.Error())
	}

	if len(events) < 1 {
		t.Error("Submitted event not found")
	}

	testAttributeExists := false
	for _, event := range events {
		if val, ok := event.Attributes["type"]; ok && val == "test" {
			testAttributeExists = true
		}
	}

	if !testAttributeExists {
		t.Error("Attribute \"type\" is missing")
	}

	c.Close()
}

func TestMultiTCP(t *testing.T) {
	c, err := Dial("tcp", "localhost:5555")
	if err != nil {
		t.Fatal(err.Error())
	}

	err = c.SendMulti([]*Event{
		&Event{
			State:      "success",
			Host:       "raidman",
			Service:    "tcp-multi-1",
			Metric:     42,
			Ttl:        1,
			Tags:       []string{"tcp", "test", "raidman", "multi"},
			Attributes: map[string]string{"type": "test"},
		},
		&Event{
			State:      "success",
			Host:       "raidman",
			Service:    "tcp-multi-2",
			Metric:     42,
			Ttl:        1,
			Tags:       []string{"tcp", "test", "raidman", "multi"},
			Attributes: map[string]string{"type": "test"},
		},
	})
	if err != nil {
		t.Error(err.Error())
	}

	events, err := c.Query("tagged \"test\" and tagged \"multi\"")
	if err != nil {
		t.Error(err.Error())
	}

	if len(events) != 2 {
		t.Error("Submitted event not found")
	}

	c.Close()
}

func TestMetricIsInt64(t *testing.T) {
	c, err := Dial("tcp", "localhost:5555")
	if err != nil {
		t.Fatal(err.Error())
	}

	var int64metric int64 = 9223372036854775807

	var event = &Event{
		State:      "success",
		Host:       "raidman",
		Service:    "tcp",
		Metric:     int64metric,
		Ttl:        1,
		Tags:       []string{"tcp", "test", "raidman"},
		Attributes: map[string]string{"type": "test"},
	}

	err = c.Send(event)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestUDP(t *testing.T) {
	c, err := Dial("udp", "localhost:5555")
	if err != nil {
		t.Fatal(err.Error())
	}
	var event = &Event{
		State:   "warning",
		Host:    "raidman",
		Service: "udp",
		Metric:  3.4,
		Ttl:     10.7,
	}

	err = c.Send(event)
	if err != nil {
		t.Error(err.Error())
	}
	c.Close()
}

func TestTCPWithoutHost(t *testing.T) {
	c, err := Dial("tcp", "localhost:5555")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer c.Close()

	var event = &Event{
		State:   "success",
		Service: "tcp-host-not-set",
		Ttl:     5,
	}

	err = c.Send(event)
	if err != nil {
		t.Error(err.Error())
	}

	events, err := c.Query("service = \"tcp-host-not-set\"")
	if err != nil {
		t.Error(err.Error())
	}

	if len(events) < 1 {
		t.Error("Submitted event not found")
	}

	for _, e := range events {
		if e.Host == "" {
			t.Error("Default host name is not set")
		}
	}
}

func TestIsZero(t *testing.T) {
	event := &Event{
		Time: 1,
	}
	elem := reflect.ValueOf(event).Elem()
	eventType := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		name := eventType.Field(i).Name
		if name == "Time" {
			if isZero(field) {
				t.Error("Time should not be zero")
			}
		} else {
			if !isZero(field) {
				t.Errorf("%s should be zero", name)
			}
		}
	}
}

func TestDialer(t *testing.T) {
	proxyAddr := "localhost:9999"
	os.Setenv("RIEMANN_PROXY", "socks5://"+proxyAddr)
	defer os.Unsetenv("RIEMANN_PROXY")
	dialer, err := newDialer()
	if err != nil {
		t.Error(err.Error())
	}
	val := reflect.Indirect(reflect.ValueOf(dialer))
	// this is a horrible hack but proxy.Dialer exports nothing.
	addr := fmt.Sprintf("%s", val.FieldByName("addr"))
	if addr != proxyAddr {
		t.Errorf("RIEMANN_PROXY is set and is %s but dialer's proxy is %s", proxyAddr, addr)
	}
}

func BenchmarkTCP(b *testing.B) {
	c, err := Dial("tcp", "localhost:5555")

	var event = &Event{
		State:   "good",
		Host:    "raidman",
		Service: "benchmark",
	}

	if err == nil {
		for i := 0; i < b.N; i++ {
			c.Send(event)
		}
	}
	c.Close()
}

func BenchmarkUDP(b *testing.B) {
	c, err := Dial("udp", "localhost:5555")

	var event = &Event{
		State:   "good",
		Host:    "raidman",
		Service: "benchmark",
	}

	if err == nil {
		for i := 0; i < b.N; i++ {
			c.Send(event)
		}
	}
	c.Close()
}

func BenchmarkConcurrentTCP(b *testing.B) {
	c, err := Dial("tcp", "localhost:5555")

	var event = &Event{
		Host:    "raidman",
		Service: "tcp_concurrent",
		Tags:    []string{"concurrent", "tcp", "benchmark"},
	}

	ch := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		go func(metric int) {
			event.Metric = metric
			err = c.Send(event)
			ch <- i
		}(i)
	}
	<-ch

	c.Close()
}

func BenchmarkConcurrentUDP(b *testing.B) {
	c, err := Dial("udp", "localhost:5555")

	var event = &Event{
		Host:    "raidman",
		Service: "udp_concurrent",
		Tags:    []string{"concurrent", "udp", "benchmark"},
	}

	ch := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		go func(metric int) {
			event.Metric = metric
			err = c.Send(event)
			ch <- i
		}(i)
	}
	<-ch

	c.Close()
}
