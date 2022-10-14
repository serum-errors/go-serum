package serum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ToJSON is a helper function to turn any error into JSON.
// It is suitable to use in implementing `encoding/json.Marshaler.MarshalJSON` if implementing your own error types,
// and it is used for that purpose in the ErrorValue type provided by this package.
// If is also suitable for freestanding use on any error value (even non-Serum error values).
//
// If the error is a Serum error (per ErrorInterface),
// we'll serialize it completely, including message, details, code, and cause, all distinctly.
// If the error is not a Serum error, we'll serialize it anyway,
// but fudge as necessary to produce a result that will at least be Serum serial spec compliant and able to be deserialized as a Serum error.
//
// In fudge mode: the golang type will appear as part of the serum code;
// the `Error() string` will be used as a message;
// `errors.Unwrap` will be used to find a cause; etc.
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
	if cause := errors.Unwrap(err); cause != nil && !isEmptyValue(reflect.ValueOf(cause)) {
		buf.WriteString(`, "cause":`)
		if causeJson, err := ToJSON(cause); err != nil {
			return nil, err
		} else {
			buf.Write(causeJson)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// ---

func (e *ErrorValue) UnmarshalJSON(b []byte) error {
	var target struct {
		// Yes, this is almost serum.Data itself, but:
		// - this uses the unexported 'pairs' type,
		// - this needs a concrete type for the cause, or things don't fly right,
		// - and not having the json tags on the serum.Data type just seems wise, since it doesn't actually use them and there's no sense in misleading a reader.
		Code    string      `json:"code"`
		Message string      `json:"message,omitempty"`
		Details pairs       `json:"details,omitempty"`
		Cause   *ErrorValue `json:"cause,omitempty"`
	}
	if err := json.Unmarshal(b, &target); err != nil {
		return err
	}
	e.Data.Code = target.Code
	e.Data.Message = target.Message
	e.Data.Details = target.Details
	e.Data.Cause = target.Cause
	return nil
}

func (e *ErrorValue) MarshalJSON() ([]byte, error) {
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
		if err := encoder.Encode(a[i][0]); err != nil {
			return err
		}
		buf.WriteByte(':')
		if err := encoder.Encode(a[i][1]); err != nil {
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

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}
