package ah_trap

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"unsafe"
	"encoding/binary"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/common/ahutil"
)

const sampleConfig = `
[[inputs.ah_trap]]
  interval = "10s"
`
type TrapPlugin struct {
	acc		   telegraf.Accumulator
	wg                  sync.WaitGroup
}

const EVT_SOCK = "/tmp/ah_telegraf.sock"

func (t *TrapPlugin) SampleConfig() string {
	return  sampleConfig
}

func (t *TrapPlugin) Description() string {
	return "Trap listener plugin over Unix socket"
}

func (t *TrapPlugin) Init() error {
	return nil
}



func (t *TrapPlugin) Start(acc telegraf.Accumulator) error {
	_ = os.RemoveAll(EVT_SOCK)

	conn, err := net.ListenPacket("unixgram", EVT_SOCK)
	if err != nil {
		return fmt.Errorf("socket listen error: %v", err)
	}

	_ = os.Chmod(EVT_SOCK, 0666)

	t.acc = acc
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.trapListener(conn)
	}()

	return nil
}

func (t *TrapPlugin) Stop() {
	t.wg.Wait()
}

/*
Gathers the generic trap that were part of AH_TRAP_MSG_TYPE type
*/
func (t *TrapPlugin) Gather_Ah_Logen(trap AhTrapMsg, acc telegraf.Accumulator) error {

	switch AhTrapType(trap.TrapType) {
	case AH_FAILURE_TRAP_TYPE:
		var failure AhFailureTrap
		rawSize := int(unsafe.Sizeof(failure))
		copy((*[1 << 10]byte)(unsafe.Pointer(&failure))[:rawSize], trap.Union[:rawSize])

		acc.AddFields("TrapEvent", map[string]interface{}{
			"trapObjName_failureTrap":          ahutil.CleanCString(failure.Name[:]),
			"cause_failureTrap":             failure.Cause,
			"set_failureTrap":               failure.Set,
			"severityLevel_trapMessage_failureTrap": trap.Level,
			"msgId_trapMessage_failureTrap": trap.MsgID,
			"desc_trapMessage_failureTrap":  ahutil.CleanCString(trap.Desc[:]),
		}, nil)

	case AH_THRESHOLD_TRAP_TYPE:
		var ahthreshold AhThresholdTrap
		rawSize := int(unsafe.Sizeof(ahthreshold))
		copy((*[1 << 10]byte)(unsafe.Pointer(&ahthreshold))[:rawSize], trap.Union[:rawSize])

		acc.AddFields("TrapEvent", map[string]interface{}{
			"trapObjName_thresholdTrap":          ahutil.CleanCString(ahthreshold.Name[:]),
			"curVal_thresholdTrap":            ahthreshold.CurVal,
			"thresholdHigh_thresholdTrap":     ahthreshold.ThresholdHigh,
			"thresholdLow_thresholdTrap":      ahthreshold.ThresholdLow,
			"severityLevel_trapMessage_thresholdTrap": trap.Level,
			"msgId_trapMessage_thresholdTrap": trap.MsgID,
			"desc_trapMessage_thresholdTrap":  ahutil.CleanCString(trap.Desc[:]),
		}, nil)

        case AH_STATE_CHANGE_TRAP_TYPE:
                var ahstatechange AhStateChangeTrap
                rawSize := int(unsafe.Sizeof(ahstatechange))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahstatechange))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_stateChangeTrap":          ahutil.CleanCString(ahstatechange.Name[:]),
                        "preState_stateChangeTrap":          ahstatechange.PreState,
                        "curState_stateChangeTrap":          ahstatechange.CurState,
                        "opMode_stateChangeTrap":            ahstatechange.OperationMode,
                        "severityLevel_trapMessage_stateChangeTrap": trap.Level,
                        "msgId_trapMessage_stateChangeTrap": trap.MsgID,
                        "desc_trapMessage_stateChangeTrap":  ahutil.CleanCString(trap.Desc[:]),
                }, nil)

	 case AH_CONNECTION_CHANGE_TRAP_TYPE:
                var ahconnectionchange AhConnectionChangeTrap
                rawSize := int(unsafe.Sizeof(ahconnectionchange))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahconnectionchange))[:rawSize], trap.Union[:rawSize])

		gatherConnectionChangeTrap(ahconnectionchange, trap, acc)

	case AH_IDP_AP_EVENT_TRAP_TYPE:
                var ahidpapevent AhIdpApEventTrap
                rawSize := int(unsafe.Sizeof(ahidpapevent))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahidpapevent))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_idpApEventTrap":          ahutil.CleanCString(ahidpapevent.Name[:]),
                        "ifIndex_idpApEventTrap":           ahidpapevent.IfIndex,
                        "remoteId_idpApEventTrap":          ahutil.FormatMac(ahidpapevent.RemoteID),
                        "idpType_idpApEventTrap":           ahidpapevent.IdpType,
			"idpChannel_idpApEventTrap":        ahidpapevent.IdpChannel,
			"idpRssi_idpApEventTrap":           ahidpapevent.IdpRSSI,
			"idpCompliance_idpApEventTrap":     ahidpapevent.IdpCompliance,
			"ssid_idpApEventTrap":              ahutil.CleanCString(ahidpapevent.SSID[:]),
			"stationType_idpApEventTrap":       ahidpapevent.StationType,
			"stationData_idpApEventTrap":       ahidpapevent.StationData,
			"idpRemoved_idpApEventTrap":        ahidpapevent.IdpRemoved,
			"idpInnet_idpApEventTrap":          ahidpapevent.IdpInNet,
                        "severityLevel_trapMessage_idpApEventTrap": trap.Level,
                        "msgId_trapMessage_idpApEventTrap": trap.MsgID,
                        "desc_trapMessage_idpApEventTrap":  ahutil.CleanCString(trap.Desc[:]),
                }, nil)

	case AH_CLIENT_INFO_TRAP_TYPE:
                var ahclientinfo AhClientInfoTrap
                rawSize := int(unsafe.Sizeof(ahclientinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahclientinfo))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_clientInfoTrap":          ahutil.CleanCString(ahclientinfo.Name[:]),
                        "ssid_clientInfoTrap":              ahutil.CleanCString(ahclientinfo.Ssid[:]),
			"clientMac_clientInfoTrap":         ahutil.FormatMac(ahclientinfo.ClientMac),
                        "hostName_clientInfoTrap":          ahutil.CleanCString(ahclientinfo.HostName[:]),
			"userName_clientInfoTrap":          ahutil.CleanCString(ahclientinfo.UserName[:]),
			"clientIp_clientInfoTrap":          ahutil.IntToIpv4(ahclientinfo.ClientIP),
			"mgtStatus_clientInfoTrap":         ahclientinfo.MgtStus,
			"staAddr6Num_clientInfoTrap":       ahclientinfo.StaAddr6Num,
			"staAddr6_clientInfoTrap":          intToIPv6(ahclientinfo.StaAddr6[:], int(ahclientinfo.StaAddr6Num)),
                        "severityLevel_trapMessage_clientInfoTrap": trap.Level,
                        "msgId_trapMessage_clientInfoTrap": trap.MsgID,
                        "desc_trapMessage_clientInfoTrap":  ahutil.CleanCString(trap.Desc[:]),
                }, nil)

         case AH_POWER_INFO_TRAP_TYPE:
                var ahpowerinfo AhPowerInfoTrap
                rawSize := int(unsafe.Sizeof(ahpowerinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahpowerinfo))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_powerInfoTrap":           ahutil.CleanCString(ahpowerinfo.Name[:]),
                        "powerSrc_powerInfoTrap":           ahpowerinfo.PowerSrc,
                        "eth0On_powerInfoTrap":             ahpowerinfo.Eth0On,
                        "eth1On_powerInfoTrap":             ahpowerinfo.Eth1On,
                        "eth0Pwr_powerInfoTrap":            ahpowerinfo.Eth0Pwr,
                        "eth1Pwr_powerInfoTrap":            ahpowerinfo.Eth1Pwr,
                        "eth0Speed_powerInfoTrap":          ahpowerinfo.Eth0Speed,
                        "eth1Speed_powerInfoTrap":          ahpowerinfo.Eth1Speed,
                        "wifi0Setting_powerInfoTrap":       ahpowerinfo.Wifi0Setting,
                        "wifi1Setting_powerInfoTrap":       ahpowerinfo.Wifi1Setting,
                        "wifi2Setting_powerInfoTrap":       ahpowerinfo.Wifi2Setting,
			"severityLevel_trapMessage_powerInfoTrap":  trap.Level,
                        "msgId_trapMessage_powerInfoTrap":  trap.MsgID,
			"desc_trapMessage_powerInfoTrap":   ahutil.CleanCString(trap.Desc[:]),
			}, nil)

	case AH_CHANNEL_POWER_TRAP_TYPE:
                var ahchannelpowerinfo AhChannelPowerTrap
                rawSize := int(unsafe.Sizeof(ahchannelpowerinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahchannelpowerinfo))[:rawSize], trap.Union[:rawSize])


                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_channelPowerTrap":          ahutil.CleanCString(ahchannelpowerinfo.Name[:]),
                        "ifIndex_channelPowerTrap":           ahchannelpowerinfo.IfIndex,
                        "channel_channelPowerTrap":           ahchannelpowerinfo.RadioChannel,
                        "txPwr_channelPowerTrap":             ahchannelpowerinfo.RadioTxPower,
                        "beaconInterval_channelPowerTrap":    ahchannelpowerinfo.BeaconInterval,
                        "channelStrfmt_channelPowerTrap":     ahchannelpowerinfo.ChnlStrfmt,
                        "powerStrfmt_channelPowerTrap":       ahchannelpowerinfo.PwrStrfmt,
			"radioEirp_channelPowerTrap":         ahutil.CleanCString(ahchannelpowerinfo.RadioEirp[:]),
                        "reason_channelPowerTrap":            ahchannelpowerinfo.Reason,
                        "severityLevel_trapMessage_channelPowerTrap":  trap.Level,
                        "msgId_trapMessage_channelPowerTrap":  trap.MsgID,
                        "desc_trapMessage_channelPowerTrap":   ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_IDP_MITIGATE_TRAP_TYPE:
                var ahidpmitigate AhIdpMitigateTrap
                rawSize := int(unsafe.Sizeof(ahidpmitigate))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahidpmitigate))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_idpMitigateTrap":           ahutil.CleanCString(ahidpmitigate.Name[:]),
                        "ifIndex_idpMitigateTrap":            ahidpmitigate.IfIndex,
			"remoteId_idpMitigateTrap":           ahutil.FormatMac(ahidpmitigate.RemoteID),
                        "bssid_idpMitigateTrap":              ahutil.FormatMac(ahidpmitigate.BSSID),
                        "removed_idpMitigateTrap":            ahidpmitigate.Removed,
                        "discoverAge_idpMitigateTrap":        ahidpmitigate.DiscoverAge,
                        "updateAge_idpMitigateTrap":          ahidpmitigate.UpdateAge,
                        "severityLevel_trapMessage_idpMitigateTrap":  trap.Level,
                        "msgId_trapMessage_idpMitigateTrap":  trap.MsgID,
                        "desc_trapMessage_idpMitigateTrap":   ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_INTERFERENCE_ALERT_TRAP_TYPE:
                var ahinterference AhInterferenceAlertTrap
                rawSize := int(unsafe.Sizeof(ahinterference))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahinterference))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_interferenceAlertTrap":            ahutil.CleanCString(ahinterference.Name[:]),
                        "ifIndex_interferenceAlertTrap":             ahinterference.IfIndex,
                        "interferenceThres_interferenceAlertTrap":   ahinterference.InterferenceThres,
                        "aveInterference_interferenceAlertTrap":     ahinterference.AveInterference,
                        "shortInterference_interferenceAlertTrap":   ahinterference.ShortInterference,
                        "snapInterference_interferenceAlertTrap":    ahinterference.SnapInterference,
                        "crcErrRateThreshold_interferenceAlertTrap": ahinterference.CRCErrRateThres,
                        "crcErrRate_interferenceAlertTrap":          ahinterference.CRCErrRate,
                        "failureSet_interferenceAlertTrap":                 ahinterference.Set,
                        "severityLevel_trapMessage_interferenceAlertTrap":   trap.Level,
                        "msgId_trapMessage_interferenceAlertTrap":   trap.MsgID,
                        "desc_trapMessage_interferenceAlertTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_BW_SENTINEL_TRAP_TYPE:
                var ahbwsentinel AhBwSentinelTrap
                rawSize := int(unsafe.Sizeof(ahbwsentinel))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahbwsentinel))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_bwSentinelTrap":            ahutil.CleanCString(ahbwsentinel.Name[:]),
                        "ifIndex_bwSentinelTrap":             ahbwsentinel.IfIndex,
                        "clientMac_bwSentinelTrap":           ahutil.FormatMac(ahbwsentinel.ClientMac),
                        "bwSentinelStatus_bwSentinelTrap":    ahbwsentinel.BwSentinelStatus,
                        "gbw_bwSentinelTrap":		      ahbwsentinel.GBW,
                        "actualBw_bwSentinelTrap":            ahbwsentinel.ActualBW,
                        "actionTaken_bwSentinelTrap":         ahbwsentinel.ActionTaken,
                        "channelUtil_bwSentinelTrap":         ahbwsentinel.ChnlUtil,
                        "interferenceUtil_bwSentinelTrap":    ahbwsentinel.InterferenceUtil,
			"txUtil_bwSentinelTrap":              ahbwsentinel.TxUtil,
			"rxUtil_bwSentinelTrap":              ahbwsentinel.RxUtil,
                        "severityLevel_trapMessage_bwSentinelTrap":   trap.Level,
                        "msgId_trapMessage_bwSentinelTrap":   trap.MsgID,
                        "desc_trapMessage_bwSentinelTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_ALARM_ALRT_TRAP_TYPE:
                var ahalarmalert AhAlarmAlrtTrap
                rawSize := int(unsafe.Sizeof(ahalarmalert))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahalarmalert))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_alarmAlertTrap":            ahutil.CleanCString(ahalarmalert.Name[:]),
                        "ifIndex_alarmAlertTrap":             ahalarmalert.IfIndex,
                        "clientMac_alarmAlertTrap":           ahutil.FormatMac(ahalarmalert.ClientMac),
                        "level_alarmAlertTrap":               ahalarmalert.Level,
						"ssid_alarmAlertTrap":                ahutil.CleanCString(ahalarmalert.SSID[:]),
                        "alertType_alarmAlertTrap":           ahalarmalert.AlertType,
                        "threshold_alarmAlertTrap":           ahalarmalert.ThresInterference,
                        "current_alarmAlertTrap":             ahalarmalert.ShortInterference,
                        "snapshot_alarmAlertTrap":            ahalarmalert.SnapInterference,
                        "failureState_alarmAlertTrap":                 ahalarmalert.Set,
                        "severityLevel_trapMessage_alarmAlertTrap":   trap.Level,
                        "msgId_trapMessage_alarmAlertTrap":   trap.MsgID,
                        "desc_trapMessage_alarmAlertTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_MESH_MGT0_VLAN_CHANGE_TRAP_TYPE:
                var ahmeshmgtvlan AhMeshMgt0VlanChangeTrap
                rawSize := int(unsafe.Sizeof(ahmeshmgtvlan))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahmeshmgtvlan))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_meshMgt0vlanChangeTrap":            ahutil.CleanCString(ahmeshmgtvlan.Name[:]),
                        "oldVlan_meshMgt0vlanChangeTrap":             ahmeshmgtvlan.OldVlan,
                        "newVlan_meshMgt0vlanChangeTrap":             ahmeshmgtvlan.NewVlan,
                        "oldNativeVlan_meshMgt0vlanChangeTrap":       ahmeshmgtvlan.OldNativeVlan,
                        "newNativeVlan_meshMgt0vlanChangeTrap":       ahmeshmgtvlan.NewNativeVlan,
                        "severityLevel_trapMessage_meshMgt0vlanChangeTrap":   trap.Level,
                        "msgId_trapMessage_meshMgt0vlanChangeTrap":   trap.MsgID,
                        "desc_trapMessage_meshMgt0vlanChangeTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_MESH_STABLE_STAGE_TRAP_TYPE:
                var ahmeshstable AhMeshStableStageTrap
                rawSize := int(unsafe.Sizeof(ahmeshstable))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahmeshstable))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_meshStableStageTrap":            ahutil.CleanCString(ahmeshstable.Name[:]),
                        "meshStableStage_meshStableStageTrap":     ahmeshstable. MeshStableStage,
                        "meshDataRate_meshStableStageTrap":        ahmeshstable.MeshDataRate,
                        "level_trapMessage_meshStableStageTrap":   trap.Level,
                        "msgId_trapMessage_meshStableStageTrap":   trap.MsgID,
                        "desc_trapMessage_meshStableStageTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	case AH_KEY_FULL_ALARM_TRAP_TYPE:
                var ahkeyfullalarm AhKeyFullAlarmTrap
                rawSize := int(unsafe.Sizeof(ahkeyfullalarm))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahkeyfullalarm))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapEvent", map[string]interface{}{
                        "trapObjName_keyFullAlarmTrap":            ahutil.CleanCString(ahkeyfullalarm.Name[:]),
                        "ifIndex_keyFullAlarmTrap":             ahkeyfullalarm.IfIndex,
                        "bssid_keyFullAlarmTrap":               ahkeyfullalarm.BSSID,
						"clientMac_keyFullAlarmTrap":           ahkeyfullalarm.ClientMAC,
						"gtkVlan_keyFullAlarmTrap":             ahkeyfullalarm.GtkVLAN,
                        "level_trapMessage_keyFullAlarmTrap":   trap.Level,
                        "msgId_trapMessage_keyFullAlarmTrap":   trap.MsgID,
                        "desc_trapMessage_keyFullAlarmTrap":    ahutil.CleanCString(trap.Desc[:]),
                        }, nil)

	}
	return nil
}

