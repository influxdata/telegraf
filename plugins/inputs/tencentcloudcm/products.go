package tencentcloudcm

import (
	"fmt"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
)

// Product defines cloud product
type Product interface {
	Namespace() string // Tencent Cloud CM Product Namespace
	Discover(crs *common.Credential, region, endpoint string) (instances []*monitor.Instance, err error)
	Metrics() []string // Supported metrics
}

// CVM defines Cloud Virtual Machine, see: https://intl.cloud.tencent.com/document/product/213
type CVM struct{}

// Namespace implements Product interface
func (c CVM) Namespace() string {
	return "QCE/CVM"
}

func (c CVM) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*cvm.DescribeInstancesResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "cvm", endpoint)
	client, err := cvm.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("cvm.NewClient failed, error: %s", err)

	}

	request := cvm.NewDescribeInstancesRequest()

	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)

	response, err := client.DescribeInstances(request)
	if err != nil {
		return nil, fmt.Errorf("discover instances for namespace %s failed: %s", c.Namespace(), err)
	}

	return response, nil
}

// Discover implements Product interface
func (c CVM) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	cvmInstances := []*cvm.Instance{}
	instances := []*monitor.Instance{}

	response, err := c.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}
	cvmInstances = append(cvmInstances, response.Response.InstanceSet...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(*response.Response.TotalCount/limit)+1; i++ {
		response, err := c.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		cvmInstances = append(cvmInstances, response.Response.InstanceSet...)
	}

	for _, cvmInstance := range cvmInstances {

		if cvmInstance.InstanceId == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("InstanceId"),
					Value: cvmInstance.InstanceId,
				},
			},
		})
	}

	return instances, nil
}

// Metrics implements Product interface
func (c CVM) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/6843
	return []string{
		// CPU Monitor
		"CPUUsage",
		"CpuLoadavg", "Cpuloadavg5m", "Cpuloadavg15m",
		"BaseCpuUsage",
		// GPU Monitor
		"GpuMemTotal", "GpuMemUsage", "GpuMemUsed",
		"GpuPowDraw", "GpuPowLimit", "GpuPowUsage",
		"GpuTemp",
		"GpuUtil",
		// Network Monitor
		"LanOuttraffic", "LanIntraffic", "LanOutpkg", "LanInpkg",
		"WanOuttraffic", "WanIntraffic", "WanOutpkg", "WanInpkg",
		"AccOuttraffic",
		"TcpCurrEstab",
		"TimeOffset",
		// Memory Monitor
		"MemUsed", "MemUsage",
		// Disk Monitor
		"CvmDiskUsage",
	}
}

// CDB defines TencentDB for MySQL, see: https://intl.cloud.tencent.com/document/product/236
type CDB struct {
}

// Namespace implements Product interface
func (c CDB) Namespace() string {
	return "QCE/CDB"
}

func (c CDB) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*cdb.DescribeDBInstancesResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "cdb", endpoint)
	client, err := cdb.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("cdb.NewClient failed, error: %s", err)
	}

	request := cdb.NewDescribeDBInstancesRequest()

	request.Offset = common.Uint64Ptr(uint64(offset))
	request.Limit = common.Uint64Ptr(uint64(limit))

	response, err := client.DescribeDBInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("DescribeDBInstances an API error has returned: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("DescribeDBInstances failed, error: %s", err)
	}
	return response, nil
}

func Discover() {}

// Discover implements Product interface
func (c CDB) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	cdbInstances := []*cdb.InstanceInfo{}
	instances := []*monitor.Instance{}

	response, err := c.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}
	cdbInstances = append(cdbInstances, response.Response.Items...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(*response.Response.TotalCount/limit)+1; i++ {
		response, err := c.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		cdbInstances = append(cdbInstances, response.Response.Items...)
	}

	for _, cdbInstance := range cdbInstances {

		if cdbInstance.InstanceId == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("InstanceId"),
					Value: cdbInstance.InstanceId,
				},
			},
		})
	}

	return instances, nil
}

