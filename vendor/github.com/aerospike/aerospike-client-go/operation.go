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

type operation interface {
	write(cmd BufferEx) (int, error)
}

// OperationType determines operation type
type OperationType *struct{ op byte }
type OperationSubType *int

// Valid OperationType values that can be used to create custom Operations.
// The names are self-explanatory.
var (
	READ OperationType = &struct{ op byte }{1}
	// READ_HEADER *OperationType = &struct{op:  1 }

	WRITE      OperationType = &struct{ op byte }{2}
	CDT_READ   OperationType = &struct{ op byte }{3}
	CDT_MODIFY OperationType = &struct{ op byte }{4}
	MAP_READ   OperationType = &struct{ op byte }{3}
	MAP_MODIFY OperationType = &struct{ op byte }{4}
	ADD        OperationType = &struct{ op byte }{5}
	APPEND     OperationType = &struct{ op byte }{9}
	PREPEND    OperationType = &struct{ op byte }{10}
	TOUCH      OperationType = &struct{ op byte }{11}
)

// Operation contains operation definition.
// This struct is used in client's operate() method.
type Operation struct {

	// OpType determines type of operation.
	opType OperationType
	// used in CDT commands
	opSubType OperationSubType

	encoder func(*Operation, BufferEx) (int, error)

	// binName (Optional) determines the name of bin used in operation.
	binName string

	// binValue (Optional) determines bin value used in operation.
	binValue Value

	// will be true ONLY for GetHeader() operation
	headerOnly bool

	// reused determines if the operation is cached. If so, it will cache the
	// internal bytes in binValue field and remove the encoder for maximum performance
	used bool
}

// cache uses the encoder and caches the packed operation for further use.
func (op *Operation) cache() error {
	packer := newPacker()

	if _, err := op.encoder(op, packer); err != nil {
		return err
	}

	op.binValue = BytesValue(packer.Bytes())
	op.encoder = nil // do not encode anymore; just use the cache
	op.used = false  // do not encode anymore; just use the cache
	return nil
}

// GetOpForBin creates read bin database operation.
func GetOpForBin(binName string) *Operation {
	return &Operation{opType: READ, binName: binName, binValue: NewNullValue()}
}

// GetOp creates read all record bins database operation.
func GetOp() *Operation {
	return &Operation{opType: READ, binValue: NewNullValue()}
}

// GetHeaderOp creates read record header database operation.
func GetHeaderOp() *Operation {
	return &Operation{opType: READ, headerOnly: true, binValue: NewNullValue()}
}

// PutOp creates set database operation.
func PutOp(bin *Bin) *Operation {
	return &Operation{opType: WRITE, binName: bin.Name, binValue: bin.Value}
}

// AppendOp creates string append database operation.
func AppendOp(bin *Bin) *Operation {
	return &Operation{opType: APPEND, binName: bin.Name, binValue: bin.Value}
}

// PrependOp creates string prepend database operation.
func PrependOp(bin *Bin) *Operation {
	return &Operation{opType: PREPEND, binName: bin.Name, binValue: bin.Value}
}

// AddOp creates integer add database operation.
func AddOp(bin *Bin) *Operation {
	return &Operation{opType: ADD, binName: bin.Name, binValue: bin.Value}
}

// TouchOp creates touch database operation.
func TouchOp() *Operation {
	return &Operation{opType: TOUCH, binValue: NewNullValue()}
}
