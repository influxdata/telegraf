package ec2

import (
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasicStartup(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.Tags = []string{"accountId", "instanceId"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Start(acc))
	require.NoError(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestBasicStartupInvalidTags(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.Tags = []string{"dummy", "qwerty"}
	acc := &testutil.Accumulator{}
	require.Error(t, p.Start(acc))
	require.Error(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestLoadingConfig(t *testing.T) {
	const conf = `[[processors.aws_ec2]]`

	c := config.NewConfig()
	err := c.LoadConfigData([]byte("[[processors.aws_ec2]]" + "\n" + sampleConfig))
	require.NoError(t, err)

	require.Len(t, c.Processors, 1)
}
