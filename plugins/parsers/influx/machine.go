//line plugins/parsers/influx/machine.go.rl:1
package influx

import (
	"errors"
	"io"
)

type readErr struct {
	Err error
}

func (e *readErr) Error() string {
	return e.Err.Error()
}

var (
	ErrNameParse      = errors.New("expected measurement name")
	ErrFieldParse     = errors.New("expected field")
	ErrTagParse       = errors.New("expected tag")
	ErrTimestampParse = errors.New("expected timestamp")
	ErrParse          = errors.New("parse error")
	EOF               = errors.New("EOF")
)

//line plugins/parsers/influx/machine.go.rl:318

//line plugins/parsers/influx/machine.go:33
const LineProtocol_start int = 46
const LineProtocol_first_final int = 46
const LineProtocol_error int = 0

const LineProtocol_en_main int = 46
const LineProtocol_en_discard_line int = 34
const LineProtocol_en_align int = 85
const LineProtocol_en_series int = 37

//line plugins/parsers/influx/machine.go.rl:321

type Handler interface {
	SetMeasurement(name []byte) error
	AddTag(key []byte, value []byte) error
	AddInt(key []byte, value []byte) error
	AddUint(key []byte, value []byte) error
	AddFloat(key []byte, value []byte) error
	AddString(key []byte, value []byte) error
	AddBool(key []byte, value []byte) error
	SetTimestamp(tm []byte) error
}

type machine struct {
	data         []byte
	cs           int
	p, pe, eof   int
	pb           int
	lineno       int
	sol          int
	handler      Handler
	initState    int
	key          []byte
	beginMetric  bool
	finishMetric bool
}

