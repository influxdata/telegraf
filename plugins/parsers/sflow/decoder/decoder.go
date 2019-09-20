// Package decoder provides the engine parts for decoding sflow network flow packets. It is basically a set of functors that can be instantiated to provide
// a complete engine to decode packets
package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// ItemDecoder are the base entities of the decoding engine. They decode bytes extracted from an io.Reader and record the results
// in a Recorder
type ItemDecoder interface {
	Decode(r io.Reader, rec Recorder) error
}

// Decode uses the supplied format, expressed as an ItemDecoder, and decodes the provide bytes according to that format.
// Should all proceed well then the results is provided via a string->interface{} map which may iteself contains arrays and other sub maps.
func Decode(format ItemDecoder, bytes *bytes.Buffer) (map[string]interface{}, error) {
	recorder := newDefaultRecorder()
	e := format.Decode(bytes, recorder)
	return recorder.GetRecording(), e
}

// Recorder provides the interface for ItemDecoders to write entries into
type Recorder interface {
	lookup(key string) (interface{}, bool)
	record(name string, value interface{})
	nest(key string, size uint32) NestedRecorder
}

// NestedRecorder is a type of Recoder that can iterate
type NestedRecorder interface {
	next() (Recorder, bool)
}

// WarnAndBreak answers an ItemDecoder that will log a warning and break an iteration
func WarnAndBreak(fieldName string, msgFmt string, optionalLookupFieldName string) ItemDecoder {
	return &warnAndBreakDecoder{fieldName, msgFmt, optionalLookupFieldName}
}

// Eql answers an ItemDecoder that may be used within an Alt ItemDecoder to form a case statement with optional paths
func Eql(k string, v interface{}, d ItemDecoder) *GuardItemDecoder {
	return &GuardItemDecoder{k, v, d, false}
}

// AltDefault answers an ItemDecoder that can be used as the default entry in an Alt ItemDecoder
func AltDefault(d ItemDecoder) *GuardItemDecoder {
	return &GuardItemDecoder{"", nil, d, true}
}

// UI16 will parse an unsigned 16 bit interger and record the value with the specified k name. Optional a map of functions can be supplied
// to take the decoded value as an input and provide an alternative key and value to record
func UI16(k string, fn ...func(uint16) (string, uint16)) ItemDecoder {
	return &uint16Decoder{k, fn}
}

// I32 will parse an signed 32 bit interger and record the value with the specified k name. Optional a map of functions can be supplied
// to take the decoded value as an input and provide an alternative key and value to record
func I32(k string, fn ...func(int32) (string, int32)) ItemDecoder {
	return &int32Decoder{k, fn, nil}
}

// UI32 will parse an unsigned 32 bit interger and record the value with the specified k name. Optional a map of functions can be supplied
// to take the decoded value as an input and provide an alternative key and value to record
func UI32(k string, fn ...func(uint32) (string, uint32)) ItemDecoder {
	return &uint32Decoder{k, fn, nil}
}

// UI32Mapped will parse an unsigned 32 bit interger and record the value with the specified k name having ampped the value to a one
// of the strings in the proviced map
func UI32Mapped(k string, toMap map[uint32]string) ItemDecoder {
	return &uint32Decoder{k, nil, toMap}
}

// Seq answers an ItemDecoder that will sequentially decoded the provided list of ItemDecoders
func Seq(Decoders ...ItemDecoder) ItemDecoder {
	return &seqDecoder{Decoders}
}

// Alt answers an ItemDecoder that takes a sequence of GuardItemDecoders, will check them one by one to see
// which one accepts the path the decode and will decode using that selected Gaurded Decoder
func Alt(ident string, Decoders ...*GuardItemDecoder) ItemDecoder {
	return &altDecoder{ident, Decoders}
}

// UI64 will parse an unsigned 364 bit interger and record the value with the specified k name. Optional a map of functions can be supplied
// to take the decoded value as an input and provide an alternative key and value to record
func UI64(k string, fn ...func(uint64) (string, uint64)) ItemDecoder {
	return &uint64Decoder{k, fn, nil}
}

