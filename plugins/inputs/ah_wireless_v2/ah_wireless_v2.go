package ah_wireless_v2

import (
	"log"
	"os"
	"bytes"
	"strconv"
	"syscall"
	"strings"
	"sync"
	"net"
	"time"
	"os/exec"
	"fmt"
	"sort"
	"unsafe"
	"runtime/debug"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/common/ahutil"
	"golang.org/x/sys/unix"
)

var (
	offsetsMutex = new(sync.Mutex)
	newLineByte  = []byte("\n")
	rrmid int = 0
)


type Ah_wireless struct {
	fd			int
	fe_fd			uintptr
	eth_fd			uintptr
	intf_m			map[string]map[string]string
	arp_m			map[string]string
	Ifname			[]string	`toml:"ifname"`
	Eth_ioctl		uint64		`toml:"eth_ioctl"`
	Scount			uint8		`toml:"scount"`
	Test_rf_stats_enable	  uint8		`toml:"test_rf_stats_enable"`
	Test_client_stats_enable  uint8		`toml:"test_client_stats_enable"`
	Test_device_stats_enable  uint8		`toml:"test_device_stats_enable"`
	Test_network_stats_enable uint8		`toml:"test_network_stats_enable"`
	closed			chan		struct{}
	numclient		[4]int
	timer_count		uint8
	entity			map[string]map[string]saved_stats
	Log			telegraf.Logger `toml:"-"`
	last_rf_stat		[4]awestats
	last_alarm_int          [4]alarm_int
	last_alarm 		[4]alarm
	last_ut_data		[4]utilization_data
	last_clt_stat		[4][]ah_ieee80211_sta_stats_item
	last_sq			map[string]map[int]map[int]ah_signal_quality_stats
	wg			sync.WaitGroup
	if_stats		[AH_MAX_WIRED]stats_interface_data
	ethx_stats		[AH_MAX_WIRED]stats_ethx_data
	nw_health		network_health_data
	nw_service		network_service_data
	fw_stats		firewall_stats_data
}

/*
 * Convert MHz frequency to IEEE channel number.
 */
func freqToChan(freq uint16) uint16 {
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

func ah_pow_10(exponent int) int64 {
    var result int64 = 1
    var factor float64

    if 0 == exponent {
        return 1
    }

    if exponent > 0 {
        factor = 10
    } else {
        factor = 0.1
        exponent = (0 - exponent)
    }

    for (exponent > 0) {
        result = int64((float64(result) * factor))
        exponent = exponent - 1
    }
    return result
}

/*
 * Retrieve Channel Width from Phymode
 */
func getChannelWidth(phymode uint32) uint32 {

    switch phymode{
		case 	IEEE80211_MODE_AUTO,
				IEEE80211_MODE_11A,
				IEEE80211_MODE_11B,
				IEEE80211_MODE_11G,
				IEEE80211_MODE_FH,
				IEEE80211_MODE_TURBO_A,
				IEEE80211_MODE_TURBO_G,
				IEEE80211_MODE_11NA_HT20,
				IEEE80211_MODE_11NG_HT20,
				IEEE80211_MODE_11AC_VHT20,
				IEEE80211_MODE_11AX_2G_HE20,
				IEEE80211_MODE_11AX_5G_HE20,
				IEEE80211_MODE_11AX_6G_HE20,
				IEEE80211_MODE_11BE_2G_EHT20,
				IEEE80211_MODE_11BE_5G_EHT20,
				IEEE80211_MODE_11BE_6G_EHT20 :

					return IEEE80211_CWM_WIDTH20

		case	IEEE80211_MODE_11NA_HT40PLUS,
				IEEE80211_MODE_11NA_HT40MINUS,
				IEEE80211_MODE_11NG_HT40PLUS,
				IEEE80211_MODE_11NG_HT40MINUS,
				IEEE80211_MODE_11NG_HT40,
				IEEE80211_MODE_11NA_HT40,
				IEEE80211_MODE_11AC_VHT40PLUS,
				IEEE80211_MODE_11AC_VHT40MINUS,
				IEEE80211_MODE_11AC_VHT40,
				IEEE80211_MODE_11AX_2G_HE40,
				IEEE80211_MODE_11AX_5G_HE40,
				IEEE80211_MODE_11AX_6G_HE40,
				IEEE80211_MODE_11BE_2G_EHT40,
				IEEE80211_MODE_11BE_5G_EHT40,
				IEEE80211_MODE_11BE_6G_EHT40 :

					return IEEE80211_CWM_WIDTH40

		case
				IEEE80211_MODE_11AC_VHT80,
				IEEE80211_MODE_11AX_5G_HE80,
				IEEE80211_MODE_11AX_6G_HE80,
				IEEE80211_MODE_11BE_5G_EHT80,
				IEEE80211_MODE_11BE_6G_EHT80 :

					return IEEE80211_CWM_WIDTH80

		case	IEEE80211_MODE_11AC_VHT160,
				IEEE80211_MODE_11AX_5G_HE160,
				IEEE80211_MODE_11AX_6G_HE160,
				IEEE80211_MODE_11BE_5G_EHT160,
				IEEE80211_MODE_11BE_6G_EHT160 :

					return IEEE80211_CWM_WIDTH160

		case	IEEE80211_MODE_11BE_6G_EHT320,
				IEEE80211_MODE_11BE_6G_EHT320_1,
				IEEE80211_MODE_11BE_6G_EHT320_2 :

					return IEEE80211_CWM_WIDTH320

		default :
					return IEEE80211_CWM_WIDTH20
    }
}

func ah_ioctl(fd uintptr, op, argp uintptr) error {
	        _, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(op), argp)
        if errno != 0 {
                return errno
        }
        return nil
}


const sampleConfig = `
[[inputs.ah_wireless_v2]]
  interval = "5s"
  scount = 10
  ifname = ["wifi0","wifi1"]
  eth_ioctl = -6767123671
  Test_rf_stats_enable = 0
  Test_client_stats_enable = 0
  Test_device_stats_enable = 0
  Test_network_stats_enable = 0
`
func NewAh_wireless(id int) *Ah_wireless {
	var err error
	// Create RAW  Socket.
        fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
        if err != nil {
                return nil
        }


	if id != -1 {
		open(fd, id)
	}

	return &Ah_wireless{
                fd:       fd,
				timer_count: 0,
				Eth_ioctl: 0,
				Scount: 10,
				Test_rf_stats_enable: 0,
				Test_client_stats_enable: 0,
				Test_device_stats_enable: 0,
				Test_network_stats_enable: 0,
        }

}

func get_rt_sta_info(t *Ah_wireless, mac_adrs string, upid int, data rt_sta_data) rt_sta_data {
	app := "telegraf_helper"

	arg0 := mac_adrs
	arg1 := strconv.Itoa(upid)

	cmd := exec.Command(app, arg0, arg1)
	output, err := cmd.Output()

	if err != nil {
		log.Printf(err.Error())
		return data
	}


	lines := strings.Split(string(output),"\n")

	var os_line, host_line, user_line, prof_line  string

	// Loop over the line to find and extract OS and HostName ans UserName
	for _, line := range lines {
		if strings.HasPrefix(line, "OS:") {
			os_line = strings.TrimSpace(strings.TrimPrefix(line, "OS:"))
		} else if strings.HasPrefix(line, "HostName:") {
			host_line = strings.TrimSpace(strings.TrimPrefix(line, "HostName:"))
		} else if strings.HasPrefix(line, "UserName:") {
			user_line = strings.TrimSpace(strings.TrimPrefix(line, "UserName:"))
		} else if strings.HasPrefix(line, "UserProfile:") {
			prof_line = strings.TrimSpace(strings.TrimPrefix(line, "UserProfile:"))
		}
	}

	data.os =   string(os_line)
	data.hostname = string(host_line)
	data.user = string(user_line)
	data.userprofile = string(prof_line)

	return data
}

func getChannel(fd int, ifname string) int32 {

	var freq uint16 = 0
	var channel int32 = 0
    request := iwreq_freq{}
    copy(request.ifr_name[:], ah_ifname_radio2vap(ifname))

	offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), SIOCGIWFREQ, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getHDDStat ioctl data error %s",err)
				offsetsMutex.Unlock()
                return channel
        }
		offsetsMutex.Unlock()

		if (request.u.m == 0) {
			channel = 0
		} else if (request.u.e == 0) { /* BCM XXX return in channel format */
			channel = request.u.m
		} else {
			freq = uint16((int64(request.u.m)) * ah_pow_10(int(request.u.e - 6)))
			channel = int32(freqToChan(freq))
		}		

        return channel
}

func getHDDStat(fd int, ifname string, cfg ieee80211req_cfg_hdd) ah_ieee80211_hdd_stats {

        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_HDD_STATS;

        iwp := iw_point{pointer: unsafe.Pointer(&cfg)}

        request := iwreq{data: iwp}

	request.data.length = VAP_BUFF_SIZE

        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

	offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getHDDStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return cfg.hdd_stats
        }
		offsetsMutex.Unlock()

        return cfg.hdd_stats
}

func getAtrTbl(fd int, ifname string, cfg ieee80211req_cfg_atr) ah_ieee80211_atr_user {

	/* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ATR_TBL;

	iwp := iw_point{pointer: unsafe.Pointer(&cfg)}

	request := iwreq{data: iwp}

	request.data.length = VAP_BUFF_SIZE

	copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

	offsetsMutex.Lock()

	if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getAtrTbl ioctl data error %s",err)
		offsetsMutex.Unlock()
                return cfg.atr
        }

	offsetsMutex.Unlock()

	return cfg.atr

}


func getRFStat(fd int, ifname string, p awestats) awestats {

	request := IFReqData{Data: uintptr(unsafe.Pointer(&p))}
	copy(request.Name[:], ifname)

	offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), SIOCGRADIOSTATS, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getRFStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return p
        }

	offsetsMutex.Unlock()

	return p
}

func getStaStat(fd int, ifname string, cfg ieee80211req_cfg_sta) ah_ieee80211_get_wifi_sta_stats {

        /* first 4 bytes is subcmd */
        cfg.cmd = IEEE80211_GET_WIFI_STA_STATS

        iwp := iw_point{pointer: unsafe.Pointer(&cfg)}

        request := iwreq{data: iwp}

        request.data.length = VAP_BUFF_SIZE

        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

        offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getStaStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return cfg.wifi_sta_stats
        }

        offsetsMutex.Unlock()

        return cfg.wifi_sta_stats
}

func getNumAssocs (fd int, ifname string) uint32 {

	ird := iwreq_data{}

        /* first 4 bytes is subcmd */
        ird.data = IEEE80211_PARAM_NUM_ASSOCS

        request := iwreq_clt{u: ird}

        copy(request.ifr_name[:], ah_ifname_radio2vap(ifname))


        offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GETPARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getNumAssocs ioctl data error %s",err)
		offsetsMutex.Unlock()
                return 0
        }


        offsetsMutex.Unlock()

	return uint32(request.u.data)
}

func getOneStaInfo(fd int, ifname string, mac_ad [MACADDR_LEN]uint8 , cfg ieee80211req_cfg_one_sta) ah_ieee80211_sta_info {

        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ONE_STA_INFO;
	cfg.sta_info.mac = mac_ad
        iwp := iw_point{pointer: unsafe.Pointer(&cfg)}
        request := iwreq{data: iwp}
        request.data.length = VAP_BUFF_SIZE
        copy(request.ifrn_name[:], ifname)

        offsetsMutex.Lock()
        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getOneStaInfo ioctl data error %s",err)
		offsetsMutex.Unlock()
                return cfg.sta_info
        }

        offsetsMutex.Unlock()

        return cfg.sta_info

}

func getOneSta(fd int, ifname string, mac_ad [MACADDR_LEN]uint8, cfg ieee80211req_cfg_one_sta_info) unsafe.Pointer {

        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ONE_STA
        cfg.mac = mac_ad
        iwp := iw_point{pointer: unsafe.Pointer(&cfg)}
        request := iwreq{data: iwp}
        request.data.length = VAP_BUFF_SIZE
        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

        offsetsMutex.Lock()
	if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getOneSta ioctl data error %s",err)
		offsetsMutex.Unlock()
		return request.data.pointer
        }
	offsetsMutex.Unlock()

        return request.data.pointer

}

func getProcNetDev(ifname string) ah_dcd_dev_stats {
	var intfname string
	var stats ah_dcd_dev_stats
	table, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return stats;
	}

	lines := bytes.Split([]byte(table), newLineByte)
	comp := fmt.Sprintf(" %s:",ifname)

  for  _, curLine := range lines {
    if strings.Contains(string(curLine), comp) {
        fmt.Sscanf(string(curLine),
        "%s %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
                            &intfname,
                            &stats.rx_bytes,
                           &stats.rx_packets,
                           &stats.rx_errors,
                           &stats.rx_dropped,
                           &stats.rx_fifo_errors,
                           &stats.rx_frame_errors,
                           &stats.rx_compressed,  /* missing for <= 1 */
                           &stats.rx_multicast, /* missing for <= 1 */
                           &stats.tx_bytes, /* missing for 0 */
                           &stats.tx_packets,
                           &stats.tx_errors,
                           &stats.tx_dropped,
                           &stats.tx_fifo_errors,
                           &stats.collisions,
                           &stats.tx_carrier_errors,
                           &stats.tx_compressed,
                           &stats.tx_multicast,
                           &stats.rx_unicast,
                           &stats.rx_broadcast,
                           &stats.tx_unicast,
                           &stats.tx_broadcast)

    }
  }

	return stats
}

func getIfStatus(fd int, ifname string) int {
        ifr, err := unix.NewIfreq(ifname)
        if err != nil {
                log.Printf("failed to create ifreq for flags: %v", err)
                return -1
        }

        offsetsMutex.Lock()
        defer offsetsMutex.Unlock()

        if err := unix.IoctlIfreq(fd, unix.SIOCGIFFLAGS, ifr); err != nil {
                log.Printf("getIfStatus ioctl error: %s", err)
                return -1
        }

        return int(ifr.Uint16())
}

