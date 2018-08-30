//go:generate go run -tags generate snmp_mocks_generate.go
package snmp

import (
	"fmt"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/soniah/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSNMPConnection struct {
	host   string
	values map[string]interface{}
}

func (tsc *testSNMPConnection) Host() string {
	return tsc.host
}

func (tsc *testSNMPConnection) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	sp := &gosnmp.SnmpPacket{}
	for _, oid := range oids {
		v, ok := tsc.values[oid]
		if !ok {
			sp.Variables = append(sp.Variables, gosnmp.SnmpPDU{
				Name: oid,
				Type: gosnmp.NoSuchObject,
			})
			continue
		}
		sp.Variables = append(sp.Variables, gosnmp.SnmpPDU{
			Name:  oid,
			Value: v,
		})
	}
	return sp, nil
}
func (tsc *testSNMPConnection) Walk(oid string, wf gosnmp.WalkFunc) error {
	for void, v := range tsc.values {
		if void == oid || (len(void) > len(oid) && void[:len(oid)+1] == oid+".") {
			if err := wf(gosnmp.SnmpPDU{
				Name:  void,
				Value: v,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

var tsc = &testSNMPConnection{
	host: "tsc",
	values: map[string]interface{}{
		".1.0.0.0.1.1.0":     "foo",
		".1.0.0.0.1.1.1":     []byte("bar"),
		".1.0.0.0.1.1.2":     []byte(""),
		".1.0.0.0.1.102":     "bad",
		".1.0.0.0.1.2.0":     1,
		".1.0.0.0.1.2.1":     2,
		".1.0.0.0.1.2.2":     0,
		".1.0.0.0.1.3.0":     "0.123",
		".1.0.0.0.1.3.1":     "0.456",
		".1.0.0.0.1.3.2":     "0.000",
		".1.0.0.0.1.3.3":     "9.999",
		".1.0.0.0.1.5.0":     123456,
		".1.0.0.1.1":         "baz",
		".1.0.0.1.2":         234,
		".1.0.0.1.3":         []byte("byte slice"),
		".1.0.0.2.1.5.0.9.9": 11,
		".1.0.0.2.1.5.1.9.9": 22,
	},
}

func TestSampleConfig(t *testing.T) {
	conf := struct {
		Inputs struct {
			Snmp []*Snmp
		}
	}{}
	err := toml.Unmarshal([]byte("[[inputs.snmp]]\n"+(*Snmp)(nil).SampleConfig()), &conf)
	assert.NoError(t, err)

	s := Snmp{
		Agents:         []string{"127.0.0.1:161"},
		Timeout:        internal.Duration{Duration: 5 * time.Second},
		Version:        2,
		Community:      "public",
		MaxRepetitions: 10,
		Retries:        3,

		Name: "system",
		Fields: []Field{
			{Name: "hostname", Oid: ".1.0.0.1.1"},
			{Name: "uptime", Oid: ".1.0.0.1.2"},
			{Name: "load", Oid: ".1.0.0.1.3"},
			{Oid: "HOST-RESOURCES-MIB::hrMemorySize"},
		},
		Tables: []Table{
			{
				Name:        "remote_servers",
				InheritTags: []string{"hostname"},
				Fields: []Field{
					{Name: "server", Oid: ".1.0.0.0.1.0", IsTag: true},
					{Name: "connections", Oid: ".1.0.0.0.1.1"},
					{Name: "latency", Oid: ".1.0.0.0.1.2"},
				},
			},
			{
				Oid: "HOST-RESOURCES-MIB::hrNetworkTable",
			},
		},
	}
	assert.Equal(t, &s, conf.Inputs.Snmp[0])
}

func TestFieldInit(t *testing.T) {
	translations := []struct {
		inputOid           string
		inputName          string
		inputConversion    string
		expectedOid        string
		expectedName       string
		expectedConversion string
	}{
		{".1.2.3", "foo", "", ".1.2.3", "foo", ""},
		{".iso.2.3", "foo", "", ".1.2.3", "foo", ""},
		{".1.0.0.0.1.1", "", "", ".1.0.0.0.1.1", "server", ""},
		{".1.0.0.0.1.1.0", "", "", ".1.0.0.0.1.1.0", "server.0", ""},
		{".999", "", "", ".999", ".999", ""},
		{"TEST::server", "", "", ".1.0.0.0.1.1", "server", ""},
		{"TEST::server.0", "", "", ".1.0.0.0.1.1.0", "server.0", ""},
		{"TEST::server", "foo", "", ".1.0.0.0.1.1", "foo", ""},
		{"IF-MIB::ifPhysAddress.1", "", "", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "hwaddr"},
		{"IF-MIB::ifPhysAddress.1", "", "none", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "none"},
		{"BRIDGE-MIB::dot1dTpFdbAddress.1", "", "", ".1.3.6.1.2.1.17.4.3.1.1.1", "dot1dTpFdbAddress.1", "hwaddr"},
		{"TCP-MIB::tcpConnectionLocalAddress.1", "", "", ".1.3.6.1.2.1.6.19.1.2.1", "tcpConnectionLocalAddress.1", "ipaddr"},
	}

	for _, txl := range translations {
		f := Field{Oid: txl.inputOid, Name: txl.inputName, Conversion: txl.inputConversion}
		err := f.init()
		if !assert.NoError(t, err, "inputOid='%s' inputName='%s'", txl.inputOid, txl.inputName) {
			continue
		}
		assert.Equal(t, txl.expectedOid, f.Oid, "inputOid='%s' inputName='%s' inputConversion='%s'", txl.inputOid, txl.inputName, txl.inputConversion)
		assert.Equal(t, txl.expectedName, f.Name, "inputOid='%s' inputName='%s' inputConversion='%s'", txl.inputOid, txl.inputName, txl.inputConversion)
	}
}

func TestTableInit(t *testing.T) {
	tbl := Table{
		Oid: ".1.0.0.0",
		Fields: []Field{
			{Oid: ".999", Name: "foo"},
			{Oid: "TEST::description", Name: "description", IsTag: true},
		},
	}
	err := tbl.init()
	require.NoError(t, err)

	assert.Equal(t, "testTable", tbl.Name)

	assert.Len(t, tbl.Fields, 5)
	assert.Contains(t, tbl.Fields, Field{Oid: ".999", Name: "foo", initialized: true})
	assert.Contains(t, tbl.Fields, Field{Oid: ".1.0.0.0.1.1", Name: "server", IsTag: true, initialized: true})
	assert.Contains(t, tbl.Fields, Field{Oid: ".1.0.0.0.1.2", Name: "connections", initialized: true})
	assert.Contains(t, tbl.Fields, Field{Oid: ".1.0.0.0.1.3", Name: "latency", initialized: true})
	assert.Contains(t, tbl.Fields, Field{Oid: ".1.0.0.0.1.4", Name: "description", IsTag: true, initialized: true})
}

func TestSnmpInit(t *testing.T) {
	s := &Snmp{
		Tables: []Table{
			{Oid: "TEST::testTable"},
		},
		Fields: []Field{
			{Oid: "TEST::hostname"},
		},
	}

	err := s.init()
	require.NoError(t, err)

	assert.Len(t, s.Tables[0].Fields, 4)
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.0.0.0.1.1", Name: "server", IsTag: true, initialized: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.0.0.0.1.2", Name: "connections", initialized: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.0.0.0.1.3", Name: "latency", initialized: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.0.0.0.1.4", Name: "description", initialized: true})

	assert.Equal(t, Field{
		Oid:         ".1.0.0.1.1",
		Name:        "hostname",
		initialized: true,
	}, s.Fields[0])
}

func TestSnmpInit_noTranslate(t *testing.T) {
	// override execCommand so it returns exec.ErrNotFound
	defer func(ec func(string, ...string) *exec.Cmd) { execCommand = ec }(execCommand)
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("snmptranslateExecErrNotFound")
	}

	s := &Snmp{
		Fields: []Field{
			{Oid: ".1.1.1.1", Name: "one", IsTag: true},
			{Oid: ".1.1.1.2", Name: "two"},
			{Oid: ".1.1.1.3"},
		},
		Tables: []Table{
			{Fields: []Field{
				{Oid: ".1.1.1.4", Name: "four", IsTag: true},
				{Oid: ".1.1.1.5", Name: "five"},
				{Oid: ".1.1.1.6"},
			}},
		},
	}

	err := s.init()
	require.NoError(t, err)

	assert.Equal(t, ".1.1.1.1", s.Fields[0].Oid)
	assert.Equal(t, "one", s.Fields[0].Name)
	assert.Equal(t, true, s.Fields[0].IsTag)

	assert.Equal(t, ".1.1.1.2", s.Fields[1].Oid)
	assert.Equal(t, "two", s.Fields[1].Name)
	assert.Equal(t, false, s.Fields[1].IsTag)

	assert.Equal(t, ".1.1.1.3", s.Fields[2].Oid)
	assert.Equal(t, ".1.1.1.3", s.Fields[2].Name)
	assert.Equal(t, false, s.Fields[2].IsTag)

	assert.Equal(t, ".1.1.1.4", s.Tables[0].Fields[0].Oid)
	assert.Equal(t, "four", s.Tables[0].Fields[0].Name)
	assert.Equal(t, true, s.Tables[0].Fields[0].IsTag)

	assert.Equal(t, ".1.1.1.5", s.Tables[0].Fields[1].Oid)
	assert.Equal(t, "five", s.Tables[0].Fields[1].Name)
	assert.Equal(t, false, s.Tables[0].Fields[1].IsTag)

	assert.Equal(t, ".1.1.1.6", s.Tables[0].Fields[2].Oid)
	assert.Equal(t, ".1.1.1.6", s.Tables[0].Fields[2].Name)
	assert.Equal(t, false, s.Tables[0].Fields[2].IsTag)
}

func TestGetSNMPConnection_v2(t *testing.T) {
	s := &Snmp{
		Agents:    []string{"1.2.3.4:567", "1.2.3.4"},
		Timeout:   internal.Duration{Duration: 3 * time.Second},
		Retries:   4,
		Version:   2,
		Community: "foo",
	}
	err := s.init()
	require.NoError(t, err)

	gsc, err := s.getConnection(0)
	require.NoError(t, err)
	gs := gsc.(gosnmpWrapper)
	assert.Equal(t, "1.2.3.4", gs.Target)
	assert.EqualValues(t, 567, gs.Port)
	assert.Equal(t, gosnmp.Version2c, gs.Version)
	assert.Equal(t, "foo", gs.Community)

	gsc, err = s.getConnection(1)
	require.NoError(t, err)
	gs = gsc.(gosnmpWrapper)
	assert.Equal(t, "1.2.3.4", gs.Target)
	assert.EqualValues(t, 161, gs.Port)
}

func TestGetSNMPConnection_v3(t *testing.T) {
	s := &Snmp{
		Agents:         []string{"1.2.3.4"},
		Version:        3,
		MaxRepetitions: 20,
		ContextName:    "mycontext",
		SecLevel:       "authPriv",
		SecName:        "myuser",
		AuthProtocol:   "md5",
		AuthPassword:   "password123",
		PrivProtocol:   "des",
		PrivPassword:   "321drowssap",
		EngineID:       "myengineid",
		EngineBoots:    1,
		EngineTime:     2,
	}
	err := s.init()
	require.NoError(t, err)

	gsc, err := s.getConnection(0)
	require.NoError(t, err)
	gs := gsc.(gosnmpWrapper)
	assert.Equal(t, gs.Version, gosnmp.Version3)
	sp := gs.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	assert.Equal(t, "1.2.3.4", gsc.Host())
	assert.EqualValues(t, 20, gs.MaxRepetitions)
	assert.Equal(t, "mycontext", gs.ContextName)
	assert.Equal(t, gosnmp.AuthPriv, gs.MsgFlags&gosnmp.AuthPriv)
	assert.Equal(t, "myuser", sp.UserName)
	assert.Equal(t, gosnmp.MD5, sp.AuthenticationProtocol)
	assert.Equal(t, "password123", sp.AuthenticationPassphrase)
	assert.Equal(t, gosnmp.DES, sp.PrivacyProtocol)
	assert.Equal(t, "321drowssap", sp.PrivacyPassphrase)
	assert.Equal(t, "myengineid", sp.AuthoritativeEngineID)
	assert.EqualValues(t, 1, sp.AuthoritativeEngineBoots)
	assert.EqualValues(t, 2, sp.AuthoritativeEngineTime)
}

func TestGetSNMPConnection_caching(t *testing.T) {
	s := &Snmp{
		Agents: []string{"1.2.3.4", "1.2.3.5", "1.2.3.5"},
	}
	err := s.init()
	require.NoError(t, err)
	gs1, err := s.getConnection(0)
	require.NoError(t, err)
	gs2, err := s.getConnection(0)
	require.NoError(t, err)
	gs3, err := s.getConnection(1)
	require.NoError(t, err)
	gs4, err := s.getConnection(2)
	require.NoError(t, err)
	assert.True(t, gs1 == gs2)
	assert.False(t, gs2 == gs3)
	assert.False(t, gs3 == gs4)
}

func TestGosnmpWrapper_walk_retry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test due to random failures.")
	}
	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	defer srvr.Close()
	require.NoError(t, err)
	reqCount := 0
	// Set up a WaitGroup to wait for the server goroutine to exit and protect
	// reqCount.
	// Even though simultaneous access is impossible because the server will be
	// blocked on ReadFrom, without this the race detector gets unhappy.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		for {
			_, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}
			reqCount++

			srvr.WriteTo([]byte{'X'}, addr) // will cause decoding error
		}
	}()

	gs := &gosnmp.GoSNMP{
		Target:    srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:      uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Version:   gosnmp.Version2c,
		Community: "public",
		Timeout:   time.Millisecond * 10,
		Retries:   1,
	}
	err = gs.Connect()
	require.NoError(t, err)
	conn := gs.Conn

	gsw := gosnmpWrapper{gs}
	err = gsw.Walk(".1.0.0", func(_ gosnmp.SnmpPDU) error { return nil })
	srvr.Close()
	wg.Wait()
	assert.Error(t, err)
	assert.False(t, gs.Conn == conn)
	assert.Equal(t, (gs.Retries+1)*2, reqCount)
}

