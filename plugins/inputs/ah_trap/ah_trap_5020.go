// +build AP5020

package ah_trap

import (
	"log"
	"unsafe"
	"fmt"
	"strings"
	"strconv"
	"os/exec"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/ahutil"
)

type AhStaLeaveStatsTrap struct {
    TrapId          uint8
    DataLen         uint16
    ObjNameLen      uint8
    ObjName         [MAX_OBJ_NAME_LEN]byte
    ReasonCode      uint32
    DesLen          uint8
    Describable     [MAX_DESCRIBLE_LEN]uint8
    DisassocTime    uint32
    IfIndex         uint32
    Mac             [MACADDR_LEN]uint8
    Rssi            uint32
    LinkupTime      uint32
    AuthMethod      uint8
    EncryptMethod   uint8
    MacProtocol     uint8
    CwpUsed         uint8
    Vlan            uint32
    UserProfileId   uint32
    Channel         uint32
    LastTxrate      uint32
    LastRxrate      uint32
    RxDataFrames    uint32
    RxDataOctets    uint32
    RxMgtFrames     uint32
    RxUcFrames      uint32
    RxMcFrames      uint32
    RxBcFrames      uint32
    RxMicFailure    uint32
    TxDataFrames    uint32
    TxMgtFrames     uint32
    TxDataOctets    uint32
    TxUcFrames      uint32
    TxMcFrames      uint32
    TxBcFrames      uint32
    Ip              uint32
    HostName        [AH_CAPWAP_STAT_NAME_MAX_LEN + 1]byte
    SsidName        [AH_CAPWAP_STAT_NAME_MAX_LEN + 1]byte
    UserName        [AH_CAPWAP_STAT_NAME_MAX_LEN + 1]byte
    TxBeDataFrames  uint32
    TxBgDataFrames  uint32
    TxViDataFrames  uint32
    TxVoDataFrames  uint32
    RxAirTime       uint64
    TxAirTime       uint64
    ClientBssid     [MACADDR_LEN]uint8
    Ts              uint32
    IfName          [AH_CAPWAP_STAT_NAME_MAX_LEN + 1]byte  // char array
    StaAddr6Num     uint8
    StaAddr6        [AH_MAX_NUM_STA_ADDRS6]AhStaAddr6
    _               [4]byte // Padding: Forces total size to 528 bytes for 8-byte alignment
}



func (t *TrapPlugin) Ah_send_sta_leave_trap(trapType uint32, trapBuf [600]byte, acc telegraf.Accumulator) error {
	var staLeave AhStaLeaveStatsTrap
	var description string
	var if_name string

	copy((*[unsafe.Sizeof(staLeave)]byte)(unsafe.Pointer(&staLeave))[:], trapBuf[:unsafe.Sizeof(staLeave)])

	if_name = ahutil.CleanCString(staLeave.IfName[:])

	ssid,_ := ahutil.GetSSIDbyIfName(if_name)
	bssid,_ := ahutil.GetBSSIDbyIfName(if_name)
	
	description = fmt.Sprintf("Station %02x%02x:%02x%02x:%02x%02x is deauthenticated from %02x%02x:%02x%02x:%02x%02x thru SSID %s\n",
			staLeave.Mac[0], staLeave.Mac[1], staLeave.Mac[2], staLeave.Mac[3], staLeave.Mac[4], staLeave.Mac[5],
			bssid[0], bssid[1], bssid[2], bssid[3], bssid[4], bssid[5],
			ssid)

	acc.AddFields("TrapEvent", map[string]interface{}{
		"ifIndex_keys_staLeaveStatsTrap":		staLeave.IfIndex,
		"name_keys_staLeaveStatsTrap":			if_name,

		"trapId_staLeaveStatsTrap":				staLeave.TrapId,

		"objectName_staLeaveStatsTrap":			ahutil.CleanCString(staLeave.ObjName[:]),
		"reasonCode_staLeaveStatsTrap":			staLeave.ReasonCode,
		"description_staLeaveStatsTrap":		description,
		"disassocTime_staLeaveStatsTrap":		staLeave.DisassocTime,
		"mac_staLeaveStatsTrap":				ahutil.FormatMac(staLeave.Mac),
		"rssi_staLeaveStatsTrap":				staLeave.Rssi,
		"linkUptime_staLeaveStatsTrap":			staLeave.LinkupTime,
		"clientAuthMethod_staLeaveStatsTrap":		staLeave.AuthMethod,
		"clientEncryptMethod_staLeaveStatsTrap":	staLeave.EncryptMethod,
		"clientMacProto_staLeaveStatsTrap":		staLeave.MacProtocol,
		"clientCwpUsed_staLeaveStatsTrap":		staLeave.CwpUsed,
		"clientVlan_staLeaveStatsTrap":			staLeave.Vlan,
		"clientChannel_staLeaveStatsTrap":		staLeave.Channel,
		"lastTxrate_staLeaveStatsTrap":			staLeave.LastTxrate,
		"lastRxrate_staLeaveStatsTrap":			staLeave.LastRxrate,

		"rxdataframes_staLeaveStatsTrap":	staLeave.RxDataFrames,
		"txdataframes_staLeaveStatsTrap":	staLeave.TxDataFrames,
		"rxdatabytes_staLeaveStatsTrap":	staLeave.RxDataOctets,
		"txdatabytes_staLeaveStatsTrap":	staLeave.TxDataOctets,
		"rxmgmtframes_staLeaveStatsTrap":	staLeave.RxMgtFrames,
		"txmgmtframes_staLeaveStatsTrap":	staLeave.TxMgtFrames,
		"rxucframes_staLeaveStatsTrap":		staLeave.RxUcFrames,
		"txucframes_staLeaveStatsTrap":		staLeave.TxUcFrames,
		"rxmcframes_staLeaveStatsTrap":		staLeave.RxMcFrames,
		"txmcframes_staLeaveStatsTrap":		staLeave.TxMcFrames,
		"rxbcframes_staLeaveStatsTrap":		staLeave.RxBcFrames,
		"txbcframes_staLeaveStatsTrap":		staLeave.TxBcFrames,
		"rxmicfailures_staLeaveStatsTrap":	staLeave.RxMicFailure,
		"clientIp_staLeaveStatsTrap":		ahutil.IntToIpv4(staLeave.Ip),
		"hostName_staLeaveStatsTrap":		ahutil.CleanCString(staLeave.HostName[:]),
		"userName_staLeaveStatsTrap":		ahutil.CleanCString(staLeave.UserName[:]),
		"ssid_staLeaveStatsTrap":			ssid,
		"bssid_staLeaveStatsTrap":			ahutil.FormatMac(bssid),
		"txbeframes_staLeaveStatsTrap":		staLeave.TxBeDataFrames,
		"txbgframes_staLeaveStatsTrap":		staLeave.TxBgDataFrames,
		"txviframes_staLeaveStatsTrap":		staLeave.TxViDataFrames,
		"txvovframes_staLeaveStatsTrap":	staLeave.TxVoDataFrames,
		"rxairtime_staLeaveStatsTrap":		staLeave.RxAirTime,
		"txairtime_staLeaveStatsTrap":		staLeave.TxAirTime,
		"ts_staLeaveStatsTrap":				staLeave.Ts,
		"staAddr6Num_staLeaveStatsTrap":	staLeave.StaAddr6Num,
		"staAddr6_staLeaveStatsTrap":		IntToIPv6_1(staLeave.StaAddr6[:], int(staLeave.StaAddr6Num)),
	}, nil)
	return nil
}