// AsrtMax answer an ItemDecoder that doesn't decode any data but will check a previously recorded value to see that it falls
// below a specified limit
func AsrtMax(srcKey string, value interface{}, location string, debugInsteadOfWarn bool) ItemDecoder {
	return &asrtMaxDecoder{srcKey, value, location, debugInsteadOfWarn}
}

// Asgn answers an ItemDecoder that will take one recorded value and popualte a new one (copy) with the same value
func Asgn(srcKey string, dstKey string) ItemDecoder {
	return &asgnDecoder{srcKey, dstKey}
}

// Sub answers an ItemDecoder that will lookup the specified key as a length of following buffer to extra for encapsulated decoding
func Sub(k string, d ItemDecoder) ItemDecoder {
	return &subBuffDecoder{k, []ItemDecoder{d}}
}

// Bin answers an ItemDecoder that will decode the specified number of bytes. Optionally the bytes may be put through a single mapping
// function to convert them to another type
func Bin(k string, l int, fn ...func([]byte) interface{}) ItemDecoder {
	if len(fn) > 0 {
		if len(fn) > 1 {
			panic("too manhy functions")
		}
		return &binDecoder{k, l, fn[0]}
	} else {
		return &binDecoder{k, l, nil}
	}
}

// Nest answers an ItenDecoder that will nest the subsequent decoder within a named sub-tree
func Nest(k string, d ItemDecoder) ItemDecoder {
	return &nestItemDecoder{k, d}
}

// Iter answers an ItemDecoder that will iterate number of times recored in the k value and record the entries against an array with name n in the parent
func Iter(n string, k string, d ItemDecoder) *IterItemDecoder {
	return &IterItemDecoder{n, k, d}
}

// UnwrapError is a type of error that is used by the WarnAndBreak to trigger a break in iteration of the decode
type UnwrapError string

// GuardItemDecoder is a type of ItemDecoder that is gaurded by a condition and forms part of the Alt ItemDecoder
type GuardItemDecoder struct {
	key            string
	equals         interface{}
	Decoder        ItemDecoder
	ifAllElseFails bool
}

// Decode is the ItemDecoder implementation for GuardItemDecoder
func (d *GuardItemDecoder) Decode(r io.Reader, rec Recorder) error {
	return d.Decoder.Decode(r, rec)
}

// IterItemDecoder is an ItemDecoder that will iterate a given number of times
type IterItemDecoder struct {
	name        string
	key         string
	ItemDecoder ItemDecoder
}

// Decode is ItemDecoder implementation for IterItemDecoder
func (d *IterItemDecoder) Decode(r io.Reader, rec Recorder) error {
	key := d.key
	v, ok := rec.lookup(key)
	if ok {
		switch t := v.(type) {
		case uint32:
			nestRec := rec.nest(d.name, t)
			for i := 0; uint32(i) < t; i++ {
				nestedRec, _ := nestRec.next()
				if e := d.ItemDecoder.Decode(r, nestedRec); e != nil {
					return e
				}
			}
			return nil
		default:
			return fmt.Errorf("unhandled type %T at name(%s) key(%s)", v, d.name, d.key)
		}
	}
	return fmt.Errorf("unable to find key %s", key)
}

func (d *GuardItemDecoder) accept(rec Recorder) (bool, error) {
	v, ok := rec.lookup(d.key)
	if ok {
		switch t := v.(type) {
		case uint16:
			if t == d.equals {
				return true, nil
			}
		case uint32:
			if t == d.equals {
				return true, nil
			}
		case *uint32:
			if *t == d.equals {
				return true, nil
			}
		case string:
			if t == d.equals {
				return true, nil
			}
		default:
			return false, fmt.Errorf("unhandled type %T", v)
		}
	} else if !d.ifAllElseFails {
		fmt.Println("couldn't find", d.key)
	}
	if d.ifAllElseFails {
		return true, nil
	}
	return false, nil
}

type defaultRecorder struct {
	recording map[string]interface{}
	parent    Recorder
}

func newDefaultRecorder() *defaultRecorder {
	return &defaultRecorder{recording: make(map[string]interface{})}
}

func (r *defaultRecorder) nest(key string, size uint32) NestedRecorder {
	v := make([]map[string]interface{}, size)
	r.record(key, v)
	return &defaultNestedRecorder{v, 0, r}
}

