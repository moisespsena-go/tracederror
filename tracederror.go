package tracederror

import (
	"fmt"
	"runtime/debug"

	"github.com/pkg/errors"
)

type Causer interface {
	Cause() error
}

type TracedError interface {
	error
	Causer
	Trace() []byte
}

type tracedError struct {
	error
	trace []byte
}

func badErrOrMsg() {
	panic("bad errOrMsg type")
}

func New(errOrMsg interface{}, trace ...[]byte) TracedError {
	var err error
	switch t := errOrMsg.(type) {
	case string:
		err = errors.New(t)
	case error:
		err = t
	default:
		badErrOrMsg()
	}
	var t []byte
	for _, t = range trace {
	}
	if len(t) == 0 {
		if _, ok := err.(TracedError); !ok {
			t = debug.Stack()
		}
	}
	return &tracedError{err, t}
}

func Wrap(err TracedError, msg string, args ...interface{}) TracedError {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &tracedError{errors.Wrap(err, msg), err.Trace()}
}

func TracedWrap(err interface{}, msg string, args ...interface{}) TracedError {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	switch t := err.(type) {
	case TracedError:
		return Wrap(t, msg)
	case error:
		return New(errors.Wrap(t, msg))
	default:
		return New(errors.Wrap(fmt.Errorf("error::%T = %v", err), msg))
	}
}

func Traced(err interface{}) TracedError {
	switch t := err.(type) {
	case TracedError:
		return t
	case error:
		return New(t)
	default:
		return New(fmt.Errorf("error::%T = %v", err))
	}
}

func (this *tracedError) Cause() error {
	return this.error
}

func (this *tracedError) Trace() []byte {
	if this.trace == nil {
		return this.error.(TracedError).Trace()
	}
	return this.trace
}
