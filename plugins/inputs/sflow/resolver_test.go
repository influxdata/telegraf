package sflow

import (
	"testing"
)

func Test_goodDNSProcessor(t *testing.T) {
	str := map[string]string{
		"fi-es-he6-z4-e02-qfx03-dev.netdevice.nesc.nokia.net":     "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev-em0-0.transit.nesc.nokia.net": "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net":               "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
	}

	p := newDNSProcessor(`(.*)(?:(?:-e.[0-9]-[0-9]\.transit)|(?:\.netdevice))(.*)`)

	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor(`s/(.*)(?:(?:-e.[0-9]-[0-9]\.transit)|(?:\.netdevice))(.*)/$1$2`)
	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor("")
	for k := range str {
		transformed := p.transform(k)
		if transformed != k {
			t.Fatalf("actual %s != expected %s", transformed, k)
		}
	}
}

func Test_badDNSProcessor(t *testing.T) {
	str := map[string]string{
		"fi-es-he6-z4-e02-qfx03-dev.netdevice.nesc.nokia.net":     "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
		"fi-es-he6-z4-e02-qfx03-dev-em0-0.transit.nesc.nokia.net": "fi-es-he6-z4-e02-qfx03-dev.nesc.nokia.net",
	}

	p := newDNSProcessor(`(.*)`)

	for k, v := range str {
		transformed := p.transform(k)
		if transformed == v {
			t.Fatalf("actual %s == expected %s and that is a surprise", transformed, v)
		}
	}
}

func Test_deanDNSProcessor(t *testing.T) {
	str := map[string]string{
		"192.168.0.49": "deans-laptop",
		"192.168.0.50": "192.168.0.50",
	}

	p := newDNSProcessor(`s/192.168.0.49/deans-laptop`)
	for k, v := range str {
		transformed := p.transform(k)
		if transformed != v {
			t.Fatalf("actual %s != expected %s", transformed, v)
		}
	}

	p = newDNSProcessor("")
	for k := range str {
		transformed := p.transform(k)
		if transformed != k {
			t.Fatalf("actual %s != expected %s", transformed, k)
		}
	}
}
