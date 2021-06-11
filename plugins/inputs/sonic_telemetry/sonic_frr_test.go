package sonic_telemetry_gnmi

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

const BgpIpV4NbrJson = `
{
  "76.76.76.2":{
    "remoteAs":100,
    "localAs":100,
    "nbrInternalLink":true,
    "hostname":"sonic",
    "bgpVersion":4,
    "remoteRouterId":"2.2.2.2",
    "localRouterId":"1.1.1.1",
    "bgpState":"Established",
    "bgpTimerUp":311534000,
    "bgpTimerUpMsec":311534000,
    "bgpTimerUpString":"3d14h32m",
    "bgpTimerUpEstablishedEpoch":1587666919,
    "bgpTimerLastRead":58000,
    "bgpTimerLastWrite":13000,
    "bgpInUpdateElapsedTimeMsecs":52332000,
    "bgpTimerHoldTimeMsecs":180000,
    "bgpTimerKeepAliveIntervalMsecs":60000,
    "neighborCapabilities":{
      "4byteAs":"advertisedAndReceived",
      "addPath":{
        "ipv4Unicast":{
          "rxAdvertisedAndReceived":true
        },
        "l2VpnEvpn":{
          "rxAdvertisedAndReceived":true
        }
      },
      "routeRefresh":"advertisedAndReceivedOldNew",
      "multiprotocolExtensions":{
        "ipv4Unicast":{
          "advertisedAndReceived":true
        },
        "l2VpnEvpn":{
          "advertisedAndReceived":true
        }
      },
      "hostName":{
        "advHostName":"sonic",
        "advDomainName":"n\/a",
        "rcvHostName":"sonic",
        "rcvDomainName":"n\/a"
      },
      "gracefulRestart":"advertisedAndReceived",
      "gracefulRestartRemoteTimerMsecs":240000,
      "addressFamiliesByPeer":"none"
    },
    "gracefulRestartInfo":{
      "endOfRibSend":{
        "ipv4Unicast":true,
        "l2VpnEvpn":true
      },
      "endOfRibRecv":{
        "ipv4Unicast":true,
        "l2VpnEvpn":true
      }
    },
    "messageStats":{
      "depthInq":0,
      "depthOutq":0,
      "opensSent":3,
      "opensRecv":3,
      "notificationsSent":4,
      "notificationsRecv":0,
      "updatesSent":11,
      "updatesRecv":11,
      "keepalivesSent":5195,
      "keepalivesRecv":5194,
      "routeRefreshSent":0,
      "routeRefreshRecv":0,
      "capabilitySent":0,
      "capabilityRecv":0,
      "totalSent":5213,
      "totalRecv":5208
    },
    "minBtwnAdvertisementRunsTimerMsecs":0,
    "addressFamilyInfo":{
      "ipv4Unicast":{
        "updateGroupId":2,
        "subGroupId":2,
        "packetQueueLength":0,
        "commAttriSentToNbr":"extendedAndStandard",
        "acceptedPrefixCounter":4,
        "sentPrefixCounter":4
      },
      "l2VpnEvpn":{
        "updateGroupId":3,
        "subGroupId":3,
        "packetQueueLength":0,
        "unchangedNextHopPropogatedToNbr":true,
        "commAttriSentToNbr":"extendedAndStandard",
        "advertiseAllVnis":true,
        "acceptedPrefixCounter":0,
        "sentPrefixCounter":0
      }
    },
    "connectionsEstablished":2,
    "connectionsDropped":1,
    "lastResetTimerMsecs":52336000,
    "lastResetDueTo":"No AFI\/SAFI activated for peer",
    "lastResetCode":30,
    "hostLocal":"76.76.76.1",
    "portLocal":179,
    "hostForeign":"76.76.76.2",
    "portForeign":38378,
    "nexthop":"76.76.76.1",
    "nexthopGlobal":"::",
    "nexthopLocal":"::",
    "bgpConnection":"sharedNetwork",
    "connectRetryTimer":120,
    "readThread":"on",
    "writeThread":"on"
  }
}
`
const BgpIpV6NbrJson = `
{
  "76::2":{
    "remoteAs":100,
    "localAs":100,
    "nbrInternalLink":true,
    "hostname":"sonic",
    "bgpVersion":4,
    "remoteRouterId":"2.2.2.2",
    "localRouterId":"1.1.1.1",
    "bgpState":"Established",
    "bgpTimerUp":22000,
    "bgpTimerUpMsec":22000,
    "bgpTimerUpString":"00:00:22",
    "bgpTimerUpEstablishedEpoch":1588043399,
    "bgpTimerLastRead":20000,
    "bgpTimerLastWrite":20000,
    "bgpInUpdateElapsedTimeMsecs":20000,
    "bgpTimerHoldTimeMsecs":180000,
    "bgpTimerKeepAliveIntervalMsecs":60000,
    "neighborCapabilities":{
      "4byteAs":"advertisedAndReceived",
      "addPath":{
        "ipv4Unicast":{
          "rxAdvertisedAndReceived":true
        },
        "ipv6Unicast":{
          "rxAdvertisedAndReceived":true
        }
      },
      "routeRefresh":"advertisedAndReceivedOldNew",
      "multiprotocolExtensions":{
        "ipv4Unicast":{
          "advertisedAndReceived":true
        },
        "ipv6Unicast":{
          "advertisedAndReceived":true
        }
      },
      "hostName":{
        "advHostName":"sonic",
        "advDomainName":"n\/a",
        "rcvHostName":"sonic",
        "rcvDomainName":"n\/a"
      },
      "gracefulRestart":"advertisedAndReceived",
      "gracefulRestartRemoteTimerMsecs":240000,
      "addressFamiliesByPeer":"none"
    },
    "gracefulRestartInfo":{
      "endOfRibSend":{
        "ipv4Unicast":true,
        "ipv6Unicast":true
      },
      "endOfRibRecv":{
        "ipv4Unicast":true,
        "ipv6Unicast":true
      }
    },
    "messageStats":{
      "depthInq":0,
      "depthOutq":0,
      "opensSent":3,
      "opensRecv":3,
      "notificationsSent":4,
      "notificationsRecv":0,
      "updatesSent":11,
      "updatesRecv":11,
      "keepalivesSent":2,
      "keepalivesRecv":2,
      "routeRefreshSent":0,
      "routeRefreshRecv":0,
      "capabilitySent":0,
      "capabilityRecv":0,
      "totalSent":20,
      "totalRecv":16
    },
    "minBtwnAdvertisementRunsTimerMsecs":0,
    "addressFamilyInfo":{
      "ipv4Unicast":{
        "updateGroupId":2,
        "subGroupId":2,
        "packetQueueLength":0,
        "commAttriSentToNbr":"extendedAndStandard",
        "acceptedPrefixCounter":4,
        "sentPrefixCounter":4
      },
      "ipv6Unicast":{
        "updateGroupId":4,
        "subGroupId":6,
        "packetQueueLength":0,
        "commAttriSentToNbr":"extendedAndStandard",
        "acceptedPrefixCounter":0,
        "sentPrefixCounter":0
      }
    },
    "connectionsEstablished":2,
    "connectionsDropped":1,
    "lastResetTimerMsecs":24000,
    "lastResetDueTo":"No AFI\/SAFI activated for peer",
    "lastResetCode":30,
    "hostLocal":"76::1",
    "portLocal":179,
    "hostForeign":"76::2",
    "portForeign":38224,
    "nexthop":"76.76.76.1",
    "nexthopGlobal":"76::1",
    "nexthopLocal":"fe80::4e76:25ff:fee5:d940",
    "bgpConnection":"sharedNetwork",
    "connectRetryTimer":120,
    "readThread":"on",
    "writeThread":"on"
  }
}
`