func TestGosnmpWrapper_get_retry(t *testing.T) {
	// TODO: Fix this test
	t.Skip("Test failing too often, skip for now and revisit later.")
	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	defer srvr.Close()
	require.NoError(t, err)
	reqCount := 0
	// Set up a WaitGroup to wait for the server goroutine to exit and protect
	// reqCount.
	// Even though simultaneous access is impossible because the server will be
	// blocked on ReadFrom, without this the race detector gets unhappy.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		for {
			_, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}
			reqCount++

			srvr.WriteTo([]byte{'X'}, addr) // will cause decoding error
		}
	}()

	gs := &gosnmp.GoSNMP{
		Target:    srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:      uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Version:   gosnmp.Version2c,
		Community: "public",
		Timeout:   time.Millisecond * 10,
		Retries:   1,
	}
	err = gs.Connect()
	require.NoError(t, err)
	conn := gs.Conn

	gsw := gosnmpWrapper{gs}
	_, err = gsw.Get([]string{".1.0.0"})
	srvr.Close()
	wg.Wait()
	assert.Error(t, err)
	assert.False(t, gs.Conn == conn)
	assert.Equal(t, (gs.Retries+1)*2, reqCount)
}

func TestTableBuild_walk(t *testing.T) {
	tbl := Table{
		Name:       "mytable",
		IndexAsTag: true,
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.0.1.2",
			},
			{
				Name:       "myfield3",
				Oid:        ".1.0.0.0.1.3",
				Conversion: "float",
			},
			{
				Name:           "myfield4",
				Oid:            ".1.0.0.2.1.5",
				OidIndexSuffix: ".9.9",
			},
			{
				Name:           "myfield5",
				Oid:            ".1.0.0.2.1.5",
				OidIndexLength: 1,
			},
		},
	}

	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)

	assert.Equal(t, tb.Name, "mytable")
	rtr1 := RTableRow{
		Tags: map[string]string{
			"myfield1": "foo",
			"index":    "0",
		},
		Fields: map[string]interface{}{
			"myfield2": 1,
			"myfield3": float64(0.123),
			"myfield4": 11,
			"myfield5": 11,
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"myfield1": "bar",
			"index":    "1",
		},
		Fields: map[string]interface{}{
			"myfield2": 2,
			"myfield3": float64(0.456),
			"myfield4": 22,
			"myfield5": 22,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"index": "2",
		},
		Fields: map[string]interface{}{
			"myfield2": 0,
			"myfield3": float64(0.0),
		},
	}
	rtr4 := RTableRow{
		Tags: map[string]string{
			"index": "3",
		},
		Fields: map[string]interface{}{
			"myfield3": float64(9.999),
		},
	}
	assert.Len(t, tb.Rows, 4)
	assert.Contains(t, tb.Rows, rtr1)
	assert.Contains(t, tb.Rows, rtr2)
	assert.Contains(t, tb.Rows, rtr3)
	assert.Contains(t, tb.Rows, rtr4)
}

