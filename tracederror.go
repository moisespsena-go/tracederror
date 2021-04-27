package tracederror

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

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

type SimpleStackTracer interface {
	StackTrace() errors.StackTrace
}

type StackTracer interface {
	TracedError
	StackTrace() errors.StackTrace
}

type stackTraceError struct {
	error
	trace []byte
}

func (this stackTraceError) Cause() error {
	return this.error
}

func stackTracerToBytes(t SimpleStackTracer) []byte {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = b[:bytes.IndexByte(b, '\n')+1]

	var buf = bytes.NewBuffer(b)
	for _, f := range t.StackTrace() {
		fmt.Fprintf(buf, "%+v\n", f)
	}
	return []byte(strings.ReplaceAll(buf.String(), "\n\t", "()\n\t"))
}

func (this stackTraceError) Trace() []byte {
	return this.trace
}

func New(errOrMsg interface{}, trace ...[]byte) TracedError {
	var err error
	switch t := errOrMsg.(type) {
	case TracedError:
		return t
	case string:
		err = errors.New(t)
	case interface {
		error
		SimpleStackTracer
	}:
		var trace []byte
		var te error

		if c, ok := t.(Causer); ok {
			te = c.Cause()
		}

	do:
		for te != nil {
			if c, ok := te.(Causer); ok {
				switch t2 := c.(type) {
				case TracedError:
					trace = t2.Trace()
					break do
				case Causer:
					te = t2.Cause()
				case SimpleStackTracer:
					trace = stackTracerToBytes(t2.(SimpleStackTracer))
					break do
				default:
					break do
				}
			} else {
				break
			}
		}
		if trace == nil {
			trace = stackTracerToBytes(t.(SimpleStackTracer))
		}
		te2 := &stackTraceError{t, trace}
		return te2
	case Causer:
		if te, ok := t.Cause().(TracedError); ok {
			trace = [][]byte{te.Trace()}
		}
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
	case string:
		return New(errors.New(t))
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
