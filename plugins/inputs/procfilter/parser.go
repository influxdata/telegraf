package procfilter

/* Hope this simple handwritten parser will be easier to maintain/contribute to
than a classic yacc one.
The micro configuration language is quite simple to parse anyway.
https://blog.gopheracademy.com/advent-2014/parsers-lexers/
*/
import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser represents a parser.
type Parser struct {
	s            *Scanner
	n2m          map[string]*measurement // ie for measurements
	n2f          map[string]filter       // named (user declared) filters
	f2n          map[filter]string       // from a filter to its name
	filters      []filter                // sequence of filters to process in this order before emitting the measurements
	measurements []*measurement
	buf          struct {
		tok tokenType // last read token
		lit string    // last read literal
		n   int       // buffer size (max=1)
	}
}

var currentParser *Parser // ugly, should be par of the filter context but would need a lot of code boiler, until proven otherwise this is good enough.

// NewParser returns a new instance of Parser. Beware only one active parser at a time.
func NewParser(r io.Reader) *Parser {
	p := Parser{}
	currentParser = &p
	p.s = newScanner(r)
	p.n2f = map[string]filter{}
	p.n2m = map[string]*measurement{}
	p.n2f["all"] = new(allFilter)
	p.n2f["unmeasured"] = new(unpackFilter)
	p.f2n = map[filter]string{}
	p.measurements = []*measurement{}
	p.filters = []filter{}
	return &p
}

// syntaxError build a syntax error string with position information
func (p *Parser) syntaxError(msg string) error {
	return fmt.Errorf("%s (%s)", msg, p.s.posInfo())
}

// syntaxErrorGlobal build a syntax error string with no position information
func (p *Parser) syntaxErrorGlobal(msg string) error {
	return fmt.Errorf("%s", msg)
}

// addFilter add a filter to the set of known (named) filters. Checks for multiple declarations.
func (p *Parser) addFilter(name string, f filter) error {
	old := p.n2f[name]
	if old != nil {
		return p.syntaxError(fmt.Sprintf("filter %q already declared", name))
	}
	p.n2f[name] = f
	p.f2n[f] = name
	p.filters = append(p.filters, f)
	return nil
}

// addMeasurement add a measurement to the set of known (named) measurements. Checks for multiple declarations.
func (p *Parser) addMeasurement(name string, m *measurement) error {
	old := p.n2m[name]
	if old != nil {
		return p.syntaxError(fmt.Sprintf("measurement %q already declared", name))
	}
	p.n2m[name] = m
	p.measurements = append(p.measurements, m)
	p.filters = append(p.filters, m.f) // a measurement is also a filter
	return nil
}

// namedFilter search for a previously declared (named) filter
func (p *Parser) namedFilter(name string) (filter, error) {
	f := p.n2f[name]
	if f == nil {
		return nil, p.syntaxError(fmt.Sprintf("unknown filter %q", name))
	}
	return f, nil
}

// funcFilter get a new filter object corresponding to the function name
func (p *Parser) funcFilter(funcName string) (filter, error) {
	f := name2FuncFilter(funcName)
	if f == nil {
		return nil, p.syntaxError(fmt.Sprintf("unknown filter %q", funcName))
	}
	return f, nil
}

// parse parse a script mande of comments, filters or measurements definitions. The proper order in which to evaluate filters and measurements is stored in the parser.
func (p *Parser) Parse() error {
	for {
		tok, name := p.scanIgnoreWhitespace()
		// end of declarations
		if tok == tTEOF {
			if len(p.measurements) == 0 {
				logErr("You need at least a measurement (m = ...), to emit something.")
			}
			return nil
		}
		// First token should be an indentifier
		if tok != tTIdentifier {
			return p.syntaxError(fmt.Sprintf("found %q, expecting identifier", name))
		}
		tok, lit := p.scanIgnoreWhitespace()
		// Only two constructs in this micro language:
		// ident <- filter
		// ident = measurement <- filter
		switch tok {
		case tTLeftArrow: // [name <-] filter
			f, err := p.parseFilter(name)
			if err != nil {
				return err
			}
			err = p.addFilter(name, f) // this filter is named, store for later retrieval
			if err != nil {
				return err
			}
		case tTEqual: // name = measurement <- filter
			//if
			m, err := p.parseMeasurement(name)
			if err != nil {
				return err
			}
			tok, lit := p.scanIgnoreWhitespace()
			if tok != tTLeftArrow {
				return p.syntaxError(fmt.Sprintf("found %q, expecting '<-'", lit))
			}
			f, err := p.parseFilter("")
			if err != nil {
				return err
			}
			m.f = f
			err = p.addMeasurement(name, m)
			if err != nil {
				return err
			}
		default:
			return p.syntaxError(fmt.Sprintf("found %q, expecting '=' or '<-'", lit))
		}
	}
}

