package tencentcloudcm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
)

// Product defines cloud product
type Product interface {
	// Namespace returns the namespac for the product, which can be found
	// in https://intl.cloud.tencent.com/document/product/248/6140
	Namespace() string

	// Metrics returns the supported metrics for the product, which can
	// be found in https://intl.cloud.tencent.com/document/product/248/6140
	Metrics() []string

	// Keys maps the instance key name to the Cloud Monitor required
	// dimension key name, e.g., product DC(Direct Connect) has instance
	// key name 'DirectConnectId' which maps to 'directConnectId' as the
	// dimension key name for Cloud Monitor request.
	Keys() map[string]string

	// Discover discovers the product instances
	Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error)
}

// DC defines Direct Connect, see: https://intl.cloud.tencent.com/document/product/216
type DC struct{}

// Namespace implements Product interface
func (d DC) Namespace() string {
	return "QCE/DC"
}

// Metrics implements Product interface
func (d DC) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/10994
	return []string{"OutBandwidth", "InBandwidth"}
}

// Keys implements Product interface
func (d DC) Keys() map[string]string {
	return map[string]string{
		"DirectConnectId": "directConnectId",
	}
}

func mapInterfaceToString(mInterfaceList []map[string]interface{}) []map[string]string {
	mStringList := []map[string]string{}
	for _, mInterface := range mInterfaceList {
		mString := map[string]string{}
		for k, v := range mInterface {
			if v != nil && !strings.Contains(strings.ToLower(k), "time") && fmt.Sprintf("%v", v) != "" {
				valueType := reflect.TypeOf(v)
				switch valueType.Kind() {
				case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					mString[k] = fmt.Sprintf("%v", v)
				case reflect.Slice:
					val, ok := v.([]interface{})
					if ok {
						if len(val) > 0 {
							mString[k] = fmt.Sprintf("%v", val[0])
						}
					}
				case reflect.Map:
					vJSON, err := json.Marshal(v)
					if err != nil {
						continue
					}
					mString[k] = string(vJSON)
				}
			}
		}
		mStringList = append(mStringList, mString)
	}
	return mStringList
}

// Discover implements Product interface
func (d DC) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "dc", endpoint)
	client, err := dc.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("dc.NewClient failed, error: %s", err)
	}

	request := dc.NewDescribeDirectConnectTunnelsRequest()
	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)

	response, err := client.DescribeDirectConnectTunnels(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", d.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.DirectConnectTunnelSet)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", d.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", d.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// CVM defines Cloud Virtual Machine, see: https://intl.cloud.tencent.com/document/product/213
type CVM struct{}

// Namespace implements Product interface
func (c CVM) Namespace() string {
	return "QCE/CVM"
}

// Metrics implements Product interface
func (c CVM) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/6843
	return []string{
		// CPU Monitor
		"CPUUsage", "CpuLoadavg", "Cpuloadavg5m", "Cpuloadavg15m", "BaseCpuUsage",
		// GPU Monitor
		"GpuMemTotal", "GpuMemUsage", "GpuMemUsed",
		"GpuPowDraw", "GpuPowLimit", "GpuPowUsage",
		"GpuTemp", "GpuUtil",
		// Network Monitor
		"LanOuttraffic", "LanIntraffic", "LanOutpkg", "LanInpkg",
		"WanOuttraffic", "WanIntraffic", "WanOutpkg", "WanInpkg",
		"AccOuttraffic", "TcpCurrEstab", "TimeOffset",
		// Memory Monitor
		"MemUsed", "MemUsage",
		// Disk Monitor
		"CvmDiskUsage",
	}
}

// Keys implements Product interface
func (c CVM) Keys() map[string]string {
	return map[string]string{
		"InstanceId": "InstanceId",
	}
}

