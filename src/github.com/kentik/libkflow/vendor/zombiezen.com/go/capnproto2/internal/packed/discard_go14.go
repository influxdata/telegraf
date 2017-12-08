// +build !go1.5

package packed

import (
	"bufio"
	"io"
	"io/ioutil"
)

func discard(r *bufio.Reader, n int) {
	io.CopyN(ioutil.Discard, r, int64(n))
}
