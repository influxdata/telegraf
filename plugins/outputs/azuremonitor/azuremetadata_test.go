package azuremonitor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetMetadata(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}
	metadata, err := azureMetadata.GetInstanceMetadata()

	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.NotEmpty(t, metadata.AzureResourceID)
	require.NotEmpty(t, metadata.Compute.Location)

	// if err != nil {
	// 	t.Logf("could not get metadata: %v\n", err)
	// } else {
	// 	t.Logf("resource id  \n%s", metadata.AzureResourceID)
	// 	t.Logf("metadata is \n%v", metadata)
	// }

	//fmt.Printf("metadata is \n%v", metadata)
}

func TestGetTOKEN(t *testing.T) {
	azureMetadata := &AzureInstanceMetadata{}

	resourceID := "https://ingestion.monitor.azure.com/"
	token, err := azureMetadata.GetMsiToken("", resourceID)

	require.NoError(t, err)
	require.NotEmpty(t, token.AccessToken)
	require.EqualValues(t, token.Resource, resourceID)

	t.Logf("token is %+v\n", token)
	t.Logf("expiry time is %s\n", token.ExpiresAt().Format(time.RFC3339))
	t.Logf("expiry duration is %s\n", token.ExpiresInDuration().String())
	t.Logf("resource is %s\n", token.Resource)

}
