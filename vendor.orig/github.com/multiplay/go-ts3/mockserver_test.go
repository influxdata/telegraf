package ts3

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cmdQuit = "quit"
	banner  = `Welcome to the TeamSpeak 3 ServerQuery interface, type "help" for a list of commands and "help <command>" for information on a specific command.`

	errUnknownCmd = `error id=256 msg=command\snot\sfound`
	errOK         = `error id=0 msg=ok`
)

var (
	commands = map[string]string{
		"version":                     "version=3.0.12.2 build=1455547898 platform=FreeBSD",
		"login":                       "",
		"logout":                      "",
		"use":                         "",
		"serverlist":                  `virtualserver_id=1 virtualserver_port=10677 virtualserver_status=online virtualserver_clientsonline=1 virtualserver_queryclientsonline=1 virtualserver_maxclients=35 virtualserver_uptime=12345025 virtualserver_name=Server\s#1 virtualserver_autostart=1 virtualserver_machine_id=1 virtualserver_unique_identifier=uniq1|virtualserver_id=2 virtualserver_port=10617 virtualserver_status=online virtualserver_clientsonline=3 virtualserver_queryclientsonline=2 virtualserver_maxclients=10 virtualserver_uptime=3165117 virtualserver_name=Server\s#2 virtualserver_autostart=1 virtualserver_machine_id=1 virtualserver_unique_identifier=uniq2`,
		"serverinfo":                  `virtualserver_antiflood_points_needed_command_block=150 virtualserver_antiflood_points_needed_ip_block=250 virtualserver_antiflood_points_tick_reduce=5 virtualserver_channel_temp_delete_delay_default=0 virtualserver_codec_encryption_mode=0 virtualserver_complain_autoban_count=5 virtualserver_complain_autoban_time=1200 virtualserver_complain_remove_time=3600 virtualserver_created=0 virtualserver_default_channel_admin_group=1 virtualserver_default_channel_group=4 virtualserver_default_server_group=5 virtualserver_download_quota=18446744073709551615 virtualserver_filebase=files virtualserver_flag_password=0 virtualserver_hostbanner_gfx_interval=0 virtualserver_hostbanner_gfx_url virtualserver_hostbanner_mode=0 virtualserver_hostbanner_url virtualserver_hostbutton_gfx_url virtualserver_hostbutton_tooltip=Multiplay\sGame\sServers virtualserver_hostbutton_url=http:\/\/www.multiplaygameservers.com virtualserver_hostmessage virtualserver_hostmessage_mode=0 virtualserver_icon_id=0 virtualserver_log_channel=0 virtualserver_log_client=0 virtualserver_log_filetransfer=0 virtualserver_log_permissions=1 virtualserver_log_query=0 virtualserver_log_server=0 virtualserver_max_download_total_bandwidth=18446744073709551615 virtualserver_max_upload_total_bandwidth=18446744073709551615 virtualserver_maxclients=32 virtualserver_min_android_version=0 virtualserver_min_client_version=0 virtualserver_min_clients_in_channel_before_forced_silence=100 virtualserver_min_ios_version=0 virtualserver_name=Test\sServer virtualserver_name_phonetic virtualserver_needed_identity_security_level=8 virtualserver_password virtualserver_priority_speaker_dimm_modificator=-18.0000 virtualserver_reserved_slots=0 virtualserver_status=template virtualserver_unique_identifier virtualserver_upload_quota=18446744073709551615 virtualserver_weblist_enabled=1 virtualserver_welcomemessage=Welcome\sto\sTeamSpeak,\scheck\s[URL]www.teamspeak.com[\/URL]\sfor\slatest\sinfos.`,
		"servercreate":                `sid=2 virtualserver_port=9988 token=eKnFZQ9EK7G7MhtuQB6+N2B1PNZZ6OZL3ycDp2OW`,
		"serveridgetbyport":           `server_id=1`,
		"servergrouplist":             `sgid=1 name=Guest\sServer\sQuery type=2 iconid=0 savedb=0 sortid=0 namemode=0 n_modifyp=0 n_member_addp=0 n_member_removep=0|sgid=2 name=Admin\sServer\sQuery type=2 iconid=500 savedb=1 sortid=0 namemode=0 n_modifyp=100 n_member_addp=100 n_member_removep=100`,
		"privilegekeylist":            `token=zTfamFVhiMEzhTl49KrOVYaMilHPDQEBQOJFh6qX token_type=0 token_id1=17395 token_id2=0 token_created=1499948005 token_description`,
		"privilegekeyadd":             `token=zTfamFVhiMEzhTl49KrOVYaMilHPgQEBQOJFh6qX`,
		"serverdelete":                "",
		"serverstop":                  "",
		"serverstart":                 "",
		"serveredit":                  "",
		"instanceinfo":                "serverinstance_database_version=26 serverinstance_filetransfer_port=30033 serverinstance_max_download_total_bandwidth=18446744073709551615 serverinstance_max_upload_total_bandwidth=18446744073709551615 serverinstance_guest_serverquery_group=1 serverinstance_serverquery_flood_commands=50 serverinstance_serverquery_flood_time=3 serverinstance_serverquery_ban_time=600 serverinstance_template_serveradmin_group=3 serverinstance_template_serverdefault_group=5 serverinstance_template_channeladmin_group=1 serverinstance_template_channeldefault_group=4 serverinstance_permissions_version=19 serverinstance_pending_connections_per_ip=0",
		"serverrequestconnectioninfo": "connection_filetransfer_bandwidth_sent=0 connection_filetransfer_bandwidth_received=0 connection_filetransfer_bytes_sent_total=617 connection_filetransfer_bytes_received_total=0 connection_packets_sent_total=926413 connection_bytes_sent_total=92911395 connection_packets_received_total=650335 connection_bytes_received_total=61940731 connection_bandwidth_sent_last_second_total=0 connection_bandwidth_sent_last_minute_total=0 connection_bandwidth_received_last_second_total=0 connection_bandwidth_received_last_minute_total=0 connection_connected_time=49408 connection_packetloss_total=0.0000 connection_ping=0.0000",
		"channellist":                 "cid=499 pid=0 channel_order=0 channel_name=Default\\sChannel total_clients=1 channel_needed_subscribe_power=0",
		"clientlist":                  "clid=5 cid=7 client_database_id=40 client_nickname=ScP client_type=0 client_away=1 client_away_message=not\\shere",
		"clientdblist":                "cldbid=7 client_unique_identifier=DZhdQU58qyooEK4Fr8Ly738hEmc= client_nickname=MuhChy client_created=1259147468 client_lastconnected=1259421233",
		"whoami":                      "virtualserver_status=online virtualserver_id=18 virtualserver_unique_identifier=gNITtWtKs9+Uh3L4LKv8\\/YHsn5c= virtualserver_port=9987 client_id=94 client_channel_id=432 client_nickname=serveradmin\\sfrom\\s127.0.0.1:49725 client_database_id=1 client_login_name=serveradmin client_unique_identifier=serveradmin client_origin_server_id=0",
		cmdQuit:                       "",
	}
)

