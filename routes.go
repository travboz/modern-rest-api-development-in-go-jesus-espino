package main

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/travboz/modern-rest-api-dev/shopping-list-api/docs" // for swagger api docs
)

func SetupRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /v1/lists", addCacheHeaders(authRequired(handleFetchAllLists)))
	mux.HandleFunc("POST /v1/lists", adminRoleRequired("admin", handleCreateList))
	mux.HandleFunc("GET /v1/lists/{id}", authRequired(handleFetchListById))
	mux.HandleFunc("PUT /v1/lists/{id}", adminRoleRequired("admin", handleUpdateList))
	mux.HandleFunc("DELETE /v1/lists/{id}", adminRoleRequired("admin", handleDeleteList))
	mux.HandleFunc("PATCH /v1/lists/{id}", adminRoleRequired("admin", handlePartialUpdateList))
	mux.HandleFunc("POST /v1/lists/{id}/push", adminRoleRequired("admin", handleListPush))

	mux.HandleFunc("POST /v1/login", handleLogin)

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8888/swagger/doc.json"),
	))

}