/*
Gather remaining traps that were not part of AH_TRAP_MSG_TYPE type
All the new traps should be added here
*/
func (t *TrapPlugin) Gather_Ah_send_trap(trapType uint32, trapBuf [256]byte, acc telegraf.Accumulator) error {
	switch trapType {

	case AH_FA_MVLAN_TRAP_TYPE:
		var mvlan AhFaMvlanChangeTrap
		copy((*[unsafe.Sizeof(mvlan)]byte)(unsafe.Pointer(&mvlan))[:], trapBuf[:unsafe.Sizeof(mvlan)])

		acc.AddFields("TrapEvent", map[string]interface{}{
			"trapType_faMvlanTrap":           mvlan.TrapType,
			"mgmtVlan_faMvlanTrap":           mvlan.MgmtVlan,
			"nativeVlan_faMvlanTrap":         mvlan.NativeVlan,
			"nativeTagged_faMvlanTrap":       mvlan.NativeTagged,
			"systemId_faMvlanTrap":           fmt.Sprintf("%X", mvlan.SystemID),
		}, nil)

	case AH_MSG_TRAP_DFS_BANG:
		var dfs AhTgrafDfsTrap
		copy((*[unsafe.Sizeof(dfs)]byte)(unsafe.Pointer(&dfs))[:], trapBuf[:unsafe.Sizeof(dfs)])

		acc.AddFields("TrapEvent", map[string]interface{}{
			"trapType_dfsBangTrap":    dfs.TrapType,
			"trapId_dfsBangTrap":      dfs.TrapId,
			"name_dfsBangTrap":        ahutil.CleanCString(dfs.IfName[:]),
			"desc_dfsBangTrap":        ahutil.CleanCString(dfs.Desc[:]),
		}, nil)

	}

	return nil
}