func getIfIndex(fd int, ifname string) int {
        ifr, err := unix.NewIfreq(ifname)

        if err != nil {
                log.Printf("failed to create ifreq  %v", err)
        }

	offsetsMutex.Lock()

        if err := unix.IoctlIfreq(fd, unix.SIOCGIFINDEX, ifr); err != nil {
                log.Printf("getIfIndex ioctl error %s",err)
		offsetsMutex.Unlock()
                return -1
        }

	offsetsMutex.Unlock()
	return int(ifr.Uint32())

}

func getEthLink(t *Ah_wireless, fd uintptr, iName string) int32 {

	var eth_ioctl = t.Eth_ioctl

	infr := ifreq_eth{}
	copy(infr.ifr_name[:], iName)

	infr.ifru_ivalue = 0
	ecmd := ah_ethif_cmd{ifr : infr}
	copy(ecmd.ifname[:], iName)

	ecmd.cmd = AH_SIOCCIFGETLINK

	offsetsMutex.Lock()

	if err := ah_ioctl(uintptr(fd), uintptr(eth_ioctl) /*AH_ETHIF_IOCTL_CMD*/, uintptr(unsafe.Pointer(&ecmd))); err != nil {

		log.Printf("getEthLink ioctl data error %s",err)
		offsetsMutex.Unlock()
		return -1
	}
	offsetsMutex.Unlock()
	link := ecmd.ifr.ifru_ivalue

	return link

}

func getEthStatus(t *Ah_wireless, fd uintptr, iName string) int32 {

	var eth_ioctl = t.Eth_ioctl

	infr := ifreq_eth{}
	copy(infr.ifr_name[:], iName)

	infr.ifru_ivalue = 0
	ecmd := ah_ethif_cmd{ifr : infr}
	copy(ecmd.ifname[:], iName)

	ecmd.cmd = AH_SIOCCIFGETSTATUS

	offsetsMutex.Lock()

	if err := ah_ioctl(uintptr(fd), uintptr(eth_ioctl) /*AH_ETHIF_IOCTL_CMD*/, uintptr(unsafe.Pointer(&ecmd))); err != nil {

		log.Printf("getEthStatus ioctl data error %s",err)
		offsetsMutex.Unlock()
		return -1
	}
	offsetsMutex.Unlock()
	link := ecmd.ifr.ifru_ivalue

	return link

}

func get_radio_band(t *Ah_wireless, ifname string)  string {

	app := "wl"

	arg0 := "-i"
	arg1 := ifname
	arg2 := "band"

	cmd := exec.Command(app, arg0, arg1, arg2)
	output, err := cmd.Output()

	if err != nil {
		log.Printf(err.Error())
		return "INVALID"
	}

	lines := strings.Split(string(output),"\n")

	switch lines[0] {
		case "a":
			return "5G"
		case "b":
			return "2.4G"
		case "6g":
			return "6G"
		default:
			return "INVALID"
	}
}

func load_ssid(t *Ah_wireless, ifname string) {
	// Get a list of all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error getting network interfaces: %v", err)
	}

	// Iterate through the interfaces and print their details

	for _, iface := range interfaces {
		// Check if the interface is a wireless interface (contains "wifi." in its name)
		if strings.Contains(iface.Name, ifname + ".") {

			app := "wl"
			arg0 := "-i"
			arg1 := iface.Name
			arg2 := "status"

			cmd := exec.Command(app, arg0, arg1, arg2)
			output, err := cmd.Output()

			if err != nil {
				continue
			}

			lines := strings.Split(string(output),"\n")
			temp  := strings.Split(lines[0]," ")

			ssid := strings.Trim(temp[1], "\"")
			t.intf_m[ifname][ssid] = iface.Name
		}
	}

}

func getFeIpnetScore(fd uintptr, clmac [MACADDR_LEN]uint8) int32 {

	msg := ah_flow_get_sta_net_health_msg{
					mac: clmac,
					net_health_score: 0,
				}
	ihdr := ah_fe_ioctl_hdr{
					retval: -1,
					msg_type: AH_GET_STATION_NETWORK_HEALTH,
					msg_size: uint16(unsafe.Sizeof(msg)),
				}
        dev_msg := ah_fw_dev_msg{
					hdr: ihdr,
					data: msg,
				}

        offsetsMutex.Lock()

        if err := ah_ioctl(fd, AH_FE_IOCTL_FLOW, uintptr(unsafe.Pointer(&dev_msg))); err != nil {
                log.Printf("getFeIpnetScore ioctl data error %s",err)
		offsetsMutex.Unlock()
                return -1
	}

	offsetsMutex.Unlock()

	if dev_msg.hdr.retval < 0 {
		log.Printf("Open ioctl data erro")
		return -1
	}

	return dev_msg.data.net_health_score
}

func getFeServerIp(fd uintptr, clmac [MACADDR_LEN]uint8, dev_msg ah_fw_dev_ip_msg) ah_flow_get_sta_server_ip_msg {

        msg := ah_flow_get_sta_server_ip_msg{
                                        mac: clmac,
                                }
        ihdr := ah_fe_ioctl_hdr{
                                        retval: -1,
                                        msg_type: AH_FLOW_GET_STATION_SERVER_IP,
                                        msg_size: uint16(unsafe.Sizeof(msg)),
                                }

		dev_msg.hdr = ihdr
		dev_msg.data = msg

        offsetsMutex.Lock()

        if err := ah_ioctl(fd, AH_FE_IOCTL_FLOW, uintptr(unsafe.Pointer(&dev_msg))); err != nil {
                log.Printf("getFeServerIp ioctl data error %s",err)
                offsetsMutex.Unlock()
                return dev_msg.data
        }

        offsetsMutex.Unlock()


        if dev_msg.hdr.retval < 0 {
                log.Printf("Open ioctl data erro")
                return dev_msg.data
        }

        return dev_msg.data
}

func open(fd, id int) *Ah_wireless {

	return &Ah_wireless{fd: fd, closed: make(chan struct{})}
}

func (t *Ah_wireless) SampleConfig() string {
	return sampleConfig
}

func (t *Ah_wireless) Description() string {
	return "Hive OS wireless stat"
}

func (t *Ah_wireless) Init() error {
	return nil
}


func dumpOutput(outfile string , outline string, append int) error {

	var f *os.File
	var err error

	if append == 1 {
		f, err = os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	} else {
		f, err = os.Create(outfile)
	}
	if err != nil {
		log.Printf(fmt.Sprint(err))
		return err
	}

	if append == 1 {
		f.WriteString("\n\n\n\n")
		if err != nil {
			f.Close()
			log.Printf(fmt.Sprint(err))
			return err
		}
	}

	f.WriteString(outline)

	if err != nil {
		f.Close()
		log.Printf(fmt.Sprint(err))
		return err
	}

	err = f.Close()
	if err != nil {
		log.Printf(fmt.Sprint(err))
		return err
	}

	return nil

}

func prepareAndDumpOutput(outfile string ,fields map[string]interface{}) error {

	var s string

	for k, v := range fields {
		if  fmt.Sprint(v) == "0" { // Check if the value is zero
			delete(fields, k)
		}
	}

	keys := make([]string, 0, len(fields))

	for k := range fields{
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		s = s + k + " : " + fmt.Sprint(fields[k]) + "\n"
	}

	s = s + "---------------------------------------------------------------------------------------------\n"

	dumpOutput(outfile, s, 1)
	return nil
}

func Gather_Rf_Avg(t *Ah_wireless, acc telegraf.Accumulator) error {
	var ii int
	ii = 0
	for _, intfName := range t.Ifname {

		var rfstat awestats

		var atrSt ieee80211req_cfg_atr
		var atrStat ah_ieee80211_atr_user

		var idx			int

		var tx_total	int64
		var rx_total	int64
		var tot_tx_bitrate_retries uint32
		var tot_rx_bitrate_retries uint32

		rfstat  = getRFStat(t.fd, intfName, rfstat)

		atrStat = getAtrTbl(t.fd, intfName, atrSt)

		/* We need check and aggregation Tx/Rx bit rate distribution
 		* prcentage, if the bit rate equal in radio interface or client reporting.
 		*/

		for i := 0; i < NS_HW_RATE_SIZE; i++{

			if ((rfstat.ast_rx_rate_stats[i].ns_rateKbps == 0) && (rfstat.ast_tx_rate_stats[i].ns_rateKbps == 0)) {
				continue
			}

			for j := (i+1); j < NS_HW_RATE_SIZE; j++ {
				if ((rfstat.ast_rx_rate_stats[i].ns_rateKbps != 0) && (rfstat.ast_rx_rate_stats[i].ns_rateKbps == rfstat.ast_rx_rate_stats[j].ns_rateKbps)) {
					rfstat.ast_rx_rate_stats[i].ns_unicasts += rfstat.ast_rx_rate_stats[j].ns_unicasts;
					rfstat.ast_rx_rate_stats[i].ns_retries += rfstat.ast_rx_rate_stats[j].ns_retries;
					rfstat.ast_rx_rate_stats[j].ns_rateKbps = 0;
					rfstat.ast_rx_rate_stats[j].ns_unicasts = 0;
					rfstat.ast_rx_rate_stats[j].ns_retries = 0;
				}
				if ((rfstat.ast_tx_rate_stats[i].ns_rateKbps != 0) && (rfstat.ast_tx_rate_stats[i].ns_rateKbps == rfstat.ast_tx_rate_stats[j].ns_rateKbps)) {
					rfstat.ast_tx_rate_stats[i].ns_unicasts += rfstat.ast_tx_rate_stats[j].ns_unicasts;
					rfstat.ast_tx_rate_stats[i].ns_retries += rfstat.ast_tx_rate_stats[j].ns_retries;
					rfstat.ast_tx_rate_stats[j].ns_rateKbps = 0;
					rfstat.ast_tx_rate_stats[j].ns_unicasts = 0;
					rfstat.ast_tx_rate_stats[j].ns_retries = 0;
				}
			}
		}



/* Rate Calculation Copied from DCD code */

	/* Tx/Rx bit rate distribution */
	for idx = 0; idx < NS_HW_RATE_SIZE; idx++ {
		tx_total += int64(reportGetDiff(rfstat.ast_tx_rate_stats[idx].ns_unicasts,
			t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_unicasts))

		rx_total += int64(reportGetDiff(rfstat.ast_rx_rate_stats[idx].ns_unicasts,
			t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_unicasts))

		tot_tx_bitrate_retries += reportGetDiff(rfstat.ast_tx_rate_stats[idx].ns_retries,
			t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_retries)

		tot_rx_bitrate_retries += reportGetDiff(rfstat.ast_rx_rate_stats[idx].ns_retries,
			t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_retries)

	}

/* Rate calculation copied from DCD code */

			if (t.last_ut_data[ii].noise_min == 0) || (t.last_ut_data[ii].noise_min >= rfstat.ast_noise_floor) {
				t.last_ut_data[ii].noise_min = rfstat.ast_noise_floor
			}
			if (t.last_ut_data[ii].noise_max == 0 ) || (t.last_ut_data[ii].noise_max <= rfstat.ast_noise_floor) {
				t.last_ut_data[ii].noise_max = rfstat.ast_noise_floor
			}
			t.last_ut_data[ii].noise_avg = (t.last_ut_data[ii].noise_avg + rfstat.ast_noise_floor)/2

			if (t.last_ut_data[ii].crc_err_rate_min == 0 ) || (t.last_ut_data[ii].crc_err_rate_min >= rfstat.phy_stats.ast_rx_crcerr) {
				t.last_ut_data[ii].crc_err_rate_min = rfstat.phy_stats.ast_rx_crcerr
			}
			if (t.last_ut_data[ii].crc_err_rate_max == 0 ) || (t.last_ut_data[ii].crc_err_rate_max <= rfstat.phy_stats.ast_rx_crcerr) {
				t.last_ut_data[ii].crc_err_rate_max = rfstat.phy_stats.ast_rx_crcerr
			}
			t.last_ut_data[ii].crc_err_rate_avg = (t.last_ut_data[ii].crc_err_rate_avg + rfstat.phy_stats.ast_rx_crcerr)/2

			if atrStat.count > 0 {

				rx_util := atrStat.atr_info[atrStat.count - 1].rxf_pcnt
				tx_util := atrStat.atr_info[atrStat.count - 1].txf_pcnt

				total_util := atrStat.atr_info[atrStat.count - 1].rxc_pcnt

				var chan_util int32
				var interface_utiliation int32
				if total_util > 100 {
					chan_util = 100
				} else {
					chan_util = int32(total_util)
				}

				/* Calculate Utilization */
				if (total_util > (rx_util + tx_util)) {
					interface_utiliation = int32(total_util) - int32(rx_util) - int32(tx_util)
				} else {
					interface_utiliation = 0
				}
				if (t.last_ut_data[ii].intfer_util_min == 0) || (t.last_ut_data[ii].intfer_util_min >= interface_utiliation) {
					t.last_ut_data[ii].intfer_util_min = interface_utiliation
				}
				if (t.last_ut_data[ii].intfer_util_max == 0) || (t.last_ut_data[ii].intfer_util_max <= interface_utiliation) {
					t.last_ut_data[ii].intfer_util_max = interface_utiliation
				}
				t.last_ut_data[ii].intfer_util_avg = (t.last_ut_data[ii].intfer_util_avg + interface_utiliation)/2

				if (t.last_ut_data[ii].chan_util_min == 0 ) || (t.last_ut_data[ii].chan_util_min >= chan_util) {
					t.last_ut_data[ii].chan_util_min = chan_util
				}
				if (t.last_ut_data[ii].chan_util_max == 0 ) || (t.last_ut_data[ii].chan_util_max <= chan_util) {
					t.last_ut_data[ii].chan_util_max = chan_util
				}
				t.last_ut_data[ii].chan_util_avg = (t.last_ut_data[ii].chan_util_avg + chan_util)/2

				if (t.last_ut_data[ii].tx_util_min == 0) || (t.last_ut_data[ii].tx_util_min >= tx_util) {
					t.last_ut_data[ii].tx_util_min = tx_util
				}
				if (t.last_ut_data[ii].tx_util_max == 0) || (t.last_ut_data[ii].tx_util_max <= tx_util) {
					t.last_ut_data[ii].tx_util_max = tx_util
				}
				t.last_ut_data[ii].tx_util_avg = (t.last_ut_data[ii].tx_util_avg + tx_util)/2

				if (t.last_ut_data[ii].rx_util_min == 0) || (t.last_ut_data[ii].rx_util_min >= rx_util) {
					t.last_ut_data[ii].rx_util_min = rx_util
				}
				if (t.last_ut_data[ii].rx_util_max == 0) || (t.last_ut_data[ii].rx_util_max <= rx_util) {
					t.last_ut_data[ii].rx_util_max = rx_util
				}
				t.last_ut_data[ii].rx_util_avg = (t.last_ut_data[ii].rx_util_avg + rx_util)/2


				if (t.last_ut_data[ii].rx_ibss_util_min == 0) || (t.last_ut_data[ii].rx_ibss_util_min >= atrStat.atr_info[atrStat.count - 1].rxf_inbss) {
					t.last_ut_data[ii].rx_ibss_util_min = atrStat.atr_info[atrStat.count - 1].rxf_inbss
				}
				if (t.last_ut_data[ii].rx_ibss_util_max == 0) || (t.last_ut_data[ii].rx_ibss_util_max <= atrStat.atr_info[atrStat.count - 1].rxf_inbss) {
					t.last_ut_data[ii].rx_ibss_util_max = atrStat.atr_info[atrStat.count - 1].rxf_inbss
				}
				t.last_ut_data[ii].rx_ibss_util_avg = (t.last_ut_data[ii].rx_ibss_util_avg + atrStat.atr_info[atrStat.count - 1].rxf_inbss)/2

				if (t.last_ut_data[ii].rx_obss_util_min == 0) || (t.last_ut_data[ii].rx_obss_util_min >= atrStat.atr_info[atrStat.count - 1].rxf_obss) {
					t.last_ut_data[ii].rx_obss_util_min = atrStat.atr_info[atrStat.count - 1].rxf_obss
				}
				if (t.last_ut_data[ii].rx_obss_util_max == 0) || (t.last_ut_data[ii].rx_obss_util_max <= atrStat.atr_info[atrStat.count - 1].rxf_obss) {
					t.last_ut_data[ii].rx_obss_util_max = atrStat.atr_info[atrStat.count - 1].rxf_obss
				}
				t.last_ut_data[ii].rx_obss_util_avg = (t.last_ut_data[ii].rx_obss_util_avg + atrStat.atr_info[atrStat.count - 1].rxf_obss)/2


                                if (t.last_ut_data[ii].wifi_i_util_min == 0) || (t.last_ut_data[ii].wifi_i_util_min >= atrStat.atr_info[atrStat.count-1].wifi_interference) {
                                        t.last_ut_data[ii].wifi_i_util_min = atrStat.atr_info[atrStat.count-1].wifi_interference
                                }
                                if (t.last_ut_data[ii].wifi_i_util_max == 0) || (t.last_ut_data[ii].wifi_i_util_max <= atrStat.atr_info[atrStat.count-1].wifi_interference) {
                                        t.last_ut_data[ii].wifi_i_util_max = atrStat.atr_info[atrStat.count-1].wifi_interference
                                }
                                t.last_ut_data[ii].wifi_i_util_avg = (t.last_ut_data[ii].wifi_i_util_avg + atrStat.atr_info[atrStat.count-1].wifi_interference) / 2

				/* Calculate Utilization */

			}

			t.last_rf_stat[ii] = rfstat
			ii++
		}

		return nil
}



