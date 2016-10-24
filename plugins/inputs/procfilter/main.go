// parse project main.go
package procfilter

import (
	"fmt"
	"strings"
)

// test conf
const conf string = `// comment to remove
	a<-name(apache)
	a_meas=tags(name) values(args,cpu) <- a
	`

func main() {
	r := strings.NewReader(conf)
	scanner := newScanner(r)
	for {
		tok, s := scanner.scan()
		fmt.Printf("%v %q", tok, s)
		if tok == tTIllegal || tok == tTEOF {
			break
		}
	}
}