// Metrics implements Product interface
func (c CDB) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/11006
	return []string{
		// Resource Monitor
		"BytesReceived", "BytesSent", "Capacity", "CPUUseRate", "IOPS",
		"MemoryUse", "MemoryUseRate", "RealCapacity", "VolumeRate",
		// Engine Monitor - MyISAM
		"KeyCacheHitRate", "KeyCacheUseRate",
		// Engine Monitor - InnoDB
		"InnodbCacheHitRate", "InnodbCacheUseRate", "InnodbNumOpenFiles",
		"InnodbOsFileReads", "InnodbOsFileWrites", "InnodbOsFsyncs",
		// Engine Monitor - Connections
		"ConnectionUseRate", "MaxConnections", "Qps", "ThreadsConnected", "Tps",
		// Engine Monitor - Access
		"ComDelete", "ComInsert", "ComReplace", "ComUpdate",
		"Queries", "QueryRate",
		"SelectCount", "SelectScan", "SlowQueries",
		// Engine Monitor - Table
		"CreatedTmpTables", "TableLocksWaited",
	}
}

// Redis defines TencentDB for Redis, see: https://intl.cloud.tencent.com/document/product/239
type Redis struct {
}

// Namespace implements Product interface
func (r Redis) Namespace() string {
	return "QCE/REDIS"
}

func (r Redis) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*redis.DescribeInstancesResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "redis", endpoint)
	client, err := redis.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("redis.NewClient failed, error: %s", err)
	}

	request := redis.NewDescribeInstancesRequest()

	request.Limit = common.Uint64Ptr(uint64(limit))
	request.Offset = common.Uint64Ptr(uint64(offset))

	response, err := client.DescribeInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("DescribeInstances an API error has returned: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed, error: %s", err)
	}
	return response, nil
}

// Discover implements Product interface
func (r Redis) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	redisInstances := []*redis.InstanceSet{}
	instances := []*monitor.Instance{}

	response, err := r.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}
	redisInstances = append(redisInstances, response.Response.InstanceSet...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(*response.Response.TotalCount/limit)+1; i++ {
		response, err := r.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		redisInstances = append(redisInstances, response.Response.InstanceSet...)
	}

	for _, redisInstance := range redisInstances {

		if redisInstance.InstanceId == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("instanceid"),
					Value: redisInstance.InstanceId,
				},
			},
		})
	}

	return instances, nil
}

// Metrics implements Product interface
func (r Redis) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/34640
	return []string{
		// Redis Monitor
		"CpuUsMin", "StorageMin", "StorageUsMin",
		"KeysMin", "ExpiredKeysMin", "EvictedKeysMin",
		"ConnectionsMin", "ConnectionsUsMin",
		"InFlowMin", "InFlowUsMin", "OutFlowMin", "OutFlowUsMin",
		"LatencyMin", "LatencyGetMin", "LatencySetMin", "LatencyOtherMin",
		"QpsMin", "StatGetMin", "StatSetMin", "StatOtherMin",
		"BigValueMin", "SlowQueryMin", "StatSuccessMin", "StatMissedMin",
		"CmdErrMin", "CacheHitRatioMin",
	}
}

// LBPublic defines Cloud Load Balancer, see: https://intl.cloud.tencent.com/document/product/214
type LBPublic struct{}

// Namespace implements Product interface
func (l LBPublic) Namespace() string {
	return "QCE/LB_PUBLIC"
}

func (l LBPublic) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*clb.DescribeLoadBalancersResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "clb", endpoint)
	client, err := clb.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("clb.NewClient failed, error: %s", err)
	}

	request := clb.NewDescribeLoadBalancersRequest()

	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)
	request.LoadBalancerType = common.StringPtr("OPEN")

	response, err := client.DescribeLoadBalancers(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("DescribeInstances an API error has returned: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed, error: %s", err)
	}
	return response, nil
}

// Discover implements Product interface
func (l LBPublic) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	lbPublicInstances := []*clb.LoadBalancer{}
	instances := []*monitor.Instance{}

	response, err := l.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}
	lbPublicInstances = append(lbPublicInstances, response.Response.LoadBalancerSet...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(int64(*response.Response.TotalCount)/limit)+1; i++ {
		response, err := l.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		lbPublicInstances = append(lbPublicInstances, response.Response.LoadBalancerSet...)
	}

	for _, lbPlubicInstance := range lbPublicInstances {

		if len(lbPlubicInstance.LoadBalancerVips) == 0 || lbPlubicInstance.LoadBalancerVips[0] == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("vip"),
					Value: lbPlubicInstance.LoadBalancerVips[0],
				},
			},
		})

	}

	return instances, nil
}

// Metrics implements Product interface
func (l LBPublic) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/10997
	return []string{
		"AccOuttraffic", "ClbHttp3xx", "ClbHttp404", "ClbHttp4xx", "ClbHttp502", "ClbHttp5xx",
		"ConNum", "Http2xx", "Http3xx", "Http404", "Http4xx", "Http502", "Http5xx", "InactiveConn",
		"InPkg", "InTraffic", "NewConn", "OutPkg", "OutTraffic",
		"ReqAvg", "ReqMax", "RspAvg", "RspMax", "RspTimeout",
		"SuccReq", "TotalReq",
	}
}

