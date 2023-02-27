package grafana_dashboard

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
)

type AlexanderzobninZabbixHostGroupResponseItem struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name"`
}

type AlexanderzobninZabbixHostGroupResponse struct {
	Result []*AlexanderzobninZabbixHostGroupResponseItem `json:"result,omitempty"`
}

type AlexanderzobninZabbixHostGroupRequestParams struct {
	Output    []string `json:"output"`
	RealHosts bool     `json:"real_hosts"`
	SortField string   `json:"sortfield"`
}

type AlexanderzobninZabbixHostGroupRequest struct {
	DatasourceID uint                                        `json:"datasourceId"`
	Method       string                                      `json:"method"`
	Params       AlexanderzobninZabbixHostGroupRequestParams `json:"params"`
}

type AlexanderzobninZabbixHostResponseItem struct {
	Host   string                                        `json:"host"`
	HostID string                                        `json:"hostid"`
	Name   string                                        `json:"name"`
	Groups []*AlexanderzobninZabbixHostGroupResponseItem `json:"groups"`
}

type AlexanderzobninZabbixHostResponse struct {
	Result []*AlexanderzobninZabbixHostResponseItem `json:"result,omitempty"`
}

type AlexanderzobninZabbixHostRequestParams struct {
	GroupIDs     []string `json:"groupids"`
	Output       []string `json:"output"`
	SortField    string   `json:"sortfield"`
	SelectGroups []string `json:"selectGroups"`
}

type AlexanderzobninZabbixHostRequest struct {
	DatasourceID uint                                   `json:"datasourceId"`
	Method       string                                 `json:"method"`
	Params       AlexanderzobninZabbixHostRequestParams `json:"params"`
}

type AlexanderzobninZabbixItemResponseItem struct {
	ItemID string `json:"itemid"`
	HostID string `json:"hostid"`
	Name   string `json:"name"`
	Units  string `json:"units"`
}

type AlexanderzobninZabbixItemResponse struct {
	Result []*AlexanderzobninZabbixItemResponseItem `json:"result,omitempty"`
}

type AlexanderzobninZabbixItemRequestParams struct {
	HostIDs     []string               `json:"hostids"`
	Filter      map[string]interface{} `json:"filter"`
	Output      []string               `json:"output"`
	SelectHosts []string               `json:"selectHosts"`
	SortField   string                 `json:"sortfield"`
	WebItems    bool                   `json:"webitems"`
}

type AlexanderzobninZabbixItemRequest struct {
	DatasourceID uint                                   `json:"datasourceId"`
	Method       string                                 `json:"method"`
	Params       AlexanderzobninZabbixItemRequestParams `json:"params"`
}

type AlexanderzobninZabbixHistoryRequestParams struct {
	ItemIDs   []string `json:"itemids"`
	History   string   `json:"history"`
	Output    string   `json:"output"`
	SortField string   `json:"sortfield"`
	SortOrder string   `json:"sortorder"`
	TimeFrom  int      `json:"time_from"`
	TimeTill  int      `json:"time_till"`
}

type AlexanderzobninZabbixHistoryRequest struct {
	DatasourceID uint                                      `json:"datasourceId"`
	Method       string                                    `json:"method"`
	Params       AlexanderzobninZabbixHistoryRequestParams `json:"params"`
}

type AlexanderzobninZabbixHistoryResponseItem struct {
	Clock  string `json:"clock"`
	ItemID string `json:"itemid"`
	NS     string `json:"ns"`
	Value  string `json:"value"`
}

type AlexanderzobninZabbixHistoryResponse struct {
	Result []*AlexanderzobninZabbixHistoryResponseItem `json:"result,omitempty"`
}

type AlexanderzobninZabbix struct {
	log     telegraf.Logger
	grafana *Grafana
}

func (az *AlexanderzobninZabbix) getFilter(s interface{}) string {
	filter := ""
	if s != nil {
		mm, ok := s.(map[string]interface{})
		if ok {
			f, ok := mm["filter"].(string)
			if ok && f != "" {
				filter = strings.ReplaceAll(f, "/", "")
			}
		}
	}
	return filter
}

