package ah_trap

import (
	"bytes"
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

func cleanCString(b []byte) string {
    if i := bytes.IndexByte(b, 0); i != -1 {
        return string(b[:i])
    }
    return string(b)
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

		acc.AddFields("TrapType", map[string]interface{}{
			"name_failureTrapType":              cleanCString(failure.Name[:]),
			"cause_failureTrapType":             failure.Cause,
			"set_failureTrapType":               failure.Set,
			"level_trapMessage_failureTrapType": trap.Level,
			"msgId_trapMessage_failureTrapType": trap.MsgID,
			"desc_trapMessage_failureTrapType":  cleanCString(trap.Desc[:]),
		}, nil)

	case AH_THRESHOLD_TRAP_TYPE:
		var ahthreshold AhThresholdTrap
		rawSize := int(unsafe.Sizeof(ahthreshold))
		copy((*[1 << 10]byte)(unsafe.Pointer(&ahthreshold))[:rawSize], trap.Union[:rawSize])

		acc.AddFields("TrapType", map[string]interface{}{
			"name_thresholdTrapType":              cleanCString(ahthreshold.Name[:]),
			"curVal_thresholdTrapType":            ahthreshold.CurVal,
			"thresholdHigh_thresholdTrapType":     ahthreshold.ThresholdHigh,
			"thresholdLow_thresholdTrapType":      ahthreshold.ThresholdLow,
			"level_trapMessage_thresholdTrapType": trap.Level,
			"msgId_trapMessage_thresholdTrapType": trap.MsgID,
			"desc_trapMessage_thresholdTrapType":  cleanCString(trap.Desc[:]),
		}, nil)

        case AH_STATE_CHANGE_TRAP_TYPE:
                var ahstatechange AhStateChangeTrap
                rawSize := int(unsafe.Sizeof(ahstatechange))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahstatechange))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_stateChangeTrapType":              cleanCString(ahstatechange.Name[:]),
                        "preState_stateChangeTrapType":          ahstatechange.PreState,
                        "curState_stateChangeTrapType":          ahstatechange.CurState,
                        "opMode_stateChangeTrapType":            ahstatechange.OperationMode,
                        "level_trapMessage_stateChangeTrapType": trap.Level,
                        "msgId_trapMessage_stateChangeTrapType": trap.MsgID,
                        "desc_trapMessage_stateChangeTrapType":  cleanCString(trap.Desc[:]),
                }, nil)

	 case AH_CONNECTION_CHANGE_TRAP_TYPE:
                var ahconnectionchange AhConnectionChangeTrap
                rawSize := int(unsafe.Sizeof(ahconnectionchange))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahconnectionchange))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_connectionChangeTrapType":              cleanCString(ahconnectionchange.Name[:]),
                        "ssid_connectionChangeTrapType":              ahconnectionchange.Ssid,
                        "hostName_connectionChangeTrapType":          ahconnectionchange.HostName,
                        "userName_connectionChangeTrapType":          ahconnectionchange.UserName,
			"objectType_connectionChangeTrapType":        ahconnectionchange.ObjectType,
			"remoteId_connectionChangeTrapType":          formatMac(ahconnectionchange.RemoteID),
			"bssid_connectionChangeTrapType":             formatMac(ahconnectionchange.BSSID),
			"curState_connectionChangeTrapType":          ahconnectionchange.CurState,
			"clientIp_connectionChangeTrapType":          intToIPv4(ahconnectionchange.ClientIP),
			"clientAuthMethod_connectionChangeTrapType":  ahconnectionchange.ClientAuthMethod,
			"clientEncryptMethod_connectionChangeTrapType": ahconnectionchange.ClientEncryptMethod,
			"clientMacProto_connectionChangeTrapType":    ahconnectionchange.ClientMacProto,
			"clientVlan_connectionChangeTrapType":        ahconnectionchange.ClientVLAN,
			"clientUpId_connectionChangeTrapType":        ahconnectionchange.ClientUPID,
			"clientChannel_connectionChangeTrapType":     ahconnectionchange.ClientChannel,
			"clientCwpUsed_connectionChangeTrapType":     ahconnectionchange.ClientCWPUsed,
			"assocTime_connectionChangeTrapType":         ahconnectionchange.AssociationTime,
			"ifIndex_keys_connectionChangeTrapType":      ahconnectionchange.IfIndex,
                        "name_keys_connectionChangeTrapType":         cleanCString(ahconnectionchange.IfName[:]),
			"rssi_connectionChangeTrapType":              ahconnectionchange.RSSI,
			"snr_connectionChangeTrapType":               ahconnectionchange.SNR,
			"profile_connectionChangeTrapType":           cleanCString(ahconnectionchange.ProfName[:]),
			"authUsed_connectionChangeTrapType":          ahconnectionchange.ClientMacBasedAuthUsed,
			"os_connectionChangeTrapType":                cleanCString(ahconnectionchange.OS[:]),
			"option55_connectionChangeTrapType":          cleanCString(ahconnectionchange.Option55[:]),
			"mgtStatus_connectionChangeTrapType":         ahconnectionchange.MgtStus,
			"staAddr6Num_connectionChangeTrapType":       ahconnectionchange.StaAddr6Num,
			"staAddr6_connectionChangeTrapType":          intToIPv6(ahconnectionchange.StaAddr6[:], int(ahconnectionchange.StaAddr6Num)),
			"deauthReason_connectionChangeTrapType":      ahconnectionchange.DeauthReason,
			"roamTime_connectionChangeTrapType":          ahconnectionchange.RoamTime,
			"authTime_connectionChangeTrapType":          ahconnectionchange.AuthTime,
			"radioProfile_connectionChangeTrapType":      cleanCString(ahconnectionchange.RadioProf[:]),
			"negotiateKbps_connectionChangeTrapType":     ahconnectionchange.NegotiateKbps,
                        "level_trapMessage_connectionChangeTrapType": trap.Level,
                        "msgId_trapMessage_connectionChangeTrapType": trap.MsgID,
                        "desc_trapMessage_connectionChangeTrapType":  cleanCString(trap.Desc[:]),
                }, nil)

	case AH_IDP_AP_EVENT_TRAP_TYPE:
                var ahidpapevent AhIdpApEventTrap
                rawSize := int(unsafe.Sizeof(ahidpapevent))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahidpapevent))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_idpApEventTrapType":              cleanCString(ahidpapevent.Name[:]),
                        "ifIndex_idpApEventTrapType":           ahidpapevent.IfIndex,
                        "remoteId_idpApEventTrapType":          formatMac(ahidpapevent.RemoteID),
                        "idpType_idpApEventTrapType":           ahidpapevent.IdpType,
			"idpChannel_idpApEventTrapType":        ahidpapevent.IdpChannel,
			"idpRssi_idpApEventTrapType":           ahidpapevent.IdpRSSI,
			"idpCompliance_idpApEventTrapType":     ahidpapevent.IdpCompliance,
			"ssid_idpApEventTrapType":              cleanCString(ahidpapevent.SSID[:]),
			"stationType_idpApEventTrapType":       ahidpapevent.StationType,
			"stationData_idpApEventTrapType":       ahidpapevent.StationData,
			"idpRemoved_idpApEventTrapType":        ahidpapevent.IdpRemoved,
			"idpInnet_idpApEventTrapType":          ahidpapevent.IdpInNet,
                        "level_trapMessage_idpApEventTrapType": trap.Level,
                        "msgId_trapMessage_idpApEventTrapType": trap.MsgID,
                        "desc_trapMessage_idpApEventTrapType":  cleanCString(trap.Desc[:]),
                }, nil)

	case AH_CLIENT_INFO_TRAP_TYPE:
                var ahclientinfo AhClientInfoTrap
                rawSize := int(unsafe.Sizeof(ahclientinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahclientinfo))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_clientInfoTrapType":              cleanCString(ahclientinfo.Name[:]),
                        "ssid_clientInfoTrapType":              ahclientinfo.Ssid,
			"clientMac_clientInfoTrapType":         formatMac(ahclientinfo.ClientMac),
                        "hostName_clientInfoTrapType":          cleanCString(ahclientinfo.HostName[:]),
			"userName_clientInfoTrapType":          cleanCString(ahclientinfo.UserName[:]),
			"clientIp_clientInfoTrapType":          intToIPv4(ahclientinfo.ClientIP),
			"mgtStatus_clientInfoTrapType":         ahclientinfo.MgtStus,
			"staAddr6Num_clientInfoTrapType":       ahclientinfo.StaAddr6Num,
			"staAddr6_clientInfoTrapType":          intToIPv6(ahclientinfo.StaAddr6[:], int(ahclientinfo.StaAddr6Num)),
                        "level_trapMessage_clientInfoTrapType": trap.Level,
                        "msgId_trapMessage_clientInfoTrapType": trap.MsgID,
                        "desc_trapMessage_clientInfoTrapType":  cleanCString(trap.Desc[:]),
                }, nil)

         case AH_POWER_INFO_TRAP_TYPE:
                var ahpowerinfo AhPowerInfoTrap
                rawSize := int(unsafe.Sizeof(ahpowerinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahpowerinfo))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_powerInfoTrapType":               cleanCString(ahpowerinfo.Name[:]),
                        "powerSrc_powerInfoTrapType":           ahpowerinfo.PowerSrc,
                        "eth0On_powerInfoTrapType":             ahpowerinfo.Eth0On,
                        "eth1On_powerInfoTrapType":             ahpowerinfo.Eth1On,
                        "eth0Pwr_powerInfoTrapType":            ahpowerinfo.Eth0Pwr,
                        "eth1Pwr_powerInfoTrapType":            ahpowerinfo.Eth1Pwr,
                        "eth0Speed_powerInfoTrapType":          ahpowerinfo.Eth0Speed,
                        "eth1Speed_powerInfoTrapType":          ahpowerinfo.Eth1Speed,
                        "wifi0Setting_powerInfoTrapType":       ahpowerinfo.Wifi0Setting,
                        "wifi1Setting_powerInfoTrapType":       ahpowerinfo.Wifi1Setting,
                        "wifi2Setting_powerInfoTrapType":       ahpowerinfo.Wifi2Setting,
			"level_trapMessage_powerInfoTrapType":  trap.Level,
                        "msgId_trapMessage_powerInfoTrapType":  trap.MsgID,
			"desc_trapMessage_powerInfoTrapType":   cleanCString(trap.Desc[:]),
			}, nil)

	case AH_CHANNEL_POWER_TRAP_TYPE:
                var ahchannelpowerinfo AhChannelPowerTrap
                rawSize := int(unsafe.Sizeof(ahchannelpowerinfo))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahchannelpowerinfo))[:rawSize], trap.Union[:rawSize])


                acc.AddFields("TrapType", map[string]interface{}{
                        "name_channelPowerTrapType":              cleanCString(ahchannelpowerinfo.Name[:]),
                        "ifIndex_channelPowerTrapType":           ahchannelpowerinfo.IfIndex,
                        "channel_channelPowerTrapType":           ahchannelpowerinfo.RadioChannel,
                        "txPwr_channelPowerTrapType":             ahchannelpowerinfo.RadioTxPower,
                        "beaconInterval_channelPowerTrapType":    ahchannelpowerinfo.BeaconInterval,
                        "channelStrfmt_channelPowerTrapType":     ahchannelpowerinfo.ChnlStrfmt,
                        "powerStrfmt_channelPowerTrapType":       ahchannelpowerinfo.PwrStrfmt,
			"radioEirp_channelPowerTrapType":         cleanCString(ahchannelpowerinfo.RadioEirp[:]),
                        "reason_channelPowerTrapType":            ahchannelpowerinfo.Reason,
                        "level_trapMessage_channelPowerTrapType":  trap.Level,
                        "msgId_trapMessage_channelPowerTrapType":  trap.MsgID,
                        "desc_trapMessage_channelPowerTrapType":   cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_IDP_MITIGATE_TRAP_TYPE:
                var ahidpmitigate AhIdpMitigateTrap
                rawSize := int(unsafe.Sizeof(ahidpmitigate))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahidpmitigate))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_idpMitigateTrapType":               cleanCString(ahidpmitigate.Name[:]),
                        "ifIndex_idpMitigateTrapType":            ahidpmitigate.IfIndex,
			"remoteId_idpMitigateTrapType":           formatMac(ahidpmitigate.RemoteID),
                        "bssid_idpMitigateTrapType":              formatMac(ahidpmitigate.BSSID),
                        "removed_idpMitigateTrapType":            ahidpmitigate.Removed,
                        "discoverAge_idpMitigateTrapType":        ahidpmitigate.DiscoverAge,
                        "updateAge_idpMitigateTrapType":          ahidpmitigate.UpdateAge,
                        "level_trapMessage_idpMitigateTrapType":  trap.Level,
                        "msgId_trapMessage_idpMitigateTrapType":  trap.MsgID,
                        "desc_trapMessage_idpMitigateTrapType":   cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_INTERFERENCE_ALERT_TRAP_TYPE:
                var ahinterference AhInterferenceAlertTrap
                rawSize := int(unsafe.Sizeof(ahinterference))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahinterference))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_interferenceAlertTrapType":                cleanCString(ahinterference.Name[:]),
                        "ifIndex_interferenceAlertTrapType":             ahinterference.IfIndex,
                        "interferenceThres_interferenceAlertTrapType":   ahinterference.InterferenceThres,
                        "aveInterference_interferenceAlertTrapType":     ahinterference.AveInterference,
                        "shortInterference_interferenceAlertTrapType":   ahinterference.ShortInterference,
                        "snapInterference_interferenceAlertTrapType":    ahinterference.SnapInterference,
                        "crcErrRateThreshold_interferenceAlertTrapType": ahinterference.CRCErrRateThres,
			"crcErrRate_interferenceAlertTrapType":          ahinterference.CRCErrRate,
			"set_interferenceAlertTrapType":                 ahinterference.Set,
                        "level_trapMessage_interferenceAlertTrapType":   trap.Level,
                        "msgId_trapMessage_interferenceAlertTrapType":   trap.MsgID,
                        "desc_trapMessage_interferenceAlertTrapType":    cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_BW_SENTINEL_TRAP_TYPE:
		log.Printf("[ah_trap] Bw sentinel trap 2")
                var ahbwsentinel AhBwSentinelTrap
                rawSize := int(unsafe.Sizeof(ahbwsentinel))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahbwsentinel))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_bwSentinelTrapType":                cleanCString(ahbwsentinel.Name[:]),
                        "ifIndex_bwSentinelTrapType":             ahbwsentinel.IfIndex,
                        "clientMac_bwSentinelTrapType":           formatMac(ahbwsentinel.ClientMac),
                        "bwSentinelStatus_bwSentinelTrapType":    ahbwsentinel.BwSentinelStatus,
                        "gbw_bwSentinelTrapType":		  ahbwsentinel.GBW,
                        "actualBw_bwSentinelTrapType":            ahbwsentinel.ActualBW,
                        "actionTaken_bwSentinelTrapType":         ahbwsentinel.ActionTaken,
                        "channelUtil_bwSentinelTrapType":         ahbwsentinel.ChnlUtil,
                        "interferenceUtil_bwSentinelTrapType":    ahbwsentinel.InterferenceUtil,
			"txUtil_bwSentinelTrapType":              ahbwsentinel.TxUtil,
			"rxUtil_bwSentinelTrapType":              ahbwsentinel.RxUtil,
                        "level_trapMessage_bwSentinelTrapType":   trap.Level,
                        "msgId_trapMessage_bwSentinelTrapType":   trap.MsgID,
                        "desc_trapMessage_bwSentinelTrapType":    cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_ALARM_ALRT_TRAP_TYPE:
                var ahalarmalert AhAlarmAlrtTrap
                rawSize := int(unsafe.Sizeof(ahalarmalert))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahalarmalert))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_alarmAlertTrapType":                cleanCString(ahalarmalert.Name[:]),
                        "ifIndex_alarmAlertTrapType":             ahalarmalert.IfIndex,
                        "clientMac_alarmAlertTrapType":           formatMac(ahalarmalert.ClientMac),
                        "level_alarmAlertTrapType":               ahalarmalert.Level,
			"ssid_alarmAlertTrapType":                cleanCString(ahalarmalert.SSID[:]),
                        "alertType_alarmAlertTrapType":           ahalarmalert.AlertType,
                        "threshold_alarmAlertTrapType":           ahalarmalert.ThresInterference,
                        "current_alarmAlertTrapType":             ahalarmalert.ShortInterference,
                        "snap_alarmAlertTrapType":                ahalarmalert.SnapInterference,
                        "set_alarmAlertTrapType":                 ahalarmalert.Set,
                        "level_trapMessage_alarmAlertTrapType":   trap.Level,
                        "msgId_trapMessage_alarmAlertTrapType":   trap.MsgID,
                        "desc_trapMessage_alarmAlertTrapType":    cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_MESH_MGT0_VLAN_CHANGE_TRAP_TYPE:
                var ahmeshmgtvlan AhMeshMgt0VlanChangeTrap
                rawSize := int(unsafe.Sizeof(ahmeshmgtvlan))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahmeshmgtvlan))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_meshMgt0vlanChangeTrapType":                cleanCString(ahmeshmgtvlan.Name[:]),
                        "oldVlan_meshMgt0vlanChangeTrapType":             ahmeshmgtvlan.OldVlan,
                        "newVlan_meshMgt0vlanChangeTrapType":             ahmeshmgtvlan.NewVlan,
                        "oldNativeVlan_meshMgt0vlanChangeTrapType":       ahmeshmgtvlan.OldNativeVlan,
                        "newNativeVlan_meshMgt0vlanChangeTrapType":       ahmeshmgtvlan.NewNativeVlan,
                        "level_trapMessage_meshMgt0vlanChangeTrapType":   trap.Level,
                        "msgId_trapMessage_meshMgt0vlanChangeTrapType":   trap.MsgID,
                        "desc_trapMessage_meshMgt0vlanChangeTrapType":    cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_MESH_STABLE_STAGE_TRAP_TYPE:
                var ahmeshstable AhMeshStableStageTrap
                rawSize := int(unsafe.Sizeof(ahmeshstable))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahmeshstable))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_meshStableStageTrapType":                cleanCString(ahmeshstable.Name[:]),
                        "meshStableStage_meshStableStageTrapType":     ahmeshstable. MeshStableStage,
                        "meshDataRate_meshStableStageTrapType":        ahmeshstable.MeshDataRate,
                        "level_trapMessage_meshStableStageTrapType":   trap.Level,
                        "msgId_trapMessage_meshStableStageTrapType":   trap.MsgID,
                        "desc_trapMessage_meshStableStageTrapType":    cleanCString(trap.Desc[:]),
                        }, nil)

	case AH_KEY_FULL_ALARM_TRAP_TYPE:
                var ahkeyfullalarm AhKeyFullAlarmTrap
                rawSize := int(unsafe.Sizeof(ahkeyfullalarm))
                copy((*[1 << 10]byte)(unsafe.Pointer(&ahkeyfullalarm))[:rawSize], trap.Union[:rawSize])

                acc.AddFields("TrapType", map[string]interface{}{
                        "name_keyFullAlarmTrapType":                cleanCString(ahkeyfullalarm.Name[:]),
                        "ifIndex_keyFullAlarmTrapType":             ahkeyfullalarm.IfIndex,
                        "bssid_keyFullAlarmTrapType":               ahkeyfullalarm.BSSID,
			"clientMac_keyFullAlarmTrapType":           ahkeyfullalarm.ClientMAC,
			"gtkVlan_keyFullAlarmTrapType":             ahkeyfullalarm.GtkVLAN,
                        "level_trapMessage_keyFullAlarmTrapType":   trap.Level,
                        "msgId_trapMessage_keyFullAlarmTrapType":   trap.MsgID,
                        "desc_trapMessage_keyFullAlarmTrapType":    cleanCString(trap.Desc[:]),
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

		acc.AddFields("TrapType", map[string]interface{}{
			"trapType_faMvlanTrapType":           mvlan.TrapType,
			"mgmtVlan_faMvlanTrapType":           mvlan.MgmtVlan,
			"nativeVlan_faMvlanTrapType":         mvlan.NativeVlan,
			"nativeTagged_faMvlanTrapType":       mvlan.NativeTagged,
			"systemId_faMvlanTrapType":           fmt.Sprintf("%X", mvlan.SystemID),
		}, nil)

	case AH_MSG_TRAP_DFS_BANG:
		var dfs AhTgrafDfsTrap
		copy((*[unsafe.Sizeof(dfs)]byte)(unsafe.Pointer(&dfs))[:], trapBuf[:unsafe.Sizeof(dfs)])

		acc.AddFields("TrapType", map[string]interface{}{
			"trapType_dfsBangTrapType":    dfs.TrapType,
			"name_dfsBangTrapType":        cleanCString(dfs.IfName[:]),
			"desc_dfsBangTrapType":        cleanCString(dfs.Desc[:]),
		}, nil)

	}

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
