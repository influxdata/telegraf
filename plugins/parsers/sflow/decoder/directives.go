package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Directive is a Decode Directive, the basic building block of a decoder
type Directive interface {

	// execute performs the function of the decode directive. If DecodeContext is nil then the
	// ask is to check that a subsequent execution (with non nill DecodeContext) is expted to work.
	execute(*bytes.Buffer, *DecodeContext) error

	// Reset the internal state of the decode director to the initial state prior to any decoding.
	// This allows decode directives to be reused and ensure there is no residual state carried over
	Reset()
}

// ValueDirective is a decode directive that extracts some data from the packet, an integer or byte maybe,
// which it then processes by using it, for example, as the counter for the number of iterations to perform
// of downstream decode directives.
//
// A ValueDirective can be used to either Switch, Iter(ate), Encapsulate or Do mutually exclusively.
type ValueDirective interface {
	Directive

	// Switch attaches a set of conditional decode directives downstream of this decode directive
	Switch(paths ...CaseValueDirective) ValueDirective

	// Iter attaches a single downstream decode directive that will be executed repeatedly according to the iteration count
	Iter(maxIterations uint32, dd Directive) ValueDirective

	// Encapsulated will form a new buffer of the encapsulated length and pass that buffer on to the downsstream decode directive
	Encapsulated(maxSize uint32, dd Directive) ValueDirective

	// Ref records this decode directive in the passed reference
	Ref(*interface{}) ValueDirective

	// Do attaches a Decode Operation - these are uses of the decoded information to perform work on, transform, write out etc.
	Do(ddo DirectiveOp) ValueDirective
}

type valueDirective struct {
	reference *valueDirective

	value    interface{}
	noDecode bool

	cases            []CaseValueDirective
	iter             Directive
	maxIterations    uint32
	encapsulated     Directive
	maxEncapsulation uint32
	ops              []DirectiveOp
	err              error

	location string

	resetFn func(interface{})
}

func (dd *valueDirective) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	dc.tracef("%s execute\n", dd.location)

	if dd.reference == nil && !dd.noDecode {
		if e := binary.Read(buffer, binary.BigEndian, dd.value); e != nil {
			return e
		}
	}

	// Switch downstream?
	if dd.cases != nil && len(dd.cases) > 0 {
		for i, c := range dd.cases {
			if c.equals(dd.value) {
				dc.tracef("%s selected case %d\n", dd.location, i)
				return c.execute(buffer, dc)
			}
		}
		switch v := dd.value.(type) {
		case *uint32:
			return fmt.Errorf("(%T).Switch,unmatched case %d - created at %s", v, *v, dd.location)
		case *uint16:
			return fmt.Errorf("(%T).Switch,unmatched case %d - created at %s", v, *v, dd.location)
		default:
			return fmt.Errorf("(%T).Switch,unmatched case %v - created at %s", dd.value, dd.value, dd.location)
		}
	}

	// Iter downstream?
	if dd.iter != nil {
		fn := func(id interface{}) error {
			dc.tracef("%s iteration %+v\n", dd.location, id)
			if e := dd.iter.execute(buffer, dc); e != nil {
				return e
			}
			return nil
		}
		switch v := dd.value.(type) {
		case *uint32:
			if *v > dd.maxIterations {
				return fmt.Errorf("iter at %s exceeds configured max - value %d, limit %d", dd.location, *v, dd.maxIterations)
			}
			for i := uint32(0); i < *v; i++ {
				if e := fn(i); e != nil {
					return e
				}
			}
		default:
			// Can't actually get here if .Iter method check types (and it does)
			return fmt.Errorf("(%T).Iter, cannot iterator over this type at %s", dd.value, dd.location)
		}
	}

	// Encapsualted downstream>
	if dd.encapsulated != nil {
		switch v := dd.value.(type) {
		case *uint32:
			if *v > dd.maxEncapsulation {
				return fmt.Errorf("encap at %s exceeds configured max - value %d, limit %d", dd.location, *v, dd.maxEncapsulation)
			}
			dc.tracef("%s encapsulated\n", dd.location)
			return dd.encapsulated.execute(bytes.NewBuffer(buffer.Next(int(*v))), dc)
		}
	}

	// Perform the attached operations
	for i, op := range dd.ops {
		dc.tracef("%s do(%d)\n", dd.location, i)
		if err := op.process(dc, dd.value); err != nil {
			return err
		}
	}

	return nil
}

