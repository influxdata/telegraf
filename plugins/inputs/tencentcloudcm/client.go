package tencentcloudcm

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// These codes is forked and tweaked from original tencentcloud-sdk-go.
// Optimized struct field tags to fix incompatibility issue when using 'go vet'
// source: https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/monitor/v20180724/client.go

// Client defines cloud monitor sdk client
type Client struct {
	common.Client
}

// APIVersion defines cloud monitor API version
const APIVersion = "2018-07-24"

// NewGetMonitorDataRequest factory
func NewGetMonitorDataRequest() (request *GetMonitorDataRequest) {
	request = &GetMonitorDataRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("monitor", APIVersion, "GetMonitorData")
	return
}

// NewGetMonitorDataResponse factory
func NewGetMonitorDataResponse() (response *GetMonitorDataResponse) {
	response = &GetMonitorDataResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 获取云产品的监控数据。传入产品的命名空间、对象维度描述和监控指标即可获得相应的监控数据。
// 接口调用频率限制为：20次/秒，1200次/分钟。单请求最多可支持批量拉取10个实例的监控数据，单请求的数据点数限制为1440个。
// 若您需要调用的指标、对象较多，可能存在因限频出现拉取失败的情况，建议尽量将请求按时间维度均摊。
// Details please refer to https://intl.cloud.tencent.com/document/product/248/33881
func (c *Client) GetMonitorData(request *GetMonitorDataRequest) (response *GetMonitorDataResponse, err error) {
	if request == nil {
		request = NewGetMonitorDataRequest()
	}
	response = NewGetMonitorDataResponse()
	err = c.Send(request, response)
	return
}
