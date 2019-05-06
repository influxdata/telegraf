/*
Copyright (c) 2014-2018 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"github.com/vmware/govmomi/vim25/types"
	"reflect"
	"time"
)

type VsanPerfGetSupportedEntityTypes VsanPerfGetSupportedEntityTypesRequestType

func init() {
	t["VsanPerfGetSupportedEntityTypes"] = reflect.TypeOf((*VsanPerfGetSupportedEntityTypes)(nil)).Elem()
}

type VsanPerfGetSupportedEntityTypesRequestType struct {
	This types.ManagedObjectReference `xml:"_this"`
}

func init() {
	t["VsanPerfGetSupportedEntityTypesRequestType"] = reflect.TypeOf((*VsanPerfGetSupportedEntityTypesRequestType)(nil)).Elem()
}

type VsanPerfGetSupportedEntityTypesResponse struct {
	Returnval []VsanPerfEntityType `xml:"returnval,omitempty"`
}

type VsanPerfEntityType struct {
	DynamicData

	Name        string      `xml:"name"`
	Id          string      `xml:"id"`
	Graphs      interface{} `xml:"graphs"`
	Description string      `xml:"description,omitempty"`
}

// Cluster health summary
type VsanQueryVcClusterHealthSummary VsanQueryVcClusterHealthSummaryRequestType

func init() {
	t["VsanQueryVcClusterHealthSummary"] = reflect.TypeOf((*VsanQueryVcClusterHealthSummary)(nil)).Elem()
}

type VsanQueryVcClusterHealthSummaryRequestType struct {
	This            types.ManagedObjectReference `xml:"_this"`
	Cluster         types.ManagedObjectReference `xml:"cluster"`
	VmCreateTimeout int32                        `xml:"vmCreateTimeout,omitempty"`
	ObjUuids        []string                     `xml:"objUuids,omitempty"`
	IncludeObjUuids *bool                        `xml:"includeObjUuids"`
	Fields          []string                     `xml:"fields,omitempty"`
	FetchFromCache  *bool                        `xml:"fetchFromCache"`
	Perspective     string                       `xml:"perspective,omitempty"`
}

func init() {
	t["VsanQueryVcClusterHealthSummaryRequestType"] = reflect.TypeOf((*VsanQueryVcClusterHealthSummaryRequestType)(nil)).Elem()
}

type VsanQueryVcClusterHealthSummaryResponse struct {
	Returnval VsanClusterHealthSummary `xml:"returnval"`
}

type VsanClusterHealthSummary struct {
	DynamicData

	ClusterStatus            interface{} `xml:"clusterStatus,omitempty"`
	Timestamp                *time.Time  `xml:"timestamp"`
	ClusterVersions          interface{} `xml:"clusterVersions,omitempty"`
	ObjectHealth             interface{} `xml:"objectHealth,omitempty"`
	VmHealth                 interface{} `xml:"vmHealth,omitempty"`
	NetworkHealth            interface{} `xml:"networkHealth,omitempty"`
	LimitHealth              interface{} `xml:"limitHealth,omitempty"`
	AdvCfgSync               interface{} `xml:"advCfgSync,omitempty"`
	CreateVmHealth           interface{} `xml:"createVmHealth,omitempty"`
	PhysicalDisksHealth      interface{} `xml:"physicalDisksHealth,omitempty"`
	EncryptionHealth         interface{} `xml:"encryptionHealth,omitempty"`
	HclInfo                  interface{} `xml:"hclInfo,omitempty"`
	Groups                   interface{} `xml:"groups,omitempty"`
	OverallHealth            string      `xml:"overallHealth"`
	OverallHealthDescription string      `xml:"overallHealthDescription"`
	ClomdLiveness            interface{} `xml:"clomdLiveness,omitempty"`
	DiskBalance              interface{} `xml:"diskBalance,omitempty"`
	GenericCluster           interface{} `xml:"genericCluster,omitempty"`
	NetworkConfig            interface{} `xml:"networkConfig,omitempty"`
	VsanConfig               interface{} `xml:"vsanConfig,omitempty,typeattr"`
	BurnInTest               interface{} `xml:"burnInTest,omitempty"`
}

func init() {
	t["VsanClusterHealthSummary"] = reflect.TypeOf((*VsanClusterHealthSummary)(nil)).Elem()
}

// Space Usage
type DynamicData struct {
}

type VsanQuerySpaceUsage VsanQuerySpaceUsageRequestType

func init() {
	t["VsanQuerySpaceUsage"] = reflect.TypeOf((*VsanQuerySpaceUsage)(nil)).Elem()
}

type VsanQuerySpaceUsageRequestType struct {
	This    types.ManagedObjectReference `xml:"_this"`
	Cluster types.ManagedObjectReference `xml:"cluster"`
}

func init() {
	t["VsanQuerySpaceUsageRequestType"] = reflect.TypeOf((*VsanQuerySpaceUsageRequestType)(nil)).Elem()
}

type VsanQuerySpaceUsageResponse struct {
	Returnval VsanSpaceUsage `xml:"returnval"`
}

type VsanSpaceUsage struct {
	DynamicData

	TotalCapacityB int64       `xml:"totalCapacityB"`
	FreeCapacityB  int64       `xml:"freeCapacityB,omitempty"`
	SpaceOverview  interface{} `xml:"spaceOverview,omitempty"`
	SpaceDetail    interface{} `xml:"spaceDetail,omitempty"`
}

// Performance
type VsanPerfQueryPerf VsanPerfQueryPerfRequestType

func init() {
	t["VsanPerfQueryPerf"] = reflect.TypeOf((*VsanPerfQueryPerf)(nil)).Elem()
}

type VsanPerfQueryPerfRequestType struct {
	This       types.ManagedObjectReference  `xml:"_this"`
	QuerySpecs []VsanPerfQuerySpec           `xml:"querySpecs"`
	Cluster    *types.ManagedObjectReference `xml:"cluster,omitempty"`
}

func init() {
	t["VsanPerfQueryPerfRequestType"] = reflect.TypeOf((*VsanPerfQueryPerfRequestType)(nil)).Elem()
}

type VsanPerfQueryPerfResponse struct {
	Returnval []VsanPerfEntityMetricCSV `xml:"returnval"`
}

type VsanPerfQuerySpec struct {
	DynamicData

	EntityRefId string     `xml:"entityRefId"`
	StartTime   *time.Time `xml:"startTime"`
	EndTime     *time.Time `xml:"endTime"`
	Group       string     `xml:"group,omitempty"`
	Labels      []string   `xml:"labels,omitempty"`
	Interval    int32      `xml:"interval,omitempty"`
}

type VsanPerfEntityMetricCSV struct {
	DynamicData

	EntityRefId string                    `xml:"entityRefId"`
	SampleInfo  string                    `xml:"sampleInfo,omitempty"`
	Value       []VsanPerfMetricSeriesCSV `xml:"value,omitempty"`
}

type VsanPerfMetricSeriesCSV struct {
	DynamicData

	MetricId  VsanPerfMetricId `xml:"metricId"`
	Threshold interface{}      `xml:"threshold,omitempty"`
	Values    string           `xml:"values,omitempty"`
}

func init() {
	t["VsanPerfEntityMetricCSV"] = reflect.TypeOf((*VsanPerfEntityMetricCSV)(nil)).Elem()
}

type VsanPerfMetricId struct {
	DynamicData

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

// Syncing summary
type VsanQuerySyncingVsanObjects VsanQuerySyncingVsanObjectsRequestType

func init() {
	t["VsanQuerySyncingVsanObjects"] = reflect.TypeOf((*VsanQuerySyncingVsanObjects)(nil)).Elem()
}

type VsanQuerySyncingVsanObjectsRequestType struct {
	This           types.ManagedObjectReference `xml:"_this"`
	Uuids          []string                     `xml:"uuids,omitempty"`
	Start          int32                        `xml:"start,omitempty"`
	Limit          *int32                       `xml:"limit"`
	IncludeSummary *bool                        `xml:"includeSummary"`
}

func init() {
	t["VsanQuerySyncingVsanObjectsRequestType"] = reflect.TypeOf((*VsanQuerySyncingVsanObjectsRequestType)(nil)).Elem()
}

type VsanQuerySyncingVsanObjectsResponse struct {
	Returnval VsanHostVsanObjectSyncQueryResult `xml:"returnval"`
}

type VsanHostVsanObjectSyncQueryResult struct {
	DynamicData

	TotalObjectsToSync int64                         `xml:"totalObjectsToSync,omitempty"`
	TotalBytesToSync   int64                         `xml:"totalBytesToSync,omitempty"`
	TotalRecoveryETA   int64                         `xml:"totalRecoveryETA,omitempty"`
	Objects            []VsanHostVsanObjectSyncState `xml:"objects,omitempty"`
}

func init() {
	t["VsanHostVsanObjectSyncQueryResult"] = reflect.TypeOf((*VsanHostVsanObjectSyncQueryResult)(nil)).Elem()
}

type VsanHostVsanObjectSyncState struct {
	DynamicData

	Uuid       string                       `xml:"uuid"`
	Components []VsanHostComponentSyncState `xml:"components"`
}

func init() {
	t["VsanHostVsanObjectSyncState"] = reflect.TypeOf((*VsanHostVsanObjectSyncState)(nil)).Elem()
}

type VsanHostComponentSyncState struct {
	DynamicData

	Uuid        string   `xml:"uuid"`
	DiskUuid    string   `xml:"diskUuid"`
	HostUuid    string   `xml:"hostUuid"`
	BytesToSync int64    `xml:"bytesToSync"`
	RecoveryETA int64    `xml:"recoveryETA,omitempty"`
	Reasons     []string `xml:"reasons,omitempty"`
}

func init() {
	t["VsanHostComponentSyncState"] = reflect.TypeOf((*VsanHostComponentSyncState)(nil)).Elem()
}
