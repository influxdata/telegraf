// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"fmt"
	"reflect"
	"time"

	. "github.com/aerospike/aerospike-client-go/types"
	xrand "github.com/aerospike/aerospike-client-go/types/rand"
	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

const (
	_MAX_BUFFER_SIZE = 1024 * 1024 * 10 // 10 MB
	_CHUNK_SIZE      = 4096
)

type multiCommand interface {
	Stop()
}

type baseMultiCommand struct {
	baseCommand

	terminationError ResultCode

	recordset *Recordset

	terminationErrorType ResultCode

	errChan chan error

	resObjType     reflect.Type
	resObjMappings map[string]string
	selectCases    []reflect.SelectCase
}

var multiObjectParser func(
	cmd *baseMultiCommand,
	obj reflect.Value,
	opCount int,
	fieldCount int,
	generation uint32,
	expiration uint32,
) error

var prepareReflectionData func(cmd *baseMultiCommand)

func newMultiCommand(node *Node, recordset *Recordset) *baseMultiCommand {
	cmd := &baseMultiCommand{
		baseCommand: baseCommand{
			node:    node,
			oneShot: true,
		},
		recordset: recordset,
	}

	if prepareReflectionData != nil {
		prepareReflectionData(cmd)
	}
	return cmd
}

func (cmd *baseMultiCommand) getNode(ifc command) (*Node, error) {
	return cmd.node, nil
}

func (cmd *baseMultiCommand) getConnection(timeout time.Duration) (*Connection, error) {
	return cmd.node.getConnectionWithHint(timeout, byte(xrand.Int64()%256))
}

func (cmd *baseMultiCommand) putConnection(conn *Connection) {
	cmd.node.putConnectionWithHint(conn, byte(xrand.Int64()%256))
}

