package t128_graphql

import (
	"fmt"
	"reflect"

	"github.com/Jeffail/gabs"
)

//ProcessedResponse stores the processed fields and tags for injection into telegraf accumulator
type ProcessedResponse struct {
	Fields map[string]interface{}
	Tags   map[string]string
}

/*
ProcessResponse takes in a query response, pulls out the desired data and stores it in a struct. To find the data,
ProcessResponse traverses the response tree recursively and appropriately merges the data returned by each child.

Args:
	jsonData example - see example below
	fields example - map[string]string{
			"/data/allRouters/nodes/peers/nodes/paths/adjacentAddress": "adjacent-address",
			"/data/allRouters/nodes/peers/nodes/paths/uptime": "uptime"
		}
	tags example - map[string]string{
			"/data/allRouters/nodes/peers/nodes/name": "peer-name",
		}

Example:
	For the the input given above along with the following jsonData input

	{"data":
		{
			"allRouters": {
			"nodes": [
				{
				"peers": {
					"nodes": [
					{
						"name": "AZDCBBP1",
						"paths": [
							{
								"uptime": 188333176,
								"adjacentAddress": "12.51.52.30"
							},
							{
								"uptime": 82247253,
								"adjacentAddress": "12.51.52.30"
							}
						]
					},
					{
						"name": "AZDCLTEP1",
						"paths": [
							{
								"uptime": 162241794,
								"adjacentAddress": "12.51.52.22"
							},
							{
								"uptime": 82247352,
								"adjacentAddress": "12.51.52.22"
							}
						]
					}]
				}}
			]
		}
	}

	ProcessResponse() will produce the following output

	[]*plugin.ProcessedResponse{
		&plugin.ProcessedResponse{
			Fields: map[string]interface{}{"adjacent-address": "12.51.52.30", "uptime": 188333176.0},
			Tags:   map[string]string{"peer-name": "AZDCBBP1"},
		},
		&plugin.ProcessedResponse{
			Fields: map[string]interface{}{"adjacent-address": "12.51.52.30", "uptime": 82247253.0},
			Tags:   map[string]string{"peer-name": "AZDCBBP1"},
		},
		&plugin.ProcessedResponse{
			Fields: map[string]interface{}{"adjacent-address": "12.51.52.22", "uptime": 162241794.0},
			Tags:   map[string]string{"peer-name": "AZDCLTEP1"},
		},
		&plugin.ProcessedResponse{
			Fields: map[string]interface{}{"adjacent-address": "12.51.52.22", "uptime": 82247352.0},
			Tags:   map[string]string{"peer-name": "AZDCLTEP1"},
		},
	}

Definitions:
	leaf - a single tag/field stored in *ProcessedResponse
	branch - any *ProcessedResponse that isn't a leaf
*/
func ProcessResponse(jsonData *gabs.Container, collector string, fields map[string]string, tags map[string]string) ([]*ProcessedResponse, error) {
	processedResponses := processNode(jsonData, "", fields, tags)

	if len(processedResponses) == 1 && getResponseSize(processedResponses[0]) == 0 {
		return nil, fmt.Errorf("no data collected for collector %s", collector)
	}

	return processedResponses, nil
}

func processNode(jsonData *gabs.Container, path string, fields map[string]string, tags map[string]string) []*ProcessedResponse {
	processedNode, err := processChildren(jsonData, "map", path, fields, tags)
	if err == nil {
		return processedNode
	}

	processedNode, err = processChildren(jsonData, "list", path, fields, tags)
	if err == nil {
		return processedNode
	}

	leaf, err := collectLeaf(jsonData.Data(), "field", path, fields)
	if err == nil {
		processedNode = append(processedNode, leaf)
		return processedNode
	}

	leaf, err = collectLeaf(jsonData.Data(), "tag", path, tags)
	if err == nil {
		processedNode = append(processedNode, leaf)
	}

	return processedNode
}

