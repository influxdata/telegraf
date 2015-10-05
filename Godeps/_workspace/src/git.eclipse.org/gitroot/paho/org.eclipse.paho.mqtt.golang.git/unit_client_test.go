/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"log"
	"net/http"
	"os"
	"testing"

	_ "net/http/pprof"
)

func init() {
	DEBUG = log.New(os.Stderr, "DEBUG    ", log.Ltime)
	WARN = log.New(os.Stderr, "WARNING  ", log.Ltime)
	CRITICAL = log.New(os.Stderr, "CRITICAL ", log.Ltime)
	ERROR = log.New(os.Stderr, "ERROR    ", log.Ltime)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func Test_NewClient_simple(t *testing.T) {
	ops := NewClientOptions().SetClientID("foo").AddBroker("tcp://10.10.0.1:1883")
	c := NewClient(ops)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	if c.options.ClientID != "foo" {
		t.Fatalf("bad client id")
	}

	if c.options.Servers[0].Scheme != "tcp" {
		t.Fatalf("bad server scheme")
	}

	if c.options.Servers[0].Host != "10.10.0.1:1883" {
		t.Fatalf("bad server host")
	}
}
