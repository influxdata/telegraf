package ipset

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func assertIpsetEntries(t *testing.T, acc *testutil.Accumulator, numMetrics int, numEntries []int, numIps []int) {
	if len(acc.Errors) > 0 {
		t.Errorf("Errors in Accumulator")
	}

	if len(acc.Metrics) != numMetrics {
		t.Errorf("Expected %d metric", numMetrics)
	}

	for index := range numEntries {
		if len(acc.Metrics[index].Fields) != 2 {
			t.Errorf("Expected 2 fields")
		}
		if acc.Metrics[index].Fields["num_entries"] != numEntries[index] {
			t.Errorf("Expected num_entries length to be %d", numEntries[index])
		}

		if acc.Metrics[index].Fields["num_ips"] != numIps[index] {
			t.Errorf("Expected num_ips length to be %d", numIps[index])
		}
	}
}

func TestIpsetEntries(t *testing.T) {
	acc := new(testutil.Accumulator)

	i := NewIpsetEntries(acc)
	i.addLine("create mylist hash:net family inet hashsize 16384 maxelem 131072 timeout 300 bucketsize 12 initval 0x4effa9ad")
	i.addLine("add mylist 89.101.238.143 timeout 161558")
	i.addLine("add mylist 122.224.15.166 timeout 186758")
	i.addLine("add mylist 47.128.40.145 timeout 431559")

	i.commit()

	assertIpsetEntries(t, acc, 1, []int{3}, []int{3})
}

func TestIpsetEntriesCidr(t *testing.T) {
	acc := new(testutil.Accumulator)

	i := NewIpsetEntries(acc)
	i.addLine("create mylist0 hash:net family inet hashsize 16384 maxelem 131072 timeout 300 bucketsize 12 initval 0x4effa9ad")
	i.addLine("add mylist0 89.101.238.143 timeout 161558")
	i.addLine("add mylist0 122.224.5.0/24 timeout 186758")
	i.addLine("add mylist0 47.128.40.145 timeout 431559")

	i.addLine("create mylist1 hash:net family inet hashsize 16384 maxelem 131072 timeout 300 bucketsize 12 initval 0x4effa9ad")
	i.addLine("add mylist1 90.101.238.143 timeout 161558")
	i.addLine("add mylist1 44.128.40.145 timeout 431559")
	i.addLine("add mylist1 122.224.5.0/8 timeout 186758")
	i.addLine("add mylist1 45.128.40.145 timeout 431559")

	i.commit()

	assertIpsetEntries(t, acc, 2, []int{3, 4}, []int{256, 16777217})
}
