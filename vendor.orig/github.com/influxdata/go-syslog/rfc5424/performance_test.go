package rfc5424

import (
	"testing"
)

// This is here to avoid compiler optimizations that
// could remove the actual call we are benchmarking
// during benchmarks
var benchParseResult *SyslogMessage

type benchCase struct {
	input []byte
	label string
}

var benchCases = []benchCase{
	{
		label: "[no] empty input",
		input: []byte(``),
	},
	{
		label: "[no] multiple syslog messages on multiple lines",
		input: []byte("<1>1 - - - - - -\x0A<2>1 - - - - - -"),
	},
	{
		label: "[no] impossible timestamp",
		input: []byte(`<101>11 2003-09-31T22:14:15.003Z`),
	},
	{
		label: "[no] malformed structured data",
		input: []byte("<1>1 - - - - - X"),
	},
	{
		label: "[no] with duplicated structured data id",
		input: []byte("<165>3 2003-10-11T22:14:15.003Z example.com evnts - ID27 [id1][id1]"),
	},
	{
		label: "[ok] minimal",
		input: []byte(`<1>1 - - - - - -`),
	},
	{
		label: "[ok] average message",
		input: []byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice - - [origin x-service="someservice"][meta sequenceId="14125553"] 127.0.0.1 - - 1456029177 "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
	},
	{
		label: "[ok] complicated message",
		input: []byte(`<78>1 2016-01-15T00:04:01Z host1 CROND 10391 - [meta sequenceId="29" sequenceBlah="foo"][my key="value"] some_message`),
	},
	{
		label: "[ok] very long message",
		input: []byte(`<190>1 2016-02-21T01:19:11+00:00 batch6sj - - - [meta sequenceId="21881798" x-group="37051387"][origin x-service="tracking"] metascutellar conversationalist nephralgic exogenetic graphy streng outtaken acouasm amateurism prenotice Lyonese bedull antigrammatical diosphenol gastriloquial bayoneteer sweetener naggy roughhouser dighter addend sulphacid uneffectless ferroprussiate reveal Mazdaist plaudite Australasian distributival wiseman rumness Seidel topazine shahdom sinsion mesmerically pinguedinous ophthalmotonometer scuppler wound eciliate expectedly carriwitchet dictatorialism bindweb pyelitic idic atule kokoon poultryproof rusticial seedlip nitrosate splenadenoma holobenthic uneternal Phocaean epigenic doubtlessly indirection torticollar robomb adoptedly outspeak wappenschawing talalgia Goop domitic savola unstrafed carded unmagnified mythologically orchester obliteration imperialine undisobeyed galvanoplastical cycloplegia quinquennia foremean umbonal marcgraviaceous happenstance theoretical necropoles wayworn Igbira pseudoangelic raising unfrounced lamasary centaurial Japanolatry microlepidoptera`),
	},
	{
		label: "[ok] all max length and complete",
		input: []byte(`<191>999 2018-12-31T23:59:59.999999-23:59 abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab abcdefghilmnopqrstuvzabcdefghilm [an@id key1="val1" key2="val2"][another@id key1="val1"] Some message "GET"`),
	},
	{
		label: "[ok] all max length except structured data and message",
		input: []byte(`<191>999 2018-12-31T23:59:59.999999-23:59 abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab abcdefghilmnopqrstuvzabcdefghilm -`),
	},
	{
		label: "[ok] minimal with message containing newline",
		input: []byte("<1>1 - - - - - - x\x0Ay"),
	},
	{
		label: "[ok] w/o procid, w/o structured data, with message starting with BOM",
		input: []byte("<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - \xEF\xBB\xBF'su root' failed for lonvick on /dev/pts/8"),
	},
	{
		label: "[ok] minimal with UTF-8 message",
		input: []byte("<0>1 - - - - - - ⠊⠀⠉⠁⠝⠀⠑⠁⠞⠀⠛⠇⠁⠎⠎⠀⠁⠝⠙⠀⠊⠞⠀⠙⠕⠑⠎⠝⠞⠀⠓⠥⠗⠞⠀⠍⠑"),
	},
	{
		label: "[ok] with structured data id, w/o structured data params",
		input: []byte(`<29>50 2016-01-15T01:00:43Z hn S - - [my@id]`),
	},
	{
		label: "[ok] with multiple structured data",
		input: []byte(`<29>50 2016-01-15T01:00:43Z hn S - - [my@id1 k="v"][my@id2 c="val"]`),
	},
	{
		label: "[ok] with escaped backslash within structured data param value, with message",
		input: []byte(`<29>50 2016-01-15T01:00:43Z hn S - - [meta es="\\valid"] 1452819643`),
	},
	{
		label: "[ok] with UTF-8 structured data param value, with message",
		input: []byte(`<78>1 2016-01-15T00:04:01+00:00 host1 CROND 10391 - [sdid x="⌘"] some_message`),
	},
}

func BenchmarkParse(b *testing.B) {
	for _, tc := range benchCases {
		tc := tc
		bestEffort := true
		b.Run(rxpad(tc.label, 50), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchParseResult, _ = NewMachine().Parse(tc.input, &bestEffort)
			}
		})
	}
}