func NewMachine(handler Handler) *machine {
	m := &machine{
		handler:   handler,
		initState: LineProtocol_en_align,
	}

//line plugins/parsers/influx/machine.go.rl:354

//line plugins/parsers/influx/machine.go.rl:355

//line plugins/parsers/influx/machine.go.rl:356

//line plugins/parsers/influx/machine.go.rl:357

//line plugins/parsers/influx/machine.go.rl:358

//line plugins/parsers/influx/machine.go.rl:359

//line plugins/parsers/influx/machine.go:90
	{
		(m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:360

	return m
}

func NewSeriesMachine(handler Handler) *machine {
	m := &machine{
		handler:   handler,
		initState: LineProtocol_en_series,
	}

//line plugins/parsers/influx/machine.go.rl:371

//line plugins/parsers/influx/machine.go.rl:372

//line plugins/parsers/influx/machine.go.rl:373

//line plugins/parsers/influx/machine.go.rl:374

//line plugins/parsers/influx/machine.go.rl:375

//line plugins/parsers/influx/machine.go:117
	{
		(m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:376

	return m
}

func (m *machine) SetData(data []byte) {
	m.data = data
	m.p = 0
	m.pb = 0
	m.lineno = 1
	m.sol = 0
	m.pe = len(data)
	m.eof = len(data)
	m.key = nil
	m.beginMetric = false
	m.finishMetric = false

//line plugins/parsers/influx/machine.go:140
	{
		(m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:393
	m.cs = m.initState
}

// Next parses the next metric line and returns nil if it was successfully
// processed.  If the line contains a syntax error an error is returned,
// otherwise if the end of file is reached before finding a metric line then
// EOF is returned.
func (m *machine) Next() error {
	if m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	m.key = nil
	m.beginMetric = false
	m.finishMetric = false

	return m.exec()
}

func (m *machine) exec() error {
	var err error

//line plugins/parsers/influx/machine.go:168
	{
		if (m.p) == (m.pe) {
			goto _test_eof
		}
		goto _resume

	_again:
		switch m.cs {
		case 46:
			goto st46
		case 1:
			goto st1
		case 2:
			goto st2
		case 3:
			goto st3
		case 0:
			goto st0
		case 4:
			goto st4
		case 5:
			goto st5
		case 6:
			goto st6
		case 47:
			goto st47
		case 48:
			goto st48
		case 49:
			goto st49
		case 7:
			goto st7
		case 8:
			goto st8
		case 9:
			goto st9
		case 10:
			goto st10
		case 50:
			goto st50
		case 51:
			goto st51
		case 52:
			goto st52
		case 53:
			goto st53
		case 54:
			goto st54
		case 55:
			goto st55
		case 56:
			goto st56
		case 57:
			goto st57
		case 58:
			goto st58
		case 59:
			goto st59
		case 60:
			goto st60
		case 61:
			goto st61
		case 62:
			goto st62
		case 63:
			goto st63
		case 64:
			goto st64
		case 65:
			goto st65
		case 66:
			goto st66
		case 67:
			goto st67
		case 68:
			goto st68
		case 69:
			goto st69
		case 11:
			goto st11
		case 12:
			goto st12
		case 13:
			goto st13
		case 14:
			goto st14
		case 15:
			goto st15
		case 70:
			goto st70
		case 16:
			goto st16
		case 17:
			goto st17
		case 71:
			goto st71
		case 72:
			goto st72
		case 73:
			goto st73
		case 74:
			goto st74
		case 75:
			goto st75
		case 76:
			goto st76
		case 77:
			goto st77
		case 78:
			goto st78
		case 79:
			goto st79
		case 18:
			goto st18
		case 19:
			goto st19
		case 20:
			goto st20
		case 80:
			goto st80
		case 21:
			goto st21
		case 22:
			goto st22
		case 23:
			goto st23
		case 81:
			goto st81
		case 24:
			goto st24
		case 25:
			goto st25
		case 82:
			goto st82
		case 83:
			goto st83
		case 26:
			goto st26
		case 27:
			goto st27
		case 28:
			goto st28
		case 29:
			goto st29
		case 30:
			goto st30
		case 31:
			goto st31
		case 32:
			goto st32
		case 33:
			goto st33
		case 34:
			goto st34
		case 84:
			goto st84
		case 37:
			goto st37
		case 86:
			goto st86
		case 87:
			goto st87
		case 38:
			goto st38
		case 39:
			goto st39
		case 40:
			goto st40
		case 41:
			goto st41
		case 88:
			goto st88
		case 42:
			goto st42
		case 89:
			goto st89
		case 43:
			goto st43
		case 44:
			goto st44
		case 45:
			goto st45
		case 85:
			goto st85
		case 35:
			goto st35
		case 36:
			goto st36
		}

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof
		}
	_resume:
		switch m.cs {
		case 46:
			goto st_case_46
		case 1:
			goto st_case_1
		case 2:
			goto st_case_2
		case 3:
			goto st_case_3
		case 0:
			goto st_case_0
		case 4:
			goto st_case_4
		case 5:
			goto st_case_5
		case 6:
			goto st_case_6
		case 47:
			goto st_case_47
		case 48:
			goto st_case_48
		case 49:
			goto st_case_49
		case 7:
			goto st_case_7
		case 8:
			goto st_case_8
		case 9:
			goto st_case_9
		case 10:
			goto st_case_10
		case 50:
			goto st_case_50
		case 51:
			goto st_case_51
		case 52:
			goto st_case_52
		case 53:
			goto st_case_53
		case 54:
			goto st_case_54
		case 55:
			goto st_case_55
		case 56:
			goto st_case_56
		case 57:
			goto st_case_57
		case 58:
			goto st_case_58
		case 59:
			goto st_case_59
		case 60:
			goto st_case_60
		case 61:
			goto st_case_61
		case 62:
			goto st_case_62
		case 63:
			goto st_case_63
		case 64:
			goto st_case_64
		case 65:
			goto st_case_65
		case 66:
			goto st_case_66
		case 67:
			goto st_case_67
		case 68:
			goto st_case_68
		case 69:
			goto st_case_69
		case 11:
			goto st_case_11
		case 12:
			goto st_case_12
		case 13:
			goto st_case_13
		case 14:
			goto st_case_14
		case 15:
			goto st_case_15
		case 70:
			goto st_case_70
		case 16:
			goto st_case_16
		case 17:
			goto st_case_17
		case 71:
			goto st_case_71
		case 72:
			goto st_case_72
		case 73:
			goto st_case_73
		case 74:
			goto st_case_74
		case 75:
			goto st_case_75
		case 76:
			goto st_case_76
		case 77:
			goto st_case_77
		case 78:
			goto st_case_78
		case 79:
			goto st_case_79
		case 18:
			goto st_case_18
		case 19:
			goto st_case_19
		case 20:
			goto st_case_20
		case 80:
			goto st_case_80
		case 21:
			goto st_case_21
		case 22:
			goto st_case_22
		case 23:
			goto st_case_23
		case 81:
			goto st_case_81
		case 24:
			goto st_case_24
		case 25:
			goto st_case_25
		case 82:
			goto st_case_82
		case 83:
			goto st_case_83
		case 26:
			goto st_case_26
		case 27:
			goto st_case_27
		case 28:
			goto st_case_28
		case 29:
			goto st_case_29
		case 30:
			goto st_case_30
		case 31:
			goto st_case_31
		case 32:
			goto st_case_32
		case 33:
			goto st_case_33
		case 34:
			goto st_case_34
		case 84:
			goto st_case_84
		case 37:
			goto st_case_37
		case 86:
			goto st_case_86
		case 87:
			goto st_case_87
		case 38:
			goto st_case_38
		case 39:
			goto st_case_39
		case 40:
			goto st_case_40
		case 41:
			goto st_case_41
		case 88:
			goto st_case_88
		case 42:
			goto st_case_42
		case 89:
			goto st_case_89
		case 43:
			goto st_case_43
		case 44:
			goto st_case_44
		case 45:
			goto st_case_45
		case 85:
			goto st_case_85
		case 35:
			goto st_case_35
		case 36:
			goto st_case_36
		}
		goto st_out
	st46:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof46
		}
	st_case_46:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr31
		case 13:
			goto tr31
		case 32:
			goto tr80
		case 35:
			goto tr31
		case 44:
			goto tr31
		case 92:
			goto tr81
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr80
		}
		goto tr79
	tr29:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st1
	tr79:
//line plugins/parsers/influx/machine.go.rl:82

		m.beginMetric = true

//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st1
	st1:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof1
		}
	st_case_1:
//line plugins/parsers/influx/machine.go:590
		switch (m.data)[(m.p)] {
		case 10:
			goto tr2
		case 13:
			goto tr2
		case 32:
			goto tr1
		case 44:
			goto tr3
		case 92:
			goto st8
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr1
		}
		goto st1
	tr1:
		(m.cs) = 2
//line plugins/parsers/influx/machine.go.rl:86

		err = m.handler.SetMeasurement(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr56:
		(m.cs) = 2
//line plugins/parsers/influx/machine.go.rl:99

		err = m.handler.AddTag(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st2:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof2
		}
	st_case_2:
//line plugins/parsers/influx/machine.go:638
		switch (m.data)[(m.p)] {
		case 10:
			goto tr7
		case 13:
			goto tr7
		case 32:
			goto st2
		case 44:
			goto tr7
		case 61:
			goto tr7
		case 92:
			goto tr8
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st2
		}
		goto tr5
	tr5:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st3
	st3:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof3
		}
	st_case_3:
//line plugins/parsers/influx/machine.go:668
		switch (m.data)[(m.p)] {
		case 32:
			goto tr7
		case 44:
			goto tr7
		case 61:
			goto tr10
		case 92:
			goto st12
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr7
		}
		goto st3
	tr2:
		(m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:46

		err = ErrTagParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

		goto _again
	tr7:
		(m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:39

		err = ErrFieldParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

		goto _again
	tr31:
		(m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:32

		err = ErrNameParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

		goto _again
	tr35:
		(m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:53

		err = ErrTimestampParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

		goto _again
	tr82:
		(m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:39

		err = ErrFieldParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

//line plugins/parsers/influx/machine.go.rl:53

		err = ErrTimestampParse
		(m.p)--

		(m.cs) = 34
		{
			(m.p)++
			goto _out
		}

		goto _again
	tr135:
//line plugins/parsers/influx/machine.go.rl:73

		(m.p)--

		{
			goto st46
		}

		goto st0
//line plugins/parsers/influx/machine.go:754
	st_case_0:
	st0:
		(m.cs) = 0
		goto _out
	tr10:
//line plugins/parsers/influx/machine.go.rl:108

		m.key = m.text()

		goto st4
	st4:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof4
		}
	st_case_4:
//line plugins/parsers/influx/machine.go:770
		switch (m.data)[(m.p)] {
		case 34:
			goto st5
		case 45:
			goto tr13
		case 46:
			goto tr14
		case 48:
			goto tr15
		case 70:
			goto tr17
		case 84:
			goto tr18
		case 102:
			goto tr19
		case 116:
			goto tr20
		}
		if 49 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto tr16
		}
		goto tr7
	st5:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof5
		}
	st_case_5:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr22
		case 34:
			goto tr23
		case 92:
			goto tr24
		}
		goto tr21
	tr21:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st6
	tr22:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto st6
	tr26:
//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto st6
	st6:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof6
		}
	st_case_6:
//line plugins/parsers/influx/machine.go:838
		switch (m.data)[(m.p)] {
		case 10:
			goto tr26
		case 34:
			goto tr27
		case 92:
			goto st13
		}
		goto st6
	tr23:
		(m.cs) = 47
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:148

		err = m.handler.AddString(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr27:
		(m.cs) = 47
//line plugins/parsers/influx/machine.go.rl:148

		err = m.handler.AddString(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st47:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof47
		}
	st_case_47:
//line plugins/parsers/influx/machine.go:883
		switch (m.data)[(m.p)] {
		case 10:
			goto tr34
		case 13:
			goto st9
		case 32:
			goto st48
		case 44:
			goto st11
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st48
		}
		goto tr82
	tr110:
		(m.cs) = 48
//line plugins/parsers/influx/machine.go.rl:130

		err = m.handler.AddFloat(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr117:
		(m.cs) = 48
//line plugins/parsers/influx/machine.go.rl:112

		err = m.handler.AddInt(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr122:
		(m.cs) = 48
//line plugins/parsers/influx/machine.go.rl:121

		err = m.handler.AddUint(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr127:
		(m.cs) = 48
//line plugins/parsers/influx/machine.go.rl:139

		err = m.handler.AddBool(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st48:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof48
		}
	st_case_48:
//line plugins/parsers/influx/machine.go:955
		switch (m.data)[(m.p)] {
		case 10:
			goto tr34
		case 13:
			goto st9
		case 32:
			goto st48
		case 45:
			goto tr86
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto tr87
			}
		case (m.data)[(m.p)] >= 9:
			goto st48
		}
		goto tr35
	tr34:
//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto st49
	tr89:
		(m.cs) = 49
//line plugins/parsers/influx/machine.go.rl:157

		err = m.handler.SetTimestamp(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	tr111:
		(m.cs) = 49
//line plugins/parsers/influx/machine.go.rl:130

		err = m.handler.AddFloat(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	tr118:
		(m.cs) = 49
//line plugins/parsers/influx/machine.go.rl:112

		err = m.handler.AddInt(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	tr123:
		(m.cs) = 49
//line plugins/parsers/influx/machine.go.rl:121

		err = m.handler.AddUint(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	tr128:
		(m.cs) = 49
//line plugins/parsers/influx/machine.go.rl:139

		err = m.handler.AddBool(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	st49:
//line plugins/parsers/influx/machine.go.rl:172

		m.finishMetric = true
		(m.cs) = 85
		{
			(m.p)++
			goto _out
		}

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof49
		}
	st_case_49:
//line plugins/parsers/influx/machine.go:1089
		switch (m.data)[(m.p)] {
		case 10:
			goto tr31
		case 13:
			goto tr31
		case 32:
			goto st7
		case 35:
			goto tr31
		case 44:
			goto tr31
		case 92:
			goto tr32
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st7
		}
		goto tr29
	tr80:
//line plugins/parsers/influx/machine.go.rl:82

		m.beginMetric = true

		goto st7
	st7:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof7
		}
	st_case_7:
//line plugins/parsers/influx/machine.go:1119
		switch (m.data)[(m.p)] {
		case 10:
			goto tr31
		case 13:
			goto tr31
		case 32:
			goto st7
		case 35:
			goto tr31
		case 44:
			goto tr31
		case 92:
			goto tr32
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st7
		}
		goto tr29
	tr32:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st8
	tr81:
//line plugins/parsers/influx/machine.go.rl:82

		m.beginMetric = true

//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st8
	st8:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof8
		}
	st_case_8:
//line plugins/parsers/influx/machine.go:1159
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto st0
		}
		goto st1
	tr90:
		(m.cs) = 9
//line plugins/parsers/influx/machine.go.rl:157

		err = m.handler.SetTimestamp(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr112:
		(m.cs) = 9
//line plugins/parsers/influx/machine.go.rl:130

		err = m.handler.AddFloat(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr119:
		(m.cs) = 9
//line plugins/parsers/influx/machine.go.rl:112

		err = m.handler.AddInt(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr124:
		(m.cs) = 9
//line plugins/parsers/influx/machine.go.rl:121

		err = m.handler.AddUint(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr129:
		(m.cs) = 9
//line plugins/parsers/influx/machine.go.rl:139

		err = m.handler.AddBool(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st9:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof9
		}
	st_case_9:
//line plugins/parsers/influx/machine.go:1234
		if (m.data)[(m.p)] == 10 {
			goto tr34
		}
		goto st0
	tr86:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st10
	st10:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof10
		}
	st_case_10:
//line plugins/parsers/influx/machine.go:1250
		if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto st50
		}
		goto tr35
	tr87:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st50
	st50:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof50
		}
	st_case_50:
//line plugins/parsers/influx/machine.go:1266
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st52
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	tr88:
		(m.cs) = 51
//line plugins/parsers/influx/machine.go.rl:157

		err = m.handler.SetTimestamp(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st51:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof51
		}
	st_case_51:
//line plugins/parsers/influx/machine.go:1302
		switch (m.data)[(m.p)] {
		case 10:
			goto tr34
		case 13:
			goto st9
		case 32:
			goto st51
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st51
		}
		goto st0
	st52:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof52
		}
	st_case_52:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st53
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st53:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof53
		}
	st_case_53:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st54
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st54:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof54
		}
	st_case_54:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st55
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st55:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof55
		}
	st_case_55:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st56
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st56:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof56
		}
	st_case_56:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st57
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st57:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof57
		}
	st_case_57:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st58
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st58:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof58
		}
	st_case_58:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st59
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st59:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof59
		}
	st_case_59:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st60
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st60:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof60
		}
	st_case_60:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st61
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st61:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof61
		}
	st_case_61:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st62
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st62:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof62
		}
	st_case_62:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st63
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st63:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof63
		}
	st_case_63:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st64
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st64:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof64
		}
	st_case_64:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st65
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st65:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof65
		}
	st_case_65:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st66
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st66:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof66
		}
	st_case_66:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st67
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st67:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof67
		}
	st_case_67:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st68
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st68:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof68
		}
	st_case_68:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st69
			}
		case (m.data)[(m.p)] >= 9:
			goto tr88
		}
		goto tr35
	st69:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof69
		}
	st_case_69:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr89
		case 13:
			goto tr90
		case 32:
			goto tr88
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr88
		}
		goto tr35
	tr113:
		(m.cs) = 11
