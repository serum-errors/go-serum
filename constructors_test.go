package serum_test

import (
	"encoding/json"
	"fmt"

	"github.com/serum-errors/go-serum"
)

func ExampleErrorf() {
	const ErrFoobar = "demo-error-foobar"
	err := serum.Errorf(ErrFoobar, "freetext goes here (%s)", "and can interpolate")

	fmt.Printf("the error as a string:\n\t%v\n", err)
	jb, jsonErr := json.MarshalIndent(err, "\t", "\t")
	if jsonErr != nil {
		panic(jsonErr)
	}
	fmt.Printf("the error as json:\n\t%s\n", jb)

	// Output:
	// the error as a string:
	// 	demo-error-foobar: freetext goes here (and can interpolate)
	// the error as json:
	// 	{
	// 		"code": "demo-error-foobar",
	// 		"message": "freetext goes here (and can interpolate)"
	// 	}
}
