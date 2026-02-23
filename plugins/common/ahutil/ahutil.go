package ahutil

import (
	"golang.org/x/sys/unix"
	"syscall"
	"os"
	"fmt"
	"strings"
	"bufio"
	"strconv"
	"log"
	"net"
	"os/exec"
	"bytes"
)

func Mystr() string {
	return "RF Stat Input Plugin Output"
}

const (
	AH_FE_DEV_NAME =		"/dev/fe"
	SIOCIWFIRSTPRIV =		0x8BE0
	SIOCGIWFREQ =			0x8B05
	SIOCGRADIOSTATS =		unix.SIOCDEVPRIVATE + 1
	IEEE80211_IOCTL_GETPARAM =	SIOCIWFIRSTPRIV + 1
	IEEE80211_RATE_MAXSIZE =	36
	WME_NUM_AC =			4
	VAP_BUFF_SIZE =			3090
	NS_HW_RATE_SIZE =		192
	AH_IEEE80211_ATR_MAX =		96
	AH_GET_STATION_NETWORK_HEALTH =	157
	AH_FLOW_GET_STATION_SERVER_IP =	200
	AH_FE_IOCTL_FLOW =		0x4000
	IEEE80211_IOCTL_GENERIC_PARAM = unix.SIOCDEVPRIVATE + 15
	IEEE80211_MESHID_LEN =		32
	AH_IEEE80211_STANAME_LEN =	16
	IEEE80211_PARAM_NUM_ASSOCS =	608
	MACADDR_LEN =			6
	AH_MAX_SSID_LEN =		32
	AH_TX_NSS_MAX =			4
	AH_SQ_GROUP_MAX =		8
	AH_MAX_DNS_LIST =		4
	RF_STAT_OUT_FILE =		"/tmp/rfStatOut"
	CLT_STAT_OUT_FILE =		"/tmp/clientStatOut"
	NW_STAT_OUT_FILE =		"/tmp/NetworkStatOut"
	DEV_STAT_OUT_FILE =		"/tmp/DeviceStatOut"
	EVT_SOCK =			"/tmp/ah_telegraf.sock"
	AH_MAX_ETH =			2
	AH_MAX_WLAN =			4
	ETH_IOCTL_FILE = 		"/dev/ah_ethif_ctl"
	AH_ETHIF_IOCTL_MAGIC =	'E'
	AH_SIOCCIFGETLINK = 	0x0009
	AH_SIOCCIFGETSTATUS =	0x0005
	NETWORK_MAX_COUNT =		30
	AH_MAX_RADIUS_NUM =		128
	AH_MAX_ACCESS_VIF_PER_RADIO = 15
	AH_MAX_LOG_SERVER =		4
)

