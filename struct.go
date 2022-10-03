package serum

// ErrorValue is a concrete type that implements the Serum conventions for errors.
//
// It can contain message and details fields in addition to the essential "code" field,
// and implements convenient features like automatic synthesis of a good message for golang `Error() string`,
// as well as supporting json marshalling and unmarshalling.
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

func (e *ErrorValue) Code() string         { return e.Data.Code }
func (e *ErrorValue) Message() string      { return e.Data.Message }
func (e *ErrorValue) Details() [][2]string { return e.Data.Details }
func (e *ErrorValue) Unwrap() error        { return e.Data.Cause }
func (e *ErrorValue) Error() string        { return SynthesizeString(e) }
