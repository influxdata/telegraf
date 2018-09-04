// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build gofuzz

package batch

func Fuzz(data []byte) int {
	Split(string(data), "GO")
	return 0
}
