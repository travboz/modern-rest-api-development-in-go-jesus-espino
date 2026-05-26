package main

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
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

func adminRoleRequired(roleRequired string, next http.HandlerFunc) http.HandlerFunc {
	return authRequired(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = token[7:]

		// by this point, the auth header exists, there is a non-expired session for the user and that user exists

		role, err := repository.GetUserRoleFromSession(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if role != roleRequired {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}

// We create our CORS middleware with the API domain and port, and pass a list of the allowed methods and headers.
// Also, we define MaxAge there to allow caching pre-flight requests for 300 seconds. We use the middleware created
// to wrap our mux variable, adding CORS to all our APIs.
func corsWrapper(mux *http.ServeMux) http.Handler {
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
		},
		MaxAge: 300,
	})

	return corsMiddleware.Handler(mux)
}
