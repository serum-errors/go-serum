/*
The serum package provides helper functions and handy types for error handling
that works in accordance with the Serum Errors Convention.

You don't need to use this package to implement the Serum Errors Convention!
(A key goal of the Convention is that you *do not* need any specific implementation;
the convention is based on the serial forms,
and in Golang, even all of the static analysis tooling is based on interfaces.)
However, you may find it handy.
*/
package serum

import (
	"path"
	"reflect"
	"sort"
	"strings"
)

// ErrorInterface is the minimal interface that must be implemented to be a Serum error.
// This is also the interface that go-serum-analyzer is looking for, if you use that tool
// (although not by name; matching the pattern is sufficient).
//
// This interface is not often seen in user code.
// Usually, we still recommend user code mostly refers to "error", as is normal in golang.
// Functions throughout this package will accept and return error as a parameter,
// and apply serum behaviors to those values by use of feature detection,
// which removes all need to refer to this interface in user code.
type ErrorInterface interface {
	error
	Code() string
}

// Code will access and return the error code for any Serum-style error.
//
// This function takes the general "error" type and feature-detects for Serum behaviors,
// but still has fallback behaviors for any error value.
//
// If the given error is _not_ recognizably Serum-styled, a code string will be invented on the fly.
// This invented code string will have the prefix "bestguess-golang-" followed by a munge of the golang type name.
// This fallback is meant to be minimally functional and help find the source of coding mistakes that lead to its creation,
// but should not be seen in a well-formed program.
func Code(err error) string {
	// If it's Serum: great.
	if e2, ok := err.(ErrorInterface); ok {
		code := e2.Code()
		// Quick sanity check we're not getting empty string.  Propagate nonsense instead of silence if so; not good, but should help the problem be noticed faster.
		if code == "" {
			return "?!"
		}
		return code
	}
	// If it's not: we'll attempt to do something useful from the golang type name.
	// "Useful" might be a stretch, but this should at least help a developer find their questionable code quickly.
	// We do not commit to the stability of this string.  A program of well-defined errors should not encounter this path.
	rt := reflect.TypeOf(err).Elem()
	return "bestguess-golang-" + path.Base(rt.PkgPath()) + "-" + rt.Name()
}

// ---

type ErrorInterfaceWithMessage interface {
	ErrorInterface
	Message() string
}

type ErrorInterfaceWithDetailsOrdered interface {
	ErrorInterface
	Details() [][2]string
}

type ErrorInterfaceWithDetailsMap interface {
	ErrorInterface
	Details() map[string]string
}

type ErrorInterfaceWithCause interface {
	ErrorInterface
	Unwrap() error
}

// DetailsMap returns the details of an error as a map.
//
// This function takes the general "error" type and feature-detects for Serum behaviors,
// but still has fallback behaviors for any error value.
//
// Note that you may wish to use the Details function instead,
// if the original ordering of the details fields is important;
// because it uses golang maps, this function is not order-preserving.
//
// If the given error is not recognizably Serum-styled,
// this function returns an empty map.
//
// The map should not be mutated; it may be the original memory from the error value.
func DetailsMap(err error) map[string]string {
	if e2, ok := err.(ErrorInterfaceWithDetailsMap); ok {
		return e2.Details()
	}
	if e2, ok := err.(ErrorInterfaceWithDetailsOrdered); ok {
		l := e2.Details()
		m := make(map[string]string, len(l))
		for _, ent := range l {
			m[ent[0]] = ent[1]
		}
		return m
	}
	return map[string]string{}
}

