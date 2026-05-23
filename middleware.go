package main

import (
	"net/http"
	"strings"
)

// This middleware checks the Authorization header to get the user token,
// then checks if the token is correctly formatted, if the token exists in
// the sessions and the session is valid, and, finally, if the user exists.
// If everything is okay, it lets the request pass to the next handler.
// If not, it will return a 401 Unauthorized status code.
func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		// check that there is a token in the auth header
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// we have a token
		token = token[7:]
		// check if there is a session that exists for this token
		_, err := repository.GetSession(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// user has a valid session, and is a user
		next(w, r)
	}
}

func adminRoleRequired(next http.HandlerFunc) http.HandlerFunc {
	return authRequired(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = token[7:]

		// by this point, the auth header exists, there is a non-expired session for the user and that user exists

		userRole, err := repository.GetUserRoleFromSession(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if userRole != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}
