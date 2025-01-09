package slogerr_test

import (
	"errors"
	"fmt"

	"github.com/gregwebs/errors/slogerr"
)

func ExampleStructuredError() {
	err := slogerr.Wraps(
		errors.New("cause"),
		"structured",
		"key", "value",
		"int", 1,
	)

	fmt.Println(err.Error())
	// Output: structured key=value int=1: cause
}

func ExampleNew() {
	err := slogerr.New(
		"cause",
		"key", "value",
		"int", 1,
	)

	fmt.Println(err.Error())
	// Output: cause key=value int=1
}

func ExampleSlogRecord() {
	err := slogerr.Wraps(
		errors.New("cause"),
		"structured",
		"key", "value",
		"int", 1,
	)

	rec := slogerr.SlogRecord(err)
	fmt.Println(rec.Message)
	// Output:
	// structured: cause
}

func ExampleWraps() {
	// Create a base error
	baseErr := fmt.Errorf("database connection failed")

	// Wrap the error with additional structured information
	wrappedErr := slogerr.Wraps(baseErr, "user authentication failed",
		"user_id", "123",
		"attempt", 3,
	)

	// Print the error
	fmt.Println(wrappedErr)
	// Output: user authentication failed user_id=123 attempt=3: database connection failed
}
