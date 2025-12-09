package ah_airrm

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

const (
	AH_USHORT_MAX  =		65535
	AH_IEEE80211_ACSP_MAX_NBRS = 384
	MAX_NEIGHBOR_NUM =			512
	IEEE80211_NWID_LEN =		32
	IEEE80211_MESHID_LEN =		32
	MACADDR_LEN = 6
	AH_IEEE80211_GET_AIRRM_TBL = 222
	AH_IEEE80211_GET_AIRRM_TBL_AP4020 = 223
	AH_IEEE80211_GET_AIRRM_TBL_AP5020 = 227
	AH_IEEE80211_GET_AIRRM_TBL_AP5000 = 216
	AH_IEEE80211_GET_AIRRM_TBL_AP3000 = 213
)

type ieee80211req_cfg_nbr struct{
		cmd		uint32
		resv	uint32			/* Let following struct to align to 64bit for 64 bit machine,
                                otherwise copy bgscan and lb status may with wrong value */
		buff	[126984]byte
}

type  iw_point struct
{
	pointer unsafe.Pointer
	length	uint16         /* number of fields or size in bytes */
	flags	uint16     /* Optional params */
}

type iwreq struct
{
	ifrn_name	[unix.IFNAMSIZ]byte    /* if name, e.g. "eth0" */
	data	iw_point
}


// AIRRM neighbor table structure
type ah_ieee80211_airrm_nbr_tbl_t struct {
	num_nbrs		uint32							// Number of neighbors
	_pad0			[4]byte							// Padding for 64bit alignment
	nbr_tbl			[MAX_NEIGHBOR_NUM + 1]ah_ieee80211_airrm_nbr_t		// Variable length array
}

// AIRRM neighbor information structure
type ah_ieee80211_airrm_nbr_t struct {
	rrmId					uint32			// unique identifier for RRM
	_pad0					[4]byte			// pad to align timestamp (was implicit)
	timestamp				uint64			// Date and timestamp
	extremeAP				uint8			// Is AP managed by Extreme?
	rssi					int8			// RSS value in dBm
	txPower                                 int8                    // Transmit power in dBm
	bssid					[MACADDR_LEN]byte		// BSSID
	radioMode				uint8			// Radio mode: 0=Access, 1=Sensor, 2=Backhaul, 3=Dual
    ssid					[32]byte		// SSID
    channelUtilization		uint8			// Channel utilization percentage
	interferenceUtilization	uint8			// Non-WIFI interference
	obssUtilization			uint8			// OBSS utilization
	wifiInterferenceUtilization	uint8		// WiFi interference utilization
	aggregationSize			uint16			// Aggregation size
	packetErrorRate			uint8			// Packet error rate
	_pad1					[1]byte			// pad for 2-byte alignment of clientCount
	clientCount				uint16			// Client count
	frequency				uint32			// Frequency in MHz
	channelWidth			uint16			// Channel width in MHz
	_pad2					[6]byte			// padding to make total size 80 bytes (multiple of 8)
}

