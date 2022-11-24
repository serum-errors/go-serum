the go-serum library
====================

The `go-serum` library is an easy implementation of the [Serum Errors Specification](https://github.com/serum-errors/serum-spec) for use in Golang development.

The Serum errors spec is meant to be a "just enough" spec -- easy to adopt, easy to extend, easy to describe.
It specifies enough to be meaningful, but not so much that it becomes complicated.

This implementation is meant to be similarly "just enough":

- it's a golang type;
- it implements interfaces for `error` and also interfaces that let tools like [go-serum-analyzer](https://github.com/serum-errors/go-serum-analyzer) do static analysis for you;
- it implements serialization to JSON;
- and that's about it.

The library is written with the trust you can put those basics to good use.

The library also provides package-scope functions which can be used to access any of the attributes of a Serum-convention error --
`Code`, `Message`, `Details`, etc -- which also work on any golang `error`, making incremental adoption easy.


Examples!
---------

We'll give a couple of examples of creating errors, in increasing order of complexity.
Then, at the bottom, a quick example of how we suggest _handling_ errors.

This is golang code to produce an error:

```go
serum.Errorf("myapp-error-foobar", "this is a foobar error, with more info: %s", "somethingsomething")
```

If you print the result as JSON, you'll get:

```json
{
	"code": "myapp-error-foobar",
	"message": "this is a foobar error, with more info: somethingsomething"
}
```

You can use the `%w` syntax to wrap other errors, too -- just like with standard `fmt.Errorf`:

```go
serum.Errorf("myapp-error-frobnoz", "this is a bigger error, with cause: %w", otherErrorAbove)
```

If you print the result as JSON, you'll get:

```json
{
	"code": "myapp-error-frobnoz",
	"message": "this is a bigger error, with cause: this is a foobar error, with more info: somethingsomething",
	"cause": {
		"code": "myapp-error-foobar",
		"message": "this is a foobar error, with more info: somethingsomething"
	}
}
```

(Note that the templating of messages is resolved in advance at all times.
So typically, to a user, you just print the outermost message.)

What's above is just the shorthand API.

You can also produce richer errors:

```go
serum.Error("myapp-error-jobnotfound",
	serum.WithMessageTemplate("job ID {{ID}} not found"),
	serum.WithDetail("ID", "asdf-qwer-zxcv"),
)
```

The result of this, as JSON, is:

```json
{
	"code": "myapp-error-jobnotfound",
	"message": "job ID asdf-qwer-zxcv not found",
	"details": {
		"ID": "asdf-qwer-zxcv"
	}
}
```

Notice how with this syntax, you could attach details to the error.
This makes for easier programmatic transmission of complex, rich errors.
The brief templating syntax -- `{{this}}` -- just substitutes in values.
It means the message prepared for human readers can still include the details, without the developer having to repeat themself too much.

(Note that you can use `WithMessageLiteral` instead of `WithMessageTemplate`, if you don't want to use the templating system at all!)

The templating language is not rich (intentionally!  You shouldn't be doing complex logic during error production!),
but it does support a few critical things, like quoting:

```go
serum.Error("myapp-error-withquotedstuff",
	serum.WithMessageTemplate("message detail {{thedetail | q}} should be quoted"),
	serum.WithDetail("thedetail", "whee! wow!"),
)
```

(A pipe character -- `|` -- is how we insert a formatting directive; and "q" means "quote this".)

If you stringify this (i.e. with just `.Error()`), you'll get:

```text
myapp-error-withquotedstuff: message detail "whee! wow!" should be quoted
```

When we serialize this one as JSON, notice that the the value in the details map is unquoted (it's still a clear value on its own!), but the composed message is quoted (which then ends up escaped in JSON):

```json
{
	"code": "demo-error-withquotes",
	"message": "message detail \"whee! wow!\" should be quoted",
	"details": {
		"thedetail": "whee! wow!"
	}
}
```

Now how do we handle all these errors?
Easy: the typical way is to switch on their "code" field.
That looks like this:

```go
switch serum.Code(theError) {
	case "myapp-error-foobar":
		// ...handle foobar...
	case "myapp-error-frobnoz":
		// ...handle frobnoz...
	case "myapp-error-jobnotfound":
		// ...handle jobnotfound...
	default:
		panic("unhandled error :(") // shouldn't happen because go-serum-analyzer can catch it at compile time!  :D
}
```

Status
------

This library is considered in "beta" status.  Please try it out, and see if it suits your needs.

The API may change in the future, as we discover more about how to make it the smoothest it can be.
However, we will take any changes carefully, as we do understand that this library may end up at the base of deep dependency tress;
we will definitely aim to minimize breaking changes, provide smooth migration windows,
and generally avoid creating any painful "diamond problems" in dependency graphs.


But this is too heavyweight...
------------------------------

Well, most people don't say that :)  It depends only on the standard library!

But yes, we allowed several dependencies from the standard library to creep in.
Namely, `encoding/json` and `reflect`.
(And `strings` and `strconv`, though usually people don't mind that.)

**You can most definitely implement the Serum conventions without such dependencies.**
But this library uses them.
Using them makes it possible for us to give you something that's easier to use than if we had avoided those stdlib packages.
Especially in the case of JSON: playing nice with stdlib's `encoding/json` just feels enormously valuable.

If you would like a variant of this library without those dependencies,
you can write another package that does exactly what you want!  It's certainly possible.
Or, patches/PRs for adding build tags to conditionally remove those features from this library would likely be accepted as well.


License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT
