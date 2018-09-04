package dns

import (
	"encoding/hex"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestCompressLength(t *testing.T) {
	m := new(Msg)
	m.SetQuestion("miek.nl", TypeMX)
	ul := m.Len()
	m.Compress = true
	if ul != m.Len() {
		t.Fatalf("should be equal")
	}
}

// Does the predicted length match final packed length?
func TestMsgCompressLength(t *testing.T) {
	makeMsg := func(question string, ans, ns, e []RR) *Msg {
		msg := new(Msg)
		msg.SetQuestion(Fqdn(question), TypeANY)
		msg.Answer = append(msg.Answer, ans...)
		msg.Ns = append(msg.Ns, ns...)
		msg.Extra = append(msg.Extra, e...)
		msg.Compress = true
		return msg
	}

	name1 := "12345678901234567890123456789012345.12345678.123."
	rrA := testRR(name1 + " 3600 IN A 192.0.2.1")
	rrMx := testRR(name1 + " 3600 IN MX 10 " + name1)
	tests := []*Msg{
		makeMsg(name1, []RR{rrA}, nil, nil),
		makeMsg(name1, []RR{rrMx, rrMx}, nil, nil)}

	for _, msg := range tests {
		predicted := msg.Len()
		buf, err := msg.Pack()
		if err != nil {
			t.Error(err)
		}
		if predicted < len(buf) {
			t.Errorf("predicted compressed length is wrong: predicted %s (len=%d) %d, actual %d",
				msg.Question[0].Name, len(msg.Answer), predicted, len(buf))
		}
	}
}

func TestMsgLength(t *testing.T) {
	makeMsg := func(question string, ans, ns, e []RR) *Msg {
		msg := new(Msg)
		msg.Compress = true
		msg.SetQuestion(Fqdn(question), TypeANY)
		msg.Answer = append(msg.Answer, ans...)
		msg.Ns = append(msg.Ns, ns...)
		msg.Extra = append(msg.Extra, e...)
		return msg
	}

	name1 := "12345678901234567890123456789012345.12345678.123."
	rrA := testRR(name1 + " 3600 IN A 192.0.2.1")
	rrMx := testRR(name1 + " 3600 IN MX 10 " + name1)
	tests := []*Msg{
		makeMsg(name1, []RR{rrA}, nil, nil),
		makeMsg(name1, []RR{rrMx, rrMx}, nil, nil)}

	for _, msg := range tests {
		predicted := msg.Len()
		buf, err := msg.Pack()
		if err != nil {
			t.Error(err)
		}
		if predicted < len(buf) {
			t.Errorf("predicted length is wrong: predicted %s (len=%d), actual %d",
				msg.Question[0].Name, predicted, len(buf))
		}
	}
}

func TestCompressionLenHelper(t *testing.T) {
	c := make(map[string]int)
	compressionLenHelper(c, "example.com", 12)
	if c["example.com"] != 12 {
		t.Errorf("bad %d", c["example.com"])
	}
	if c["com"] != 20 {
		t.Errorf("bad %d", c["com"])
	}

	// Test boundaries
	c = make(map[string]int)
	// foo label starts at 16379
	// com label starts at 16384
	compressionLenHelper(c, "foo.com", 16379)
	if c["foo.com"] != 16379 {
		t.Errorf("bad %d", c["foo.com"])
	}
	// com label is accessible
	if c["com"] != 16383 {
		t.Errorf("bad %d", c["com"])
	}

	c = make(map[string]int)
	// foo label starts at 16379
	// com label starts at 16385 => outside range
	compressionLenHelper(c, "foo.com", 16380)
	if c["foo.com"] != 16380 {
		t.Errorf("bad %d", c["foo.com"])
	}
	// com label is NOT accessible
	if c["com"] != 0 {
		t.Errorf("bad %d", c["com"])
	}

	c = make(map[string]int)
	compressionLenHelper(c, "example.com", 16375)
	if c["example.com"] != 16375 {
		t.Errorf("bad %d", c["example.com"])
	}
	// com starts AFTER 16384
	if c["com"] != 16383 {
		t.Errorf("bad %d", c["com"])
	}

	c = make(map[string]int)
	compressionLenHelper(c, "example.com", 16376)
	if c["example.com"] != 16376 {
		t.Errorf("bad %d", c["example.com"])
	}
	// com starts AFTER 16384
	if c["com"] != 0 {
		t.Errorf("bad %d", c["com"])
	}
}

func TestCompressionLenSearch(t *testing.T) {
	c := make(map[string]int)
	compressed, ok, fullSize := compressionLenSearch(c, "a.b.org.")
	if compressed != 0 || ok || fullSize != 14 {
		panic(fmt.Errorf("Failed: compressed:=%d, ok:=%v, fullSize:=%d", compressed, ok, fullSize))
	}
	c["org."] = 3
	compressed, ok, fullSize = compressionLenSearch(c, "a.b.org.")
	if compressed != 4 || !ok || fullSize != 8 {
		panic(fmt.Errorf("Failed: compressed:=%d, ok:=%v, fullSize:=%d", compressed, ok, fullSize))
	}
	c["b.org."] = 5
	compressed, ok, fullSize = compressionLenSearch(c, "a.b.org.")
	if compressed != 6 || !ok || fullSize != 4 {
		panic(fmt.Errorf("Failed: compressed:=%d, ok:=%v, fullSize:=%d", compressed, ok, fullSize))
	}
	// Not found long compression
	c["x.b.org."] = 5
	compressed, ok, fullSize = compressionLenSearch(c, "a.b.org.")
	if compressed != 6 || !ok || fullSize != 4 {
		panic(fmt.Errorf("Failed: compressed:=%d, ok:=%v, fullSize:=%d", compressed, ok, fullSize))
	}
	// Found long compression
	c["a.b.org."] = 5
	compressed, ok, fullSize = compressionLenSearch(c, "a.b.org.")
	if compressed != 8 || !ok || fullSize != 0 {
		panic(fmt.Errorf("Failed: compressed:=%d, ok:=%v, fullSize:=%d", compressed, ok, fullSize))
	}
}

func TestMsgLength2(t *testing.T) {
	// Serialized replies
	var testMessages = []string{
		// google.com. IN A?
		"064e81800001000b0004000506676f6f676c6503636f6d0000010001c00c00010001000000050004adc22986c00c00010001000000050004adc22987c00c00010001000000050004adc22988c00c00010001000000050004adc22989c00c00010001000000050004adc2298ec00c00010001000000050004adc22980c00c00010001000000050004adc22981c00c00010001000000050004adc22982c00c00010001000000050004adc22983c00c00010001000000050004adc22984c00c00010001000000050004adc22985c00c00020001000000050006036e7331c00cc00c00020001000000050006036e7332c00cc00c00020001000000050006036e7333c00cc00c00020001000000050006036e7334c00cc0d800010001000000050004d8ef200ac0ea00010001000000050004d8ef220ac0fc00010001000000050004d8ef240ac10e00010001000000050004d8ef260a0000290500000000050000",
		// amazon.com. IN A? (reply has no EDNS0 record)
		// TODO(miek): this one is off-by-one, need to find out why
		//"6de1818000010004000a000806616d617a6f6e03636f6d0000010001c00c000100010000000500044815c2d4c00c000100010000000500044815d7e8c00c00010001000000050004b02062a6c00c00010001000000050004cdfbf236c00c000200010000000500140570646e733408756c747261646e73036f726700c00c000200010000000500150570646e733508756c747261646e7304696e666f00c00c000200010000000500160570646e733608756c747261646e7302636f02756b00c00c00020001000000050014036e7331037033310664796e656374036e657400c00c00020001000000050006036e7332c0cfc00c00020001000000050006036e7333c0cfc00c00020001000000050006036e7334c0cfc00c000200010000000500110570646e733108756c747261646e73c0dac00c000200010000000500080570646e7332c127c00c000200010000000500080570646e7333c06ec0cb00010001000000050004d04e461fc0eb00010001000000050004cc0dfa1fc0fd00010001000000050004d04e471fc10f00010001000000050004cc0dfb1fc12100010001000000050004cc4a6c01c121001c000100000005001020010502f3ff00000000000000000001c13e00010001000000050004cc4a6d01c13e001c0001000000050010261000a1101400000000000000000001",
		// yahoo.com. IN A?
		"fc2d81800001000300070008057961686f6f03636f6d0000010001c00c00010001000000050004628afd6dc00c00010001000000050004628bb718c00c00010001000000050004cebe242dc00c00020001000000050006036e7336c00cc00c00020001000000050006036e7338c00cc00c00020001000000050006036e7331c00cc00c00020001000000050006036e7332c00cc00c00020001000000050006036e7333c00cc00c00020001000000050006036e7334c00cc00c00020001000000050006036e7335c00cc07b0001000100000005000444b48310c08d00010001000000050004448eff10c09f00010001000000050004cb54dd35c0b100010001000000050004628a0b9dc0c30001000100000005000477a0f77cc05700010001000000050004ca2bdfaac06900010001000000050004caa568160000290500000000050000",
		// microsoft.com. IN A?
		"f4368180000100020005000b096d6963726f736f667403636f6d0000010001c00c0001000100000005000440040b25c00c0001000100000005000441373ac9c00c0002000100000005000e036e7331046d736674036e657400c00c00020001000000050006036e7332c04fc00c00020001000000050006036e7333c04fc00c00020001000000050006036e7334c04fc00c00020001000000050006036e7335c04fc04b000100010000000500044137253ec04b001c00010000000500102a010111200500000000000000010001c0650001000100000005000440043badc065001c00010000000500102a010111200600060000000000010001c07700010001000000050004d5c7b435c077001c00010000000500102a010111202000000000000000010001c08900010001000000050004cf2e4bfec089001c00010000000500102404f800200300000000000000010001c09b000100010000000500044137e28cc09b001c00010000000500102a010111200f000100000000000100010000290500000000050000",
		// google.com. IN MX?
		"724b8180000100050004000b06676f6f676c6503636f6d00000f0001c00c000f000100000005000c000a056173706d78016cc00cc00c000f0001000000050009001404616c7431c02ac00c000f0001000000050009001e04616c7432c02ac00c000f0001000000050009002804616c7433c02ac00c000f0001000000050009003204616c7434c02ac00c00020001000000050006036e7332c00cc00c00020001000000050006036e7333c00cc00c00020001000000050006036e7334c00cc00c00020001000000050006036e7331c00cc02a00010001000000050004adc2421bc02a001c00010000000500102a00145040080c01000000000000001bc04200010001000000050004adc2461bc05700010001000000050004adc2451bc06c000100010000000500044a7d8f1bc081000100010000000500044a7d191bc0ca00010001000000050004d8ef200ac09400010001000000050004d8ef220ac0a600010001000000050004d8ef240ac0b800010001000000050004d8ef260a0000290500000000050000",
		// reddit.com. IN A?
		"12b98180000100080000000c0672656464697403636f6d0000020001c00c0002000100000005000f046175733204616b616d036e657400c00c000200010000000500070475736534c02dc00c000200010000000500070475737733c02dc00c000200010000000500070475737735c02dc00c00020001000000050008056173696131c02dc00c00020001000000050008056173696139c02dc00c00020001000000050008056e73312d31c02dc00c0002000100000005000a076e73312d313935c02dc02800010001000000050004c30a242ec04300010001000000050004451f1d39c05600010001000000050004451f3bc7c0690001000100000005000460073240c07c000100010000000500046007fb81c090000100010000000500047c283484c090001c00010000000500102a0226f0006700000000000000000064c0a400010001000000050004c16c5b01c0a4001c000100000005001026001401000200000000000000000001c0b800010001000000050004c16c5bc3c0b8001c0001000000050010260014010002000000000000000000c30000290500000000050000",
	}

	for i, hexData := range testMessages {
		// we won't fail the decoding of the hex
		input, _ := hex.DecodeString(hexData)

		m := new(Msg)
		m.Unpack(input)
		m.Compress = true
		lenComp := m.Len()
		b, _ := m.Pack()
		pacComp := len(b)
		m.Compress = false
		lenUnComp := m.Len()
		b, _ = m.Pack()
		pacUnComp := len(b)
		if pacComp+1 != lenComp {
			t.Errorf("msg.Len(compressed)=%d actual=%d for test %d", lenComp, pacComp, i)
		}
		if pacUnComp+1 != lenUnComp {
			t.Errorf("msg.Len(uncompressed)=%d actual=%d for test %d", lenUnComp, pacUnComp, i)
		}
	}
}

func TestMsgLengthCompressionMalformed(t *testing.T) {
	// SOA with empty hostmaster, which is illegal
	soa := &SOA{Hdr: RR_Header{Name: ".", Rrtype: TypeSOA, Class: ClassINET, Ttl: 12345},
		Ns:      ".",
		Mbox:    "",
		Serial:  0,
		Refresh: 28800,
		Retry:   7200,
		Expire:  604800,
		Minttl:  60}
	m := new(Msg)
	m.Compress = true
	m.Ns = []RR{soa}
	m.Len() // Should not crash.
}

func TestMsgCompressLength2(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion(Fqdn("bliep."), TypeANY)
	msg.Answer = append(msg.Answer, &SRV{Hdr: RR_Header{Name: "blaat.", Rrtype: 0x21, Class: 0x1, Ttl: 0x3c}, Port: 0x4c57, Target: "foo.bar."})
	msg.Extra = append(msg.Extra, &A{Hdr: RR_Header{Name: "foo.bar.", Rrtype: 0x1, Class: 0x1, Ttl: 0x3c}, A: net.IP{0xac, 0x11, 0x0, 0x3}})
	predicted := msg.Len()
	buf, err := msg.Pack()
	if err != nil {
		t.Error(err)
	}
	if predicted != len(buf) {
		t.Errorf("predicted compressed length is wrong: predicted %s (len=%d) %d, actual %d",
			msg.Question[0].Name, len(msg.Answer), predicted, len(buf))
	}
}

func TestMsgCompressLengthLargeRecords(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion("my.service.acme.", TypeSRV)
	j := 1
	for i := 0; i < 250; i++ {
		target := fmt.Sprintf("host-redis-1-%d.test.acme.com.node.dc1.consul.", i)
		msg.Answer = append(msg.Answer, &SRV{Hdr: RR_Header{Name: "redis.service.consul.", Class: 1, Rrtype: TypeSRV, Ttl: 0x3c}, Port: 0x4c57, Target: target})
		msg.Extra = append(msg.Extra, &CNAME{Hdr: RR_Header{Name: target, Class: 1, Rrtype: TypeCNAME, Ttl: 0x3c}, Target: fmt.Sprintf("fx.168.%d.%d.", j, i)})
	}
	predicted := msg.Len()
	buf, err := msg.Pack()
	if err != nil {
		t.Error(err)
	}
	if predicted != len(buf) {
		t.Fatalf("predicted compressed length is wrong: predicted %s (len=%d) %d, actual %d", msg.Question[0].Name, len(msg.Answer), predicted, len(buf))
	}
}

func TestCompareCompressionMapsForANY(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion("a.service.acme.", TypeANY)
	// Be sure to have more than 14bits
	for i := 0; i < 2000; i++ {
		target := fmt.Sprintf("host.app-%d.x%d.test.acme.", i%250, i)
		msg.Answer = append(msg.Answer, &AAAA{Hdr: RR_Header{Name: target, Rrtype: TypeAAAA, Class: ClassINET, Ttl: 0x3c}, AAAA: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(i / 255), byte(i % 255)}})
		msg.Answer = append(msg.Answer, &A{Hdr: RR_Header{Name: target, Rrtype: TypeA, Class: ClassINET, Ttl: 0x3c}, A: net.IP{127, 0, byte(i / 255), byte(i % 255)}})
		if msg.Len() > 16384 {
			break
		}
	}
	for labelSize := 0; labelSize < 63; labelSize++ {
		msg.SetQuestion(fmt.Sprintf("a%s.service.acme.", strings.Repeat("x", labelSize)), TypeANY)

		compressionFake := make(map[string]int)
		lenFake := compressedLenWithCompressionMap(msg, compressionFake)

		compressionReal := make(map[string]int)
		buf, err := msg.packBufferWithCompressionMap(nil, compressionReal)
		if err != nil {
			t.Fatal(err)
		}
		if lenFake != len(buf) {
			t.Fatalf("padding= %d ; Predicted len := %d != real:= %d", labelSize, lenFake, len(buf))
		}
		if !reflect.DeepEqual(compressionFake, compressionReal) {
			t.Fatalf("padding= %d ; Fake Compression Map != Real Compression Map\n*** Real:= %v\n\n***Fake:= %v", labelSize, compressionReal, compressionFake)
		}
	}
}

