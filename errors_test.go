package errors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		err  string
		want error
	}{
		{"", fmt.Errorf("")},
		{"foo", fmt.Errorf("foo")},
		{"foo", New("foo")},
		{"string with format specifiers: %v", errors.New("string with format specifiers: %v")},
	}

	for _, tt := range tests {
		got := New(tt.err)
		if got.Error() != tt.want.Error() {
			t.Errorf("New.Error(): got: %q, want %q", got, tt.want)
		}
	}
}

func TestWrapNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrap(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

type nilError struct{}

func (nilError) Error() string { return "nil error" }

func TestCause(t *testing.T) {
	x := New("error")
	tests := []struct {
		name string
		err  error
		want error
	}{{
		name: "all nil",
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		name: " explicit nil error is nil",
		err:  (error)(nil),
		want: nil,
	}, {
		name: "typed nil is nil",
		err:  (*nilError)(nil),
		want: (*nilError)(nil),
	}, {
		name: "uncaused error is unaffected",
		err:  io.EOF,
		want: io.EOF,
	}, {
		name: "caused error returns cause",
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		name: "errors.New self",
		err:  x,
		want: x,
	}, {
		name: "nil With",
		err:  WrapNoStack(nil, "whoops"),
		want: nil,
	}, {
		name: "WrapNoStack",
		err:  WrapNoStack(io.EOF, "whoops"),
		want: io.EOF,
	}, {
		name: "AddStack nil",
		err:  AddStack(nil),
		want: nil,
	}, {
		name: "AddStack",
		err:  AddStack(io.EOF),
		want: io.EOF,
	}}

	for _, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %s: got %#v, want %#v", tt.name, got, tt.want)
		}
	}
}

func TestWrapfNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrapf(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrapf(io.EOF, "read error without format specifiers"), "client error", "client error: read error without format specifiers: EOF"},
		{Wrapf(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrapf(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{Errorf("read error without format specifiers"), "read error without format specifiers"},
		{Errorf("read error with %d format specifier", 1), "read error with 1 format specifier"},
		{Errorf("wrapped error %w", errors.New("wrapped")), "wrapped error wrapped"},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("Errorf(%v): got: %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestAddStackNil(t *testing.T) {
	got := AddStack(nil)
	if got != nil {
		t.Errorf("AddStack(nil): got %#v, expected nil", got)
	}
	got = AddStack(nil)
	if got != nil {
		t.Errorf("AddStack(nil): got %#v, expected nil", got)
	}
}

func TestAddStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{AddStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := AddStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("AddStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestAddStackSkip(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{AddStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := AddStackSkip(tt.err, 2).Error()
		if got != tt.want {
			t.Errorf("AddStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestGetStackTracer(t *testing.T) {
	orig := io.EOF
	if GetStackTracer(orig) != nil {
		t.Errorf("GetStackTracer: got: %v, want %v", GetStackTracer(orig), nil)
	}
	stacked := AddStack(orig)
	if GetStackTracer(stacked).(error) != stacked {
		t.Errorf("GetStackTracer(stacked): got: %v, want %v", GetStackTracer(stacked), stacked)
	}
	final := AddStack(stacked)
	if GetStackTracer(final).(error) != stacked {
		t.Errorf("GetStackTracer(final): got: %v, want %v", GetStackTracer(final), stacked)
	}
}

func TestAddStackDedup(t *testing.T) {
	stacked := AddStack(io.EOF)
	err := AddStack(stacked)
	if err != stacked {
		t.Errorf("AddStack: got: %+v, want %+v", err, stacked)
	}
}

func TestWrapNoStackNil(t *testing.T) {
	got := WrapNoStack(nil, "no error")
	if got != nil {
		t.Errorf("WrapNoStack(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrapNoStack(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WrapNoStack(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WrapNoStack(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WrapNoStack(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

// errors.New, etc values are not expected to be compared by value
// but the change in errors#27 made them incomparable. Assert that
// various kinds of errors have a functional equality operator, even
// if the result of that equality is always false.
func TestErrorEquality(t *testing.T) {
	vals := []error{
		nil,
		io.EOF,
		errors.New("EOF"),
		New("EOF"),
		Errorf("EOF"),
		Wrap(io.EOF, "EOF"),
		Wrapf(io.EOF, "EOF%d", 2),
		WrapNoStack(nil, "whoops"),
		WrapNoStack(io.EOF, "whoops"),
		AddStack(io.EOF),
		AddStack(nil),
		AddStack(io.EOF),
		AddStack(nil),
	}

	for i := range vals {
		for j := range vals {
			_ = vals[i] == vals[j] // mustn't panic
		}
	}
}

type errWalkTest struct {
	cause error
	sub   []error
	v     int
}

func (e *errWalkTest) Error() string {
	return strconv.Itoa(e.v)
}

func (e *errWalkTest) Unwrap() error {
	return e.cause
}

func (e *errWalkTest) Errors() []error {
	return e.sub
}

func testFind(err error, v int) bool {
	return WalkDeep(err, func(err error) bool {
		e := err.(*errWalkTest)
		return e.v == v
	})
}

func TestWalkDeep(t *testing.T) {
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

	if !testFind(err, 11) {
		t.Errorf("not found in first cause chain")
	}

	if !testFind(err, 22) {
		t.Errorf("not found in siblings")
	}

	if testFind(err, 32) {
		t.Errorf("found not exists")
	}
}

type FindMe struct {
	a int
}

func (fm FindMe) Error() string {
	return "you found me!"
}

func TestAsType(t *testing.T) {
	var err error
	var errAs FindMe
	var found bool
	var errorValue = 1
	err = FindMe{a: errorValue}
	errAs, found = AsType[FindMe](err)
	if !found || errAs.a != errorValue {
		t.Errorf("dif not find error 0 levels deep")
	}

	err = Wrap(err, "wrapped up")
	errAs, found = AsType[FindMe](err)
	if !found || errAs.a != errorValue {
		t.Errorf("did not find error 1 levels deep")
	}

	err = nilError{}
	errAs, found = AsType[FindMe](err)
	if found {
		t.Errorf("should not have found a different error type")
	}
}

func TestFormatWrapped(t *testing.T) {
	bottom := New("underlying")
	wrapped := Wrap(bottom, "wrapped")
	if fmt.Sprintf("%v", wrapped) != "wrapped: underlying" {
		t.Errorf("Unexpected wrapping format: %v", wrapped)
	}
	if strings.HasPrefix(fmt.Sprintf("%+v", wrapped), "wrapped: underlying") {
		t.Errorf("Unexpected wrapping format: %+v", wrapped)
	}
	unwrapped := Unwrap(wrapped)
	got := fmt.Sprintf("%v", unwrapped)
	if got != "underlying" {
		t.Errorf("Unexpected unwrapping format, got: %s, wrapped: %v", got, wrapped)
	}
	if !strings.HasPrefix(fmt.Sprintf("%+v", unwrapped), "underlying") {
		t.Errorf("Unexpected unwrapping format: %+v", wrapped)
	}
}

type WrappedInPlace struct {
	*ErrorWrap
}

func TestErrorWrapper(t *testing.T) {
	err := WrappedInPlace{&ErrorWrap{New("underlying")}}
	if err.Error() != "underlying" {
		t.Errorf("Error()")
	}
	err.WrapError(WrapFn("wrap"))
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

type ErrArray []error

func (ea ErrArray) Error() string {
	return errors.Join(ea).Error()
}

func TestIsNil(t *testing.T) {
	var err error = (*nilError)(nil)
	got := Wrap(err, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
	if IsNil(nil) == false {
		t.Errorf("IsNil expected true")
	}
	if IsNil(err) == false {
		t.Errorf("IsNil expected true")
	}
	if IsNil(ErrArray([]error{})) == false {
		t.Errorf("IsNil expected true")
	}

	if IsNil(nilError{}) == true {
		t.Errorf("IsNil expected false")
	}
	if IsNil(ErrArray([]error{nilError{}})) == true {
		t.Errorf("IsNil expected false")
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
