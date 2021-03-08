package rx

import "fmt"

var _ error = (*PanicError)(nil)

type PanicError struct {
	msg interface{}
}

func (p PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.msg)
}

func ErrPanic(msg interface{}) error {
	return PanicError{msg: msg}
}