// newLockListener creates a new listener on the local IP.
func newLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// server is a mock TeamSpeak 3 server
type server struct {
	Addr     string
	Listener net.Listener

	t         *testing.T
	conns     map[net.Conn]struct{}
	done      chan struct{}
	wg        sync.WaitGroup
	noHeader  bool
	noBanner  bool
	failConn  bool
	badHeader bool
	mtx       sync.Mutex
}

// sconn represents a server connection
type sconn struct {
	id int
	net.Conn
}

// newServer returns a running server or nil if an error occurred.
func newServer(t *testing.T) *server {
	s := newServerStopped(t)
	s.Start()

	return s
}

// newServerStopped returns a stopped servers or nil if an error occurred.
func newServerStopped(t *testing.T) *server {
	l, err := newLocalListener()
	if !assert.NoError(t, err) {
		return nil
	}

	s := &server{
		Listener: l,
		conns:    make(map[net.Conn]struct{}),
		done:     make(chan struct{}),
		t:        t,
	}
	s.Addr = s.Listener.Addr().String()
	return s
}

// Start starts the server.
func (s *server) Start() {
	s.wg.Add(1)
	go s.serve()
}

// server processes incoming requests until signaled to stop with Close.
func (s *server) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.running() {
				assert.NoError(s.t, err)
			}
			return
		}
		s.wg.Add(1)
		go s.handle(conn)
	}
}

// writeResponse writes the given msg followed by an error (ok) response.
// If msg is empty the only the error (ok) rsponse is sent.
func (s *server) writeResponse(c *sconn, msg string) error {
	if msg != "" {
		if err := s.write(c.Conn, msg); err != nil {
			return err
		}
	}

	return s.write(c.Conn, errOK)
}

// write writes msg to conn.
func (s *server) write(conn net.Conn, msg string) error {
	_, err := conn.Write([]byte(msg + "\n\r"))
	if s.running() {
		assert.NoError(s.t, err)
	}

	return err
}

// running returns true unless Close has been called, false otherwise.
func (s *server) running() bool {
	select {
	case <-s.done:
		return false
	default:
		return true
	}
}

// handle handles a client connection.
func (s *server) handle(conn net.Conn) {
	s.mtx.Lock()
	s.conns[conn] = struct{}{}
	s.mtx.Unlock()
	defer func() {
		s.closeConn(conn)
		s.wg.Done()
	}()

	if s.failConn {
		return
	}

	sc := bufio.NewScanner(bufio.NewReader(conn))
	sc.Split(bufio.ScanLines)

	if !s.noHeader {
		if s.badHeader {
			if err := s.write(conn, "bad"); err != nil {
				return
			}
		} else {
			if err := s.write(conn, connectHeader); err != nil {
				return
			}
		}

		if !s.noBanner {
			if err := s.write(conn, banner); err != nil {
				return
			}
		}
	}

	c := &sconn{Conn: conn}
	for sc.Scan() {
		l := sc.Text()
		parts := strings.Split(l, " ")
		resp, ok := commands[parts[0]]
		var err error
		if ok {
			err = s.writeResponse(c, resp)
		} else if parts[0] == "disconnect" {
			return
		} else {
			err = s.write(c, errUnknownCmd)
		}
		if err != nil || parts[0] == cmdQuit {
			return
		}
	}

	if err := sc.Err(); err != nil && s.running() {
		assert.NoError(s.t, err)
	}
}

// closeConn closes a client connection and removes it from our map of connections.
func (s *server) closeConn(conn net.Conn) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	conn.Close() // nolint: errcheck
	delete(s.conns, conn)
}

// Close cleanly shuts down the server.
func (s *server) Close() error {
	close(s.done)
	err := s.Listener.Close()
	s.mtx.Lock()
	for c := range s.conns {
		if err2 := c.Close(); err2 != nil && err == nil {
			err = err2
		}
	}
	s.mtx.Unlock()
	s.wg.Wait()

	return err
}
