package kinesis_output

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Verify that we can connect kinesis endpoint. This test allows for a chain of credential
	// so that various authentication methods can pass depending on the system that executes.
	Config := &aws.Config{
		Region: aws.String("us-west-1"),
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
	resp, err := svc.ListStreams(KinesisParams)

	fmt.Println(resp)

	require.NoError(t, err)
}
