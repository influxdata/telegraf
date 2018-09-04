package stringutil_test

import (
	"reflect"
	"testing"

	"github.com/naoina/go-stringutil"
)

var commonTestCasesForToUpperCamelCase = []struct {
	input, expect string
}{
	{"", ""},
	{"thequickbrownfoxoverthelazydog", "Thequickbrownfoxoverthelazydog"},
	{"thequickbrownfoxoverthelazydoG", "ThequickbrownfoxoverthelazydoG"},
	{"thequickbrownfoxoverthelazydo_g", "ThequickbrownfoxoverthelazydoG"},
	{"TheQuickBrownFoxJumpsOverTheLazyDog", "TheQuickBrownFoxJumpsOverTheLazyDog"},
	{"the_quick_brown_fox_jumps_over_the_lazy_dog", "TheQuickBrownFoxJumpsOverTheLazyDog"},
	{"the_Quick_Brown_Fox_Jumps_Over_The_Lazy_Dog", "TheQuickBrownFoxJumpsOverTheLazyDog"},
	{"api_server", "APIServer"},
	{"a_t_api", "ATAPI"},
	{"atapi", "Atapi"},
	{"web_ui", "WebUI"},
	{"api", "API"},
	{"ascii", "ASCII"},
	{"cpu", "CPU"},
	{"csrf", "CSRF"},
	{"css", "CSS"},
	{"dns", "DNS"},
	{"eof", "EOF"},
	{"guid", "GUID"},
	{"html", "HTML"},
	{"http", "HTTP"},
	{"https", "HTTPS"},
	{"id", "ID"},
	{"ip", "IP"},
	{"json", "JSON"},
	{"lhs", "LHS"},
	{"qps", "QPS"},
	{"ram", "RAM"},
	{"rhs", "RHS"},
	{"rpc", "RPC"},
	{"sla", "SLA"},
	{"smtp", "SMTP"},
	{"sql", "SQL"},
	{"ssh", "SSH"},
	{"tcp", "TCP"},
	{"tls", "TLS"},
	{"ttl", "TTL"},
	{"udp", "UDP"},
	{"ui", "UI"},
	{"uid", "UID"},
	{"uuid", "UUID"},
	{"uri", "URI"},
	{"url", "URL"},
	{"utf8", "UTF8"},
	{"vm", "VM"},
	{"xml", "XML"},
	{"xsrf", "XSRF"},
	{"xss", "XSS"},
}

func TestToUpperCamelCase(t *testing.T) {
	for _, v := range append(commonTestCasesForToUpperCamelCase, []struct {
		input, expect string
	}{
		{"ｔｈｅ_ｑｕｉｃｋ_ｂｒｏｗｎ_ｆｏｘ_ｏｖｅｒ_ｔｈｅ_ｌａｚｙ_ｄｏｇ", "ＴｈｅＱｕｉｃｋＢｒｏｗｎＦｏｘＯｖｅｒＴｈｅＬａｚｙＤｏｇ"},
	}...) {
		actual := stringutil.ToUpperCamelCase(v.input)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`stringutil.ToUpperCamelCase(%#v) => %#v; want %#v`, v.input, actual, expect)
		}
	}
}

func TestToUpperCamelCaseASCII(t *testing.T) {
	for _, v := range commonTestCasesForToUpperCamelCase {
		actual := stringutil.ToUpperCamelCaseASCII(v.input)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`stringutil.ToUpperCamelCaseASCII(%#v) => %#v; want %#v`, v.input, actual, expect)
		}
	}
}