/*
Helper function to convert state value to string for SSID bind/unbind
*/
func stateToString(state uint8) string {
       switch state {
       case 0:
               return "UNBIND"
       case 1:
               return "BIND"
       default:
               return fmt.Sprintf("UNKNOWN(%d)", state)
       }
}

func (t *TrapPlugin) Ah_send_ssid_bind_unbind_trap(trapType uint32, trapBuf [600]byte, acc telegraf.Accumulator) error {
    var ssidBindUnbind AhTgrafSsidBindUnbindTrap
    copy((*[unsafe.Sizeof(ssidBindUnbind)]byte)(unsafe.Pointer(&ssidBindUnbind))[:], trapBuf[:unsafe.Sizeof(ssidBindUnbind)])

    acc.AddFields("TrapEvent", map[string]interface{}{
        "trapType_ssidBindUnbindTrap":    ssidBindUnbind.TrapType,
        "trapId_ssidBindUnbindTrap":      ssidBindUnbind.TrapID,
        "ifName_ssidBindUnbindTrap":      ahutil.CleanCString(ssidBindUnbind.IfName[:]),
        "ifIndex_ssidBindUnbindTrap":     ssidBindUnbind.IfIndex,
        "description_ssidBindUnbindTrap": ahutil.CleanCString(ssidBindUnbind.Description[:]),
        "bssidMac_ssidBindUnbindTrap":    ahutil.FormatMac(ssidBindUnbind.BssidMAC),
        "ssid_ssidBindUnbindTrap":        ahutil.CleanCString(ssidBindUnbind.SSID[:]),
	"state_ssidBindUnbindTrap":       stateToString(ssidBindUnbind.State),
    }, nil)
    return nil
}

