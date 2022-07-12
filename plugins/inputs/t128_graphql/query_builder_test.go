package t128_graphql_test

import (
	"strings"
	"testing"

	plugin "github.com/influxdata/telegraf/plugins/inputs/t128_graphql"
	"github.com/stretchr/testify/require"
)

const (
	ValidQueryDoubleTag   = "query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field\ntest-tag-1\ntest-tag-2}}}}}}}"
	ValidQueryDoubleField = "query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\ntest-field-1\ntest-field-2\ntest-tag}}}}}}}"
	ValidQueryNestedTag   = "query {\nallRouters(name:\"ComboEast\"){\nnodes{\nnodes(name:\"east-combo\"){\nnodes{\narp{\nnodes{\nstate{\ntest-tag-2}\ntest-field\ntest-tag-1}}}}}}}"
)

var QueryFormationTestCases = []struct {
	Name          string
	ConfigIn      *plugin.Config
	ExpectedQuery string
}{
	{
		Name:          "build simple query single tag",
		ConfigIn:      getTestConfigWithPredicates("(name:\"ComboEast\")", "(name:\"east-combo\")"),
		ExpectedQuery: ValidQuerySingleTag,
	},
	{
		Name: "build simple query double tag",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(name:\"ComboEast\")",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag-1": "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: ValidQueryDoubleTag,
	},
	{
		Name: "build simple query double field",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(name:\"ComboEast\")",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field-1": "test-field-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field-2": "test-field-2",
			},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag": "test-tag",
			},
		},
		ExpectedQuery: ValidQueryDoubleField,
	},
	{
		Name: "build query nested tag",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(name:\"ComboEast\")",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag-1":       "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: ValidQueryNestedTag,
	},
	{
		Name: "build query multi-level-nested tag",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(name:\"ComboEast\")",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.test-tag-1":               "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.state2.state3.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: strings.ReplaceAll(`query {
			allRouters(name:"ComboEast"){
			nodes{
			nodes(name:"east-combo"){
			nodes{
			arp{
			nodes{
			state1{
			state2{
			state3{
			test-tag-2}}
			test-tag-1}
			test-field}}}}}}}`, "\t", ""),
	},
	{
		Name: "build query list predicate",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(names:[\"wan\",\"lan\"])",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.test-tag-1":               "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.state2.state3.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: strings.ReplaceAll(`query {
			allRouters(names:["wan","lan"]){
			nodes{
			nodes(name:"east-combo"){
			nodes{
			arp{
			nodes{
			state1{
			state2{
			state3{
			test-tag-2}}
			test-tag-1}
			test-field}}}}}}}`, "\t", ""),
	},
	{
		Name: "build query multi-value predicates",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(names:[\"wan\",\"lan\"],key2:\"value2\")",
				".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.test-tag-1":               "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.state2.state3.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: strings.ReplaceAll(`query {
			allRouters(names:["wan","lan"],key2:"value2"){
			nodes{
			nodes(name:"east-combo"){
			nodes{
			arp{
			nodes{
			state1{
			state2{
			state3{
			test-tag-2}}
			test-tag-1}
			test-field}}}}}}}`, "\t", ""),
	},
	{
		Name: "build complex",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allRouters.$predicate":             "(names:[\"wan\",\"lan\"],key2:\"value2\")",
				".data.allRouters.nodes.nodes.$predicate": "(names:[\"east-combo\",\"west-combo\"])",
			},
			Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
			Tags: map[string]string{
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.test-tag-1":               "test-tag-1",
				".data.allRouters.nodes.nodes.nodes.arp.nodes.state1.state2.state3.test-tag-2": "test-tag-2",
			},
		},
		ExpectedQuery: strings.ReplaceAll(`query {
			allRouters(names:["wan","lan"],key2:"value2"){
			nodes{
			nodes(names:["east-combo","west-combo"]){
			nodes{
			arp{
			nodes{
			state1{
			state2{
			state3{
			test-tag-2}}
			test-tag-1}
			test-field}}}}}}}`, "\t", ""),
	},
	{
		Name: "build query with absolute path tags and fields",
		ConfigIn: &plugin.Config{
			Predicates: map[string]string{
				".data.allServices.nodes.timeSeriesAnalytic.$predicate": "(metric:BANDWIDTH,router:\"${ROUTER}\",transform:AVERAGE,resolution:1000,startTime:\"now-180\",endTime:\"now\")",
			},
			Fields: map[string]string{
				".data.allServices.nodes.timeSeriesAnalytic.value":     "value",
				".data.allServices.nodes.timeSeriesAnalytic.timestamp": "timestamp",
				".data.allServices.nodes.other":                        "other-field",
			},
			Tags: map[string]string{
				".data.allServices.nodes.timeSeriesAnalytic.test-tag": "test-tag",
				".data.allServices.nodes.name":                        "name",
			},
		},
		ExpectedQuery: strings.ReplaceAll(`query {
			allServices{
			nodes{
			name
			other
			timeSeriesAnalytic(metric:BANDWIDTH,router:"${ROUTER}",transform:AVERAGE,resolution:1000,startTime:"now-180",endTime:"now"){
			test-tag
			timestamp
			value}}}}`, "\t", ""),
	},
}

func TestT128GraphqlQueryFormation(t *testing.T) {
	for _, testCase := range QueryFormationTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			query := plugin.BuildQuery(testCase.ConfigIn)
			require.Equal(t, testCase.ExpectedQuery, query)
		})
	}
}
