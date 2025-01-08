package errwrap

import (
	"fmt"
	"reflect"
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

func isNil(err error) bool {
	if err == nil {
		return true
	}

	v := reflect.ValueOf(err)
	k := v.Kind()
	switch k {
	case reflect.Pointer, reflect.UnsafePointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.IsNil() || v.Len() == 0
	}

	return false
}