// Details returns the details key-values of an error as a slice of pairs of strings.
//
// This function takes the general "error" type and feature-detects for Serum behaviors,
// but still has fallback behaviors for any error value.
//
// If the given error is not recognizably Serum-styled,
// this function returns nil.
//
// Note that you may also be able to use the DetailsMap to get the same content as a golang map, for convenience,
// but be aware that access mechanism does not support order-preservation, and may often be slightly slower performance.
//
// The result should not be mutated; it may be the original memory from the error value.
func Details(err error) [][2]string {
	if e2, ok := err.(ErrorInterfaceWithDetailsOrdered); ok {
		return e2.Details()
	}
	if e2, ok := err.(ErrorInterfaceWithDetailsMap); ok {
		m := e2.Details()
		l := make([][2]string, len(m))
		for k, v := range m {
			l = append(l, [2]string{k, v})
		}
		sort.Sort(pairs(l))
		return l
	}
	return nil
}

type pairs [][2]string

func (a pairs) Len() int           { return len(a) }
func (a pairs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a pairs) Less(i, j int) bool { return a[i][0] < a[j][0] }

// Detail gets a detail value out of an error, or returns empty string if there isn't one.
// It's functionally equal to `DetailsMap()[whichDetail]`, but may be more efficient.
func Detail(err error, whichDetail string) string {
	if e2, ok := err.(ErrorInterfaceWithDetailsMap); ok {
		return e2.Details()[whichDetail]
	}
	if e2, ok := err.(ErrorInterfaceWithDetailsOrdered); ok {
		for _, ent := range e2.Details() {
			if ent[0] == whichDetail {
				return ent[1]
			}
		}
	}
	return ""
}

// ---

// ... below might belong in a different package; they're for helping you write types.

// SynthesizeString generates a string for an error, suitable for return as the golang `Error() string` result.
// SynthesizeString will detect properties of a Serum error, and synthesize a string using them.
// The string will contain the code, the message, and the string of the cause if present,
// in roughly the form "{code}[: {message}][: caused by: {cause}]".
// Entries from a details map will not be present (unless the message includes them), as per the Serum standard's recommendation.
//
// You can use this function to implement the `Error() string` method of a Serum error type conveniently.
//
// The resultant string is hoped to be human-readable.
// It is not expected to be mechanically parsible.
// The form is primarily meant to match Golang community norms; it is not a Serum convention.
//
// The exact behavior of this function may change over time.
// For example, currently, it disregards all linebreaks (it neither strips nor introduces them itself),
// but in the future, if a Serum convention for multiline errors is introduced, then this function will likely change in behavior to match.
func SynthesizeString(err ErrorInterface) string {
	var sb strings.Builder
	sb.WriteString(err.Code())
	if e2, ok := err.(ErrorInterfaceWithMessage); ok {
		msg := e2.Message()
		if msg != "" {
			sb.WriteString(": ")
			sb.WriteString(msg)
		}
	}
	if e2, ok := err.(ErrorInterfaceWithCause); ok {
		cause := e2.Unwrap()
		if cause != nil {
			sb.WriteString(": caused by: ")
			sb.WriteString(cause.Error())
		}
	}
	return sb.String()
}

/*
Not actually sure the following is valuable enough to take on a templating package dependency.

// ErrorWithMessageTemplateInterface describes an error that has a message template.
// These aren't a Serum Convention standard; it's a convenience feature.
// The template can refer to keys in the details map.
// Having a template attached to a type via a constant method is a way
// to avoid having to write a custom constructor function.
type ErrorInterfaceWithMessageTemplate interface {
	ErrorInterface
	Template() string
}

// SynthesizeMessage produces a message string from an error.
// (Note: this is not the entire string that describes an error; see SynthesizeString.)
// If the error has a message template (per ErrorInterfaceWithMessageTemplate), the template will be evaluated;
// if there is no template, a "k1=v1, k2=v2" string will be produced as a fallback.
//
// If there's already a message (per ErrorInterfaceWithMessage), this function disregards it.
// This is because this function is meant primarily to help implement the Message function;
// so, to call Message would be prone to result in endless loops in practice.
func SynthesizeMessage(err ErrorInterface) string {
	panic("not yet implemented")
}

*/