func (t *TrapPlugin) Ah_send_bssid_spoofing_trap(trapType uint32, trapBuf [AH_TRAP_SIZE_300]byte, acc telegraf.Accumulator) error {
    var bssidSpoofing AhTgrafBSSIDSpoofingTrap
    copy((*[unsafe.Sizeof(bssidSpoofing)]byte)(unsafe.Pointer(&bssidSpoofing))[:], trapBuf[:unsafe.Sizeof(bssidSpoofing)])

    acc.AddFields("TrapEvent", map[string]interface{}{
        "trapId_bssidSpoofingTrap":       bssidSpoofing.TrapID,
        "ifName_bssidSpoofingTrap":       ahutil.CleanCString(bssidSpoofing.IfName[:]),
        "description_bssidSpoofingTrap":  ahutil.CleanCString(bssidSpoofing.Description[:]),
        "ifIndex_bssidSpoofingTrap":      bssidSpoofing.IfIndex,
        "bssidMac_bssidSpoofingTrap":     ahutil.FormatMac(bssidSpoofing.BssidMAC),
        "attackMac_bssidSpoofingTrap":    ahutil.FormatMac(bssidSpoofing.AttackMAC),
        "attackCount_bssidSpoofingTrap":  bssidSpoofing.AttackCount,
        "protocol_bssidSpoofingTrap":     bssidSpoofing.Protocol,
        "severity_bssidSpoofingTrap":     bssidSpoofing.Severity,
        "sourceIp_bssidSpoofingTrap":     ahutil.IntToIpv4(bssidSpoofing.SourceIP),
        "targetIp_bssidSpoofingTrap":     ahutil.IntToIpv4(bssidSpoofing.TargetIP),
    }, nil)
    return nil
}

