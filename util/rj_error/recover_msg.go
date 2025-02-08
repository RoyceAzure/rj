package rj_error

import (
	"fmt"
	"runtime/debug"
)

func GetRecoverMsg(recover any) string {
	var errMsg string
	stack := debug.Stack()
	switch v := recover.(type) {
	case error:
		errMsg = v.Error()
	case string:
		errMsg = v
	default:
		errMsg = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("Error: %s\nStack Trace:\n%s", errMsg, stack)
}
