package main

import (
	"encoding/json"
	"net/http"
)

func handleCreateList(w http.ResponseWriter, r *http.Request) {
	var list ShoppingList
	err := json.NewDecoder(r.Body).Decode(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allData = append(allData, list)
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
