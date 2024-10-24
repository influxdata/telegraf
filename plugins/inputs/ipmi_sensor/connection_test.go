package ipmi_sensor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	testData := []struct {
		addr string
		con  *connection
	}{
		{
			"USERID:PASSW0RD@lan(192.168.1.1)",
			&connection{
				hostname:  "192.168.1.1",
				username:  "USERID",
				password:  "PASSW0RD",
				intf:      "lan",
				privilege: "USER",
				hexKey:    "0001",
			},
		},
		{
			"USERID:PASS:!@#$%^&*(234)_+W0RD@lan(192.168.1.1)",
			&connection{
				hostname:  "192.168.1.1",
				username:  "USERID",
				password:  "PASS:!@#$%^&*(234)_+W0RD",
				intf:      "lan",
				privilege: "USER",
				hexKey:    "0001",
			},
		},
		// test connection doesn't panic if incorrect symbol used
		{
			"USERID@PASSW0RD@lan(192.168.1.1)",
			&connection{
				hostname:  "192.168.1.1",
				username:  "",
				password:  "",
				intf:      "lan",
				privilege: "USER",
				hexKey:    "0001",
			},
		},
	}

	for _, v := range testData {
		require.EqualValues(t, v.con, newConnection(v.addr, "USER", "0001"))
	}
}

func TestGetCommandOptions(t *testing.T) {
	testData := []struct {
		connection *connection
		options    []string
	}{
		{
			&connection{
				hostname:  "192.168.1.1",
				username:  "user",
				password:  "password",
				intf:      "lan",
				privilege: "USER",
				hexKey:    "0001",
			},
			[]string{"-H", "192.168.1.1", "-U", "user", "-P", "password", "-I", "lan", "-y", "0001", "-L", "USER"},
		},
		{
			&connection{
				hostname:  "192.168.1.1",
				username:  "user",
				password:  "password",
				intf:      "lan",
				privilege: "USER",
				hexKey:    "",
			},
			[]string{"-H", "192.168.1.1", "-U", "user", "-P", "password", "-I", "lan", "-L", "USER"},
		},
	}

	for _, data := range testData {
		require.EqualValues(t, data.options, data.connection.options())
	}
}
