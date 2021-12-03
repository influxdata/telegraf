package phpfpm

import (
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

// Create an fcgi client
func newFcgiClient(h string, args ...interface{}) (*conn, error) {
	var con net.Conn
	if len(args) != 1 {
		return nil, errors.New("fcgi: not enough params")
	}

	var err error
	switch args[0].(type) {
	case int:
		addr := h + ":" + strconv.FormatInt(int64(args[0].(int)), 10)
		con, err = net.Dial("tcp", addr)
	case string:
		laddr := net.UnixAddr{Name: args[0].(string), Net: h}
		con, err = net.DialUnix(h, nil, &laddr)
	default:
		err = errors.New("fcgi: we only accept int (port) or string (socket) params")
	}
	fcgi := &conn{
		rwc: con,
	}

	return fcgi, err
}

func (c *conn) Request(env map[string]string, requestData string) (retout []byte, reterr []byte, err error) {
	defer c.rwc.Close()
	var reqID uint16 = 1

	err = c.writeBeginRequest(reqID, uint16(roleResponder), 0)
	if err != nil {
		return nil, nil, err
	}

	err = c.writePairs(typeParams, reqID, env)
	if err != nil {
		return nil, nil, err
	}

	if len(requestData) > 0 {
		if err = c.writeRecord(typeStdin, reqID, []byte(requestData)); err != nil {
			return nil, nil, err
		}
	}

	rec := &record{}
	var err1 error

	// receive until EOF or FCGI_END_REQUEST
READ_LOOP:
	for {
		err1 = rec.read(c.rwc)
		if err1 != nil && strings.Contains(err1.Error(), "use of closed network connection") {
			if err1 != io.EOF {
				err = err1
			}
			break
		}

		switch {
		case rec.h.Type == typeStdout:
			retout = append(retout, rec.content()...)
		case rec.h.Type == typeStderr:
			reterr = append(reterr, rec.content()...)
		case rec.h.Type == typeEndRequest:
			fallthrough
		default:
			break READ_LOOP
		}
	}

	return retout, reterr, err
}
