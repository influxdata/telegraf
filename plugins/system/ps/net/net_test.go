package net

import (
	"fmt"
	"testing"
)

func TestAddrString(t *testing.T) {
	v := Addr{IP: "192.168.0.1", Port: 8000}

	s := fmt.Sprintf("%v", v)
	if s != "{\"ip\":\"192.168.0.1\",\"port\":8000}" {
		t.Errorf("Addr string is invalid: %v", v)
	}
}

func TestNetIOCountersStatString(t *testing.T) {
	v := NetIOCountersStat{
		Name:      "test",
		BytesSent: 100,
	}
	e := `{"name":"test","bytes_sent":100,"bytes_recv":0,"packets_sent":0,"packets_recv":0,"errin":0,"errout":0,"dropin":0,"dropout":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("NetIOCountersStat string is invalid: %v", v)
	}
}

func TestNetConnectionStatString(t *testing.T) {
	v := NetConnectionStat{
		Fd:     10,
		Family: 10,
		Type:   10,
	}
	e := `{"fd":10,"family":10,"type":10,"localaddr":{"ip":"","port":0},"remoteaddr":{"ip":"","port":0},"status":"","pid":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("NetConnectionStat string is invalid: %v", v)
	}

}

func TestNetIOCountersAll(t *testing.T) {
	v, err := NetIOCounters(false)
	per, err := NetIOCounters(true)
	if err != nil {
		t.Errorf("Could not get NetIOCounters: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("Could not get NetIOCounters: %v", v)
	}
	if v[0].Name != "all" {
		t.Errorf("Invalid NetIOCounters: %v", v)
	}
	var pr uint64
	for _, p := range per {
		pr += p.PacketsRecv
	}
	if v[0].PacketsRecv != pr {
		t.Errorf("invalid sum value: %v, %v", v[0].PacketsRecv, pr)
	}
}

func TestNetIOCountersPerNic(t *testing.T) {
	v, err := NetIOCounters(true)
	if err != nil {
		t.Errorf("Could not get NetIOCounters: %v", err)
	}
	if len(v) == 0 {
		t.Errorf("Could not get NetIOCounters: %v", v)
	}
	for _, vv := range v {
		if vv.Name == "" {
			t.Errorf("Invalid NetIOCounters: %v", vv)
		}
	}
}

func Test_getNetIOCountersAll(t *testing.T) {
	n := []NetIOCountersStat{
		NetIOCountersStat{
			Name:        "a",
			BytesRecv:   10,
			PacketsRecv: 10,
		},
		NetIOCountersStat{
			Name:        "b",
			BytesRecv:   10,
			PacketsRecv: 10,
			Errin:       10,
		},
	}
	ret, err := getNetIOCountersAll(n)
	if err != nil {
		t.Error(err)
	}
	if len(ret) != 1 {
		t.Errorf("invalid return count")
	}
	if ret[0].Name != "all" {
		t.Errorf("invalid return name")
	}
	if ret[0].BytesRecv != 20 {
		t.Errorf("invalid count bytesrecv")
	}
	if ret[0].Errin != 10 {
		t.Errorf("invalid count errin")
	}
}

func TestNetInterfaces(t *testing.T) {
	v, err := NetInterfaces()
	if err != nil {
		t.Errorf("Could not get NetInterfaceStat: %v", err)
	}
	if len(v) == 0 {
		t.Errorf("Could not get NetInterfaceStat: %v", err)
	}
	for _, vv := range v {
		if vv.Name == "" {
			t.Errorf("Invalid NetInterface: %v", vv)
		}
	}
}
