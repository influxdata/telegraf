package openldap

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenldapNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap {
		Host: "127.0.0.1",
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err) // test that we didn't return an error
	assert.Zero(t, acc.NFields()) // test that we set an error
	assert.NotEmpty(t, acc.Errors) // test that we set an error
}

func TestOpenldapGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap {
		Host: testutil.GetLocalHost(),
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	// do more stuff
}

func TestOpenldapStartTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	o := &Openldap {
		Host: testutil.GetLocalHost(),
		Tls: true,
		TlsSkipverify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
}

func TestOpenldapBind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap {
		Host: testutil.GetLocalHost(),
		Tls: true,
		TlsSkipverify: true,
		BindDn: "cn=manager,cn=config",
		BindPassword: "secret",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
}

