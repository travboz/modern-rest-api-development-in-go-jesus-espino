package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var allData []ShoppingList

func seedShoppingList() error {
	err := repository.Empty()
	if err != nil {
		return err
	}

	lists := []ShoppingList{
		{ID: 1, Name: "Saturday shopping list", Items: []string{"bread", "ice cream", "milk", "pasta", "toothpaste", "eggs", "soap", "detergent"}},
		{ID: 2, Name: "Hamburger night", Items: []string{"beef patties", "burger rolls", "eggs", "bacon", "tomatoes", "sliced cheese", "bbq sauce", "beetroot", "butter", "lettuce"}},
	}

	for _, l := range lists {
		err := repository.CreateNewShoppingList(&l)
		if err != nil {
			return err
		}
	}

	return nil
}

var repository *Repository

func main() {
	var err error
	repository, err = NewRepository("./data/database.db")
	if err != nil {
		log.Println("Unable to open the database:", err.Error())
		os.Exit(1)
	}

	if err := repository.Init(); err != nil {
		log.Println("Unable to initialise the database:", err.Error())
		os.Exit(1)
	}

	err = seedShoppingList()
	if err != nil {
		log.Println("Unable to seed the database:", err.Error())
		os.Exit(1)
	}

	log.Println("Successfully seeded shopping lists")

	port := 8888

	http.HandleFunc("GET /v1/lists", handleFetchAllLists)
	http.HandleFunc("POST /v1/lists", adminRoleRequired("admin", handleCreateList))
	http.HandleFunc("GET /v1/lists/{id}", authRequired(handleFetchListById))
	http.HandleFunc("PUT /v1/lists/{id}", adminRoleRequired("admin", handleUpdateList))
	http.HandleFunc("DELETE /v1/lists/{id}", adminRoleRequired("admin", handleDeleteList))
	http.HandleFunc("PATCH /v1/lists/{id}", adminRoleRequired("admin", handlePartialUpdateList))
	http.HandleFunc("POST /v1/lists/{id}/push", adminRoleRequired("admin", handleListPush))

	http.HandleFunc("POST /login", handleLogin)

	log.Printf("listening on port :%d", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