func TestTableBuild_noWalk(t *testing.T) {
	tbl := Table{
		Name: "mytable",
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
			{
				Name:  "myfield3",
				Oid:   ".1.0.0.1.2",
				IsTag: true,
			},
			{
				Name: "empty",
				Oid:  ".1.0.0.0.1.1.2",
			},
			{
				Name: "noexist",
				Oid:  ".1.2.3.4.5",
			},
		},
	}

	tb, err := tbl.Build(tsc, false)
	require.NoError(t, err)

	rtr := RTableRow{
		Tags:   map[string]string{"myfield1": "baz", "myfield3": "234"},
		Fields: map[string]interface{}{"myfield2": 234},
	}
	assert.Len(t, tb.Rows, 1)
	assert.Contains(t, tb.Rows, rtr)
}

func TestGather(t *testing.T) {
	s := &Snmp{
		Agents: []string{"TestGather"},
		Name:   "mytable",
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
			{
				Name: "myfield3",
				Oid:  "1.0.0.1.1",
			},
		},
		Tables: []Table{
			{
				Name:        "myOtherTable",
				InheritTags: []string{"myfield1"},
				Fields: []Field{
					{
						Name: "myOtherField",
						Oid:  ".1.0.0.0.1.5",
					},
				},
			},
		},

		connectionCache: []snmpConnection{
			tsc,
		},
		initialized: true,
	}
	acc := &testutil.Accumulator{}

	tstart := time.Now()
	s.Gather(acc)
	tstop := time.Now()

	require.Len(t, acc.Metrics, 2)

	m := acc.Metrics[0]
	assert.Equal(t, "mytable", m.Measurement)
	assert.Equal(t, "tsc", m.Tags["agent_host"])
	assert.Equal(t, "baz", m.Tags["myfield1"])
	assert.Len(t, m.Fields, 2)
	assert.Equal(t, 234, m.Fields["myfield2"])
	assert.Equal(t, "baz", m.Fields["myfield3"])
	assert.True(t, tstart.Before(m.Time))
	assert.True(t, tstop.After(m.Time))

	m2 := acc.Metrics[1]
	assert.Equal(t, "myOtherTable", m2.Measurement)
	assert.Equal(t, "tsc", m2.Tags["agent_host"])
	assert.Equal(t, "baz", m2.Tags["myfield1"])
	assert.Len(t, m2.Fields, 1)
	assert.Equal(t, 123456, m2.Fields["myOtherField"])
}

