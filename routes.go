package main

import "net/http"

func SetupRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /v1/lists", addCacheHeaders(authRequired(handleFetchAllLists)))
	mux.HandleFunc("POST /v1/lists", adminRoleRequired("admin", handleCreateList))
	mux.HandleFunc("GET /v1/lists/{id}", authRequired(handleFetchListById))
	mux.HandleFunc("PUT /v1/lists/{id}", adminRoleRequired("admin", handleUpdateList))
	mux.HandleFunc("DELETE /v1/lists/{id}", adminRoleRequired("admin", handleDeleteList))
	mux.HandleFunc("PATCH /v1/lists/{id}", adminRoleRequired("admin", handlePartialUpdateList))
	mux.HandleFunc("POST /v1/lists/{id}/push", adminRoleRequired("admin", handleListPush))

	mux.HandleFunc("POST /v1/login", handleLogin)
}