func TestCompareCompressionMapsForSRV(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion("a.service.acme.", TypeSRV)
	// Be sure to have more than 14bits
	for i := 0; i < 2000; i++ {
		target := fmt.Sprintf("host.app-%d.x%d.test.acme.", i%250, i)
		msg.Answer = append(msg.Answer, &SRV{Hdr: RR_Header{Name: "redis.service.consul.", Class: ClassINET, Rrtype: TypeSRV, Ttl: 0x3c}, Port: 0x4c57, Target: target})
		msg.Extra = append(msg.Extra, &A{Hdr: RR_Header{Name: target, Rrtype: TypeA, Class: ClassINET, Ttl: 0x3c}, A: net.IP{127, 0, byte(i / 255), byte(i % 255)}})
		if msg.Len() > 16384 {
			break
		}
	}
	for labelSize := 0; labelSize < 63; labelSize++ {
		msg.SetQuestion(fmt.Sprintf("a%s.service.acme.", strings.Repeat("x", labelSize)), TypeAAAA)

		compressionFake := make(map[string]int)
		lenFake := compressedLenWithCompressionMap(msg, compressionFake)

		compressionReal := make(map[string]int)
		buf, err := msg.packBufferWithCompressionMap(nil, compressionReal)
		if err != nil {
			t.Fatal(err)
		}
		if lenFake != len(buf) {
			t.Fatalf("padding= %d ; Predicted len := %d != real:= %d", labelSize, lenFake, len(buf))
		}
		if !reflect.DeepEqual(compressionFake, compressionReal) {
			t.Fatalf("padding= %d ; Fake Compression Map != Real Compression Map\n*** Real:= %v\n\n***Fake:= %v", labelSize, compressionReal, compressionFake)
		}
	}
}

