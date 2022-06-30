package ipmi_sensor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	testData := []struct {
		addr string
		con  *Connection
	}{
		{
			"USERID:PASSW0RD@lan(192.168.1.1)",
			&Connection{
				Hostname:  "192.168.1.1",
				Username:  "USERID",
				Password:  "PASSW0RD",
				Interface: "lan",
				Privilege: "USER",
				HexKey:    "0001",
			},
		},
		{
			"USERID:PASS:!@#$%^&*(234)_+W0RD@lan(192.168.1.1)",
			&Connection{
				Hostname:  "192.168.1.1",
				Username:  "USERID",
				Password:  "PASS:!@#$%^&*(234)_+W0RD",
				Interface: "lan",
				Privilege: "USER",
				HexKey:    "0001",
			},
		},
		// test connection doesn't panic if incorrect symbol used
		{
			"USERID@PASSW0RD@lan(192.168.1.1)",
			&Connection{
				Hostname:  "192.168.1.1",
				Username:  "",
				Password:  "",
				Interface: "lan",
				Privilege: "USER",
				HexKey:    "0001",
			},
		},
	}

	for _, v := range testData {
		require.EqualValues(t, v.con, NewConnection(v.addr, "USER", "0001"))
	}
}

func TestGetCommandOptions(t *testing.T) {
	testData := []struct {
		connection *Connection
		options    []string
	}{
		{
			&Connection{
				Hostname:  "192.168.1.1",
				Username:  "user",
				Password:  "password",
				Interface: "lan",
				Privilege: "USER",
				HexKey:    "0001",
			},
			[]string{"-H", "192.168.1.1", "-U", "user", "-P", "password", "-I", "lan", "-y", "0001", "-L", "USER"},
		},
		{
			&Connection{
				Hostname:  "192.168.1.1",
				Username:  "user",
				Password:  "password",
				Interface: "lan",
				Privilege: "USER",
				HexKey:    "",
			},
			[]string{"-H", "192.168.1.1", "-U", "user", "-P", "password", "-I", "lan", "-L", "USER"},
		},
	}

	for _, data := range testData {
		require.EqualValues(t, data.options, data.connection.options())
	}
}
