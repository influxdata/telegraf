// +build AP3000 AP5000

package ah_trap

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/ahutil"
)

// gatherConnectionChangeTrap handles CONNECTION_CHANGE trap without MLO fields for AP3000/AP5000
func gatherConnectionChangeTrap(ahconnectionchange AhConnectionChangeTrap, trap AhTrapMsg, acc telegraf.Accumulator) {
	acc.AddFields("TrapEvent", map[string]interface{}{
		"trapObjName_connectionChangeTrap":          ahutil.CleanCString(ahconnectionchange.Name[:]),
		"ssid_connectionChangeTrap":              ahutil.CleanCString(ahconnectionchange.Ssid[:]),
		"hostName_connectionChangeTrap":          ahutil.CleanCString(ahconnectionchange.HostName[:]),
		"userName_connectionChangeTrap":          ahutil.CleanCString(ahconnectionchange.UserName[:]),
		"objectType_connectionChangeTrap":        ahconnectionchange.ObjectType,
		"remoteId_connectionChangeTrap":          ahutil.FormatMac(ahconnectionchange.RemoteID),
		"bssid_connectionChangeTrap":             ahutil.FormatMac(ahconnectionchange.BSSID),
		"curState_connectionChangeTrap":          ahconnectionchange.CurState,
		"clientIp_connectionChangeTrap":          ahutil.IntToIpv4(ahconnectionchange.ClientIP),
		"clientAuthMethod_connectionChangeTrap":  ahconnectionchange.ClientAuthMethod,
		"clientEncryptMethod_connectionChangeTrap": ahconnectionchange.ClientEncryptMethod,
		"clientMacProto_connectionChangeTrap":    ahconnectionchange.ClientMacProto,
		"clientVlan_connectionChangeTrap":        ahconnectionchange.ClientVLAN,
		"clientUpId_connectionChangeTrap":        ahconnectionchange.ClientUPID,
		"clientChannel_connectionChangeTrap":     ahconnectionchange.ClientChannel,
		"clientCwpUsed_connectionChangeTrap":     ahconnectionchange.ClientCWPUsed,
		"associationTime_connectionChangeTrap":   ahconnectionchange.AssociationTime,
		"ifIndex_keys_connectionChangeTrap":      ahconnectionchange.IfIndex,
		"name_keys_connectionChangeTrap":         ahutil.CleanCString(ahconnectionchange.IfName[:]),
		"rssi_connectionChangeTrap":              ahconnectionchange.RSSI,
		"snr_connectionChangeTrap":               ahconnectionchange.SNR,
		"profile_connectionChangeTrap":           ahutil.CleanCString(ahconnectionchange.ProfName[:]),
		"authUsed_connectionChangeTrap":          ahconnectionchange.ClientMacBasedAuthUsed,
		"os_connectionChangeTrap":                ahutil.CleanCString(ahconnectionchange.OS[:]),
		"option55_connectionChangeTrap":          ahutil.CleanCString(ahconnectionchange.Option55[:]),
		"mgtStatus_connectionChangeTrap":         ahconnectionchange.MgtStus,
		"staAddr6Num_connectionChangeTrap":       ahconnectionchange.StaAddr6Num,
		"staAddr6_connectionChangeTrap":          intToIPv6(ahconnectionchange.StaAddr6[:], int(ahconnectionchange.StaAddr6Num)),
		"deauthReason_connectionChangeTrap":      ahconnectionchange.DeauthReason,
		"roamTime_connectionChangeTrap":          ahconnectionchange.RoamTime,
		"assocTime_connectionChangeTrap":         ahconnectionchange.AssocTime,
		"authTime_connectionChangeTrap":          ahconnectionchange.AuthTime,
		"radioProfile_connectionChangeTrap":      ahutil.CleanCString(ahconnectionchange.RadioProf[:]),
		"negotiateKbps_connectionChangeTrap":     ahconnectionchange.NegotiateKbps,
		// No MLO fields for AP3000/AP5000
		"severityLevel_trapMessage_connectionChangeTrap": trap.Level,
		"msgId_trapMessage_connectionChangeTrap": trap.MsgID,
		"desc_trapMessage_connectionChangeTrap":  ahutil.CleanCString(trap.Desc[:]),
	}, nil)
}
