package jti_openconfig_telemetry

import "sort"

type dataGroup struct {
	numKeys int
	tags    map[string]string
	data    map[string]interface{}
}

// Sort the data groups by number of keys
type collectionByKeys []dataGroup

func (a collectionByKeys) Len() int           { return len(a) }
func (a collectionByKeys) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a collectionByKeys) Less(i, j int) bool { return a[i].numKeys < a[j].numKeys }

// Checks to see if there is already a group with these tags and returns its index. Returns -1 if unavailable.
func (a collectionByKeys) isAvailable(tags map[string]string) *dataGroup {
	sort.Sort(a)

	// Iterate through all the groups and see if we have group with these tags
	for _, group := range a {
		// Since already sorted, match with only groups with N keys
		if group.numKeys < len(tags) {
			continue
		} else if group.numKeys > len(tags) {
			break
		}

		matchFound := true
		for k, v := range tags {
			val, ok := group.tags[k]
			if !ok || val != v {
				matchFound = false
				break
			}
		}

		if matchFound {
			return &group
		}
	}
	return nil
}

// Inserts into already existing group or creates a new group
func (a collectionByKeys) insert(tags map[string]string, data map[string]interface{}) collectionByKeys {
	// If there is already a group with this set of tags, insert into it. Otherwise create a new group and insert
	if group := a.isAvailable(tags); group != nil {
		for k, v := range data {
			group.data[k] = v
		}
	} else {
		a = append(a, dataGroup{len(tags), tags, data})
	}

	return a
}
