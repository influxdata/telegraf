//go:build AP4020 || AP5020
// +build AP4020 AP5020

package ah_trap

// AhConnectionChangeTrap for AP4020/AP5020 (with MLO support)
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
	_                      [3]byte // Padding for 4-byte alignment
	StaAddr6               [AH_MAX_NUM_STA_ADDRS6][16]byte
	DeauthReason           int32
	RoamTime               int32
	AssocTime              int32
	AuthTime               int32
	RadioProf              [AH_MAX_NAME_LEN + 1]byte
	NegotiateKbps          uint32
	// MLO fields (only for AP4020/AP5020)
	IsMloAssoc        uint8
	ClientMode        uint8
	Band              uint8
	StaMldAddr        [MACADDR_LEN]byte
	StaLinkAddr       [MACADDR_LEN]byte
	ApLinkAddr        [MACADDR_LEN]byte
	ClientAssocBitmap uint8
}

// AhTrapMsg for AP4020/AP5020 (larger union size to accommodate MLO fields)
type AhTrapMsg struct {
	TrapType uint32
	Union    [840]byte // Increased size to accommodate MLO fields (840 + safety margin)
	Level    int32
	MsgID    int32
	Desc     [256]byte
}
