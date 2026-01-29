package aliyun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAliyunRegionList(t *testing.T) {
	require.NotEmpty(t, aliyunRegionList, "AliyunRegionList should not be empty")

	expectedRegions := []string{
		"cn-qingdao",
		"cn-beijing",
		"cn-hangzhou",
		"cn-shanghai",
		"cn-shenzhen",
		"cn-hongkong",
		"us-west-1",
		"us-east-1",
		"eu-central-1",
		"ap-southeast-1",
	}

	for _, expectedRegion := range expectedRegions {
		require.Contains(t, aliyunRegionList, expectedRegion, "Region list should contain %s", expectedRegion)
	}
}

func TestDefaultRegions(t *testing.T) {
	regions := DefaultRegions()

	require.NotEmpty(t, regions)
	require.Len(t, regions, len(aliyunRegionList))

	regions[0] = "modified-region"
	regions = append(regions, "new-region")

	require.NotEqual(t, "modified-region", aliyunRegionList[0])
	require.Len(t, aliyunRegionList, len(regions)-1)
}

func TestDefaultRegionsContent(t *testing.T) {
	regions := DefaultRegions()

	for i, region := range aliyunRegionList {
		require.Equal(t, region, regions[i], "Region at index %d should match", i)
	}
}

func TestRegionListCompleteness(t *testing.T) {
	require.Len(t, aliyunRegionList, 21, "Should have 21 regions")

	regionSet := make(map[string]bool)
	for _, region := range aliyunRegionList {
		require.False(t, regionSet[region], "Region %s should not be duplicated", region)
		regionSet[region] = true
	}
}

func TestRegionFormat(t *testing.T) {
	for _, region := range aliyunRegionList {
		require.NotEmpty(t, region, "Region should not be empty")
		require.Contains(t, region, "-", "Region should contain a hyphen: %s", region)
	}
}
