package zabbix

import (
	"strings"
)

func ItemValueFloatJSON() string {
	return strings.Replace(`
	{
		"host": {
			"host": "10.248.65.148",
			"name": "server1"
		},
		"groups": [
			"UPS",
			"SNMP"
		],
		"item_tags": [
			{
				"tag": "component",
				"value": "health"
			},
			{
				"tag": "component",
				"value": "network"
			}
		],
		"itemid": 54988,
		"name": "ICMP: ICMP ping",
		"clock": 1712352621,
		"ns": 304061973,
		"value": 1,
		"type": 3
	}`, "\n", " ", -1)
}

func ItemValueTextJSON() string {
	return strings.Replace(`
	{
		"host": {
		"host": "10.248.65.148",
		"name": "server1"
		},
		"groups": [
			"UPS",
			"SNMP"
		],
		"item_tags": [
		{
		"tag": "component",
		"value": "health"
		},
		{
		"tag": "component",
		"value": "network"
		}
		],
		"itemid": 54988,
		"name": "ICMP: ICMP ping",
		"clock": 1712352621,
		"ns": 304061973,
		"value": "up",
		"type": 4
		}`, "\n", " ", -1)
}
