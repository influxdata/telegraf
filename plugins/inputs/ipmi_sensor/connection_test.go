package ipmi_sensor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type conTest struct {
	Got  string
	Want *Connection
}

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
			},
		},
	}

	for _, v := range testData {
		assert.Equal(t, v.con, NewConnection(v.addr, "USER"))
	}
}
