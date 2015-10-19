package telegraf

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/plugins/exec"
	"github.com/influxdb/telegraf/plugins/kafka_consumer"
	"github.com/influxdb/telegraf/plugins/procstat"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestConfig_fieldMatch(t *testing.T) {
	assert := assert.New(t)

	matchFunc := fieldMatch("testfield")
	assert.True(matchFunc("testField"), "testfield should match testField")
	assert.True(matchFunc("TestField"), "testfield should match TestField")
	assert.True(matchFunc("TESTFIELD"), "testfield should match TESTFIELD")
	assert.False(matchFunc("OtherField"), "testfield should not match OtherField")

	matchFunc = fieldMatch("test_field")
	assert.True(matchFunc("testField"), "test_field should match testField")
	assert.True(matchFunc("TestField"), "test_field should match TestField")
	assert.True(matchFunc("TESTFIELD"), "test_field should match TESTFIELD")
	assert.False(matchFunc("OtherField"), "test_field should not match OtherField")
}

type subTest struct {
	AField       string
	AnotherField int
}
type test struct {
	StringField     string
	IntegerField    int
	FloatField      float32
	BooleanField    bool
	DatetimeField   time.Time
	ArrayField      []string
	TableArrayField []subTest
}

type MergeStructSuite struct {
	suite.Suite
	EmptyStruct       *test
	FullStruct        *test
	AnotherFullStruct *test
	AllFields         []string
}

func (s *MergeStructSuite) SetupSuite() {
	s.AllFields = []string{"string_field", "integer_field", "float_field", "boolean_field", "date_time_field", "array_field", "table_array_field"}
}

func (s *MergeStructSuite) SetupTest() {
	s.EmptyStruct = &test{
		ArrayField:      []string{},
		TableArrayField: []subTest{},
	}
	s.FullStruct = &test{
		StringField:   "one",
		IntegerField:  1,
		FloatField:    1.1,
		BooleanField:  false,
		DatetimeField: time.Date(1963, time.August, 28, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"one", "two", "three"},
		TableArrayField: []subTest{
			subTest{
				AField:       "one",
				AnotherField: 1,
			},
			subTest{
				AField:       "two",
				AnotherField: 2,
			},
		},
	}
	s.AnotherFullStruct = &test{
		StringField:   "two",
		IntegerField:  2,
		FloatField:    2.2,
		BooleanField:  true,
		DatetimeField: time.Date(1965, time.March, 25, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"four", "five", "six"},
		TableArrayField: []subTest{
			subTest{
				AField:       "three",
				AnotherField: 3,
			},
			subTest{
				AField:       "four",
				AnotherField: 4,
			},
		},
	}
}

func (s *MergeStructSuite) TestEmptyMerge() {
	err := mergeStruct(s.EmptyStruct, s.FullStruct, s.AllFields)
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(s.FullStruct, s.EmptyStruct, fmt.Sprintf("Full merge of %v onto an empty struct failed.", s.FullStruct))
}

func (s *MergeStructSuite) TestFullMerge() {
	result := &test{
		StringField:   "two",
		IntegerField:  2,
		FloatField:    2.2,
		BooleanField:  true,
		DatetimeField: time.Date(1965, time.March, 25, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"one", "two", "three", "four", "five", "six"},
		TableArrayField: []subTest{
			subTest{
				AField:       "one",
				AnotherField: 1,
			},
			subTest{
				AField:       "two",
				AnotherField: 2,
			},
			subTest{
				AField:       "three",
				AnotherField: 3,
			},
			subTest{
				AField:       "four",
				AnotherField: 4,
			},
		},
	}

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct, s.AllFields)
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct, fmt.Sprintf("Full merge of %v onto FullStruct failed.", s.AnotherFullStruct))
	s.T().Log("hi")
}

func (s *MergeStructSuite) TestPartialMergeWithoutSlices() {
	result := &test{
		StringField:   "two",
		IntegerField:  1,
		FloatField:    2.2,
		BooleanField:  false,
		DatetimeField: time.Date(1965, time.March, 25, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"one", "two", "three"},
		TableArrayField: []subTest{
			subTest{
				AField:       "one",
				AnotherField: 1,
			},
			subTest{
				AField:       "two",
				AnotherField: 2,
			},
		},
	}

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct, []string{"string_field", "float_field", "date_time_field"})
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct, fmt.Sprintf("Partial merge without slices of %v onto FullStruct failed.", s.AnotherFullStruct))
}