func (cmd *baseMultiCommand) drainConn(receiveSize int) error {
	// consume the rest of the input buffer from the socket
	if cmd.dataOffset < receiveSize && cmd.conn.IsConnected() {
		if err := cmd.readBytes(receiveSize - cmd.dataOffset); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *baseMultiCommand) parseResult(ifc command, conn *Connection) error {
	// Read socket into receive buffer one record at a time.  Do not read entire receive size
	// because the receive buffer would be too big.
	status := true

	var err error

	for status {
		// Read header.
		if _, err = cmd.conn.Read(cmd.dataBuffer, 8); err != nil {
			return err
		}

		size := Buffer.BytesToInt64(cmd.dataBuffer, 0)

		// Validate header to make sure we are at the beginning of a message
		if err := cmd.validateHeader(size); err != nil {
			return err
		}

		receiveSize := int(size & 0xFFFFFFFFFFFF)
		if receiveSize > 0 {
			status, err = ifc.parseRecordResults(ifc, receiveSize)
			cmd.drainConn(receiveSize)
			if err != nil {
				return err
			}
		} else {
			status = false
		}
	}

	return nil
}

func (cmd *baseMultiCommand) parseKey(fieldCount int) (*Key, error) {
	var digest [20]byte
	var namespace, setName string
	var userKey Value
	var err error

	for i := 0; i < fieldCount; i++ {
		if err = cmd.readBytes(4); err != nil {
			return nil, err
		}

		fieldlen := int(Buffer.BytesToUint32(cmd.dataBuffer, 0))
		if err = cmd.readBytes(fieldlen); err != nil {
			return nil, err
		}

		fieldtype := FieldType(cmd.dataBuffer[0])
		size := fieldlen - 1

		switch fieldtype {
		case DIGEST_RIPE:
			copy(digest[:], cmd.dataBuffer[1:size+1])
		case NAMESPACE:
			namespace = string(cmd.dataBuffer[1 : size+1])
		case TABLE:
			setName = string(cmd.dataBuffer[1 : size+1])
		case KEY:
			if userKey, err = bytesToKeyValue(int(cmd.dataBuffer[1]), cmd.dataBuffer, 2, size-1); err != nil {
				return nil, err
			}
		}
	}

	return &Key{namespace: namespace, setName: setName, digest: digest, userKey: userKey}, nil
}

func (cmd *baseMultiCommand) readBytes(length int) error {
	// Corrupted data streams can result in a huge length.
	// Do a sanity check here.
	if length > MaxBufferSize || length < 0 {
		return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid readBytes length: %d", length))
	}

	if length > cap(cmd.dataBuffer) {
		cmd.dataBuffer = make([]byte, length)
	}

	if n, err := cmd.conn.Read(cmd.dataBuffer[:length], length); err != nil {
		return fmt.Errorf("Requested to read %d bytes, but %d was read. (%v)", length, n, err)
	}

	cmd.dataOffset += length
	return nil
}

func (cmd *baseMultiCommand) parseRecordResults(ifc command, receiveSize int) (bool, error) {
	// Read/parse remaining message bytes one record at a time.
	cmd.dataOffset = 0

	for cmd.dataOffset < receiveSize {
		if err := cmd.readBytes(int(_MSG_REMAINING_HEADER_SIZE)); err != nil {
			err = newNodeError(cmd.node, err)
			return false, err
		}
		resultCode := ResultCode(cmd.dataBuffer[5] & 0xFF)

		if resultCode != 0 {
			if resultCode == KEY_NOT_FOUND_ERROR {
				return false, nil
			}
			err := NewAerospikeError(resultCode)
			err = newNodeError(cmd.node, err)
			return false, err
		}

		info3 := int(cmd.dataBuffer[3])

		// If cmd is the end marker of the response, do not proceed further
		if (info3 & _INFO3_LAST) == _INFO3_LAST {
			return false, nil
		}

		generation := Buffer.BytesToUint32(cmd.dataBuffer, 6)
		expiration := TTL(Buffer.BytesToUint32(cmd.dataBuffer, 10))
		fieldCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 18))
		opCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 20))

		key, err := cmd.parseKey(fieldCount)
		if err != nil {
			err = newNodeError(cmd.node, err)
			return false, err
		}

		// if there is a recordset, process the record traditionally
		// otherwise, it is supposed to be a record channel
		if cmd.selectCases == nil {
			// Parse bins.
			var bins BinMap

			for i := 0; i < opCount; i++ {
				if err := cmd.readBytes(8); err != nil {
					err = newNodeError(cmd.node, err)
					return false, err
				}

				opSize := int(Buffer.BytesToUint32(cmd.dataBuffer, 0))
				particleType := int(cmd.dataBuffer[5])
				nameSize := int(cmd.dataBuffer[7])

				if err := cmd.readBytes(nameSize); err != nil {
					err = newNodeError(cmd.node, err)
					return false, err
				}
				name := string(cmd.dataBuffer[:nameSize])

				particleBytesSize := int((opSize - (4 + nameSize)))
				if err = cmd.readBytes(particleBytesSize); err != nil {
					err = newNodeError(cmd.node, err)
					return false, err
				}
				value, err := bytesToParticle(particleType, cmd.dataBuffer, 0, particleBytesSize)
				if err != nil {
					err = newNodeError(cmd.node, err)
					return false, err
				}

				if bins == nil {
					bins = make(BinMap, opCount)
				}
				bins[name] = value
			}

			// If the channel is full and it blocks, we don't want this command to
			// block forever, or panic in case the channel is closed in the meantime.
			select {
			// send back the result on the async channel
			case cmd.recordset.Records <- newRecord(cmd.node, key, bins, generation, expiration):
			case <-cmd.recordset.cancelled:
				return false, NewAerospikeError(cmd.terminationErrorType)
			}
		} else if multiObjectParser != nil {
			obj := reflect.New(cmd.resObjType)
			if err := multiObjectParser(cmd, obj, opCount, fieldCount, generation, expiration); err != nil {
				err = newNodeError(cmd.node, err)
				return false, err
			}

			// set the object to send
			cmd.selectCases[0].Send = obj

			chosen, _, _ := reflect.Select(cmd.selectCases)
			switch chosen {
			case 0: // object sent
			case 1: // cancel channel is closed
				return false, NewAerospikeError(cmd.terminationErrorType)
			}
		}
	}

	return true, nil
}