const (
	AH_IEEE80211_CMD_NONE = iota
	AH_IEEE80211_SET_STA_VLAN          /* set station valn */
	AH_IEEE80211_GET_STA_VLAN          /* get station valn */
	AH_IEEE80211_SET_DOSCFG            /* set dos config */
	AH_IEEE80211_GET_DOSCFG            /* get dos config */
	AH_IEEE80211_FF_ADD_CFG            /* add ff */
	AH_IEEE80211_FF_DEL_CFG            /* del ff */
	AH_IEEE80211_FF_GET_CFG            /* get ff */
	AH_IEEE80211_FF_GET_CFG_BYID       /* get ff by id */
	AH_IEEE80211_FF_GET_CNT            /* get ff cnt */
	AH_IEEE80211_GET_LB_STATUS
	AH_IEEE80211_GET_LB_DATA
	AH_IEEE80211_CLR_STA_STATS
	AH_IEEE80211_CLR_80211_STATS
	AH_IEEE80211_SET_STA_UPID
	AH_IEEE80211_GET_STA_UPID
	AH_IEEE80211_SET_NODE_ID
	AH_IEEE80211_GET_ACSPIE
	AH_IEEE80211_GET_ACSPNBR_TBL
	AH_IEEE80211_QUERY_ACSPNBR
	AH_IEEE80211_SET_BGSCAN_CFG
	AH_IEEE80211_GET_BGSCAN_CFG
	AH_IEEE80211_GET_BGSCAN_STATS
	AH_IEEE80211_GET_IDP_AP_TBL
	AH_IEEE80211_SET_IDP_CFG
	AH_IEEE80211_GET_IDP_CFG
	AH_IEEE80211_GET_ALL_STA
	AH_IEEE80211_GET_ONE_STA
	AH_IEEE80211_GET_LB_STATS
	AH_IEEE80211_DEAUTH_STA
	AH_IEEE80211_SEND_PROBEREQ
	AH_IEEE80211_SET_STANAME
	AH_IEEE80211_START_IDP_IND         /* start IDP in-network detection */
	AH_IEEE80211_SET_AUTHMODE
	AH_IEEE80211_SEND_MICFAIL
	AH_IEEE80211_GET_NBR_RATE
	AH_IEEE80211_GET_ONE_NBR_RATE
	AH_IEEE80211_SET_IDP_AP_CAT        /* set idp ap category */
	AH_IEEE80211_SET_ADDI_AUTH_FLAG
	AH_IEEE80211_GET_ONE_STA_INFO
	AH_IEEE80211_SET_CONTX_MODE        /* enable continuous tx */
	AH_IEEE80211_CLR_CONTX_MODE        /* disable continuous tx */
	AH_IEEE80211_GET_NBR_AVGRSSI
	AH_IEEE80211_SET_TX_RATESET        /* set tx rate set */
	AH_IEEE80211_GET_TX_RATESET        /* get tx rate set */
	AH_IEEE80211_SET_RX_ONLY           /* enable rx only test */
	AH_IEEE80211_CLR_RX_ONLY           /* disable rx only test */
	AH_IEEE80211_SET_LCS_TAG           /* start/stop tag report */
	AH_IEEE80211_SET_LCS_MU            /* start/stop mu report */
	AH_IEEE80211_GET_LCS_TAG_STATS     /* get lcs tag statistics */
	AH_IEEE80211_CLR_LCS_TAG_STATS     /* clear lcs tag statistics */
	AH_IEEE80211_GET_LCS_MISC_STATS    /* get lcs misc statistics */
	AH_IEEE80211_CLR_LCS_MISC_STATS    /* clear lcs misc statistics */
	AH_IEEE80211_SET_LCS_RATE_LIMIT    /* set lcs rate limit threshold */
	AH_IEEE80211_GET_LCS_RATE_LIMIT    /* set lcs rate limit threshold */
	AH_IEEE80211_SET_PKT_CPT           /* set pkt capture parameters */
	AH_IEEE80211_GET_PKT_CPT_STATS     /* get pkt capture related stats */
	AH_IEEE80211_GET_AIRTIME_STATS
	AH_IEEE80211_SET_DISTANCE          /* set radio operating distance in meter */
	AH_IEEE80211_GET_DISTANCE          /* get radio operating distance in meter */
	AH_IEEE80211_SET_ANT_PORT          /* set antenna port to be used for tx/rx */
	AH_IEEE80211_GET_ANT_PORT          /* get antenna port to be used for tx/rx */
	AH_IEEE80211_GET_ATR_TBL
	AH_IEEE80211_SEND_MGMT_FRAME       /* asking driver to send a mgmt frame */
	AH_IEEE80211_EXEC_MITIGATE_AP      /* exec mitigate action on rogue AP */
	AH_IEEE80211_GET_IDP_STA_TBL       /* get clients connect to rogue AP */
	AH_IEEE80211_CLR_ROGUE_AP          /* clear rogue AP(s) */
	AH_IEEE80211_SET_LTA_TRACK_OBJ     /* set lta track object */
	AH_IEEE80211_GET_LTA_RPT_TBL       /* get lta station report table */
	AH_IEEE80211_GET_LTA_STATS         /* get lta statistics */
	AH_IEEE80211_CLR_LTA_STATS         /* clear lta statistics */
	AH_IEEE80211_GET_CCA               /* get cca stats, don't use later */
	AH_IEEE80211_SET_MESH_JOIN_REQ     /* set mesh join req */
	AH_IEEE80211_SET_ASM_BEHAVIOR      /* set ASM behavior criteria */
	AH_IEEE80211_SET_ASM_ACTION        /* set ASM action criteria */
	AH_IEEE80211_GET_CONN_BNCHMRK      /* get client connectivity benchmark score */
	AH_IEEE80211_SET_CONN_BNCHMRK      /* set client connectivity benchmark score */
	AH_IEEE80211_SET_STA_IDLE_TIME     /* set client idle_out time, in second*/
	AH_IEEE80211_GET_STA_IDLE_TIME     /* get client idle_out time, in second*/
	IEEE80211_GET_WIFI_STA_STATS       /* get all clients' stats for one radio */
	AH_IEEE80211_SET_HDD_CFG           /* get hdd config */
	AH_IEEE80211_GET_HDD_CFG           /* set hdd config */
	AH_IEEE80211_GET_HDD_STATS         /* get hdd statistics */
	AH_IEEE80211_CLR_HDD_STATS         /* clear hdd statistics */
	AH_IEEE80211_GET_HDD_RAIDO_LOAD    /* get hdd radio load  */
	AH_IEEE80211_GET_HDD_STA_RLS       /* get client local release flag */
	AH_IEEE80211_ADD_HDD_5G_ENTRY      /* add an entry to 5g client list (test command) */
	AH_IEEE80211_CLR_HDD_5G_TBL        /* clear hdd 5G capable client table node(s) */
	AH_IEEE80211_CLR_HDD_SUP_TBL       /* clear hdd suppress table node(s) */
	AH_IEEE80211_GET_HDD_SUP_TBL       /* get hdd suppress table */
	AH_IEEE80211_GET_HDD_NT            /* get hdd table HDD-2 */
	AH_IEEE80211_ADD_HDD_BPS_OUI       /* add hdd bcast probe suppression OUI entry */
	AH_IEEE80211_DEL_HDD_BPS_OUI       /* delete hdd bcast probe suppression OUI entry */
	AH_IEEE80211_GET_HDD_BPS_OUI_COUNT /* get count of hdd bcast probe suppression OUI entries */
	AH_IEEE80211_GET_HDD_BPS_OUI       /* get hdd bcast probe suppression OUI entries */
	AH_IEEE80211_GET_CYCLE_COUNTS      /* get cycle counts */
	AH_IEEE80211_GET_RD_MAX_TX_PWR     /* get RF regulatory domain max tx power limit */
	AH_IEEE80211_SET_CHANNEL_COST      /* set channel cost */
	AH_IEEE80211_SET_EXCLUDED_CHANNELS /* set list of excluded channels (ACSP) */
	AH_IEEE80211_SET_SPECTRAL_CHAN     /* set spectral scan channel */
	AH_IEEE80211_SET_SPECTRAL_ACTION   /* start/stop spectral scan */
	AH_IEEE80211_SET_MCAST_UCAST_CONV_MODE    /* Configure multicast to unicast conversion mode (DISABLED/ALWAYS/AUTO) */
	AH_IEEE80211_GET_MCAST_UCAST_CONV_MODE    /* Get current value of multicast to unicast conversion mode */
	AH_IEEE80211_SET_MCAST_CONV_CU_THRESH     /* Set Multicast conversion CU Threshold */
	AH_IEEE80211_GET_MCAST_CONV_CU_THRESH     /* Get Multicast conversion CU Threshold */
	AH_IEEE80211_SET_MCAST_CONV_MEMBER_THRESH /* Set Multicast conversion Membership Count Threshold */
	AH_IEEE80211_GET_MCAST_CONV_MEMBER_THRESH /* Get Multicast conversion Membership Count Threshold */
	AH_IEEE80211_GET_MCAST_CONV_STATS_SIZE    /* Get Multicast Group Statistics structure size */
	AH_IEEE80211_GET_MCAST_CONV_STATS         /* Get Multicast Group Statistics */
	AH_IEEE80211_CLR_MCAST_CONV_STATS         /* Clear Multicast Group Statistics */
	AH_IEEE80211_SEND_ANT_ALIGN_REQ_FRM
	AH_IEEE80211_GET_ANT_ALIGN_RSP_DATA
	AH_IEEE80211_SET_IDP_STA_CAT
	AH_IEEE80211_SET_RRM_QUIET_PARA         /* RRM QUIET PARA */
	AH_IEEE80211_GET_ADMCTL_TSINFO          /* Get Admctl Tsinfo */
	AH_IEEE80211_GET_ADMCTL_TSINFO_LIST     /* Get Admctl Tsinfo LIST*/
	AH_IEEE80211_SET_ADMCTL_BW
	AH_IEEE80211_SET_MDIE                   /* Set MDIE PARAM */
	AH_IEEE80211_SET_PRESENCE_CFG
	AH_IEEE80211_GET_PRESENCE_CFG
	AH_IEEE80211_GET_PRESENCE_DATA
	AH_IEEE80211_SET_SENSOR_CFG
	AH_IEEE80211_SET_VHT_MCS_MAP       /* set mcs map in VHT capability IE */
	AH_IEEE80211_GET_VHT_MCS_MAP       /* get mcs map in VHT capability IE */
	AH_IEEE80211_SET_ACSP_NODE_IPV6
//#ifdef AH_RADIO_BCM
	AH_IEEE80211_SET_BRCM_MSGLEVEL
	AH_IEEE80211_GET_BRCM_MSGLEVEL
	AH_IEEE80211_SET_BRCM_EMSGLEVEL
	AH_IEEE80211_GET_BRCM_EMSGLEVEL
	AH_IEEE80211_SET_BRCM_PHYMSGLEVEL
	AH_IEEE80211_GET_BRCM_PHYMSGLEVEL
	AH_IEEE80211_SET_AWEMSGLEVEL
	AH_IEEE80211_GET_AWEMSGLEVEL
	AH_IEEE80211_SET_IEMMSGLEVEL
	AH_IEEE80211_GET_IEMMSGLEVEL
	AH_IEEE80211_SET_TXBF_MODE
	AH_IEEE80211_GET_TXBF_MODE
	AH_IEEE80211_SET_VHT2G_MODE
	AH_IEEE80211_GET_VHT2G_MODE
//#ifdef AH_SUPPORT_ANTENNA_TYPE
	AH_IEEE80211_SET_ANTENNA_TYPE
//#endif
	AH_IEEE80211_DBGREQ_SENDBCNRPT
	AH_IEEE80211_DBGREQ_SENDTSMRPT
	AH_IEEE80211_DBGREQ_SENDDELTS
	AH_IEEE80211_SET_TPC
	AH_IEEE80211_DBGREQ_SENDBSTMREQ /* WNM Operation */
//#endif
//#ifdef AH_SUPPORT_DOS
	AH_IEEE80211_SET_DOS_EXT_ENABLE
	AH_IEEE80211_GET_DOS_EXT_ENABLE
	AH_IEEE80211_ADD_DOS_EXT_BLOCK
	AH_IEEE80211_DEL_DOS_EXT_BLOCK
	AH_IEEE80211_GET_DOS_EXT_USRLIST
	AH_IEEE80211_GET_DOS_EXT_BANLIST
//#endif
//#ifdef AH_SUPPORT_MUMIMO
	AH_IEEE80211_GET_MUMIMO_COUNTERS
	AH_IEEE80211_CLEAR_MUMIMO_COUNTERS
	AH_IEEE80211_GET_MUMIMO_CANDIDATES
	AH_IEEE80211_SET_MUTX
	AH_IEEE80211_SET_MU_FEATURES
	AH_IEEE80211_SET_MUMIMO_STA_RXCHAINS
	AH_IEEE80211_SET_MU_SOUNDING_INTERVAL
//#endif
//#ifdef AH_ZWDFS_SUPPORT
	AH_IEEE80211_SET_ZEROWAIT_DFS
//#ifdef AH_ZWDFS_CFG_SUPPORT
	AH_IEEE80211_SET_ZWDFS_CFG
//#endif
//#endif
//#ifdef AH_AIRIQ_SUPPORT
	AH_IEEE80211_SET_SPECTRAL_INTERVAL   /* set spectral scan interval */
	AH_IEEE80211_SET_SPECTRAL_AIRIQ_MODE /* set spectral airiq mode */
	AH_IEEE80211_GET_SPECTRAL_FREQ_INFO  /* get spectral freq info */
	AH_IEEE80211_GET_OPERATE_TX_CHAIN    /* get operate Tx chain */
	AH_IEEE80211_GET_OPERATE_RX_CHAIN    /* get operate Rx chain */
//#endif
	AH_IEEE80211_GET_RADIO_MODE
//#ifdef AH_SUPPORT_SDR
	AH_IEEE80211_GET_SDR_NBR_TBL
//#endif
//#ifdef AH_SUPPORT_DYNAMIC_CHANNEL_WIDTH
	AH_IEEE80211_SET_DYNAMIC_CHANNEL_WIDTH
	AH_IEEE80211_SET_DYNAMIC_CHANNEL_WIDTH_THRESH
	AH_IEEE80211_GET_DYNAMIC_CHANNEL_WIDTH_HISTORY
	AH_IEEE80211_CLEAR_DYNAMIC_CHANNEL_WIDTH_HISTORY
//#endif
	AH_IEEE80211_SET_RADARARGS           /* set PHY radar parameters */
	AH_IEEE80211_SET_RADARTHR            /* set PHY radar threshhold parameters */
	AH_IEEE80211_INCR_DFS_SENSE_RANGE    /* increase DFS sence range */
	AH_IEEE80211_RESET_DFS_SENSE_RANGE   /* reset DFS sense range */
	AH_IEEE80211_SET_ACSP_ACTIVE         /* set acsp active flag */
	AH_IEEE80211_GET_POWER_LIMIT         /* get hw power limit */
	AH_IEEE80211_GET_NBR_MAX_RATE        /* get nbr max tx rate in Kbps */
	AH_IEEE80211_GET_ONE_ACSP_NBR
	AH_IEEE80211_GET_CHNL_SCAN_STATS     /* get channel scan results(cca/chanim stats) */
//#ifdef AH_SUPPORT_HOSTNAME_IN_BCN
	AH_IEEE80211_GET_HOSTNAME_IN_BCN_ENABLE    /* get hostname-in-beacon enable */
	AH_IEEE80211_SET_HOSTNAME_IN_BCN_ENABLE    /* enable/disable hostname in beacon */
//#endif
//#ifdef AH_SUPPORT_RX_SENSITIVITY
	AH_IEEE80211_SET_ED_THRESHOLD        /* set energy detection threshold */
	AH_IEEE80211_SET_RXDESENS            /* set rx sensitivity */
//#endif
//#ifdef AH_SUPPORT_PER_VLAN_GTK
	AH_IEEE80211_GET_ALL_STA_VLAN_GTK
//#endif
	AH_IEEE80211_SET_RRM_NBR_LIST_DUAL_BAND
	AH_IEEE80211_SET_RRM_NBR_LIST_BAND_MAX
	AH_IEEE80211_SET_BSS_TRANS_DISASSOC_IMMT
	AH_IEEE80211_SET_BSS_TRANS_DISASSOC_IMMT_TIMER
	AH_IEEE80211_SET_11KV_FAILED_ACTIONS
//#ifdef AH_NETWORK360_CLT_CAPS
	AH_IEEE80211_CLT_CAPS_DEL_STA
//#endif
//#ifdef AH_NETWORK360_WIFI_STATS
	AH_IEEE80211_SQ_GROUP_RANGE
//#endif
//#ifdef AH_SUPPORT_11AX
	AH_IEEE80211_SET_HE_MCS_MAP       /* set he mcs map */
	AH_IEEE80211_GET_HE_MCS_MAP       /* get he mcs map */
	AH_IEEE80211_SET_HE_OFDMA_DL		/* configure downlink OFDMA */
	AH_IEEE80211_GET_HE_OFDMA_DL		/* get the status of downlink OFDMA */
//#ifdef AH_SUPPORT_11AX_FEATURES
	AH_IEEE80211_SET_HE_OFDMA_UL		/* configure uplink OFDMA */
	AH_IEEE80211_GET_HE_OFDMA_UL		/* get the status of uplink OFDMA */
	AH_IEEE80211_SET_HE_BSS_COLOR		/* configure BSS color */
	AH_IEEE80211_GET_HE_BSS_COLOR		/* get the status of BSS color */
	AH_IEEE80211_SET_TWT		        /* configure TWT */
	AH_IEEE80211_GET_TWT		        /* get the status of TWT */
//#endif
//#endif
	AH_IEEE80211_GET_CHNL_STATS     /* get channel util, tx util and rx util etc */
	AH_IEEE80211_GET_TXCORE         /* get txcore of radio */
	AH_IEEE80211_GET_HW_PWR_LIMIT   /* get hw power limit of radio */
	AH_IEEE80211_GET_DFS_STATE      /* get DFS state of radio */
//#ifdef AH_SUPPORT_ADSP_SENSOR
	AH_IEEE80211_SET_ADSP_SENSOR_CFG
	AH_IEEE80211_GET_ADSP_SENSOR_CFG
	AH_IEEE80211_SET_WIPSK_CHANSPEC
//#endif
	AH_IEEE80211_SEND_CSA		/* send channel switch announcement */
	AH_IEEE80211_GET_IF_COUNTRY           /* get interface brcm cc */
	AH_IEEE80211_SET_RSNXE_CAP		/* set RSN Extended IE capabilities*/
//#ifdef AH_SUPPORT_AFC
	AH_IEEE80211_GET_AFC_INFO	/* get AFC info */
//#endif
//#ifdef AH_SUPPORT_11MC
    AH_IEEE80211_SET_FTM           /* set FTM and LCI */
//#endif
//#if AH_SUPPORT_USEG
	AH_IEEE80211_SET_USEG_BCAST_FILTER   /* set PCG broadcast filter */
	AH_IEEE80211_SET_USEG_MCAST_FILTER   /* set PCG multicast filter */
//#endif
)

