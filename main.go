package main

import (
	"fmt"
	"log"
	"net/http"
)

var allData []ShoppingList

func seedShoppingList() {
	lists := []ShoppingList{
		{ID: 1, Name: "Saturday shopping list", Items: []string{"bread", "ice cream", "milk", "pasta", "toothpaste", "eggs", "soap", "detergent"}},
		{ID: 2, Name: "Hamburger night", Items: []string{"beef patties", "burger rolls", "eggs", "bacon", "tomatoes", "sliced cheese", "bbq sauce", "beetroot", "butter", "lettuce"}},
	}

	for _, l := range lists {
		allData = append(allData, l)
	}

	log.Println("seeded shopping lists")
}

func main() {
	port := 8888

	http.HandleFunc("POST /v1/lists", handleCreateList)
	http.HandleFunc("GET /v1/lists", handleFetchAllLists)
	http.HandleFunc("DELETE /v1/lists/{id}", handleDeleteList)
	http.HandleFunc("PUT /v1/lists/{id}", handleUpdateList)
	http.HandleFunc("PATCH /v1/lists/{id}", handlePartialUpdateList)
	http.HandleFunc("GET /v1/lists/{id}", handleFetchListById)

	seedShoppingList()

	log.Printf("listening on port :%d", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
