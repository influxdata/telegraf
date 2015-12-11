package kinesis_output

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	k := &KinesisOutput{
		Region: "us-west-2",
	}

	// Verify that we can connect kinesis endpoint. This test allows for a chain of credential
	// so that various authentication methods can pass depending on the system that executes.
	Config := &aws.Config{
		Region: aws.String(k.Region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}),
	}
	svc := kinesis.New(session.New(Config))

	KinesisParams := &kinesis.ListStreamsInput{
		Limit: aws.Int64(1)}
	_, err := svc.ListStreams(KinesisParams)

	if err != nil {
		t.Error("Unable to connect to Kinesis")
	}
	require.NoError(t, err)
}

func TestFormatMetric(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	k := &KinesisOutput{
		Format: "string",
	}

	p := testutil.MockBatchPoints().Points()[0]

	valid_string := "test1,tag1=value1 value=1 1257894000000000000"
	func_string, err := FormatMetric(k, p)

	if func_string != valid_string {
		t.Error("Expected ", valid_string)
	}
	require.NoError(t, err)

	k = &KinesisOutput{
		Format: "custom",
	}

	valid_custom := "test1,map[tag1:value1],test1,tag1=value1 value=1 1257894000000000000"
	func_custom, err := FormatMetric(k, p)

	if func_custom != valid_custom {
		t.Error("Expected ", valid_custom)
	}
	require.NoError(t, err)
}