func TestBGPNbr(t *testing.T) {
	var acc testutil.Accumulator
	sf := &SonicFRR{}
	//acc.SetDebug(true)

	// Define VRF
	vrfs := []struct {
		Name    string
		AfiSafi []string
	}{
		{
			"default",
			[]string{"ipv4", "ipv6"},
		},
	}

	expectedNbrCntTags := map[string]interface{}{
		"ipv4": map[string]string{
			"Vrf":        "default",
			"AddrFamily": "ipv4",
		},
		"ipv6": map[string]string{
			"Vrf":        "default",
			"AddrFamily": "ipv6",
		},
	}

	expectedNbrCntFlds := map[string]interface{}{
		"ipv4": map[string]interface{}{
			"totalNbrs":   int32(1),
			"totalNbrsUp": int32(1),
		},
		"ipv6": map[string]interface{}{
			"totalNbrs":   int32(1),
			"totalNbrsUp": int32(1),
		},
	}

	expectedNbrsTags := map[string]interface{}{
		"ipv4": map[string]string{
			"AddrFamily":  "ipv4",
			"BGPNeighbor": "76.76.76.2",
			"Vrf":         "default",
		},
		"ipv6": map[string]string{
			"AddrFamily":  "ipv6",
			"BGPNeighbor": "76::2",
			"Vrf":         "default",
		},
	}

	expectedNbrsFlds := map[string]interface{}{
		"ipv4": map[string]interface{}{
			"ipv4PrefixRecv": float64(4),
			"ipv4PrefixSent": float64(4),
			"localAs":        float64(100),
			"localRouterID":  "1.1.1.1",
			"remoteAs":       float64(100),
			"remoteRouterID": "2.2.2.2",
			"state":          "Established",
			"totalMsgsRecv":  float64(5208),
			"totalMsgsSent":  float64(5213),
			"uptime":         "3d14h32m",
		},

		"ipv6": map[string]interface{}{
			"ipv4PrefixRecv": float64(4),
			"ipv4PrefixSent": float64(4),
			"ipv6PrefixRecv": float64(0),
			"ipv6PrefixSent": float64(0),
			"localAs":        float64(100),
			"localRouterID":  "1.1.1.1",
			"remoteAs":       float64(100),
			"remoteRouterID": "2.2.2.2",
			"state":          "Established",
			"totalMsgsRecv":  float64(16),
			"totalMsgsSent":  float64(20),
			"uptime":         "00:00:22",
		},
	}

	for _, vrf := range vrfs {
		vrf_name := vrf.Name
		for _, afisafi := range vrf.AfiSafi {
			bgp_nbr_json := make(map[string]interface{})

			var err error
			if afisafi == "ipv4" {
				err = json.Unmarshal([]byte(BgpIpV4NbrJson), &bgp_nbr_json)
			} else if afisafi == "ipv6" {
				err = json.Unmarshal([]byte(BgpIpV6NbrJson), &bgp_nbr_json)
			} else {
				acc.AddError(fmt.Errorf("Unsupported Address Family '%s' Json", afisafi))
				return
			}
			if err != nil {
				acc.AddError(fmt.Errorf("Unmarshal of '%s' Json", afisafi))
				return
			}
			var nbr_cnt, nbr_up_cnt int32
			for nbr_key := range bgp_nbr_json {
				if nbr_data_json, ok := bgp_nbr_json[nbr_key].(map[string]interface{}); ok {
					err, nbr_state := sf.gatherNbrInfo(vrf_name, nbr_key, afisafi, nbr_data_json, &acc)
					if err != nil {
						acc.AddError(fmt.Errorf("Unable to Gather NbrInfo for vrf:'%s' nbr:'%s' afisafi:'%s' err: '%s'",
							vrf_name, nbr_key, afisafi, err))
					}
					nbr_cnt++
					if nbr_state == true {
						nbr_up_cnt++
					}
				}
			}
			fields := map[string]interface{}{
				"totalNbrs":   nbr_cnt,
				"totalNbrsUp": nbr_up_cnt,
			}

			tags := map[string]string{
				"Vrf":        vrf_name,
				"AddrFamily": afisafi,
			}
			acc.AddFields("BGPNbrCount", fields, tags)

			acc.AssertContainsTaggedFields(t, "BGPNeighbors",
				expectedNbrsFlds[afisafi].(map[string]interface{}),
				expectedNbrsTags[afisafi].(map[string]string))
			acc.AssertContainsTaggedFields(t, "BGPNbrCount",
				expectedNbrCntFlds[afisafi].(map[string]interface{}),
				expectedNbrCntTags[afisafi].(map[string]string))
		}
	}
}
