package errors_test

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/gregwebs/errors"
	"github.com/gregwebs/errors/stackfmt"
)

func ExampleNew() {
	err := errors.New("whoops")
	fmt.Println(err)
	// Output: whoops
}

// Returns the first 'n' lines of a given string, where each line is separated by '\n'.
func firstNLines(s string, n int) string {
	allLines := strings.SplitN(s, "\n", n+1)
	if n > len(allLines) {
		n = len(allLines)
	}
	return strings.Join(allLines[0:n], "\n")
}

func functionLines(s string) string {
	lines := []string{}
	for _, line := range strings.SplitAfter(s, "\n") {
		// function name
		if strings.HasPrefix(line, "github.com") {
			lines = append(lines, line)
			// File name (different on different machine)
		} else if strings.HasPrefix(line, "\t") || len(lines) == 0 {
			continue
		} else {
			break
		}
	}
	return strings.Join(lines, "")
}

func ExampleWrapNoStack() {
	cause := stderrors.New("whoops")
	err := errors.WrapNoStack(cause, "oh noes")
	fmt.Printf("%+v", err)
	// Output:
	// whoops
	// oh noes
}

func ExampleAddStack() {
	// stderrors is the standard library errors: no stack trace
	cause := stderrors.New("whoops")
	fmt.Print(firstNLines(fmt.Sprintf("%+v\n", cause), 2))
	// Add a stack trace to it
	err := errors.AddStack(cause)
	fmt.Print(firstNLines(fmt.Sprintf("%+v\n", err), 2))
	// Output:
	// whoops
	// whoops
	// github.com/gregwebs/errors_test.ExampleAddStack
}

func ExampleAddStackSkip() {
	// stderrors is the standard library errors: no stack trace
	inner := func() {
		cause := stderrors.New("whoops")
		err := errors.AddStack(cause)
		fmt.Print(functionLines(fmt.Sprintf("%+v\n", err)))

		fmt.Println("---")

		// Add a stack trace to it
		err = errors.AddStackSkip(cause, 1)
		fmt.Print(functionLines(fmt.Sprintf("%+v\n", err)))
	}
	inner()
	// Output:
	// github.com/gregwebs/errors_test.ExampleAddStackSkip.func1
	// github.com/gregwebs/errors_test.ExampleAddStackSkip
	// ---
	// github.com/gregwebs/errors_test.ExampleAddStackSkip
}

func ExampleWrap() {
	cause := errors.New("whoops")
	err := errors.Wrap(cause, "oh noes")
	fmt.Println(err)
	// Output: oh noes: whoops
}

func newWrappedErr() error {
	e1 := errors.New("cause")
	e2 := errors.Wrap(e1, "inner")
	e3 := errors.Wrap(e2, "middle")
	return errors.Wrap(e3, "outer")
}

func ExampleCause() {
	err := newWrappedErr()
	fmt.Println(err)
	fmt.Println(errors.Cause(err))
	// Output: outer: middle: inner: cause
	// cause
}

func ExampleWrapf() {
	cause := errors.New("whoops")
	err := errors.Wrapf(cause, "oh noes #%d", 2)
	fmt.Println(err)
	// Output: oh noes #2: whoops
}

func ExampleErrorf() {
	err := errors.Errorf("whoops: %s", "foo")
	verbose := fmt.Sprintf("%+v", err)
	fmt.Print(strings.Join(strings.SplitN(verbose, "\n", 3)[0:2], "\n"))
	// Output:
	// whoops: foo
	// github.com/gregwebs/errors_test.ExampleErrorf
}

func Example_stackTrace() {
	type stackTracer interface {
		StackTrace() stackfmt.StackTrace
	}

	err, ok := errors.Cause(newWrappedErr()).(stackTracer)
	if !ok {
		panic("oops, err does not implement stackTracer")
	}

	st := err.StackTrace()
	fmt.Print(functionLines(fmt.Sprintf("%+v", st)))
	// Output:
	// github.com/gregwebs/errors_test.newWrappedErr
	// github.com/gregwebs/errors_test.Example_stackTrace
}

type ErrEmpty struct{}

func (et ErrEmpty) Error() string {
	return "empty"
}

func ExampleAsType() {
	err := errors.Wrap(ErrEmpty{}, "failed")
	cause, _ := errors.AsType[ErrEmpty](err)
	fmt.Printf("%v", cause)
	// Output: empty
}

type errorGroup struct {
	errs []error
}

func (eg *errorGroup) Error() string {
	return errors.Join(eg.errs...).Error()
}

func (eg *errorGroup) Errors() []error { return eg.errs }

func ExampleErrors() {
	var eg errorGroup
	eg.errs = append(eg.errs, errors.New("error1"))
	eg.errs = append(eg.errs, errors.New("error2"))

	fmt.Println(errors.Errors(nil))
	fmt.Println(errors.Errors(errors.New("test")))
	fmt.Println(errors.Errors(&eg))
	// Output:
	// []
	// []
	// [error1 error2]
}

func ExampleIsNil() {
	var empty *ErrEmpty = nil //nolint:staticcheck
	var err error = empty
	fmt.Println(err == nil) //nolint:staticcheck
	fmt.Println(errors.IsNil(err))
	// Output:
	// false
	// true
}

type inplace struct {
	*errors.ErrorWrap
}

func ExampleWrapInPlace() {
	err := inplace{errors.NewErrorWrap(errors.New("original error"))}

	// Wrap the error in place
	wrapped := errors.WrapInPlace(err, errors.WrapFn("wrapped"))

	// Print the error and whether it was wrapped in place
	fmt.Printf("Wrapped in place: %v\n", wrapped)
	fmt.Printf("Error: %v\n", err)

	// Try with a regular error that doesn't implement ErrorWrapper
	regularErr := errors.New("regular error")
	wrapped = errors.WrapInPlace(regularErr, errors.WrapFn("wrapped"))

	// Print the result for regular error
	fmt.Printf("Regular error wrapped in place: %v\n", wrapped)
	fmt.Printf("Regular error: %v\n", regularErr)

	// Output:
	// Wrapped in place: true
	// Error: wrapped: original error
	// Regular error wrapped in place: false
	// Regular error: regular error
}