func Gather_Rf_Stat(t *Ah_wireless, acc telegraf.Accumulator) error {
	var ii int
	ii = 0

	for _, intfName := range t.Ifname {

		if !ahutil.Check_Vap_Status(intfName) {
			log.Printf("VAP %s is not up, rfStat not collected\n", intfName)
			continue
		}

		var rfstat awestats
		var ifindex int
		var chann int32
		var atrSt ieee80211req_cfg_atr
		var atrStat ah_ieee80211_atr_user
		var hddStat ah_ieee80211_hdd_stats
		var hdd ieee80211req_cfg_hdd

		var idx			int
		var tmp_count1	int64
		var tmp_count2	int64
		var tx_total	int64
		var rx_total	int64
		var tmp_count3	int32
		var tmp_count4	int32
		var tot_tx_bitrate_retries uint32
		var tot_rx_bitrate_retries uint32

		var rf_report	ah_dcd_stats_report_int_data

		rfstat  = getRFStat(t.fd, intfName, rfstat)

		ifindex = getIfIndex(t.fd, intfName)
		if (ifindex <= 0) {
			continue
		}

		atrStat = getAtrTbl(t.fd, intfName, atrSt)

		hddStat = getHDDStat(t.fd, intfName, hdd)

		chann = 0
		chann = getChannel(t.fd, intfName);

		/* We need check and aggregation Tx/Rx bit rate distribution
 		* prcentage, if the bit rate equal in radio interface or client reporting.
 		*/

		for i := 0; i < NS_HW_RATE_SIZE; i++{

			if ((rfstat.ast_rx_rate_stats[i].ns_rateKbps == 0) && (rfstat.ast_tx_rate_stats[i].ns_rateKbps == 0)) {
				continue
			}

			for j := (i+1); j < NS_HW_RATE_SIZE; j++ {
				if ((rfstat.ast_rx_rate_stats[i].ns_rateKbps != 0) && (rfstat.ast_rx_rate_stats[i].ns_rateKbps == rfstat.ast_rx_rate_stats[j].ns_rateKbps)) {
					rfstat.ast_rx_rate_stats[i].ns_unicasts += rfstat.ast_rx_rate_stats[j].ns_unicasts;
					rfstat.ast_rx_rate_stats[i].ns_retries += rfstat.ast_rx_rate_stats[j].ns_retries;
					rfstat.ast_rx_rate_stats[j].ns_rateKbps = 0;
					rfstat.ast_rx_rate_stats[j].ns_unicasts = 0;
					rfstat.ast_rx_rate_stats[j].ns_retries = 0;
				}
				if ((rfstat.ast_tx_rate_stats[i].ns_rateKbps != 0) && (rfstat.ast_tx_rate_stats[i].ns_rateKbps == rfstat.ast_tx_rate_stats[j].ns_rateKbps)) {
					rfstat.ast_tx_rate_stats[i].ns_unicasts += rfstat.ast_tx_rate_stats[j].ns_unicasts;
					rfstat.ast_tx_rate_stats[i].ns_retries += rfstat.ast_tx_rate_stats[j].ns_retries;
					rfstat.ast_tx_rate_stats[j].ns_rateKbps = 0;
					rfstat.ast_tx_rate_stats[j].ns_unicasts = 0;
					rfstat.ast_tx_rate_stats[j].ns_retries = 0;
				}
			}
		}



/* Rate Calculation Copied from DCD code */

	/* Tx/Rx bit rate distribution */
	for idx = 0; idx < NS_HW_RATE_SIZE; idx++ {
		tx_total += int64(reportGetDiff(rfstat.ast_tx_rate_stats[idx].ns_unicasts,
			t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_unicasts))

		rx_total += int64(reportGetDiff(rfstat.ast_rx_rate_stats[idx].ns_unicasts,
			t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_unicasts))

		tot_tx_bitrate_retries += reportGetDiff(rfstat.ast_tx_rate_stats[idx].ns_retries,
			t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_retries)

		tot_rx_bitrate_retries += reportGetDiff(rfstat.ast_rx_rate_stats[idx].ns_retries,
			t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_retries)

	}

	for idx = 0; idx < NS_HW_RATE_SIZE; idx++ {

		tmp_count3 = int32(rfstat.ast_tx_rate_stats[idx].ns_unicasts - t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_unicasts)
		if (tx_total > 0 && tmp_count3 > 0) {
			rf_report.tx_bit_rate[idx].rate_dtn = uint8((int64(tmp_count3) * 100) / tx_total)
		} else {
			rf_report.tx_bit_rate[idx].rate_dtn = 0;
		}
		tmp_count4 = int32(rfstat.ast_rx_rate_stats[idx].ns_unicasts - t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_unicasts)
		if (rx_total > 0 && tmp_count4 > 0) {
			rf_report.rx_bit_rate[idx].rate_dtn = uint8((int64(tmp_count4) * 100) / rx_total)
		} else {
			rf_report.rx_bit_rate[idx].rate_dtn = 0;
		}

		/* Tx/Rx bit rate success distribution */
		tmp_count1 = int64(rfstat.ast_tx_rate_stats[idx].ns_retries - t.last_rf_stat[ii].ast_tx_rate_stats[idx].ns_retries)
		tmp_count2 = tmp_count1 + int64(tmp_count3)
		if (tmp_count2 > 0 && rf_report.tx_bit_rate[idx].rate_dtn > 0) {
			rf_report.tx_bit_rate[idx].rate_suc_dtn = uint8((int64(tmp_count3) * 100) / tmp_count2)
			if (rf_report.tx_bit_rate[idx].rate_suc_dtn > 100) {
				rf_report.tx_bit_rate[idx].rate_suc_dtn  = 100
				log.Printf("stats report int data process: rate_suc_dtn1 is more than 100%\n")
			}
		} else {
			rf_report.tx_bit_rate[idx].rate_suc_dtn = 0;
		}

		tmp_count1 = int64(rfstat.ast_rx_rate_stats[idx].ns_retries - t.last_rf_stat[ii].ast_rx_rate_stats[idx].ns_retries)
		tmp_count2 = tmp_count1 + int64(tmp_count4)
		if (tmp_count2 > 0 && rf_report.rx_bit_rate[idx].rate_dtn > 0) {
			rf_report.rx_bit_rate[idx].rate_suc_dtn = uint8((int64(tmp_count4) * 100) / tmp_count2)
			if (rf_report.rx_bit_rate[idx].rate_suc_dtn > 100) {
				rf_report.rx_bit_rate[idx].rate_suc_dtn = 100;
				log.Printf("stats report int data process: rate_suc_dtn2 is more than 100%\n");
			}

		} else {
			rf_report.rx_bit_rate[idx].rate_suc_dtn = 0;
		}
		rf_report.tx_bit_rate[idx].kbps = rfstat.ast_tx_rate_stats[idx].ns_rateKbps;
		rf_report.rx_bit_rate[idx].kbps = rfstat.ast_rx_rate_stats[idx].ns_rateKbps;
	}


/* Rate calculation copied from DCD code */

		fields := map[string]interface{}{

			"name_keys":					intfName,
			"ifIndex_keys":					ifindex,

		}

			fields["band"] = get_radio_band(t, intfName)
			fields["rrmId"] = rrmid
			fields["txPower"] = ahutil.GetTxPower(intfName)

			if (t.last_ut_data[ii].noise_min == 0) || (t.last_ut_data[ii].noise_min >= rfstat.ast_noise_floor) {
				t.last_ut_data[ii].noise_min = rfstat.ast_noise_floor
			}
			if (t.last_ut_data[ii].noise_max == 0 ) || (t.last_ut_data[ii].noise_max <= rfstat.ast_noise_floor) {
				t.last_ut_data[ii].noise_max = rfstat.ast_noise_floor
			}
			t.last_ut_data[ii].noise_avg = (t.last_ut_data[ii].noise_avg + rfstat.ast_noise_floor)/2

			if (t.last_ut_data[ii].crc_err_rate_min == 0 ) || (t.last_ut_data[ii].crc_err_rate_min >= rfstat.phy_stats.ast_rx_crcerr) {
				t.last_ut_data[ii].crc_err_rate_min = rfstat.phy_stats.ast_rx_crcerr
			}
			if (t.last_ut_data[ii].crc_err_rate_max == 0 ) || (t.last_ut_data[ii].crc_err_rate_max <= rfstat.phy_stats.ast_rx_crcerr) {
				t.last_ut_data[ii].crc_err_rate_max = rfstat.phy_stats.ast_rx_crcerr
			}
			t.last_ut_data[ii].crc_err_rate_avg = (t.last_ut_data[ii].crc_err_rate_avg + rfstat.phy_stats.ast_rx_crcerr)/2

			if atrStat.count > 0 {

				rx_util := atrStat.atr_info[atrStat.count - 1].rxf_pcnt
				tx_util := atrStat.atr_info[atrStat.count - 1].txf_pcnt

				total_util := atrStat.atr_info[atrStat.count - 1].rxc_pcnt

				var chan_util int32
				var interface_utiliation int32
				if total_util > 100 {
					chan_util = 100
				} else {
					chan_util = int32(total_util)
				}

				/* Calculate Utilization */
				if (total_util > (rx_util + tx_util)) {
					interface_utiliation = int32(total_util) - int32(rx_util) - int32(tx_util)
				} else {
					interface_utiliation = 0
				}
				if (t.last_ut_data[ii].intfer_util_min == 0) || (t.last_ut_data[ii].intfer_util_min >= interface_utiliation) {
					t.last_ut_data[ii].intfer_util_min = interface_utiliation
				}
				if (t.last_ut_data[ii].intfer_util_max == 0) || (t.last_ut_data[ii].intfer_util_max <= interface_utiliation) {
					t.last_ut_data[ii].intfer_util_max = interface_utiliation
				}
				t.last_ut_data[ii].intfer_util_avg = (t.last_ut_data[ii].intfer_util_avg + interface_utiliation)/2

				if (t.last_ut_data[ii].chan_util_min == 0 ) || (t.last_ut_data[ii].chan_util_min >= chan_util) {
					t.last_ut_data[ii].chan_util_min = chan_util
				}
				if (t.last_ut_data[ii].chan_util_max == 0 ) || (t.last_ut_data[ii].chan_util_max <= chan_util) {
					t.last_ut_data[ii].chan_util_max = chan_util
				}
				t.last_ut_data[ii].chan_util_avg = (t.last_ut_data[ii].chan_util_avg + chan_util)/2

				if (t.last_ut_data[ii].tx_util_min == 0) || (t.last_ut_data[ii].tx_util_min >= tx_util) {
					t.last_ut_data[ii].tx_util_min = tx_util
				}
				if (t.last_ut_data[ii].tx_util_max == 0) || (t.last_ut_data[ii].tx_util_max <= tx_util) {
					t.last_ut_data[ii].tx_util_max = tx_util
				}
				t.last_ut_data[ii].tx_util_avg = (t.last_ut_data[ii].tx_util_avg + tx_util)/2

				if (t.last_ut_data[ii].rx_util_min == 0) || (t.last_ut_data[ii].rx_util_min >= rx_util) {
					t.last_ut_data[ii].rx_util_min = rx_util
				}
				if (t.last_ut_data[ii].rx_util_max == 0) || (t.last_ut_data[ii].rx_util_max <= rx_util) {
					t.last_ut_data[ii].rx_util_max = rx_util
				}
				t.last_ut_data[ii].rx_util_avg = (t.last_ut_data[ii].rx_util_avg + rx_util)/2


				if (t.last_ut_data[ii].rx_ibss_util_min == 0) || (t.last_ut_data[ii].rx_ibss_util_min >= atrStat.atr_info[atrStat.count - 1].rxf_inbss) {
					t.last_ut_data[ii].rx_ibss_util_min = atrStat.atr_info[atrStat.count - 1].rxf_inbss
				}
				if (t.last_ut_data[ii].rx_ibss_util_max == 0) || (t.last_ut_data[ii].rx_ibss_util_max <= atrStat.atr_info[atrStat.count - 1].rxf_inbss) {
					t.last_ut_data[ii].rx_ibss_util_max = atrStat.atr_info[atrStat.count - 1].rxf_inbss
				}
				t.last_ut_data[ii].rx_ibss_util_avg = (t.last_ut_data[ii].rx_ibss_util_avg + atrStat.atr_info[atrStat.count - 1].rxf_inbss)/2

				if (t.last_ut_data[ii].rx_obss_util_min == 0) || (t.last_ut_data[ii].rx_obss_util_min >= atrStat.atr_info[atrStat.count - 1].rxf_obss) {
					t.last_ut_data[ii].rx_obss_util_min = atrStat.atr_info[atrStat.count - 1].rxf_obss
				}
				if (t.last_ut_data[ii].rx_obss_util_max == 0) || (t.last_ut_data[ii].rx_obss_util_max <= atrStat.atr_info[atrStat.count - 1].rxf_obss) {
					t.last_ut_data[ii].rx_obss_util_max = atrStat.atr_info[atrStat.count - 1].rxf_obss
				}
				t.last_ut_data[ii].rx_obss_util_avg = (t.last_ut_data[ii].rx_obss_util_avg + atrStat.atr_info[atrStat.count - 1].rxf_obss)/2

				if (t.last_ut_data[ii].wifi_i_util_min == 0) || (t.last_ut_data[ii].wifi_i_util_min >= atrStat.atr_info[atrStat.count-1].wifi_interference) {
					t.last_ut_data[ii].wifi_i_util_min = atrStat.atr_info[atrStat.count-1].wifi_interference
				}
				if (t.last_ut_data[ii].wifi_i_util_max == 0) || (t.last_ut_data[ii].wifi_i_util_max <= atrStat.atr_info[atrStat.count-1].wifi_interference) {
					t.last_ut_data[ii].wifi_i_util_max = atrStat.atr_info[atrStat.count-1].wifi_interference
				}
				t.last_ut_data[ii].wifi_i_util_avg = (t.last_ut_data[ii].wifi_i_util_avg + atrStat.atr_info[atrStat.count-1].wifi_interference) / 2
				/* Calculate Utilization */


				fields["interferenceUtilization_min"]		= t.last_ut_data[ii].intfer_util_min
				fields["interferenceUtilization_max"]		= t.last_ut_data[ii].intfer_util_max
				fields["interferenceUtilization_avg"]		= t.last_ut_data[ii].intfer_util_avg


				fields["channelUtilization_min"]			= t.last_ut_data[ii].chan_util_min
				fields["channelUtilization_max"]			= t.last_ut_data[ii].chan_util_max
				fields["channelUtilization_avg"]			= t.last_ut_data[ii].chan_util_avg

				fields["txUtilization_min"]					= t.last_ut_data[ii].tx_util_min
				fields["txUtilization_max"]					= t.last_ut_data[ii].tx_util_max
				fields["txUtilization_avg"]					= t.last_ut_data[ii].tx_util_avg

				fields["rxUtilization_min"]					= t.last_ut_data[ii].rx_util_min
				fields["rxUtilization_max"]					= t.last_ut_data[ii].rx_util_max
				fields["rxUtilization_avg"]					= t.last_ut_data[ii].rx_util_avg

				fields["rxInbssUtilization_min"]			= t.last_ut_data[ii].rx_ibss_util_min
				fields["rxInbssUtilization_max"]			= t.last_ut_data[ii].rx_ibss_util_max
				fields["rxInbssUtilization_avg"]			= t.last_ut_data[ii].rx_ibss_util_avg

				fields["rxObssUtilization_min"]				= t.last_ut_data[ii].rx_obss_util_min
				fields["rxObssUtilization_max"]				= t.last_ut_data[ii].rx_obss_util_max
				fields["rxObssUtilization_avg"]				= t.last_ut_data[ii].rx_obss_util_avg

				fields["wifinterferenceUtilization_min"]			= t.last_ut_data[ii].wifi_i_util_min
				fields["wifinterferenceUtilization_max"]			= t.last_ut_data[ii].wifi_i_util_max
				fields["wifinterferenceUtilization_avg"]			= t.last_ut_data[ii].wifi_i_util_avg

			} else {
				fields["channelUtilization_min"]			= 0
				fields["channelUtilization_max"]			= 0
				fields["channelUtilization_avg"]			= 0

				fields["txUtilization_min"]				= 0
				fields["txUtilization_max"]				= 0
				fields["txUtilization_avg"]				= 0

				fields["rxUtilization_min"]				= 0
				fields["rxUtilization_max"]				= 0
				fields["rxUtilization_avg"]				= 0

				fields["rxInbssUtilization_min"]			= 0
				fields["rxInbssUtilization_max"]			= 0
				fields["rxInbssUtilization_avg"]			= 0

				fields["rxObssUtilization_min"]				= 0
				fields["rxObssUtilization_max"]				= 0
				fields["rxObssUtilization_avg"]				= 0

				fields["wifinterferenceUtilization_min"]			= 0
				fields["wifinterferenceUtilization_max"]			= 0
				fields["wifinterferenceUtilization_avg"]			= 0

			}

//			fields["alarmFlag"] =  t.last_alarm_int[ii].alarm

			fields["wifinterferenceUtilization_min"]			= t.last_ut_data[ii].wifi_i_util_min
			fields["wifinterferenceUtilization_max"]			= t.last_ut_data[ii].wifi_i_util_max
			fields["wifinterferenceUtilization_avg"]			= t.last_ut_data[ii].wifi_i_util_avg

			fields["noise_min"]						= t.last_ut_data[ii].noise_min
			fields["noise_max"]						= t.last_ut_data[ii].noise_max
			fields["noise_avg"]						= t.last_ut_data[ii].noise_avg

			fields["crcErrorRate_min"]					= t.last_ut_data[ii].crc_err_rate_min
			fields["crcErrorRate_max"]					= t.last_ut_data[ii].crc_err_rate_max
			fields["crcErrorRate_avg"]					= t.last_ut_data[ii].crc_err_rate_avg


			fields["txPackets"]						= rfstat.ast_as.ast_tx_packets
			fields["txErrors"] 						= rfstat.ast_as.ast_tx_xretries + rfstat.ast_as.ast_tx_fifoerr
			fields["txDropped"]						= rfstat.ast_as.ast_tx_nobuf +  rfstat.ast_as.ast_tx_nobufmgt
			fields["txHwDropped"]						= rfstat.ast_as.ast_tx_shortpre + rfstat.ast_as.ast_tx_xretries + rfstat.ast_as.ast_tx_fifoerr
			fields["txSwDropped"]						= rfstat.ast_tx_blckd_drops
			fields["txBytes"]						 = rfstat.ast_as.ast_tx_bytes;
			fields["txRetryCount"]						= rfstat.phy_stats.ast_tx_shortretry + rfstat.phy_stats.ast_tx_longretry

			fields["txRate_min"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps
			fields["txRate_max"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps
			fields["txRate_avg"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps

			fields["txUnicastPackets"]					= rfstat.ast_tx_rate_stats[0].ns_unicasts
			fields["txMulticastPackets"]					= rfstat.ast_as.ast_tx_mcast
			fields["txMulticastBytes"]					= rfstat.ast_as.ast_tx_mcast_bytes
			fields["txBcastBytes"]						= rfstat.ast_as.ast_tx_bcast_bytes
			fields["txBcastPackets"]					= rfstat.ast_as.ast_tx_bcast

			fields["rxPackets"]						= rfstat.ast_as.ast_rx_num_data + rfstat.ast_as.ast_rx_num_mgmt + rfstat.ast_as.ast_rx_num_ctl
			fields["rxErrors"]						= rfstat.phy_stats.ast_rx_phyerr + rfstat.phy_stats.ast_rx_fifoerr + uint64(rfstat.ast_as.ast_rx_badcrypt) + uint64(rfstat.ast_as.ast_rx_badmic)
			fields["rxDropped"]						= rfstat.phy_stats.ast_rx_tooshort + uint64(rfstat.ast_as.ast_rx_nobuf) + rfstat.phy_stats.ast_rx_toobig
			fields["rxBytes"]						= rfstat.ast_as.ast_rx_bytes
			fields["rxRetryCount"]						= rfstat.ast_rx_retry

			fields["rxRate_min"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps
			fields["rxRate_max"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps
			fields["rxRate_avg"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps

			fields["rxMulticastBytes"]					= rfstat.ast_rx_mcast_bytes
			fields["rxMulticastPackets"]					= rfstat.ast_rx_mcast
			fields["rxBcastPackets"]					= rfstat.ast_rx_bcast
			fields["rxBcastBytes"]						= rfstat.ast_rx_bcast_bytes

			fields["bsSpCnt"]						= hddStat.bs_sp_cnt
			fields["snrSpCnt"]						= hddStat.snr_sp_cnt
			fields["snAnswerCnt"]						= reportGetDiff(hddStat.sn_answer_cnt, hddStat.sn_answer_cnt)
			fields["rxPrbSpCnt"]						= rfstat.is_rx_hdd_probe_sup
			fields["rxAuthCnt"]						= rfstat.is_rx_hdd_auth_sup

			fields["txBitrateSuc"]						= rfstat.ast_tx_rix_invalids
			fields["rxBitrateSuc"]						= rfstat.ast_rx_rix_invalids

			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_@%d_rxRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_@%d_rxRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_@%d_rxRateStats",i)
				if (rf_report.rx_bit_rate[i].kbps != 0) {
					fields[kbps]					= rf_report.rx_bit_rate[i].kbps
				}
				if (rf_report.rx_bit_rate[i].rate_dtn != 0) {
					fields[rateDtn]					= rf_report.rx_bit_rate[i].rate_dtn
				}
				if (rf_report.rx_bit_rate[i].rate_suc_dtn != 0) {
					fields[rateSucDtn]				= rf_report.rx_bit_rate[i].rate_suc_dtn
				}
			}


			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_@%d_txRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_@%d_txRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_@%d_txRateStats",i)
				if (rf_report.tx_bit_rate[i].kbps != 0) {
					fields[kbps]					= rf_report.tx_bit_rate[i].kbps
				}
				if (rf_report.tx_bit_rate[i].rate_dtn != 0) {
					fields[rateDtn]					= rf_report.tx_bit_rate[i].rate_dtn
				}
				if (rf_report.tx_bit_rate[i].rate_suc_dtn != 0) {
					fields[rateSucDtn]				= rf_report.tx_bit_rate[i].rate_suc_dtn
				}
			}

			fields["clientCount"]						= t.numclient[ii]
			fields["lbSpCnt"]							= hddStat.lb_sp_cnt
			fields["rxProbeSup"]						= rfstat.is_rx_hdd_probe_sup
			fields["rxUnicastPackets"]					= rfstat.ast_rx_rate_stats[0].ns_unicasts
			fields["channel"]							= chann

			acc.AddGauge("RfStats", fields, nil)

			var s string
			s = "Stats of interface " + intfName + "\n\n"
			dumpOutput(RF_STAT_OUT_FILE, s, 1)
			prepareAndDumpOutput(RF_STAT_OUT_FILE, fields)

			log.Printf("ah_wireless: radio status is processed")

			t.last_rf_stat[ii] = rfstat
			ii++
		}

		return nil
}

func Gather_Client_Stat(t *Ah_wireless, acc telegraf.Accumulator) error {
	tags := map[string]string{
	}


	fields2 := map[string]interface{}{
	}

	var ii int
	var client_mac string
	ii = 0

    currentTime := time.Now()
    for _, intfName2 := range t.Ifname {

		var cltstat ah_ieee80211_get_wifi_sta_stats
		var cltcfg ieee80211req_cfg_sta 
		var ifindex2 int
		var numassoc int
		var stainfo ah_ieee80211_sta_info
		var stacfg ieee80211req_cfg_one_sta

		var tot_rx_tx uint32
		var tot_rate_frame uint32
		var tot_pcnt int64
		var conn_score int64
		var tx_total		int64
		var rx_total		int64
		var tx_retries		uint32
		//var tx_retry_rate	uchar
		var idx				int
		var tmp_count1		int32
		var tmp_count2		int32
		var tmp_count3		uint32
		var tmp_count4		uint32
		var tmp_count5		uint64
		var tmp_count6		uint64
		var rf_report		ah_dcd_stats_report_int_data
		var tot_tx_bitrate_retries uint32
		var tot_rx_bitrate_retries	uint32

		var client_ssid string

		numassoc = int(getNumAssocs(t.fd, intfName2))

		t.numclient[ii] = numassoc

		if(numassoc == 0) {
			ii++
			continue
		}

		/* Dynamically resize last_clt_stat based on number of clients connected */
		if t.last_clt_stat[ii] == nil || len(t.last_clt_stat[ii]) != numassoc {
			t.last_clt_stat[ii] = make([]ah_ieee80211_sta_stats_item, numassoc)
		}

		clt_item := make([]ah_ieee80211_sta_stats_item, numassoc)


		ifindex2 = getIfIndex(t.fd, intfName2)
		if(ifindex2 <= 0 ) {
			continue
		}

		cltcfg.wifi_sta_stats.count = uint16(numassoc)
		cltcfg.wifi_sta_stats.pointer = unsafe.Pointer(&clt_item[0])

		cltstat = getStaStat(t.fd, intfName2, cltcfg)

		for cn := 0; cn < numassoc; cn++ {

			/* Declare clt_sq at the beginning of client loop to ensure it's in scope */
			var clt_sq [AH_SQ_TYPE_MAX][AH_SQ_GROUP_MAX]ah_signal_quality_stats

			//Re initialiing all the temp variable for the next iteration
			tot_rx_tx		= 0
			tot_rate_frame		= 0
			tot_pcnt		= 0
			conn_score		= 0
			tx_total		= 0
			rx_total		= 0
			tx_retries		= 0

			tmp_count1		= 0
			tmp_count2		= 0
			tmp_count3		= 0
			tmp_count4		= 0
			tmp_count5		= 0
			tmp_count6		= 0

			tot_tx_bitrate_retries  = 0
			tot_rx_bitrate_retries	= 0


			client_ssid = strings.TrimSpace(string(bytes.Trim(clt_item[cn].ns_ssid[:], "\x00")))

			if idx := strings.IndexByte(client_ssid, '\x00'); idx >= 0 {
				client_ssid = client_ssid[:idx]
			}

			if(clt_item[cn].ns_mac[0] !=0 || clt_item[cn].ns_mac[1] !=0 || clt_item[cn].ns_mac[2] !=0 || clt_item[cn].ns_mac[3] !=0 || clt_item[cn].ns_mac[4] != 0 || clt_item[cn].ns_mac[5]!=0) {
				cintfName := t.intf_m[intfName2][client_ssid]
				client_mac = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])


				stainfo = getOneStaInfo(t.fd, cintfName, clt_item[cn].ns_mac, stacfg)

				if stainfo.rssi == 0 {
					continue
				}
			} else {
				continue
			}

			f := init_fe()
			ipnet_score := getFeIpnetScore(f.Fd(), clt_item[cn].ns_mac)
			var ipmsg ah_fw_dev_ip_msg
			sta_ip := getFeServerIp(f.Fd(), clt_item[cn].ns_mac, ipmsg)
			f.Close()

			//client_mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])
			var onestcfg ieee80211req_cfg_one_sta_info
			cfgptr := getOneSta(t.fd, intfName2, clt_item[cn].ns_mac, onestcfg)

			if(cfgptr == nil) {
				continue
			}

			/* Calculation for Signal Quality as per DCD stats */
			var changed bool
			changed = false

			if(t.last_sq[client_mac] == nil) {
				t.last_sq[client_mac] = make(map[int]map[int]ah_signal_quality_stats)

				for i := 0; i < AH_SQ_TYPE_MAX; i++ {
					for j := 0; j < AH_SQ_GROUP_MAX; j++ {
						clt_sq[i][j].asqrange.min = clt_item[cn].ns_sq_group[i][j].asqrange.min
						clt_sq[i][j].asqrange.max = clt_item[cn].ns_sq_group[i][j].asqrange.max
						clt_sq[i][j].count = clt_item[cn].ns_sq_group[i][j].count
					}
				}

			} else {

				for i := 0; i < AH_SQ_TYPE_MAX; i++ {
					for j := 0; j < AH_SQ_GROUP_MAX; j++ {
						if ((t.last_sq[client_mac][i][j].asqrange.min != clt_item[cn].ns_sq_group[i][j].asqrange.min) ||
							(t.last_sq[client_mac][i][j].asqrange.max != clt_item[cn].ns_sq_group[i][j].asqrange.max)) {
							/* the range is changed, just reset snapshot */
							changed = true;
							break;
						}
					}
				}

				for i := 0; i < AH_SQ_TYPE_MAX; i++ {
					for j := 0; j < AH_SQ_GROUP_MAX; j++ {
						clt_sq[i][j].asqrange.min = clt_item[cn].ns_sq_group[i][j].asqrange.min
						clt_sq[i][j].asqrange.max = clt_item[cn].ns_sq_group[i][j].asqrange.max
						if (changed) {
							clt_sq[i][j].count = clt_item[cn].ns_sq_group[i][j].count;
						} else if (clt_item[cn].ns_sq_group[i][j].count > t.last_sq[client_mac][i][j].count) {
							clt_sq[i][j].count = clt_item[cn].ns_sq_group[i][j].count - t.last_sq[client_mac][i][j].count;
						} else {
							clt_sq[i][j].count = 0;
						}
					}
				}

			}

			for i := 0; i < AH_SQ_TYPE_MAX; i++ {
				t.last_sq[client_mac][i] = make(map[int]ah_signal_quality_stats)
				for j := 0; j < AH_SQ_GROUP_MAX; j++ {
					t.last_sq[client_mac][i][j] = clt_item[cn].ns_sq_group[i][j]
				}
			}

			/* Calculation for Signal Quality as per DCD stats end */

			var onesta *ieee80211req_sta_info = (*ieee80211req_sta_info)(cfgptr)

			var clt_last_stats saved_stats = t.entity[intfName2][client_mac]

			/* We need check and aggregation Tx/Rx bit rate distribution
 			* prcentage, if the bit rate equal in radio interface or client reporting.
 			*/

			for i := 0; i < NS_HW_RATE_SIZE; i++{

				if ((clt_item[cn].ns_rx_rate_stats[i].ns_rateKbps == 0) && (clt_item[cn].ns_tx_rate_stats[i].ns_rateKbps == 0)) {
					continue
				}

				for j := (i+1); j < NS_HW_RATE_SIZE; j++ {
					if ((clt_item[cn].ns_rx_rate_stats[i].ns_rateKbps != 0) && (clt_item[cn].ns_rx_rate_stats[i].ns_rateKbps == clt_item[cn].ns_rx_rate_stats[j].ns_rateKbps)) {
						clt_item[cn].ns_rx_rate_stats[i].ns_unicasts += clt_item[cn].ns_rx_rate_stats[j].ns_unicasts;
						clt_item[cn].ns_rx_rate_stats[i].ns_retries += clt_item[cn].ns_rx_rate_stats[j].ns_retries;
						clt_item[cn].ns_rx_rate_stats[j].ns_rateKbps = 0;
						clt_item[cn].ns_rx_rate_stats[j].ns_unicasts = 0;
						clt_item[cn].ns_rx_rate_stats[j].ns_retries = 0;
					}
					if ((clt_item[cn].ns_tx_rate_stats[i].ns_rateKbps != 0) && (clt_item[cn].ns_tx_rate_stats[i].ns_rateKbps == clt_item[cn].ns_tx_rate_stats[j].ns_rateKbps)) {
						clt_item[cn].ns_tx_rate_stats[i].ns_unicasts += clt_item[cn].ns_tx_rate_stats[j].ns_unicasts;
						clt_item[cn].ns_tx_rate_stats[i].ns_retries += clt_item[cn].ns_tx_rate_stats[j].ns_retries;
						clt_item[cn].ns_tx_rate_stats[j].ns_rateKbps = 0;
						clt_item[cn].ns_tx_rate_stats[j].ns_unicasts = 0;
						clt_item[cn].ns_tx_rate_stats[j].ns_retries = 0;
					}
				}
			}

			/* Rate stat from DCD */
			for idx = 0; idx < NS_HW_RATE_SIZE; idx++ {
				if ( clt_item[cn].ns_tx_rate_stats[idx].ns_unicasts > t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_unicasts) {
					tx_total += int64(clt_item[cn].ns_tx_rate_stats[idx].ns_unicasts - t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_unicasts)
				}

				if (clt_item[cn].ns_rx_rate_stats[idx].ns_unicasts > t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_unicasts) {
					rx_total += int64(clt_item[cn].ns_rx_rate_stats[idx].ns_unicasts - t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_unicasts)
				}

				if (clt_item[cn].ns_tx_rate_stats[idx].ns_retries > t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_retries) {
					tot_tx_bitrate_retries += clt_item[cn].ns_tx_rate_stats[idx].ns_retries - t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_retries;
				}

				if (clt_item[cn].ns_rx_rate_stats[idx].ns_retries > t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_retries) {
					tot_rx_bitrate_retries += clt_item[cn].ns_rx_rate_stats[idx].ns_retries - t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_retries;
				}
			}

			tx_ok := tx_total
			rx_ok := rx_total

			tot_unicasts := rx_ok + tx_ok
			if tot_unicasts > 100 {
				tot_unicasts /= 100
			} else {
				tot_unicasts = 1
			}
			/* Tx/Rx bit rate distribution */
			for idx = 0; idx < NS_HW_RATE_SIZE; idx++ {
				tmp_count1 = int32(clt_item[cn].ns_tx_rate_stats[idx].ns_unicasts - t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_unicasts)
				if (tx_total > 0 && tmp_count1 > 0) {
					rf_report.tx_bit_rate[idx].rate_dtn = uint8((int64(tmp_count1) * 100) / tx_total)
				} else {
					rf_report.tx_bit_rate[idx].rate_dtn = 0;
				}

				tmp_count2 = int32(clt_item[cn].ns_rx_rate_stats[idx].ns_unicasts - t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_unicasts)
				if (rx_total > 0 && tmp_count2 > 0) {
					rf_report.rx_bit_rate[idx].rate_dtn = uint8((int64(tmp_count2) * 100) / rx_total)
				} else {
					rf_report.rx_bit_rate[idx].rate_dtn = 0;
				}

				/* Tx/Rx bit rate success distribution */
				tmp_count3 = uint32(clt_item[cn].ns_tx_rate_stats[idx].ns_retries - t.last_clt_stat[ii][cn].ns_tx_rate_stats[idx].ns_retries)
				tmp_count5 = uint64(tmp_count1) + uint64(tmp_count3)
				if (tmp_count5 > 0 && rf_report.tx_bit_rate[idx].rate_dtn > 0) {
					rf_report.tx_bit_rate[idx].rate_suc_dtn = uint8((uint64(tmp_count1) * 100) / tmp_count5)
					if (rf_report.tx_bit_rate[idx].rate_suc_dtn > 100) {
						rf_report.tx_bit_rate[idx].rate_suc_dtn = 100
						log.Printf("stats report client data process: rate_suc_dtn1 is more than 100%\n")
					}
				} else {
					rf_report.tx_bit_rate[idx].rate_suc_dtn = 0;
				}
				tx_retries += tmp_count3;

				tmp_count4 = clt_item[cn].ns_rx_rate_stats[idx].ns_retries - t.last_clt_stat[ii][cn].ns_rx_rate_stats[idx].ns_retries
				tmp_count6 = uint64(tmp_count2) + uint64(tmp_count4)
				if (tmp_count6 > 0 && rf_report.rx_bit_rate[idx].rate_dtn > 0) {
					rf_report.rx_bit_rate[idx].rate_suc_dtn = uint8((uint64(tmp_count2) * 100) / tmp_count6)
					if (rf_report.rx_bit_rate[idx].rate_suc_dtn > 100) {
						rf_report.rx_bit_rate[idx].rate_suc_dtn = 100
						log.Printf("stats report client data process: rate_suc_dtn2 is more than 100%\n")
					}
				} else {
					rf_report.rx_bit_rate[idx].rate_suc_dtn = 0
				}

				rf_report.tx_bit_rate[idx].kbps = clt_item[cn].ns_tx_rate_stats[idx].ns_rateKbps
				rf_report.rx_bit_rate[idx].kbps = clt_item[cn].ns_rx_rate_stats[idx].ns_rateKbps
			}
			/* Rate stat from DCD */

			for i := 0; i < NS_HW_RATE_SIZE; i++ {
				if clt_item[cn].ns_rx_rate_stats[i].ns_unicasts != 0 || clt_item[cn].ns_tx_rate_stats[i].ns_unicasts != 0 {
					if clt_item[cn].ns_tx_rate_stats[i].ns_unicasts >= t.last_clt_stat[ii][cn].ns_tx_rate_stats[i].ns_unicasts ||
					   clt_item[cn].ns_rx_rate_stats[i].ns_unicasts >= t.last_clt_stat[ii][cn].ns_rx_rate_stats[i].ns_unicasts {
						tot_rx_tx = (clt_item[cn].ns_tx_rate_stats[i].ns_unicasts + clt_item[cn].ns_rx_rate_stats[i].ns_unicasts +
						             clt_item[cn].ns_rx_rate_stats[i].ns_retries + clt_item[cn].ns_tx_rate_stats[i].ns_retries) -
						             (t.last_clt_stat[ii][cn].ns_tx_rate_stats[i].ns_unicasts +
							      t.last_clt_stat[ii][cn].ns_rx_rate_stats[i].ns_unicasts +
						              t.last_clt_stat[ii][cn].ns_rx_rate_stats[i].ns_retries +
							      t.last_clt_stat[ii][cn].ns_tx_rate_stats[i].ns_retries)
					} else {
						tot_rx_tx = 0
					}
					tot_rate_frame += tot_rx_tx

					if tot_rx_tx > 100 {
						tot_rx_tx /= 100
					} else {
						tot_rx_tx = 1
					}
					success_count := (clt_item[cn].ns_rx_rate_stats[i].ns_unicasts + clt_item[cn].ns_tx_rate_stats[i].ns_unicasts) -
					(t.last_clt_stat[ii][cn].ns_rx_rate_stats[i].ns_unicasts + t.last_clt_stat[ii][cn].ns_tx_rate_stats[i].ns_unicasts)
					success := success_count / tot_rx_tx

					if tot_unicasts != 0 {
						tot_pcnt = int64(success_count) / tot_unicasts
					}

					rate_score := (clt_item[cn].ns_tx_rate_stats[i].ns_rateKbps) / 1000
					conn_score += (int64(rate_score) * int64(success) * tot_pcnt)
				}
			}

			var rssi int
			var radio_link_score int64
			rssi = int(stainfo.rssi) + int(stainfo.noise_floor)

			var tmp_count1 int64
			if tot_rate_frame > (600 * 20) {
				if clt_item[cn].ns_sla_bm_score > 0 {
					tmp_count1 = (50 * conn_score) / int64(clt_item[cn].ns_sla_bm_score)
					if tmp_count1 > AH_DCD_CLT_SCORE_GOOD {
						radio_link_score = AH_DCD_CLT_SCORE_GOOD
						fmt.Println("radio link score is more than 100")
					} else {
						radio_link_score = tmp_count1
					}
				} else {
					radio_link_score = AH_DCD_CLT_SCORE_GOOD
				}
			} else {
				if rssi <= -85 {
					radio_link_score = AH_DCD_CLT_SCORE_POOR
				} else if rssi < -70 {
					radio_link_score = AH_DCD_CLT_SCORE_ACCEPTABLE
				} else {
					radio_link_score = AH_DCD_CLT_SCORE_GOOD
				}
			}

			// radio link score is hard coded for AP5020
			radio_link_score = AH_DCD_CLT_SCORE_GOOD


			t.last_clt_stat[ii][cn] = clt_item[cn]

			var rt_sta rt_sta_data
			rt_sta = get_rt_sta_info(t, client_mac, int(onesta.isi_upid), rt_sta)

			fields2["ifName"]			= intfName2
			fields2["ifIndex"]			= ifindex2
			fields2["rrmId"]			= rrmid
			fields2["channel"]			= freqToChan(onesta.isi_freq)
			fields2["channelWidth"]		= getChannelWidth(onesta.isi_phymode)

			fields2["mac_keys"]			= client_mac
//			fields2["alarmFlag"]		= t.last_alarm[ii].alarm
			fields2["number"]			= cltstat.count
			fields2["ssid"]				= client_ssid
			fields2["txPackets"]		= stainfo.tx_pkts
			fields2["txBytes"]			= stainfo.tx_bytes
			fields2["txDrop"]			= clt_item[cn].ns_tx_drops
			fields2["slaDrop"]			= clt_item[cn].ns_sla_traps
			fields2["rxPackets"]		= stainfo.rx_pkts
			fields2["rxBytes"]			= stainfo.rx_bytes
			fields2["rxDrop"]			= clt_item[cn].ns_tx_drops
			fields2["avgSnr"]			= clt_item[cn].ns_snr
			fields2["psTimes"]			= clt_item[cn].ns_ps_times
			fields2["radioScore"]		= radio_link_score
			fields2["ipNetScore"]		= ipnet_score
			if ipnet_score == 0 {
				fields2["appScore"]		= ipnet_score
			} else {
				fields2["appScore"]		= clt_item[cn].ns_app_health_score
			}
			fields2["phyMode"]			= getMacProtoMode(onesta.isi_phymode)

			fields2["rssi"]				= int(stainfo.rssi) + int(stainfo.noise_floor)

			fields2["os"]				= rt_sta.os
			fields2["name"]				= strings.ReplaceAll(string(onesta.isi_name[:]), "\u0000", "")
			fields2["host"]				= rt_sta.hostname
			fields2["profName"]			= rt_sta.userprofile
			fields2["dhcpIp"]			= intToIp(sta_ip.dhcp_server)
			fields2["gwIp"]				= intToIp(sta_ip.gateway)
			fields2["dnsIp"]			= intToIp(sta_ip.dns[0].dns_ip)
			fields2["clientIp"]			= intToIp(sta_ip.client_static_ip)
			fields2["dhcpTime"]			= sta_ip.dhcp_time
//			fields2["gwTime"]			= 0							/* TBD (Needs shared memory of auth2)     */
			fields2["dnsTime"]			= sta_ip.dns[0].dns_response_time
			fields2["clientTime"]		= onesta.isi_assoc_time


			for i := 0; i < AH_TX_NSS_MAX; i++{
				txNssUsage := fmt.Sprintf("@%d_txNssUsage",i)
				fields2[txNssUsage]           = clt_item[cn].ns_tx_nss[i]
			}



			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_@%d_rxRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_@%d_rxRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_@%d_rxRateStats",i)
				if (rf_report.rx_bit_rate[i].kbps != 0) {
					fields2[kbps]			= rf_report.rx_bit_rate[i].kbps
				}
				if (rf_report.rx_bit_rate[i].rate_dtn != 0) {
					fields2[rateDtn]		= rf_report.rx_bit_rate[i].rate_dtn
				}
				if (rf_report.rx_bit_rate[i].rate_suc_dtn != 0) {
					fields2[rateSucDtn]		= rf_report.rx_bit_rate[i].rate_suc_dtn
				}
			}


			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_@%d_txRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_@%d_txRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_@%d_txRateStats",i)
				if (rf_report.tx_bit_rate[i].kbps != 0) {
					fields2[kbps]			= rf_report.tx_bit_rate[i].kbps
				}
				if (rf_report.tx_bit_rate[i].rate_dtn != 0) {
					fields2[rateDtn]		= rf_report.tx_bit_rate[i].rate_dtn
				}
				if (rf_report.tx_bit_rate[i].rate_suc_dtn != 0) {
					fields2[rateSucDtn]		= rf_report.tx_bit_rate[i].rate_suc_dtn
				}
			}


			fields2["txAirtime_min"]				= clt_last_stats.tx_airtime_min 
			fields2["txAirtime_max"]				= clt_last_stats.tx_airtime_max 
			fields2["txAirtime_avg"]				= clt_last_stats.tx_airtime_average

			fields2["rxAirtime_min"]				= clt_last_stats.rx_airtime_min
			fields2["rxAirtime_max"]				= clt_last_stats.rx_airtime_max
			fields2["rxAirtime_avg"]				= clt_last_stats.rx_airtime_average

			fields2["bwUsage_min"]					= clt_last_stats.bw_usage_min
			fields2["bwUsage_max"]					= clt_last_stats.bw_usage_max
			fields2["bwUsage_avg"]					= clt_last_stats.bw_usage_average

			for i := 0; i < AH_SQ_GROUP_MAX; i++{
				rangeMin	:=	fmt.Sprintf("rangeMin_@%d_sqRssi",i)
				rangeMax	:=	fmt.Sprintf("rangeMax_@%d_sqRssi",i)
				countt		:=	fmt.Sprintf("count_@%d_sqRssi",i)
				bucket		:=	fmt.Sprintf("bucketNum_@%d_sqRssi",i)
				fields2[rangeMin]		= clt_sq[0][i].asqrange.min
				fields2[rangeMax]		= clt_sq[0][i].asqrange.max
				fields2[countt]			= clt_sq[0][i].count
				fields2[bucket]			= i
			}

			for i := 0; i < AH_SQ_GROUP_MAX; i++{
				rangeMin	:=	fmt.Sprintf("rangeMin_@%d_sqNoise",i)
				rangeMax	:=	fmt.Sprintf("rangeMax_@%d_sqNoise",i)
				countt		:=	fmt.Sprintf("count_@%d_sqNoise",i)
				bucket		:=	fmt.Sprintf("bucketNum_@%d_sqNoise",i)
				fields2[rangeMin]		= clt_sq[1][i].asqrange.min
				fields2[rangeMax]		= clt_sq[1][i].asqrange.max
				fields2[countt]			= clt_sq[1][i].count
				fields2[bucket]			= i
			}

			for i := 0; i < AH_SQ_GROUP_MAX; i++{
				rangeMin	:=	fmt.Sprintf("rangeMin_@%d_sqSnr",i)
				rangeMax	:=	fmt.Sprintf("rangeMax_@%d_sqSnr",i)
				countt		:=	fmt.Sprintf("count_@%d_sqSnr",i)
				bucket		:=	fmt.Sprintf("bucketNum_@%d_sqSnr",i)
				fields2[rangeMin]			= clt_sq[2][i].asqrange.min
				fields2[rangeMax]			= clt_sq[2][i].asqrange.max
				fields2[countt]				= clt_sq[2][i].count
				fields2[bucket]				= i
			}

			acc.AddFields("ClientStats", fields2, tags, currentTime)


			clt_new_stats := saved_stats{}
			clt_new_stats.tx_airtime = clt_item[cn].ns_tx_airtime
			clt_new_stats.rx_airtime = clt_item[cn].ns_rx_airtime
			t.entity[intfName2][client_mac] = clt_new_stats


			var s string
			s = "Stats of client [" + client_mac + "]\n\n"
			dumpOutput(CLT_STAT_OUT_FILE, s, 1)
			prepareAndDumpOutput(CLT_STAT_OUT_FILE, fields2)

		}
		ii++

	}

	log.Printf("ah_wireless_v2: client status is processed")

	return nil
}