/*
 trapListener listens for incoming trap messages on a UDP connection,
 extracts the trap type and payload, and processes supported trap types.
*/
func (t *TrapPlugin) trapListener(conn net.PacketConn) {
	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("[ah_trap] Read error: %v", err)
			return
		}

		if n < 4 {
			log.Printf("[ah_trap] Received too short message: %d bytes", n)
			continue
		}

		trapType := binary.LittleEndian.Uint32(buf[:4])
		payload := buf[4:n]

		//Enable this log only for testing or else this will flood as channel power trap will be initaited all the time
//		log.Printf("[ah_trap] Received trap type: %d ", trapType)
		switch trapType {
		case AH_TRAP_MSG_TYPE:
			var trap AhTrapMsg
			expected := int(unsafe.Sizeof(trap))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid AhTrapMsg size: got %d, expected %d", len(payload), expected)
				continue
			}
			copy((*[unsafe.Sizeof(AhTrapMsg{})]byte)(unsafe.Pointer(&trap))[:], payload)

			if err := t.Gather_Ah_Logen(trap, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering trap: %v", err)
			}

		case AH_FA_MVLAN_TRAP_TYPE:
			var mvlan AhFaMvlanChangeTrap
			expected := int(unsafe.Sizeof(mvlan))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid FA_MVLAN size: got %d, expected %d", len(payload), expected)
				continue
			}

			var trapBuf [256]byte
			copy(trapBuf[:expected], payload)
			if err := t.Gather_Ah_send_trap(trapType, trapBuf, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering mvlan trap: %v", err)
			}

		case AH_MSG_TRAP_DFS_BANG:
			var dfs AhTgrafDfsTrap
			expected := int(unsafe.Sizeof(dfs))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid DFS BANG size: got %d, expected %d", len(payload), expected)
				continue
			}
			var trapBuf [256]byte
			copy(trapBuf[:expected], payload)
			if err := t.Gather_Ah_send_trap(trapType, trapBuf, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering DFS trap: %v", err)
			}

		case AH_MSG_TRAP_STA_LEAVE_STATS:
			var staLeave AhStaLeaveStatsTrap
			expected := int(unsafe.Sizeof(staLeave))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid STA LEAVE STATS size: got %d, expected %d", len(payload), expected)
				continue
			}
			var trapBuf [600]byte
			copy(trapBuf[:expected], payload)
			if err := t.Ah_send_sta_leave_trap(trapType, trapBuf, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering STA LEAVE STATS trap: %v", err)
			}
		case AH_MSG_TRAP_SSID_BIND_UNBIND:
			var  SsidBindUnbind AhTgrafSsidBindUnbindTrap
			expected := int(unsafe.Sizeof(SsidBindUnbind))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid SSID BIND UNBIND STATS size: got %d, expected %d", len(payload), expected)
				continue
			}
			var trapBuf [600]byte
			copy(trapBuf[:expected], payload)
			if err := t.Ah_send_ssid_bind_unbind_trap(trapType, trapBuf, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering SSID Bind Unbind trap: %v", err)
			}
		case AH_MSG_TRAP_BSSID_SPOOFING:
			var bssidSpoofing AhTgrafBSSIDSpoofingTrap
			expected := int(unsafe.Sizeof(bssidSpoofing))
			if len(payload) != expected {
				log.Printf("[ah_trap] Invalid BSSID SPOOFING size: got %d, expected %d", len(payload), expected)
				continue
			}
			var trapBuf [AH_TRAP_SIZE_300]byte
			copy(trapBuf[:expected], payload)
			if err := t.Ah_send_bssid_spoofing_trap(trapType, trapBuf, t.acc); err != nil {
				log.Printf("[ah_trap] Error gathering BSSID Spoofing trap: %v", err)
			}
		}
	}
}

func (t *TrapPlugin) Gather(acc telegraf.Accumulator) error {
	// No-op: event-driven mode
	defer func() {
               if r := recover(); r != nil {
                       stack := debug.Stack()
                       log.Printf("[ah_trap] telegraf crash recovered: %s\n", stack)
               }
       }()
       return nil
}

func init() {
	inputs.Add("ah_trap", func() telegraf.Input {
		return &TrapPlugin{}
	})
}
