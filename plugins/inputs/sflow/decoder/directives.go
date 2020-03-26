package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Directive is a Decode Directive, the basic building block of a decoder
type Directive interface {

	// Execute performs the function of the decode directive. If DecodeContext is nil then the
	// ask is to check that a subsequent execution (with non nill DecodeContext) is expted to work.
	Execute(*bytes.Buffer, *DecodeContext) error
}

type IterOption struct {
	EOFTerminateIter                   bool
	RemainingToGreaterEqualOrTerminate uint32
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
	Iter(maxIterations uint32, dd Directive, iterOptions ...IterOption) ValueDirective

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

	iterOption IterOption
}

func valueToString(in interface{}) string {
	switch v := in.(type) {
	case *uint16:
		return fmt.Sprintf("%d", *v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case *uint32:
		return fmt.Sprintf("%d", *v)
	case uint32:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", in)
	}
}

func (dd *valueDirective) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	if dd.reference == nil && !dd.noDecode {
		if e := binary.Read(buffer, binary.BigEndian, dd.value); e != nil {
			return e
		}
	}

	// Switch downstream?
	if dd.cases != nil && len(dd.cases) > 0 {
		for _, c := range dd.cases {
			if c.Equals(dd.value) {
				return c.Execute(buffer, dc)
			}
		}
		switch v := dd.value.(type) {
		case *uint32:
			return fmt.Errorf("(%T).Switch,unmatched case %d", v, *v)
		case *uint16:
			return fmt.Errorf("(%T).Switch,unmatched case %d", v, *v)
		default:
			return fmt.Errorf("(%T).Switch,unmatched case %v", dd.value, dd.value)
		}
	}

	// Iter downstream?
	if dd.iter != nil {
		fn := func(id interface{}) error {
			if dd.iterOption.RemainingToGreaterEqualOrTerminate > 0 && uint32(buffer.Len()) < dd.iterOption.RemainingToGreaterEqualOrTerminate {
				return nil
			}
			if dd.iterOption.EOFTerminateIter && buffer.Len() == 0 {
				return nil
			}
			if e := dd.iter.Execute(buffer, dc); e != nil {
				return e
			}
			return nil
		}
		switch v := dd.value.(type) {
		case *uint32:
			if *v > dd.maxIterations {
				return fmt.Errorf("iter exceeds configured max - value %d, limit %d", *v, dd.maxIterations)
			}
			for i := uint32(0); i < *v; i++ {
				if e := fn(i); e != nil {
					return e
				}
			}
		case *uint16:
			if *v > uint16(dd.maxIterations) {
				return fmt.Errorf("iter exceeds configured max - value %d, limit %d", *v, dd.maxIterations)
			}
			for i := uint16(0); i < *v; i++ {
				if e := fn(i); e != nil {
					return e
				}
			}
		default:
			// Can't actually get here if .Iter method check types (and it does)
			return fmt.Errorf("(%T).Iter, cannot iterator over this type", dd.value)
		}
	}

	// Encapsualted downstream>
	if dd.encapsulated != nil {
		switch v := dd.value.(type) {
		case *uint32:
			if *v > dd.maxEncapsulation {
				return fmt.Errorf("encap exceeds configured max - value %d, limit %d", *v, dd.maxEncapsulation)
			}
			return dd.encapsulated.Execute(bytes.NewBuffer(buffer.Next(int(*v))), dc)
		case *uint16:
			if *v > uint16(dd.maxEncapsulation) {
				return fmt.Errorf("encap exceeds configured max - value %d, limit %d", *v, dd.maxEncapsulation)
			}
			return dd.encapsulated.Execute(bytes.NewBuffer(buffer.Next(int(*v))), dc)
		}
	}

	// Perform the attached operations
	for _, op := range dd.ops {
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
		panic(fmt.Sprintf("already have switch cases assigned, cannot assign %s", change))
	}
	if dd.iter != nil {
		panic(fmt.Sprintf("already have iter assigned, cannot assign %s", change))
	}
	if dd.encapsulated != nil {
		panic(fmt.Sprintf("already have encap assigned, cannot assign %s @", change))
	}
	if checkDOs && dd.ops != nil && len(dd.ops) > 0 {
		panic(fmt.Sprintf("already have do assigned, cannot assign %s", change))
	}
}