func TestMsgCompressLengthLargeRecordsWithPaddingPermutation(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion("my.service.acme.", TypeSRV)

	for i := 0; i < 250; i++ {
		target := fmt.Sprintf("host-redis-x-%d.test.acme.com.node.dc1.consul.", i)
		msg.Answer = append(msg.Answer, &SRV{Hdr: RR_Header{Name: "redis.service.consul.", Class: 1, Rrtype: TypeSRV, Ttl: 0x3c}, Port: 0x4c57, Target: target})
		msg.Extra = append(msg.Extra, &CNAME{Hdr: RR_Header{Name: target, Class: ClassINET, Rrtype: TypeCNAME, Ttl: 0x3c}, Target: fmt.Sprintf("fx.168.x.%d.", i)})
	}
	for labelSize := 1; labelSize < 63; labelSize++ {
		msg.SetQuestion(fmt.Sprintf("my.%s.service.acme.", strings.Repeat("x", labelSize)), TypeSRV)
		predicted := msg.Len()
		buf, err := msg.Pack()
		if err != nil {
			t.Error(err)
		}
		if predicted != len(buf) {
			t.Fatalf("padding= %d ; predicted compressed length is wrong: predicted %s (len=%d) %d, actual %d", labelSize, msg.Question[0].Name, len(msg.Answer), predicted, len(buf))
		}
	}
}

func TestMsgCompressLengthLargeRecordsAllValues(t *testing.T) {
	msg := new(Msg)
	msg.Compress = true
	msg.SetQuestion("redis.service.consul.", TypeSRV)
	for i := 0; i < 900; i++ {
		target := fmt.Sprintf("host-redis-%d-%d.test.acme.com.node.dc1.consul.", i/256, i%256)
		msg.Answer = append(msg.Answer, &SRV{Hdr: RR_Header{Name: "redis.service.consul.", Class: 1, Rrtype: TypeSRV, Ttl: 0x3c}, Port: 0x4c57, Target: target})
		msg.Extra = append(msg.Extra, &CNAME{Hdr: RR_Header{Name: target, Class: ClassINET, Rrtype: TypeCNAME, Ttl: 0x3c}, Target: fmt.Sprintf("fx.168.%d.%d.", i/256, i%256)})
		predicted := msg.Len()
		buf, err := msg.Pack()
		if err != nil {
			t.Error(err)
		}
		if predicted != len(buf) {
			t.Fatalf("predicted compressed length is wrong for %d records: predicted %s (len=%d) %d, actual %d", i, msg.Question[0].Name, len(msg.Answer), predicted, len(buf))
		}
	}
}
