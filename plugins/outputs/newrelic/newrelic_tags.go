package newrelic

import(
  "sort"
  "strings"
)

type NewRelicTags struct {
	Tags *map[string]string
	SortedKeys []string
	Hostname string
}

func TagValue(tagValue string) string {
	tagValueParts := strings.Split(tagValue, "/")
	var clean_parts []string
	for _, part := range(tagValueParts) {
		if part != "" { clean_parts = append(clean_parts, part) }
	}
	if len(clean_parts) > 0 {
		tagValue = strings.ToLower(strings.Join(clean_parts, "-"))
	} else {
		tagValue = "root"
	}
	return tagValue
}

func (nrt *NewRelicTags) Fill(originalTags map[string]string) {
	nrt.SortedKeys = make([]string, 0, len(originalTags))
  tags := make(map[string]string)
  nrt.Tags = &tags
  for key, value := range originalTags {
    if key != "host" {
			nrt.SortedKeys = append(nrt.SortedKeys, key)
			(*nrt.Tags)[key] = TagValue(value)
		} else {
			nrt.Hostname = value
		}
  }
  sort.Strings(nrt.SortedKeys)
}

func (nrt *NewRelicTags) GetTag(tagKey string) string {
  return (*nrt.Tags)[tagKey]
}
