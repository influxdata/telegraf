//Copyright (c) 2019, Groupon, Inc.
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are
//met:
//
//Redistributions of source code must retain the above copyright notice,
//this list of conditions and the following disclaimer.
//
//Redistributions in binary form must reproduce the above copyright
//notice, this list of conditions and the following disclaimer in the
//documentation and/or other materials provided with the distribution.
//
//Neither the name of GROUPON nor the names of its contributors may be
//used to endorse or promote products derived from this software without
//specific prior written permission.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
//IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
//TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
//PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
//TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package tdigestagg

import (
	"github.com/influxdata/tdigest"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/bucketing"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/constants"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"

	"fmt"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestClamSameTags(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(m1)
	histogram.Add(m2)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(1, 1)
	expected.Add(3, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}
	expectedTags := map[string]string{
		"foo":                      "bar",
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_bar",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)

	// Tag that exists on data but not in aggregation list
	expectedTags[tagKeyEnv] = valueDevelopment
	acc.AssertDoesNotContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
	delete(expectedTags, tagKeyEnv)

	expectedTags[constants.TagKeySource] = "test"
	acc.AssertDoesNotContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

func TestClamTimeBucket(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mSingleTimer)
	histogram.Add(mSingleTimer)
	mSingleTimer.SetTime(mSingleTimer.Time().Add(time.Minute))
	histogram.Add(mSingleTimer)
	histogram.Push(&acc)

	// 1 aggregations expected for each time config
	assert.Equal(t, len(acc.Metrics), 2)

	centroidSet := make(map[interface{}]int)
	for _, m := range acc.Metrics {
		fred := m.Fields[constants.FieldCentroids]
		centroidSet[fred] = 1
	}

	// assert that there are two distinct centroids created, one for each time config
	assert.Equal(t, len(centroidSet), 2)
}

func TestClamAtomAsSource(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyHost,
		AtomReplacementTagKey: constants.TagKeyHost,
	}, {
		ExcludeTags:           []string{constants.TagKeyHost},
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyService,
	}}

	histogram.Add(mTimer)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	serviceTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueService,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, serviceTags)

	hostTags := map[string]string{
		constants.TagKeyAggregates: serviceTags[constants.TagKeyAggregates],
		tagKeyAZ:                   serviceTags[tagKeyAZ],
		tagKeyEnv:                  serviceTags[tagKeyEnv],
		constants.TagKeyService:    serviceTags[constants.TagKeyService],
		constants.TagKeyBucketKey:  "m1_a_ubuntu_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueHost,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, hostTags)
}

func TestClamCloudExample(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          "namespace",
		AtomReplacementTagKey: "replicaHash",
	}, {
		ExcludeTags:           []string{"replicaHash"},
		SourceTagKey:          "namespace",
		AtomReplacementTagKey: "namespace",
	}}

	histogram.Add(mCloud1)
	histogram.Add(mCloud2)
	histogram.Add(mCloud3)
	histogram.Push(&acc)

	var atom1Expected = tdigest.NewWithCompression(0)
	atom1Expected.Add(1, 1)
	atom1Expected.Add(3, 1)
	atom1Fields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(atom1Expected.Centroids()),
	}

	var atom2Expected = tdigest.NewWithCompression(0)
	atom2Expected.Add(4, 1)
	atom2Fields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(atom2Expected.Centroids()),
	}

	var atomSourceExpected = tdigest.NewWithCompression(0)
	atomSourceExpected.Add(1, 1)
	atomSourceExpected.Add(3, 1)
	atomSourceExpected.Add(4, 1)
	atomSourceFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(atomSourceExpected.Centroids()),
	}

	atom2Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   "aws",
		tagKeyEnv:                  "us-west-1",
		constants.TagKeyService:    "cloud",
		constants.TagKeyBucketKey:  "cloud_a_kubeName_aws_us-west-1_kubeName_h2_cloud",
		constants.TagKeySource:     "kubeName",
		"namespace":                "kubeName",
		constants.TagKeyAtom:       "h2",
		"replicaHash":              "h2",
	}
	acc.AssertContainsTaggedFields(t, "cloud_a", atom2Fields, atom2Tags)

	// Tags for L1 level aggregation
	atom1Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   atom2Tags[tagKeyAZ],
		tagKeyEnv:                  atom2Tags[tagKeyEnv],
		constants.TagKeyService:    atom2Tags[constants.TagKeyService],
		constants.TagKeySource:     atom2Tags[constants.TagKeySource],
		"namespace":                atom2Tags["namespace"],
		constants.TagKeyBucketKey:  "cloud_a_kubeName_aws_us-west-1_kubeName_h1_cloud",
		constants.TagKeyAtom:       "h1",
		"replicaHash":              "h1",
	}
	acc.AssertContainsTaggedFields(t, "cloud_a", atom1Fields, atom1Tags)

	atomSourceTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   "aws",
		tagKeyEnv:                  "us-west-1",
		constants.TagKeyService:    "cloud",
		constants.TagKeyBucketKey:  "cloud_a_kubeName_aws_us-west-1_kubeName_cloud",
		constants.TagKeySource:     "kubeName",
		"namespace":                "kubeName",
		constants.TagKeyAtom:       "kubeName",
	}
	acc.AssertContainsTaggedFields(t, "cloud_a", atomSourceFields, atomSourceTags)
}

func TestClamCustomAtomTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:  nil,
		SourceTagKey: constants.TagKeyService,
	}}

	histogram.Add(mAtom)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	atomTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_carbon_sea1_dev_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       "carbon",
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, atomTags)
}

func TestClamMissingAtomReplacementTag(t *testing.T) {
	var noAtom, _ = metric.New("m1",
		map[string]string{
			// This tag missing is error condition
			//TagKeyHost:    valueHost,
			tagKeyAZ:                valueSeattle,
			tagKeyEnv:               valueDevelopment,
			constants.TagKeyService: valueService,
			constants.TagKeyRollup:  "timer:*",
		},
		map[string]interface{}{
			"a": float64(9),
		},
		time.Now(),
	)

	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(noAtom)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	invalidAtomTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_MISSING_host_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       constants.MissingValueForRequiredTagPrefix + constants.TagKeyHost,
		constants.TagKeyHost:       constants.MissingValueForRequiredTagPrefix + constants.TagKeyHost,
		constants.AtomSlaViolation: constants.SlaViolationMissing,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, invalidAtomTags)
}

func TestClamMissingAtomTagNoOverride(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:  nil,
		SourceTagKey: constants.TagKeyService,
	}}

	histogram.Add(mTimer)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	L2Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       constants.MissingValueForRequiredTagPrefix + constants.TagKeyAtom,
		constants.TagKeyHost:       valueHost,
		"sla.violation.atom_tag":   "MISSING",
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, L2Tags)
}

func TestClamMissingSourceTag(t *testing.T) {
	var noSource, _ = metric.New("m1",
		map[string]string{
			constants.TagKeyHost: valueHost,
			tagKeyAZ:             valueSeattle,
			tagKeyEnv:            valueDevelopment,
			// This tag missing is error condition
			//TagKeyService: valueService,
			constants.TagKeyRollup: "timer:*",
		},
		map[string]interface{}{
			"a": float64(9),
		},
		time.Now(),
	)

	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(noSource)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	expectedTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    constants.MissingValueForRequiredTagPrefix + constants.TagKeyService,
		constants.TagKeyBucketKey:  "m1_a_MISSING_service_sea1_dev_ubuntu_MISSING_service",
		constants.TagKeySource:     constants.MissingValueForRequiredTagPrefix + constants.TagKeyService,
		constants.TagKeyHost:       valueHost,
		constants.TagKeyAtom:       valueHost,
		"sla.violation.source_tag": "MISSING",
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

func TestClamTimerTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mTimer)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	L2Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, L2Tags)
}

// Expected fields and tags copied from TestCounterTag
func TestClamDefaultTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mDefault)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	expectedTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsGauge,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

func TestClamCpuCoreQuantile(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	var expected = tdigest.NewWithCompression(0)
	for i := 0; i < 10; i++ {
		var mCpu, _ = metric.New("cpu",
			map[string]string{
				tagKeyAZ:                valueSeattle,
				tagKeyEnv:               valueDevelopment,
				constants.TagKeyHost:    valueHost,
				constants.TagKeyService: valueService,
				constants.TagKeyRollup:  "gauge:*-core",
				"core":                  strconv.Itoa(i),
			},
			map[string]interface{}{
				"used": float64(i),
			},
			time.Now(),
		)
		histogram.Add(mCpu)
		expected.Add(float64(i), 1)
	}

	histogram.Push(&acc)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	expectedTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsGauge,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "cpu_used_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "cpu_used", expectedFields, expectedTags)
}

func TestClamGaugeTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mGauge)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	expectedTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsGauge,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

func TestClamCounterTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mCounter)
	histogram.Add(mCounter)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldSum + constants.UtilityFieldModifier: 18.,
		constants.FieldCompression:                          0.,
		constants.FieldCentroids:                            fmt.Sprint(out),
	}

	expectedTags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsCounter,
		tagKeyAZ:                   valueSeattle,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

func TestClamExcludeTags(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           []string{tagKeyAZ},
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mTimer)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	L2Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_dev_ubuntu_telegraf",
		constants.TagKeySource:     valueService,
		constants.TagKeyAtom:       valueHost,
		constants.TagKeyHost:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, L2Tags)
}

func TestClamExcludeTagsAsAtom(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.UsingClam = true
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           []string{constants.TagKeyHost},
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mTimer)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldCompression: 0.,
		constants.FieldCentroids:   fmt.Sprint(out),
	}

	L2Tags := map[string]string{
		constants.TagKeyAggregates: constants.AggregationsTimer,
		tagKeyEnv:                  valueDevelopment,
		constants.TagKeyService:    valueService,
		constants.TagKeyBucketKey:  "m1_a_telegraf_sea1_dev_telegraf",
		constants.TagKeySource:     valueService,
		tagKeyAZ:                   valueSeattle,
		constants.TagKeyAtom:       valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, L2Tags)
}
