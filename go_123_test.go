//go:build go1.23

package errors

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func testFindErr(err error, v int) bool {
	for err := range UnwrapGroups(err) {
		e := err.(*errWalkTest)
		if e.v == v {
			return true
		}
	}
	return false
}

func TestUnwrapGroups(t *testing.T) {
	err := &errWalkTest{
		sub: []error{
			&errWalkTest{
				v:     10,
				cause: &errWalkTest{v: 11},
			},
			&errWalkTest{
				v:     20,
				cause: &errWalkTest{v: 21, cause: &errWalkTest{v: 22}},
			},
			&errWalkTest{
				v:     30,
				cause: &errWalkTest{v: 31},
			},
		},
	}

	if !testFindErr(err, 11) {
		t.Errorf("not found in first cause chain")
	}

	if !testFindErr(err, 22) {
		t.Errorf("not found in siblings")
	}

	if testFindErr(err, 32) {
		t.Errorf("found not exists")
	}
}

func find(origErr error, test func(error) bool) error {
	var foundErr error
	for err := range UnwrapGroups(origErr) {
		if test(err) {
			foundErr = err
			break
		}
	}
	return foundErr
}

func TestFind(t *testing.T) {
	eNew := errors.New("error")
	wrapped := Wrap(nilError{}, "nil")
	tests := []struct {
		err    error
		finder func(error) bool
		found  error
	}{
		{io.EOF, func(_ error) bool { return true }, io.EOF},
		{io.EOF, func(_ error) bool { return false }, nil},
		{io.EOF, func(err error) bool { return err == io.EOF }, io.EOF},
		{io.EOF, func(err error) bool { return err != io.EOF }, nil},

		{eNew, func(err error) bool { return true }, eNew},
		{eNew, func(err error) bool { return false }, nil},

		{nilError{}, func(err error) bool { return true }, nilError{}},
		{nilError{}, func(err error) bool { return false }, nil},
		{nilError{}, func(err error) bool { _, ok := err.(nilError); return ok }, nilError{}},

		{wrapped, func(err error) bool { return true }, wrapped},
		{wrapped, func(err error) bool { return false }, nil},
		{wrapped, func(err error) bool { _, ok := err.(nilError); return ok }, nilError{}},
	}

	for _, tt := range tests {
		got := find(tt.err, tt.finder)
		if got != tt.found {
			t.Errorf("WrapNoStack(%v): got: %q, want %q", tt.err, got, tt.found)
		}
	}
}

func ExampleUnwrapGroups() {
	err1 := New("error 1")
	err2 := New("error 2")
	group := Join(err1, err2)
	wrapped := Wrap(group, "wrapped")

	for e := range UnwrapGroups(wrapped) {
		fmt.Println(e.Error() + "\n")
	}
	// Output:
	// wrapped: error 1
	// error 2
	//
	// error 1
	// error 2
	//
	// error 1
	//
	// error 2
}
