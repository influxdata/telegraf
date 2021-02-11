// +build linux

package conntrack

import (
	"strings"
)

var (
	byteSpace   = byte(' ')
	byteEqual   = byte('=')
	byteTwoDots = byte(':')
	byteNewLine = byte('\n')
)

func newNfConntrack() *nfConntrack {
	nf := &nfConntrack{
		counters: make(map[string]int64),
		row:      &rowConn{},
		block:    make([]byte, 0),
		pos:      0,
	}
	return nf
}

type nfConntrack struct {
	counters map[string]int64
	row      *rowConn
	block    []byte
	pos      int
}

func (nf *nfConntrack) Write(b []byte) (int, error) {
	for _, v := range b {
		switch v {
		case byteSpace:
			if len(nf.block) > 0 {
				switch nf.pos {
				case 0:
					nf.row.version = string(nf.block)
				case 2:
					nf.row.dframe = string(nf.block)
				}

				nf.row.parse(nf.block, nf.pos)
				nf.block = nf.block[:0]
				nf.pos++
			}
		case byteNewLine:
			if len(nf.block) > 0 {
				nf.row.parse(nf.block, nf.pos)
				nf.block = nf.block[:0]
				switch nf.row.dframe {
				case "tcp":
					nf.counters[nf.row.dframe+"_"+strings.ToLower(nf.row.status)]++
				case "udp":
					if nf.row.unreplied {
						nf.counters[nf.row.dframe+"_unreplied"]++
					} else {
						nf.counters[nf.row.dframe]++
					}
				}
				nf.row.Reset()
			}
			nf.pos = 0
		default:
			nf.block = append(nf.block, v)
		}
	}
	return len(b), nil
}

type rowConn struct {
	version   string // ipv4, ipv6
	dframe    string // tcp, udp
	status    string // TIME_WAIT, ESTABLISHED...
	unreplied bool   // [UNREPLIED]
}

func (row *rowConn) parse(block []byte, pos int) {
	switch pos {
	case 0:
		row.version = string(block)
	case 2:
		row.dframe = string(block)
	default:
		switch row.dframe {
		case "tcp":
			row.parseTCP(block, pos)
		case "udp":
			row.parseUDP(block, pos)
		}
	}

}

func (row *rowConn) parseTCP(block []byte, pos int) {
	// 0 ipv4
	// 1 2
	// 2 tcp
	// 3 6
	// 4 32
	// 5 TIME_WAIT
	// 6 src=192.168.0.1
	// 7 dst=8.8.8.8
	// 8 sport=12842
	// 9 dport=80
	// 10 src=8.8.8.8
	// 11 dst=10.255.244.244
	// 12 sport=80
	// 13 dport=12842
	// 14 [ASSURED]
	// 15 mark=0
	// 16 zone=0
	// 17 use=2
	switch pos {
	case 5:
		row.status = string(block)
	}
}

func (row *rowConn) parseUDP(block []byte, pos int) {
	// 0 ipv4
	// 1 2
	// 2 udp
	// 3 17
	// 4 16
	// 5 src=192.168.0.1
	// 6 dst=8.8.8.8
	// 7 sport=51162
	// 8 dport=123
	// 9 [UNREPLIED]
	// 10 src=8.8.8.8
	// 11 dst=10.255.244.37
	// 12 sport=123
	// 13 dport=28166
	switch pos {
	case 9:
		if string(block) == "[UNREPLIED]" {
			row.unreplied = true
		}
	}
}

func (r *rowConn) Reset() {
	n := &rowConn{}
	*r = *n
}
