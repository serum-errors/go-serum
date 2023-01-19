package serum

// ErrorValue is a concrete type that implements the Serum conventions for errors.
//
// It can contain message and details fields in addition to the essential "code" field,
// and implements convenient features like automatic synthesis of a good message for golang `Error() string`,
// as well as supporting json marshalling and unmarshalling.
//
// Accessor methods can be used to inspect the values inside this type,
// but typically, the package-scope functions in serum should be used instead --
// `serum.Code`, `serum.Message`, `serum.Details`, etc --
// because they are easier to use without refering to any concrete types.
// (Using the package-scope functions will save you from any syntactical line-noise of casting!)
//
// The fields of this type are exported, but mutating them is inadvisable.
// (The go-serum-analyzer tool becomes much less useful if you do so;
// it does not support tracking the effects of such mutations.)
type ErrorValue struct {
	Data
}

// Data is the body of the ErrorValue type.
//
// It is a separate type mainly for naming purposes
// (it allows us to have the same name for fields here as we use for the accessor methods on ErrorValue).
//
// Most user code will not see this type.
// Although it is exported, and referencing it is allowed, it is not usually necessary.
// User code can construct these values if desired,
// but using constructor functions from the go-serum package is often syntactically easier.
// User code may access these values directly if it's known that the code is handling ErrorValue concretely,
// but most code is not writen in such a way, and the serum accessor functions are used instead.
type Data struct {
	Code    string
	Message string
	Details [][2]string
	Cause   ErrorInterface
}

// Code returns the Serum errorcode.  Use the `serum.Code` package function to access this without referring to the concrete type.
func (e *ErrorValue) Code() string { return e.Data.Code }

// Message returns the Serum message.  Use the `serum.Message` package function to access this without referring to the concrete type.
func (e *ErrorValue) Message() string { return e.Data.Message }

// Details returns the Serum details key-values.  Use the `serum.Details` or `serum.DetailsMap` package function to access this without referring to the concrete type.
func (e *ErrorValue) Details() [][2]string { return e.Data.Details }

// Unwrap returns the Serum cause.  Use the `serum.Cause` package function, or the golang `errors.Unwrap` function, to access this without referring to the concrete type.
func (e *ErrorValue) Unwrap() error { return e.Data.Cause }

// Error implements the golang error interface.  The returned string will contain the code, the message if present, and the string of the cause.  Per Serum convention, it does not include any of the details fields.
func (e *ErrorValue) Error() string { return SynthesizeString(e) }

// Is implements errors.Is so that it works for non-serum errors
// This allows non-serum-aware packages to take serum errors if they use errors.Is for error comparisons
func (e *ErrorValue) Is(target error) bool {
	if Code(e) != Code(target) {
		return false
	}
	if Message(e) != Message(target) {
		return false
	}
	// We don't check detail map because it _should_ be synthesized into message.
	// We should not unwrap here because errors.Is handles unwrapping.
	return true
}
 