//line plugins/parsers/influx/machine.go.rl:130

		err = m.handler.AddFloat(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr120:
		(m.cs) = 11
//line plugins/parsers/influx/machine.go.rl:112

		err = m.handler.AddInt(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr125:
		(m.cs) = 11
//line plugins/parsers/influx/machine.go.rl:121

		err = m.handler.AddUint(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr130:
		(m.cs) = 11
//line plugins/parsers/influx/machine.go.rl:139

		err = m.handler.AddBool(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st11:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof11
		}
	st_case_11:
//line plugins/parsers/influx/machine.go:1763
		switch (m.data)[(m.p)] {
		case 32:
			goto tr7
		case 44:
			goto tr7
		case 61:
			goto tr7
		case 92:
			goto tr8
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr7
		}
		goto tr5
	tr8:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st12
	st12:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof12
		}
	st_case_12:
//line plugins/parsers/influx/machine.go:1789
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr7
		}
		goto st3
	tr24:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st13
	st13:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof13
		}
	st_case_13:
//line plugins/parsers/influx/machine.go:1805
		switch (m.data)[(m.p)] {
		case 34:
			goto st6
		case 92:
			goto st6
		}
		goto tr7
	tr13:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st14
	st14:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof14
		}
	st_case_14:
//line plugins/parsers/influx/machine.go:1824
		switch (m.data)[(m.p)] {
		case 46:
			goto st15
		case 48:
			goto st72
		}
		if 49 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto st75
		}
		goto tr7
	tr14:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st15
	st15:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof15
		}
	st_case_15:
