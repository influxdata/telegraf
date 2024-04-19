package zabbix

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(zb *ZabbixWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	w := httptest.NewRecorder()
	w.Code = 500

	zb.eventHandler(w, req)

	return w
}

func TestFloatItem(t *testing.T) {
	var acc testutil.Accumulator
	zb := &ZabbixWebhook{Path: "/zabbix", acc: &acc}
	resp := postWebhooks(zb, ItemValueFloatJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"value": 1.0,
	}

	tags := map[string]string{
		"item":          "ICMP: ICMP ping",
		"host_raw":      "10.248.65.148",
		"hostgroups":    "UPS,SNMP",
		"hostname":      "server1",
		"tag_component": "health,network",
		"itemid":        "54988",
	}

	acc.AssertContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)
}

func TestStringItem(t *testing.T) {
	var acc testutil.Accumulator
	zb := &ZabbixWebhook{Path: "/zabbix", acc: &acc}
	resp := postWebhooks(zb, ItemValueTextJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"text": "up",
	}

	tags := map[string]string{
		"item":          "ICMP: ICMP ping",
		"host_raw":      "10.248.65.148",
		"hostname":      "server1",
		"hostgroups":    "UPS,SNMP",
		"tag_component": "health,network",
		"itemid":        "54988",
	}

	acc.AssertContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)
}

func TestMultibleItems(t *testing.T) {
	var acc testutil.Accumulator
	zb := &ZabbixWebhook{Path: "/zabbix", acc: &acc}
	resp := postWebhooks(zb, ItemValueFloatJSON()+"\n"+ItemValueTextJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"value": 1.0,
	}

	tags := map[string]string{
		"item":          "ICMP: ICMP ping",
		"host_raw":      "10.248.65.148",
		"hostgroups":    "UPS,SNMP",
		"hostname":      "server1",
		"tag_component": "health,network",
		"itemid":        "54988",
	}

	acc.AssertContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)

	fields = map[string]interface{}{
		"text": "up",
	}
	acc.AssertContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)
}

func TestIgnoreTextItems(t *testing.T) {
	var acc testutil.Accumulator
	zb := &ZabbixWebhook{Path: "/zabbix", IgnoreText: true, acc: &acc}
	resp := postWebhooks(zb, ItemValueFloatJSON()+"\n"+ItemValueTextJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"value": 1.0,
	}

	tags := map[string]string{
		"item":          "ICMP: ICMP ping",
		"host_raw":      "10.248.65.148",
		"hostgroups":    "UPS,SNMP",
		"hostname":      "server1",
		"tag_component": "health,network",
		"itemid":        "54988",
	}

	acc.AssertContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)

	fields = map[string]interface{}{
		"text": "up",
	}
	acc.AssertDoesNotContainsTaggedFields(t, "zabbix_component_health_network", fields, tags)
}
