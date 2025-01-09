package errwrap_test

import (
	"errors"
	"fmt"

	"github.com/gregwebs/errors/errwrap"
)

type inplace struct {
	*errwrap.ErrorWrap
}

func wrapFn(msg string) func(error) error {
	return func(err error) error { return fmt.Errorf("%s: %w", msg, err) }
}

func ExampleWrapInPlace() {
	err := inplace{errwrap.NewErrorWrap(errors.New("original error"))}

	// Wrap the error in place
	wrapped := errwrap.WrapInPlace(err, wrapFn("wrapped"))

	// Print the error and whether it was wrapped in place
	fmt.Printf("Wrapped in place: %v\n", wrapped)
	fmt.Printf("Error: %v\n", err)

	// Try with a regular error that doesn't implement ErrorWrapper
	regularErr := errors.New("regular error")
	wrapped = errwrap.WrapInPlace(regularErr, wrapFn("wrapped"))

	// Print the result for regular error
	fmt.Printf("Regular error wrapped in place: %v\n", wrapped)
	fmt.Printf("Regular error: %v\n", regularErr)

	// Output:
	// Wrapped in place: true
	// Error: wrapped: original error
	// Regular error wrapped in place: false
	// Regular error: regular error
}

type errorGroup struct {
	errs []error
}

func (eg *errorGroup) Error() string {
	return errors.Join(eg.errs...).Error()
}

func (eg *errorGroup) Unwrap() []error { return eg.errs }

func ExampleUnwraps() {
	var eg errorGroup
	eg.errs = append(eg.errs, errors.New("error1"))
	eg.errs = append(eg.errs, errors.New("error2"))

	fmt.Println(errwrap.Errors(nil))
	fmt.Println(errwrap.Unwraps(errors.New("test")))
	fmt.Println(errwrap.Unwraps(&eg))
	// Output:
	// []
	// []
	// [error1 error2]
}

type ErrEmpty struct{}

func (et ErrEmpty) Error() string {
	return "empty"
}
