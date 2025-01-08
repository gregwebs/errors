package errwrap

import (
	"errors"
	"fmt"
	"testing"
)

type WrappedInPlace struct {
	*ErrorWrap
}

func TestErrorWrapper(t *testing.T) {
	err := WrappedInPlace{&ErrorWrap{errors.New("underlying")}}
	if err.Error() != "underlying" {
		t.Errorf("Error()")
	}
	err.WrapError(wrapFn("wrap"))
	if err.Error() != "wrap: underlying" {
		t.Errorf("wrap Error()")
	}

	err.WrapError(WrapfFn("wrapf %d", 1))
	if s := err.Error(); s != "wrapf 1: wrap: underlying" {
		t.Errorf("wrapf Error() %s", s)
	}

	/*
		err.WrapError(WrapsFn("wraps", "i", 2))
		if s := err.Error(); s != "wraps i=2: wrapf 1: wrap: underlying" {
			t.Errorf("wrapf Error() %s", s)
		}
	*/
}

// wrapFn returns a wrapping function that calls Wrap
func wrapFn(msg string) func(error) error {
	return func(err error) error { return fmt.Errorf("%s: %w", msg, err) }
}

// wrapFn returns a wrapping function that calls Wrapf
func WrapfFn(msg string, args ...interface{}) func(error) error {
	return func(err error) error {
		fmt.Println(args...)
		return fmt.Errorf(msg+": %w", append(args, err)...)
	}
}

func TestWalkDeepComplexTree(t *testing.T) {
	err := &errWalkTest{v: 1, cause: &errWalkTest{
		sub: []error{
			&errWalkTest{
				v:     10,
				cause: &errWalkTest{v: 11},
			},
			&errWalkTest{
				v: 20,
				sub: []error{
					&errWalkTest{v: 21},
					&errWalkTest{v: 22},
				},
			},
			&errWalkTest{
				v:     30,
				cause: &errWalkTest{v: 31},
			},
		},
	}}

	assertFind := func(v int, comment string) {
		if !testFind(err, v) {
			t.Errorf("%d not found in the error: %s", v, comment)
		}
	}
	assertNotFind := func(v int, comment string) {
		if testFind(err, v) {
			t.Errorf("%d found in the error, but not expected: %s", v, comment)
		}
	}

	assertFind(1, "shallow search")
	assertFind(11, "deep search A1")
	assertFind(21, "deep search A2")
	assertFind(22, "deep search B1")
	assertNotFind(23, "deep search Neg")
	assertFind(31, "deep search B2")
	assertNotFind(32, "deep search Neg")
	assertFind(30, "Tree node A")
	assertFind(20, "Tree node with many children")
}