func (s *MergeStructSuite) TestPartialMergeWithSlices() {
	result := &test{
		StringField:   "two",
		IntegerField:  1,
		FloatField:    2.2,
		BooleanField:  false,
		DatetimeField: time.Date(1965, time.March, 25, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"one", "two", "three"},
		TableArrayField: []subTest{
			subTest{
				AField:       "one",
				AnotherField: 1,
			},
			subTest{
				AField:       "two",
				AnotherField: 2,
			},
			subTest{
				AField:       "three",
				AnotherField: 3,
			},
			subTest{
				AField:       "four",
				AnotherField: 4,
			},
		},
	}

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct, []string{"string_field", "float_field", "date_time_field", "table_array_field"})
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct, fmt.Sprintf("Partial merge with slices of %v onto FullStruct failed.", s.AnotherFullStruct))
}

func TestConfig_mergeStruct(t *testing.T) {
	suite.Run(t, new(MergeStructSuite))
}

func TestConfig_parsePlugin(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/single_plugin.toml")
	if err != nil {
		t.Error(err)
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		t.Error(err)
	}

	c := &Config{
		plugins:                      make(map[string]plugins.Plugin),
		pluginConfigurations:         make(map[string]*ConfiguredPlugin),
		pluginFieldsSet:              make(map[string][]string),
		pluginConfigurationFieldsSet: make(map[string][]string),
	}

	subtbl := tbl.Fields["kafka"].(*ast.Table)
	err = c.parsePlugin("kafka", subtbl)

	kafka := plugins.Plugins["kafka"]().(*kafka_consumer.Kafka)
	kafka.ConsumerGroupName = "telegraf_metrics_consumers"
	kafka.Topic = "topic_with_metrics"
	kafka.ZookeeperPeers = []string{"test.example.com:2181"}
	kafka.BatchSize = 1000

	kConfig := &ConfiguredPlugin{
		Name: "kafka",
		Drop: []string{"other", "stuff"},
		Pass: []string{"some", "strings"},
		TagDrop: []TagFilter{
			TagFilter{
				Name:   "badtag",
				Filter: []string{"othertag"},
			},
		},
		TagPass: []TagFilter{
			TagFilter{
				Name:   "goodtag",
				Filter: []string{"mytag"},
			},
		},
		Interval: 5 * time.Second,
	}

	assert.Equal(t, kafka, c.plugins["kafka"], "Testdata did not produce a correct kafka struct.")
	assert.Equal(t, kConfig, c.pluginConfigurations["kafka"], "Testdata did not produce correct kafka metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c, err := LoadConfig("./testdata/telegraf-agent.toml")
	if err != nil {
		t.Error(err)
	}
	err = c.LoadDirectory("./testdata/subconfig")
	if err != nil {
		t.Error(err)
	}

	kafka := plugins.Plugins["kafka"]().(*kafka_consumer.Kafka)
	kafka.ConsumerGroupName = "telegraf_metrics_consumers"
	kafka.Topic = "topic_with_metrics"
	kafka.ZookeeperPeers = []string{"localhost:2181", "test.example.com:2181"}
	kafka.BatchSize = 10000

	kConfig := &ConfiguredPlugin{
		Name: "kafka",
		Drop: []string{"other", "stuff"},
		Pass: []string{"some", "strings"},
		TagDrop: []TagFilter{
			TagFilter{
				Name:   "badtag",
				Filter: []string{"othertag"},
			},
		},
		TagPass: []TagFilter{
			TagFilter{
				Name:   "goodtag",
				Filter: []string{"mytag"},
			},
		},
		Interval: 5 * time.Second,
	}

	ex := plugins.Plugins["exec"]().(*exec.Exec)
	ex.Commands = []*exec.Command{
		&exec.Command{
			Command: "/usr/bin/mycollector --foo=bar",
			Name:    "mycollector",
		},
		&exec.Command{
			Command: "/usr/bin/myothercollector --foo=bar",
			Name:    "myothercollector",
		},
	}

	eConfig := &ConfiguredPlugin{Name: "exec"}

	pstat := plugins.Plugins["procstat"]().(*procstat.Procstat)
	pstat.Specifications = []*procstat.Specification{
		&procstat.Specification{
			PidFile: "/var/run/grafana-server.pid",
		},
		&procstat.Specification{
			PidFile: "/var/run/influxdb/influxd.pid",
		},
	}

	pConfig := &ConfiguredPlugin{Name: "procstat"}

	assert.Equal(t, kafka, c.plugins["kafka"], "Merged Testdata did not produce a correct kafka struct.")
	assert.Equal(t, kConfig, c.pluginConfigurations["kafka"], "Merged Testdata did not produce correct kafka metadata.")

	assert.Equal(t, ex, c.plugins["exec"], "Merged Testdata did not produce a correct exec struct.")
	assert.Equal(t, eConfig, c.pluginConfigurations["exec"], "Merged Testdata did not produce correct exec metadata.")

	assert.Equal(t, pstat, c.plugins["procstat"], "Merged Testdata did not produce a correct procstat struct.")
	assert.Equal(t, pConfig, c.pluginConfigurations["procstat"], "Merged Testdata did not produce correct procstat metadata.")
}
