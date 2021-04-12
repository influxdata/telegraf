package hddtemp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	l := serve(t, []byte("|/dev/sda|foobar|36|C|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())
	require.NoError(t, err)

	expected := []Disk{
		{
			DeviceName:  "sda",
			Model:       "foobar",
			Temperature: 36,
			Unit:        "C",
		},
	}
	require.Equal(t, expected, disks, "disks' slice is different from expected")
}

func TestFetchWrongAddress(t *testing.T) {
	_, err := New().Fetch("127.0.0.1:1")
	require.Error(t, err)
}

func TestFetchStatus(t *testing.T) {
	l := serve(t, []byte("|/dev/sda|foobar|SLP|C|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())
	require.NoError(t, err)

	expected := []Disk{
		{
			DeviceName:  "sda",
			Model:       "foobar",
			Temperature: 0,
			Unit:        "C",
			Status:      "SLP",
		},
	}
	require.Equal(t, expected, disks, "disks' slice is different from expected")
}

func TestFetchTwoDisks(t *testing.T) {
	l := serve(t, []byte("|/dev/hda|ST380011A|46|C||/dev/hdd|ST340016A|SLP|*|"))
	defer l.Close()

	disks, err := New().Fetch(l.Addr().String())
	require.NoError(t, err)

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
	require.Equal(t, expected, disks, "disks' slice is different from expected")
}

func serve(t *testing.T, data []byte) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func(t *testing.T) {
		conn, err := l.Accept()
		require.NoError(t, err)

		_, err = conn.Write(data)
		require.NoError(t, err)
		require.NoError(t, conn.Close())
	}(t)

	return l
}