func Gather_AirTime(t *Ah_wireless, acc telegraf.Accumulator) error {

	for _, intfName2 := range t.Ifname {

		var numassoc1 int
		var client_mac1 string
		var cltcfg ieee80211req_cfg_sta 

		numassoc1 = int(getNumAssocs(t.fd, intfName2))

		if(numassoc1 == 0) {
			continue
		}

		log.Printf("ah_wireless: calculating airtime for %d sta\n",numassoc1)

		clt_item := make([]ah_ieee80211_sta_stats_item, numassoc1)

		cltcfg.wifi_sta_stats.count = uint16(numassoc1)
		cltcfg.wifi_sta_stats.pointer = unsafe.Pointer(&clt_item[0])
		getStaStat(t.fd, intfName2, cltcfg)

		for cn := 0; cn < numassoc1; cn++ {

			if(clt_item[cn].ns_mac[0] !=0 || clt_item[cn].ns_mac[1] !=0 || clt_item[cn].ns_mac[2] !=0 || clt_item[cn].ns_mac[3] !=0 || clt_item[cn].ns_mac[4] != 0 || clt_item[cn].ns_mac[5]!=0) {

				client_mac1 = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])
			} else {

				continue
			}

			var clt_last_stats saved_stats = t.entity[intfName2][client_mac1]

			clt_new_stats := saved_stats{}

			/* Calculate tx airtime min, max, average */

			if ((clt_last_stats.tx_airtime_min > clt_item[cn].ns_tx_airtime) || (clt_last_stats.tx_airtime_min == 0) ) {
				clt_new_stats.tx_airtime_min = clt_item[cn].ns_tx_airtime - clt_last_stats.tx_airtime
			} else {
				clt_new_stats.tx_airtime_min = clt_last_stats.tx_airtime_min
			}

			if (clt_last_stats.tx_airtime_max < clt_item[cn].ns_tx_airtime ) {
				clt_new_stats.tx_airtime_max = clt_item[cn].ns_tx_airtime - clt_last_stats.tx_airtime
			} else {
				clt_new_stats.tx_airtime_max = clt_last_stats.tx_airtime_max
			}

			clt_new_stats.tx_airtime_average = ((clt_last_stats.tx_airtime_average + clt_new_stats.tx_airtime_min + clt_new_stats.tx_airtime_max)/3)

			/* Calculate rx airtime min, max, average */

			if ((clt_last_stats.rx_airtime_min > clt_item[cn].ns_rx_airtime) || (clt_last_stats.rx_airtime_min == 0) ) {
				clt_new_stats.rx_airtime_min = clt_item[cn].ns_rx_airtime - clt_last_stats.rx_airtime
			} else {
				clt_new_stats.rx_airtime_min = clt_last_stats.rx_airtime_min
			}

			if (clt_last_stats.rx_airtime_max < clt_item[cn].ns_rx_airtime ) {
				clt_new_stats.rx_airtime_max = clt_item[cn].ns_rx_airtime - clt_last_stats.rx_airtime
			} else {
				clt_new_stats.rx_airtime_max = clt_last_stats.rx_airtime_max
			}

			clt_new_stats.rx_airtime_average = ((clt_last_stats.rx_airtime_average + clt_new_stats.rx_airtime_min + clt_new_stats.rx_airtime_max)/3)


				clt_new_stats.tx_airtime_min	= ((clt_new_stats.tx_airtime_min /10) / (60 *1000))
				clt_new_stats.tx_airtime_max	= ((clt_new_stats.tx_airtime_max /10) / (60 *1000))
				clt_new_stats.tx_airtime_average	= ((clt_new_stats.tx_airtime_average /10) / (60 *1000))

				clt_new_stats.rx_airtime_min	= ((clt_new_stats.rx_airtime_min /10) / (60 *1000))
				clt_new_stats.rx_airtime_max	= ((clt_new_stats.rx_airtime_max /10) / (60 *1000))
				clt_new_stats.rx_airtime_average	= ((clt_new_stats.rx_airtime_average /10) / (60 *1000))

				if (clt_new_stats.tx_airtime_min > 100) {
					clt_new_stats.tx_airtime_min = 100
				}

				if (clt_new_stats.tx_airtime_max > 100) {
					clt_new_stats.tx_airtime_max = 100
				}

				if (clt_new_stats.tx_airtime_average > 100) {
					clt_new_stats.tx_airtime_average = 100
				}

				if (clt_new_stats.rx_airtime_min > 100) {
					clt_new_stats.rx_airtime_min = 100
				}

				if (clt_new_stats.rx_airtime_max > 100) {
					clt_new_stats.rx_airtime_max = 100
				}

				if (clt_new_stats.rx_airtime_average > 100) {
					clt_new_stats.rx_airtime_average = 100
				}

			/* Calculate bandwidth usage min, max, average */

			bw_usage := (((clt_item[cn].ns_tx_bytes + clt_item[cn].ns_rx_bytes) * 8) / (60)) / 1000;

			if ((clt_last_stats.bw_usage_min > bw_usage) || (clt_last_stats.bw_usage_min == 0)) {
				clt_new_stats.bw_usage_min = bw_usage
			} else {
				clt_new_stats.bw_usage_min = clt_last_stats.bw_usage_min
			}

			if (clt_last_stats.bw_usage_max < bw_usage) {
				clt_new_stats.bw_usage_max = bw_usage
			} else {
				clt_new_stats.bw_usage_max = clt_last_stats.bw_usage_max
			}

			clt_new_stats.tx_airtime = clt_last_stats.tx_airtime
			clt_new_stats.rx_airtime = clt_last_stats.rx_airtime

			clt_new_stats.bw_usage_average = ((clt_last_stats.bw_usage_average + clt_new_stats.bw_usage_min + clt_new_stats.bw_usage_max)/3)
			t.entity[intfName2][client_mac1] = clt_new_stats
		}

	}
	log.Printf("telegraf ah_wireless calculated airtime avg!\n")

	return nil
}

