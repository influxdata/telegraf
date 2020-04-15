package k8s

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ericchiang/k8s/runtime"
	"github.com/golang/protobuf/proto"
)

const (
	contentTypePB   = "application/vnd.kubernetes.protobuf"
	contentTypeJSON = "application/json"
)

func contentTypeFor(i interface{}) string {
	if _, ok := i.(proto.Message); ok {
		return contentTypePB
	}
	return contentTypeJSON
}

// marshal encodes an object and returns the content type of that resource
// and the marshaled representation.
//
// marshal prefers protobuf encoding, but falls back to JSON.
func marshal(i interface{}) (string, []byte, error) {
	if _, ok := i.(proto.Message); ok {
		data, err := marshalPB(i)
		return contentTypePB, data, err
	}
	data, err := json.Marshal(i)
	return contentTypeJSON, data, err
}

// unmarshal decoded an object given the content type of the encoded form.
func unmarshal(data []byte, contentType string, i interface{}) error {
	msg, isPBMsg := i.(proto.Message)
	if contentType == contentTypePB && isPBMsg {
		if err := unmarshalPB(data, msg); err != nil {
			return fmt.Errorf("decode protobuf: %v", err)
		}
		return nil
	}
	if isPBMsg {
		// only decode into JSON of a protobuf message if the type
		// explicitly implements json.Unmarshaler
		if _, ok := i.(json.Unmarshaler); !ok {
			return fmt.Errorf("cannot decode json payload into protobuf object %T", i)
		}
	}
	if err := json.Unmarshal(data, i); err != nil {
		return fmt.Errorf("decode json: %v", err)
	}
	return nil
}

var magicBytes = []byte{0x6b, 0x38, 0x73, 0x00}

func unmarshalPB(b []byte, msg proto.Message) error {
	if len(b) < len(magicBytes) {
		return errors.New("payload is not a kubernetes protobuf object")
	}
	if !bytes.Equal(b[:len(magicBytes)], magicBytes) {
		return errors.New("payload is not a kubernetes protobuf object")
	}

	u := new(runtime.Unknown)
	if err := u.Unmarshal(b[len(magicBytes):]); err != nil {
		return fmt.Errorf("unmarshal unknown: %v", err)
	}
	return proto.Unmarshal(u.Raw, msg)
}

func marshalPB(obj interface{}) ([]byte, error) {
	message, ok := obj.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("expected obj of type proto.Message, got %T", obj)
	}
	payload, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	// The URL path informs the API server what the API group, version, and resource
	// of the object. We don't need to specify it here to talk to the API server.
	body, err := (&runtime.Unknown{Raw: payload}).Marshal()
	if err != nil {
		return nil, err
	}

	d := make([]byte, len(magicBytes)+len(body))
	copy(d[:len(magicBytes)], magicBytes)
	copy(d[len(magicBytes):], body)
	return d, nil
}
