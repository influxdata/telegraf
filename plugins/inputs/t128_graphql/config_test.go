package t128_graphql_test

import (
	"testing"

	plugin "github.com/influxdata/telegraf/plugins/inputs/t128_graphql"
	"github.com/stretchr/testify/require"
)

var JSONPathFormationTestCases = []struct {
	Name           string
	EntryPoint     string
	Fields         map[string]string
	Tags           map[string]string
	ExpectedOutput *plugin.Config
}{
	{
		Name:           "process simple input",
		EntryPoint:     "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:         getTestFields(),
		Tags:           getTestTags(),
		ExpectedOutput: getTestConfigWithPredicates("(name:\"ComboEast\")", "(name:\"east-combo\")"),
	},
	{
		Name:           "process predicate with list",
		EntryPoint:     "allRouters(names:['wan','lan'])/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:         getTestFields(),
		Tags:           getTestTags(),
		ExpectedOutput: getTestConfigWithPredicates("(names:[\"wan\",\"lan\"])", "(name:\"east-combo\")"),
	},
	{
		Name:           "process multi-value predicates",
		EntryPoint:     "allRouters(names:['wan', 'lan'], key2:'value2')/nodes/nodes(name:'east-combo')/nodes/arp/nodes",
		Fields:         getTestFields(),
		Tags:           getTestTags(),
		ExpectedOutput: getTestConfigWithPredicates("(names:[\"wan\",\"lan\"],key2:\"value2\")", "(name:\"east-combo\")"),
	},
	{
		Name:           "process complex config",
		EntryPoint:     "allRouters(names:['wan', 'lan'], key2:'value2')/nodes/nodes(names:['east-combo', 'west-combo'])/nodes/arp/nodes",
		Fields:         getTestFields(),
		Tags:           getTestTags(),
		ExpectedOutput: getTestConfigWithPredicates("(names:[\"wan\",\"lan\"],key2:\"value2\")", "(names:[\"east-combo\",\"west-combo\"])"),
	},
}

func getTestConfigWithPredicates(pred1 string, pred2 string) *plugin.Config {
	return &plugin.Config{
		Predicates: map[string]string{
			".data.allRouters.$predicate":             pred1,
			".data.allRouters.nodes.nodes.$predicate": pred2,
		},
		Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
		Tags:   map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag": "test-tag"},
	}
}

func TestT128GraphqlEntryPointParsing(t *testing.T) {
	for _, testCase := range JSONPathFormationTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			parsedEntryPoint := plugin.LoadConfig(testCase.EntryPoint, testCase.Fields, testCase.Tags)
			require.Equal(t, testCase.ExpectedOutput, parsedEntryPoint)
		})
	}
}

func getTestFields() map[string]string {
	return map[string]string{"test-field": "test-field"}
}

func getTestTags() map[string]string {
	return map[string]string{"test-tag": "test-tag"}
}
