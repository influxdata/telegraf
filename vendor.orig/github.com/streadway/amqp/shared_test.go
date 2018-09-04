// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"encoding/hex"
	"io"
	"testing"
)

type pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (p pipe) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func (p pipe) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

func (p pipe) Close() error {
	p.r.Close()
	p.w.Close()
	return nil
}

type logIO struct {
	t      *testing.T
	prefix string
	proxy  io.ReadWriteCloser
}

func (log *logIO) Read(p []byte) (n int, err error) {
	log.t.Logf("%s reading %d\n", log.prefix, len(p))
	n, err = log.proxy.Read(p)
	if err != nil {
		log.t.Logf("%s read %x: %v\n", log.prefix, p[0:n], err)
	} else {
		log.t.Logf("%s read:\n%s\n", log.prefix, hex.Dump(p[0:n]))
		//fmt.Printf("%s read:\n%s\n", log.prefix, hex.Dump(p[0:n]))
	}
	return
}

func (log *logIO) Write(p []byte) (n int, err error) {
	log.t.Logf("%s writing %d\n", log.prefix, len(p))
	n, err = log.proxy.Write(p)
	if err != nil {
		log.t.Logf("%s write %d, %x: %v\n", log.prefix, len(p), p[0:n], err)
	} else {
		log.t.Logf("%s write %d:\n%s", log.prefix, len(p), hex.Dump(p[0:n]))
		//fmt.Printf("%s write %d:\n%s", log.prefix, len(p), hex.Dump(p[0:n]))
	}
	return
}

func (log *logIO) Close() (err error) {
	err = log.proxy.Close()
	if err != nil {
		log.t.Logf("%s close : %v\n", log.prefix, err)
	} else {
		log.t.Logf("%s close\n", log.prefix)
	}
	return
}
