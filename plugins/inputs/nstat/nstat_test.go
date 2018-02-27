package nstat

import "testing"

func TestLoadUglyTable(t *testing.T) {
	uglyStr := `IpExt: InNoRoutes InTruncatedPkts InMcastPkts InCEPkts
	IpExt: 332 433718 0 2660494435`
	parsed := map[string]interface{}{
		"IpExtInNoRoutes":      int64(332),
		"IpExtInTruncatedPkts": int64(433718),
		"IpExtInMcastPkts":     int64(0),
		"IpExtInCEPkts":        int64(2660494435),
	}

	got, err := loadUglyTable([]byte(uglyStr), true)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) == 0 {
		t.Fatalf("want %+v, got %+v", parsed, got)
	}

	for key := range parsed {
		if parsed[key].(int64) != got[key].(int64) {
			t.Fatalf("want %+v, got %+v", parsed[key], got[key])
		}
	}
}

func TestLoadGoodTable(t *testing.T) {
	goodStr := `Ip6InReceives                   	11707
				Ip6InTooBigErrors               	0
				Ip6InDelivers                   	62
				Ip6InMcastOctets                	1242966`

	parsed := map[string]interface{}{
		"Ip6InReceives":     int64(11707),
		"Ip6InTooBigErrors": int64(0),
		"Ip6InDelivers":     int64(62),
		"Ip6InMcastOctets":  int64(1242966),
	}
	got, err := loadGoodTable([]byte(goodStr), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatalf("want %+v, got %+v", parsed, got)
	}

	for key := range parsed {
		if parsed[key].(int64) != got[key].(int64) {
			t.Fatalf("want %+v, got %+v", parsed[key], got[key])
		}
	}
}