var commonTestCasesForToSnakeCase = []struct {
	input, expect string
}{
	{"", ""},
	{"thequickbrownfoxjumpsoverthelazydog", "thequickbrownfoxjumpsoverthelazydog"},
	{"Thequickbrownfoxjumpsoverthelazydog", "thequickbrownfoxjumpsoverthelazydog"},
	{"ThequickbrownfoxjumpsoverthelazydoG", "thequickbrownfoxjumpsoverthelazydo_g"},
	{"TheQuickBrownFoxJumpsOverTheLazyDog", "the_quick_brown_fox_jumps_over_the_lazy_dog"},
	{"the_quick_brown_fox_jumps_over_the_lazy_dog", "the_quick_brown_fox_jumps_over_the_lazy_dog"},
	{"APIServer", "api_server"},
	{"ATAPI", "a_t_api"},
	{"Atapi", "atapi"},
	{"WebUI", "web_ui"},
	{"API", "api"},
	{"ASCII", "ascii"},
	{"CPU", "cpu"},
	{"CSRF", "csrf"},
	{"CSS", "css"},
	{"DNS", "dns"},
	{"EOF", "eof"},
	{"GUID", "guid"},
	{"HTML", "html"},
	{"HTTP", "http"},
	{"HTTPS", "https"},
	{"ID", "id"},
	{"ip", "ip"},
	{"JSON", "json"},
	{"LHS", "lhs"},
	{"QPS", "qps"},
	{"RAM", "ram"},
	{"RHS", "rhs"},
	{"RPC", "rpc"},
	{"SLA", "sla"},
	{"SMTP", "smtp"},
	{"SQL", "sql"},
	{"SSH", "ssh"},
	{"TCP", "tcp"},
	{"TLS", "tls"},
	{"TTL", "ttl"},
	{"UDP", "udp"},
	{"UI", "ui"},
	{"UID", "uid"},
	{"UUID", "uuid"},
	{"URI", "uri"},
	{"URL", "url"},
	{"UTF8", "utf8"},
	{"VM", "vm"},
	{"XML", "xml"},
	{"XSRF", "xsrf"},
	{"XSS", "xss"},
}

func TestToSnakeCase(t *testing.T) {
	for _, v := range append(commonTestCasesForToSnakeCase, []struct {
		input, expect string
	}{
		{"ＴｈｅＱｕｉｃｋＢｒｏｗｎＦｏｘＯｖｅｒＴｈｅＬａｚｙＤｏｇ", "ｔｈｅ_ｑｕｉｃｋ_ｂｒｏｗｎ_ｆｏｘ_ｏｖｅｒ_ｔｈｅ_ｌａｚｙ_ｄｏｇ"},
	}...) {
		actual := stringutil.ToSnakeCase(v.input)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`stringutil.ToSnakeCase(%#v) => %#v; want %#v`, v.input, actual, expect)
		}
	}
}

func TestToSnakeCaseASCII(t *testing.T) {
	for _, v := range commonTestCasesForToSnakeCase {
		actual := stringutil.ToSnakeCaseASCII(v.input)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`stringutil.ToSnakeCaseASCII(%#v) => %#v; want %#v`, v.input, actual, expect)
		}
	}
}

func TestAddCommonInitialismWithToUpperCamelCase(t *testing.T) {
	input := "test_case"
	actual := stringutil.ToUpperCamelCase(input)
	expect := "TestCase"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`ToUpperCamelCase(%#v) with AddCommonInitialism => %#v; want %#v`, input, actual, expect)
	}
	stringutil.AddCommonInitialism("TEST", "CASE")
	defer stringutil.DelCommonInitialism("TEST", "CASE")
	actual = stringutil.ToUpperCamelCase(input)
	expect = "TESTCASE"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`ToUpperCamelCase(%#v) with AddCommonInitialism => %#v; want %#v`, input, actual, expect)
	}
}

func TestAddCommonInitialismWithToSnakeCase(t *testing.T) {
	input := "TESTCase"
	actual := stringutil.ToSnakeCase(input)
	expect := "t_e_s_t_case"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`ToSnakeCase(%#v) with AddCommonInitialism => %#v; want %#v`, input, actual, expect)
	}
	stringutil.AddCommonInitialism("TEST", "CASE")
	defer stringutil.DelCommonInitialism("TEST", "CASE")
	actual = stringutil.ToSnakeCase(input)
	expect = "test_case"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`ToSnakeCase(%#v) with AddCommonInitialism => %#v; want %#v`, input, actual, expect)
	}
}