func Gather_EthernetInterfaceStats(t *Ah_wireless) error {

	var ethdevstats ah_dcd_dev_stats

	interfaces := []string{}

        for i := 0; i < AH_MAX_ETH; i++ {
                interfaces = append(interfaces, fmt.Sprintf("eth%d", i))
        }

        interfaces = append(interfaces, "agg0", "red0")


        for i, ifName := range interfaces{

		ethdevstats = getProcNetDev(ifName)

		t.if_stats[i].ifname 			= ifName

		t.if_stats[i].rx_unicast		= reportGetDiff64(uint64(ethdevstats.rx_unicast), t.if_stats[i].rx_unicast)
		t.if_stats[i].rx_broadcast		= reportGetDiff64(uint64(ethdevstats.rx_broadcast), t.if_stats[i].rx_broadcast)
		t.if_stats[i].rx_multicast		= reportGetDiff64(uint64(ethdevstats.rx_multicast), t.if_stats[i].rx_multicast)
		t.if_stats[i].rx_bytes                  = reportGetDiff64(uint64(ethdevstats.rx_bytes), t.if_stats[i].rx_bytes)
		t.if_stats[i].rx_errors                 = reportGetDiff64(uint64(ethdevstats.rx_errors), t.if_stats[i].rx_errors)
		t.if_stats[i].rx_dropped                = reportGetDiff64(uint64(ethdevstats.rx_dropped), t.if_stats[i].rx_dropped)
		t.if_stats[i].tx_unicast		= reportGetDiff64(uint64(ethdevstats.tx_unicast), t.if_stats[i].tx_unicast)
		t.if_stats[i].tx_broadcast		= reportGetDiff64(uint64(ethdevstats.tx_broadcast), t.if_stats[i].tx_broadcast)
		t.if_stats[i].tx_multicast		= reportGetDiff64(uint64(ethdevstats.tx_multicast), t.if_stats[i].tx_multicast)
		t.if_stats[i].tx_bytes                  = reportGetDiff64(uint64(ethdevstats.tx_bytes), t.if_stats[i].tx_bytes)
		t.if_stats[i].tx_errors                 = reportGetDiff64(uint64(ethdevstats.tx_errors), t.if_stats[i].tx_errors)
		t.if_stats[i].tx_dropped                = reportGetDiff64(uint64(ethdevstats.tx_dropped), t.if_stats[i].tx_dropped)
		t.ethx_stats[i].ifname                  = ifName


                var link_status, eth_status, speed, duplex int32

                if ifName == "agg0" || ifName == "red0" {
                        var memberIfstatus int32
                        var maxMemberSpeed int32  = ETH_MII_LINK_DOWN
                        var maxMemberDuplex int32 = ETH_MII_LINK_DOWN

                        ifStatus := getIfStatus(t.fd, ifName)

                        if (ifStatus & IFF_UP) != 0 && (ifStatus & IFF_RUNNING) != 0 {
                                link_status = ETH_SET_MII_LINK_UP
				memberIfstatus = AH_IF_STATUS
                        } else {
                                link_status = ETH_SET_MII_LINK_DOWN
				memberIfstatus = ETH_MII_LINK_DOWN
                        }

                        if ( link_status == ETH_SET_MII_LINK_DOWN){
                            t.ethx_stats[i].duplex = "LINK_DOWN"
                            t.ethx_stats[i].speed = "LINK_DOWN"
                            continue
                        }

                        if memberIfstatus != ETH_MII_LINK_DOWN {
                            speed := memberIfstatus & ETH_MII_SPEED_MASK
                            duplex := memberIfstatus & ETH_MII_DUPLEX_MASK

                            if maxMemberSpeed < speed {
                                maxMemberSpeed = speed
                            }
                            if maxMemberDuplex < duplex {
                                maxMemberDuplex = duplex
                            }
                        }
                        speed = maxMemberSpeed | maxMemberDuplex
                        duplex = speed

	        } else{
			f := init_ethf()
                        link_status = getEthLink(t, f.Fd(), ifName)
                        eth_status = getEthStatus(t, f.Fd(), ifName)
                        f.Close()
                        if link_status == ETH_SET_MII_LINK_DOWN {
                                t.ethx_stats[i].duplex = "LINK_DOWN"
                                t.ethx_stats[i].speed = "LINK_DOWN"
                                continue
                        }

                        speed = eth_status
                        duplex = eth_status
                }

		if((duplex & ETH_MII_DUPLEX_FULL) > 0) {
			t.ethx_stats[i].duplex = "FULL"
		} else {
			t.ethx_stats[i].duplex = "HALF"
		}

		if((speed & ETH_MII_SPEED_10000M) > 0) {
			t.ethx_stats[i].speed = "10000M"
		} else if ((speed & ETH_MII_SPEED_5000M) > 0) {
			t.ethx_stats[i].speed = "5000M"
		} else if ((speed & ETH_MII_SPEED_2500M) > 0) {
			t.ethx_stats[i].speed = "2500M"
		} else if ((speed & ETH_MII_SPEED_1000M) > 0) {
			t.ethx_stats[i].speed = "1000M"
		} else if ((speed & ETH_MII_SPEED_100M) > 0) {
			t.ethx_stats[i].speed = "100M"
		} else {
			t.ethx_stats[i].speed = "10M"
		}


	}

	return nil
}