func TestGather_host(t *testing.T) {
	s := &Snmp{
		Agents: []string{"TestGather"},
		Name:   "mytable",
		Fields: []Field{
			{
				Name:  "host",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
		},

		connectionCache: []snmpConnection{
			tsc,
		},
		initialized: true,
	}

	acc := &testutil.Accumulator{}

	s.Gather(acc)

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]
	assert.Equal(t, "baz", m.Tags["host"])
}

func TestFieldConvert(t *testing.T) {
	testTable := []struct {
		input    interface{}
		conv     string
		expected interface{}
	}{
		{[]byte("foo"), "", string("foo")},
		{"0.123", "float", float64(0.123)},
		{[]byte("0.123"), "float", float64(0.123)},
		{float32(0.123), "float", float64(float32(0.123))},
		{float64(0.123), "float", float64(0.123)},
		{123, "float", float64(123)},
		{123, "float(0)", float64(123)},
		{123, "float(4)", float64(0.0123)},
		{int8(123), "float(3)", float64(0.123)},
		{int16(123), "float(3)", float64(0.123)},
		{int32(123), "float(3)", float64(0.123)},
		{int64(123), "float(3)", float64(0.123)},
		{uint(123), "float(3)", float64(0.123)},
		{uint8(123), "float(3)", float64(0.123)},
		{uint16(123), "float(3)", float64(0.123)},
		{uint32(123), "float(3)", float64(0.123)},
		{uint64(123), "float(3)", float64(0.123)},
		{"123", "int", int64(123)},
		{[]byte("123"), "int", int64(123)},
		{float32(12.3), "int", int64(12)},
		{float64(12.3), "int", int64(12)},
		{int(123), "int", int64(123)},
		{int8(123), "int", int64(123)},
		{int16(123), "int", int64(123)},
		{int32(123), "int", int64(123)},
		{int64(123), "int", int64(123)},
		{uint(123), "int", int64(123)},
		{uint8(123), "int", int64(123)},
		{uint16(123), "int", int64(123)},
		{uint32(123), "int", int64(123)},
		{uint64(123), "int", int64(123)},
		{[]byte("abcdef"), "hwaddr", "61:62:63:64:65:66"},
		{"abcdef", "hwaddr", "61:62:63:64:65:66"},
		{[]byte("abcd"), "ipaddr", "97.98.99.100"},
		{"abcd", "ipaddr", "97.98.99.100"},
		{[]byte("abcdefghijklmnop"), "ipaddr", "6162:6364:6566:6768:696a:6b6c:6d6e:6f70"},
	}

	for _, tc := range testTable {
		act, err := fieldConvert(tc.conv, tc.input)
		if !assert.NoError(t, err, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected) {
			continue
		}
		assert.EqualValues(t, tc.expected, act, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected)
	}
}

