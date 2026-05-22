package main

import (
	"net/http"
	"strings"
	"time"
)

// This middleware checks the Authorization header to get the user token,
// then checks if the token is correctly formatted, if the token exists in
// the sessions and the session is valid, and, finally, if the user exists.
// If everything is okay, it lets the request pass to the next handler.
// If not, it will return a 401 Unauthorized status code.
func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token = token[7:]
		if sessions[token] == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if sessions[token].Expires.Before(time.Now()) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, exists := allUsers[sessions[token].Username]
		if !exists {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