func (az *AlexanderzobninZabbix) getHostGroupIDs(dsID uint, group interface{}) ([]string, []*AlexanderzobninZabbixHostGroupResponseItem, error) {
	filter := az.getFilter(group)

	request := AlexanderzobninZabbixHostGroupRequest{
		DatasourceID: dsID,
		Method:       "hostgroup.get",
		Params: AlexanderzobninZabbixHostGroupRequestParams{
			Output:    []string{"name"},
			RealHosts: true,
			SortField: "name",
		},
	}

	b, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("/api/datasources/%d/resources/zabbix-api", dsID)
	raw, code, err := az.grafana.httpPost(url, nil, b)
	if err != nil {
		return nil, nil, err
	}
	if code != 200 {
		return nil, nil, fmt.Errorf("AlexanderzobninZabbix HTTP error %d: returns %s", code, raw)
	}

	var res AlexanderzobninZabbixHostGroupResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, nil, err
	}

	var hostGroupIDs []string
	var hostGroups []*AlexanderzobninZabbixHostGroupResponseItem
	for _, v := range res.Result {
		if v != nil {
			if m, _ := regexp.MatchString(filter, v.Name); m {
				hostGroupIDs = append(hostGroupIDs, v.GroupID)
				hostGroups = append(hostGroups, v)
			}
		}
	}
	return hostGroupIDs, hostGroups, nil
}

func (az *AlexanderzobninZabbix) getHostIDs(dsID uint, hostGroupIDs []string, host interface{}) ([]string, []*AlexanderzobninZabbixHostResponseItem, error) {
	filter := az.getFilter(host)

	request := AlexanderzobninZabbixHostRequest{
		DatasourceID: dsID,
		Method:       "host.get",
		Params: AlexanderzobninZabbixHostRequestParams{
			GroupIDs: hostGroupIDs,
			Output:   []string{"name", "host"},
			//SelectGroups: []string{"name", "groupid"},
			SortField: "name",
		},
	}

	b, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("/api/datasources/%d/resources/zabbix-api", dsID)
	raw, code, err := az.grafana.httpPost(url, nil, b)
	if err != nil {
		return nil, nil, err
	}
	if code != 200 {
		return nil, nil, fmt.Errorf("AlexanderzobninZabbix HTTP error %d: returns %s", code, raw)
	}

	var res AlexanderzobninZabbixHostResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, nil, err
	}

	var hostIDs []string
	var hosts []*AlexanderzobninZabbixHostResponseItem
	for _, v := range res.Result {
		if v != nil {
			if m, _ := regexp.MatchString(filter, v.Name); m {
				hostIDs = append(hostIDs, v.HostID)
				hosts = append(hosts, v)
			}
		}
	}
	return hostIDs, hosts, nil
}

func (az *AlexanderzobninZabbix) getItemIDs(dsID uint, hostIDs []string, item interface{}) ([]string, []*AlexanderzobninZabbixItemResponseItem, error) {
	filter := az.getFilter(item)

	filterMap := make(map[string]interface{})
	filterMap["value_type"] = []int{0, 3}

	request := AlexanderzobninZabbixItemRequest{
		DatasourceID: dsID,
		Method:       "item.get",
		Params: AlexanderzobninZabbixItemRequestParams{
			HostIDs:   hostIDs,
			Filter:    filterMap,
			Output:    []string{"name", "hostid", "units"},
			SortField: "name",
			WebItems:  true,
		},
	}

	b, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("/api/datasources/%d/resources/zabbix-api", dsID)
	raw, code, err := az.grafana.httpPost(url, nil, b)
	if err != nil {
		return nil, nil, err
	}
	if code != 200 {
		return nil, nil, fmt.Errorf("AlexanderzobninZabbix HTTP error %d: returns %s", code, raw)
	}

	var res AlexanderzobninZabbixItemResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, nil, err
	}

	var itemIDs []string
	var items []*AlexanderzobninZabbixItemResponseItem
	for _, v := range res.Result {
		if v != nil {
			// compare through regex AND directly
			if m, _ := regexp.MatchString(filter, v.Name); m || filter == v.Name {
				itemIDs = append(itemIDs, v.ItemID)
				items = append(items, v)
			}
		}
	}
	return itemIDs, items, nil
}

