package syslog

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	"github.com/influxdata/telegraf/testutil"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

type testCasePacket struct {
	name           string
	data           []byte
	wantBestEffort telegraf.Metric
	wantStrict     telegraf.Metric
	werr           bool
}

type testCaseStream struct {
	name           string
	data           []byte
	wantBestEffort []telegraf.Metric
	wantStrict     []telegraf.Metric
	werr           int // how many errors we expect in the strict mode?
}

func newUDPSyslogReceiver(address string, bestEffort bool, rfc syslogRFC) *Syslog {
	return &Syslog{
		Address: address,
		now: func() time.Time {
			return defaultTime
		},
		BestEffort:     bestEffort,
		SyslogStandard: rfc,
		Separator:      "_",
	}
}

func newTCPSyslogReceiver(address string, keepAlive *config.Duration, maxConn int, bestEffort bool, f framing.Framing) *Syslog {
	d := config.Duration(defaultReadTimeout)
	s := &Syslog{
		Address: address,
		now: func() time.Time {
			return defaultTime
		},
		Framing:        f,
		ReadTimeout:    &d,
		BestEffort:     bestEffort,
		SyslogStandard: syslogRFC5424,
		Separator:      "_",
	}
	if keepAlive != nil {
		s.KeepAlivePeriod = keepAlive
	}
	if maxConn > 0 {
		s.MaxConnections = maxConn
	}

	return s
}
