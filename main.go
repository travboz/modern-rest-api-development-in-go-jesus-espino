package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	lru "github.com/hashicorp/golang-lru/v2"
)

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

var (
	repository RepositoryInterface
	listsCache *lru.Cache[string, *ShoppingList]
)

func main() {
	var err error
	listsCache, err = lru.New[string, *ShoppingList](128)
	if err != nil {
		log.Println("Unable to initialise the lists cache:", "erorr", err.Error())
	}

	repository, err = NewRepository("./data/database.db")
	if err != nil {
		log.Println("Unable to open the database:", "error", err.Error())
		os.Exit(1)
	}

	if err := repository.Init(); err != nil {
		log.Println("Unable to initialise the database:", "error", err.Error())
		os.Exit(1)
	}

	err = seedShoppingList()
	if err != nil {
		log.Println("Unable to seed the database:", "error", err.Error())
		os.Exit(1)
	}

	log.Println("Successfully seeded shopping lists")

	port := fmt.Sprintf(":%d", 8888)

	mux := http.NewServeMux()

	SetupRoutes(mux)
	handler := corsWrapper(mux)

	server := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	log.Println("listening on port", "addr", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
