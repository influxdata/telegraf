package types

import (
	"reflect"
	"time"

	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// PerfMetricId defines
type PerfMetricId struct {
	CounterId int32  `xml:"counterId"`
	Instance  string `xml:"instance"`
}

func init() {
	t["PerfMetricId"] = reflect.TypeOf((*PerfMetricId)(nil)).Elem()
}

type VsanPerfQuerySpec struct {
	EntityRefId string     `xml:"entityRefId"`
	StartTime   *time.Time `xml:"startTime"`
	EndTime     *time.Time `xml:"endTime"`
	Group       string     `xml:"group,omitempty"`
	Labels      []string   `xml:"labels,omitempty"`
	Interval    int32      `xml:"interval,omitempty"`
}

func init() {
	t["VsanPerfQuerySpec"] = reflect.TypeOf((*VsanPerfQuerySpec)(nil)).Elem()
}

type PerfQuerySpec struct {
	Entity     types.ManagedObjectReference `xml:"entity"`
	StartTime  *time.Time                   `xml:"startTime"`
	EndTime    *time.Time                   `xml:"endTime"`
	MaxSample  int32                        `xml:"maxSample,omitempty"`
	MetricId   []PerfMetricId               `xml:"metricId,omitempty"`
	IntervalId int32                        `xml:"intervalId,omitempty"`
	Format     string                       `xml:"format,omitempty"`
}

func init() {
	t["PerfQuerySpec"] = reflect.TypeOf((*PerfQuerySpec)(nil)).Elem()
}

type VsanPerfQueryPerf VsanPerfQueryPerfRequestType

func init() {
	t["VsanPerfQueryPerf"] = reflect.TypeOf((*VsanPerfQueryPerf)(nil)).Elem()
}

type VsanPerfQueryPerfRequestType struct {
	This       types.ManagedObjectReference `xml:"_this"`
	QuerySpecs []VsanPerfQuerySpec          `xml:"querySpecs"`
	Cluster    types.ManagedObjectReference `xml:"cluster,omitempty"`
}

func init() {
	t["VsanPerfQueryPerfRequestType"] = reflect.TypeOf((*VsanPerfQueryPerfRequestType)(nil)).Elem()
}

type VsanPerfMetricId struct {
	Label                  string `xml:"label"`
	Group                  string `xml:"group,omitempty"`
	RollupType             string `xml:"rollupType,omitempty"`
	StatsType              string `xml:"statsType,omitempty"`
	Name                   string `xml:"name,omitempty"`
	Description            string `xml:"description,omitempty"`
	MetricsCollectInterval int32  `xml:"metricsCollectInterval,omitempty"`
}

func init() {
	t["VsanPerfMetricId"] = reflect.TypeOf((*VsanPerfMetricId)(nil)).Elem()
}

type VsanPerfThreshold struct {
	Direction string `xml:"direction"`
	Yellow    string `xml:"yellow,omitempty"`
	Red       string `xml:"red,omitempty"`
}

func init() {
	t["VsanPerfThreshold"] = reflect.TypeOf((*VsanPerfThreshold)(nil)).Elem()
}

type VsanPerfMetricSeriesCSV struct {
	MetricId  VsanPerfMetricId   `xml:"metricId"`
	Threshold *VsanPerfThreshold `xml:"threshold,omitempty"`
	Values    string             `xml:"values,omitempty"`
}

func init() {
	t["VsanPerfMetricSeriesCSV"] = reflect.TypeOf((*VsanPerfMetricSeriesCSV)(nil)).Elem()
}

type VsanPerfEntityMetricCSV struct {
	EntityRefId string                    `xml:"entityRefId"`
	SampleInfo  string                    `xml:"sampleInfo,omitempty"`
	Value       []VsanPerfMetricSeriesCSV `xml:"value,omitempty"`
}

func init() {
	t["VsanPerfEntityMetricCSV"] = reflect.TypeOf((*VsanPerfEntityMetricCSV)(nil)).Elem()
}

type VsanPerfQueryPerfResponse struct {
	Returnval []VsanPerfEntityMetricCSV `xml:"returnval"`
}

type VsanPerfQueryPerfBody struct {
	Req    *VsanPerfQueryPerf         `xml:"urn:vsan VsanPerfQueryPerf,omitempty"`
	Res    *VsanPerfQueryPerfResponse `xml:"urn:vsan VsanPerfQueryPerfResponse,omitempty"`
	Fault_ *soap.Fault                `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault,omitempty"`
}

// VsanPerfQueryPerfBody must implement HasFault interface
func (b *VsanPerfQueryPerfBody) Fault() *soap.Fault { return b.Fault_ }

func init() {
	t["VsanPerfQueryPerfBody"] = reflect.TypeOf((*VsanPerfQueryPerfBody)(nil)).Elem()
}
