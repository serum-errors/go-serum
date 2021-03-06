package serum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// ToJSON is a helper function to turn any error into JSON.
// It is suitable to use in implementing `encoding/json.Marshaler.MarshalJSON` if implementing your own error types.
// If is also suitable for freestanding use on any error value (even non-Serum error values).
//
// If the error is a Serum error (per ErrorInterface),
// we'll serialize it completely, including message, details, code, and cause, all distinctly.
// If the error is not a Serum error, we'll serialize it anyway,
// but fudge as necessary to produce a result that will at least be Serum serial spec compliant and able to be deserialized as a Serum error
// (the golang type will appear as part of the code, and the `Error() string` will be used as if a message, and `errors.Unwrap` will be used to find a cause, etc).
func ToJSON(err error) ([]byte, error) {
	// Error handling throughout this function would appear lax; it is not.
	// Where we are encoding to a buffer, and know we are handling only strings, errors from encode are not really possible, and so the branch to check is omitted.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	buf.WriteString(`{"code":`)
	encoder.Encode(Code(err))
	if _, ok := err.(ErrorInterface); ok {
		if e2, ok := err.(ErrorInterfaceWithMessage); ok {
			msg := e2.Message()
			if msg != "" {
				buf.WriteString(`, "message":`)
				encoder.Encode(msg)
			}
		}
	} else {
		buf.WriteString(`, "message":`)
		encoder.Encode(err.Error())
	}
	if details := Details(err); details != nil {
		buf.WriteString(`, "details":`)
		pairs(details).marshalJSON(&buf)
	}
	if cause := errors.Unwrap(err); cause != nil {
		buf.WriteString(`, "cause":`)
		if err := encoder.Encode(cause); err != nil {
			return nil, err
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// ---

func (e *ErrorStruct) UnmarshalJSON(b []byte) error {
	var target struct {
		// Yes, this is almost ErrorStruct itself, but:
		// - this uses the unexported pairs type,
		// - this needs a concrete type for the cause, or things don't fly right,
		// - and not having the json tags on the ErrorStruct type just seems wise, since it doesn't actually use them and there's no sense in misleading a reader.
		TheCode    string       `json:"code"`
		TheMessage string       `json:"message,omitempty"`
		TheDetails pairs        `json:"details,omitempty"`
		TheCause   *ErrorStruct `json:"cause,omitempty"`
	}
	if err := json.Unmarshal(b, &target); err != nil {
		return err
	}
	e.TheCode = target.TheCode
	e.TheMessage = target.TheMessage
	e.TheDetails = target.TheDetails
	e.TheCause = target.TheCause
	return nil
}

func (e *ErrorStruct) MarshalJSON() ([]byte, error) {
	return ToJSON(e)
}

// ---

// MarshalJSON on the pairs type is a kludge to get ordered map behavior.
func (a pairs) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := a.marshalJSON(&buf)
	return buf.Bytes(), err
}

// marshalJSON on the pairs type is a kludge within a kludge because the stdlib interfaces for this are ridiculous.
func (a pairs) marshalJSON(buf *bytes.Buffer) error {
	buf.WriteByte('{')
	encoder := json.NewEncoder(buf)
	for i := range a {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := encoder.Encode(a[0]); err != nil {
			return err
		}
		buf.WriteByte(':')
		if err := encoder.Encode(a[1]); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

func (a *pairs) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	if tok, err := dec.Token(); err != nil {
		return err
	} else if tok != '{' {
		return fmt.Errorf("deserializing a serum error: details field must be a map")
	}
	for {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return nil
		}
		key := token.(string)
		token, err = dec.Token()
		if err != nil {
			return err
		}
		if value, ok := token.(string); !ok {
			return fmt.Errorf("deserializing a serum error: only strings are permitted in details map values")
		} else {
			*a = append(*a, [2]string{key, value})
		}
	}
}
