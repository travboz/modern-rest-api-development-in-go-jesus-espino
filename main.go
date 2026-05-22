package main

import (
	"fmt"
	"log"
	"net/http"
)

var allData []ShoppingList

func main() {
	port := 8888

	http.HandleFunc("POST /v1/lists", handleCreateList)

	log.Printf("listening on port :%d", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
