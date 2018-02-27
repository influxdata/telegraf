package teamspeak

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

const welcome = `Welcome to the TeamSpeak 3 ServerQuery interface, type "help" for a list of commands and "help <command>" for information on a specific command.`
const ok = `error id=0 msg=ok`
const errorMsg = `error id=256 msg=command\snot\sfound`

var cmd = map[string]string{
	"login":                       "",
	"use":                         "",
	"serverinfo":                  `virtualserver_unique_identifier=a1vn9PLF8CMIU virtualserver_name=Testserver virtualserver_welcomemessage=Test virtualserver_platform=Linux virtualserver_version=3.0.13.8\s[Build:\s1500452811] virtualserver_maxclients=32 virtualserver_password virtualserver_clientsonline=2 virtualserver_channelsonline=1 virtualserver_created=1507400243 virtualserver_uptime=148 virtualserver_codec_encryption_mode=0 virtualserver_hostmessage virtualserver_hostmessage_mode=0 virtualserver_filebase=files\/virtualserver_1 virtualserver_default_server_group=8 virtualserver_default_channel_group=8 virtualserver_flag_password=0 virtualserver_default_channel_admin_group=5 virtualserver_max_download_total_bandwidth=18446744073709551615 virtualserver_max_upload_total_bandwidth=18446744073709551615 virtualserver_hostbanner_url virtualserver_hostbanner_gfx_url virtualserver_hostbanner_gfx_interval=0 virtualserver_complain_autoban_count=5 virtualserver_complain_autoban_time=1200 virtualserver_complain_remove_time=3600 virtualserver_min_clients_in_channel_before_forced_silence=100 virtualserver_priority_speaker_dimm_modificator=-18.0000 virtualserver_id=1 virtualserver_antiflood_points_tick_reduce=5 virtualserver_antiflood_points_needed_command_block=150 virtualserver_antiflood_points_needed_ip_block=250 virtualserver_client_connections=1 virtualserver_query_client_connections=1 virtualserver_hostbutton_tooltip virtualserver_hostbutton_url virtualserver_hostbutton_gfx_url virtualserver_queryclientsonline=1 virtualserver_download_quota=18446744073709551615 virtualserver_upload_quota=18446744073709551615 virtualserver_month_bytes_downloaded=0 virtualserver_month_bytes_uploaded=0 virtualserver_total_bytes_downloaded=0 virtualserver_total_bytes_uploaded=0 virtualserver_port=9987 virtualserver_autostart=1 virtualserver_machine_id virtualserver_needed_identity_security_level=8 virtualserver_log_client=0 virtualserver_log_query=0 virtualserver_log_channel=0 virtualserver_log_permissions=1 virtualserver_log_server=0 virtualserver_log_filetransfer=0 virtualserver_min_client_version=1445512488 virtualserver_name_phonetic virtualserver_icon_id=0 virtualserver_reserved_slots=0 virtualserver_total_packetloss_speech=0.0000 virtualserver_total_packetloss_keepalive=0.0000 virtualserver_total_packetloss_control=0.0000 virtualserver_total_packetloss_total=0.0000 virtualserver_total_ping=1.0000 virtualserver_ip=0.0.0.0,\s:: virtualserver_weblist_enabled=1 virtualserver_ask_for_privilegekey=0 virtualserver_hostbanner_mode=0 virtualserver_channel_temp_delete_delay_default=0 virtualserver_min_android_version=1407159763 virtualserver_min_ios_version=1407159763 virtualserver_status=online connection_filetransfer_bandwidth_sent=0 connection_filetransfer_bandwidth_received=0 connection_filetransfer_bytes_sent_total=0 connection_filetransfer_bytes_received_total=0 connection_packets_sent_speech=0 connection_bytes_sent_speech=0 connection_packets_received_speech=0 connection_bytes_received_speech=0 connection_packets_sent_keepalive=261 connection_bytes_sent_keepalive=10701 connection_packets_received_keepalive=261 connection_bytes_received_keepalive=10961 connection_packets_sent_control=54 connection_bytes_sent_control=15143 connection_packets_received_control=55 connection_bytes_received_control=4239 connection_packets_sent_total=315 connection_bytes_sent_total=25844 connection_packets_received_total=316 connection_bytes_received_total=15200 connection_bandwidth_sent_last_second_total=81 connection_bandwidth_sent_last_minute_total=141 connection_bandwidth_received_last_second_total=83 connection_bandwidth_received_last_minute_total=98`,
	"serverrequestconnectioninfo": `connection_filetransfer_bandwidth_sent=0 connection_filetransfer_bandwidth_received=0 connection_filetransfer_bytes_sent_total=0 connection_filetransfer_bytes_received_total=0 connection_packets_sent_total=369 connection_bytes_sent_total=28058 connection_packets_received_total=370 connection_bytes_received_total=17468 connection_bandwidth_sent_last_second_total=81 connection_bandwidth_sent_last_minute_total=109 connection_bandwidth_received_last_second_total=83 connection_bandwidth_received_last_minute_total=94 connection_connected_time=174 connection_packetloss_total=0.0000 connection_ping=1.0000`,
}

func TestGather(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Initializing test server failed")
	}
	defer l.Close()

	go handleRequest(l, t)

	var acc testutil.Accumulator
	testConfig := Teamspeak{
		Server:         l.Addr().String(),
		Username:       "serveradmin",
		Password:       "test",
		VirtualServers: []int{1},
	}
	err = testConfig.Gather(&acc)

	if err != nil {
		t.Fatalf("Gather returned error. Error: %s\n", err)
	}

	fields := map[string]interface{}{
		"uptime":                 int(148),
		"clients_online":         int(2),
		"total_ping":             float32(1.0000),
		"total_packet_loss":      float64(0.0000),
		"packets_sent_total":     uint64(369),
		"packets_received_total": uint64(370),
		"bytes_sent_total":       uint64(28058),
		"bytes_received_total":   uint64(17468),
	}

	acc.AssertContainsFields(t, "teamspeak", fields)
}

func handleRequest(l net.Listener, t *testing.T) {
	c, err := l.Accept()
	if err != nil {
		t.Fatal("Error accepting test connection")
	}
	c.Write([]byte("TS3\n\r" + welcome + "\n\r"))
	for {
		msg, _, err := bufio.NewReader(c).ReadLine()
		if err != nil {
			return
		}
		r, exists := cmd[strings.Split(string(msg), " ")[0]]

		if exists {
			switch r {
			case "":
				c.Write([]byte(ok + "\n\r"))
			case "quit":
				c.Write([]byte(ok + "\n\r"))
				c.Close()
				return
			default:
				c.Write([]byte(r + "\n\r" + ok + "\n\r"))
			}
		} else {
			c.Write([]byte(errorMsg + "\n\r"))
		}
	}
}
