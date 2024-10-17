package ah_wireless

import (
	"log"
	"os"
	"bytes"
	"strconv"
	"syscall"
	"strings"
	"sync"
	"time"
	"os/exec"
	"fmt"
	"sort"
	"unsafe"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/unix"
)

var (
	offsetsMutex = new(sync.Mutex)
	newLineByte  = []byte("\n")
)


type Ah_wireless struct {
	fd			int
	fe_fd			uintptr
	intf_m			map[string]map[string]string
	arp_m			map[string]string
	Ifname			[]string	`toml:"ifname"`
	closed			chan		struct{}
	numclient		[4]int
	timer_count		uint8
	entity			map[string]map[string]unsafe.Pointer
	Log			telegraf.Logger `toml:"-"`
	last_rf_stat		[4]awestats
	last_ut_data		[4]utilization_data
	last_clt_stat		[4][50]ah_ieee80211_sta_stats_item
}


func ah_ioctl(fd uintptr, op, argp uintptr) error {
	        _, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(op), argp)
        if errno != 0 {
                return errno
        }
        return nil
}


const sampleConfig = `
[[inputs.ah_wireless]]
  interval = "5s"
  ifname = ["wifi0","wifi1"]
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
        }

}

func get_rt_sta_info(t *Ah_wireless, mac_adrs string) *rt_sta_data {
	app := "telegraf_helper"

	arg0 := mac_adrs

	cmd := exec.Command(app, arg0)
	output, err := cmd.Output()

	if err != nil {
		log.Printf(err.Error())
		return nil
	}


	lines := strings.Split(string(output),"\n")

	data := rt_sta_data{}

	var os_line, host_line, user_line  string

	// Loop over the line to find and extract OS and HostName ans UserName
	for _, line := range lines {
		if strings.HasPrefix(line, "OS:") {
			os_line = strings.TrimSpace(strings.TrimPrefix(line, "OS:"))
		} else if strings.HasPrefix(line, "HostName:") {
			host_line = strings.TrimSpace(strings.TrimPrefix(line, "HostName:"))
		}else if strings.HasPrefix(line, "UserName:") {
			user_line = strings.TrimSpace(strings.TrimPrefix(line, "UserName:"))
		}
	}

	data.os =   string(os_line)
	data.hostname = string(host_line)
	data.user = string(user_line)


	return &data
}

func getHDDStat(fd int, ifname string) *ah_ieee80211_hdd_stats {

        var cfg *ieee80211req_cfg_hdd
        cfg = new(ieee80211req_cfg_hdd)


        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_HDD_STATS;

        iwp := iw_point{pointer: unsafe.Pointer(cfg)}

        request := iwreq{data: iwp}

	request.data.length = VAP_BUFF_SIZE

        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

	offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getHDDStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return nil
        }
		offsetsMutex.Unlock()

        return &cfg.hdd_stats

}

func getAtrTbl(fd int, ifname string) *ah_ieee80211_atr_user {
	var cfg *ieee80211req_cfg_atr
        cfg = new(ieee80211req_cfg_atr)

	/* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ATR_TBL;

	iwp := iw_point{pointer: unsafe.Pointer(cfg)}

	request := iwreq{data: iwp}

	request.data.length = VAP_BUFF_SIZE

	copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

	offsetsMutex.Lock()

	if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getAtrTbl ioctl data error %s",err)
		offsetsMutex.Unlock()
                return nil
        }

	offsetsMutex.Unlock()

	return &cfg.atr

}


func getRFStat(fd int, ifname string) *awestats {

	var p *awestats
	p = new(awestats)

	request := IFReqData{Data: uintptr(unsafe.Pointer(p))}
	copy(request.Name[:], ifname)

	offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), SIOCGRADIOSTATS, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getRFStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return nil
        }

	offsetsMutex.Unlock()

	return p
}

func getStaStat(fd int, ifname string, buf unsafe.Pointer,count int) *ah_ieee80211_get_wifi_sta_stats {

        var cfg *ieee80211req_cfg_sta
        cfg = new(ieee80211req_cfg_sta)

        /* first 4 bytes is subcmd */
        cfg.cmd = IEEE80211_GET_WIFI_STA_STATS
		cfg.wifi_sta_stats.count = uint16(count)
		cfg.wifi_sta_stats.pointer = buf

        iwp := iw_point{pointer: unsafe.Pointer(cfg)}

        request := iwreq{data: iwp}

        request.data.length = VAP_BUFF_SIZE

        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

        offsetsMutex.Lock()

        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getStaStat ioctl data error %s",err)
		offsetsMutex.Unlock()
                return nil
        }

        offsetsMutex.Unlock()

        return &cfg.wifi_sta_stats
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

func getOneStaInfo(fd int, ifname string, mac_ad [MACADDR_LEN]uint8) *ah_ieee80211_sta_info {
        var cfg *ieee80211req_cfg_one_sta
        cfg = new(ieee80211req_cfg_one_sta)

        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ONE_STA_INFO;
	cfg.sta_info.mac = mac_ad
        iwp := iw_point{pointer: unsafe.Pointer(cfg)}
        request := iwreq{data: iwp}
        request.data.length = VAP_BUFF_SIZE
        copy(request.ifrn_name[:], ifname)

        offsetsMutex.Lock()
        if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getOneStaInfo ioctl data error %s",err)
		offsetsMutex.Unlock()
                return nil
        }

        offsetsMutex.Unlock()

        return &cfg.sta_info

}

func getOneSta(fd int, ifname string, mac_ad [MACADDR_LEN]uint8) unsafe.Pointer {
        var cfg *ieee80211req_cfg_one_sta_info
        cfg = new(ieee80211req_cfg_one_sta_info)

        /* first 4 bytes is subcmd */
        cfg.cmd = AH_IEEE80211_GET_ONE_STA
        cfg.mac = mac_ad
        iwp := iw_point{pointer: unsafe.Pointer(cfg)}
        request := iwreq{data: iwp}
        request.data.length = VAP_BUFF_SIZE
        copy(request.ifrn_name[:], ah_ifname_radio2vap(ifname))

        offsetsMutex.Lock()
	if err := ah_ioctl(uintptr(fd), IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
                log.Printf("getOneSta ioctl data error %s",err)
		offsetsMutex.Unlock()
		return nil
        }
	offsetsMutex.Unlock()

        return request.data.pointer

}

func getProcNetDev(ifname string) *ah_dcd_dev_stats {
	table, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil;
	}

	lines := bytes.Split([]byte(table), newLineByte)

	var intfname string
	var stats = new(ah_dcd_dev_stats)
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

func load_ssid(t *Ah_wireless, ifname string) {

	for i := 1; i < 1024; i++ {
		vifname := ifname + "." + strconv.Itoa(i)

		app := "wl"

		arg0 := "-i"
		arg1 := vifname
		arg2 := "status"
		//  arg3 := "| grep \"SSID: \"\\\""
		//log.Printf(app + " " + arg0 + " " + arg1 + " " + arg2)

		cmd := exec.Command(app, arg0, arg1, arg2)
		output, err := cmd.Output()

		if err != nil {
			log.Printf(err.Error())
			return
		}

		lines := strings.Split(string(output),"\n")

		temp  := strings.Split(lines[0]," ")

		ssid := strings.Trim(temp[1], "\"")
		t.intf_m[ifname][ssid] = vifname
	}
}

func load_arp_table(t *Ah_wireless) {

	app := "arp"
	arg := "-v"

	cmd := exec.Command(app, arg)

	arp_str, err := cmd.Output()

	if err != nil {
		log.Printf(err.Error())
		return
	}

	arp_lines := strings.Split(string(arp_str),"\n")

	for i :=0; i<len(arp_lines); i++ {
		if len(arp_lines[i]) > 1 {
			arp_eliments := strings.Split(arp_lines[i]," ")
			t.arp_m[arp_eliments[3]] = arp_eliments[1]
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

func getFeServerIp(fd uintptr, clmac [MACADDR_LEN]uint8) *ah_flow_get_sta_server_ip_msg {

        msg := ah_flow_get_sta_server_ip_msg{
                                        mac: clmac,
                                }
        ihdr := ah_fe_ioctl_hdr{
                                        retval: -1,
                                        msg_type: AH_FLOW_GET_STATION_SERVER_IP,
                                        msg_size: uint16(unsafe.Sizeof(msg)),
                                }
        dev_msg := ah_fw_dev_ip_msg{
                                        hdr: ihdr,
                                        data: msg,
                                }

        offsetsMutex.Lock()

        if err := ah_ioctl(fd, AH_FE_IOCTL_FLOW, uintptr(unsafe.Pointer(&dev_msg))); err != nil {
                log.Printf("getFeServerIp ioctl data error %s",err)
                offsetsMutex.Unlock()
                return nil
        }

        offsetsMutex.Unlock()


        if dev_msg.hdr.retval < 0 {
                log.Printf("Open ioctl data erro")
                return nil
        }

        return &dev_msg.data
}

func open(fd, id int) *Ah_wireless {

	//getProcNetDev("wifi1")
	//defer unix.Close(fd)

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

func Gather_Rf_Avg(t *Ah_wireless, acc telegraf.Accumulator) error {
	var ii int
	ii = 0
	for _, intfName := range t.Ifname {

		var rfstat *awestats
		var devstats *ah_dcd_dev_stats
		var ifindex int
		var atrStat *ah_ieee80211_atr_user
		var hddStat *ah_ieee80211_hdd_stats

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

		rfstat  = getRFStat(t.fd, intfName)
		if (rfstat == nil) {
			continue
		}

		ifindex = getIfIndex(t.fd, intfName)
		if (ifindex <= 0) {
			continue
		}
		devstats = getProcNetDev(intfName)
		if (devstats == nil) {
			continue
		}
		atrStat = getAtrTbl(t.fd, intfName)
		if (atrStat == nil) {
			continue
		}
		hddStat = getHDDStat(t.fd, intfName)
		if (hddStat == nil) {
			continue
		}


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

				/* Calculate Utilization */

			}

			t.last_rf_stat[ii] = *rfstat
			ii++
		}

		return nil
}



func Gather_Rf_Stat(t *Ah_wireless, acc telegraf.Accumulator) error {
	var ii int
	ii = 0
	for _, intfName := range t.Ifname {

		var rfstat *awestats
		var devstats *ah_dcd_dev_stats
		var ifindex int
		var atrStat *ah_ieee80211_atr_user
		var hddStat *ah_ieee80211_hdd_stats

		var idx			int
		var tmp_count1	int64
		var tmp_count2	int64
		//var tx_ok		uint64
		//var rx_ok		uint64
		var tx_total	int64
		var rx_total	int64
		var tmp_count3	int32
		var tmp_count4	int32
		var tot_tx_bitrate_retries uint32
		var tot_rx_bitrate_retries uint32

		var rf_report	ah_dcd_stats_report_int_data

		rfstat  = getRFStat(t.fd, intfName)
		if (rfstat == nil) {
			continue
		}

		ifindex = getIfIndex(t.fd, intfName)
		if (ifindex <= 0) {
			continue
		}
		devstats = getProcNetDev(intfName)
		if (devstats == nil) {
			continue
		}
		atrStat = getAtrTbl(t.fd, intfName)
		if (atrStat == nil) {
			continue
		}
		hddStat = getHDDStat(t.fd, intfName)
		if (hddStat == nil) {
			continue
		}


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
	//tx_ok = uint64(tx_total)
	//rx_ok = uint64(rx_total)



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
			"ifindex_keys":					ifindex,

		}

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

			fields["wifinterferenceUtilization_min"]			= t.last_ut_data[ii].wifi_i_util_min
			fields["wifinterferenceUtilization_max"]			= t.last_ut_data[ii].wifi_i_util_max
			fields["wifinterferenceUtilization_avg"]			= t.last_ut_data[ii].wifi_i_util_avg

			fields["noise_min"]						= t.last_ut_data[ii].noise_min
			fields["noise_max"]						= t.last_ut_data[ii].noise_max
			fields["noise_avg"]						= t.last_ut_data[ii].noise_avg

			fields["crcErrorRate_min"]					= t.last_ut_data[ii].crc_err_rate_min
			fields["crcErrorRate_max"]					= t.last_ut_data[ii].crc_err_rate_max
			fields["crcErrorRate_avg"]					= t.last_ut_data[ii].crc_err_rate_avg


			fields["txPackets"]						= devstats.tx_packets
			fields["txErrors"]						= devstats.tx_errors
			fields["txDropped"]						= devstats.tx_dropped
			fields["txHwDropped"]						= rfstat.ast_as.ast_tx_shortpre + rfstat.ast_as.ast_tx_xretries + rfstat.ast_as.ast_tx_fifoerr
			fields["txSwDropped"]						= devstats.tx_dropped
			fields["txBytes"]						= devstats.tx_bytes
			fields["txRetryCount"]						= rfstat.phy_stats.ast_tx_shortretry + rfstat.phy_stats.ast_tx_longretry

			fields["txRate_min"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps
			fields["txRate_max"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps
			fields["txRate_avg"]						= rfstat.ast_tx_rate_stats[0].ns_rateKbps

			fields["txUnicastPackets"]					= rfstat.ast_tx_rate_stats[0].ns_unicasts
			fields["txMulticastPackets"]					= devstats.tx_multicast
			fields["txMulticastBytes"]					= rfstat.ast_as.ast_tx_mcast_bytes
			fields["txBcastBytes"]						= rfstat.ast_as.ast_tx_bcast_bytes
			fields["txBcastPackets"]					= devstats.tx_broadcast

			fields["rxPackets"]						= devstats.rx_packets
			fields["rxErrors"]						= devstats.rx_errors
			fields["rxDropped"]						= devstats.rx_dropped
			fields["rxBytes"]						= devstats.rx_bytes
			fields["rxRetryCount"]						= rfstat.ast_rx_retry

			fields["rxRate_min"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps
			fields["rxRate_max"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps
			fields["rxRate_avg"]						= rfstat.ast_rx_rate_stats[0].ns_rateKbps

			fields["rxMulticastBytes"]					= rfstat.ast_rx_mcast_bytes
			fields["rxMulticastPackets"]					= devstats.rx_multicast
			fields["rxBcastPackets"]					= devstats.rx_broadcast
			fields["rxBcastBytes"]						= rfstat.ast_rx_bcast_bytes

			fields["bsSpCnt"]						= hddStat.bs_sp_cnt
			fields["snrSpCnt"]						= hddStat.snr_sp_cnt
			fields["snAnswerCnt"]						= reportGetDiff(hddStat.sn_answer_cnt, hddStat.sn_answer_cnt)
			fields["rxPrbSpCnt"]						= rfstat.is_rx_hdd_probe_sup
			fields["rxAuthCnt"]						= rfstat.is_rx_hdd_auth_sup

			fields["txBitrateSuc"]						= rfstat.ast_tx_rix_invalids
			fields["rxBitrateSuc"]						= rfstat.ast_rx_rix_invalids

				for i := 0; i < NS_HW_RATE_SIZE; i++{
					kbps := fmt.Sprintf("kbps_%d_rxRateStats",i)
					rateDtn := fmt.Sprintf("rateDtn_%d_rxRateStats",i)
					rateSucDtn := fmt.Sprintf("rateSucDtn_%d_rxRateStats",i)
					fields[kbps]					= rf_report.rx_bit_rate[i].kbps
					fields[rateDtn]					= rf_report.rx_bit_rate[i].rate_dtn
					fields[rateSucDtn]				= rf_report.rx_bit_rate[i].rate_suc_dtn
				}


				for i := 0; i < NS_HW_RATE_SIZE; i++{
					kbps := fmt.Sprintf("kbps_%d_txRateStats",i)
					rateDtn := fmt.Sprintf("rateDtn_%d_txRateStats",i)
					rateSucDtn := fmt.Sprintf("rateSucDtn_%d_txRateStats",i)
					fields[kbps]					= rf_report.tx_bit_rate[i].kbps
					fields[rateDtn]					= rf_report.tx_bit_rate[i].rate_dtn
					fields[rateSucDtn]				= rf_report.tx_bit_rate[i].rate_suc_dtn
				}

			fields["clientCount"]						= t.numclient[ii]
			fields["lbSpCnt"]							= hddStat.lb_sp_cnt
			fields["rxProbeSup"]						= rfstat.is_rx_hdd_probe_sup
			fields["rxSwDropped"]						= devstats.rx_dropped
			fields["rxUnicastPackets"]					= rfstat.ast_rx_rate_stats[0].ns_unicasts


			acc.AddGauge("RfStats", fields, nil)

			var s string

			s = "Stats of interface " + intfName + "\n\n"

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

			log.Printf("ah_wireless: radio status is processed")

			dumpOutput(RF_STAT_OUT_FILE, s, 1)

			t.last_rf_stat[ii] = *rfstat
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


    for _, intfName2 := range t.Ifname {

		var cltstat *ah_ieee80211_get_wifi_sta_stats
		var ifindex2 int
		var numassoc int
		var stainfo *ah_ieee80211_sta_info

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


		clt_item := make([]ah_ieee80211_sta_stats_item, numassoc)


		ifindex2 = getIfIndex(t.fd, intfName2)
		if(ifindex2 <= 0 ) {
			continue
		}

		cltstat = getStaStat(t.fd, intfName2, unsafe.Pointer(&clt_item[0]),  numassoc)

		for cn := 0; cn < numassoc; cn++ {
			//if ( clt_item[cn] == nil) {
			//	continue
			//}
			client_ssid = string(bytes.Trim(clt_item[cn].ns_ssid[:], "\x00"))

			if(clt_item[cn].ns_mac[0] !=0 || clt_item[cn].ns_mac[1] !=0 || clt_item[cn].ns_mac[2] !=0 || clt_item[cn].ns_mac[3] !=0 || clt_item[cn].ns_mac[4] != 0 || clt_item[cn].ns_mac[5]!=0) {
				cintfName := t.intf_m[intfName2][client_ssid]
				client_mac = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])


				stainfo = getOneStaInfo(t.fd, cintfName, clt_item[cn].ns_mac)

				if(stainfo==nil) {
					log.Printf("Error in getOneStaInfo")
					continue
				}

				if stainfo.rssi == 0 {
					continue
				}
			} else {
				stainfo = nil
				continue
			}

			f := init_fe()
			ipnet_score := getFeIpnetScore(f.Fd(), clt_item[cn].ns_mac)
			sta_ip := getFeServerIp(f.Fd(), clt_item[cn].ns_mac)
			f.Close()

			//client_mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])
			cfgptr := getOneSta(t.fd, intfName2, clt_item[cn].ns_mac)

			if(cfgptr == nil) {
				continue
			}

			var onesta *ieee80211req_sta_info = (*ieee80211req_sta_info)(cfgptr)

			var clt_last_stats *saved_stats = (*saved_stats)(t.entity[intfName2][client_mac])

/*
			if (clt_last_stats != nil) {
				clt_last_stats.tx_airtime_min	= ((clt_last_stats.tx_airtime_min /10) / (60 *1000))
				clt_last_stats.tx_airtime_max	= ((clt_last_stats.tx_airtime_max /10) / (60 *1000))
				clt_last_stats.tx_airtime_average	= ((clt_last_stats.tx_airtime_average /10) / (60 *1000))

				clt_last_stats.rx_airtime_min	= ((clt_last_stats.rx_airtime_min /10) / (60 *1000))
				clt_last_stats.rx_airtime_max	= ((clt_last_stats.rx_airtime_max /10) / (60 *1000))
				clt_last_stats.rx_airtime_average	= ((clt_last_stats.rx_airtime_average /10) / (60 *1000))

				if (clt_last_stats.tx_airtime_min > 100) {
					clt_last_stats.tx_airtime_min = 100
				}

				if (clt_last_stats.tx_airtime_max > 100) {
					clt_last_stats.tx_airtime_max = 100
				}

				if (clt_last_stats.tx_airtime_average > 100) {
					clt_last_stats.tx_airtime_average = 100
				}

				if (clt_last_stats.rx_airtime_min > 100) {
					clt_last_stats.rx_airtime_min = 100
				}

				if (clt_last_stats.rx_airtime_max > 100) {
					clt_last_stats.rx_airtime_max = 100
				}

				if (clt_last_stats.rx_airtime_average > 100) {
					clt_last_stats.rx_airtime_average = 100
				}
			}
*/

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
					conn_score = (int64(rate_score) * int64(success) * tot_pcnt)
				}
			}
			var rssi int
			var radio_link_score int64
			if (stainfo != nil) {
				rssi = int(stainfo.rssi) + int(stainfo.noise_floor)
			} else {
				rssi = 0
			}
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
			t.last_clt_stat[ii][cn] = clt_item[cn]

			var rt_sta *rt_sta_data
			rt_sta = get_rt_sta_info(t, client_mac)

			fields2["ifname"]               = intfName2
			fields2["ifIndex"]              = ifindex2

			fields2["mac_keys"]		= client_mac

			fields2["number"]		= cltstat.count
			fields2["ssid"]			= client_ssid
                        fields2["txPackets"]		= stainfo.tx_pkts
                        fields2["txBytes"]		= stainfo.tx_bytes
                        fields2["txDrop"]		= clt_item[cn].ns_tx_drops
                        fields2["slaDrop"]		= clt_item[cn].ns_sla_traps
                        fields2["rxPackets"]		= stainfo.rx_pkts
                        fields2["rxBytes"]		= stainfo.rx_bytes
                        fields2["rxDrop"]		= clt_item[cn].ns_tx_drops
                        fields2["avgSnr"]		= clt_item[cn].ns_snr
                        fields2["psTimes"]		= clt_item[cn].ns_ps_times
                        fields2["radioScore"]		= radio_link_score
                        fields2["ipNetScore"]		= ipnet_score
			if ipnet_score == 0 {
				fields2["appScore"]	= ipnet_score
			} else {
				fields2["appScore"]	= clt_item[cn].ns_app_health_score
			}
                        fields2["phyMode"]		= getMacProtoMode(onesta.isi_phymode)
			if(stainfo != nil) {
				fields2["rssi"]		= int(stainfo.rssi) + int(stainfo.noise_floor)
			} else {
				fields2["rssi"]		= 0
			}
                        fields2["os"]			= rt_sta.os
			fields2["name"]			= string(onesta.isi_name[:])
                        fields2["host"]			= rt_sta.hostname
                        fields2["profName"]		= "default-profile"			/* TBD (Needs shared memory of dcd)	*/
                        fields2["dhcpIp"]		= intToIp(sta_ip.dhcp_server)
			fields2["gwIp"]			= intToIp(sta_ip.gateway)
                        fields2["dnsIp"]		= intToIp(sta_ip.dns[0].dns_ip)
			fields2["clientIp"]		= intToIp(sta_ip.client_static_ip)
                        fields2["dhcpTime"]		= sta_ip.dhcp_time
                        fields2["gwTime"]		= 0					/* TBD (Needs shared memory of auth2)     */
                        fields2["dnsTime"]		= sta_ip.dns[0].dns_response_time
                        fields2["clientTime"]		= onesta.isi_assoc_time


			for i := 0; i < AH_TX_NSS_MAX; i++{
				txNssUsage := fmt.Sprintf("txNssUsage_%d",i)
				fields2[txNssUsage]           = clt_item[cn].ns_tx_nss[i]
			}



			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_%d_rxRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_%d_rxRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_%d_rxRateStats",i)
				fields2[kbps]			= rf_report.rx_bit_rate[i].kbps
				fields2[rateDtn]		= rf_report.rx_bit_rate[i].rate_dtn
				fields2[rateSucDtn]		= rf_report.rx_bit_rate[i].rate_suc_dtn
			}


			for i := 0; i < NS_HW_RATE_SIZE; i++{
				kbps := fmt.Sprintf("kbps_%d_txRateStats",i)
				rateDtn := fmt.Sprintf("rateDtn_%d_txRateStats",i)
				rateSucDtn := fmt.Sprintf("rateSucDtn_%d_txRateStats",i)
				fields2[kbps]			= rf_report.tx_bit_rate[i].kbps
				fields2[rateDtn]		= rf_report.tx_bit_rate[i].rate_dtn
				fields2[rateSucDtn]		= rf_report.tx_bit_rate[i].rate_suc_dtn
			}

			if (clt_last_stats != nil) {
				fields2["txAirtime_min"]				= clt_last_stats.tx_airtime_min 
				fields2["txAirtime_max"]				= clt_last_stats.tx_airtime_max 
				fields2["txAirtime_avg"]				= clt_last_stats.tx_airtime_average

				fields2["rxAirtime_min"]				= clt_last_stats.rx_airtime_min
				fields2["rxAirtime_max"]				= clt_last_stats.rx_airtime_max
				fields2["rxAirtime_avg"]				= clt_last_stats.rx_airtime_average

				fields2["bwUsage_min"]					= clt_last_stats.bw_usage_min
				fields2["bwUsage_max"]					= clt_last_stats.bw_usage_max
				fields2["bwUsage_avg"]					= clt_last_stats.bw_usage_average
			}
			acc.AddFields("ClientStats", fields2, tags, time.Now())


			clt_new_stats := saved_stats{}
			clt_new_stats.tx_airtime = clt_item[cn].ns_tx_airtime
			clt_new_stats.rx_airtime = clt_item[cn].ns_rx_airtime
			t.entity[intfName2][client_mac] = unsafe.Pointer(&clt_new_stats)


			var s string

			s = "Stats of client [" + client_mac + "]\n\n"

			for k, v := range fields2 {
				if  fmt.Sprint(v) == "0" { // Check if the value is zero
					delete(fields2, k)
				}
			}

			keys := make([]string, 0, len(fields2))

			for k := range fields2{
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for _, k := range keys {
				s = s + k + " : " + fmt.Sprint(fields2[k]) + "\n"
			}

			s = s + "---------------------------------------------------------------------------------------------\n"


			dumpOutput(CLT_STAT_OUT_FILE, s, 1)

		}
		ii++

	}

	log.Printf("ah_wireless: client status is processed")

	return nil
}

func Gather_AirTime(t *Ah_wireless, acc telegraf.Accumulator) error {

	for _, intfName2 := range t.Ifname {

		var numassoc1 int
		var client_mac1 string

		numassoc1 = int(getNumAssocs(t.fd, intfName2))

		if(numassoc1 == 0) {
			continue
		}


		clt_item := make([]ah_ieee80211_sta_stats_item, numassoc1)

		getStaStat(t.fd, intfName2, unsafe.Pointer(&clt_item[0]),  numassoc1)

		for cn := 0; cn < numassoc1; cn++ {

			if(clt_item[cn].ns_mac[0] !=0 || clt_item[cn].ns_mac[1] !=0 || clt_item[cn].ns_mac[2] !=0 || clt_item[cn].ns_mac[3] !=0 || clt_item[cn].ns_mac[4] != 0 || clt_item[cn].ns_mac[5]!=0) {

				client_mac1 = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",clt_item[cn].ns_mac[0],clt_item[cn].ns_mac[1],clt_item[cn].ns_mac[2],clt_item[cn].ns_mac[3],clt_item[cn].ns_mac[4],clt_item[cn].ns_mac[5])
			} else {

				continue
			}

			var clt_last_stats *saved_stats = (*saved_stats)(t.entity[intfName2][client_mac1])

			if(clt_last_stats == nil) {
				clt_new_stats := saved_stats{
								tx_airtime_min:0,
								tx_airtime_max:0,
								tx_airtime_average:0,
								rx_airtime_min:0,
								rx_airtime_max:0,
								rx_airtime_average:0,
								bw_usage_min:0,
								bw_usage_max:0,
								bw_usage_average:0,
								tx_airtime:0,
								rx_airtime:0}
				t.entity[intfName2][client_mac1] = unsafe.Pointer(&clt_new_stats)
				return nil
			}

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
			t.entity[intfName2][client_mac1] = unsafe.Pointer(&clt_new_stats)
		}

	}

	return nil
}

func (t *Ah_wireless) Gather(acc telegraf.Accumulator) error {
	if t.timer_count == 9 {
		dumpOutput(RF_STAT_OUT_FILE, "RF Stat Input Plugin Output", 0)
		dumpOutput(CLT_STAT_OUT_FILE, "Client Stat Input Plugin Output", 0)
		for _, intfName := range t.Ifname {
			t.intf_m[intfName] = make(map[string]string)
			load_ssid(t, intfName)
		}
		Gather_Client_Stat(t, acc)
		Gather_Rf_Stat(t, acc)
		t.timer_count = 0

		t.last_rf_stat =  [4]awestats{}
		t.last_ut_data  = [4]utilization_data{}
		t.last_clt_stat = [4][50]ah_ieee80211_sta_stats_item{}

	} else {
		Gather_AirTime(t,acc)
		Gather_Rf_Avg(t,acc)
		t.timer_count++
	}

	return nil
}



func (t *Ah_wireless) Start(acc telegraf.Accumulator) error {
	t.intf_m = make(map[string]map[string]string)
	t.entity = make(map[string]map[string]unsafe.Pointer)

	for _, intfName := range t.Ifname {
		t.entity[intfName] = make(map[string]unsafe.Pointer)
//		load_ssid(t, intfName)
	}
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


func (t *Ah_wireless) Stop() {
	unix.Close(t.fd)
}


func init() {
	inputs.Add("ah_wireless", func() telegraf.Input {
		return NewAh_wireless(1)
	})
}
