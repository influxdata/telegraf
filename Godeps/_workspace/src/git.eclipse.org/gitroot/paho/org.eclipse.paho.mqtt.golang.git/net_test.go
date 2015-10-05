package mqtt

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func Test_openConnection(t *testing.T) {
	_, err := strconv.Atoi("")
	e := fmt.Errorf(" : %s", err)
	t.Errorf("%#v", e)

	e1 := errors.New("hogehoge %s")
	t.Errorf("%#v", e1)
}
