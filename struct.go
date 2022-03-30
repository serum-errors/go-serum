package serum

type ErrorStruct struct {
	TheCode    string
	TheMessage string
	TheDetails [][2]string
	TheCause   ErrorInterface
}

func (e *ErrorStruct) Code() string         { return e.TheCode }
func (e *ErrorStruct) Message() string      { return e.TheMessage }
func (e *ErrorStruct) Details() [][2]string { return e.TheDetails }
func (e *ErrorStruct) Unwrap() error        { return e.TheCause }
func (e *ErrorStruct) Error() string        { return SynthesizeString(e) }
