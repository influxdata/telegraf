package zabbixlld

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

type Operations interface{}
type OperationAdd []telegraf.Metric
type OperationPush struct{}
type OperationCheck []telegraf.Metric

func TestGenerateDataValues(t *testing.T) {
	lldTags := LLDTags{
		map[string]string{"foo": "A", "bar": "B"},
		map[string]string{"foo": "X", "bar": "Y"},
	}

	resp := lldTags.generateDataValues()

	expected := map[string][]map[string]string{
		"data": {
			{"{#FOO}": "A", "{#BAR}": "B"},
			{"{#FOO}": "X", "{#BAR}": "Y"},
		},
	}

	assert.Equal(t, expected, resp)
}

func TestLLDKey(t *testing.T) {
	tests := map[string]struct {
		Name   string
		Tags   map[string]string
		Result string
		Error  error
	}{
		"metric with emtpy measurement name return an error": {
			Name:   "",
			Tags:   map[string]string{"foo": "bar"},
			Result: "",
			Error:  fmt.Errorf("empty measurement name"),
		},
		"metric with no tags return an error": {
			Name:   "disk",
			Tags:   map[string]string{},
			Result: "",
			Error:  fmt.Errorf("metric without tags"),
		},
		"metric with one tag": {
			Name: "disk",
			Tags: map[string]string{
				"foo": "bar",
			},
			Result: "disk.foo",
			Error:  nil,
		},
		"metric with two tags sorted": {
			Name: "disk",
			Tags: map[string]string{
				"foo": "bar",
				"zaz": "oof",
			},
			Result: "disk.foo.zaz",
			Error:  nil,
		},
		"metric with two tags inverse sorted": {
			Name: "disk",
			Tags: map[string]string{
				"zaz": "oof",
				"foo": "bar",
			},
			Result: "disk.foo.zaz",
			Error:  nil,
		},
		"metric with three tags not sorted": {
			Name: "net",
			Tags: map[string]string{
				"bxx": "2",
				"cxx": "3",
				"axx": "1",
			},
			Result: "net.axx.bxx.cxx",
			Error:  nil,
		},
		"empty tags are ignored": {
			Name: "net",
			Tags: map[string]string{
				"bxx": "",
				"cxx": "3",
				"axx": "1",
			},
			Result: "net.axx.cxx",
			Error:  nil,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			id, err := generateLLDKey(test.Name, test.Tags)
			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Result, id)
		})
	}
}