func Send_NetworkStats(t *Ah_wireless, acc telegraf.Accumulator) error {

	var id int

	id = 0

	for i := 0; i < (AH_MAX_WIRED); i++{

		if ( i >= NETWORK_MAX_COUNT ) {
			return nil
		}

		if(len(t.if_stats[i].ifname) < 2) {
			continue //Invalid interface name
		}

		iface, err := net.InterfaceByName(t.if_stats[i].ifname)
		if err != nil {
			id = 0
			log.Printf("Error getting index\n")
		} else {
			id = iface.Index
		}

		fields := map[string]interface{}{

			"name_keys":					t.if_stats[i].ifname,
			"ifIndex_keys":					id,

		}

		fields["rxUnicastPackets"]		= t.if_stats[i].rx_unicast
		fields["rxMulticastPackets"]	= t.if_stats[i].rx_multicast
		fields["rxBcastPackets"]		= t.if_stats[i].rx_broadcast
		fields["rxBytes"]	                = t.if_stats[i].rx_bytes
		fields["rxErrors"]                      = t.if_stats[i].rx_errors
		fields["rxDropped"]                     = t.if_stats[i].rx_dropped
		fields["txUnicastPackets"]		= t.if_stats[i].tx_unicast
		fields["txMulticastPackets"]	= t.if_stats[i].tx_multicast
		fields["txBcastPackets"]		= t.if_stats[i].tx_broadcast
		fields["txBytes"]                       = t.if_stats[i].tx_bytes
		fields["txErrors"]                      = t.if_stats[i].tx_errors
		fields["txDropped"]                     = t.if_stats[i].tx_dropped

		if len(strings.TrimSpace(t.ethx_stats[i].duplex)) > 0 {
			fields["duplex"]				= t.ethx_stats[i].duplex
		}
		if len(strings.TrimSpace(t.ethx_stats[i].speed)) > 0 {
			fields["speed"]					= t.ethx_stats[i].speed
		}

		acc.AddGauge("NetworkStats", fields, nil)

		var s string
		s = "Stats of interface " + t.if_stats[i].ifname + "\n\n"
		dumpOutput(NW_STAT_OUT_FILE, s, 1)
		prepareAndDumpOutput(NW_STAT_OUT_FILE, fields)
		log.Printf("network status is processed")

	}
	return nil
}

