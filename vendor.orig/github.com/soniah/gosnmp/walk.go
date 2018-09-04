// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package gosnmp

import (
	"fmt"
	"strings"
)

func (x *GoSNMP) walk(getRequestType PDUType, rootOid string, walkFn WalkFunc) error {
	if rootOid == "" || rootOid == "." {
		rootOid = baseOid
	}

	if !strings.HasPrefix(rootOid, ".") {
		rootOid = string(".") + rootOid
	}

	oid := rootOid
	requests := 0
	maxReps := x.MaxRepetitions
	if maxReps == 0 {
		maxReps = defaultMaxRepetitions
	}

RequestLoop:
	for {

		requests++

		var response *SnmpPacket
		var err error

		switch getRequestType {
		case GetBulkRequest:
			response, err = x.GetBulk([]string{oid}, uint8(x.NonRepeaters), uint8(maxReps))
		case GetNextRequest:
			response, err = x.GetNext([]string{oid})
		case GetRequest:
			response, err = x.Get([]string{oid})
		default:
			response, err = nil, fmt.Errorf("Unsupported request type: %d", getRequestType)
		}

		if err != nil {
			return err
		}
		if len(response.Variables) == 0 {
			break RequestLoop
		}

		if response.Error == NoSuchName {
			x.Logger.Print("Walk terminated with NoSuchName")
			break RequestLoop
		}

		for k, v := range response.Variables {
			if v.Type == EndOfMibView || v.Type == NoSuchObject || v.Type == NoSuchInstance {
				x.Logger.Printf("BulkWalk terminated with type 0x%x", v.Type)
				break RequestLoop
			}
			if !strings.HasPrefix(v.Name, rootOid+".") {
				// Not in the requested root range.
				// if this is the first request, and the first variable in that request
				// and this condition is triggered - the first result is out of range
				// need to perform a regular get request
				// this request has been too narrowly defined to be found with a getNext
				// Issue #78 #93
				if requests == 1 && k == 0 {
					getRequestType = GetRequest
					continue RequestLoop
				}
				break RequestLoop
			}
			if v.Name == oid {
				return fmt.Errorf("OID not increasing: %s", v.Name)
			}
			// Report our pdu
			if err := walkFn(v); err != nil {
				return err
			}
		}
		// Save last oid for next request
		oid = response.Variables[len(response.Variables)-1].Name
	}
	x.Logger.Printf("BulkWalk completed in %d requests", requests)
	return nil
}

func (x *GoSNMP) walkAll(getRequestType PDUType, rootOid string) (results []SnmpPDU, err error) {
	err = x.walk(getRequestType, rootOid, func(dataUnit SnmpPDU) error {
		results = append(results, dataUnit)
		return nil
	})
	return results, err
}