func (r *defaultRecorder) lookup(key string) (interface{}, bool) {
	v, ok := r.recording[key]
	if !ok && r.parent != nil {
		// look in our parent
		return r.parent.lookup(key)
	}
	return v, ok

}
func (r *defaultRecorder) record(name string, value interface{}) {
	r.recording[name] = value
}

func (r *defaultRecorder) GetRecording() map[string]interface{} {
	return r.recording
}

type defaultNestedRecorder struct {
	ary    []map[string]interface{}
	index  int
	parent Recorder
}

func (d *defaultNestedRecorder) next() (Recorder, bool) {
	if d.index == len(d.ary) {
		return nil, false
	}
	result := make(map[string]interface{})
	d.ary[d.index] = result
	if d.index < len(d.ary)-1 {
		d.index++
		d.ary[d.index] = make(map[string]interface{})
	}
	return &defaultRecorder{result, d.parent}, true
}

type seqDecoder struct {
	Decoders []ItemDecoder
}

func (d *seqDecoder) Decode(r io.Reader, rec Recorder) error {
	for _, sd := range d.Decoders {
		e := sd.Decode(r, rec)
		if e != nil {
			return e
		}
	}
	return nil
}

type altDecoder struct {
	ident    string
	Decoders []*GuardItemDecoder
}

func (d *altDecoder) Decode(r io.Reader, rec Recorder) error {
	for _, cd := range d.Decoders {
		a, e := cd.accept(rec)
		if e != nil {
			return e
		}
		if a {
			return cd.Decode(r, rec)
		}
	}
	return fmt.Errorf("Non of the alternatives accept %s", d.ident)
}

func (e UnwrapError) Error() string {
	return string(e)
}

type warnAndBreakDecoder struct {
	fieldName               string
	msgFmt                  string
	optionalLookupFieldName string
}

func (d *warnAndBreakDecoder) Decode(r io.Reader, rec Recorder) error {
	var v interface{}
	if d.optionalLookupFieldName != "" {
		v, _ = rec.lookup(d.optionalLookupFieldName)
	}
	rec.record(d.fieldName, fmt.Sprintf(d.msgFmt, v))
	return UnwrapError(fmt.Sprintf(d.msgFmt, v))
}

type nestItemDecoder struct {
	name        string
	ItemDecoder ItemDecoder
}

func (d *nestItemDecoder) Decode(r io.Reader, rec Recorder) error {
	nestRec := rec.nest(d.name, 1)
	nestedRec, _ := nestRec.next()
	return d.ItemDecoder.Decode(r, nestedRec)
}

type uint16Decoder struct {
	name string
	fn   []func(uint16) (string, uint16)
}

func (d *uint16Decoder) Decode(r io.Reader, rec Recorder) error {
	var value uint16
	err := binary.Read(r, binary.BigEndian, &value)
	if err != nil {
		return err
	}
	if d.name != "" {
		rec.record(d.name, value)
	}
	for _, f := range d.fn {
		n, v := f(value)
		rec.record(n, v)
	}
	return nil
}

type int32Decoder struct {
	name  string
	fn    []func(int32) (string, int32)
	toMap map[int32]string
}

func (d *int32Decoder) Decode(r io.Reader, rec Recorder) error {
	var value int32
	err := binary.Read(r, binary.BigEndian, &value)
	if err != nil {
		return err
	}
	if d.name != "" {
		if d.toMap != nil {
			mappedValue, ok := d.toMap[value]
			if !ok {
				return fmt.Errorf("unable to map %d", value)
			}
			rec.record(d.name, mappedValue)
		} else {
			rec.record(d.name, value)
		}
	}
	for _, f := range d.fn {
		n, v := f(value)
		rec.record(n, v)
	}
	return nil
}

type uint32Decoder struct {
	name  string
	fn    []func(uint32) (string, uint32)
	toMap map[uint32]string
}

func (d *uint32Decoder) Decode(r io.Reader, rec Recorder) error {
	var value uint32
	err := binary.Read(r, binary.BigEndian, &value)
	if err != nil {
		return err
	}
	if d.name != "" {
		if d.toMap != nil {
			mappedValue, ok := d.toMap[value]
			if !ok {
				return fmt.Errorf("unable to map %d", value)
			}
			rec.record(d.name, mappedValue)
		} else {
			rec.record(d.name, value)
		}
	}
	for _, f := range d.fn {
		n, v := f(value)
		rec.record(n, v)
	}
	return nil
}

