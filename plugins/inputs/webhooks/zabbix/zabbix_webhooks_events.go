package zabbix

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

type host struct {
	Host string `json:"host"`
	Name string `json:"name"`
}

type tag struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type zabbix_item struct {
	Value    any      `json:"value"`
	Itemname string   `json:"name"`
	Groups   []string `json:"groups"`
	Type     int      `json:"type"`
	Host     host     `json:"host"`
	ItemTags []tag    `json:"item_tags"`
	Itemid   int      `json:"itemid"`
	Clock    int64    `json:clock`
	Ns       int64    `json:ns`
}

func (ni *zabbix_item) NameFromTag(t string) string {
	var tags []string
	for _, s := range ni.ItemTags {
		if strings.ToLower(s.Tag) == t {
			tags = append(tags, s.Value)
		}
	}
	sort.Strings(tags)
	return t + "_" + strings.Join(tags[:], "_")
}

func (ni *zabbix_item) Tags() map[string]string {
	sort.Strings(ni.Groups)
	res := map[string]string{
		"item":       ni.Itemname,
		"host_raw":   ni.Host.Host,
		"hostname":   ni.Host.Name,
		"hostgroups": strings.Join(ni.Groups[:], ","),
		"itemid":     strconv.Itoa(ni.Itemid),
	}
	tag_map := map[string][]string{}
	for _, s := range ni.ItemTags {
		var tag_name = "tag_" + s.Tag
		if _, ok := tag_map[tag_name]; ok {
			tag_map[tag_name] = append(tag_map[tag_name], s.Value)
		} else {
			tag_map[tag_name] = []string{s.Value}
		}
	}
	for k, v := range tag_map {
		sort.Strings(v)
		res[k] = strings.Join(v[:], ",")
	}

	return res
}

func (ni *zabbix_item) Fields() map[string]interface{} {
	if ni.Type == 1 || ni.Type == 2 || ni.Type == 4 {
		return map[string]interface{}{
			"text": ni.Value,
		}
	}
	return map[string]interface{}{
		"value": ni.Value,
	}
}

func (ni *zabbix_item) Time() time.Time {
	return time.Unix(ni.Clock, ni.Ns)
}