//line plugins/parsers/influx/machine.go:1846
		if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto st70
		}
		goto tr7
	st70:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof70
		}
	st_case_70:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 69:
			goto st16
		case 101:
			goto st16
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st70
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	st16:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof16
		}
	st_case_16:
		switch (m.data)[(m.p)] {
		case 34:
			goto st17
		case 43:
			goto st17
		case 45:
			goto st17
		}
		if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto st71
		}
		goto tr7
	st17:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof17
		}
	st_case_17:
		if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
			goto st71
		}
		goto tr7
	st71:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof71
		}
	st_case_71:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st71
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	st72:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof72
		}
	st_case_72:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 46:
			goto st70
		case 69:
			goto st16
		case 101:
			goto st16
		case 105:
			goto st74
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st73
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	st73:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof73
		}
	st_case_73:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 46:
			goto st70
		case 69:
			goto st16
		case 101:
			goto st16
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st73
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	st74:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof74
		}
	st_case_74:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr118
		case 13:
			goto tr119
		case 32:
			goto tr117
		case 44:
			goto tr120
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr117
		}
		goto tr82
	st75:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof75
		}
	st_case_75:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 46:
			goto st70
		case 69:
			goto st16
		case 101:
			goto st16
		case 105:
			goto st74
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st75
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	tr15:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st76
	st76:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof76
		}
	st_case_76:
//line plugins/parsers/influx/machine.go:2053
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 46:
			goto st70
		case 69:
			goto st16
		case 101:
			goto st16
		case 105:
			goto st74
		case 117:
			goto st77
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st73
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	st77:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof77
		}
	st_case_77:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr123
		case 13:
			goto tr124
		case 32:
			goto tr122
		case 44:
			goto tr125
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr122
		}
		goto tr82
	tr16:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st78
	st78:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof78
		}
	st_case_78:
//line plugins/parsers/influx/machine.go:2113
		switch (m.data)[(m.p)] {
		case 10:
			goto tr111
		case 13:
			goto tr112
		case 32:
			goto tr110
		case 44:
			goto tr113
		case 46:
			goto st70
		case 69:
			goto st16
		case 101:
			goto st16
		case 105:
			goto st74
		case 117:
			goto st77
		}
		switch {
		case (m.data)[(m.p)] > 12:
			if 48 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 57 {
				goto st78
			}
		case (m.data)[(m.p)] >= 9:
			goto tr110
		}
		goto tr82
	tr17:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st79
	st79:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof79
		}
	st_case_79:
