package smc_ipmi

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
			"USERID:PASSW0RD@(192.168.1.1)",
			&Connection{
				Hostname: "192.168.1.1",
				Username: "USERID",
				Password: "PASSW0RD",
			},
		},
	}

	for _, v := range testData {
		assert.Equal(t, v.con, NewConnection(v.addr))
	}
}
