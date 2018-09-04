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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"

	ParticleType "github.com/aerospike/aerospike-client-go/types/particle_type"
	// Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

const (
	// Flags commented out are not supported by cmd client.
	// Contains a read operation.
	_INFO1_READ int = (1 << 0)
	// Get all bins.
	_INFO1_GET_ALL int = (1 << 1)

	// Do not read the bins
	_INFO1_NOBINDATA int = (1 << 5)

	// Involve all replicas in read operation.
	_INFO1_CONSISTENCY_ALL = (1 << 6)

	// Create or update record
	_INFO2_WRITE int = (1 << 0)
	// Fling a record into the belly of Moloch.
	_INFO2_DELETE int = (1 << 1)
	// Update if expected generation == old.
	_INFO2_GENERATION int = (1 << 2)
	// Update if new generation >= old, good for restore.
	_INFO2_GENERATION_GT int = (1 << 3)
	// Transaction resulting in record deletion leaves tombstone (Enterprise only).
	_INFO2_DURABLE_DELETE int = (1 << 4)
	// Create only. Fail if record already exists.
	_INFO2_CREATE_ONLY int = (1 << 5)

	// Return a result for every operation.
	_INFO2_RESPOND_ALL_OPS int = (1 << 7)

	// This is the last of a multi-part message.
	_INFO3_LAST int = (1 << 0)
	// Commit to master only before declaring success.
	_INFO3_COMMIT_MASTER int = (1 << 1)
	// Update only. Merge bins.
	_INFO3_UPDATE_ONLY int = (1 << 3)

	// Create or completely replace record.
	_INFO3_CREATE_OR_REPLACE int = (1 << 4)
	// Completely replace existing record only.
	_INFO3_REPLACE_ONLY int = (1 << 5)

	_MSG_TOTAL_HEADER_SIZE     uint8 = 30
	_FIELD_HEADER_SIZE         uint8 = 5
	_OPERATION_HEADER_SIZE     uint8 = 8
	_MSG_REMAINING_HEADER_SIZE uint8 = 22
	_DIGEST_SIZE               uint8 = 20
	_CL_MSG_VERSION            int64 = 2
	_AS_MSG_TYPE               int64 = 3
)

// command interface describes all commands available
type command interface {
	getPolicy(ifc command) Policy

	writeBuffer(ifc command) error
	getNode(ifc command) (*Node, error)
	getConnection(timeout time.Duration) (*Connection, error)
	putConnection(conn *Connection)
	parseResult(ifc command, conn *Connection) error
	parseRecordResults(ifc command, receiveSize int) (bool, error)

	execute(ifc command) error
	// Executes the command
	Execute() error
}

// Holds data buffer for the command
type baseCommand struct {
	node *Node
	conn *Connection

	dataBuffer []byte
	dataOffset int

	// oneShot determines if streaming commands like query, scan or queryAggregate
	// are not retried if they error out mid-parsing
	oneShot bool
}

// Writes the command for write operations
func (cmd *baseCommand) setWrite(policy *WritePolicy, operation OperationType, key *Key, bins []*Bin, binMap BinMap) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, policy.SendKey)
	if err != nil {
		return err
	}

	if binMap == nil {
		for i := range bins {
			if err := cmd.estimateOperationSizeForBin(bins[i]); err != nil {
				return err
			}
		}
	} else {
		for name, value := range binMap {
			if err := cmd.estimateOperationSizeForBinNameAndValue(name, value); err != nil {
				return err
			}
		}
	}

	if err := cmd.sizeBuffer(); err != nil {
		return err
	}

	if binMap == nil {
		cmd.writeHeaderWithPolicy(policy, 0, _INFO2_WRITE, fieldCount, len(bins))
	} else {
		cmd.writeHeaderWithPolicy(policy, 0, _INFO2_WRITE, fieldCount, len(binMap))
	}

	cmd.writeKey(key, policy.SendKey)

	if binMap == nil {
		for i := range bins {
			if err := cmd.writeOperationForBin(bins[i], operation); err != nil {
				return err
			}
		}
	} else {
		for name, value := range binMap {
			if err := cmd.writeOperationForBinNameAndValue(name, value, operation); err != nil {
				return err
			}
		}
	}

	cmd.end()

	return nil
}

// Writes the command for delete operations
func (cmd *baseCommand) setDelete(policy *WritePolicy, key *Key) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, false)
	if err != nil {
		return err
	}
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}
	cmd.writeHeaderWithPolicy(policy, 0, _INFO2_WRITE|_INFO2_DELETE, fieldCount, 0)
	cmd.writeKey(key, false)
	cmd.end()
	return nil

}

// Writes the command for touch operations
func (cmd *baseCommand) setTouch(policy *WritePolicy, key *Key) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, policy.SendKey)
	if err != nil {
		return err
	}

	cmd.estimateOperationSize()
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}
	cmd.writeHeaderWithPolicy(policy, 0, _INFO2_WRITE, fieldCount, 1)
	cmd.writeKey(key, policy.SendKey)
	cmd.writeOperationForOperationType(TOUCH)
	cmd.end()
	return nil

}

// Writes the command for exist operations
func (cmd *baseCommand) setExists(policy *BasePolicy, key *Key) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, false)
	if err != nil {
		return err
	}
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}
	cmd.writeHeader(policy, _INFO1_READ|_INFO1_NOBINDATA, 0, fieldCount, 0)
	cmd.writeKey(key, false)
	cmd.end()
	return nil

}

