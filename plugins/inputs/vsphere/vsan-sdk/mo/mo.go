package mo

import (
	"github.com/vmware/govmomi/vim25/types"
	"reflect"
)

type PerformanceManager struct {
	Self types.ManagedObjectReference

	Description        types.PerformanceDescription `mo:"description"`
	HistoricalInterval []types.PerfInterval         `mo:"historicalInterval"`
	PerfCounter        []types.PerfCounterInfo      `mo:"perfCounter"`
}

func (m PerformanceManager) Reference() types.ManagedObjectReference {
	return m.Self
}

func init() {
	types.Add("PerformanceManager", reflect.TypeOf((*PerformanceManager)(nil)).Elem())
}