func Ah_ioctl(fd uintptr, op, argp uintptr) error {
	        _, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(op), argp)
        if errno != 0 {
                return errno
        }
        return nil
}

func Ah_ifname_radio2vap(radio_name string ) string {
    switch {
    case radio_name == "wifi0":
        return "wifi0.1"
    case radio_name == "wifi1":
        return "wifi1.1"
    case radio_name == "wifi2":
        return "wifi2.1"
    default:
        return "invalid"
    }
}

/*
 * Convert MHz frequency to IEEE channel number.
 */
func FreqToChan(freq uint16) uint16 {
		if freq < 2412 {
			return 0
		}
		if freq > 7125 {
			return 0
		}
		if freq == 2484 {
			return 14
		}
		if freq < 2484 {
			return (freq-2407)/5
		}
		if freq < 5000 {
			return 15+((freq-2512)/20)
		}
		if freq < 5935 {
			return (freq - 5000)/5
		}
		if freq == 5935 {
			return 2
		}
		return (freq -5950)/5
}

func GetAPName() string {
    file, err := os.Open("/f/system_info/hw_info")
    if err != nil {
        return "NA"
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if strings.Contains(scanner.Text(), "Product name:") {
            pl := scanner.Text()
            fields := strings.Fields(pl)
            APN := fields[2][:6]
            return APN
        }
    }

    if err := scanner.Err(); err != nil {
        return "NA"
    }
    return "NA"
}