// panickIfNotBlackCanvas checks the state of this value directive to see if it is has
// alrady been configured in a manner inconsistent with another configuration change
func (dd *valueDirective) panickIfNotBlackCanvas(change string, checkDOs bool) {
	if dd.cases != nil {
		panic(fmt.Sprintf("already have switch cases assigned, cannot assign %s @ %s", change, dd.location))
	}
	if dd.iter != nil {
		panic(fmt.Sprintf("already have iter assigned, cannot assign %s @ %s", change, dd.location))
	}
	if dd.encapsulated != nil {
		panic(fmt.Sprintf("already have encap assigned, cannot assign %s @ %s", change, dd.location))
	}
	if checkDOs && dd.ops != nil && len(dd.ops) > 0 {
		panic(fmt.Sprintf("already have do assigned, cannot assign %s @ %s", change, dd.location))
	}
}

func (dd *valueDirective) Switch(paths ...CaseValueDirective) ValueDirective {
	dd.panickIfNotBlackCanvas("new switch", true)
	dd.cases = paths
	return dd
}

func (dd *valueDirective) Iter(maxIterations uint32, iter Directive) ValueDirective {
	dd.panickIfNotBlackCanvas("new iter", true)
	switch dd.value.(type) {
	case *uint32:
	default:
		panic(fmt.Sprintf("cannot iterate a %T", dd.value))
	}

	dd.iter = iter
	dd.maxIterations = maxIterations
	return dd
}

func (dd *valueDirective) Encapsulated(maxSize uint32, encapsulated Directive) ValueDirective {
	dd.panickIfNotBlackCanvas("new encapsulated", true)
	switch dd.value.(type) {
	case *uint32:
	default:
		panic(fmt.Sprintf("cannot encapsulated on a %T", dd.value))
	}

	dd.encapsulated = encapsulated
	dd.maxEncapsulation = maxSize
	return dd
}

func (dd *valueDirective) Do(ddo DirectiveOp) ValueDirective {
	dd.panickIfNotBlackCanvas("new do", false)
	for {
		if ddo.prev() == nil {
			break
		}
		ddo = ddo.prev()
	}
	if err := ddo.process(nil, dd.value); err != nil {
		panic(fmt.Sprintf("directive operation %T cannot process %T - %s", ddo, dd.value, err))
	}
	if dd.ops == nil {
		dd.ops = make([]DirectiveOp, 0, 5)
	}
	dd.ops = append(dd.ops, ddo)

	return dd
}

func (dd *valueDirective) Ref(ref *interface{}) ValueDirective {
	if *ref != nil {
		panic("ref already assigned, not overwritting")
	}
	*ref = dd
	return dd
}

func (dd *valueDirective) Reset() {
	if dd.resetFn != nil {
		dd.resetFn(dd.value)
	}
	for _, c := range dd.cases {
		c.Reset()
	}
	if dd.iter != nil {
		dd.iter.Reset()
	}
	if dd.encapsulated != nil {
		dd.encapsulated.Reset()
	}
}

// errorDirective a decode directive that reports an error
type errorDirective struct {
	Directive
}

func (dd *errorDirective) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	return fmt.Errorf("Error Directive")
}

// CaseValueDirective is a decode directive that also has a switch/case test
type CaseValueDirective interface {
	Directive
	equals(interface{}) bool
}

type caseValueDirective struct {
	caseValue interface{}
	isDefault bool
	equalsDd  Directive
}

