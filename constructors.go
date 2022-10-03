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

// Error is a constructor for new Serum-style error values,
// supporting use of templated messages, and attachment of details,
// causes, and the enter suite of Serum features.
//
// The error code parameter is required, and all other parameters are optional.
// The `serum.With*()` functions are used to create descriptions of messages
// and details attachements, and these are then provided as varargs as desired.
//
// See the examples for usage.
//
// Errors:
//
//   - param: ecode -- the error code to construct.
//
func Error(ecode string, params ...WithConstruction) error {
	res := &ErrorValue{Data{
		Code: ecode,
	}}
	var doLast WithConstruction
	for _, param := range params {
		switch {
		case param.msgLiteral != "":
			res.Data.Message = param.msgLiteral
		case param.msgTemplate != nil:
			doLast = param // Need to get all the details assembled first.
		case param.detailKey != "":
			res.Data.Details = append(res.Data.Details, [2]string{param.detailKey, param.detailValue})
		case param.cause != nil:
			res.Data.Cause = param.cause
		}
	}
	if doLast.msgTemplate != nil {
		res.Data.Message = interpolate(doLast.msgTemplate, res.Data.Details)
	}
	return res
}

// WithMessageTemplate is part of the system for constructing an error
// with the serum.Error function.
//
// WithMessageTemplate describes how to produce a message for the error,
// and allows values from the error's details to be incorporated into the message smoothly.
//
// The templates used are very simple: "{{x}}" will look up the detail labelled "x"
// and place the corresponding value in that position in the string.
// The templates will never error; if a detail is not found with the given label,
// then the output will just contain the template syntax and missing label
// (e.g., "{{x}}" will be emitted as output).
//
// See the examples of the Error function for complete demonstrations of usage.
func WithMessageTemplate(tmpl string) WithConstruction {
	return WithConstruction{msgTemplate: parse(tmpl)}
}

// WithMessageLiteral is part of the system for constructing an error
// with the serum.Error function.
//
// In contrast with WithMessageTemplate, this string is always passed
// verbatim into the message of the error.
func WithMessageLiteral(s string) WithConstruction {
	return WithConstruction{msgLiteral: s}
}

// WithDetail is part of the system for constructing an error
// with the serum.Error function.
// It allows attaching simple key:value string pairs to the error.
//
// In addition to being stored as details in Serum convention
// (e.g., these values will be serialized as entries in a map when serializing the error),
// the WithMessageTemplate system can reference the detail values.
//
// See the examples of the Error function for complete demonstrations of usage.
func WithDetail(key, value string) WithConstruction {
	return WithConstruction{detailKey: key, detailValue: value}
}

// WithDetail is part of the system for constructing an error
// with the serum.Error function.
// It can accept any golang error value and will attach it as a cause
// to the newly produced Serum error.
//
// As with Errorf's behavior when attaching causes, if the given error
// is not already Serum-style error, it will be coerced into one.
// This may result in a generated error code, which is prefixed with
// the string "bestguess-golang-" and some type name information.
func WithCause(cause error) WithConstruction {
	return WithConstruction{cause: Standardize(cause)}
}

// WithConstruction is a data carrier type used as part of the Error constructor system.
// It is not usually seen directly in user code; only passed between the
// With*() functions, and directly into the Error constructor function.
//
// See the examples of the Error function for complete demonstrations of usage.
type WithConstruction struct {
	msgLiteral  string
	msgTemplate []parsed
	detailKey   string
	detailValue string
	cause       ErrorInterface
}
