package ds389

import (
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test389dsStartTLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
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

func Test389dsNoConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
		Host: "nosuchhost",
		Port: 389,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)        // test that we didn't return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that we set an error
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
		SSL:                "ldaps",
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
		BindPassword: "secret", //shoul be something else
		Dbtomonitor:  []string{"userRoot"},
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)
	require.NoError(t, err)
	assert.True(t, acc.HasInt64Field("ds389", "userroot_dncachehits"), "Has an integer field called userroot_dncachehits")
}

func Test389dsInvalidSSL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	o := &ds389{
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

func commonTests(t *testing.T, o *ds389, acc *testutil.Accumulator) {
	assert.Empty(t, acc.Errors, "accumulator had no errors")
	assert.True(t, acc.HasMeasurement("ds389"), "Has a measurement called 'ds389'")
	assert.Equal(t, o.Host, acc.TagValue("ds389", "server"), "Has a tag value of server=o.Host")
	assert.Equal(t, strconv.Itoa(o.Port), acc.TagValue("ds389", "port"), "Has a tag value of port=o.Port")
	assert.True(t, acc.HasInt64Field("ds389", "totalconnections"), "Has an integer field called totalconnections")
}
