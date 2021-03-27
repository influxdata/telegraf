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
	p.ImdsTags = []string{"accountId", "instanceId"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Start(acc))
	require.NoError(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestBasicStartupWithEC2Tags(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	p.EC2Tags = []string{"Name"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Start(acc))
	require.NoError(t, p.Stop())

	require.Len(t, acc.GetTelegrafMetrics(), 0)
	require.Len(t, acc.Errors, 0)
}

func TestBasicInitNoTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{}
	err := p.Init()
	require.Error(t, err)
}

func TestBasicInitInvalidTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"dummy", "qwerty"}
	err := p.Init()
	require.Error(t, err)
}

func TestLoadingConfig(t *testing.T) {
	confFile := []byte("[[processors.aws_ec2]]" + "\n" + sampleConfig)
	c := config.NewConfig()
	err := c.LoadConfigData(confFile)
	require.NoError(t, err)

	require.Len(t, c.Processors, 1)
}
