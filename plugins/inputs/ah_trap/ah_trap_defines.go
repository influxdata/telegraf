package ah_trap


import (
	"net"
)

type AhTrapType uint32

const (

	AH_MAX_TRAP_OBJ_NAME     = 64
	AH_MAX_TRAP_IF_NAME      =  16
	AH_MAX_TRAP_SSID_NAME    = 32
	AH_MAX_TRAP_HOST_NAME    = 32
	AH_MAX_TRAP_USER_NAME    = 128
	AH_MAX_TRAP_PROF_NAME    = 32
	AH_MAX_NAME_LEN          = 32
	AH_UCHAR_MAX             = 255
	AH_MAX_NUM_STA_ADDRS6    = 5
	AH_TRAP_MSG_TYPE         = 1
	AH_FA_MVLAN_TRAP_TYPE    = 124
	TRAP_DCRPT_LEN           = 96
	AH_MSG_TRAP_DFS_BANG     = 12
	AH_MSG_TRAP_STA_LEAVE_STATS = 6
	MACADDR_LEN		 = 6
	MAX_DESCRIBLE_LEN	 = 128
	AH_CAPWAP_STAT_NAME_MAX_LEN = 32
	MAX_OBJ_NAME_LEN		 = 4
)

const (
    AH_FAILURE_TRAP_TYPE AhTrapType = iota + 1
    AH_THRESHOLD_TRAP_TYPE
    AH_STATE_CHANGE_TRAP_TYPE
    AH_CONNECTION_CHANGE_TRAP_TYPE
    AH_IDP_AP_EVENT_TRAP_TYPE
    AH_CLIENT_INFO_TRAP_TYPE
    AH_POWER_INFO_TRAP_TYPE
    AH_CHANNEL_POWER_TRAP_TYPE
    AH_IDP_MITIGATE_TRAP_TYPE
    AH_INTERFERENCE_ALERT_TRAP_TYPE
    AH_BW_SENTINEL_TRAP_TYPE
    AH_ALARM_ALRT_TRAP_TYPE
    AH_MESH_MGT0_VLAN_CHANGE_TRAP_TYPE
    AH_KEY_FULL_ALARM_TRAP_TYPE
    AH_MESH_STABLE_STAGE_TRAP_TYPE
)

type AhFaMvlanChangeTrap struct {
	TrapType      uint8
	SystemID      [10]uint8
	NativeTagged uint8
	MgmtVlan      uint16
	NativeVlan    uint16
}

type AhTgrafDfsTrap struct {
	TrapType  uint8
	TrapId    uint8
//	DataLen   uint16
//	IfNameLen uint8
	IfName    [AH_MAX_TRAP_IF_NAME + 1]byte
//	DescLen   uint8
	Desc      [TRAP_DCRPT_LEN]byte
}

type AhFailureTrap struct {
    Name  [AH_MAX_TRAP_OBJ_NAME+1]byte
    Cause int32
    Set   int32
}

type AhThresholdTrap struct {
	Name           [AH_MAX_TRAP_OBJ_NAME+1]byte
	CurVal         int32
	ThresholdHigh  int32
	ThresholdLow   int32
}

type AhStateChangeTrap struct {
    Name          [AH_MAX_TRAP_OBJ_NAME+1]byte
    PreState      int32
    CurState      int32
    OperationMode int32
}

type AhConnectionChangeTrap struct {
	Name                   [AH_MAX_TRAP_OBJ_NAME + 1]byte
	Ssid                   [AH_MAX_TRAP_SSID_NAME + 1]byte
	HostName               [AH_MAX_TRAP_HOST_NAME + 1]byte
	UserName               [AH_MAX_TRAP_USER_NAME + 1]byte
	IfIndex                int32
	ObjectType             int32
	RemoteID               [6]byte
	BSSID                  [6]byte
	CurState               int32
	ClientIP               uint32
	ClientAuthMethod       int32
	ClientEncryptMethod    int32
	ClientMacProto         int32
	ClientVLAN             int32
	ClientUPID             int32
	ClientChannel          int32
	ClientCWPUsed          int32
	AssociationTime        uint32
	IfName                 [AH_MAX_TRAP_IF_NAME + 1]byte
	RSSI                   int32
	ProfName               [AH_MAX_TRAP_PROF_NAME + 1]byte
	SNR                    int32
	ClientMacBasedAuthUsed byte
	OS                     [AH_MAX_NAME_LEN + 1]byte
	Option55               [AH_UCHAR_MAX + 1]byte
	MgtStus                uint16
	StaAddr6Num            uint8
	_                      [3]byte  // Padding for 4-byte alignment
	StaAddr6               [AH_MAX_NUM_STA_ADDRS6][16]byte
	DeauthReason           int32
	RoamTime               int32
	AssocTime              int32
	AuthTime               int32
	RadioProf              [AH_MAX_NAME_LEN + 1]byte
	NegotiateKbps          uint32
}

