package ah_airrm

import (
	"log"
	"os"
	"sync"
	"strings"
	"bytes"
	"time"
	"fmt"
	"net"
	"unsafe"
	"runtime/debug"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/common/ahutil"
	"golang.org/x/sys/unix"
)

var (
	aiMutex = new(sync.Mutex)
)

type Ah_airrm struct {
	fd			int
	apn		string
	Ifname		[]string	`toml:"ifname"`
	Log			telegraf.Logger `toml:"-"`
	wg			sync.WaitGroup
	Test_airrm_enable  uint8	`toml:"test_airrm_enable"`
	Enable_periodic    bool     `toml:"enable_periodic"`
}

const sampleConfig = `
[[inputs.ah_airrm]]
  interval = "5s"
  ifname = ["wifi0","wifi1"]
  test_airrm_enable = 0
  enable_periodic = false
`
func NewAh_airrm(id int) *Ah_airrm {
	var err error

	apn := ahutil.GetAPName()
	// Create RAW  Socket.
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return nil
	}
	if id != -1 {
		open(fd, id)
	}

	return &Ah_airrm{
        fd:	fd,
	apn:	apn,
	Test_airrm_enable: 0,
	Enable_periodic: false,
	}

}

func (ai *Ah_airrm) runAirrmOneShot(flag *uint8, collect func()) {
    if flag == nil || *flag != 1 {
        return
    }
    if collect != nil {
        collect()
    }
    *flag = 0
}

func getAirrmNbrTbl(ai *Ah_airrm, ifname string, cfg ieee80211req_cfg_nbr) unsafe.Pointer {

	var size int

	/* first 4 bytes is subcmd */
        switch ai.apn {
                case "AP4020":
                        cfg.cmd = AH_IEEE80211_GET_AIRRM_TBL_AP4020
                case "AP5020":
                        cfg.cmd = AH_IEEE80211_GET_AIRRM_TBL_AP5020
                case "AP5000","AP5010":
                        cfg.cmd = AH_IEEE80211_GET_AIRRM_TBL_AP5000
                case "AP3000":
                        cfg.cmd = AH_IEEE80211_GET_AIRRM_TBL_AP3000
                default:
                        cfg.cmd = AH_IEEE80211_GET_AIRRM_TBL
        }

	iwp := iw_point{pointer: unsafe.Pointer(&cfg)}
	s := ah_ieee80211_airrm_nbr_tbl_t{}
	size = int(unsafe.Sizeof(s))

	request := iwreq{data: iwp}

	if size > AH_USHORT_MAX {
		request.data.length  = AH_USHORT_MAX
		request.data.flags = uint16(size - AH_USHORT_MAX)
	} else {
		request.data.length  = uint16(size)
		request.data.flags = 9
	}

	copy(request.ifrn_name[:], ahutil.Ah_ifname_radio2vap(ifname))

	aiMutex.Lock()

	if err := ahutil.Ah_ioctl(uintptr(ai.fd), ahutil.IEEE80211_IOCTL_GENERIC_PARAM, uintptr(unsafe.Pointer(&request))); err != nil {
		log.Printf("getAirrmNbrTbl ioctl data error %s",err)
		aiMutex.Unlock()
		return request.data.pointer
	}
	aiMutex.Unlock()

	return request.data.pointer
}

func open(fd, id int) *Ah_airrm {
	return &Ah_airrm{fd: fd}
}

func (ai *Ah_airrm) SampleConfig() string {
	return sampleConfig
}

func (ai *Ah_airrm) Description() string {
	return "Hive OS wireless stat"
}

func (ai *Ah_airrm) Init() error {
	return nil
}