//line plugins/parsers/influx/machine.go:2154
		switch (m.data)[(m.p)] {
		case 10:
			goto tr128
		case 13:
			goto tr129
		case 32:
			goto tr127
		case 44:
			goto tr130
		case 65:
			goto st18
		case 97:
			goto st21
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr127
		}
		goto tr82
	st18:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof18
		}
	st_case_18:
		if (m.data)[(m.p)] == 76 {
			goto st19
		}
		goto tr7
	st19:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof19
		}
	st_case_19:
		if (m.data)[(m.p)] == 83 {
			goto st20
		}
		goto tr7
	st20:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof20
		}
	st_case_20:
		if (m.data)[(m.p)] == 69 {
			goto st80
		}
		goto tr7
	st80:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof80
		}
	st_case_80:
		switch (m.data)[(m.p)] {
		case 10:
			goto tr128
		case 13:
			goto tr129
		case 32:
			goto tr127
		case 44:
			goto tr130
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr127
		}
		goto tr82
	st21:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof21
		}
	st_case_21:
		if (m.data)[(m.p)] == 108 {
			goto st22
		}
		goto tr7
	st22:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof22
		}
	st_case_22:
		if (m.data)[(m.p)] == 115 {
			goto st23
		}
		goto tr7
	st23:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof23
		}
	st_case_23:
		if (m.data)[(m.p)] == 101 {
			goto st80
		}
		goto tr7
	tr18:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st81
	st81:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof81
		}
	st_case_81:
//line plugins/parsers/influx/machine.go:2257
		switch (m.data)[(m.p)] {
		case 10:
			goto tr128
		case 13:
			goto tr129
		case 32:
			goto tr127
		case 44:
			goto tr130
		case 82:
			goto st24
		case 114:
			goto st25
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr127
		}
		goto tr82
	st24:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof24
		}
	st_case_24:
		if (m.data)[(m.p)] == 85 {
			goto st20
		}
		goto tr7
	st25:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof25
		}
	st_case_25:
		if (m.data)[(m.p)] == 117 {
			goto st23
		}
		goto tr7
	tr19:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st82
	st82:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof82
		}
	st_case_82:
//line plugins/parsers/influx/machine.go:2305
		switch (m.data)[(m.p)] {
		case 10:
			goto tr128
		case 13:
			goto tr129
		case 32:
			goto tr127
		case 44:
			goto tr130
		case 97:
			goto st21
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr127
		}
		goto tr82
	tr20:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st83
	st83:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof83
		}
	st_case_83:
//line plugins/parsers/influx/machine.go:2333
		switch (m.data)[(m.p)] {
		case 10:
			goto tr128
		case 13:
			goto tr129
		case 32:
			goto tr127
		case 44:
			goto tr130
		case 114:
			goto st25
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr127
		}
		goto tr82
	tr3:
		(m.cs) = 26
//line plugins/parsers/influx/machine.go.rl:86

		err = m.handler.SetMeasurement(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr57:
		(m.cs) = 26
//line plugins/parsers/influx/machine.go.rl:99

		err = m.handler.AddTag(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st26:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof26
		}
	st_case_26:
//line plugins/parsers/influx/machine.go:2381
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr49
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto tr48
	tr48:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st27
	st27:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof27
		}
	st_case_27:
//line plugins/parsers/influx/machine.go:2407
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr51
		case 92:
			goto st32
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st27
	tr51:
//line plugins/parsers/influx/machine.go.rl:95

		m.key = m.text()

		goto st28
	st28:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof28
		}
	st_case_28:
//line plugins/parsers/influx/machine.go:2433
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr54
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto tr53
	tr53:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st29
	st29:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof29
		}
	st_case_29:
//line plugins/parsers/influx/machine.go:2459
		switch (m.data)[(m.p)] {
		case 10:
			goto tr2
		case 13:
			goto tr2
		case 32:
			goto tr56
		case 44:
			goto tr57
		case 61:
			goto tr2
		case 92:
			goto st30
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr56
		}
		goto st29
	tr54:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st30
	st30:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof30
		}
	st_case_30:
//line plugins/parsers/influx/machine.go:2489
		if (m.data)[(m.p)] == 92 {
			goto st31
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st29
	st31:
//line plugins/parsers/influx/machine.go.rl:248
		(m.p)--

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof31
		}
	st_case_31:
//line plugins/parsers/influx/machine.go:2505
		switch (m.data)[(m.p)] {
		case 10:
			goto tr2
		case 13:
			goto tr2
		case 32:
			goto tr56
		case 44:
			goto tr57
		case 61:
			goto tr2
		case 92:
			goto st30
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr56
		}
		goto st29
	tr49:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st32
	st32:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof32
		}
	st_case_32:
//line plugins/parsers/influx/machine.go:2535
		if (m.data)[(m.p)] == 92 {
			goto st33
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st27
	st33:
//line plugins/parsers/influx/machine.go.rl:248
		(m.p)--

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof33
		}
	st_case_33:
//line plugins/parsers/influx/machine.go:2551
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr51
		case 92:
			goto st32
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st27
	st34:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof34
		}
	st_case_34:
		if (m.data)[(m.p)] == 10 {
			goto tr62
		}
		goto st34
	tr62:
//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

//line plugins/parsers/influx/machine.go.rl:78

		{
			goto st85
		}

		goto st84
	st84:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof84
		}
	st_case_84:
//line plugins/parsers/influx/machine.go:2592
		goto st0
	st37:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof37
		}
	st_case_37:
		switch (m.data)[(m.p)] {
		case 32:
			goto tr31
		case 35:
			goto tr31
		case 44:
			goto tr31
		case 92:
			goto tr66
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr31
		}
		goto tr65
	tr65:
