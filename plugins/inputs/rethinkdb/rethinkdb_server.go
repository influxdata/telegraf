package rethinkdb

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/gorethink/gorethink.v3"

	"github.com/influxdata/telegraf"
)

type Server struct {
	URL          *url.URL
	session      *gorethink.Session
	serverStatus serverStatus
}

func (s *Server) gatherData(acc telegraf.Accumulator) error {
	if err := s.getServerStatus(); err != nil {
		return fmt.Errorf("failed to get server_status, %s", err)
	}

	if err := s.validateVersion(); err != nil {
		return fmt.Errorf("failed version validation, %s", err.Error())
	}

	if err := s.addClusterStats(acc); err != nil {
		return fmt.Errorf("error adding cluster stats, %s", err.Error())
	}

	if err := s.addMemberStats(acc); err != nil {
		return fmt.Errorf("error adding member stats, %s", err.Error())
	}

	if err := s.addTablesStats(acc); err != nil {
		return fmt.Errorf("error adding table stats, %s", err.Error())
	}

	return nil
}

func (s *Server) validateVersion() error {
	if s.serverStatus.Process.Version == "" {
		return errors.New("could not determine the RethinkDB server version: process.version key missing")
	}

	versionRegexp := regexp.MustCompile(`\d.\d.\d`)
	versionString := versionRegexp.FindString(s.serverStatus.Process.Version)
	if versionString == "" {
		return fmt.Errorf("could not determine the RethinkDB server version: malformed version string (%v)", s.serverStatus.Process.Version)
	}

	majorVersion, err := strconv.Atoi(strings.Split(versionString, "")[0])
	if err != nil || majorVersion < 2 {
		return fmt.Errorf("unsupported major version %s", versionString)
	}
	return nil
}

func (s *Server) getServerStatus() error {
	cursor, err := gorethink.DB("rethinkdb").Table("server_status").Run(s.session)
	if err != nil {
		return err
	}

	if cursor.IsNil() {
		return errors.New("could not determine the RethinkDB server version: no rows returned from the server_status table")
	}
	defer cursor.Close()
	var serverStatuses []serverStatus
	err = cursor.All(&serverStatuses)
	if err != nil {
		return errors.New("could not parse server_status results")
	}
	host, port, err := net.SplitHostPort(s.URL.Host)
	if err != nil {
		return fmt.Errorf("unable to determine provided hostname from %s", s.URL.Host)
	}
	driverPort, _ := strconv.Atoi(port)
	for _, ss := range serverStatuses {
		for _, address := range ss.Network.Addresses {
			if address.Host == host && ss.Network.DriverPort == driverPort {
				s.serverStatus = ss
				return nil
			}
		}
	}

	return fmt.Errorf("unable to determine host id from server_status with %s", s.URL.Host)
}

func (s *Server) getDefaultTags() map[string]string {
	tags := make(map[string]string)
	tags["rethinkdb_host"] = s.URL.Host
	tags["rethinkdb_hostname"] = s.serverStatus.Network.Hostname
	return tags
}

var ClusterTracking = []string{
	"active_clients",
	"clients",
	"queries_per_sec",
	"read_docs_per_sec",
	"written_docs_per_sec",
}

func (s *Server) addClusterStats(acc telegraf.Accumulator) error {
	cursor, err := gorethink.DB("rethinkdb").Table("stats").Get([]string{"cluster"}).Run(s.session)
	if err != nil {
		return fmt.Errorf("cluster stats query error, %s", err.Error())
	}
	defer cursor.Close()
	var clusterStats stats
	if err := cursor.One(&clusterStats); err != nil {
		return fmt.Errorf("failure to parse cluster stats, %s", err.Error())
	}

	tags := s.getDefaultTags()
	tags["type"] = "cluster"
	clusterStats.Engine.AddEngineStats(ClusterTracking, acc, tags)
	return nil
}

var MemberTracking = []string{
	"active_clients",
	"clients",
	"queries_per_sec",
	"total_queries",
	"read_docs_per_sec",
	"total_reads",
	"written_docs_per_sec",
	"total_writes",
}

func (s *Server) addMemberStats(acc telegraf.Accumulator) error {
	cursor, err := gorethink.DB("rethinkdb").Table("stats").Get([]string{"server", s.serverStatus.ID}).Run(s.session)
	if err != nil {
		return fmt.Errorf("member stats query error, %s", err.Error())
	}
	defer cursor.Close()
	var memberStats stats
	if err := cursor.One(&memberStats); err != nil {
		return fmt.Errorf("failure to parse member stats, %s", err.Error())
	}

	tags := s.getDefaultTags()
	tags["type"] = "member"
	memberStats.Engine.AddEngineStats(MemberTracking, acc, tags)
	return nil
}

var TableTracking = []string{
	"read_docs_per_sec",
	"total_reads",
	"written_docs_per_sec",
	"total_writes",
}

func (s *Server) addTablesStats(acc telegraf.Accumulator) error {
	tablesCursor, err := gorethink.DB("rethinkdb").Table("table_status").Run(s.session)
	if err != nil {
		return fmt.Errorf("table stats query error, %s", err.Error())
	}

	defer tablesCursor.Close()
	var tables []tableStatus
	err = tablesCursor.All(&tables)
	if err != nil {
		return errors.New("could not parse table_status results")
	}
	for _, table := range tables {
		err = s.addTableStats(acc, table)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) addTableStats(acc telegraf.Accumulator, table tableStatus) error {
	cursor, err := gorethink.DB("rethinkdb").Table("stats").
		Get([]string{"table_server", table.ID, s.serverStatus.ID}).
		Run(s.session)
	if err != nil {
		return fmt.Errorf("table stats query error, %s", err.Error())
	}
	defer cursor.Close()

	var ts tableStats
	if err := cursor.One(&ts); err != nil {
		return fmt.Errorf("failure to parse table stats, %s", err.Error())
	}

	tags := s.getDefaultTags()
	tags["type"] = "data"
	tags["ns"] = fmt.Sprintf("%s.%s", table.DB, table.Name)
	ts.Engine.AddEngineStats(TableTracking, acc, tags)
	ts.Storage.AddStats(acc, tags)

	return nil
}
