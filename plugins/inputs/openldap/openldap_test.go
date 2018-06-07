package openldap

import (
	"strconv"
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
				DN:         "cn=Total,cn=Connections,cn=Monitor",
				Attributes: []*ldap.EntryAttribute{{Name: "monitorCounter", Values: []string{"1"}}},
			},
		},
		Referrals: []string{},
		Controls:  []ldap.Control{},
	}

	o := &Openldap{
		Host: "localhost",
		Port: 389,
	}

	gatherSearchResult(&mockSearchResult, o, &acc)
	commonTests(t, o, &acc)
}

func TestOpenldapNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host: "nosuchhost",
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)        // test that we didn't return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that we set an error
}

func TestOpenldapGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host: testutil.GetLocalHost(),
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapStartTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host:               testutil.GetLocalHost(),
		Port:               389,
		SSL:                "starttls",
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapLDAPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		SSL:                "ldaps",
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestOpenldapInvalidSSL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		SSL:                "invalid",
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)        // test that we didn't return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that we set an error
}

func TestOpenldapBind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host:               testutil.GetLocalHost(),
		Port:               389,
		SSL:                "",
		InsecureSkipVerify: true,
		BindDn:             "cn=manager,cn=config",
		BindPassword:       "secret",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func commonTests(t *testing.T, o *Openldap, acc *testutil.Accumulator) {
	assert.Empty(t, acc.Errors, "accumulator had no errors")
	assert.True(t, acc.HasMeasurement("openldap"), "Has a measurement called 'openldap'")
	assert.Equal(t, o.Host, acc.TagValue("openldap", "server"), "Has a tag value of server=o.Host")
	assert.Equal(t, strconv.Itoa(o.Port), acc.TagValue("openldap", "port"), "Has a tag value of port=o.Port")
	assert.True(t, acc.HasInt64Field("openldap", "total_connections"), "Has an integer field called total_connections")
}

func TestOpenldapReverseMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Openldap{
		Host:               testutil.GetLocalHost(),
		Port:               389,
		SSL:                "",
		InsecureSkipVerify: true,
		BindDn:             "cn=manager,cn=config",
		BindPassword:       "secret",
		ReverseMetricNames: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	assert.True(t, acc.HasInt64Field("openldap", "connections_total"), "Has an integer field called connections_total")
}
