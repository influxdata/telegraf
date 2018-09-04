// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// The purpose of these tests is to validate gosnmp's public APIs.
//
// IMPORTANT: If you're modifying _any_ existing code in this file, you
// should be asking yourself about API compatibility!

// +build all api

package gosnmp_test // force external view

import (
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"

	"github.com/soniah/gosnmp"
)

func TestAPIConfigTypes(t *testing.T) {
	g := &gosnmp.GoSNMP{}
	g.Target = ""
	g.Port = 0
	g.Community = ""
	g.Version = gosnmp.Version1
	g.Version = gosnmp.Version2c
	g.Timeout = time.Duration(0)
	g.Retries = 0
	g.Logger = log.New(ioutil.Discard, "", 0)
	g.MaxOids = 0
	g.MaxRepetitions = 0
	g.NonRepeaters = 0

	var c net.Conn
	c = g.Conn
	_ = c
}

func TestAPIDefault(t *testing.T) {
	var g *gosnmp.GoSNMP
	g = gosnmp.Default
	_ = g
}

func TestAPIConnectMethodSignature(t *testing.T) {
	var f func() error
	f = gosnmp.Default.Connect
	_ = f
}

func TestAPIGetMethodSignature(t *testing.T) {
	var f func([]string) (*gosnmp.SnmpPacket, error)
	f = gosnmp.Default.Get
	_ = f
}

func TestAPISetMethodSignature(t *testing.T) {
	var f func([]gosnmp.SnmpPDU) (*gosnmp.SnmpPacket, error)
	f = gosnmp.Default.Set
	_ = f
}

func TestAPIGetNextMethodSignature(t *testing.T) {
	var f func([]string) (*gosnmp.SnmpPacket, error)
	f = gosnmp.Default.GetNext
	_ = f
}

func TestAPIBulkWalkMethodSignature(t *testing.T) {
	var f func(string, gosnmp.WalkFunc) error
	f = gosnmp.Default.BulkWalk
	_ = f
}

func TestAPIBulkWalkAllMethodSignature(t *testing.T) {
	var f func(string) ([]gosnmp.SnmpPDU, error)
	f = gosnmp.Default.BulkWalkAll
	_ = f
}

func TestAPIWalkMethodSignature(t *testing.T) {
	var f func(string, gosnmp.WalkFunc) error
	f = gosnmp.Default.Walk
	_ = f
}

func TestAPIWalkAllMethodSignature(t *testing.T) {
	var f func(string) ([]gosnmp.SnmpPDU, error)
	f = gosnmp.Default.WalkAll
	_ = f
}

func TestAPIWalkFuncSignature(t *testing.T) {
	var f gosnmp.WalkFunc
	f = func(du gosnmp.SnmpPDU) (err error) { return }
	_ = f
}
