package aliyun

import (
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/stretchr/testify/require"
)

// Mock types for testing
type mockDiscoveryRequest struct {
	*requests.RpcRequest
}

func TestDiscoveryToolFields(t *testing.T) {
	dt := &DiscoveryTool{
		Req:                map[string]DiscoveryRequest{},
		Cli:                map[string]AliyunSdkClient{},
		RespRootKey:        "TestRoot",
		RespObjectIDKey:    "TestID",
		RateLimit:          100,
		ReqDefaultPageSize: 20,
		DataChan:           make(chan map[string]interface{}, 1),
	}

	require.NotNil(t, dt.Req)
	require.NotNil(t, dt.Cli)
	require.Equal(t, "TestRoot", dt.RespRootKey)
	require.Equal(t, "TestID", dt.RespObjectIDKey)
	require.Equal(t, 100, dt.RateLimit)
	require.Equal(t, 20, dt.ReqDefaultPageSize)
	require.NotNil(t, dt.DataChan)
}

func TestParsedDiscoveryResponseFields(t *testing.T) {
	parsed := &ParsedDiscoveryResponse{
		Data:       []interface{}{map[string]interface{}{"id": "test"}},
		TotalCount: 10,
		PageSize:   5,
		PageNumber: 2,
	}

	require.Len(t, parsed.Data, 1)
	require.Equal(t, 10, parsed.TotalCount)
	require.Equal(t, 5, parsed.PageSize)
	require.Equal(t, 2, parsed.PageNumber)
}

func TestGetRPCReqFromDiscoveryRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         DiscoveryRequest
		expectError bool
	}{
		{
			name: "valid RPC request",
			req: &mockDiscoveryRequest{
				RpcRequest: &requests.RpcRequest{},
			},
			expectError: false,
		},
		{
			name:        "nil request",
			req:         (*mockDiscoveryRequest)(nil),
			expectError: true,
		},
		{
			name: "struct without RpcRequest field",
			req: &struct {
				Field string
			}{Field: "test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcReq, err := getRPCReqFromDiscoveryRequest(tt.req)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, rpcReq)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rpcReq)
			}
		})
	}
}

func TestDiscoveryInterfaces(t *testing.T) {
	var req DiscoveryRequest

	req = &mockDiscoveryRequest{RpcRequest: &requests.RpcRequest{}}
	require.NotNil(t, req)

	req = "string"
	require.NotNil(t, req)

	req = 123
	require.NotNil(t, req)
}

func TestDiscoveryToolCreation(t *testing.T) {
	dataChan := make(chan map[string]interface{}, 10)

	dt := &DiscoveryTool{
		Req:                make(map[string]DiscoveryRequest),
		Cli:                make(map[string]AliyunSdkClient),
		RespRootKey:        "Instances",
		RespObjectIDKey:    "InstanceId",
		RateLimit:          50,
		ReqDefaultPageSize: 100,
		DataChan:           dataChan,
	}

	require.NotNil(t, dt)
	require.Empty(t, dt.Req)
	require.Empty(t, dt.Cli)
	require.Equal(t, "Instances", dt.RespRootKey)
	require.Equal(t, "InstanceId", dt.RespObjectIDKey)
	require.Equal(t, 50, dt.RateLimit)
	require.Equal(t, 100, dt.ReqDefaultPageSize)
	require.Equal(t, dataChan, dt.DataChan)
}

func TestDiscoveryRequestMapOperations(t *testing.T) {
	reqMap := make(map[string]DiscoveryRequest)

	reqMap["region1"] = &mockDiscoveryRequest{RpcRequest: &requests.RpcRequest{}}
	reqMap["region2"] = "string-request"

	require.Len(t, reqMap, 2)
	require.NotNil(t, reqMap["region1"])
	require.NotNil(t, reqMap["region2"])

	delete(reqMap, "region1")
	require.Len(t, reqMap, 1)
}
