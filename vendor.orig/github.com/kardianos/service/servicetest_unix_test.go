// Copyright 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package service_test

import (
	"os"
	"testing"
)

func interruptProcess(t *testing.T) {
	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("FindProcess: %s", err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("Signal: %s", err)
	}
}
