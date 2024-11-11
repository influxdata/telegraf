package mikrotik

import (
	"encoding/json"
	"fmt"
	"strings"
)

var ignoreCommentsFunction func(commonData) bool

func createPropLists() (metricsPropList string, resourcesPropList string, routerboardPropList string) {
	metricsPropList = ".proplist=" + strings.Join(append(tagFields, valueFields...), ",")
	resourcesPropList = ".proplist=" + strings.Join(systemResources, ",")
	routerboardPropList = ".proplist=" + strings.Join(systemRouterBoard, ",")
	return metricsPropList, resourcesPropList, routerboardPropList
}

func binToCommon(b []byte) (c common, err error) {
	errCommon := json.Unmarshal(b, &c)
	if errCommon == nil {
		return c, nil
	}

	cd := commonData{}
	err = json.Unmarshal(b, &cd)
	if err != nil {
		return c, fmt.Errorf("data could not be unmarshalled neither into common nor into commonData structures. %w %w", errCommon, err)
	}
	c = append(c, cd)

	return c, err
}

func basicCommentAndDisableFilter(commentsToIgnore []string) func(commonData) bool {
	return func(c commonData) bool {
		if c["disabled"] == "true" {
			return false
		}
		for _, comment := range commentsToIgnore {
			if strings.Contains(c["comment"], comment) {
				return false
			}
		}

		return true
	}
}