// parseMeasurement parse a measurement.
// eg: [m =] tags(user) fields(cpu,rss) [<-]
func (p *Parser) parseMeasurement(name string) (*measurement, error) {
	var fields, tags []string
	for {
		tok, _ := p.scanIgnoreWhitespace()
		p.unscan()
		if tok == tTLeftArrow {
			break
		}
		ident, err := p.parseIdentifier("")
		if err != nil {
			return nil, err
		}
		switch ident {
		case "tags", "tag":
			if tags != nil {
				return nil, p.syntaxError(fmt.Sprintf("found more than one tag declaration for measurement %q", name))
			}
			err = p.parseSymbol('(')
			if err != nil {
				return nil, err
			}
			tags, err = p.parseIdentifierList()
			if err != nil {
				return nil, err
			}
		case "fields", "field":
			if fields != nil {
				return nil, p.syntaxError(fmt.Sprintf("found more than one tag declaration for measurement %q", name))
			}
			err = p.parseSymbol('(')
			if err != nil {
				return nil, err
			}
			fields, err = p.parseIdentifierList()
			if err != nil {
				return nil, err
			}
		default:
			return nil, p.syntaxError(fmt.Sprintf("found %q, expecting a ´tag´ or ´field´ for measurement %q", ident, name))
		}
	}
	if tags == nil && fields == nil {
		return nil, p.syntaxError(fmt.Sprintf("found no tag or field to emit for measurement %q", name))
	}
	m := new(measurement)
	m.name = name
	m.tags = tags
	m.fields = fields
	return m, nil
}

