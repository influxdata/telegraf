package t128_graphql

import (
	"strings"
)

//Config stores paths to fields, tags and predicates to be used by BuildQuery and ProcessResponse
type Config struct {
	Predicates map[string]string
	Fields     map[string]string
	Tags       map[string]string
}

/*
LoadConfig converts a telegraf config into paths to predicates, fields and tags to be used by BuildQuery and ProcessResponse
Paths correspond to keys in the output to avoid collisions and because they are used as lookups in ProcessResponse

Args:
	entryPoint example - "allRouters(name:'ComboEast')/nodes/nodes(name:'east-combo')/nodes/arp/nodes"
	fieldsIn example - map[string]string{"test-field": "test-field"}
	tagsIn example - map[string]string{"test-tag": "test-tag"}

Example:
	For the example input above, LoadConfig() will produce the following Config

	*Config{
		Predicates: map[string]string{
			".data.allRouters.$predicate":             "(name:\"ComboEast\")",
			".data.allRouters.nodes.nodes.$predicate": "(name:\"east-combo\")",
		},
		Fields: map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-field": "test-field"},
		Tags:   map[string]string{".data.allRouters.nodes.nodes.nodes.arp.nodes.test-tag": "test-tag"},
	}
*/
func LoadConfig(entryPoint string, fieldsIn map[string]string, tagsIn map[string]string) *Config {
	config := &Config{}
	path := ".data."
	predicates := map[string]string{}

	pathElements := strings.Split(entryPoint, "/")
	for _, element := range pathElements {
		parenthesisIdx := strings.Index(element, "(")
		if parenthesisIdx > 0 {
			path += element[:parenthesisIdx]
			predicatePath := path + "." + predicateTag + "predicate"
			predicates[predicatePath] = formatPredicate(element[parenthesisIdx:])
			path += "."
		} else {
			path += element + "."
		}
	}

	config.Predicates = predicates
	config.Fields = formatPaths(fieldsIn, path)
	config.Tags = formatPaths(tagsIn, path)

	return config
}

//needed because users configure tags & fields with paths starting at entry_point with "/" instead of "."
func formatPaths(items map[string]string, basePath string) map[string]string {
	newMap := make(map[string]string)
	replacer := strings.NewReplacer("/", ".")
	for name, partialPath := range items {
		newMap[basePath+replacer.Replace(partialPath)] = name
	}
	return newMap
}

//needed to strip whitespace and to replace ' with \"
func formatPredicate(predicate string) string {
	replacer := strings.NewReplacer(" ", "", "'", "\"")
	return replacer.Replace(predicate)
}