type uint64Decoder struct {
	name  string
	fn    []func(uint64) (string, uint64)
	toMap map[uint64]string
}

func (d *uint64Decoder) Decode(r io.Reader, rec Recorder) error {
	var value uint64
	err := binary.Read(r, binary.BigEndian, &value)
	if err != nil {
		return err
	}
	if d.name != "" {
		if d.toMap != nil {
			mappedValue, ok := d.toMap[value]
			if !ok {
				return fmt.Errorf("unable to map %d", value)
			}
			rec.record(d.name, mappedValue)
		} else {
			rec.record(d.name, value)
		}
	}
	for _, f := range d.fn {
		n, v := f(value)
		rec.record(n, v)
	}
	return nil
}

type binDecoder struct {
	name string
	size int
	fn   func([]byte) interface{}
}

func (b *binDecoder) Decode(r io.Reader, rec Recorder) error {
	v := make([]byte, b.size)
	e := binary.Read(r, binary.BigEndian, v)
	if e != nil {
		return e
	}
	if b.fn != nil {
		rec.record(b.name, b.fn(v))
	} else {
		rec.record(b.name, v)
	}
	return nil
}

type subBuffDecoder struct {
	key        string
	processors []ItemDecoder
}

func (s *subBuffDecoder) Decode(r io.Reader, rec Recorder) error {
	length, ok := rec.lookup(s.key)
	if ok {
		buff, ok := r.(*bytes.Buffer)
		if !ok {
			return fmt.Errorf("can't convert to *bytes.Buffer %T", r)
		}
		lUint32, ok := length.(uint32)
		lengthInt := int(lUint32)
		if !ok {
			return fmt.Errorf("can't convert to uint32 %T", length)
		}
		sampleReader := bytes.NewBuffer(buff.Next(lengthInt))
		for _, p := range s.processors {
			if p != nil {
				e := p.Decode(sampleReader, rec)
				if e != nil {
					if _, ok := e.(UnwrapError); ok {
						// It is an UnwrapError to stop processing any further decoders at this level and continue (return no error)
						// at the higher level
						return nil
					}
					return e
				}
			}
		}
		return nil
	}
	return fmt.Errorf("unabl to find sub length value from key %s", s.key)
}

type asgnDecoder struct {
	srcKey string
	dstKey string
}

func (d *asgnDecoder) Decode(r io.Reader, rec Recorder) error {
	v, ok := rec.lookup(d.srcKey)
	if ok {
		rec.record(d.dstKey, v)
		return nil
	}
	return fmt.Errorf("assigne cannot find source %s", d.srcKey)
}

type asrtMaxDecoder struct {
	srcKey             string
	value              interface{}
	location           string
	debugInsteadOfWarn bool
}

func (d *asrtMaxDecoder) Decode(_ io.Reader, rec Recorder) error {
	v, ok := rec.lookup(d.srcKey)
	if ok {
		notGreater := false
		var validTest bool //:= false
		switch t := v.(type) {
		case uint:
			valueAsUint, ok := d.value.(uint)
			validTest = ok
			notGreater = t <= valueAsUint
		case uint16:
			valueAsUint16, ok := d.value.(uint16)
			validTest = ok
			notGreater = t <= valueAsUint16
		case uint32:
			valueAsUint32, ok := d.value.(uint32)
			validTest = ok
			notGreater = t <= valueAsUint32
		}
		if validTest {
			if !notGreater {
				msg := fmt.Sprintf("asrtMax %s = %v > %v @ %s", d.srcKey, v, d.value, d.location)
				if d.debugInsteadOfWarn {
					log.Printf("D! %s", msg)
				} else {
					log.Printf("W! %s", msg)
				}
				//os.Exit(1)
				return UnwrapError(msg)
			}
			return nil
		}
		return fmt.Errorf("artMax unable to compare %T and %T @ %s", v, d.value, d.location)
	}
	return fmt.Errorf("asrtMax cannot find source %s @ %s", d.srcKey, d.location)
}