// parseFilter parse a filter
// eg: top(5,rss,name("apache"))
func (p *Parser) parseFilter(name string) (filter, error) {
	ident, err := p.parseIdentifier("")
	if err != nil {
		return nil, err
	}
	// a filter is a simple identifier (the name of a filter)
	// or a funcion with its parameterst
	// eg:  top(5,cpu,apache)
	err = p.parseSymbol('(')
	if err != nil {
		// {name} [^(]
		p.unscan() // keep this token for later
		if name == ident {
			return nil, p.syntaxError(fmt.Sprintf("you can't use %q filter in its own definition", name))
		}
		// Name designing a known filter?
		f, err := p.namedFilter(ident)
		return f, err
	} else {
		// {ident(} p1[,px]*)
		f, err := p.funcFilter(ident)
		if err != nil {
			return nil, err
		}
		err = f.Parse(p)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

// parseIdentifier consume an identifier (and lowercase it)
func (p *Parser) parseIdentifier(ident string) (string, error) {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != tTIdentifier {
		p.unscan()
		return "", p.syntaxError(fmt.Sprintf("found %q, expecting an identifier", lit))
	} else if ident != "" && lit != ident {
		// lit is not the proper identifier
		p.unscan()
		return "", p.syntaxError(fmt.Sprintf("found %q, expecting %q identifier", lit, ident))
	}
	lit = strings.ToLower(lit)
	return lit, nil
}

// parseNumber parse a nimber
func (p *Parser) parseInt() (int64, error) {
	tok, lit := p.scanIgnoreWhitespace()
	if tok != tTNumber {
		p.unscan()
		return 0, p.syntaxError(fmt.Sprintf("found %q, expecting a number", lit))
	}
	i, err := strconv.ParseInt(lit, 10, 64)
	if err != nil {
		// should not happend if the scan is OK
		return 0, p.syntaxError(fmt.Sprintf("unable to convert %q to an integer value", lit))
	}
	return i, nil
}

// parseIdentifierList parse a list of identifiers
// eg: a,b,c) note the last ) that will be cosummed
func (p *Parser) parseIdentifierList() ([]string, error) {
	var il []string
	for {
		tok, lit := p.scanIgnoreWhitespace()
		if tok != tTIdentifier {
			p.unscan()
			return nil, p.syntaxError(fmt.Sprintf("found %q, expecting an identifier", lit))
		}
		il = append(il, lit)
		tok, lit = p.scanIgnoreWhitespace()
		if tok != tTComma && tok != tTRightPar {
			p.unscan()
			return nil, p.syntaxError(fmt.Sprintf("found %q, expecting ',' or ')", lit))
		} else if tok == tTRightPar {
			return il, nil
		}
	}
}

// parseSymbol parse the desired symbol
func (p *Parser) parseSymbol(ch rune) error {
	_, lit := p.scanIgnoreWhitespace()
	if len(lit) != 1 || lit[0] != byte(ch) {
		p.unscan()
		return p.syntaxError(fmt.Sprintf("found %q, expecting %q", lit, ch))
	}
	return nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok tokenType, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tokenType, string) {
	for {
		tok, lit := p.scan()
		if tok != tTComment && tok != tTWhitespace {
			//fmt.Printf("lit:'%s'\n", lit)
			return tok, lit
		}
	}
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// Parse a filter
func (p *Parser) parseArgFilter(pa *filter) error {
	a, err := p.parseFilter("")
	if err != nil {
		return err
	}
	*pa = a
	return p.parseArgSep() // parse a filter declaration or if next token is a ) assume this is the last optional parameer all that repalce a filter with all processes
}

// Parse an optional last filter argument. If not present (found a ')) then assume this is a allFilter
func (p *Parser) parseArgLastFilter(pa *filter) error {
	// parse a filter declaration or if next token is a ) assume this is the last optional parameer all that repalce a filter with all processes
	tok, _ := p.scan()
	p.unscan()
	if tok == tTRightPar {
		// special case with optional ,all) filter as last argument
		*pa = new(allFilter)
		return nil
	}
	a, err := p.parseFilter("")
	if err != nil {
		return err
	}
	*pa = a
	return nil
}

// Consumes a list of 'atLeast' filters until a ')' (not consummed). If atLeast is 0 and nothing is found fill with the optional All filter.
func (p *Parser) parseArgFilterList(pa *[]filter, atLeast int) error {
	// f1[,f2]*) the last ) is left in the scanner
	*pa = make([]filter, 0, 2) // most of the time we have <=2 inputs
	for {
		tok, _ := p.scan()
		p.unscan()
		if tok == tTRightPar {
			if len(*pa) < atLeast {
				return p.syntaxError(fmt.Sprintf("need at least %d filter arguments (found %d)", atLeast, len(*pa)))
			}
			// No filters but we accept 0 args => insert the optional all filter.
			// eg: top(5,rss) equivalent to top(5,rss,all)
			if atLeast == 0 && len(*pa) == 0 {
				*pa = append(*pa, p.n2f["all"])
			}
			return nil
		}
		a, err := p.parseFilter("")
		if err != nil {
			return err
		}
		*pa = append(*pa, a)
		err = p.parseArgSep()
		if err != nil {
			return err
		}
	}
}

func (p *Parser) parseArgIdentifier(pa *string) error {
	a, err := p.parseIdentifier("")
	if err != nil {
		return err
	}
	*pa = a
	return p.parseArgSep()
}

func (p *Parser) parseArgString(pa *string) error {
	tok, a := p.scanIgnoreWhitespace()
	if tok != tTString {
		return p.syntaxError(fmt.Sprintf("found %q, expecting a string", a))
	}
	*pa = a
	return p.parseArgSep()
}

func (p *Parser) parseArgStregexp(pa **stregexp) error {
	var a *stregexp
	var err error
	tok, lit := p.scanIgnoreWhitespace()
	tokSymb, _ := p.scan()
	var invert bool
	if tokSymb == tTBang {
		invert = true
	} else {
		p.unscan()
	}
	if tok == tTRegexp {
		a, err = NewStregexp(lit, true, invert)
		if err != nil {
			return err
		}
	} else if tok == tTString {
		a, err = NewStregexp(lit, false, invert)
	} else {
		return p.syntaxError(fmt.Sprintf("found %q, expecting a string (or a regexp)", a))
	}
	*pa = a
	return p.parseArgSep()
}

func (p *Parser) parseArgInt(pa *int64) error {
	a, err := p.parseInt()
	if err != nil {
		return err
	}
	*pa = a
	return p.parseArgSep()
}

// Consume a , but keep the ) in the scanner, other tokens are syntax errors
func (p *Parser) parseArgSep() error {
	tok, lit := p.scan()
	if tok == tTComma {
		return nil
	} else if tok == tTRightPar {
		p.unscan()
		return nil
	}
	return p.syntaxError(fmt.Sprintf("found %q, expecting ','  or ')'", lit))
}
