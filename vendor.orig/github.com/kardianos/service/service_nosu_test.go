// Copyright 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// +build !su

package service_test

import "testing"

func TestInstallRunRestartStopRemove(t *testing.T) {
	t.Skip("skipping test as not running as root/admin (Build tag: su)")
}