//line plugins/parsers/influx/machine.go.rl:82

		m.beginMetric = true

//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st86
	st86:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof86
		}
	st_case_86:
//line plugins/parsers/influx/machine.go:2628
		switch (m.data)[(m.p)] {
		case 10:
			goto tr138
		case 13:
			goto tr139
		case 32:
			goto tr2
		case 44:
			goto tr140
		case 92:
			goto st45
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr2
		}
		goto st86
	tr67:
//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto st87
	tr138:
		(m.cs) = 87
//line plugins/parsers/influx/machine.go.rl:86

		err = m.handler.SetMeasurement(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	tr142:
		(m.cs) = 87
//line plugins/parsers/influx/machine.go.rl:99

		err = m.handler.AddTag(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto _again
	st87:
//line plugins/parsers/influx/machine.go.rl:172

		m.finishMetric = true
		(m.cs) = 85
		{
			(m.p)++
			goto _out
		}

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof87
		}
	st_case_87:
//line plugins/parsers/influx/machine.go:2702
		goto st0
	tr139:
		(m.cs) = 38
//line plugins/parsers/influx/machine.go.rl:86

		err = m.handler.SetMeasurement(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr143:
		(m.cs) = 38
//line plugins/parsers/influx/machine.go.rl:99

		err = m.handler.AddTag(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st38:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof38
		}
	st_case_38:
//line plugins/parsers/influx/machine.go:2735
		if (m.data)[(m.p)] == 10 {
			goto tr67
		}
		goto st0
	tr140:
		(m.cs) = 39
//line plugins/parsers/influx/machine.go.rl:86

		err = m.handler.SetMeasurement(m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	tr144:
		(m.cs) = 39
//line plugins/parsers/influx/machine.go.rl:99

		err = m.handler.AddTag(m.key, m.text())
		if err != nil {
			(m.p)--

			(m.cs) = 34
			{
				(m.p)++
				goto _out
			}
		}

		goto _again
	st39:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof39
		}
	st_case_39:
//line plugins/parsers/influx/machine.go:2771
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr69
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto tr68
	tr68:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st40
	st40:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof40
		}
	st_case_40:
//line plugins/parsers/influx/machine.go:2797
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr71
		case 92:
			goto st43
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st40
	tr71:
//line plugins/parsers/influx/machine.go.rl:95

		m.key = m.text()

		goto st41
	st41:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof41
		}
	st_case_41:
//line plugins/parsers/influx/machine.go:2823
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr74
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto tr73
	tr73:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st88
	st88:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof88
		}
	st_case_88:
//line plugins/parsers/influx/machine.go:2849
		switch (m.data)[(m.p)] {
		case 10:
			goto tr142
		case 13:
			goto tr143
		case 32:
			goto tr2
		case 44:
			goto tr144
		case 61:
			goto tr2
		case 92:
			goto st42
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr2
		}
		goto st88
	tr74:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st42
	st42:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof42
		}
	st_case_42:
//line plugins/parsers/influx/machine.go:2879
		if (m.data)[(m.p)] == 92 {
			goto st89
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st88
	st89:
//line plugins/parsers/influx/machine.go.rl:248
		(m.p)--

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof89
		}
	st_case_89:
//line plugins/parsers/influx/machine.go:2895
		switch (m.data)[(m.p)] {
		case 10:
			goto tr142
		case 13:
			goto tr143
		case 32:
			goto tr2
		case 44:
			goto tr144
		case 61:
			goto tr2
		case 92:
			goto st42
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto tr2
		}
		goto st88
	tr69:
//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st43
	st43:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof43
		}
	st_case_43:
//line plugins/parsers/influx/machine.go:2925
		if (m.data)[(m.p)] == 92 {
			goto st44
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st40
	st44:
//line plugins/parsers/influx/machine.go.rl:248
		(m.p)--

		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof44
		}
	st_case_44:
//line plugins/parsers/influx/machine.go:2941
		switch (m.data)[(m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr71
		case 92:
			goto st43
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto tr2
		}
		goto st40
	tr66:
//line plugins/parsers/influx/machine.go.rl:82

		m.beginMetric = true

//line plugins/parsers/influx/machine.go.rl:28

		m.pb = m.p

		goto st45
	st45:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof45
		}
	st_case_45:
//line plugins/parsers/influx/machine.go:2971
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 13 {
			goto st0
		}
		goto st86
	tr63:
//line plugins/parsers/influx/machine.go.rl:166

		m.lineno++
		m.sol = m.p
		m.sol++ // next char will be the first column in the line

		goto st85
	st85:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof85
		}
	st_case_85:
//line plugins/parsers/influx/machine.go:2989
		switch (m.data)[(m.p)] {
		case 10:
			goto tr63
		case 13:
			goto st35
		case 32:
			goto st85
		case 35:
			goto st36
		}
		if 9 <= (m.data)[(m.p)] && (m.data)[(m.p)] <= 12 {
			goto st85
		}
		goto tr135
	st35:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof35
		}
	st_case_35:
		if (m.data)[(m.p)] == 10 {
			goto tr63
		}
		goto st0
	st36:
		if (m.p)++; (m.p) == (m.pe) {
			goto _test_eof36
		}
	st_case_36:
		if (m.data)[(m.p)] == 10 {
			goto tr63
		}
		goto st36
	st_out:
	_test_eof46:
		(m.cs) = 46
		goto _test_eof
	_test_eof1:
		(m.cs) = 1
		goto _test_eof
	_test_eof2:
		(m.cs) = 2
		goto _test_eof
	_test_eof3:
		(m.cs) = 3
		goto _test_eof
	_test_eof4:
		(m.cs) = 4
		goto _test_eof
	_test_eof5:
		(m.cs) = 5
		goto _test_eof
	_test_eof6:
		(m.cs) = 6
		goto _test_eof
	_test_eof47:
		(m.cs) = 47
		goto _test_eof
	_test_eof48:
		(m.cs) = 48
		goto _test_eof
	_test_eof49:
		(m.cs) = 49
		goto _test_eof
	_test_eof7:
		(m.cs) = 7
		goto _test_eof
	_test_eof8:
		(m.cs) = 8
		goto _test_eof
	_test_eof9:
		(m.cs) = 9
		goto _test_eof
	_test_eof10:
		(m.cs) = 10
		goto _test_eof
	_test_eof50:
		(m.cs) = 50
		goto _test_eof
	_test_eof51:
		(m.cs) = 51
		goto _test_eof
	_test_eof52:
		(m.cs) = 52
		goto _test_eof
	_test_eof53:
		(m.cs) = 53
		goto _test_eof
	_test_eof54:
		(m.cs) = 54
		goto _test_eof
	_test_eof55:
		(m.cs) = 55
		goto _test_eof
	_test_eof56:
		(m.cs) = 56
		goto _test_eof
	_test_eof57:
		(m.cs) = 57
		goto _test_eof
	_test_eof58:
		(m.cs) = 58
		goto _test_eof
	_test_eof59:
		(m.cs) = 59
		goto _test_eof
	_test_eof60:
		(m.cs) = 60
		goto _test_eof
	_test_eof61:
		(m.cs) = 61
		goto _test_eof
	_test_eof62:
		(m.cs) = 62
		goto _test_eof
	_test_eof63:
		(m.cs) = 63
		goto _test_eof
	_test_eof64:
		(m.cs) = 64
		goto _test_eof
	_test_eof65:
		(m.cs) = 65
		goto _test_eof
	_test_eof66:
		(m.cs) = 66
		goto _test_eof
	_test_eof67:
		(m.cs) = 67
		goto _test_eof
	_test_eof68:
		(m.cs) = 68
		goto _test_eof
	_test_eof69:
		(m.cs) = 69
		goto _test_eof
	_test_eof11:
		(m.cs) = 11
		goto _test_eof
	_test_eof12:
		(m.cs) = 12
		goto _test_eof
	_test_eof13:
		(m.cs) = 13
		goto _test_eof
	_test_eof14:
		(m.cs) = 14
		goto _test_eof
	_test_eof15:
		(m.cs) = 15
		goto _test_eof
	_test_eof70:
		(m.cs) = 70
		goto _test_eof
	_test_eof16:
		(m.cs) = 16
		goto _test_eof
	_test_eof17:
		(m.cs) = 17
		goto _test_eof
	_test_eof71:
		(m.cs) = 71
		goto _test_eof
	_test_eof72:
		(m.cs) = 72
		goto _test_eof
	_test_eof73:
		(m.cs) = 73
		goto _test_eof
	_test_eof74:
		(m.cs) = 74
		goto _test_eof
	_test_eof75:
		(m.cs) = 75
		goto _test_eof
	_test_eof76:
		(m.cs) = 76
		goto _test_eof
	_test_eof77:
		(m.cs) = 77
		goto _test_eof
	_test_eof78:
		(m.cs) = 78
		goto _test_eof
	_test_eof79:
		(m.cs) = 79
		goto _test_eof
	_test_eof18:
		(m.cs) = 18
		goto _test_eof
	_test_eof19:
		(m.cs) = 19
		goto _test_eof
	_test_eof20:
		(m.cs) = 20
		goto _test_eof
	_test_eof80:
		(m.cs) = 80
		goto _test_eof
	_test_eof21:
		(m.cs) = 21
		goto _test_eof
	_test_eof22:
		(m.cs) = 22
		goto _test_eof
	_test_eof23:
		(m.cs) = 23
		goto _test_eof
	_test_eof81:
		(m.cs) = 81
		goto _test_eof
	_test_eof24:
		(m.cs) = 24
		goto _test_eof
	_test_eof25:
		(m.cs) = 25
		goto _test_eof
	_test_eof82:
		(m.cs) = 82
		goto _test_eof
	_test_eof83:
		(m.cs) = 83
		goto _test_eof
	_test_eof26:
		(m.cs) = 26
		goto _test_eof
	_test_eof27:
		(m.cs) = 27
		goto _test_eof
	_test_eof28:
		(m.cs) = 28
		goto _test_eof
	_test_eof29:
		(m.cs) = 29
		goto _test_eof
	_test_eof30:
		(m.cs) = 30
		goto _test_eof
	_test_eof31:
		(m.cs) = 31
		goto _test_eof
	_test_eof32:
		(m.cs) = 32
		goto _test_eof
	_test_eof33:
		(m.cs) = 33
		goto _test_eof
	_test_eof34:
		(m.cs) = 34
		goto _test_eof
	_test_eof84:
		(m.cs) = 84
		goto _test_eof
	_test_eof37:
		(m.cs) = 37
		goto _test_eof
	_test_eof86:
		(m.cs) = 86
		goto _test_eof
	_test_eof87:
		(m.cs) = 87
		goto _test_eof
	_test_eof38:
		(m.cs) = 38
		goto _test_eof
	_test_eof39:
		(m.cs) = 39
		goto _test_eof
	_test_eof40:
		(m.cs) = 40
		goto _test_eof
	_test_eof41:
		(m.cs) = 41
		goto _test_eof
	_test_eof88:
		(m.cs) = 88
		goto _test_eof
	_test_eof42:
		(m.cs) = 42
		goto _test_eof
	_test_eof89:
		(m.cs) = 89
		goto _test_eof
	_test_eof43:
		(m.cs) = 43
		goto _test_eof
	_test_eof44:
		(m.cs) = 44
		goto _test_eof
	_test_eof45:
		(m.cs) = 45
		goto _test_eof
	_test_eof85:
		(m.cs) = 85
		goto _test_eof
	_test_eof35:
		(m.cs) = 35
		goto _test_eof
	_test_eof36:
		(m.cs) = 36
		goto _test_eof

	_test_eof:
		{
		}
		if (m.p) == (m.eof) {
			switch m.cs {
			case 7, 37:
//line plugins/parsers/influx/machine.go.rl:32

				err = ErrNameParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 2, 3, 4, 5, 6, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25:
//line plugins/parsers/influx/machine.go.rl:39

				err = ErrFieldParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 26, 27, 28, 30, 32, 33, 39, 40, 41, 42, 43, 44:
//line plugins/parsers/influx/machine.go.rl:46

				err = ErrTagParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 10:
//line plugins/parsers/influx/machine.go.rl:53

				err = ErrTimestampParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 86:
//line plugins/parsers/influx/machine.go.rl:86

				err = m.handler.SetMeasurement(m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

			case 88, 89:
//line plugins/parsers/influx/machine.go.rl:99

				err = m.handler.AddTag(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

			case 47, 48, 49, 51:
//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 46:
//line plugins/parsers/influx/machine.go.rl:82

				m.beginMetric = true

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 1:
//line plugins/parsers/influx/machine.go.rl:86

				err = m.handler.SetMeasurement(m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:46

				err = ErrTagParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 29, 31:
//line plugins/parsers/influx/machine.go.rl:99

				err = m.handler.AddTag(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:46

				err = ErrTagParse
				(m.p)--

				(m.cs) = 34
				{
					(m.p)++
					(m.cs) = 0
					goto _out
				}

			case 74:
//line plugins/parsers/influx/machine.go.rl:112

				err = m.handler.AddInt(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 77:
//line plugins/parsers/influx/machine.go.rl:121

				err = m.handler.AddUint(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 70, 71, 72, 73, 75, 76, 78:
//line plugins/parsers/influx/machine.go.rl:130

				err = m.handler.AddFloat(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 79, 80, 81, 82, 83:
//line plugins/parsers/influx/machine.go.rl:139

				err = m.handler.AddBool(m.key, m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

			case 50, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69:
//line plugins/parsers/influx/machine.go.rl:157

				err = m.handler.SetTimestamp(m.text())
				if err != nil {
					(m.p)--

					(m.cs) = 34
					{
						(m.p)++
						(m.cs) = 0
						goto _out
					}
				}

//line plugins/parsers/influx/machine.go.rl:178

				m.finishMetric = true

//line plugins/parsers/influx/machine.go:3301
			}
		}

	_out:
		{
		}
	}

//line plugins/parsers/influx/machine.go.rl:415

	if err != nil {
		return err
	}

	// This would indicate an error in the machine that was reported with a
	// more specific error.  We return a generic error but this should
	// possibly be a panic.
	if m.cs == 0 {
		m.cs = LineProtocol_en_discard_line
		return ErrParse
	}

	// If we haven't found a metric line yet and we reached the EOF, report it
	// now.  This happens when the data ends with a comment or whitespace.
	//
	// Otherwise we have successfully parsed a metric line, so if we are at
	// the EOF we will report it the next call.
	if !m.beginMetric && m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	return nil
}

// Position returns the current byte offset into the data.
func (m *machine) Position() int {
	return m.p
}

// LineOffset returns the byte offset of the current line.
func (m *machine) LineOffset() int {
	return m.sol
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (m *machine) LineNumber() int {
	return m.lineno
}

// Column returns the current column.
func (m *machine) Column() int {
	lineOffset := m.p - m.sol
	return lineOffset + 1
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}

type streamMachine struct {
	machine *machine
	reader  io.Reader
}

func NewStreamMachine(r io.Reader, handler Handler) *streamMachine {
	m := &streamMachine{
		machine: NewMachine(handler),
		reader:  r,
	}

	m.machine.SetData(make([]byte, 1024))
	m.machine.pe = 0
	m.machine.eof = -1
	return m
}

func (m *streamMachine) Next() error {
	// Check if we are already at EOF, this should only happen if called again
	// after already returning EOF.
	if m.machine.p == m.machine.pe && m.machine.pe == m.machine.eof {
		return EOF
	}

	copy(m.machine.data, m.machine.data[m.machine.p:])
	m.machine.pe = m.machine.pe - m.machine.p
	m.machine.sol = m.machine.sol - m.machine.p
	m.machine.pb = 0
	m.machine.p = 0
	m.machine.eof = -1

	m.machine.key = nil
	m.machine.beginMetric = false
	m.machine.finishMetric = false

	for {
		err := m.machine.exec()
		if err != nil {
			return err
		}

		// If we have successfully parsed a full metric line break out
		if m.machine.finishMetric {
			break
		}

		// Expand the buffer if it is full
		if m.machine.pe == len(m.machine.data) {
			expanded := make([]byte, 2*len(m.machine.data))
			copy(expanded, m.machine.data)
			m.machine.data = expanded
		}

		n, err := m.reader.Read(m.machine.data[m.machine.pe:])
		if n == 0 && err == io.EOF {
			m.machine.eof = m.machine.pe
		} else if err != nil && err != io.EOF {
			// After the reader returns an error this function shouldn't be
			// called again.  This will cause the machine to return EOF this
			// is done.
			m.machine.p = m.machine.pe
			m.machine.eof = m.machine.pe
			return &readErr{Err: err}
		}

		m.machine.pe += n

	}

	return nil
}

// Position returns the current byte offset into the data.
func (m *streamMachine) Position() int {
	return m.machine.Position()
}

// LineOffset returns the byte offset of the current line.
func (m *streamMachine) LineOffset() int {
	return m.machine.LineOffset()
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (m *streamMachine) LineNumber() int {
	return m.machine.LineNumber()
}

// Column returns the current column.
func (m *streamMachine) Column() int {
	return m.machine.Column()
}

// LineText returns the text of the current line that has been parsed so far.
func (m *streamMachine) LineText() string {
	return string(m.machine.data[0:m.machine.p])
}