func (c CVM) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "cvm", endpoint)
	client, err := cvm.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("cvm.NewClient failed, error: %s", err)
	}

	request := cvm.NewDescribeInstancesRequest()
	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)

	response, err := client.DescribeInstances(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", c.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.InstanceSet)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", c.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", c.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// CDB defines TencentDB for MySQL, see: https://intl.cloud.tencent.com/document/product/236
type CDB struct{}

// Namespace implements Product interface
func (c CDB) Namespace() string {
	return "QCE/CDB"
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

// Keys implements Product interface
func (c CDB) Keys() map[string]string {
	return map[string]string{
		"InstanceId": "InstanceId",
	}
}

// Discover implements Product interface
func (c CDB) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "cdb", endpoint)
	client, err := cdb.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("cdb.NewClient failed, error: %s", err)
	}

	request := cdb.NewDescribeDBInstancesRequest()

	request.Offset = common.Uint64Ptr(uint64(offset))
	request.Limit = common.Uint64Ptr(uint64(limit))

	response, err := client.DescribeDBInstances(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", c.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.Items)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", c.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", c.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// Redis defines TencentDB for Redis, see: https://intl.cloud.tencent.com/document/product/239
type Redis struct{}

// Namespace implements Product interface
func (r Redis) Namespace() string {
	return "QCE/REDIS"
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

// Keys implements Product interface
func (r Redis) Keys() map[string]string {
	return map[string]string{
		"InstanceId": "instanceid",
	}
}

// Discover implements Product interface
func (r Redis) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "redis", endpoint)
	client, err := redis.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("redis.NewClient failed, error: %s", err)
	}

	request := redis.NewDescribeInstancesRequest()
	request.Limit = common.Uint64Ptr(uint64(limit))
	request.Offset = common.Uint64Ptr(uint64(offset))

	response, err := client.DescribeInstances(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", r.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.InstanceSet)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", r.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", r.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// LBPublic defines Cloud Load Balancer, see: https://intl.cloud.tencent.com/document/product/214
type LBPublic struct{}

// Namespace implements Product interface
func (l LBPublic) Namespace() string {
	return "QCE/LB_PUBLIC"
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

// Keys implements Product interface
func (l LBPublic) Keys() map[string]string {
	return map[string]string{
		"LoadBalancerVips": "vip",
	}
}

// Discover implements Product interface
func (l LBPublic) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "clb", endpoint)
	client, err := clb.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("clb.NewClient failed, error: %s", err)
	}

	request := clb.NewDescribeLoadBalancersRequest()
	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)
	request.LoadBalancerType = common.StringPtr("OPEN")

	response, err := client.DescribeLoadBalancers(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", l.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.LoadBalancerSet)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", l.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", l.Namespace(), err)
	}
	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// LBPrivate defines Cloud Load Balancer, see: https://intl.cloud.tencent.com/document/product/214
type LBPrivate struct{}

// Namespace implements Product interface
func (l *LBPrivate) Namespace() string {
	return "QCE/LB_PRIVATE"
}

// Metrics implements Product interface
func (l *LBPrivate) Metrics() []string {
	// see: https://intl.cloud.tencent.com/document/product/248/34639
	return []string{
		"Connum", "NewConn", "Intraffic", "Outtraffic", "Inpkg", "Outpkg",
	}
}

// Keys implements Product interface
func (l LBPrivate) Keys() map[string]string {
	return map[string]string{
		"LoadBalancerVips": "vip",
		"VpcId":            "vpcId",
	}
}

// Discover implements Product interface
func (l *LBPrivate) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "clb", endpoint)
	client, err := clb.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("clb.NewClient failed, error: %s", err)
	}

	request := clb.NewDescribeLoadBalancersRequest()

	request.Offset = common.Int64Ptr(offset)
	request.Limit = common.Int64Ptr(limit)
	request.LoadBalancerType = common.StringPtr("INTERNAL")

	response, err := client.DescribeLoadBalancers(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", l.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.LoadBalancerSet)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", l.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", l.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}

// CES defines Elasticsearch Service, see: https://intl.cloud.tencent.com/document/product/845
type CES struct{}

// Namespace implements Product interface
func (c *CES) Namespace() string {
	return "QCE/CES"
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

// Keys implements Product interface
func (c CES) Keys() map[string]string {
	return map[string]string{
		"InstanceId": "uInstanceId",
	}
}

func (c *CES) Discover(crs *common.Credential, region, endpoint string, offset, limit int64) (uint64, []map[string]string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("%s.%s", "es", endpoint)
	client, err := es.NewClient(crs, region, cpf)
	if err != nil {
		return 0, nil, fmt.Errorf("es.NewClient failed, error: %s", err)
	}

	request := es.NewDescribeInstancesRequest()
	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(100)

	response, err := client.DescribeInstances(request)
	if err != nil {
		return 0, nil, fmt.Errorf("discover instances for namespace %s failed: %s", c.Namespace(), err)
	}

	var instances []map[string]interface{}
	instancesJSON, err := json.Marshal(response.Response.InstanceList)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Marshal response for namespace %s failed: %s", c.Namespace(), err)
	}
	err = json.Unmarshal(instancesJSON, &instances)
	if err != nil {
		return 0, nil, fmt.Errorf("json.Unmarshal response for namespace %s failed: %s", c.Namespace(), err)
	}

	return uint64(*response.Response.TotalCount), mapInterfaceToString(instances), nil
}
