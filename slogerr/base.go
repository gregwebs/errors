package slogerr

import (
	"fmt"
)

func formatterPlusV(s fmt.State, verb rune, err error) {
	if f, ok := err.(fmt.Formatter); ok {
		f.Format(s, verb)
	} else {
		fmt.Fprintf(s, "%+v", err)
	}
}

type unwrapper interface {
	Unwrap() error
}

type unwraps interface {
	Unwrap() []error
}

type errorUnwrap interface {
	Unwrap() error
	// ErrorNoUnwrap is the error message component of the wrapping
	// It will be a prefix of Error()
	// If there is no message in the wrapping then this can return an empty string
	ErrorNoUnwrap() string
}

type stackTraceAware interface {
	HasStack() bool
}