func Send_DeviceStats(t *Ah_wireless, acc telegraf.Accumulator) error {

		fields := map[string]interface{}{

		}

		fields["trackIp"]		= intToIp(uint32(t.nw_health.track_ip))
		fields["trackLatency"]	= t.nw_health.track_latency
		fields["gwIp"]			= intToIp(uint32(t.nw_health.gw_ip))
		fields["gwMac"]			= t.nw_health.gw_mac
		fields["gwLatency"]		= t.nw_health.gw_latency
		fields["gwTtl"]			= t.nw_health.gw_ttl
		fields["txIpv4Packets"]	= t.nw_health.snd_ipv4_packet
		fields["txIpv4Bytes"]	= t.nw_health.snd_ipv4_byte
		fields["txIpv6Packets"]	= t.nw_health.snd_ipv6_packet
		fields["txIpv6Bytes"]	= t.nw_health.snd_ipv6_byte
		fields["rxIpv4Packets"]	= t.nw_health.rec_ipv4_packet
		fields["rxIpv4Bytes"]	= t.nw_health.rec_ipv4_byte
		fields["rxIpv6Packets"]	= t.nw_health.rec_ipv6_packet
		fields["rxIpv6Bytes"]	= t.nw_health.rec_ipv6_byte

		fields["dhcpIp"]		= intToIp(t.nw_service.dhcp_ip)
		fields["dhcpTime"]		= t.nw_service.dhcp_time

		for j := 0; j < 16; j++{
			if t.nw_service.dns_ip[j] > 0 {
				dnsip	:=	fmt.Sprintf("dnsIp_@%d_dnsServer",j)
				dnstime	:=	fmt.Sprintf("dnsTime_@%d_dnsServer",j)

				fields[dnsip]		= intToIp(t.nw_service.dns_ip[j])
				fields[dnstime]	= t.nw_service.dns_time[j]
			}
		}

		fields["ntpServer"]		= t.nw_service.ntp_server
		fields["ntpLatency"]	= t.nw_service.ntp_latency

		for j := 0; j < int(t.nw_service.syslog_sev_num); j++{
			server	:=	fmt.Sprintf("name_@%d_syslogServer",j)
			latency	:=	fmt.Sprintf("latency_@%d_syslogServer",j)

			fields[server]		= t.nw_service.syslog_server[j]
			fields[latency]	= t.nw_service.syslog_latency[j]

		}

		for j := 0; j < int(t.nw_service.cwp_external_num); j++{
			server	:=	fmt.Sprintf("name_@%d_cwpServer",j)
			latency	:=	fmt.Sprintf("latency_@%d_cwpServer",j)

			fields[server]		= t.nw_service.cwp_external_name[j]
			fields[latency]	= t.nw_service.cwp_latency[j]

		}

		for j := 0; j < int(t.nw_service.radius_sev_num); j++{
			if j < AH_MAX_RADIUS_NUM {
				server	:=	fmt.Sprintf("name_@%d_radiusServer",j)
				latency	:=	fmt.Sprintf("latency_@%d_radiusServer",j)

				fields[server]		= t.nw_service.radius_server[j]
				fields[latency]	= t.nw_service.radius_latency[j]
			}
		}

		acc.AddGauge("DeviceStats", fields, nil)

		var s string
		s = "-----------------------------------------------\n\n"
		dumpOutput(DEV_STAT_OUT_FILE, s, 1)
		prepareAndDumpOutput(DEV_STAT_OUT_FILE, fields)
		log.Printf("device status is processed")
		return nil
}

func Gather_Network_Health(t *Ah_wireless) error {

	table, err := os.ReadFile("/tmp/dcd_stat_network_health")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	var stats = new(network_health_data)

    fmt.Sscanf(string(lines[0]),
    "%d %d %d %s %d %d %d %d %d %d %d %d %d %d %d",

                           &stats.track_ip,
                           &stats.track_latency,
                           &stats.gw_ip,
                           &stats.gw_mac,
                           &stats.gw_latency,
                           &stats.gw_ttl,
                           &stats.if_data_num,
                           &stats.rec_ipv4_packet,
                           &stats.rec_ipv4_byte,
                           &stats.rec_ipv6_packet,
                           &stats.rec_ipv6_byte,
                           &stats.snd_ipv4_packet,
                           &stats.snd_ipv4_byte,
                           &stats.snd_ipv6_packet,
                           &stats.snd_ipv6_byte)


	t.nw_health.track_ip = stats.track_ip
	t.nw_health.track_latency = stats.track_latency
	t.nw_health.gw_ip = stats.gw_ip
	t.nw_health.gw_mac = stats.gw_mac
	t.nw_health.gw_latency = stats.gw_latency
	t.nw_health.gw_ttl = stats.gw_ttl
	t.nw_health.if_data_num = stats.if_data_num
	t.nw_health.rec_ipv4_packet = stats.rec_ipv4_packet
	t.nw_health.rec_ipv4_byte = stats.rec_ipv4_byte
	t.nw_health.rec_ipv6_packet = stats.rec_ipv6_packet
	t.nw_health.rec_ipv6_byte = stats.rec_ipv6_byte
	t.nw_health.snd_ipv4_packet = stats.snd_ipv4_packet
	t.nw_health.snd_ipv4_byte = stats.snd_ipv4_byte
	t.nw_health.snd_ipv6_packet = stats.snd_ipv6_packet
	t.nw_health.snd_ipv6_byte = stats.snd_ipv6_byte

	return nil
}

