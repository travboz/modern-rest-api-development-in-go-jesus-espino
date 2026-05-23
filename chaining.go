package main

import (
	"net/http"
	"slices"
)

// Custom middleware chaining type to allow for dependency-free chaining (as in no external packages)
type middlewareChain []func(http.HandlerFunc) http.HandlerFunc

func (c middlewareChain) then(h http.HandlerFunc) http.HandlerFunc {
	// wrap the middleware from the end backwards.
	// so, a chain like this: {requestID, logRequest, authenticate} would look like:
	// 		requestID(	logRequest(	authenticate(h)	)	)
	for _, mw := range slices.Backward(c) {
		// wrap h in all the middleware backwards
		h = mw(h)
	}

	return h
}

func (c middlewareChain) thenFunc(h http.HandlerFunc) http.Handler {
	return c.then(h)
}