// Writes the command for get operations (all bins)
func (cmd *baseCommand) setReadForKeyOnly(policy *BasePolicy, key *Key) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, false)
	if err != nil {
		return err
	}
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}
	cmd.writeHeader(policy, _INFO1_READ|_INFO1_GET_ALL, 0, fieldCount, 0)
	cmd.writeKey(key, false)
	cmd.end()
	return nil

}

// Writes the command for get operations (specified bins)
func (cmd *baseCommand) setRead(policy *BasePolicy, key *Key, binNames []string) (err error) {
	if len(binNames) > 0 {
		cmd.begin()
		fieldCount, err := cmd.estimateKeySize(key, false)
		if err != nil {
			return err
		}

		for i := range binNames {
			cmd.estimateOperationSizeForBinName(binNames[i])
		}
		if err = cmd.sizeBuffer(); err != nil {
			return nil
		}
		cmd.writeHeader(policy, _INFO1_READ, 0, fieldCount, len(binNames))
		cmd.writeKey(key, false)

		for i := range binNames {
			cmd.writeOperationForBinName(binNames[i], READ)
		}
		cmd.end()
	} else {
		err = cmd.setReadForKeyOnly(policy, key)
	}

	return err
}

// Writes the command for getting metadata operations
func (cmd *baseCommand) setReadHeader(policy *BasePolicy, key *Key) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, false)
	if err != nil {
		return err
	}
	cmd.estimateOperationSizeForBinName("")
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}

	cmd.writeHeader(policy, _INFO1_READ|_INFO1_NOBINDATA, 0, fieldCount, 1)

	cmd.writeKey(key, false)
	cmd.writeOperationForBinName("", READ)
	cmd.end()
	return nil

}

// Implements different command operations
func (cmd *baseCommand) setOperate(policy *WritePolicy, key *Key, operations []*Operation) error {
	if len(operations) == 0 {
		return NewAerospikeError(PARAMETER_ERROR, "No operations were passed.")
	}

	cmd.begin()
	fieldCount := 0
	readAttr := 0
	writeAttr := 0
	readBin := false
	readHeader := false
	RespondPerEachOp := policy.RespondPerEachOp

	for i := range operations {
		switch operations[i].opType {
		case MAP_READ:
			// Map operations require RespondPerEachOp to be true.
			RespondPerEachOp = true
			// Fall through to read.
			fallthrough
		case READ, CDT_READ:
			if !operations[i].headerOnly {
				readAttr |= _INFO1_READ

				// Read all bins if no bin is specified.
				if operations[i].binName == "" {
					readAttr |= _INFO1_GET_ALL
				}
				readBin = true
			} else {
				readAttr |= _INFO1_READ
				readHeader = true
			}
		case MAP_MODIFY:
			// Map operations require RespondPerEachOp to be true.
			RespondPerEachOp = true
			// Fall through to default.
			fallthrough
		default:
			writeAttr = _INFO2_WRITE
		}
		cmd.estimateOperationSizeForOperation(operations[i])
	}

	ksz, err := cmd.estimateKeySize(key, policy.SendKey && writeAttr != 0)
	if err != nil {
		return err
	}
	fieldCount += ksz

	if err := cmd.sizeBuffer(); err != nil {
		return err
	}

	if readHeader && !readBin {
		readAttr |= _INFO1_NOBINDATA
	}

	if RespondPerEachOp {
		writeAttr |= _INFO2_RESPOND_ALL_OPS
	}

	if writeAttr != 0 {
		cmd.writeHeaderWithPolicy(policy, readAttr, writeAttr, fieldCount, len(operations))
	} else {
		cmd.writeHeader(&policy.BasePolicy, readAttr, writeAttr, fieldCount, len(operations))
	}
	cmd.writeKey(key, policy.SendKey && writeAttr != 0)

	for _, operation := range operations {
		if err := cmd.writeOperationForOperation(operation); err != nil {
			return err
		}
	}

	cmd.end()

	return nil
}

func (cmd *baseCommand) setUdf(policy *WritePolicy, key *Key, packageName string, functionName string, args *ValueArray) error {
	cmd.begin()
	fieldCount, err := cmd.estimateKeySize(key, policy.SendKey)
	if err != nil {
		return err
	}

	fc, err := cmd.estimateUdfSize(packageName, functionName, args)
	if err != nil {
		return err
	}
	fieldCount += fc

	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}

	cmd.writeHeaderWithPolicy(policy, 0, _INFO2_WRITE, fieldCount, 0)
	cmd.writeKey(key, policy.SendKey)
	cmd.writeFieldString(packageName, UDF_PACKAGE_NAME)
	cmd.writeFieldString(functionName, UDF_FUNCTION)
	cmd.writeUdfArgs(args)
	cmd.end()

	return nil
}

func (cmd *baseCommand) setBatchExists(policy *BasePolicy, keys []*Key, batch *batchNamespace) error {
	// Estimate buffer size
	cmd.begin()
	byteSize := batch.offsetSize * int(_DIGEST_SIZE)

	cmd.dataOffset += len(*batch.namespace) +
		int(_FIELD_HEADER_SIZE) + byteSize + int(_FIELD_HEADER_SIZE)
	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}

	cmd.writeHeader(policy, _INFO1_READ|_INFO1_NOBINDATA, 0, 2, 0)
	cmd.writeFieldString(*batch.namespace, NAMESPACE)
	cmd.writeFieldHeader(byteSize, DIGEST_RIPE_ARRAY)

	offsets := batch.offsets
	max := batch.offsetSize

	for i := 0; i < max; i++ {
		key := keys[offsets[i]]
		copy(cmd.dataBuffer[cmd.dataOffset:], key.digest[:])
		cmd.dataOffset += len(key.digest)
	}
	cmd.end()

	return nil
}

func (cmd *baseCommand) setBatchGet(policy *BasePolicy, keys []*Key, batch *batchNamespace, binNames map[string]struct{}, readAttr int) error {
	// Estimate buffer size
	cmd.begin()
	byteSize := batch.offsetSize * int(_DIGEST_SIZE)

	cmd.dataOffset += len(*batch.namespace) +
		int(_FIELD_HEADER_SIZE) + byteSize + int(_FIELD_HEADER_SIZE)

	for binName := range binNames {
		cmd.estimateOperationSizeForBinName(binName)
	}

	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}

	operationCount := len(binNames)
	cmd.writeHeader(policy, readAttr, 0, 2, operationCount)
	cmd.writeFieldString(*batch.namespace, NAMESPACE)
	cmd.writeFieldHeader(byteSize, DIGEST_RIPE_ARRAY)

	offsets := batch.offsets
	max := batch.offsetSize

	for i := 0; i < max; i++ {
		key := keys[offsets[i]]
		copy(cmd.dataBuffer[cmd.dataOffset:], key.digest[:])
		cmd.dataOffset += len(key.digest)
	}

	for binName := range binNames {
		cmd.writeOperationForBinName(binName, READ)
	}
	cmd.end()

	return nil
}

func (cmd *baseCommand) setScan(policy *ScanPolicy, namespace *string, setName *string, binNames []string, taskId uint64) error {
	cmd.begin()
	fieldCount := 0
	// predExpsSize := 0

	if namespace != nil {
		cmd.dataOffset += len(*namespace) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	if setName != nil {
		cmd.dataOffset += len(*setName) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	// Estimate scan options size.
	cmd.dataOffset += 2 + int(_FIELD_HEADER_SIZE)
	fieldCount++

	// Estimate scan timeout size.
	cmd.dataOffset += 4 + int(_FIELD_HEADER_SIZE)
	fieldCount++

	// Allocate space for TaskId field.
	cmd.dataOffset += 8 + int(_FIELD_HEADER_SIZE)
	fieldCount++

	if binNames != nil {
		for i := range binNames {
			cmd.estimateOperationSizeForBinName(binNames[i])
		}
	}

	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}
	readAttr := _INFO1_READ

	if !policy.IncludeBinData {
		readAttr |= _INFO1_NOBINDATA
	}

	operationCount := 0
	if binNames != nil {
		operationCount = len(binNames)
	}
	cmd.writeHeader(policy.BasePolicy, readAttr, 0, fieldCount, operationCount)

	if namespace != nil {
		cmd.writeFieldString(*namespace, NAMESPACE)
	}

	if setName != nil {
		cmd.writeFieldString(*setName, TABLE)
	}

	cmd.writeFieldHeader(2, SCAN_OPTIONS)
	priority := byte(policy.Priority)
	priority <<= 4

	if policy.FailOnClusterChange {
		priority |= 0x08
	}

	if policy.IncludeLDT {
		priority |= 0x02
	}

	cmd.WriteByte(priority)
	cmd.WriteByte(byte(policy.ScanPercent))

	// Write scan timeout
	cmd.writeFieldHeader(4, SCAN_TIMEOUT)
	cmd.WriteInt32(int32(policy.ServerSocketTimeout / time.Millisecond)) // in milliseconds

	cmd.writeFieldHeader(8, TRAN_ID)
	cmd.WriteUint64(taskId)

	if binNames != nil {
		for i := range binNames {
			cmd.writeOperationForBinName(binNames[i], READ)
		}
	}

	cmd.end()

	return nil
}

func (cmd *baseCommand) setQuery(policy *QueryPolicy, statement *Statement, write bool) (err error) {
	fieldCount := 0
	filterSize := 0
	binNameSize := 0
	predExpsSize := 0

	cmd.begin()

	if statement.Namespace != "" {
		cmd.dataOffset += len(statement.Namespace) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	if statement.IndexName != "" {
		cmd.dataOffset += len(statement.IndexName) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	if statement.SetName != "" {
		cmd.dataOffset += len(statement.SetName) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	// Allocate space for TaskId field.
	cmd.dataOffset += 8 + int(_FIELD_HEADER_SIZE)
	fieldCount++

	if len(statement.Filters) > 0 {
		if len(statement.Filters) > 1 {
			return NewAerospikeError(PARAMETER_ERROR, "Aerospike server currently supports only one filter.")
		} else if len(statement.Filters) == 1 {
			idxType := statement.Filters[0].IndexCollectionType()

			if idxType != ICT_DEFAULT {
				cmd.dataOffset += int(_FIELD_HEADER_SIZE) + 1
				fieldCount++
			}
		}

		cmd.dataOffset += int(_FIELD_HEADER_SIZE)
		filterSize++ // num filters

		for _, filter := range statement.Filters {
			sz, err := filter.estimateSize()
			if err != nil {
				return err
			}
			filterSize += sz
		}
		cmd.dataOffset += filterSize
		fieldCount++

		// Query bin names are specified as a field (Scan bin names are specified later as operations)
		if len(statement.BinNames) > 0 {
			cmd.dataOffset += int(_FIELD_HEADER_SIZE)
			binNameSize++ // num bin names

			for _, binName := range statement.BinNames {
				binNameSize += len(binName) + 1
			}
			cmd.dataOffset += binNameSize
			fieldCount++
		}
	} else {
		// Calling query with no filters is more efficiently handled by a primary index scan.
		// Estimate scan options size.
		cmd.dataOffset += (2 + int(_FIELD_HEADER_SIZE))
		fieldCount++
	}

	if len(statement.predExps) > 0 {
		cmd.dataOffset += int(_FIELD_HEADER_SIZE)
		for _, predexp := range statement.predExps {
			predExpsSize += predexp.marshaledSize()
		}
		cmd.dataOffset += predExpsSize
		fieldCount++
	}

	var functionArgs *ValueArray
	if statement.functionName != "" {
		cmd.dataOffset += int(_FIELD_HEADER_SIZE) + 1 // udf type
		cmd.dataOffset += len(statement.packageName) + int(_FIELD_HEADER_SIZE)
		cmd.dataOffset += len(statement.functionName) + int(_FIELD_HEADER_SIZE)

		fasz := 0
		if len(statement.functionArgs) > 0 {
			functionArgs = NewValueArray(statement.functionArgs)
			fasz, err = functionArgs.estimateSize()
			if err != nil {
				return err
			}
		}

		cmd.dataOffset += int(_FIELD_HEADER_SIZE) + fasz
		fieldCount += 4
	}

	if len(statement.Filters) == 0 {
		if len(statement.BinNames) > 0 {
			for _, binName := range statement.BinNames {
				cmd.estimateOperationSizeForBinName(binName)
			}
		}
	}

	if err := cmd.sizeBuffer(); err != nil {
		return nil
	}

	operationCount := 0
	if len(statement.Filters) == 0 && len(statement.BinNames) > 0 {
		operationCount = len(statement.BinNames)
	}

	if write {
		cmd.writeHeader(policy.BasePolicy, _INFO1_READ, _INFO2_WRITE, fieldCount, operationCount)
	} else {
		cmd.writeHeader(policy.BasePolicy, _INFO1_READ, 0, fieldCount, operationCount)
	}

	if statement.Namespace != "" {
		cmd.writeFieldString(statement.Namespace, NAMESPACE)
	}

	if statement.IndexName != "" {
		cmd.writeFieldString(statement.IndexName, INDEX_NAME)
	}

	if statement.SetName != "" {
		cmd.writeFieldString(statement.SetName, TABLE)
	}

	cmd.writeFieldHeader(8, TRAN_ID)
	cmd.WriteUint64(statement.TaskId)

	if len(statement.Filters) > 0 {
		if len(statement.Filters) >= 1 {
			idxType := statement.Filters[0].IndexCollectionType()

			if idxType != ICT_DEFAULT {
				cmd.writeFieldHeader(1, INDEX_TYPE)
				cmd.WriteByte(byte(idxType))
			}
		}

		cmd.writeFieldHeader(filterSize, INDEX_RANGE)
		cmd.WriteByte(byte(len(statement.Filters)))

		for _, filter := range statement.Filters {
			_, err := filter.write(cmd)
			if err != nil {
				return err
			}
		}

		if len(statement.BinNames) > 0 {
			cmd.writeFieldHeader(binNameSize, QUERY_BINLIST)
			cmd.WriteByte(byte(len(statement.BinNames)))

			for _, binName := range statement.BinNames {
				len := copy(cmd.dataBuffer[cmd.dataOffset+1:], binName)
				cmd.dataBuffer[cmd.dataOffset] = byte(len)
				cmd.dataOffset += len + 1
			}
		}
	} else {
		// Calling query with no filters is more efficiently handled by a primary index scan.
		cmd.writeFieldHeader(2, SCAN_OPTIONS)
		priority := byte(policy.Priority)
		priority <<= 4
		cmd.WriteByte(priority)
		cmd.WriteByte(byte(100))
	}

	if len(statement.predExps) > 0 {
		cmd.writeFieldHeader(predExpsSize, PREDEXP)
		for _, predexp := range statement.predExps {
			if err := predexp.marshal(cmd); err != nil {
				return err
			}
		}
	}

	if statement.functionName != "" {
		cmd.writeFieldHeader(1, UDF_OP)
		if statement.returnData {
			cmd.dataBuffer[cmd.dataOffset] = byte(1)
		} else {
			cmd.dataBuffer[cmd.dataOffset] = byte(2)
		}
		cmd.dataOffset++

		cmd.writeFieldString(statement.packageName, UDF_PACKAGE_NAME)
		cmd.writeFieldString(statement.functionName, UDF_FUNCTION)
		cmd.writeUdfArgs(functionArgs)
	}

	// scan binNames come last
	if len(statement.Filters) == 0 {
		if len(statement.BinNames) > 0 {
			for _, binName := range statement.BinNames {
				cmd.writeOperationForBinName(binName, READ)
			}
		}
	}

	cmd.end()

	return nil
}

func (cmd *baseCommand) estimateKeySize(key *Key, sendKey bool) (int, error) {
	fieldCount := 0

	if key.namespace != "" {
		cmd.dataOffset += len(key.namespace) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	if key.setName != "" {
		cmd.dataOffset += len(key.setName) + int(_FIELD_HEADER_SIZE)
		fieldCount++
	}

	cmd.dataOffset += int(_DIGEST_SIZE + _FIELD_HEADER_SIZE)
	fieldCount++

	if sendKey {
		// field header size + key size
		sz, err := key.userKey.estimateSize()
		if err != nil {
			return sz, err
		}
		cmd.dataOffset += sz + int(_FIELD_HEADER_SIZE) + 1
		fieldCount++
	}

	return fieldCount, nil
}

func (cmd *baseCommand) estimateUdfSize(packageName string, functionName string, args *ValueArray) (int, error) {
	cmd.dataOffset += len(packageName) + int(_FIELD_HEADER_SIZE)
	cmd.dataOffset += len(functionName) + int(_FIELD_HEADER_SIZE)

	sz, err := args.estimateSize()
	if err != nil {
		return 0, err
	}

	// fmt.Println(args, sz)

	cmd.dataOffset += sz + int(_FIELD_HEADER_SIZE)
	return 3, nil
}

func (cmd *baseCommand) estimateOperationSizeForBin(bin *Bin) error {
	cmd.dataOffset += len(bin.Name) + int(_OPERATION_HEADER_SIZE)
	sz, err := bin.Value.estimateSize()
	if err != nil {
		return err
	}
	cmd.dataOffset += sz
	return nil
}

func (cmd *baseCommand) estimateOperationSizeForBinNameAndValue(name string, value interface{}) error {
	cmd.dataOffset += len(name) + int(_OPERATION_HEADER_SIZE)
	sz, err := NewValue(value).estimateSize()
	if err != nil {
		return err
	}
	cmd.dataOffset += sz
	return nil
}

func (cmd *baseCommand) estimateOperationSizeForOperation(operation *Operation) error {
	binLen := len(operation.binName)
	cmd.dataOffset += binLen + int(_OPERATION_HEADER_SIZE)

	if operation.encoder == nil {
		if operation.binValue != nil {
			sz, err := operation.binValue.estimateSize()
			if err != nil {
				return err
			}
			cmd.dataOffset += sz
		}
	} else {
		sz, err := operation.encoder(operation, nil)
		if err != nil {
			return err
		}
		cmd.dataOffset += sz
	}
	return nil
}

func (cmd *baseCommand) estimateOperationSizeForBinName(binName string) {
	cmd.dataOffset += len(binName) + int(_OPERATION_HEADER_SIZE)
}

func (cmd *baseCommand) estimateOperationSize() {
	cmd.dataOffset += int(_OPERATION_HEADER_SIZE)
}

// Generic header write.
func (cmd *baseCommand) writeHeader(policy *BasePolicy, readAttr int, writeAttr int, fieldCount int, operationCount int) {

	if policy.ConsistencyLevel == CONSISTENCY_ALL {
		readAttr |= _INFO1_CONSISTENCY_ALL
	}

	// Write all header data except total size which must be written last.
	cmd.dataBuffer[8] = _MSG_REMAINING_HEADER_SIZE // Message header length.
	cmd.dataBuffer[9] = byte(readAttr)
	cmd.dataBuffer[10] = byte(writeAttr)

	for i := 11; i < 26; i++ {
		cmd.dataBuffer[i] = 0
	}
	cmd.dataOffset = 26
	cmd.WriteInt16(int16(fieldCount))
	cmd.WriteInt16(int16(operationCount))
	cmd.dataOffset = int(_MSG_TOTAL_HEADER_SIZE)
}

// Header write for write operations.
func (cmd *baseCommand) writeHeaderWithPolicy(policy *WritePolicy, readAttr int, writeAttr int, fieldCount int, operationCount int) {
	// Set flags.
	generation := uint32(0)
	infoAttr := 0

	switch policy.RecordExistsAction {
	case UPDATE:
	case UPDATE_ONLY:
		infoAttr |= _INFO3_UPDATE_ONLY
	case REPLACE:
		infoAttr |= _INFO3_CREATE_OR_REPLACE
	case REPLACE_ONLY:
		infoAttr |= _INFO3_REPLACE_ONLY
	case CREATE_ONLY:
		writeAttr |= _INFO2_CREATE_ONLY
	}

	switch policy.GenerationPolicy {
	case NONE:
	case EXPECT_GEN_EQUAL:
		generation = policy.Generation
		writeAttr |= _INFO2_GENERATION
	case EXPECT_GEN_GT:
		generation = policy.Generation
		writeAttr |= _INFO2_GENERATION_GT
	}

	if policy.CommitLevel == COMMIT_MASTER {
		infoAttr |= _INFO3_COMMIT_MASTER
	}

	if policy.ConsistencyLevel == CONSISTENCY_ALL {
		readAttr |= _INFO1_CONSISTENCY_ALL
	}

	if policy.DurableDelete {
		writeAttr |= _INFO2_DURABLE_DELETE
	}

	// Write all header data except total size which must be written last.
	cmd.dataBuffer[8] = _MSG_REMAINING_HEADER_SIZE // Message header length.
	cmd.dataBuffer[9] = byte(readAttr)
	cmd.dataBuffer[10] = byte(writeAttr)
	cmd.dataBuffer[11] = byte(infoAttr)
	cmd.dataBuffer[12] = 0 // unused
	cmd.dataBuffer[13] = 0 // clear the result code
	cmd.dataOffset = 14
	cmd.WriteUint32(generation)
	cmd.dataOffset = 18
	cmd.WriteUint32(policy.Expiration)

	// Initialize timeout. It will be written later.
	cmd.dataBuffer[22] = 0
	cmd.dataBuffer[23] = 0
	cmd.dataBuffer[24] = 0
	cmd.dataBuffer[25] = 0

	cmd.dataOffset = 26
	cmd.WriteInt16(int16(fieldCount))
	cmd.WriteInt16(int16(operationCount))
	cmd.dataOffset = int(_MSG_TOTAL_HEADER_SIZE)
}

func (cmd *baseCommand) writeKey(key *Key, sendKey bool) {
	// Write key into buffer.
	if key.namespace != "" {
		cmd.writeFieldString(key.namespace, NAMESPACE)
	}

	if key.setName != "" {
		cmd.writeFieldString(key.setName, TABLE)
	}

	cmd.writeFieldBytes(key.digest[:], DIGEST_RIPE)

	if sendKey {
		cmd.writeFieldValue(key.userKey, KEY)
	}
}

func (cmd *baseCommand) writeOperationForBin(bin *Bin, operation OperationType) error {
	nameLength := copy(cmd.dataBuffer[(cmd.dataOffset+int(_OPERATION_HEADER_SIZE)):], bin.Name)

	// check for float support
	cmd.checkServerCompatibility(bin.Value)

	valueLength, err := bin.Value.estimateSize()
	if err != nil {
		return err
	}

	cmd.WriteInt32(int32(nameLength + valueLength + 4))
	cmd.WriteByte((operation.op))
	cmd.WriteByte((byte(bin.Value.GetType())))
	cmd.WriteByte((byte(0)))
	cmd.WriteByte((byte(nameLength)))
	cmd.dataOffset += nameLength
	_, err = bin.Value.write(cmd)
	return err
}

func (cmd *baseCommand) writeOperationForBinNameAndValue(name string, val interface{}, operation OperationType) error {
	nameLength := copy(cmd.dataBuffer[(cmd.dataOffset+int(_OPERATION_HEADER_SIZE)):], name)

	v := NewValue(val)

	// check for float support
	cmd.checkServerCompatibility(v)

	valueLength, err := v.estimateSize()
	if err != nil {
		return err
	}

	cmd.WriteInt32(int32(nameLength + valueLength + 4))
	cmd.WriteByte((operation.op))
	cmd.WriteByte((byte(v.GetType())))
	cmd.WriteByte((byte(0)))
	cmd.WriteByte((byte(nameLength)))
	cmd.dataOffset += nameLength
	_, err = v.write(cmd)
	return err
}

func (cmd *baseCommand) writeOperationForOperation(operation *Operation) error {
	nameLength := copy(cmd.dataBuffer[(cmd.dataOffset+int(_OPERATION_HEADER_SIZE)):], operation.binName)

	// check for float support
	cmd.checkServerCompatibility(operation.binValue)

	if operation.used {
		// cahce will set the used flag to false again
		operation.cache()
	}

	if operation.encoder == nil {
		valueLength, err := operation.binValue.estimateSize()
		if err != nil {
			return err
		}

		cmd.WriteInt32(int32(nameLength + valueLength + 4))
		cmd.WriteByte((operation.opType.op))
		cmd.WriteByte((byte(operation.binValue.GetType())))
		cmd.WriteByte((byte(0)))
		cmd.WriteByte((byte(nameLength)))
		cmd.dataOffset += nameLength
		_, err = operation.binValue.write(cmd)
		return err
	} else {
		valueLength, err := operation.encoder(operation, nil)
		if err != nil {
			return err
		}

		cmd.WriteInt32(int32(nameLength + valueLength + 4))
		cmd.WriteByte((operation.opType.op))
		cmd.WriteByte((byte(ParticleType.BLOB)))
		cmd.WriteByte((byte(0)))
		cmd.WriteByte((byte(nameLength)))
		cmd.dataOffset += nameLength
		_, err = operation.encoder(operation, cmd)
		//mark the operation as used, so that it will be cached the next time it is used
		operation.used = err == nil
		return err
	}
}

func (cmd *baseCommand) writeOperationForBinName(name string, operation OperationType) {
	nameLength := copy(cmd.dataBuffer[(cmd.dataOffset+int(_OPERATION_HEADER_SIZE)):], name)
	cmd.WriteInt32(int32(nameLength + 4))
	cmd.WriteByte((operation.op))
	cmd.WriteByte(byte(0))
	cmd.WriteByte(byte(0))
	cmd.WriteByte(byte(nameLength))
	cmd.dataOffset += nameLength
}

func (cmd *baseCommand) writeOperationForOperationType(operation OperationType) {
	cmd.WriteInt32(int32(4))
	cmd.WriteByte(operation.op)
	cmd.WriteByte(0)
	cmd.WriteByte(0)
	cmd.WriteByte(0)
}

// TODO: Remove this method and move it to the appropriate VALUE method
func (cmd *baseCommand) checkServerCompatibility(val Value) {
	if val == nil {
		return
	}

	// check for float support
	switch val.GetType() {
	case ParticleType.FLOAT:
		if !cmd.node.supportsFloat.Get() {
			panic("This cluster node doesn't support double precision floating-point values.")
		}
	case ParticleType.GEOJSON:
		if !cmd.node.supportsGeo.Get() {
			panic("This cluster node doesn't support geo-spatial features.")
		}
	}
}

func (cmd *baseCommand) writeFieldValue(value Value, ftype FieldType) error {
	// check for float support
	cmd.checkServerCompatibility(value)

	vlen, err := value.estimateSize()
	if err != nil {
		return err
	}
	cmd.writeFieldHeader(vlen+1, ftype)
	cmd.WriteByte(byte(value.GetType()))

	_, err = value.write(cmd)
	return err
}

func (cmd *baseCommand) writeUdfArgs(value *ValueArray) error {
	if value != nil {
		vlen, err := value.estimateSize()
		if err != nil {
			return err
		}
		cmd.writeFieldHeader(vlen, UDF_ARGLIST)
		_, err = value.pack(cmd)
		return err
	}

	cmd.writeFieldHeader(0, UDF_ARGLIST)
	return nil

}

func (cmd *baseCommand) writeFieldString(str string, ftype FieldType) {
	len := copy(cmd.dataBuffer[(cmd.dataOffset+int(_FIELD_HEADER_SIZE)):], str)
	cmd.writeFieldHeader(len, ftype)
	cmd.dataOffset += len
}

func (cmd *baseCommand) writeFieldBytes(bytes []byte, ftype FieldType) {
	copy(cmd.dataBuffer[cmd.dataOffset+int(_FIELD_HEADER_SIZE):], bytes)

	cmd.writeFieldHeader(len(bytes), ftype)
	cmd.dataOffset += len(bytes)
}

func (cmd *baseCommand) writeFieldHeader(size int, ftype FieldType) {
	cmd.WriteInt32(int32(size + 1))
	cmd.WriteByte((byte(ftype)))
}

// Int64ToBytes converts an int64 into slice of Bytes.
func (cmd *baseCommand) WriteInt64(num int64) (int, error) {
	return cmd.WriteUint64(uint64(num))
}

// Uint64ToBytes converts an uint64 into slice of Bytes.
func (cmd *baseCommand) WriteUint64(num uint64) (int, error) {
	binary.BigEndian.PutUint64(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+8], num)
	cmd.dataOffset += 8
	return 8, nil
}

// Int32ToBytes converts an int32 to a byte slice of size 4
func (cmd *baseCommand) WriteInt32(num int32) (int, error) {
	return cmd.WriteUint32(uint32(num))
}

// Uint32ToBytes converts an uint32 to a byte slice of size 4
func (cmd *baseCommand) WriteUint32(num uint32) (int, error) {
	binary.BigEndian.PutUint32(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+4], num)
	cmd.dataOffset += 4
	return 4, nil
}

// Int16ToBytes converts an int16 to slice of bytes
func (cmd *baseCommand) WriteInt16(num int16) (int, error) {
	return cmd.WriteUint16(uint16(num))
}

// Int16ToBytes converts an int16 to slice of bytes
func (cmd *baseCommand) WriteUint16(num uint16) (int, error) {
	binary.BigEndian.PutUint16(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+2], num)
	cmd.dataOffset += 2
	return 2, nil
}

func (cmd *baseCommand) WriteFloat32(float float32) (int, error) {
	bits := math.Float32bits(float)
	binary.BigEndian.PutUint32(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+4], bits)
	cmd.dataOffset += 4
	return 4, nil
}

func (cmd *baseCommand) WriteFloat64(float float64) (int, error) {
	bits := math.Float64bits(float)
	binary.BigEndian.PutUint64(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+8], bits)
	cmd.dataOffset += 8
	return 8, nil
}

func (cmd *baseCommand) WriteByte(b byte) error {
	cmd.dataBuffer[cmd.dataOffset] = b
	cmd.dataOffset++
	return nil
}

func (cmd *baseCommand) WriteString(s string) (int, error) {
	copy(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+len(s)], s)
	cmd.dataOffset += len(s)
	return len(s), nil
}

func (cmd *baseCommand) Write(b []byte) (int, error) {
	copy(cmd.dataBuffer[cmd.dataOffset:cmd.dataOffset+len(b)], b)
	cmd.dataOffset += len(b)
	return len(b), nil
}

func (cmd *baseCommand) begin() {
	cmd.dataOffset = int(_MSG_TOTAL_HEADER_SIZE)
}

func (cmd *baseCommand) sizeBuffer() error {
	return cmd.sizeBufferSz(cmd.dataOffset)
}

func (cmd *baseCommand) validateHeader(header int64) error {
	msgVersion := (uint64(header) & 0xFF00000000000000) >> 56
	if msgVersion != 2 {
		return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid Message Header: Expected version to be 2, but got %v", msgVersion))
	}

	msgType := uint64((uint64(header) & 0x00FF000000000000)) >> 49
	if !(msgType == 1 || msgType == 3) {
		return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid Message Header: Expected type to be 1 or 3, but got %v", msgType))
	}

	msgSize := int64((header & 0x0000FFFFFFFFFFFF))
	if msgSize > int64(MaxBufferSize) {
		return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid Message Header: Expected size to be under 10MiB, but got %v", msgSize))
	}

	return nil
}

var (
	// MaxBufferSize protects against allocating massive memory blocks
	// for buffers. Tweak this number if you are returning a lot of
	// LDT elements in your queries.
	MaxBufferSize = 1024 * 1024 * 10 // 10 MB
)

func (cmd *baseCommand) sizeBufferSz(size int) error {
	// Corrupted data streams can result in a huge length.
	// Do a sanity check here.
	if size > MaxBufferSize || size < 0 {
		return NewAerospikeError(PARSE_ERROR, fmt.Sprintf("Invalid size for buffer: %d", size))
	}

	if size <= len(cmd.dataBuffer) {
		// don't touch the buffer
	} else if size <= cap(cmd.dataBuffer) {
		cmd.dataBuffer = cmd.dataBuffer[:size]
	} else {
		// not enough space
		cmd.dataBuffer = make([]byte, size)
	}

	return nil
}

func (cmd *baseCommand) end() {
	var size = int64(cmd.dataOffset-8) | (_CL_MSG_VERSION << 56) | (_AS_MSG_TYPE << 48)
	// Buffer.Int64ToBytes(size, cmd.dataBuffer, 0)
	binary.BigEndian.PutUint64(cmd.dataBuffer[0:], uint64(size))
}

////////////////////////////////////

// SetCommandBufferPool can be used to customize the command Buffer Pool parameters to calibrate
// the pool for different workloads
// This method is deprecated.
func SetCommandBufferPool(poolSize, initBufSize, maxBufferSize int) {
	panic("There is no need to optimize the buffer pool anymore. Buffers have moved to Connection object.")
}

func (cmd *baseCommand) execute(ifc command) (err error) {
	policy := ifc.getPolicy(ifc).GetBasePolicy()
	iterations := -1

	// for exponential backoff
	interval := policy.SleepBetweenRetries

	// set timeout outside the loop
	deadline := time.Now().Add(policy.Timeout)

	// Execute command until successful, timed out or maximum iterations have been reached.
	for {
		// too many retries
		if iterations++; (policy.MaxRetries <= 0 && iterations > 0) || (policy.MaxRetries > 0 && iterations > policy.MaxRetries) {
			return NewAerospikeError(TIMEOUT, fmt.Sprintf("command execution timed out: Exceeded number of retries. See `Policy.MaxRetries`. (last error: %s)", err))
		}

		// Sleep before trying again, after the first iteration
		if iterations > 0 && policy.SleepBetweenRetries > 0 {
			time.Sleep(interval)
			if policy.SleepMultiplier > 1 {
				interval = time.Duration(float64(interval) * policy.SleepMultiplier)
			}
		}

		// check for command timeout
		if policy.Timeout > 0 && time.Now().After(deadline) {
			break
		}

		// set command node, so when you return a record it has the node
		cmd.node, err = ifc.getNode(ifc)
		if cmd.node == nil || !cmd.node.IsActive() || err != nil {
			// Node is currently inactive. Retry.
			continue
		}

		// cmd.conn, err = cmd.node.GetConnection(policy.Timeout)
		cmd.conn, err = ifc.getConnection(policy.Timeout)
		if err != nil {
			Logger.Warn("Node " + cmd.node.String() + ": " + err.Error())
			continue
		}

		// Assign the connection buffer to the command buffer
		cmd.dataBuffer = cmd.conn.dataBuffer

		// Set command buffer.
		err = ifc.writeBuffer(ifc)
		if err != nil {
			// All runtime exceptions are considered fatal. Do not retry.
			// Close socket to flush out possible garbage. Do not put back in pool.
			cmd.conn.Close()
			return err
		}

		// Reset timeout in send buffer (destined for server) and socket.
		// Buffer.Int32ToBytes(int32(policy.Timeout/time.Millisecond), cmd.dataBuffer, 22)
		binary.BigEndian.PutUint32(cmd.dataBuffer[22:], uint32(policy.Timeout/time.Millisecond))

		// Send command.
		_, err = cmd.conn.Write(cmd.dataBuffer[:cmd.dataOffset])
		if err != nil {
			// IO errors are considered temporary anomalies. Retry.
			// Close socket to flush out possible garbage. Do not put back in pool.
			cmd.conn.Close()

			Logger.Warn("Node " + cmd.node.String() + ": " + err.Error())
			continue
		}

		// Parse results.
		err = ifc.parseResult(ifc, cmd.conn)
		if err != nil {
			if err == io.EOF {
				// IO errors are considered temporary anomalies. Retry.
				// Close socket to flush out possible garbage. Do not put back in pool.
				cmd.conn.Close()

				Logger.Warn("Node " + cmd.node.String() + ": " + err.Error())

				// retry only for non-streaming commands
				if !cmd.oneShot {
					continue
				}
			}

			// close the connection
			// cancelling/closing the batch/multi commands will return an error, which will
			// close the connection to throw away its data and signal the server about the
			// situation. We will not put back the connection in the buffer.
			if cmd.conn.IsConnected() && KeepConnection(err) {
				// Put connection back in pool.
				cmd.node.PutConnection(cmd.conn)
			} else {
				cmd.conn.Close()

			}
			return err
		}

		// in case it has grown and re-allocated
		cmd.conn.dataBuffer = cmd.dataBuffer

		// Put connection back in pool.
		// cmd.node.PutConnection(cmd.conn)
		ifc.putConnection(cmd.conn)

		// command has completed successfully.  Exit method.
		return nil

	}

	// execution timeout
	return NewAerospikeError(TIMEOUT, "command execution timed out: See `Policy.Timeout`")
}

func (cmd *baseCommand) parseRecordResults(ifc command, receiveSize int) (bool, error) {
	panic(errors.New("Abstract method. Should not end up here"))
}