func Gather_acs_nbr(ai *Ah_airrm, acc telegraf.Accumulator) error {

	var count int
	var nbr_ssid string
	count = 0
	for _, intfName := range ai.Ifname {

		var i uint32
		var nbr ieee80211req_cfg_nbr
		cfgptr := getAirrmNbrTbl(ai, intfName, nbr)
		if(cfgptr == nil) {
			return nil
		}
		var nbrtbl *ah_ieee80211_airrm_nbr_tbl_t = (*ah_ieee80211_airrm_nbr_tbl_t)(cfgptr)

		iface, err := net.InterfaceByName(intfName)
		if err != nil {
			log.Printf("Error getting index\n")
			iface.Index = 0
		}

		for i = 0; i < nbrtbl.num_nbrs; i++ {
			if i >= MAX_NEIGHBOR_NUM {
				continue
			}

			if nbrtbl.nbr_tbl[i].bssid[0] == 0 && nbrtbl.nbr_tbl[i].bssid[1] == 0 && nbrtbl.nbr_tbl[i].bssid[2] == 0 {
				continue
			}

			nbr_ssid = strings.TrimSpace(string(bytes.Trim(nbrtbl.nbr_tbl[i].ssid[:], "\u0000")))

			if idx := strings.IndexByte(nbr_ssid, '\u0000'); idx >= 0 {
				nbr_ssid = nbr_ssid[:idx]
			}

			fields := map[string]interface{}{
				"name_keys":					intfName,
				"ifIndex_keys":					iface.Index,
			}

			fields["rrmId"] =						nbrtbl.nbr_tbl[i].rrmId
			fields["bssid"] =						ahutil.MacToString(nbrtbl.nbr_tbl[i].bssid)
			fields["ssid"] =						nbr_ssid
			fields["channel"] =					ahutil.FreqToChan(uint16(nbrtbl.nbr_tbl[i].frequency))
			fields["channelWidth"] =				nbrtbl.nbr_tbl[i].channelWidth
			fields["rssi"] =					nbrtbl.nbr_tbl[i].rssi
			fields["txPower"] =                                        nbrtbl.nbr_tbl[i].txPower
			fields["extremeAP"] =					nbrtbl.nbr_tbl[i].extremeAP == 1
			fields["channelUtilization"] =			nbrtbl.nbr_tbl[i].channelUtilization
			fields["interferenceUtilization"] =		nbrtbl.nbr_tbl[i].interferenceUtilization
			fields["rxObssUtilization"] =			nbrtbl.nbr_tbl[i].obssUtilization
			fields["wifinterferenceUtilization"] =	nbrtbl.nbr_tbl[i].wifiInterferenceUtilization
			fields["packetErrorRate"] =				nbrtbl.nbr_tbl[i].packetErrorRate
			fields["aggregationSize"] =				nbrtbl.nbr_tbl[i].aggregationSize
			fields["clientCount"] =					nbrtbl.nbr_tbl[i].clientCount

			acc.AddGauge("RfNbrStats", fields, nil)
			count++
		}
	}
	log.Printf("ah_airrm: rfNbrStats(ai-rrm) is processed with %d entries\n", count)

	return nil
}



func (ai *Ah_airrm) Gather(acc telegraf.Accumulator) error {

	defer func() {
		if r := recover(); r != nil {
			currentTime := time.Now()
			crash_file := fmt.Sprintf("/tmp/telegraf_crash_%s.txt", currentTime.Format("2006_01_02_15_04_05"))
			ss := string(debug.Stack())
			log.Printf("telegraf crash: %s\n",ss)
			os.WriteFile(crash_file, debug.Stack(), 0644)
			os.Exit(128)
		}
	}()
	
    if ai.Test_airrm_enable == 1 {
        ai.runAirrmOneShot(&ai.Test_airrm_enable, func() {
            Gather_acs_nbr(ai, acc)
        })
        return nil
    }

    if ai.Enable_periodic && ai.Test_airrm_enable == 0 {
        Gather_acs_nbr(ai, acc)
    }
    return nil
}

func (ai *Ah_airrm) Start(acc telegraf.Accumulator) error {

	return nil
}


func (ai *Ah_airrm) Stop() {
	unix.Close(ai.fd)
}


func init() {
	inputs.Add("ah_airrm", func() telegraf.Input {
		return NewAh_airrm(1)
	})
}
