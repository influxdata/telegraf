package netflow

//go:generate go run ../scripts/netflow/generate-field-decoders.go

import (
	"bytes"
	"fmt"
	"math"
	"net"

	"github.com/influxdata/telegraf/plugins/parsers/network_flow/decoder"
)

const (
	metricName = "netflow"
)

type ftm struct {
	id   uint16
	name string
	decoder.ValueDirective
}

type fieldDefn struct {
	fieldType   uint16
	fieldLength uint16
	// dd is the decoder.Directive that will decode this field type within the data flow set
	dd decoder.Directive
}

type templateDefn struct {
	templateID  uint16
	fields      []fieldDefn
	totalLength uint16
	// dd is the decoder.Directive that will decode data flow sets according to this template
	dd decoder.Directive
}

type templateMapCase struct {
	//od *obsDomain
	//templateMap          map[uint16]*templateDefn
	od                   *uint32
	lastSelectedTemplate *templateDefn
}

func (tmc *templateMapCase) Equals(in interface{}) bool {
	od := obsDomains[*tmc.od]
	if od == nil {
		return false
	}
	switch v := in.(type) {
	case *uint16:
		if t, ok := od.templateMap[*v]; ok {
			tmc.lastSelectedTemplate = t
			return true
		}
	}
	return false
}

func (tmc *templateMapCase) Execute(b *bytes.Buffer, dc *decoder.DecodeContext) error {
	// TODO change Execute signature so that it takes the selector value
	if tmc.lastSelectedTemplate == nil {
		return fmt.Errorf("nil template selected")
	}
	return tmc.lastSelectedTemplate.dd.Execute(b, dc)
}

func (*templateMapCase) Reset() {
	// TODO should probably reset everything in template map
}

// obsDomain holds the template definitions for a particular observtation domain
type obsDomain struct {
	sourceID    uint32
	templateMap map[uint16]*templateDefn
}

var obsDomains map[uint32]*obsDomain

//var templateMap map[uint16]*templateDefn

func bytesToIPStr(b []byte) string {
	return net.IP(b).String()
}

func uintByLenAsF(n string) func(l uint16) decoder.Directive {
	return func(l uint16) decoder.Directive {
		return uintDecoderByLen(l).Do(decoder.AsF(n))
	}
}

func minusFourBytes() *decoder.U16ToU16DOp {
	return decoder.U16ToU16(func(in uint16) uint16 { return in - 4 })
}

func templateFormat(sourceID *uint32) decoder.Directive {

	td := &templateDefn{}

	addTemplateField := func(ft uint16, fl uint16) {
		nf := fieldDefn{fieldType: ft, fieldLength: fl}
		nf.dd = getFieldDecoder(ft, fl)
		/*
			ldd := fieldDecoderMap[ft]
			if ldd != nil {
				nf.dd = ldd(fl)
			}
			if nf.dd == nil {
				// we don't have a field specific decoder
				// use the nop decoder that will consume the bytes and do nothing with them
				nf.dd = decoder.Bytes(int(fl))
			}
		*/
		td.fields = append(td.fields, nf)
		td.totalLength += fl
	}

	completeTemplate := func(templateID uint16) {
		td.templateID = templateID
		fieldDecoders := make([]decoder.Directive, 0, len(td.fields))
		for _, fd := range td.fields {
			fieldDecoders = append(fieldDecoders, fd.dd)
		}
		infinite := uint16(math.MaxUint16)

		td.dd = decoder.Seq(
			decoder.U16().Encapsulated(
				math.MaxUint32,
				decoder.U16Value(&infinite).Iter(
					uint32(infinite),
					decoder.Seq(
						decoder.OpenMetric(metricName),
						decoder.SeqOf(fieldDecoders),
						decoder.CloseMetric(),
					),
					decoder.IterOption{RemainingToGreaterEqualOrTerminate: uint32(td.totalLength - 4)}, // -4 comes from the fact that we have already consumed 4 bytes
				),
			),
		)

		od := obsDomains[*sourceID]
		if od == nil {
			od = &obsDomain{sourceID: *sourceID, templateMap: make(map[uint16]*templateDefn)}
			obsDomains[*sourceID] = od
		}

		//templateMap[templateID] = td
		od.templateMap[templateID] = td
		td = &templateDefn{}
	}

	var templateID uint16
	var fieldType uint16
	var fieldLength uint16
	return decoder.Seq(
		decoder.U16().Do(decoder.Set(&templateID)),
		decoder.U16().Iter(
			math.MaxUint32,
			decoder.Seq(
				decoder.U16().Do(decoder.Set(&fieldType)),
				decoder.U16().Do(decoder.Set(&fieldLength)),
				decoder.Notify(func() {
					addTemplateField(fieldType, fieldLength)
				}),
			),
		),
		decoder.Notify(func() {
			completeTemplate(templateID)
		}),
	)
}

func templateFlowSet(sourceID *uint32) decoder.Directive {
	flowSetLength := new(uint16)
	maxUint16 := uint16(math.MaxUint16)
	return decoder.Seq(
		decoder.U16().Do(minusFourBytes().Set(flowSetLength)),
		decoder.U16Value(flowSetLength).Encapsulated(
			math.MaxUint32,
			decoder.U16Value(&maxUint16).Iter(
				math.MaxUint16,
				templateFormat(sourceID),
				decoder.IterOption{EOFTerminateIter: true},
			),
		),
	)
}

func flowSetFormat(sourceID *uint32) decoder.Directive {
	flowSetLength := new(uint16)
	optionsTemplateSet := decoder.Seq(
		decoder.U16().Do(decoder.U16ToU16(
			func(in uint16) uint16 { return in - 4 }).Set(flowSetLength)), // -2 * UI16
		decoder.U16Value(flowSetLength).Encapsulated(
			math.MaxUint32,
			nil, // NOP
		),
	)

	var flowSetID uint16
	var dataSetLength uint16
	return decoder.Seq(
		decoder.U16().Do(decoder.Set(&flowSetID)),
		decoder.U16Value(&flowSetID).Switch(
			decoder.Case(uint16(0), templateFlowSet(sourceID)),
			decoder.Case(uint16(1), optionsTemplateSet),

			// use the template map to selected a decoder, if we have a template for the flowSetId
			&templateMapCase{od: sourceID}, //{, // templateMap: templateMap},

			// If it gets this far then it is a data set with a template we don't know of
			// just ignore the bytes of that data set then
			decoder.DefaultCase(decoder.Seq(
				decoder.U16().Do(decoder.U16ToU16(
					func(in uint16) uint16 { return in - 4 }).Set(&dataSetLength)), // -2 * UI16
				decoder.U16Value(&dataSetLength).Encapsulated(
					math.MaxUint32,
					decoder.Nop(),
				),
			)),
		),
	)
}

// NewV10Decoder answers a decoder.Directive that will decode a NetflowV10 (and v9 via backwards conpatqbility) packet
// into Influx tags and fields
func NewV10Decoder() decoder.Directive {
	var count interface{}
	var sourceID uint32
	return decoder.Seq(
		decoder.U16().Do(decoder.U16Assert(func(in uint16) bool { return in == 10 || in == 9 }, "cannot support version %d only versions 9 & 10")),
		decoder.U16().Ref(&count),
		decoder.U32(),
		decoder.U32().Do(decoder.AsTimestamp()),
		decoder.U32(),
		decoder.U32().Do(decoder.Set(&sourceID).AsT("sourceID")),
		decoder.Ref(count).Iter(
			math.MaxInt32,
			flowSetFormat(&sourceID),
			decoder.IterOption{EOFTerminateIter: true},
		),
	)
}
