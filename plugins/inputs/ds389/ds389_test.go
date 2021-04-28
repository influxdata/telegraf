package ds389

import (
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test389dsLdap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:     testutil.GetLocalHost(),
		Protocol: "ldap",
		Port:     389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func Test389dsStartTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:     testutil.GetLocalHost(),
		Protocol: "starttls",
		Port:     389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func Test389dsNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:     "nosuchhost",
		Protocol: "ldap",
		Port:     389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.Error(t, err)
	assert.Zero(t, acc.NFields())
	assert.NotEmpty(t, acc.Errors)
}

func Test389dsGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host: testutil.GetLocalHost(),
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func Test389dsLDAPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		Protocol:           "ldaps",
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	commonTests(t, o, &acc)
}

func Test389dsLdbmAttrs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:         testutil.GetLocalHost(),
		Port:         389,
		BindDn:       "cn=Directory manager",
		BindPassword: "secret",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	assert.True(t, acc.HasInt64Field("ds389", "dbcachehitratio"), "Has an integer field called dbcachehitratio")
}

func Test389dsNetscapeRootDbAttrs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:         testutil.GetLocalHost(),
		Port:         389,
		BindDn:       "cn=Directory manager",
		BindPassword: "secret", //should be something else
		Dbtomonitor:  []string{"userRoot"},
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	assert.True(t, acc.HasInt64Field("ds389", "userroot_dncachehits"), "Has an integer field called userroot_dncachehits")
}

func Test389dsInvalidTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host:               testutil.GetLocalHost(),
		Port:               636,
		Protocol:           "invalid",
		InsecureSkipVerify: true,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	assert.Zero(t, acc.NFields())
	assert.NotEmpty(t, acc.Errors)
}

func commonTests(t *testing.T, o *ds389, acc *testutil.Accumulator) {
	assert.Empty(t, acc.Errors, "accumulator had no errors")
	assert.True(t, acc.HasMeasurement("ds389"), "Has a measurement called 'ds389'")
	assert.Equal(t, o.Host, acc.TagValue("ds389", "server"), "Has a tag value of server=o.Host")
	assert.Equal(t, strconv.Itoa(o.Port), acc.TagValue("ds389", "port"), "Has a tag value of port=o.Port")
	assert.True(t, acc.HasInt64Field("ds389", "totalconnections"), "Has an integer field called totalconnections")
}
