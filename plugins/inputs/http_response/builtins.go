package http_response

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.starlark.net/starlark"
)

func builtinMD5(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var input string
	err := starlark.UnpackArgs("builtinMD5", args, kwargs, "input", &input)
	if err != nil {
		return nil, err
	}
	return starlark.String(fmt.Sprintf("%x", md5.Sum([]byte(input)))), nil
}

func builtinSHA256(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var input string
	err := starlark.UnpackArgs("builtinSHA256", args, kwargs, "input", &input)
	if err != nil {
		return nil, err
	}
	return starlark.String(strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(input))))), nil
}

func builtinNow(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.MakeInt64(time.Now().Unix()), nil
}

func builtinRand(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	rand.Seed(time.Now().Unix())
	return starlark.Float(rand.Float64()), nil
}
