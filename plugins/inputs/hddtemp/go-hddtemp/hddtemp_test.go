package hddtemp

import (
	"net"
	"reflect"
	"testing"
)

func TestFetch(t *testing.T) {
	l := serve(t, []byte("|/dev/sda|foobar|36|C|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())

	if err != nil {
		t.Error("expecting err to be nil")
	}

	expected := []Disk{
		{
			DeviceName:  "sda",
			Model:       "foobar",
			Temperature: 36,
			Unit:        "C",
		},
	}

	if !reflect.DeepEqual(expected, disks) {
		t.Error("disks' slice is different from expected")
	}
}

func TestFetchWrongAddress(t *testing.T) {
	_, err := New().Fetch("127.0.0.1:1")

	if err == nil {
		t.Error("expecting err to be non-nil")
	}
}

func TestFetchStatus(t *testing.T) {
	l := serve(t, []byte("|/dev/sda|foobar|SLP|C|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())

	if err != nil {
		t.Error("expecting err to be nil")
	}

	expected := []Disk{
		{
			DeviceName:  "sda",
			Model:       "foobar",
			Temperature: 0,
			Unit:        "C",
			Status:      "SLP",
		},
	}

	if !reflect.DeepEqual(expected, disks) {
		t.Error("disks' slice is different from expected")
	}
}

func TestFetchTwoDisks(t *testing.T) {
	l := serve(t, []byte("|/dev/hda|ST380011A|46|C||/dev/hdd|ST340016A|SLP|*|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())

	if err != nil {
		t.Error("expecting err to be nil")
	}

	expected := []Disk{
		{
			DeviceName:  "hda",
			Model:       "ST380011A",
			Temperature: 46,
			Unit:        "C",
		},
		{
			DeviceName:  "hdd",
			Model:       "ST340016A",
			Temperature: 0,
			Unit:        "*",
			Status:      "SLP",
		},
	}

	if !reflect.DeepEqual(expected, disks) {
		t.Error("disks' slice is different from expected")
	}
}

func serve(t *testing.T, data []byte) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatal(err)
	}

	go func(t *testing.T) {
		conn, err := l.Accept()

		if err != nil {
			t.Fatal(err)
		}

		conn.Write(data)
		conn.Close()
	}(t)

	return l
}
