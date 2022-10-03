package serum

import (
	"fmt"
)

// Errorf produces new Serum-style error values, and attaches a message,
// which may use a formatting pattern.
//
// Provide the error code string as the first parameter, and a message as the second parameter.
// The message may be a format string as per `fmt.Errorf` and friends,
// and additional parameters can be given as varargs.
//
// If a %w verb is used, Errorf will take an error parameter in the args and attach it as "cause",
// similarly to the behavior of `fmt.Errorf`.
// However, if that error is not already a Serum-style error (concretely: if it does not implement ErrorInterface),
// it will be coerced into one, by use of the Standardize function.
// (We consider this coersion appropriate to perform immediately,
// because otherwise the resulting value would fail to round-trip through serialization.)
//
// Errors:
//
//   - param: ecode -- the error code to construct.
//
func Errorf(ecode string, fmtPattern string, args ...interface{}) error {
	// Literally use stdlib Errorf, then extract from its results, because replicating its parse for '%w' is nontrivial.
	fmtErr := fmt.Errorf(fmtPattern, args...)
	return &ErrorValue{Data{
		Code:    ecode,
		Message: fmtErr.Error(),
		Cause:   Standardize(Cause(fmtErr)),
	}}
}

// Standardize returns a value that's guaranteed to be a Serum-style error,
// and use the concrete type of *ErrorValue from this package.
//
// This isn't often necessary to use, because all the functions in this package
// accept any error implementation and figure out how to do the right thing --
// but is provided for your convenience if needed for creating new errors
// or just moving things into a standard memory layout for some reason.
//
// If given a value that implements the Serum interfaces,
// all data will be copied, using those interfaces to access it.
//
// If given a golang error that's not a Serum-style error at all,
// the same procedure is followed: a new value will be created,
// where the code is set to what `serum.Code` returns on the old value; etc.
// (In practice, this means you'll end up with an ErrorValue that contains a
// code string that is prefixed with "golang-bestguess-"; etc.)
//
// If given a value that is already of type *ErrorValue, it is returned unchanged.
//
// This function returns ErrorInterface rather than concretely *ErrorValue,
// to reduce the chance of creating "untyped nil" problems in practical usage,
// but it is valid to directly cast the result to *ErrorValue if you wish.
func Standardize(other error) ErrorInterface {
	if other == nil {
		return nil
	}
	if cast, ok := other.(*ErrorValue); ok {
		return cast
	}
	return &ErrorValue{Data{
		Code:    Code(other),
		Message: Message(other),
		Details: Details(other),
		Cause:   Standardize(Cause(other)),
	}}
}
