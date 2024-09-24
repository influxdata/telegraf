package ah_wireless

import (
        "golang.org/x/sys/unix"
	"unsafe"
	"fmt"
)

const (
	AH_FE_DEV_NAME =		"/dev/fe"
	SIOCIWFIRSTPRIV =		0x8BE0
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
)

const (
        AH_SQ_TYPE_RSSI = iota
        AH_SQ_TYPE_NOISE
        AH_SQ_TYPE_SNR
        AH_SQ_TYPE_MAX
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

const (
	IEEE80211_MODE_AUTO             = iota   /* autoselect */
	IEEE80211_MODE_11A             /* 5GHz, OFDM */
	IEEE80211_MODE_11B              /* 2GHz, CCK */
	IEEE80211_MODE_11G              /* 2GHz, OFDM */
	IEEE80211_MODE_FH               /* 2GHz, GFSK */
	IEEE80211_MODE_TURBO_A             /* 5GHz, OFDM, 2x clock dynamic turbo */
	IEEE80211_MODE_TURBO_G            /* 2GHz, OFDM, 2x clock  dynamic turbo*/
	IEEE80211_MODE_11NA_HT20      /* 5Ghz, HT20 */
	IEEE80211_MODE_11NG_HT20         /* 2Ghz, HT20 */
	IEEE80211_MODE_11NA_HT40PLUS      /* 5Ghz, HT40 (ext ch +1) */
	IEEE80211_MODE_11NA_HT40MINUS     /* 5Ghz, HT40 (ext ch -1) */
	IEEE80211_MODE_11NG_HT40PLUS    /* 2Ghz, HT40 (ext ch +1) */
	IEEE80211_MODE_11NG_HT40MINUS   /* 2Ghz, HT40 (ext ch -1) */
	IEEE80211_MODE_11NG_HT40     /* 2Ghz, Auto HT40 */
	IEEE80211_MODE_11NA_HT40  /* 2Ghz, Auto HT40 */
	IEEE80211_MODE_11AC_VHT20     /* 5Ghz, VHT20 */
	IEEE80211_MODE_11AC_VHT40PLUS    /* 5Ghz, VHT40 (Ext ch +1) */
	IEEE80211_MODE_11AC_VHT40MINUS    /* 5Ghz  VHT40 (Ext ch -1) */
	IEEE80211_MODE_11AC_VHT40  /* 5Ghz, VHT40 */
	IEEE80211_MODE_11AC_VHT80   /* 5Ghz, VHT80 */
	IEEE80211_MODE_11AC_VHT160   /* 5Ghz, VHT160 */
	IEEE80211_MODE_11AX_2G_HE20     /* 2Ghz, HE20 */
	IEEE80211_MODE_11AX_2G_HE40    /* 2Ghz, HE40 */

	IEEE80211_MODE_11AX_5G_HE20     /* 5Ghz, HE20 */
	IEEE80211_MODE_11AX_5G_HE40    /* 5Ghz, HE40 */
	IEEE80211_MODE_11AX_5G_HE80     /* 5Ghz, HE80 */
	IEEE80211_MODE_11AX_5G_HE160    /* 5Ghz, HE160 */

	IEEE80211_MODE_11AX_6G_HE20    /* 6Ghz, HE20 */
	IEEE80211_MODE_11AX_6G_HE40  /* 6Ghz, HE40 */
    IEEE80211_MODE_11AX_6G_HE80  /* 6Ghz, HE80 */
	IEEE80211_MODE_11AX_6G_HE160  /* 6Ghz, HE160 */
	IEEE80211_MODE_LAST
)

const (
    AH_DCD_NMS_PHY_MODE_A = iota
    AH_DCD_NMS_PHY_MODE_B
    AH_DCD_NMS_PHY_MODE_G
    AH_DCD_NMS_PHY_MODE_AC  = iota + 2
    AH_DCD_NMS_PHY_MODE_NA
    AH_DCD_NMS_PHY_MODE_NG
    AH_DCD_NMS_PHY_MODE_AX_2G
    AH_DCD_NMS_PHY_MODE_AX_5G
    AH_DCD_NMS_PHY_MODE_AX_6G
)


type IFReqData struct {
        Name [unix.IFNAMSIZ]byte
        Data uintptr
}

type  ah_ieee80211_atr_info struct{
        rxc_pcnt	int32
        rxf_pcnt	int32
	rxf_obss	int32
	rxf_inbss	int32
	rxf_nocfg	int32
        txf_pcnt	int32
        nfarray		[2]int16
        valid		int32   /* mark this bucket as either valid or invalid */
}


type  ah_ieee80211_atr_user struct{
        count		uint32   /* air-time ring buffer count */
        atr_info	[AH_IEEE80211_ATR_MAX]ah_ieee80211_atr_info
}

type ah_ieee80211_hdd_stats struct {
        bs_sp_cnt		uint32      /* band steering suppress count */
        lb_sp_cnt		uint32      /* load balance suppress count */
        snr_sp_cnt		uint32     /* weak snr suppress count */
        sn_answer_cnt	uint32  /* safety net answer (safety net check fail) count */
}

type ieee80211req_cfg_atr struct{
        cmd		uint32
        resv	uint32			/* Let following struct to align to 64bit for 64 bit machine,
									otherwise copy bgscan and lb status may with wrong value */
	atr ah_ieee80211_atr_user
	unused uint32
}

type ieee80211req_cfg_hdd struct{
		cmd		uint32
		resv	uint32			/* Let following struct to align to 64bit for 64 bit machine,
                                otherwise copy bgscan and lb status may with wrong value */
		hdd_stats ah_ieee80211_hdd_stats
		unused uint32
}

type  iw_point struct
{
//	pointer uintptr       /* Pointer to the data  (in user space) */
	pointer unsafe.Pointer
	length	uint16         /* number of fields or size in bytes */
	flags	uint16     /* Optional params */
}

type iwreq struct
{
	ifrn_name	[unix.IFNAMSIZ]byte    /* if name, e.g. "eth0" */
	data	iw_point
}

type ah_dcd_dev_stats struct{
        rx_packets uint64  /* total packets received       */
        tx_packets uint64  /* total packets transmitted    */
        rx_bytes uint64    /* total bytes received         */
        tx_bytes uint64    /* total bytes transmitted      */
        rx_errors uint32         /* bad packets received         */
        tx_errors uint32         /* packet transmit problems     */
        rx_dropped uint32        /* no space in linux buffers    */
        tx_dropped uint32        /* no space available in linux  */
        rx_multicast uint32      /* multicast packets received   */
        rx_compressed uint32
        tx_compressed uint32
        collisions uint32

        rx_unicast uint32
        rx_broadcast uint32
        tx_unicast uint32
        tx_broadcast uint32
        tx_multicast uint32


        /* detailed rx_errors: */
        rx_length_errors uint32
        rx_over_errors uint32    /* receiver ring buff overflow  */
        rx_crc_errors uint32     /* recved pkt with crc error    */
        rx_frame_errors uint32   /* recv'd frame alignment error */
        rx_fifo_errors uint32     /* recv'r fifo overrun          */
        rx_missed_errors uint32   /* receiver missed packet       */
        /* detailed tx_errors */
        tx_aborted_errors uint32
        tx_carrier_errors uint32
        tx_fifo_errors uint32
        tx_heartbeat_errors uint32
        tx_window_errors uint32
}

type ieee80211_node_rate_stats struct {
	ns_unicasts			uint32				/* tx/rx total unicasts */
	ns_retries			uint32				/* tx/rx total retries */
	ns_rateKbps			uint32				/* rate in Kpbs */
}

type utilization_data struct {
	intfer_util_min		int32
	intfer_util_max		int32
	intfer_util_avg		int32

	chan_util_min		int32
	chan_util_max		int32
	chan_util_avg		int32

	tx_util_min			int32
	tx_util_max			int32
	tx_util_avg			int32

	rx_util_min			int32
	rx_util_max			int32
	rx_util_avg			int32

	rx_ibss_util_min	int32
	rx_ibss_util_max	int32
	rx_ibss_util_avg	int32

	rx_obss_util_min	int32
	rx_obss_util_max	int32
	rx_obss_util_avg	int32

	wifi_i_util_min		int32
	wifi_i_util_max		int32
	wifi_i_util_avg		int32

	noise_min			int16
	noise_max			int16
	noise_avg			int16

	crc_err_rate_min	uint64
	crc_err_rate_max	uint64
	crc_err_rate_avg	uint64
}

type awestats struct {
	ast_as				wl_stats
	ast_uapsdqnuldepth		uint32				/* count of UAPSD QoS NULL frames in uapsd queue */
	ast_uapsdresetdrop_pkts		uint32				/* count of UAPSD frames dropped because of chip reset */
	ast_tx_xtxop			uint32				/* tx failed due to exceeding txop */
	ast_tx_xtimer			uint32				/* tx failed due to tx timer expired */
	ast_cwm_mac40to20		uint32				/* dynamic channel width change from HT40 to HT20 */
	ast_cwm_mac20to40		uint32				/* dynamic channel width change from H */
	phy_stats			wl_phy_stats			/* phy stats */
	ast_rx_mgt			uint32				/* management frames received */
	ast_rx_ctl			uint32				/* control frames received */
	ast_rx_flush			uint32				/* packet flush */
	ast_rx_bufsize			uint32				/* wrong buffer size */
	ast_rx_keyix_errors		uint32				/* keyix erros by Atheros chip */
	ast_rx_antenna_errors		uint32				/* antenaa errors by Atheros chip */
	ast_rx_short_frames		uint32				/* frames too short */
	ast_rx_ieee_queue_depth		uint32				/* per radio rx ieee80211 queue depth */
	ast_tx_failed			uint32				/* ath_tx_start failed */
	ast_tx_ps_full			uint32				/* power save queue full */
	ast_tx_buf_count		uint32				/* available tx buffer */
	ast_buf_wo_count		uint32				/* # of times buffer workaround counter kicked in */
	ast_tx_buf_max_count		uint32				/* max tx buffer */
	ast_tx_hw_othererr		uint32				/* other hw tx error */
	ast_rx_bcast			uint32				/* rx broadcast data frame */
	ast_rx_mcast			uint32				/* rx multicast data frame */
	ast_tx_wme_ac			[WME_NUM_AC]uint32		/* tx WME AC data frame */
//#ifdef AH_SUPPORT_MUMIMO
	ast_tx_wme_ac_mu		[WME_NUM_AC]uint32		/* tx WME AC MU-MIMO data frame */
//#endif
	ast_tx_drain			uint32				/* tx drain data frame */
	ast_tx_blckd_drops		uint32				/* tx drop due to tx blocked to a station */
	ast_tx_captured			uint32				/* captured total num of tx frames */
	ast_rx_captured			uint32				/* captured total num of rx frames */
	ast_tx_queue_scheds		uint32				/* txq schedules because of congestion */
	ast_beacon_stucks		uint32				/* number of beacon stucks */
	ast_pci_errors			uint32				/* number of PCI read/write timeout errors */
	ast_tx_rix_invalids		uint32				/* tx rate index invalids */
	ast_rx_rix_invalids		uint32				/* tx rate index invalids */
	ast_bandwidth			uint32				/* total bandwidth: rx + tx */

	ast_rx_rate_stats		[NS_HW_RATE_SIZE]ieee80211_node_rate_stats
	ast_tx_rate_stats		[NS_HW_RATE_SIZE]ieee80211_node_rate_stats

	ast_tx_airtime			uint64				/* tx airtime (us) */
	ast_rx_airtime			uint64				/* rx airtime (us) */
	ast_crcerr_airtime		uint64				/* crc eror airtime (us) */
	ast_noise_floor			int16
	ast_rx_mcast_bytes		uint64
	ast_rx_bcast_bytes		uint64
	ast_rx_retry			uint32

	is_rx_hdd_probe_sup		uint32
	is_rx_hdd_auth_sup		uint32

	magic					uint32
}

type wl_stats struct {
	ast_rx_bytes			uint32				/* total number of bytes received */
	ast_tx_bytes			uint32				/* total number of bytes transmitted */

	ast_tx_packets			uint32				/* packet sent on the interface */
	ast_rx_packets			uint32				/* packet received on the interface */
	ast_tx_mgmt				uint32				/* management frames transmitted */
	ast_tx_discard			uint32				/* frames discarded prior to assoc */
	ast_tx_invalid			uint32				/* frames discarded 'cuz device gone */
	ast_tx_qstop			uint32				/* tx queue stopped 'cuz full */
	ast_tx_encap			uint32				/* tx encapsulation failed */
	ast_tx_nonode			uint32				/* tx failed 'cuz no node */
	ast_tx_nobuf			uint32				/* tx failed 'cuz no tx buffer (data) */
	ast_tx_nobufmgt			uint32				/* tx failed 'cuz no tx buffer (mgmt)*/

	ast_tx_bcast			uint32
	ast_tx_mcast			uint32
	ast_tx_bcast_bytes		uint32
	ast_tx_mcast_bytes		uint32
//#ifdef AH_SUPPORT_MUMIMO
	ast_tx_mu				uint32
//#endif

	ast_tx_noack			uint32				/* tx frames with no ack marked */
	ast_tx_cts				uint32				/* tx frames with cts enabled */
	ast_tx_protect			uint32				/* tx frames with protection */
	ast_tx_xretries			uint32				/* tx failed 'cuz too many retries */

	ast_be_xmit				uint32				/* beacons transmitted */
	ast_tx_shortpre			uint32				/* tx frames with short preamble */
	ast_tx_altrate			uint32				/* tx frames with alternate rate */

	ast_tx_ok				uint32				/* tx ok data frame */
	ast_tx_fail				uint32				/* tx fail data frame */

	ast_rxorn				uint32				/* rx overrun interrupts */
	ast_txurn				uint32				/* tx underrun interrupts */

	ast_tx_fifoerr			uint32				/* tx failed 'cuz FIFO underrun */
	ast_tx_filtered			uint32				/* tx failed 'cuz xmit filtered */

	ast_rx_orn				uint32				/* rx failed 'cuz of desc overrun */
	ast_rx_badcrypt			uint32				/* rx failed 'cuz decryption */
	ast_rx_badmic			uint32				/* rx failed 'cuz MIC failure */
	ast_rx_nobuf			uint32				/* rx setup failed 'cuz no skbuff */
	ast_rx_swdecrypt		uint32				/* rx frames sw decrypted due to key miss */
	ast_rx_num_data			uint32
	ast_rx_num_mgmt			uint32
	ast_rx_num_ctl			uint32
	ast_rx_num_unknown		uint32

	ast_11n_stats			wl_11n_stats		/* 11n statistics */

	ast_rx_rssi			int8				/* last rx rssi */
	ast_chan_switch			uint32				/* no. of channel switch */
	ast_be_nobuf			uint32				/* no skbuff available for beacon */
}

type wl_11n_stats struct {
	rx_pkts					uint32				/* rx pkts */
	tx_bars					uint32				/* tx bars sent */
	tx_bars_drop			uint32				/* dropped tx bar frames */
	rx_bars					uint32				/* rx bars */
	tx_compaggr				uint32				/* tx aggregated completions */
	/* BCM XXX Not used yet */
	rx_compaggr				uint32
	txaggr_compxretry		uint32
	txunaggr_xretry			uint32
	tx_compunaggr			uint32
	tx_bawadv				uint32
	rx_aggr					uint32
	tx_retries				uint32				/* tx retries of sub frames */
	tx_xretries				uint32
}

/*
 * try to mirro ath_phy_stats here
 */
type wl_phy_stats struct {
	ast_tx_rts				uint64				/* RTS success count */
	ast_tx_shortretry		uint64				/* tx on-chip retries (short). RTSFailCnt */
	ast_tx_longretry		uint64				/* tx on-chip retries (long). DataFailCnt */
	ast_rx_tooshort			uint64				/* rx discarded 'cuz frame too short */
	ast_rx_toobig			uint64				/* rx discarded 'cuz frame too large */
        //u_int64_t   ast_rx_err	uint64 /* rx error */
	ast_rx_crcerr			uint64				/* rx failed 'cuz of bad CRC */
	ast_rx_crcerr_no_phyerr	uint64				/* rx crcerr without phy err */ //TBD
	ast_rx_fifoerr			uint64				/* rx failed 'cuz of FIFO overrun */
	ast_rx_phyerr			uint64				/* rx PHY error summary count */
        //u_int64_t   ast_rx_decrypterr	 uint64 /* rx decryption error */
        //u_int64_t   ast_rx_demicerr	uint64 /* rx demic error */
        //u_int64_t   ast_rx_demicok	uint64 /* rx demic ok */
        //u_int64_t   ast_rx_delim_pre_crcerr	uint64 /* pre-delimiter crc errors */
        //u_int64_t   ast_rx_delim_post_crcerr	uint64 /* post-delimiter crc errors */
        //u_int64_t   ast_rx_decrypt_busyerr	uint64 /* decrypt busy errors */
        //u_int64_t   ast_rx_phy	[32]uint64    /* rx PHY error per-code counts */
}

type ah_signal_quality_range struct
{
		min int8
		max int8
}

type ah_signal_quality_stats struct
{
		asqrange ah_signal_quality_range
		count uint32
}

type  ah_ieee80211_sta_stats_item struct {
		ns_mac				[MACADDR_LEN]byte            /* client mac address */
		ns_ssid				[AH_MAX_SSID_LEN + 1]byte   /* ssid associated */
		ns_snr				int8                         /* signal noise ratio */
		ns_tx_airtime		uint64                  /* tx airtime */
		ns_rx_airtime		uint64                  /* rx airtime */
		ns_tx_drops			uint32                    /* tx excessive retries, fifo err etc */
		ns_rx_drops			uint32                   /* due to: duplicate seq numbers, decrypt errors, security replay checking*/
		ns_tx_data			uint32             /* tx data frames */
		ns_tx_bytes			uint64             /* tx data count (bytes) */
		ns_rx_data			uint32                /* rx data frames */
		ns_rx_bytes			uint64             /* rx data count (bytes) */
		ns_sla_traps		uint32             /* sent how many sla violation traps */
		ns_sla_bm_score		uint32             /* sla benchmark core */
		ns_app_health_score	uint32         /* application health score */
		ns_ps_times			uint32             /* indicate client entered into power save time */
		ns_rx_probereq		uint32             /* rx probe request frames */
		ns_rx_mcast			uint32             /* rx multi/broadcast frames */

		ns_rx_rate_stats	[NS_HW_RATE_SIZE]ieee80211_node_rate_stats
		ns_tx_rate_stats	[NS_HW_RATE_SIZE]ieee80211_node_rate_stats
//#ifdef AH_CLIENT360_PHASE2_EXT_RADIO
		ns_tx_nss			[AH_TX_NSS_MAX]uint32      /* pkt number per tx spatial stream */
//#endif
//#ifdef AH_NETWORK360_WIFI_STATS
		ns_sq_group			[AH_SQ_TYPE_MAX][AH_SQ_GROUP_MAX]ah_signal_quality_stats
//#endif
		pad					[6]byte
}


type  ah_ieee80211_get_wifi_sta_stats struct {
		pointer		unsafe.Pointer
		count		uint16   /* air-time ring buffer count */
}

type ieee80211req_cfg_sta struct {
		cmd		uint32
		resv		uint32			/* Let following struct to align to 64bit for 64 bit machine,
							otherwise copy bgscan and lb status may with wrong value */
		wifi_sta_stats ah_ieee80211_get_wifi_sta_stats
}


type iwreq_data struct
{
		data uint32
}

type iwreq_clt struct
{
		ifr_name	[unix.IFNAMSIZ]byte    /* if name, e.g. "eth0" */
		u	iwreq_data
}

type ah_ieee80211_sta_info struct {
	mac				[MACADDR_LEN]uint8
	noise_floor		int16
	rssi			int32
	tx_ratekbps		int32
	tx_pkts			uint32
	tx_bytes		uint64
	rx_ratekbps		int32
	rx_pkts			uint32
	rx_bytes		uint64
	bw				uint32
}

type ieee80211req_cfg_one_sta struct{
		cmd		uint32
		resv	uint32			/* Let following struct to align to 64bit for 64 bit machine,
									otherwise copy bgscan and lb status may with wrong value */
		sta_info ah_ieee80211_sta_info
}

type ah_fw_dev_msg struct {
	hdr		ah_fe_ioctl_hdr
	data	ah_flow_get_sta_net_health_msg
}

type ah_fw_dev_ip_msg struct {
        hdr     ah_fe_ioctl_hdr
        data    ah_flow_get_sta_server_ip_msg
}




/* FE ioctl generic data structure */
type ah_fe_ioctl_hdr struct {
	retval		int32			/* return value */
	msg_type	uint16		/* sub message type */
	msg_size	uint16		/* I/O msg size not including header */
}

type ah_flow_get_sta_net_health_msg struct {
	mac			[MACADDR_LEN]uint8
	net_health_score	int32
}

type  ieee80211req_sta_info struct{
    isi_freq	uint16				/* MHz */
    isi_len		uint16			/* length (mult of 4) */
    isi_flags		uint32		 /* channel flags */
	isi_authmode	uint8			/* authentication algorithm */
    isi_rssi		int8
    isi_capinfo		uint16			/* capabilities */
    isi_erp		uint8			/* ERP element */
    isi_macaddr		[MACADDR_LEN]uint8
    isi_nrates		uint8
    isi_rates		[IEEE80211_RATE_MAXSIZE]uint8  /* negotiated rates */
    isi_txrate		uint8			/* index to isi_rates[] */
    isi_txratekbps	uint32      /* tx rate in Kbps, for 11n */
    uisi_ie_len		int16			/* IE length */
    isi_associd		uint16			/* assoc response */
    isi_txpower		uint16			/* current tx power */
    isi_vlan		uint16			/* vlan tag */
    isi_cipher		uint8
    isi_pmf		uint8                  /* 802.11w: sta in MFP (or PMF) */
    isi_assoc_time	uint32              /* sta association time */

	isi_htcap		uint16           /* HT capabilities */
	isi_rxratekbps	uint32      /* rx rate in Kbps */


	/* We use this as a common variable for legacy rates
	   and lln. We do not attempt to make it symmetrical
	   to isi xratekbps and isi xrate, which seem to be
	   separate due to legacy code. */
	/* XXX frag state? */
	/* variable length IE data */

	/* fix me XXX move to awe side */

	isi_ah_flags			uint32                /* aerohive flags */

	isi_mesh_flag			uint32               /* mesh state machine flags */
	isi_ucast_keyix			uint16             /* unicast key index */
	isi_mcast_keyix			[2]uint16      /* multicast key index */
	isi_mesh_state			uint8               /* mesh state */
	isi_mesh_role			uint8                /* mesh association role */

	isi_meshid				[IEEE80211_MESHID_LEN + 1]uint8     /* mesh id */

	isi_addition_auth_flag	uint8       /* addtional auth flag */
	isi_upid				uint32                    /* user profile ID */
	isi_phymode				uint32                 /* physical mode */
	isi_power_mode			uint32              /* power mode */

	isi_name			[AH_IEEE80211_STANAME_LEN + 1]byte
	isi_chwidth				uint32
	isi_noise_floor			int16               /* noise floor */
	isi_release_flag		int16              /* release flag for high density */


	isi_vhtcap				uint32

	isi_mu_groupid			uint8

	isi_roaming_time_start	uint32
	isi_assoc_phase_time	uint32      /* time used in assoc phase in millisecond */

	isi_negotiate_Kbps		uint32            /* negotiate rate, max supported rate in Kbps */

}

type  ieee80211req_cfg_one_sta_info struct{
	cmd		uint32
	resv		uint32          /* Let following struct to align to 64bit for 64 bit machine,
                                       otherwise copy bgscan and lb status may with wrong value */
	mac		[MACADDR_LEN]uint8
	pad		ieee80211req_sta_info

}

type ah_dns_time struct {
        dns_ip				uint32
        dns_response_time	uint32
        failure_cnt			uint32
}

type ah_flow_get_sta_server_ip_msg struct {

    mac        [MACADDR_LEN]uint8
    client_static_ip uint32
    dhcp_server    uint32
    dhcp_time   int32
    gateway    uint32
    dns     [AH_MAX_DNS_LIST]ah_dns_time
    num_dns_servers    uint8
}

type saved_stats struct {
	tx_airtime_min		uint64
	tx_airtime_max		uint64
	tx_airtime_average	uint64

	rx_airtime_min		uint64
	rx_airtime_max		uint64
	rx_airtime_average	uint64

	bw_usage_min		uint64
	bw_usage_max		uint64
	bw_usage_average	uint64

	tx_airtime			uint64
	rx_airtime			uint64
}

type ah_dcd_stats_report_rate_stats struct  {
	kbps				uint32     /* TX/RX rate Kbps */
	rate_dtn			uint8      /* TX/RX bit rate distribution */
	rate_suc_dtn		uint8     /* TX/RX bit rate sucess distribution */
}


type  ah_dcd_stats_report_int_data struct {

	tx_bit_rate			[NS_HW_RATE_SIZE]ah_dcd_stats_report_rate_stats
	rx_bit_rate			[NS_HW_RATE_SIZE]ah_dcd_stats_report_rate_stats

}

func ah_ifname_radio2vap(radio_name string ) string {
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

func getMacProtoMode(phymode uint32) uint32 {

    switch phymode{
		case  IEEE80211_MODE_11A, IEEE80211_MODE_TURBO_A:
			return AH_DCD_NMS_PHY_MODE_A

		case IEEE80211_MODE_11B:
			return AH_DCD_NMS_PHY_MODE_B

		case IEEE80211_MODE_11NA_HT20, IEEE80211_MODE_11NA_HT40PLUS, IEEE80211_MODE_11NA_HT40MINUS, IEEE80211_MODE_11NA_HT40:
			return AH_DCD_NMS_PHY_MODE_NA

		case IEEE80211_MODE_11NG_HT20, IEEE80211_MODE_11NG_HT40PLUS, IEEE80211_MODE_11NG_HT40MINUS, IEEE80211_MODE_11NG_HT40:
			return AH_DCD_NMS_PHY_MODE_NG

		case  IEEE80211_MODE_11AC_VHT20, IEEE80211_MODE_11AC_VHT40PLUS, IEEE80211_MODE_11AC_VHT40MINUS, IEEE80211_MODE_11AC_VHT40, IEEE80211_MODE_11AC_VHT80:
			return AH_DCD_NMS_PHY_MODE_AC

		case IEEE80211_MODE_11AX_2G_HE20, IEEE80211_MODE_11AX_2G_HE40:
			return AH_DCD_NMS_PHY_MODE_AX_2G

		case IEEE80211_MODE_11AX_5G_HE20, IEEE80211_MODE_11AX_5G_HE40, IEEE80211_MODE_11AX_5G_HE80, IEEE80211_MODE_11AX_5G_HE160:
			return AH_DCD_NMS_PHY_MODE_AX_5G

		case IEEE80211_MODE_11AX_6G_HE20, IEEE80211_MODE_11AX_6G_HE40, IEEE80211_MODE_11AX_6G_HE80, IEEE80211_MODE_11AX_6G_HE160:
			return AH_DCD_NMS_PHY_MODE_AX_6G

		default :
			return AH_DCD_NMS_PHY_MODE_G
    }

    return AH_DCD_NMS_PHY_MODE_G
}

func intToIp(num uint32) string {

    b := make([]byte, 4)
    b[0] = byte(num)
    b[1] = byte(num >> 8)
    b[2] = byte(num >> 16)
    b[3] = byte(num >> 24)
    return fmt.Sprintf("%d.%d.%d.%d",b[0],b[1],b[2],b[3])
}

func reportGetDiff(curr uint32, last uint32) uint32 {
	if curr >= last {
		return (curr - last)
	} else {
		return curr
	}
}
