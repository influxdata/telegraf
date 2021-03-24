package jti_openconfig_telemetry

import "sort"

type DataGroup struct {
	numKeys int
	tags    map[string]string
	data    map[string]interface{}
}

// Sort the data groups by number of keys
type CollectionByKeys []DataGroup

func (a CollectionByKeys) Len() int           { return len(a) }
func (a CollectionByKeys) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CollectionByKeys) Less(i, j int) bool { return a[i].numKeys < a[j].numKeys }

// Checks to see if there is already a group with these tags and returns its index. Returns -1 if unavailable.
func (a CollectionByKeys) IsAvailable(tags map[string]string) *DataGroup {
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
			if val, ok := group.tags[k]; ok {
				if val != v {
					matchFound = false
					break
				}
			} else {
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
func (a CollectionByKeys) Insert(tags map[string]string, data map[string]interface{}) CollectionByKeys {
	// If there is already a group with this set of tags, insert into it. Otherwise create a new group and insert
	if group := a.IsAvailable(tags); group != nil {
		for k, v := range data {
			group.data[k] = v
		}
	} else {
		a = append(a, DataGroup{len(tags), tags, data})
	}

	return a
}