func (az *AlexanderzobninZabbix) getTags(
	itemID string,
	items []*AlexanderzobninZabbixItemResponseItem,
	hosts []*AlexanderzobninZabbixHostResponseItem,
) (string, string, string) {
	var (
		item *AlexanderzobninZabbixItemResponseItem
		host *AlexanderzobninZabbixHostResponseItem
	)

	for _, i := range items {
		if i.ItemID == itemID {
			item = i
			break
		}
	}
	if item == nil {
		return "", "", ""
	}

	for _, h := range hosts {
		if h.HostID == item.HostID {
			host = h
			break
		}
	}
	if host == nil {
		return item.Name, item.Units, ""
	}
	return item.Name, item.Units, host.Name
}

func (az *AlexanderzobninZabbix) getHistory(dsID uint, itemIDs []string, history string, start int, end int) (*AlexanderzobninZabbixHistoryResponse, error) {
	request := AlexanderzobninZabbixHistoryRequest{
		DatasourceID: dsID,
		Method:       "history.get",
		Params: AlexanderzobninZabbixHistoryRequestParams{
			ItemIDs:   itemIDs,
			History:   history,
			Output:    "extend",
			SortField: "clock",
			SortOrder: "ASC",
			TimeFrom:  start,
			TimeTill:  end,
		},
	}

	b, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	az.log.Debugf("AlexanderzobninZabbix request body => %s", string(b))

	url := fmt.Sprintf("/api/datasources/%d/resources/zabbix-api", dsID)
	raw, code, err := az.grafana.httpPost(url, nil, b)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("AlexanderzobninZabbix HTTP error %d: returns %s", code, raw)
	}
	var res AlexanderzobninZabbixHistoryResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (az *AlexanderzobninZabbix) GetData(t *sdk.Target, ds *sdk.Datasource, period *GrafanaDashboardPeriod, push GrafanaDatasourcePushFunc) error {
	when := time.Now()

	hostGroupIDs, _, err := az.getHostGroupIDs(ds.ID, t.Group)
	if err != nil {
		return err
	}
	if len(hostGroupIDs) == 0 {
		az.log.Debug("AlexanderzobninZabbix has no host group IDs")
	}

	hostIDs, hosts, err := az.getHostIDs(ds.ID, hostGroupIDs, t.Host)
	if err != nil {
		return err
	}
	if len(hostIDs) == 0 {
		az.log.Debug("AlexanderzobninZabbix has no host IDs")
		return nil
	}

	itemIDs, items, err := az.getItemIDs(ds.ID, hostIDs, t.Item)
	if err != nil {
		return err
	}
	if len(itemIDs) == 0 {
		az.log.Debug("AlexanderzobninZabbix has no item IDs")
		return nil
	}

	t1, t2 := period.StartEnd()
	start := int(t1.UTC().Unix())
	end := int(t2.UTC().Unix())

	res, err := az.getHistory(ds.ID, itemIDs, "0", start, end)
	if err != nil {
		return nil
	}

	if len(res.Result) == 0 {
		az.log.Debug("AlexanderzobninZabbix has no float data. Trying int…")
		res, err = az.getHistory(ds.ID, itemIDs, "3", start, end)
		if err != nil {
			return nil
		}
	}

	if len(res.Result) == 0 {
		az.log.Debug("AlexanderzobninZabbix has no data.")
		return nil
	}

	for _, r := range res.Result {
		tags := make(map[string]string)

		item, units, host := az.getTags(r.ItemID, items, hosts)

		if item != "" {
			tags["item"] = item
		}
		if units != "" {
			tags["units"] = units
		}
		if host != "" {
			tags["host"] = host
		}

		ts, err := strconv.ParseInt(r.Clock, 0, 64)
		if err != nil {
			continue
		}

		if f, err := strconv.ParseFloat(r.Value, 64); err == nil {
			if len(t.Functions) > 0 {
				f = applyZabbixFunctions(f, t.Functions)
			}
			push(when, tags, time.Unix(ts, 0), f)
		}
	}
	return nil
}

func applyZabbixFunctions(v float64, functions []sdk.ZabbixFunction) float64 {
	res := v
	for _, f := range functions {
		switch f.Def.Name {
		case "scale":
			factor, err := strconv.ParseFloat(f.Params[0], 64)
			if err == nil {
				res = res * factor
			}
		case "offset":
			factor, err := strconv.ParseFloat(f.Params[0], 64)
			if err == nil {
				res = res + factor
			}
		}
	}
	return res
}

func NewAlexanderzobninZabbix(log telegraf.Logger, grafana *Grafana) *AlexanderzobninZabbix {
	return &AlexanderzobninZabbix{
		log:     log,
		grafana: grafana,
	}
}