func TestSnmpTranslateCache_miss(t *testing.T) {
	snmpTranslateCaches = nil
	oid := "IF-MIB::ifPhysAddress.1"
	mibName, oidNum, oidText, conversion, err := snmpTranslate(oid)
	assert.Len(t, snmpTranslateCaches, 1)
	stc := snmpTranslateCaches[oid]
	require.NotNil(t, stc)
	assert.Equal(t, mibName, stc.mibName)
	assert.Equal(t, oidNum, stc.oidNum)
	assert.Equal(t, oidText, stc.oidText)
	assert.Equal(t, conversion, stc.conversion)
	assert.Equal(t, err, stc.err)
}

func TestSnmpTranslateCache_hit(t *testing.T) {
	snmpTranslateCaches = map[string]snmpTranslateCache{
		"foo": snmpTranslateCache{
			mibName:    "a",
			oidNum:     "b",
			oidText:    "c",
			conversion: "d",
			err:        fmt.Errorf("e"),
		},
	}
	mibName, oidNum, oidText, conversion, err := snmpTranslate("foo")
	assert.Equal(t, "a", mibName)
	assert.Equal(t, "b", oidNum)
	assert.Equal(t, "c", oidText)
	assert.Equal(t, "d", conversion)
	assert.Equal(t, fmt.Errorf("e"), err)
	snmpTranslateCaches = nil
}

