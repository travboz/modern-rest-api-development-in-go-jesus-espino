// @title           Shopping List API
// @version         0.1
// @description     An API for managing shopping lists

// @host            localhost:8888
// @BasePath        /v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and the JWT token
package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/lmittmann/tint"

	_ "github.com/travboz/modern-rest-api-dev/shopping-list-api/docs" // for swagger api docs
)

func seedShoppingList() error {
	err := repository.Empty()
	if err != nil {
		return err
	}

	lists := []ShoppingList{
		{Name: "Saturday shopping list", Items: []string{"bread", "ice cream", "milk", "pasta", "toothpaste", "eggs", "soap", "detergent"}},
		{Name: "Hamburger night", Items: []string{"beef patties", "burger rolls", "eggs", "bacon", "tomatoes", "sliced cheese", "bbq sauce", "beetroot", "butter", "lettuce"}},
	}

	for _, l := range lists {
		err := repository.CreateNewShoppingList(&l)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	KitchenWithSeconds = "03:04:05 PM"
)

// tint.Err(exampleError) for RED COLOURED error hightlighting

var (
	repository RepositoryInterface
	listsCache *lru.Cache[string, *ShoppingList]
	logger     *slog.Logger
)

func main() {
	logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelInfo,
		AddSource:  true,
		TimeFormat: KitchenWithSeconds,
	}))

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