func (dd *valueDirective) Switch(paths ...CaseValueDirective) ValueDirective {
	dd.panickIfNotBlackCanvas("new switch", true)
	dd.cases = paths
	return dd
}

func (dd *valueDirective) Iter(maxIterations uint32, iter Directive, iterOptions ...IterOption) ValueDirective {
	dd.panickIfNotBlackCanvas("new iter", true)
	switch dd.value.(type) {
	case *uint32:
	case *uint16:
	default:
		panic(fmt.Sprintf("cannot iterate a %T", dd.value))
	}

	dd.iter = iter
	dd.maxIterations = maxIterations
	for _, io := range iterOptions {
		dd.iterOption = io
	}
	return dd
}

func (dd *valueDirective) Encapsulated(maxSize uint32, encapsulated Directive) ValueDirective {
	dd.panickIfNotBlackCanvas("new encapsulated", true)
	switch dd.value.(type) {
	case *uint32:
	case *uint16:
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

// errorDirective a decode directive that reports an error
type errorDirective struct {
	Directive
}

func (dd *errorDirective) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	return fmt.Errorf("Error Directive")
}

// CaseValueDirective is a decode directive that also has a switch/case test
type CaseValueDirective interface {
	Directive
	Equals(interface{}) bool
}

type caseValueDirective struct {
	caseValue interface{}
	isDefault bool
	equalsDd  Directive
}

func (dd *caseValueDirective) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	if dd.equalsDd == nil {
		return nil
	}
	return dd.equalsDd.Execute(buffer, dc)
}

func (dd *caseValueDirective) Equals(value interface{}) bool {
	if dd.isDefault {
		return true
	}
	switch ourV := dd.caseValue.(type) {
	case uint32:
		ov, ok := value.(*uint32)
		if ok {
			return ourV == *ov
		}
	case uint16:
		ov, ok := value.(*uint16)
		if ok {
			return ourV == *ov
		}
	case byte:
		ov, ok := value.([]byte)
		if ok {
			if len(ov) == 1 {
				return ourV == ov[0]
			}
		}
	}
	return false
}

// sequenceDirective is a decode directive that is a simple sequentially executed list of other decode directives
type sequenceDirective struct {
	decoders []Directive
}

func (di *sequenceDirective) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	for _, innerDD := range di.decoders {
		if err := innerDD.Execute(buffer, dc); err != nil {
			return err
		}
	}
	return nil
}

// openMetric a decode directive that opens the recording of new fields and tags
type openMetric struct {
	name string
}

func (di *openMetric) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	dc.openMetric(di.name)
	return nil
}

// closeMetric a decode directive that closes the current open metric
type closeMetric struct {
}

func (di *closeMetric) Execute(buffer *bytes.Buffer, dc *DecodeContext) error {
	dc.closeMetric()
	return nil
}

// DecodeContext provides context for the decoding of a packet and primarily acts
// as a repository for metrics that are collected during the packet decode process
type DecodeContext struct {
	metrics        []telegraf.Metric
	timeHasBeenSet bool

	// oreMetric is used to capture tags or fields that may be recored before a metric has been openned
	// these fields and tags are then copied into metrics that are then subsequently opened
	preMetric telegraf.Metric
	current   telegraf.Metric
	nano      int
}

func (dc *DecodeContext) openMetric(name string) {
	t := dc.preMetric.Time()
	if !dc.timeHasBeenSet {
		t = time.Now().Add(time.Duration(dc.nano))
	}
	m, _ := metric.New(name, make(map[string]string), make(map[string]interface{}), t)
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
	return dd.Execute(buffer, dc)
}

// GetMetrics answers the metrics that have been collected during the packet decode
func (dc *DecodeContext) GetMetrics() []telegraf.Metric {
	return dc.metrics
}

type notifyDirective struct {
	fn func()
}

func (nd *notifyDirective) Execute(_ *bytes.Buffer, dc *DecodeContext) error {
	if dc != nil {
		nd.fn()
	}
	return nil
}
