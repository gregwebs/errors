# errors

* stack traces for errors
* a more structured API for error wrapping.

There is a [blog article overview here](https://blog.gregweber.info/blog/go-errors-library/).

## Adding stack traces and context to an error

The errors.Wrap function returns a new error that adds context to the original error. For example

```go
_, err := ioutil.ReadAll(r)
if err != nil {
        return errors.Wrap(err, "read failed")
}
```

There are a few other functions available such as `Wrapf` (use a format string) and `New` and `Errorf` (create a new error).

These all add a stack trace to errors.
So generally wherever an error is generated in your code, you want to wrap it to add a stack trace.
If you don't want to bother with adding any additional metadata, you can just use `AddStack`.

You can also add context with slog attributes with `Wraps`

```go
return errors.Wraps(err, "failure", "id", 5, "key", "value)
```

## Using with standard library errors

This library is designed to be used in place of the standard errors package.
There are cases where stack traces are not desired. The suggested pattern is to rename `errors` to `stderrors` when importing.

```go
import (
        stderrors "errors"
        "github.com/gregwebs/errors"
)

var errVar = stderrors.New("stack trace not desired")
```

## Retrieving the cause of an error

Using `errors.Wrap` constructs a stack of errors, adding context to the preceding error. Depending on the nature of the error it may be necessary to reverse the operation of errors.Wrap to retrieve the original error for inspection. Any error value which implements the method `Unwrap` can be inspected by standard errors package functions or by `errors.Cause`.

`errors.Cause` will recursively retrieve the topmost error which does not implement `Unwrap`, which is assumed to be the original cause. For example:

```go
switch err := errors.Cause(err).(type) {
case *MyError:
        // handle specifically
default:
        // unknown error
}
```

## License

BSD-2-Clause