func TestAdd(t *testing.T) {
	tests := map[string]struct {
		Metrics      []telegraf.Metric
		ReceivedData map[HostLLD]LLDTags
	}{
		"metric without tags are ignored": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{},
		},
		"metric without tag host are ignored": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"foo": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{},
		},
		"metric without only tag host are ignored": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{},
		},
		"simple add of one metric with one extra tag": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{
				{"bar", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"same metric with different field values is only stored once": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo": "bar"}, map[string]interface{}{"a": 999}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{
				{"bar", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"for the same measurement and tags, the different combinations of tag values are stored under the same key": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo": "bar1"}, map[string]interface{}{"a": 0}, time.Now()),
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo": "bar2"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{
				{"bar", "disk.foo"}: {
					map[string]string{
						"foo": "bar1",
					},
					map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		"same measurement and tags for different hosts are stored in different keys": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "barA", "foo": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
				testutil.MustMetric("disk", map[string]string{"host": "barB", "foo": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{
				{"barA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
				{"barB", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"different number of tags for the same measurement are stored in different keys": {
			Metrics: []telegraf.Metric{
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo1": "bar", "foo2": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
				testutil.MustMetric("disk", map[string]string{"host": "bar", "foo1": "bar"}, map[string]interface{}{"a": 0}, time.Now()),
			},
			ReceivedData: map[HostLLD]LLDTags{
				{"bar", "disk.foo1.foo2"}: {
					map[string]string{
						"foo1": "bar",
						"foo2": "bar",
					},
				},
				{"bar", "disk.foo1"}: {
					map[string]string{
						"foo1": "bar",
					},
				},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := NewZabbixLLD()
			for _, m := range test.Metrics {
				agg.Add(m)
			}

			assert.Equal(t, test.ReceivedData, agg.receivedData)
		})
	}
}

func TestPush(t *testing.T) {
	tests := map[string]struct {
		ReceivedData         map[HostLLD]LLDTags
		PreviousReceivedData map[HostLLD]LLDTags
		Metrics              []telegraf.Metric
	}{
		"an empty ReceivedData does not generate any metric": {
			ReceivedData:         map[HostLLD]LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics:              []telegraf.Metric{},
		},
		"simple one host with one lld with one set of values": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"one host with one lld with two set of values": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar1",
					},
					map[string]string{
						"foo": "bar2",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar1"},{"{#FOO}":"bar2"}]}`},
					time.Now(),
				),
			},
		},
		"one host with one lld with one multiset of values": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.fooA.fooB.fooC"}: {
					map[string]string{
						"fooA": "bar1",
						"fooB": "bar2",
						"fooC": "bar3",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.fooA.fooB.fooC": `{"data":[{"{#FOOA}":"bar1","{#FOOB}":"bar2","{#FOOC}":"bar3"}]}`},
					time.Now(),
				),
			},
		},
		"one host with three lld with one set of values, not sorted": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
				{"hostA", "net.iface"}: {
					map[string]string{
						"iface": "eth0",
					},
				},
				{"hostA", "proc.pid"}: {
					map[string]string{
						"pid": "1234",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"proc.pid": `{"data":[{"{#PID}":"1234"}]}`},
					time.Now(),
				),
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"net.iface": `{"data":[{"{#IFACE}":"eth0"}]}`},
					time.Now(),
				),
			},
		},
		"two host with the same lld with one set of values": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
				{"hostB", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostB"},
					map[string]interface{}{"disk.foo": `{"data":[{"{#FOO}":"bar"}]}`},
					time.Now(),
				),
			},
		},
		"ignore generating a new lld if it was sent the last time": {
			ReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Metrics: []telegraf.Metric{},
		},
		"send an empty LLD if one metric has stopped being sent": {
			ReceivedData: map[HostLLD]LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "disk.foo"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Metrics: []telegraf.Metric{
				testutil.MustMetric(LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"disk.foo": `{"data":[]}`},
					time.Now(),
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := NewZabbixLLD()
			agg.receivedData = test.ReceivedData
			agg.previousReceivedData = test.PreviousReceivedData
			acc := testutil.Accumulator{}
			agg.Push(&acc)

			testutil.RequireMetricsEqual(t, test.Metrics, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestCompareAndDelete(t *testing.T) {
	tests := map[string]struct {
		Key                          HostLLD
		Tags                         LLDTags
		PreviousReceivedData         map[HostLLD]LLDTags
		Result                       bool
		PostFuncPreviousReceivedData map[HostLLD]LLDTags
	}{
		"an empty PreviousReceivedData always returns false": {
			Key:                          HostLLD{"hostA", "foo.bar"},
			Tags:                         LLDTags{},
			PreviousReceivedData:         map[HostLLD]LLDTags{},
			Result:                       false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"a key not present in PreviousReceivedData does not delete anything and returns false": {
			Key:  HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostB", "aaa.bbb"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result: false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{
				{"hostB", "aaa.bbb"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"if LLD key matches but host does not, returns false": {
			Key:  HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostB", "foo.bar"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result: false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{
				{"hostB", "foo.bar"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"if host matches, but LLD key matches does not, returns false": {
			Key:  HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "aaa.bbb"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result: false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "aaa.bbb"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"if host and LLD key matches, tags empty in arg and previousReceivedData, return true and delete entry": {
			Key:  HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {},
			},
			Result:                       true,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"if host and LLD key matches, but tags are not equal, returns false, delete entry": {
			Key:  HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result:                       false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"if host and LLD key matches, but tags values are different, returns false, delete entry": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"aaa": "bbb",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result:                       false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"if host and LLD key matches, tags partially match, returns false, delete entry": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"foo": "bar",
				},
				map[string]string{
					"aaa": "bbb",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result:                       false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"if tag values are different, result is false and entry in previousReceivedData is deleted": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"aaa": "bbb",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"aaa": "bbb1",
					},
				},
			},
			Result:                       false,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"if everything matches, return true and delete the entry in previousReceivedData": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"aaa": "bbb",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"aaa": "bbb",
					},
				},
			},
			Result:                       true,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"different tag ordering also match": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"foo": "aaa1",
					"bar": "bbb1",
				},
				map[string]string{
					"foo": "aaa2",
					"bar": "bbb2",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"bar": "bbb2",
						"foo": "aaa2",
					},
					map[string]string{
						"bar": "bbb1",
						"foo": "aaa1",
					},
				},
			},
			Result:                       true,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"different tag ordering inside a set also match": {
			Key: HostLLD{"hostA", "foo.bar"},
			Tags: LLDTags{
				map[string]string{
					"aaa1": "bbb1",
					"aaa2": "bbb2",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar"}: {
					map[string]string{
						"aaa2": "bbb2",
						"aaa1": "bbb1",
					},
				},
			},
			Result:                       true,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{},
		},
		"match and delete one entry in previousReceivedData": {
			Key: HostLLD{"hostA", "foo.bar1"},
			Tags: LLDTags{
				map[string]string{
					"foo": "bar",
				},
			},
			PreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar1"}: {
					map[string]string{
						"foo": "bar",
					},
				},
				{"hostA", "foo.bar2"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
			Result: true,
			PostFuncPreviousReceivedData: map[HostLLD]LLDTags{
				{"hostA", "foo.bar2"}: {
					map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := NewZabbixLLD()
			agg.previousReceivedData = test.PreviousReceivedData
			equal := agg.compareAndDelete(test.Key, test.Tags)

			assert.Equal(t, test.Result, equal)
			assert.Equal(t, test.PostFuncPreviousReceivedData, agg.previousReceivedData)
		})
	}
}

func TestAddAndPush(t *testing.T) {
	tests := map[string][]Operations{
		"simple Add, Push and check generated LLD metric": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
		},
		"same metric several times generate only one LLD": {
			OperationAdd{
				testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now()),
				testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now()),
				testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now()),
			},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
		},
		"after sending correctly an LLD, same tag values does not generate the same LLD": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{},
		},
		"in the Nth push, all status is reseted to be able to resend again already seen LLDs": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
		},
		"is one input stop sending metrics, an empty LLD is sent": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now())},
			OperationPush{},
			OperationCheck{testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[]}`}, time.Now())},
		},
		"different hosts sending the same metric should generate different LLDs": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostB", "foo": "bar"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now()),
				testutil.MustMetric(LLDName, map[string]string{"host": "hostB"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"bar"}]}`}, time.Now()),
			},
		},
		"same measurement with different tags should generate different LLDs": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "a"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "a", "bar": "b"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.foo": `{"data":[{"{#FOO}":"a"}]}`}, time.Now()),
				testutil.MustMetric(LLDName, map[string]string{"host": "hostA"}, map[string]interface{}{"name.bar.foo": `{"data":[{"{#BAR}":"b","{#FOO}":"a"}]}`}, time.Now()),
			},
		},
		"a set with a new combination of tag values already seen should generate a new lld": {
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "a", "bar": "b"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "x", "bar": "y"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationAdd{testutil.MustMetric("name", map[string]string{"host": "hostA", "foo": "a", "bar": "y"}, map[string]interface{}{"value": 1}, time.Now())},
			OperationPush{},
			OperationCheck{
				testutil.MustMetric(
					LLDName,
					map[string]string{"host": "hostA"},
					map[string]interface{}{"name.bar.foo": `{"data":[{"{#BAR}":"b","{#FOO}":"a"},{"{#BAR}":"y","{#FOO}":"x"},{"{#BAR}":"y","{#FOO}":"a"}]}`},
					time.Now(),
				),
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			agg := NewZabbixLLD()
			// Set to 4 to be test quicker the force push
			agg.ResetPeriod = 4
			acc := testutil.Accumulator{}

			for _, op := range test {
				switch o := (op).(type) {
				case OperationAdd:
					for _, metric := range o {
						agg.Add(metric)
					}
				case OperationPush:
					agg.Push(&acc)
					agg.Reset() // Telegraf calls Reset() after Push()
				case OperationCheck:
					testutil.RequireMetricsEqual(t, o, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
					acc.ClearMetrics()
				}
			}
		})
	}
}
