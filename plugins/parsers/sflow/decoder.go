package sflow

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func Decode(format ItemDecoder, bytes *bytes.Buffer) (map[string]interface{}, error) {
	recorder := newDefaultRecorder()
	e := format.Decode(bytes, recorder)
	if e != nil {
		return recorder.GetRecording(), e
	}
	return recorder.GetRecording(), nil
}

// Recorder probably going to do away with this, or rather not export it.
type Recorder interface {
	lookup(key string) (interface{}, bool)
	record(name string, value interface{})
	nest(key string, size uint32) NestedRecorder
}

type NestedRecorder interface {
	next() (Recorder, bool)
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

func seq(Decoders ...ItemDecoder) *seqDecoder {
	return &seqDecoder{Decoders}
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
	Decoders []*guardItemDecoder
}

func alt(ident string, Decoders ...*guardItemDecoder) ItemDecoder {
	return &altDecoder{ident, Decoders}
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

type guardItemDecoder struct {
	key            string
	equals         interface{}
	Decoder        ItemDecoder
	ifAllElseFails bool
}

func (d *guardItemDecoder) Decode(r io.Reader, rec Recorder) error {
	return d.Decoder.Decode(r, rec)
}

func (d *guardItemDecoder) accept(rec Recorder) (bool, error) {
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

// UnwrapError
type UnwrapError string

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

func warnAndBreak(fieldName string, msgFmt string, optionalLookupFieldName string) ItemDecoder {
	return &warnAndBreakDecoder{fieldName, msgFmt, optionalLookupFieldName}
}

func eql(k string, v interface{}, d ItemDecoder) *guardItemDecoder {
	return &guardItemDecoder{k, v, d, false}
}

func altDefault(d ItemDecoder) *guardItemDecoder {
	return &guardItemDecoder{"", nil, d, true}
}

func ui16(k string, fn ...func(uint16) (string, uint16)) ItemDecoder {
	return &uint16Decoder{k, fn}
}

func i32(k string, fn ...func(int32) (string, int32)) ItemDecoder {
	return &int32Decoder{k, fn, nil}
}

func ui32(k string, fn ...func(uint32) (string, uint32)) ItemDecoder {
	return &uint32Decoder{k, fn, nil}
}

func ui64(k string, fn ...func(uint64) (string, uint64)) ItemDecoder {
	return &uint64Decoder{k, fn, nil}
}

func ui32Mapped(k string, toMap map[uint32]string) ItemDecoder {
	return &uint32Decoder{k, nil, toMap}
}

func bin(k string, l int, fn ...func([]byte) interface{}) ItemDecoder {
	if len(fn) > 0 {
		if len(fn) > 1 {
			panic("too manhy functions")
		}
		return &binDecoder{k, l, fn[0]}
	} else {
		return &binDecoder{k, l, nil}
	}
}

func sub(k string, d ItemDecoder) ItemDecoder {
	return &subBuffDecoder{k, []ItemDecoder{d}}
}

func nest(k string, d ItemDecoder) ItemDecoder {
	return &nestItemDecoder{k, d}
}

type iterItemDecoder struct {
	name        string
	key         string
	max         uint32
	ItemDecoder ItemDecoder
}

func iter(n string, k string, max uint32, d ItemDecoder) *iterItemDecoder {
	return &iterItemDecoder{n, k, max, d}
}

func (d *iterItemDecoder) Decode(r io.Reader, rec Recorder) error {
	key := d.key
	v, ok := rec.lookup(key)
	if ok {
		switch t := v.(type) {
		case uint32:
			if d.max > 0 && t > d.max {
				return fmt.Errorf("iteration for key %s exceeds max of %d at %d", key, d.max, t)
			}
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

type nestItemDecoder struct {
	name        string
	ItemDecoder ItemDecoder
}

func (d *nestItemDecoder) Decode(r io.Reader, rec Recorder) error {
	nestRec := rec.nest(d.name, 1)
	nestedRec, _ := nestRec.next()
	return d.ItemDecoder.Decode(r, nestedRec)
}

type ItemDecoder interface {
	Decode(r io.Reader, rec Recorder) error
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

type ifDecoder struct {
	key     string
	value   interface{}
	Decoder ItemDecoder
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
			return fmt.Errorf("can't convert to int %T", length)
		}
		sampleReader := bytes.NewBuffer(buff.Next(lengthInt))
		for _, p := range s.processors {
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
		return nil
	}
	return fmt.Errorf("unabl to find sub length value from key %s", s.key)
}

type remainingBytesDecoder struct {
	key string
}

func (d *remainingBytesDecoder) Decode(r io.Reader, rec Recorder) error {
	buff, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("can't convert to *bytes.Buffer %T", r)
	}
	v := buff.Bytes()
	rec.record(d.key, v)
	return nil
}
func remainingBytes(key string) *remainingBytesDecoder {
	return &remainingBytesDecoder{key}
}

type asgnDecoder struct {
	srcKey string
	dstKey string
}

func asgn(srcKey string, dstKey string) ItemDecoder {
	return &asgnDecoder{srcKey, dstKey}
}

func (d *asgnDecoder) Decode(r io.Reader, rec Recorder) error {
	v, ok := rec.lookup(d.srcKey)
	if ok {
		rec.record(d.dstKey, v)
		return nil
	}
	return fmt.Errorf("assigne cannot find source %s", d.srcKey)
}
