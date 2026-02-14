package aliyuncms

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/influxdata/telegraf"
	common_aliyun "github.com/influxdata/telegraf/plugins/common/aliyun"
)

// discoveryTool wraps the common library's DiscoveryTool for CMS
type discoveryTool struct {
	*common_aliyun.DiscoveryTool
}

// newDiscoveryTool creates a discovery tool for CMS using the common library
func newDiscoveryTool(
	regions []string,
	project string,
	lg telegraf.Logger,
	credential auth.Credential,
	rateLimit int,
	discoveryInterval time.Duration,
) (*discoveryTool, error) {
	var (
		responseRootKey       string
		responseObjectIDKey   string
		err                   error
		noDiscoverySupportErr = fmt.Errorf("no discovery support for project %q", project)
	)

	if len(regions) == 0 {
		regions = common_aliyun.DefaultRegions()
		lg.Infof("'regions' is not provided! Discovery data will be queried across %d regions:\n%s",
			len(regions), strings.Join(regions, ","))
	}

	if rateLimit == 0 {
		rateLimit = 1
	}

	dscReq := make(map[string]common_aliyun.DiscoveryRequest, len(regions))
	cli := make(map[string]common_aliyun.AliyunSdkClient, len(regions))
	for _, region := range regions {
		switch project {
		case "acs_ecs_dashboard":
			dscReq[region] = ecs.CreateDescribeInstancesRequest()
			responseRootKey = "Instances"
			responseObjectIDKey = "InstanceId"
		case "acs_rds_dashboard":
			dscReq[region] = rds.CreateDescribeDBInstancesRequest()
			responseRootKey = "Items"
			responseObjectIDKey = "DBInstanceId"
		case "acs_slb_dashboard":
			dscReq[region] = slb.CreateDescribeLoadBalancersRequest()
			responseRootKey = "LoadBalancers"
			responseObjectIDKey = "LoadBalancerId"
		case "acs_memcache":
			return nil, noDiscoverySupportErr
		case "acs_ocs":
			return nil, noDiscoverySupportErr
		case "acs_oss":
			// oss is really complicated and has its own format
			return nil, noDiscoverySupportErr
		case "acs_vpc_eip":
			dscReq[region] = vpc.CreateDescribeEipAddressesRequest()
			responseRootKey = "EipAddresses"
			responseObjectIDKey = "AllocationId"
		case "acs_kvstore":
			return nil, noDiscoverySupportErr
		case "acs_mns_new":
			return nil, noDiscoverySupportErr
		case "acs_cdn":
			// API replies are in its own format.
			return nil, noDiscoverySupportErr
		case "acs_polardb":
			return nil, noDiscoverySupportErr
		case "acs_gdb":
			return nil, noDiscoverySupportErr
		case "acs_ads":
			return nil, noDiscoverySupportErr
		case "acs_mongodb":
			return nil, noDiscoverySupportErr
		case "acs_express_connect":
			return nil, noDiscoverySupportErr
		case "acs_fc":
			return nil, noDiscoverySupportErr
		case "acs_nat_gateway":
			return nil, noDiscoverySupportErr
		case "acs_sls_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_containerservice_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_vpn":
			return nil, noDiscoverySupportErr
		case "acs_bandwidth_package":
			return nil, noDiscoverySupportErr
		case "acs_cen":
			return nil, noDiscoverySupportErr
		case "acs_ens":
			return nil, noDiscoverySupportErr
		case "acs_opensearch":
			return nil, noDiscoverySupportErr
		case "acs_scdn":
			return nil, noDiscoverySupportErr
		case "acs_drds":
			return nil, noDiscoverySupportErr
		case "acs_iot":
			return nil, noDiscoverySupportErr
		case "acs_directmail":
			return nil, noDiscoverySupportErr
		case "acs_elasticsearch":
			return nil, noDiscoverySupportErr
		case "acs_ess_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_streamcompute":
			return nil, noDiscoverySupportErr
		case "acs_global_acceleration":
			return nil, noDiscoverySupportErr
		case "acs_hitsdb":
			return nil, noDiscoverySupportErr
		case "acs_kafka":
			return nil, noDiscoverySupportErr
		case "acs_openad":
			return nil, noDiscoverySupportErr
		case "acs_pcdn":
			return nil, noDiscoverySupportErr
		case "acs_dcdn":
			return nil, noDiscoverySupportErr
		case "acs_petadata":
			return nil, noDiscoverySupportErr
		case "acs_videolive":
			return nil, noDiscoverySupportErr
		case "acs_hybriddb":
			return nil, noDiscoverySupportErr
		case "acs_adb":
			return nil, noDiscoverySupportErr
		case "acs_mps":
			return nil, noDiscoverySupportErr
		case "acs_maxcompute_prepay":
			return nil, noDiscoverySupportErr
		case "acs_hdfs":
			return nil, noDiscoverySupportErr
		case "acs_ddh":
			return nil, noDiscoverySupportErr
		case "acs_hbr":
			return nil, noDiscoverySupportErr
		case "acs_hdr":
			return nil, noDiscoverySupportErr
		case "acs_cds":
			return nil, noDiscoverySupportErr
		default:
			return nil, fmt.Errorf("project %q is not recognized by discovery", project)
		}

		cli[region], err = sdk.NewClientWithOptions(region, sdk.NewConfig(), credential)
		if err != nil {
			return nil, err
		}
	}

	if len(dscReq) == 0 || len(cli) == 0 {
		return nil, fmt.Errorf("can't build discovery request for project: %q, regions: %v", project, regions)
	}

	dt := &common_aliyun.DiscoveryTool{
		Req:                dscReq,
		Cli:                cli,
		RespRootKey:        responseRootKey,
		RespObjectIDKey:    responseObjectIDKey,
		RateLimit:          rateLimit,
		Interval:           discoveryInterval,
		ReqDefaultPageSize: 20,
		DataChan:           make(chan map[string]interface{}, 1),
		Lg:                 lg,
	}

	return &discoveryTool{DiscoveryTool: dt}, nil
}

// start begins the discovery loop
func (dt *discoveryTool) start() {
	dt.Start()
}

// stop stops the discovery loop
func (dt *discoveryTool) stop() {
	dt.Stop()
}