func get_radius_server_data(t *Ah_wireless) error {
	var j uint8
	var num uint8
	var lent uint8
	var name string
	var lat int32

	table, err := os.ReadFile("/tmp/dcd_stat_radius")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	j = 0
	for  _, curLine := range lines {

        fmt.Sscanf(string(curLine[:]), "%d %d %s %d", &num, &lent, &name, &lat)

		if num >= AH_MAX_RADIUS_NUM { 
			continue
		}

		t.nw_service.radius_sev_num = num

		if (j == num) {
			break
		}

		t.nw_service.radius_sev_len[j] = lent
		t.nw_service.radius_server[j] = strings.Trim(name, "[]")
		t.nw_service.radius_latency[j] = lat
		j++


	}
	return nil
}

func get_cwp_server_data(t *Ah_wireless) error {

	var j uint8
	var num uint8
	var lent uint8
	var name string
	var lat int32

	table, err := os.ReadFile("/tmp/dcd_stat_cwp")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	j = 0
	for  _, curLine := range lines {

        fmt.Sscanf(string(curLine[:]), "%d %d %s %d", &num, &lent, &name, &lat)

		t.nw_service.cwp_external_num = num

		if (j == num) {
			break
		}

		t.nw_service.cwp_external_len[j] = lent
		t.nw_service.cwp_external_name[j] = strings.Trim(name, "[]")
		t.nw_service.cwp_latency[j] = lat
		j++

	}
	return nil
}

func get_syslog_server_data(t *Ah_wireless) error {
	var j uint8
	var num uint8
	var lent uint8
	var name string
	var lat int32

	table, err := os.ReadFile("/tmp/dcd_stat_syslog")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	j = 0
	for  _, curLine := range lines {

        fmt.Sscanf(string(curLine[:]), "%d %d %s %d", &num, &lent, &name, &lat)

		t.nw_service.syslog_sev_num = num

		if (j == num) {
			break
		}

		t.nw_service.syslog_sev_len[j] = lent
		t.nw_service.syslog_server[j] = strings.Trim(name, "[]")
		t.nw_service.syslog_latency[j] = lat
		j++

	}
	return nil
}


func get_network_dhcp_dns_data(t *Ah_wireless) error {

	table, err := os.ReadFile("/tmp/dcd_stat_dhcp_dns")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	words := strings.Fields(string(lines[0]))

	if((words == nil) ) {
		return nil
	}

	dhip, _		:= strconv.Atoi(words[0])
	dhtime, _	:= strconv.Atoi(words[1])
	dns_count,_	:= strconv.Atoi(words[2])

	ntps		:= strings.Trim(words[3], "[]")
	ntpl, _		:= strconv.Atoi(words[4])

	var j int
	j = 5
	for i := 0; i < dns_count ; i++ {
		dip, _ := strconv.Atoi(words[j])
		dtime, _ := strconv.Atoi(words[j+1])
		j = j + 2

		t.nw_service.dns_ip[i] = uint32(dip)
		t.nw_service.dns_time[i] = int32(dtime)
	}

	t.nw_service.dhcp_ip = uint32(dhip)
	t.nw_service.dhcp_time = int32(dhtime)

	t.nw_service.ntp_server = string(ntps)
	t.nw_service.ntp_latency = int32(ntpl)

	return nil
}


func Gather_Network_Service(t *Ah_wireless) error {
	get_network_dhcp_dns_data(t)
	get_radius_server_data(t)
	get_cwp_server_data(t)
	get_syslog_server_data(t)
	return nil
}

func Gather_Firewall_Stats(t *Ah_wireless) error {
	// Read IP firewall stats from file
	ip_table, err := os.ReadFile("/tmp/telegraf_dcd_ipfirewall_stats")
	if err == nil {
		lines := bytes.Split([]byte(ip_table), newLineByte)
		for _, curLine := range lines {
			if len(curLine) == 0 {
				continue
			}
			words := strings.Fields(string(curLine))
			if len(words) < 3 {
				continue
			}

			time_stamp, _ := strconv.ParseUint(words[0], 10, 32)
			coll_period, _ := strconv.ParseUint(words[1], 10, 32)
			group_cnt, _ := strconv.ParseUint(words[2], 10, 32)

			t.fw_stats.time_stamp = uint32(time_stamp)
			t.fw_stats.coll_period = uint32(coll_period)

			// Parse IP firewall groups
			t.fw_stats.ip_fw_groups = make([]firewall_acl_group, 0, group_cnt)
			idx := 3
			for i := uint64(0); i < group_cnt && idx+1 < len(words); i++ {
				name := words[idx]
				drop_cnt, _ := strconv.ParseUint(words[idx+1], 10, 32)
				t.fw_stats.ip_fw_groups = append(t.fw_stats.ip_fw_groups, firewall_acl_group{
					name:       name,
					drop_count: uint64(drop_cnt),
				})
				idx += 2
			}
			break // Only process first line
		}
	}

	// Read MAC firewall stats from file
	mac_table, err := os.ReadFile("/tmp/telegraf_dcd_macfirewall_stats")
	if err == nil {
		lines := bytes.Split([]byte(mac_table), newLineByte)
		for _, curLine := range lines {
			if len(curLine) == 0 {
				continue
			}
			words := strings.Fields(string(curLine))
			if len(words) < 3 {
				continue
			}

			group_cnt, _ := strconv.ParseUint(words[2], 10, 32)

			// Parse MAC firewall groups
			t.fw_stats.mac_fw_groups = make([]firewall_acl_group, 0, group_cnt)
			idx := 3
			for i := uint64(0); i < group_cnt && idx+1 < len(words); i++ {
				name := words[idx]
				drop_cnt, _ := strconv.ParseUint(words[idx+1], 10, 32)
				t.fw_stats.mac_fw_groups = append(t.fw_stats.mac_fw_groups, firewall_acl_group{
					name:       name,
					drop_count: uint64(drop_cnt),
				})
				idx += 2
			}
			break // Only process first line
		}
	}

	return nil
}

func Send_FirewallStats(t *Ah_wireless, acc telegraf.Accumulator) error {
	var s string
	s = "Firewall Statistics\n\n"
	dumpOutput(FW_STAT_OUT_FILE, s, 0)

	// Emit IP Firewall stats as separate entry
	if len(t.fw_stats.ip_fw_groups) > 0 {
		ipFields := map[string]interface{}{}
		ipFields["collPeriod"] = t.fw_stats.coll_period
		ipFields["timeStamp"] = t.fw_stats.time_stamp
		ipFields["firewallType"] = "IP_FIREWALL"

		// Add IP firewall groups using array notation for aclGroups
		for i, grp := range t.fw_stats.ip_fw_groups {
			name_key := fmt.Sprintf("name_@%d_aclGroups", i)
			drop_key := fmt.Sprintf("dropCount_@%d_aclGroups", i)
			ipFields[name_key] = grp.name
			ipFields[drop_key] = grp.drop_count
		}

		acc.AddGauge("FirewallStats", ipFields, nil)
		prepareAndDumpOutput(FW_STAT_OUT_FILE, ipFields)
		log.Printf("IP firewall stats processed")
	}

	// Emit MAC Firewall stats as separate entry
	if len(t.fw_stats.mac_fw_groups) > 0 {
		macFields := map[string]interface{}{}
		macFields["collPeriod"] = t.fw_stats.coll_period
		macFields["timeStamp"] = t.fw_stats.time_stamp
		macFields["firewallType"] = "MAC_FIREWALL"

		// Add MAC firewall groups using array notation for aclGroups
		for i, grp := range t.fw_stats.mac_fw_groups {
			name_key := fmt.Sprintf("name_@%d_aclGroups", i)
			drop_key := fmt.Sprintf("dropCount_@%d_aclGroups", i)
			macFields[name_key] = grp.name
			macFields[drop_key] = grp.drop_count
		}

		acc.AddGauge("FirewallStats", macFields, nil)
		prepareAndDumpOutput(FW_STAT_OUT_FILE, macFields)
		log.Printf("MAC firewall stats processed")
	}

	return nil
}

func Gather_deffer_end(t *Ah_wireless) {
	t.wg.Done()
	if r := recover(); r != nil {
		currentTime := time.Now()
		crash_file := fmt.Sprintf("/tmp/telegraf_crash_%s.txt", currentTime.Format("2006_01_02_15_04_05"))
		ss := string(debug.Stack())
		log.Printf("telegraf crash: %s\n",ss)
		os.WriteFile(crash_file, debug.Stack(), 0644)
		os.Exit(128)
	}
}

func (t *Ah_wireless) runWirelessStats(
    enableWirelessStat *uint8,
    wirelessStatOutFile string,
    wirelessStatTitle string,
    preWirelessStat func(),
    collectWirelessStat func(),
) {
    if *enableWirelessStat != 1 {
        return
    }
   
    if preWirelessStat != nil {
        preWirelessStat()
    }
	if collectWirelessStat != nil {
        collectWirelessStat()
    }
    *enableWirelessStat = 0

}


func (t *Ah_wireless) Gather(acc telegraf.Accumulator) error {

    t.wg.Add(1)

    go func() {
        defer Gather_deffer_end(t)

        // If ANY test flag set, do shared interface + ssid prep once.
        if t.Test_rf_stats_enable == 1 || t.Test_client_stats_enable == 1 ||
            t.Test_device_stats_enable == 1 || t.Test_network_stats_enable == 1 {
            for _, ifn := range t.Ifname {
                t.intf_m[ifn] = make(map[string]string)
                load_ssid(t, ifn)
            }
        
		// RF one-shot (set rrmid only here)
        if t.Test_rf_stats_enable == 1 {
            rrmid = ahutil.GetRrmId()
        }
        t.runWirelessStats(&t.Test_rf_stats_enable,
            RF_STAT_OUT_FILE,
            "RF Stat Input Plugin Output",
            nil,
            func() { Gather_Rf_Stat(t, acc) },
        )

        // Client one-shot
        t.runWirelessStats(&t.Test_client_stats_enable,
            CLT_STAT_OUT_FILE,
            "Client Stat Input Plugin Output",
            nil,
            func() { Gather_Client_Stat(t, acc) },
        )

        // Device one-shot (needs health + service first)
        t.runWirelessStats(&t.Test_device_stats_enable,
            DEV_STAT_OUT_FILE,
            "Device Stat Input Plugin Output",
            func() {
                Gather_Network_Health(t)
                Gather_Network_Service(t)
            },
            func() { Send_DeviceStats(t, acc) },
        )

        // Network one-shot (needs ethernet interface stats first)
        t.runWirelessStats(&t.Test_network_stats_enable,
            NW_STAT_OUT_FILE,
            "Network Stat Input Plugin Output",
            func() { Gather_EthernetInterfaceStats(t) },
            func() { Send_NetworkStats(t, acc) },
        )

	}

        // Periodic / aggregated path

		if t.timer_count == (t.Scount - 1) {
			dumpOutput(RF_STAT_OUT_FILE, "RF Stat Input Plugin Output", 0)
			dumpOutput(CLT_STAT_OUT_FILE, "Client Stat Input Plugin Output", 0)
			dumpOutput(NW_STAT_OUT_FILE, "Network Stat Input Plugin Output",0)
			dumpOutput(DEV_STAT_OUT_FILE, "Device Stat Input Plugin Output",0)
			dumpOutput(FW_STAT_OUT_FILE, "Firewall Stat Input Plugin Output",0)
			for _, intfName := range t.Ifname {
				t.intf_m[intfName] = make(map[string]string)
				load_ssid(t, intfName)
			}

			Gather_EthernetInterfaceStats(t)

			rrmid = ahutil.GetRrmId()

			Gather_Client_Stat(t, acc)
			Gather_Rf_Stat(t, acc)


			Gather_Network_Health(t)
			Gather_Network_Service(t)
			Gather_Firewall_Stats(t)

			Send_NetworkStats(t,acc)
			Send_DeviceStats(t,acc)
			Send_FirewallStats(t,acc)

			t.timer_count = 0

			t.last_rf_stat =  [4]awestats{}
			t.last_ut_data  = [4]utilization_data{}

		} else {
			Gather_AirTime(t,acc)
			Gather_Rf_Avg(t,acc)
			t.timer_count++
		}
	}()

	t.wg.Wait()

	return nil
}

func (t *Ah_wireless) Start(acc telegraf.Accumulator) error {
	t.intf_m	=	make(map[string]map[string]string)
	t.entity	=	make(map[string]map[string]saved_stats)
	t.last_sq	=	make(map[string]map[int]map[int]ah_signal_quality_stats)

	for _, intfName := range t.Ifname {
		t.entity[intfName] = make(map[string]saved_stats)
	}

	t.if_stats	=	[AH_MAX_WIRED]stats_interface_data{}
	t.ethx_stats =	[AH_MAX_WIRED]stats_ethx_data{}

	t.nw_health =	network_health_data{}
	t.nw_service =  network_service_data{}

	t.last_clt_stat = [4][]ah_ieee80211_sta_stats_item{}

	return nil
}

func init_fe() *os.File {
        file, err := os.Open(AH_FE_DEV_NAME)
        if err != nil {
                log.Printf("Error opening file:", err)
                return nil
        }

	// Get the current flags
        flags, err := unix.FcntlInt(file.Fd(), syscall.F_GETFD, 0)
        if err != nil {
                log.Printf("Error getting flags:", err)
                return nil
        }

        flags |= unix.FD_CLOEXEC

        // Set the close-on-exec flag
        _, err = unix.FcntlInt(file.Fd(), syscall.F_SETFD, flags)
        if err != nil {
                log.Printf("Error setting flags:", err)
                return nil
        }
	return file
}

func init_ethf() *os.File {

	file, err := os.OpenFile("/dev/ah_ethif_ctl", syscall.O_RDONLY | syscall.O_CLOEXEC, 0666);

	if err != nil {
			log.Printf("Error opening file:", err)

			return nil
	}

	return file
}

func (t *Ah_wireless) Stop() {
	unix.Close(t.fd)
}


func init() {
	inputs.Add("ah_wireless_v2", func() telegraf.Input {
		return NewAh_wireless(1)
	})
}
