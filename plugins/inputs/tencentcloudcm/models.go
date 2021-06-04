package tencentcloudcm

import (
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// These codes is forked and tweaked from original tencentcloud-sdk-go.
// Optimized struct field tags to fix incompatibility issue when using 'go vet'
// source: https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/monitor/v20180724/models.go

// GetMonitorDataRequest defines GetMonitorData request
type GetMonitorDataRequest struct {
	*tchttp.BaseRequest

	// 命名空间，如QCE/CVM。各个云产品的详细命名空间说明请参阅各个产品[监控指标](https://cloud.tencent.com/document/product/248/6140)文档
	Namespace *string `json:"Namespace,omitempty" name:"Namespace"`

	// 指标名称，如CPUUsage。各个云产品的详细指标说明请参阅各个产品[监控指标](https://cloud.tencent.com/document/product/248/6140)文档，对应的指标英文名即为MetricName
	MetricName *string `json:"MetricName,omitempty" name:"MetricName"`

	// 实例对象的维度组合，格式为key-value键值对形式的集合。如[{"Name":"InstanceId","Value":"ins-j0hk02zo"}]。各个云产品的维度请参阅各个产品[监控指标](https://cloud.tencent.com/document/product/248/6140)文档，对应的维度列即为维度组合的key,value为key对应的值
	Instances []*MonitorInstance `json:"Instances,omitempty" name:"Instances"`

	// 监控统计周期，如60。默认为取值为300，单位为s。每个指标支持的统计周期不一定相同，各个云产品支持的统计周期请参阅各个产品[监控指标](https://cloud.tencent.com/document/product/248/6140)文档，对应的统计周期列即为支持的统计周期
	Period *uint64 `json:"Period,omitempty" name:"Period"`

	// 起始时间，如2018-09-22T19:51:23+08:00
	StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

	// 结束时间，如2018-09-22T20:51:23+08:00，默认为当前时间。 EndTime不能小于StartTime
	EndTime *string `json:"EndTime,omitempty" name:"EndTime"`
}

// MonitorInstance defines monitor instances
type MonitorInstance struct {

	// 实例的维度组合
	Dimensions []*MonitorDimension `json:"Dimensions,omitempty" name:"Dimensions"`
}

// MonitorDimension defines monitor dimensions
type MonitorDimension struct {

	// 实例维度名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 实例维度值
	Value *string `json:"Value,omitempty" name:"Value"`
}

// GetMonitorDataResponse defines GetMonitorData response
type GetMonitorDataResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 统计周期
		Period *uint64 `json:"Period,omitempty" name:"Period"`

		// 指标名
		MetricName *string `json:"MetricName,omitempty" name:"MetricName"`

		// 数据点数组
		DataPoints []*MonitorDataPoint `json:"DataPoints,omitempty" name:"DataPoints"`

		// 开始时间
		StartTime *string `json:"StartTime,omitempty" name:"StartTime"`

		// 结束时间
		EndTime *string `json:"EndTime,omitempty" name:"EndTime"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

// MonitorDataPoint defines monitor data points
type MonitorDataPoint struct {

	// 实例对象维度组合
	Dimensions []*MonitorDimension `json:"Dimensions,omitempty" name:"Dimensions"`

	// 时间戳数组，表示那些时间点有数据，缺失的时间戳，没有数据点，可以理解为掉点了
	Timestamps []*float64 `json:"Timestamps,omitempty" name:"Timestamps"`

	// 监控值数组，该数组和Timestamps一一对应
	Values []*float64 `json:"Values,omitempty" name:"Values"`
}
