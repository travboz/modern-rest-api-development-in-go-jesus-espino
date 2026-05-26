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

	mux := http.NewServeMux()

	SetupRoutes(mux)
	handler := corsWrapper(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	log.Printf("listening on port :%d", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