func processChildren(jsonData *gabs.Container, mode string, path string, fields map[string]string, tags map[string]string) ([]*ProcessedResponse, error) {
	output := []*ProcessedResponse{}

	processChild := func(child *gabs.Container, path string) {
		processedChild := processNode(child, path, fields, tags)
		for _, mergedChildOutput := range mergeAll(processedChild) {
			output = append(output, mergedChildOutput)
		}
	}

	processAsMap := func() ([]*ProcessedResponse, error) {
		children, err := jsonData.ChildrenMap()
		if err == nil {
			for key, child := range children {
				processChild(child, path+"."+key)
			}
			return output, nil
		}
		return nil, fmt.Errorf("could not process map")
	}

	processAsList := func() ([]*ProcessedResponse, error) {
		children, err := jsonData.Children()
		if err == nil {
			for _, child := range children {
				processChild(child, path)
			}
			return output, nil
		}
		return nil, fmt.Errorf("could not process list")
	}

	if mode == "map" {
		return processAsMap()
	}
	return processAsList()
}

//Definitions:
//leaf - a single tag/field stored in *ProcessedResponse
//branch - any *ProcessedResponse that isn't a leaf
func collectLeaf(leaf interface{}, mode string, path string, lookup map[string]string) (*ProcessedResponse, error) {
	output := newResponse()

	if leafName, ok := lookup[path]; ok {
		if !isNil(leaf) {
			if mode == "field" {
				output.Fields[leafName] = leaf
			} else {
				output.Tags[leafName] = fmt.Sprintf("%v", leaf)
			}
			return output, nil
		}
	}
	return nil, fmt.Errorf("could not collect leaf")
}

//merges all leaves/branches at a given node
func mergeAll(itemsToMerge []*ProcessedResponse) []*ProcessedResponse {
	leaves := []*ProcessedResponse{}
	branches := []*ProcessedResponse{}

	//fill branches and leaves structs
	for _, response := range itemsToMerge {
		if getResponseSize(response) > 1 {
			branches = append(branches, response)
		} else {
			leaves = append(leaves, response)
		}
	}

	//mergeLeaves if all itemsToMerge are leaves
	if len(branches) == 0 {
		return mergeLeaves(leaves)
	}

	//mergeLeafIntoBranch if not all itemsToMerge are leaves
	for _, branch := range branches {
		for _, leaf := range leaves {
			mergeLeafIntoBranch(leaf, branch)
		}
	}

	return branches
}

//used when a node only has leaves
func mergeLeaves(leaves []*ProcessedResponse) []*ProcessedResponse {
	mergedLeaves := newResponse()
	for _, leaf := range leaves {
		_, err := mergeFields(leaf, mergedLeaves)
		if err != nil {
			return leaves
		}

		_, err = mergeTags(leaf, mergedLeaves)
		if err != nil {
			return leaves
		}
	}
	return []*ProcessedResponse{mergedLeaves}
}

//used when a node has branches and leaves
func mergeLeafIntoBranch(leaf *ProcessedResponse, branch *ProcessedResponse) *ProcessedResponse {
	newBranch := branch
	success, err := mergeFields(leaf, newBranch)
	if success > 0 || err != nil {
		return newBranch
	}
	success, err = mergeTags(leaf, newBranch)
	if success > 0 || err != nil {
		return newBranch
	}
	return branch
}

func mergeFields(source *ProcessedResponse, dest *ProcessedResponse) (int, error) {
	var err error
	success := 0
	for key, value := range source.Fields {
		if _, ok := dest.Fields[key]; ok {
			err = fmt.Errorf("a collision occurred on the merge")
		} else {
			dest.Fields[key] = value
			success++
		}
	}
	return success, err
}

func mergeTags(source *ProcessedResponse, dest *ProcessedResponse) (int, error) {
	var err error
	success := 0
	for key, value := range source.Tags {
		if _, ok := dest.Tags[key]; ok {
			err = fmt.Errorf("a collision occurred on the merge")
		} else {
			dest.Tags[key] = value
			success++
		}
	}
	return success, err
}

func getResponseSize(processedResponse *ProcessedResponse) int {
	return len(processedResponse.Fields) + len(processedResponse.Tags)
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func newResponse() *ProcessedResponse {
	return &ProcessedResponse{
		Fields: map[string]interface{}{},
		Tags:   map[string]string{},
	}
}
