package ldap_response

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ldap "gopkg.in/ldap.v2"
)

func TestLdapMockResult(t *testing.T) {
	var acc testutil.Accumulator

	mockSearchResult := ldap.SearchResult{
		Entries: []*ldap.Entry{
			{
				DN:         "cn=manager,cn=config",
				Attributes: []*ldap.EntryAttribute{{Name: "someEntry", Values: []string{"1"}}},
			},
		},
		Referrals: []string{},
		Controls:  []ldap.Control{},
	}

	l := &Ldap{
		Host:             "localhost",
		Port:             389,
		SearchAttributes: []string{"someEntry"},
	}

	fields := map[string]interface{}{
		"query_time_ms":   float64(2),
		"connect_time_ms": float64(2),
		"bind_time_ms":    float64(2),
		"total_time_ms":   float64(2),
	}
	gatherSearchResult(fields, &mockSearchResult, l, &acc)
	commonTests(t, l, &acc)
}

func TestLdapBind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	l := &Ldap{
		Host:               testutil.GetLocalHost(),
		Port:               389,
		Ssl:                "",
		InsecureSkipVerify: true,
		BindDn:             "cn=manager,cn=config",
		BindPassword:       "secret",
		SearchBase:         "cn=Monitor",
	}

	var acc testutil.Accumulator
	err := l.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, l, &acc)
}

func commonTests(t *testing.T, l *Ldap, acc *testutil.Accumulator) {
	assert.Empty(t, acc.Errors, "Expecting accumulator to have no errors")
	assert.True(t, acc.HasFloatField("ldap_response", "connect_time_ms"), "Expeting connect_time_ms field to be present")
	assert.True(t, acc.HasFloatField("ldap_response", "bind_time_ms"), "Expeting bind_time_ms field to be present")
	assert.True(t, acc.HasFloatField("ldap_response", "query_time_ms"), "Expeting query_time_ms field to be present")
	assert.True(t, acc.HasFloatField("ldap_response", "total_time_ms"), "Expeting total_time_ms field to be present")
	assert.True(t, acc.HasMeasurement("ldap_response"), "Expecting a measurement called 'ldap_response'")
	assert.Equal(t, l.Host, acc.TagValue("ldap_response", "server"), fmt.Sprintf("Expecting a tag value of server=%v", l.Host))
	assert.Equal(t, strconv.Itoa(l.Port), acc.TagValue("ldap_response", "port"), fmt.Sprintf("Expecting a tag value of port=%v", l.Port))
}

func TestLdapNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Ldap{
		Host: "nosuchhost",
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)        // test that we didn't return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that we set an error
}
func TestLdapStartTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Ldap{
		Host:               testutil.GetLocalHost(),
		Port:               389,
		Ssl:                "starttls",
		InsecureSkipVerify: true,
		SearchBase:         "cn=Monitor",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestLdapLDAPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Ldap{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		Ssl:                "ldaps",
		InsecureSkipVerify: true,
		SearchBase:         "cn=Monitor",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func TestLdapInvalidSSL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &Ldap{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		Ssl:                "invalid",
		InsecureSkipVerify: true,
		SearchBase:         "cn=Monitor",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)        // test that we didn't return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that we set an error
}
