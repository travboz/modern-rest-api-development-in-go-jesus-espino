package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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

func handleFetchAllLists(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(allData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeleteList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for i, list := range allData {
		if strconv.Itoa(list.ID) == id {
			allData = append(allData[:i], allData[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "List not found", http.StatusNotFound)
}

// handleUpdateList handles a modification of an entire shopping list
func handleUpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for i, list := range allData {
		// check if shopping list exists
		if strconv.Itoa(list.ID) == id {
			// it exists, so now decode the body
			var updatedList ShoppingList
			err := json.NewDecoder(r.Body).Decode(&updatedList)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			allData[i] = updatedList // replace the current list with the updated list

			if err := json.NewEncoder(w).Encode(updatedList); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}

	http.Error(w, "List not found", http.StatusNotFound)
}

func handlePartialUpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for i, list := range allData {
		// check if shopping list exists
		if strconv.Itoa(list.ID) == id {
			// the list exists: do work
			var patch ShoppingListPatch
			err := json.NewDecoder(r.Body).Decode(&patch)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// using nil check to see if a value was provided for that field - if nil,
			// value hasn't been given to reference
			if patch.Name != nil {
				list.Name = *patch.Name
			}

			if patch.Items != nil {
				list.Items = patch.Items
			}

			allData[i] = list

			err = json.NewEncoder(w).Encode(list)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}

	http.Error(w, "List not found", http.StatusNotFound)
}

// handleFetchListById fetches a single shopping list by its
func handleFetchListById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, list := range allData {
		// check if shopping list exists
		if strconv.Itoa(list.ID) == id {
			// it exists, so send it to the client
			data, err := json.Marshal(list)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return // because successfully written to response
		}
	}

	http.Error(w, "List not found", http.StatusNotFound)
}
