package serum_test

import (
	"encoding/json"
	"fmt"
	"strconv"

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

func ExampleError() {
	const ErrJobNotFound = "demo-error-job-not-found"
	constructor := func(param int) error {
		return serum.Error(ErrJobNotFound,
			serum.WithMessageTemplate("job ID {{ID}} not found"),
			serum.WithDetail("ID", strconv.Itoa(param)),
		)
	}
	err := constructor(12)
	fmt.Printf("the error as a string:\n\t%v\n", err)
	jb, jsonErr := json.MarshalIndent(err, "\t", "\t")
	if jsonErr != nil {
		panic(jsonErr)
	}
	fmt.Printf("the error as json:\n\t%s\n", jb)

	// Output:
	// the error as a string:
	// 	demo-error-job-not-found: job ID 12 not found
	// the error as json:
	// 	{
	// 		"code": "demo-error-job-not-found",
	// 		"message": "job ID 12 not found",
	// 		"details": {
	// 			"ID": "12"
	// 		}
	// 	}
}

func ExampleErrorWithTemplateQuoting() {
	err := serum.Error("demo-error-withquotes",
		serum.WithMessageTemplate("message detail {{thedetail|q}} should be quoted"),
		serum.WithDetail("thedetail", "whee! wow!"),
	)
	fmt.Printf("the error as a string:\n\t%v\n", err)
	jb, jsonErr := json.MarshalIndent(err, "\t", "\t")
	if jsonErr != nil {
		panic(jsonErr)
	}
	fmt.Printf("the error as json:\n\t%s\n", jb)

	// Output:
	// the error as a string:
	// 	demo-error-withquotes: message detail "whee! wow!" should be quoted
	// the error as json:
	// 	{
	// 		"code": "demo-error-withquotes",
	// 		"message": "message detail \"whee! wow!\" should be quoted",
	// 		"details": {
	// 			"thedetail": "whee! wow!"
	// 		}
	// 	}
}
