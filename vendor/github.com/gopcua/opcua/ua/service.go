// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"fmt"
	"strconv"

	"github.com/gopcua/opcua/debug"
)

// svcreg contains all known service request/response objects.
var svcreg = NewTypeRegistry()

// RegisterService registers a new service object type.
// It panics if the type or the id is already registered.
func RegisterService(typeID uint16, v interface{}) {
	id := strconv.Itoa(int(typeID))
	if err := svcreg.Register(id, v); err != nil {
		panic("Service " + err.Error())
	}
}

// ServiceTypeID returns the id of the service object type as
// registered with RegisterService. If the service object is not
// known the function returns 0.
func ServiceTypeID(v interface{}) uint16 {
	n, _ := strconv.Atoi(svcreg.Lookup(v))
	return uint16(n)
}

func DecodeService(b []byte) (*ExpandedNodeID, interface{}, error) {
	typeID := new(ExpandedNodeID)
	n, err := typeID.Decode(b)
	if err != nil {
		return nil, nil, err
	}
	b = b[n:]

	v := svcreg.New(strconv.Itoa(int(typeID.NodeID.IntID())))
	if v == nil {
		return nil, nil, StatusBadServiceUnsupported
	}

	if debug.FlagSet("packet") {
		fmt.Printf("%T: %#v\n", v, b)
	}

	_, err = Decode(b, v)
	return typeID, v, err
}
