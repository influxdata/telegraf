// Copyright (c) 2016 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package proto

// To run this command you need protobuf and the go protoc plugin:
// brew install protobuf --devel
// go get -u github.com/golang/protobuf/protoc-gen-go

//go:generate protoc --go_out=. lanz.proto