func GetRrmId() int {
	var ret int
	ret = 0
	content, err := os.ReadFile("/tmp/rrmid")
	if err != nil {
		return 0
	}
	ret, err = strconv.Atoi(string(content))
	if err != nil {
		return 0
	}
	return ret
}

func Check_Vap_Status(ifname string) bool {
	vapName := Ah_ifname_radio2vap(ifname)
	if vapName == "invalid" {
		return false
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("ahutil: net.Interfaces failed: %v", err)
		return false
	}
	for _, iface := range ifaces {
		if iface.Name == vapName {
			up := iface.Flags&net.FlagUp != 0
			return up
		}
	}

	return false
}

// Utility function to get tx power using 'wl -i ifname txpwr' command
func GetTxPower(ifname string) int8 {

	// Validation: check if ifname is not blank and contains "wifi"
	if ifname == "" || !strings.Contains(ifname, "wifi") {
		return -1
	}

	app := "wl"

	arg0 := "-i"
	arg1 := ifname
	arg2 := "txpwr"

	cmd := exec.Command(app, arg0, arg1, arg2)
	output, err := cmd.Output()
	if err != nil {
		return -1
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "dBm") && strings.Contains(line, "mw") {
			// Example line: "18.0 dBm = 63 mw." or "0.0 dBm = 1 mw"
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "dBm" && i > 0 {
					// Get the dBm value (before "dBm")
					dBmStr := fields[i-1]
					// Parse as float first to handle decimal values like "18.0"
					dBmFloat, err := strconv.ParseFloat(dBmStr, 64)
					if err == nil {
						return int8(dBmFloat)
					}
				}
			}
		}
	}
	return -1
}

