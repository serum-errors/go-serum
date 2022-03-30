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


Status
------

This library is considered in "beta" status.  Please try it out, and see if it suits your needs.

The API may change in the future, as we discover more about how to make it the smoothest it can be.
However, we will take any changes carefully, as we do understand that this library may end up at the base of deep dependency tress;
we will definitely aim to minimize breaking changes, provide smooth migration windows,
and generally avoid creating any painful "diamond problems" in dependency graphs.


But this is too heavyweight...
------------------------------

Yes, we allowed several dependencies from the standard library to creep in.
Namely, `encoding/json` and `reflect`.

**You can most definitely implement the Serum conventions without such dependencies.**

If you would like a variant of this library without those dependencies,
you can write another package that does exactly what you want,
or,
patches/PRs for adding build tags to conditionally remove those features from this library would likely be accepted as well.


License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT
