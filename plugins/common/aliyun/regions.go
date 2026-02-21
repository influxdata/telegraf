package aliyun

// aliyunRegionList contains all supported Aliyun regions
// Source: https://www.alibabacloud.com/help/doc-detail/40654.htm
var aliyunRegionList = []string{
	"cn-qingdao",
	"cn-beijing",
	"cn-zhangjiakou",
	"cn-huhehaote",
	"cn-hangzhou",
	"cn-shanghai",
	"cn-shenzhen",
	"cn-heyuan",
	"cn-chengdu",
	"cn-hongkong",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-5",
	"ap-south-1",
	"ap-northeast-1",
	"us-west-1",
	"us-east-1",
	"eu-central-1",
	"eu-west-1",
	"me-east-1",
}

// DefaultRegions returns a copy of the default region list
func DefaultRegions() []string {
	regions := make([]string, len(aliyunRegionList))
	copy(regions, aliyunRegionList)
	return regions
}
