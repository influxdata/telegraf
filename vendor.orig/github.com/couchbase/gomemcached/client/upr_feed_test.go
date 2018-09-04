package memcached

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/couchbase/gomemcached"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupBoilerPlate() (*vbStreamNegotiator, *UprFeed) {
	negotiator := &vbStreamNegotiator{}
	negotiator.initialize()

	testFeed := &UprFeed{
		vbstreams:  make(map[uint16]*UprStream),
		negotiator: *negotiator,
	}

	return negotiator, testFeed
}

func TestNegotiator(t *testing.T) {
	assert := assert.New(t)
	fmt.Println("============== Test case start: TestNegotiator =================")
	var vbno uint16 = 1
	var opaque uint16 = 2
	opaqueComposed := composeOpaque(vbno, opaque)
	var headerBuf [gomemcached.HDR_LEN]byte

	negotiator, testFeed := setupBoilerPlate()

	_, err := negotiator.getStreamFromMap(1, 2)
	assert.NotNil(err)

	negotiator.registerRequest(vbno, opaque, 3, 4, 5)
	_, err = negotiator.getStreamFromMap(vbno, opaque)
	assert.Nil(err)

	err = testFeed.validateCloseStream(vbno)
	assert.Nil(err)

	request := &gomemcached.MCRequest{Opcode: gomemcached.UPR_STREAMREQ,
		VBucket: vbno,
		Opaque:  opaqueComposed,
	}
	response := &gomemcached.MCResponse{Opcode: gomemcached.UPR_STREAMREQ,
		Opaque: opaqueComposed,
	}

	event, err := negotiator.handleStreamRequest(testFeed, headerBuf, request, 0, response)
	assert.Nil(err)
	assert.NotNil(event)

	fmt.Println("============== Test case end: TestNegotiator =================")
}

func TestNegotiatorMultiSession(t *testing.T) {
	assert := assert.New(t)
	fmt.Println("============== Test case start: TestNegotiatorMultiSession =================")
	var vbno uint16 = 1
	var opaque uint16 = 2
	opaqueComposed := composeOpaque(vbno, opaque)
	var headerBuf [gomemcached.HDR_LEN]byte

	negotiator, testFeed := setupBoilerPlate()

	negotiator.registerRequest(vbno, opaque, 3, 4, 5)
	_, err := negotiator.getStreamFromMap(vbno, opaque)
	assert.Nil(err)

	negotiator.registerRequest(vbno, opaque+1, 3, 4, 5)
	_, err = negotiator.getStreamFromMap(vbno, opaque+1)
	assert.Nil(err)

	request := &gomemcached.MCRequest{Opcode: gomemcached.UPR_STREAMREQ,
		VBucket: vbno,
		Opaque:  opaqueComposed,
	}

	// Assume a response from DCP
	rollbackNumberBuffer := new(bytes.Buffer)
	err = binary.Write(rollbackNumberBuffer, binary.BigEndian, uint64(0))
	assert.Nil(err)

	response := &gomemcached.MCResponse{Opcode: gomemcached.UPR_STREAMREQ,
		Opaque: opaqueComposed,
		Status: gomemcached.ROLLBACK,
		Body:   rollbackNumberBuffer.Bytes(),
	}

	event, err := negotiator.handleStreamRequest(testFeed, headerBuf, request, 0, response)
	assert.Nil(err)
	assert.NotNil(event)

	// After a success, the map should be empty for this one
	_, err = negotiator.getStreamFromMap(vbno, opaque)
	assert.NotNil(err)

	// The second one should still be there
	_, err = negotiator.getStreamFromMap(vbno, opaque+1)
	assert.Nil(err)

	response.Opaque = composeOpaque(vbno, opaque+1)
	event, err = negotiator.handleStreamRequest(testFeed, headerBuf, request, 0, response)
	assert.Nil(err)
	assert.NotNil(event)

	_, err = negotiator.getStreamFromMap(vbno, opaque+1)
	assert.NotNil(err)

	fmt.Println("============== Test case end: TestNegotiatorMultiSession =================")
}