func MacToString(mac [6]byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",mac[0],mac[1],mac[2],mac[3],mac[4],mac[5])
}

func GetBSSIDbyIfName(ifName string) (bssid [MACADDR_LEN]uint8, err error) {

	var ret [MACADDR_LEN]uint8
	app := "wl"
	arg0 := "-i"
	arg1 := ifName
	arg2 := "bssid"

	cmd := exec.Command(app, arg0, arg1, arg2)
	out, err := cmd.Output()
	if err != nil {
		return ret, err
	}

	bssid_str := string(out)

	bssid_str = strings.TrimSpace(bssid_str)
	split_text := strings.Split(bssid_str, ":")

	for i := 0; i < MACADDR_LEN; i++ {
		val,err := strconv.ParseUint(split_text[i], 16, 8)
		if err != nil {
			return ret, err
		}
		ret[i] = uint8(val)
	}

	return ret, nil
}

func GetSSIDbyIfName(ifName string) (ssid string, err error) {

	app := "wl"
	arg0 := "-i"
	arg1 := ifName
	arg2 := "ssid"

	cmd := exec.Command(app, arg0, arg1, arg2)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outstring := string(out)

	ssid_text := strings.Split(outstring, ":")
	ssid_ret := strings.TrimSpace(ssid_text[1])[1 : len(strings.TrimSpace(ssid_text[1]))-1]

	return ssid_ret, nil
}

func IntToIpv4(num uint32) string {

    b := make([]byte, 4)
    b[0] = byte(num)
    b[1] = byte(num >> 8)
    b[2] = byte(num >> 16)
    b[3] = byte(num >> 24)
    return fmt.Sprintf("%d.%d.%d.%d",b[0],b[1],b[2],b[3])
}


func CleanCString(b []byte) string {
    if i := bytes.IndexByte(b, 0); i != -1 {
        return string(b[:i])
    }
    return string(b)
}


func FormatMac(mac [6]byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}