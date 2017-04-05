package openldap

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenldapNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap {
		Host: "nosuchhost",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err) // test that we handled the error
	require.Zero(t, acc.NFields()) // test that we didn't get anything back
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