// LBPrivate defines Cloud Load Balancer, see: https://intl.cloud.tencent.com/document/product/214
type LBPrivate struct{}

// Namespace implements Product interface
func (l *LBPrivate) Namespace() string {
	return "QCE/LB_PRIVATE"
}

func (l *LBPrivate) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*clb.DescribeLoadBalancersResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "clb", endpoint)
	client, err := clb.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("clb.NewClient failed, error: %s", err)
	}

	request := clb.NewDescribeLoadBalancersRequest()

	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)
	request.LoadBalancerType = common.StringPtr("INTERNAL")

	response, err := client.DescribeLoadBalancers(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("DescribeInstances an API error has returned: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed, error: %s", err)
	}
	return response, nil
}

// Discover implements Product interface
func (l *LBPrivate) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	lbPrivateInstances := []*clb.LoadBalancer{}
	instances := []*monitor.Instance{}

	response, err := l.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}
	lbPrivateInstances = append(lbPrivateInstances, response.Response.LoadBalancerSet...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(int64(*response.Response.TotalCount)/limit)+1; i++ {
		response, err := l.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		lbPrivateInstances = append(lbPrivateInstances, response.Response.LoadBalancerSet...)
	}

	for _, lbPrivateInstance := range lbPrivateInstances {
		if len(lbPrivateInstance.LoadBalancerVips) == 0 || lbPrivateInstance.LoadBalancerVips[0] == nil || lbPrivateInstance.VpcId == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("vip"),
					Value: lbPrivateInstance.LoadBalancerVips[0],
				},
				{
					Name:  common.StringPtr("vpcId"),
					Value: lbPrivateInstance.VpcId,
				},
			},
		})

	}
	return instances, nil
}

// Metrics implements Product interface
func (l *LBPrivate) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/34639
	return []string{
		"Connum", "NewConn", "Intraffic", "Outtraffic", "Inpkg", "Outpkg",
	}
}

// CES defines Elasticsearch Service, see: https://intl.cloud.tencent.com/document/product/845
type CES struct{}

// Namespace implements Product interface
func (c *CES) Namespace() string {
	return "QCE/CES"
}

func (c *CES) discover(crs *common.Credential, region, endpoint string, offset, limit int64) (*es.DescribeInstancesResponse, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "es", endpoint)
	client, err := es.NewClient(crs, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("es.NewClient failed, error: %s", err)
	}

	request := es.NewDescribeInstancesRequest()

	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(100)

	response, err := client.DescribeInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return nil, fmt.Errorf("DescribeInstances an API error has returned: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed, error: %s", err)
	}
	return response, nil
}

// Discover implements Product interface
func (c *CES) Discover(crs *common.Credential, region, endpoint string) ([]*monitor.Instance, error) {
	offset, limit := int64(0), int64(100)
	esInstances := []*es.InstanceInfo{}
	instances := []*monitor.Instance{}

	response, err := c.discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return nil, err
	}

	esInstances = append(esInstances, response.Response.InstanceList...)

	// discover all instances if total count is bigger than limit
	for i := 1; i < int(int64(*response.Response.TotalCount)/limit)+1; i++ {
		response, err := c.discover(crs, region, endpoint, offset+(int64(i)*limit), limit)
		if err != nil {
			return nil, err
		}
		esInstances = append(esInstances, response.Response.InstanceList...)
	}

	for _, esInstance := range esInstances {

		if esInstance.InstanceId == nil {
			continue
		}

		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr("uInstanceId"),
					Value: esInstance.InstanceId,
				},
			},
		})

	}
	return instances, nil
}

// Metrics implements Product interface
func (c *CES) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/34642
	return []string{
		"Status", "DiskUsageAvg", "DiskUsageMax",
		"JvmMemUsageAvg", "JvmMemUsageMax", "JvmOldMemUsageAvg", "JvmOldMemUsageMax",
		"CpuUsageAvg", "CpuUsageMax", "CpuLoad1minAvg", "CpuLoad1minMax",
		"IndexLatencyAvg", "IndexLatencyMax", "SearchLatencyAvg", "SearchLatencyMax", "IndexSpeed", "SearchCompletedSpeed",
		"BulkRejectedCompletedPercent", "SearchRejectedCompletedPercent", "IndexDocs",
	}
}