func TestSnmpTableCache_miss(t *testing.T) {
	snmpTableCaches = nil
	oid := ".1.0.0.0"
	mibName, oidNum, oidText, fields, err := snmpTable(oid)
	assert.Len(t, snmpTableCaches, 1)
	stc := snmpTableCaches[oid]
	require.NotNil(t, stc)
	assert.Equal(t, mibName, stc.mibName)
	assert.Equal(t, oidNum, stc.oidNum)
	assert.Equal(t, oidText, stc.oidText)
	assert.Equal(t, fields, stc.fields)
	assert.Equal(t, err, stc.err)
}

func TestSnmpTableCache_hit(t *testing.T) {
	snmpTableCaches = map[string]snmpTableCache{
		"foo": snmpTableCache{
			mibName: "a",
			oidNum:  "b",
			oidText: "c",
			fields:  []Field{{Name: "d"}},
			err:     fmt.Errorf("e"),
		},
	}
	mibName, oidNum, oidText, fields, err := snmpTable("foo")
	assert.Equal(t, "a", mibName)
	assert.Equal(t, "b", oidNum)
	assert.Equal(t, "c", oidText)
	assert.Equal(t, []Field{{Name: "d"}}, fields)
	assert.Equal(t, fmt.Errorf("e"), err)
}

func TestError(t *testing.T) {
	e := fmt.Errorf("nested error")
	err := Errorf(e, "top error %d", 123)
	require.Error(t, err)

	ne, ok := err.(NestedError)
	require.True(t, ok)
	assert.Equal(t, e, ne.NestedErr)

	assert.Contains(t, err.Error(), "top error 123")
	assert.Contains(t, err.Error(), "nested error")
}
