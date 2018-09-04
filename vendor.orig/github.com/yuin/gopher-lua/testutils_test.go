package lua

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func positionString(level int) string {
	_, file, line, _ := runtime.Caller(level + 1)
	return fmt.Sprintf("%v:%v:", filepath.Base(file), line)
}

func errorIfNotEqual(t *testing.T, v1, v2 interface{}) {
	if v1 != v2 {
		t.Errorf("%v '%v' expected, but got '%v'", positionString(1), v1, v2)
	}
}

func errorIfFalse(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		if len(args) > 0 {
			t.Errorf("%v %v", positionString(1), fmt.Sprintf(msg, args...))
		} else {
			t.Errorf("%v %v", positionString(1), msg)
		}
	}
}

func errorIfNotNil(t *testing.T, v1 interface{}) {
	if fmt.Sprint(v1) != "<nil>" {
		t.Errorf("%v nil expected, but got '%v'", positionString(1), v1)
	}
}

func errorIfNil(t *testing.T, v1 interface{}) {
	if fmt.Sprint(v1) == "<nil>" {
		t.Errorf("%v non-nil value expected, but got nil", positionString(1))
	}
}

func errorIfScriptFail(t *testing.T, L *LState, script string) {
	if err := L.DoString(script); err != nil {
		t.Errorf("%v %v", positionString(1), err.Error())
	}
}

func errorIfGFuncFail(t *testing.T, L *LState, f LGFunction) {
	if err := L.GPCall(f, LNil); err != nil {
		t.Errorf("%v %v", positionString(1), err.Error())
	}
}

func errorIfScriptNotFail(t *testing.T, L *LState, script string, pattern string) {
	if err := L.DoString(script); err != nil {
		reg := regexp.MustCompile(pattern)
		if len(reg.FindStringIndex(err.Error())) == 0 {
			t.Errorf("%v error message '%v' does not contains given pattern string '%v'.", positionString(1), err.Error(), pattern)
			return
		}
		return
	}
	t.Errorf("%v script should fail", positionString(1))
}

func errorIfGFuncNotFail(t *testing.T, L *LState, f LGFunction, pattern string) {
	if err := L.GPCall(f, LNil); err != nil {
		reg := regexp.MustCompile(pattern)
		if len(reg.FindStringIndex(err.Error())) == 0 {
			t.Errorf("%v error message '%v' does not contains given pattern string '%v'.", positionString(1), err.Error(), pattern)
			return
		}
		return
	}
	t.Errorf("%v LGFunction should fail", positionString(1))
}