func (dd *caseValueDirective) Reset() {
	if dd.equalsDd != nil {
		dd.equalsDd.Reset()
	}
}

func (dd *caseValueDirective) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	if dd.equalsDd == nil {
		return nil
	}
	return dd.equalsDd.execute(buffer, dc)
}

func (dd *caseValueDirective) equals(value interface{}) bool {
	if dd.isDefault {
		return true
	}
	switch ourV := dd.caseValue.(type) {
	case uint32:
		ov, ok := value.(*uint32)
		if ok {
			return ourV == *ov
		}
		log.Printf("D! value not a *uint32 but %T\n", value)
	case uint16:
		ov, ok := value.(*uint16)
		if ok {
			return ourV == *ov
		}
		log.Printf("D! value not a *uint16 but %T\n", value)
	case byte:
		ov, ok := value.([]byte)
		if ok {
			if len(ov) == 1 {
				return ourV == ov[0]
			}
			log.Printf("D! value not a [1]byte but %T\n", value)
		}
		log.Printf("D! value not a [1]byte but %T\n", value)
	}
	return false
}

// sequenceDirective is a decode directive that is a simple sequentially executed list of other decode directives
type sequenceDirective struct {
	decoders []Directive
	location string
}

func (di *sequenceDirective) Reset() {
	for _, d := range di.decoders {
		d.Reset()
	}

}

func (di *sequenceDirective) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	for i, innerDD := range di.decoders {
		dc.tracef("%s seq %d\n", di.location, i)
		if err := innerDD.execute(buffer, dc); err != nil {
			return err
		}
	}
	return nil
}

// openMetric a decode directive that opens the recording of new fields and tags
type openMetric struct {
	location string
}

func (di *openMetric) Reset() {
	// NOP
}

func (di *openMetric) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	dc.tracef("%s open metric\n", di.location)
	dc.openMetric()
	return nil
}

// closeMetric a decode directive that closes the current open metric
type closeMetric struct {
	location string
}

func (di *closeMetric) Reset() {
	// NOP
}

func (di *closeMetric) execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	dc.tracef("%s close metric\n", di.location)
	dc.closeMetric()
	return nil
}

// DecodeContext provides context for the decoding of a packet and primarily acts
// as a repository for metrics that are collected during the packet decode process
type DecodeContext struct {
	metrics []telegraf.Metric

	// oreMetric is used to capture tags or fields that may be recored before a metric has been openned
	// these fields and tags are then copied into metrics that are then subsequently opened
	preMetric telegraf.Metric
	current   telegraf.Metric
	nano      int
	trace     bool
}

func (dc *DecodeContext) tracef(fmt string, v ...interface{}) {
	if dc.trace {
		log.Printf(fmt, v...)
	}
}

func (dc *DecodeContext) openMetric() {
	m, _ := metric.New("sflow", make(map[string]string), make(map[string]interface{}), time.Now().Add(time.Duration(dc.nano)))
	dc.nano++
	// make sure to copy any fields and tags that were capture prior to the metric being openned
	for t, v := range dc.preMetric.Tags() {
		m.AddTag(t, v)
	}
	for f, v := range dc.preMetric.Fields() {
		m.AddField(f, v)
	}
	dc.current = m
}

func (dc *DecodeContext) closeMetric() {
	if dc.current != nil {
		dc.metrics = append(dc.metrics, dc.current)
	}
	dc.current = nil
}

func (dc *DecodeContext) currentMetric() telegraf.Metric {
	if dc.current == nil {
		return dc.preMetric
	}
	return dc.current
}

// Decode initiates the decoding of the supplied buffer according to the root decode directive that is provided
func (dc *DecodeContext) Decode(dd Directive, buffer *bytes.Buffer) error {
	return dd.execute(buffer, dc)
}

// GetMetrics answers the metrics that have been collected during the packet decode
func (dc *DecodeContext) GetMetrics() []telegraf.Metric {
	return dc.metrics
}
