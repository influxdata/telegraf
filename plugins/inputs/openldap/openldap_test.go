package openldap

import (
	"testing"
	"gopkg.in/ldap.v2"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenldapMockResult(t *testing.T) {
	var acc testutil.Accumulator

	mockSearchResult := ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN: "cn=Total,cn=Connections,cn=Monitor",
				Attributes: []*ldap.EntryAttribute{{Name: "monitorCounter", Values: []string{"1"}}},
			},
		},
		Referrals: []string{},
		Controls: []ldap.Control{},
	}

	o := &Openldap {
		Host: "localhost",
		Port: 389,
	}

	err := gatherSearchResult(&mockSearchResult, o, &acc)

	require.NoError(t, err)
	assert.NotZero(t, acc.NFields())
	assert.True(t, acc.HasFloatField("openldap", "total_connections"))
}

func TestOpenldapNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap {
		Host: testutil.GetLocalHost(),
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err) // test that we didn't return an error
	assert.Zero(t, acc.NFields()) // test that we didn't return any fields
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
	assert.Empty(t, acc.Errors)
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

func runTests(t *testing.T, acc *testutil.Accumulator) {
	assert.True(t, acc.HasMeasurement("openldap"))
	assert.True(t, acc.HasTag("host", testutil.GetLocalHost()))
	assert.True(t, acc.HasTag("port", "389"))
	assert.NotZero(t, acc.NFields())
}
