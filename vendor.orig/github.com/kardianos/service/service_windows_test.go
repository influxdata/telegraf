// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package service

import (
	"testing"
)

func TestTimeout(t *testing.T) {
	stopSpan := getStopTimeout()
	t.Log("Max Stop Duration", stopSpan)
}
