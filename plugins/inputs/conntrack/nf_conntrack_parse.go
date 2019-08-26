// +build linux

package conntrack

import (
	"io"
	"strings"
)

var (
	byteSpace   = byte(' ')
	byteEqual   = byte('=')
	byteTwoDots = byte(':')
	byteNewLine = byte('\n')
)

func newNfConntrack(f io.Reader) *nfConntrack {
	nf := &nfConntrack{
		counters: make(map[string]int64),
	}
	nf.parseProcNfConntrack(f)
	return nf
}

type nfConntrack struct {
	counters map[string]int64
}

func (nf *nfConntrack) parseProcNfConntrack(f io.Reader) error {
	b := make([]byte, 32*1024)
	block := make([]byte, 0)
	pos := 0

	row := &rowConn{}

	stackParseFn := func(n int) {
		for _, v := range b[:n] {
			switch v {
			case byteSpace:
				if len(block) > 0 {
					switch pos {
					case 0:
						row.version = string(block)
					case 2:
						row.dframe = string(block)
					}

					row.parse(block, pos)
					block = block[:0]
					pos++
				}
			case byteNewLine:
				if len(block) > 0 {
					row.parse(block, pos)
					block = block[:0]

					switch row.dframe {
					case "tcp":
						nf.counters[row.dframe+"_"+strings.ToLower(row.status)]++
					case "udp":
						if row.unreplied {
							nf.counters[row.dframe+"_unreplied"]++
						} else {
							nf.counters[row.dframe]++
						}
					}

					row.Reset()
				}
				pos = 0
			default:
				block = append(block, v)
			}
		}
	}

	for {
		n, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				stackParseFn(n)
				return nil
			}
			return err
		}
		stackParseFn(n)
	}
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
