package aliyunrds

import (
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"github.com/influxdata/telegraf"
	common_aliyun "github.com/influxdata/telegraf/plugins/common/aliyun"
)

// discoveryTool wraps the common library's DiscoveryTool for RDS
type discoveryTool struct {
	*common_aliyun.DiscoveryTool
}

// newDiscoveryTool creates a discovery tool for RDS instances using the common library
func newDiscoveryTool(
	regions []string,
	lg telegraf.Logger,
	credential auth.Credential,
	rateLimit int,
	discoveryInterval time.Duration,
) (*discoveryTool, error) {
	if len(regions) == 0 {
		regions = common_aliyun.DefaultRegions()
		lg.Infof("'regions' is not provided! Discovery data will be queried across %d regions:\n%s",
			len(regions), strings.Join(regions, ","))
	}

	if rateLimit == 0 {
		rateLimit = 1
	}

	// Create discovery requests per region
	dscReq := make(map[string]common_aliyun.DiscoveryRequest, len(regions))
	cli := make(map[string]common_aliyun.AliyunSdkClient, len(regions))

	var err error
	for _, region := range regions {
		dscReq[region] = rds.CreateDescribeDBInstancesRequest()
		cli[region], err = sdk.NewClientWithOptions(region, sdk.NewConfig(), credential)
		if err != nil {
			return nil, err
		}
	}

	dt := &common_aliyun.DiscoveryTool{
		Req:                dscReq,
		Cli:                cli,
		RespRootKey:        "Items",
		RespObjectIDKey:    "DBInstanceId",
		RateLimit:          rateLimit,
		Interval:           discoveryInterval,
		ReqDefaultPageSize: 20,
		DataChan:           make(chan map[string]interface{}, 1),
		Lg:                 lg,
	}

	return &discoveryTool{DiscoveryTool: dt}, nil
}

// getDiscoveryDataAcrossRegions retrieves discovery data across all regions
func (dt *discoveryTool) getDiscoveryDataAcrossRegions(lmtr chan bool) (map[string]interface{}, error) {
	return dt.GetDiscoveryDataAcrossRegions(lmtr)
}

// start begins the discovery loop
func (dt *discoveryTool) start() {
	dt.Start()
}

// stop stops the discovery loop
func (dt *discoveryTool) stop() {
	dt.Stop()
}
