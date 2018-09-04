// Copyright 2012-2016 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

/*
The developer of the trapserver code (https://github.com/jda) says "I'm working
on the best level of abstraction but I'm able to receive traps from a Cisco
switch and Net-SNMP".

Pull requests welcome.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	g "github.com/soniah/gosnmp"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage:\n")
		fmt.Printf("   %s\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	tl := g.NewTrapListener()
	tl.OnNewTrap = myTrapHandler
	tl.Params = g.Default
	tl.Params.Logger = log.New(os.Stdout, "", 0)

	err := tl.Listen("0.0.0.0:9162")
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}

func myTrapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	log.Printf("got trapdata from %s\n", addr.IP)
	for _, v := range packet.Variables {
		switch v.Type {
		case g.OctetString:
			b := v.Value.([]byte)
			fmt.Printf("OID: %s, string: %x\n", v.Name, b)

		default:
			log.Printf("trap: %+v\n", v)
		}
	}
}