type AhIdpApEventTrap struct {
	Name           [AH_MAX_TRAP_OBJ_NAME + 1]byte
	IfIndex        int32
	RemoteID       [6]byte
	IdpType        int32
	IdpChannel     int32
	IdpRSSI        int32
	IdpCompliance  int32
	SSID           [AH_MAX_TRAP_SSID_NAME + 1]byte
	StationType    int32
	StationData    int32
	IdpRemoved     int32
	IdpInNet       int32
}

type AhClientInfoTrap struct {
	Name         [AH_MAX_TRAP_OBJ_NAME + 1]byte
	Ssid         [AH_MAX_TRAP_SSID_NAME + 1]byte
	ClientMac    [6]byte
	HostName     [AH_MAX_TRAP_HOST_NAME + 1]byte
	UserName     [AH_MAX_TRAP_USER_NAME + 1]byte
	ClientIP     uint32
	MgtStus      uint16
	StaAddr6Num  uint8
	StaAddr6     [AH_MAX_NUM_STA_ADDRS6][16]byte
}

type AhPowerInfoTrap struct {
    Name         [AH_MAX_TRAP_OBJ_NAME + 1]byte
    PowerSrc     int
    Eth0On       int
    Eth1On       int
    Eth0Pwr      int
    Eth1Pwr      int
    Eth0Speed    int
    Eth1Speed    int
    Wifi0Setting int
    Wifi1Setting int
    Wifi2Setting int
}

type AhChannelPowerTrap struct {
    Name           [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex        int32
    RadioChannel   int32
    RadioTxPower   int32
    BeaconInterval uint32
    ChnlStrfmt     uint16
    PwrStrfmt      uint16
    RadioEirp      [8]byte
    Reason         int32
}

type AhIdpMitigateTrap struct {
    Name         [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex      int32
    RemoteID     [6]byte
    BSSID        [6]byte
    Removed      int32
    DiscoverAge  uint32
    UpdateAge    uint32
}

type AhInterferenceAlertTrap struct {
    Name                [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex             int32
    InterferenceThres   int32
    AveInterference     int32
    ShortInterference   int32
    SnapInterference    int32
    CRCErrRateThres     int32
    CRCErrRate          int32
    Set                 int32
}

type AhBwSentinelTrap struct {
    Name              [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex           int32
    ClientMac         [6]byte
    BwSentinelStatus  int32
    GBW               int32
    ActualBW          int32
    ActionTaken       uint32
    ChnlUtil          uint8
    InterferenceUtil  uint8
    TxUtil            uint8
    RxUtil            uint8
}

type AhAlarmAlrtTrap struct {
    Name               [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex            int32
    ClientMac          [6]byte
    Level              int32
    SSID               [AH_MAX_TRAP_SSID_NAME + 1]byte
    AlertType          int32
    ThresInterference  int32
    ShortInterference  int32
    SnapInterference   int32
    Set                int32
}

type AhMeshMgt0VlanChangeTrap struct {
	Name           [AH_MAX_TRAP_OBJ_NAME + 1]byte
	OldVlan        uint16
	NewVlan        uint16
	OldNativeVlan  uint16
	NewNativeVlan  uint16
}

type AhMeshStableStageTrap struct {
	Name            [AH_MAX_TRAP_OBJ_NAME + 1]byte
	MeshStableStage int32
	MeshDataRate    int32
}

type AhKeyFullAlarmTrap struct {
    Name      [AH_MAX_TRAP_OBJ_NAME + 1]byte
    IfIndex   int32
    BSSID     [MACADDR_LEN]byte
    ClientMAC [MACADDR_LEN]byte
    GtkVLAN   uint32
}

type AhTrapMsg struct {
     TrapType uint32
     Union    [816]byte
     Level    int32
     MsgID    int32
     Desc     [256]byte
}

func intToIPv4(num uint32) string {
	ip := net.IPv4(
		byte(num>>24),
		byte(num>>16),
		byte(num>>8),
		byte(num),
	)
	return ip.String()
}

func intToIPv6(addrs [][16]byte, count int) string {
	if count > 0 && count <= len(addrs) {
		ip := net.IP(addrs[0][:])
		ipStr := ip.String()
		return ipStr
	}
	return ""
}

func IntToIPv6_1(addrs []AhStaAddr6, count int) []string {
	var result []string
	for i := 0; i < count && i < len(addrs); i++ {
		ip := net.IP(addrs[i].StaAddr6[:])
		result = append(result, ip.String())
	}
	return result
}

type AhStaAddr6 struct {
    AddrType byte      // char addr_type
    StaAddr6 [16]byte  // struct in6_addr (IPv6 address is 16 bytes)
}