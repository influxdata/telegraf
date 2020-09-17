// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uacp

import (
	"net"
	"strings"

	"github.com/gopcua/opcua/errors"
)

// ResolveEndpoint returns network type, address, and error splitted from EndpointURL.
//
// Expected format of input is "opc.tcp://<addr[:port]/path/to/somewhere"
func ResolveEndpoint(endpoint string) (network string, addr *net.TCPAddr, err error) {
	elems := strings.Split(endpoint, "/")
	if elems[0] != "opc.tcp:" {
		return "", nil, errors.Errorf("invalid endpoint %s", endpoint)
	}

	addrString := elems[2]
	if !strings.Contains(addrString, ":") {
		addrString += ":4840"
	}

	network = "tcp"
	addr, err = net.ResolveTCPAddr(network, addrString)
	switch err.(type) {
	case *net.DNSError:
		return "", nil, errors.Errorf("could not resolve address %s", addrString)
	}
	return
}
