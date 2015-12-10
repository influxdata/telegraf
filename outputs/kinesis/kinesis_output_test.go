package kinesis_output

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
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
	err := kinesis.New(session.New(Config))


	require.NoError(t, err)
}
