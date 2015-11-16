package telegraf

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/plugins/exec"
	"github.com/influxdb/telegraf/plugins/memcached"
	"github.com/influxdb/telegraf/plugins/procstat"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type subTest struct {
	AField       string
	AnotherField int
}
type test struct {
	StringField     string
	IntegerField    int
	FloatField      float32
	BooleansField   bool `toml:"boolean_field"`
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
	s.AllFields = []string{"string_field", "integer_field", "float_field",
		"boolean_field", "date_time_field", "array_field", "table_array_field"}
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
		BooleansField: false,
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
		BooleansField: true,
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
	s.Equal(s.FullStruct, s.EmptyStruct,
		fmt.Sprintf("Full merge of %v onto an empty struct failed.", s.FullStruct))
}

func (s *MergeStructSuite) TestFullMerge() {
	result := &test{
		StringField:   "two",
		IntegerField:  2,
		FloatField:    2.2,
		BooleansField: true,
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

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct, s.AllFields)
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct,
		fmt.Sprintf("Full merge of %v onto FullStruct failed.", s.AnotherFullStruct))
}

func (s *MergeStructSuite) TestPartialMergeWithoutSlices() {
	result := &test{
		StringField:   "two",
		IntegerField:  1,
		FloatField:    2.2,
		BooleansField: false,
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

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct,
		[]string{"string_field", "float_field", "date_time_field"})
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct,
		fmt.Sprintf("Partial merge without slices of %v onto FullStruct failed.",
			s.AnotherFullStruct))
}

func (s *MergeStructSuite) TestPartialMergeWithSlices() {
	result := &test{
		StringField:   "two",
		IntegerField:  1,
		FloatField:    2.2,
		BooleansField: false,
		DatetimeField: time.Date(1965, time.March, 25, 17, 0, 0, 0, time.UTC),
		ArrayField:    []string{"one", "two", "three"},
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

	err := mergeStruct(s.FullStruct, s.AnotherFullStruct,
		[]string{"string_field", "float_field", "date_time_field", "table_array_field"})
	if err != nil {
		s.T().Error(err)
	}
	s.Equal(result, s.FullStruct,
		fmt.Sprintf("Partial merge with slices of %v onto FullStruct failed.",
			s.AnotherFullStruct))
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

	subtbl := tbl.Fields["memcached"].(*ast.Table)
	err = c.parsePlugin("memcached", subtbl)

	memcached := plugins.Plugins["memcached"]().(*memcached.Memcached)
	memcached.Servers = []string{"localhost"}

	mConfig := &ConfiguredPlugin{
		Name: "memcached",
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

	assert.Equal(t, memcached, c.plugins["memcached"],
		"Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.pluginConfigurations["memcached"],
		"Testdata did not produce correct memcached metadata.")
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

	memcached := plugins.Plugins["memcached"]().(*memcached.Memcached)
	memcached.Servers = []string{"192.168.1.1"}

	mConfig := &ConfiguredPlugin{
		Name: "memcached",
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

	assert.Equal(t, memcached, c.plugins["memcached"],
		"Merged Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.pluginConfigurations["memcached"],
		"Merged Testdata did not produce correct memcached metadata.")

	assert.Equal(t, ex, c.plugins["exec"],
		"Merged Testdata did not produce a correct exec struct.")
	assert.Equal(t, eConfig, c.pluginConfigurations["exec"],
		"Merged Testdata did not produce correct exec metadata.")

	assert.Equal(t, pstat, c.plugins["procstat"],
		"Merged Testdata did not produce a correct procstat struct.")
	assert.Equal(t, pConfig, c.pluginConfigurations["procstat"],
		"Merged Testdata did not produce correct procstat metadata.")
}
