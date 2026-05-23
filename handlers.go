package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func handleCreateList(w http.ResponseWriter, r *http.Request) {
	var list ShoppingList
	err := json.NewDecoder(r.Body).Decode(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repository.CreateNewShoppingList(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleFetchAllLists(w http.ResponseWriter, r *http.Request) {
	shoppingLists, err := repository.GetAllShoppingLists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(shoppingLists)
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	err = repository.DeleteShoppingList(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateList handles an update of an entire shopping list
func handleUpdateList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	l, err := repository.GetListByID(id)
	if err != nil || l == nil {
		http.Error(w, "Shopping list not found", http.StatusNotFound)
		return
	}

	// it exists, so now decode the body
	var updatedList ShoppingList
	err = json.NewDecoder(r.Body).Decode(&updatedList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: perform update - add in repository method and finish handler

	if err := json.NewEncoder(w).Encode(updatedList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	list, err := repository.GetListByID(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleListPush(w http.ResponseWriter, r *http.Request) {
	// check if list exists
	id := r.PathValue("id")
	for i, list := range allData {
		// check if shopping list exists
		if strconv.Itoa(list.ID) == id {
			// it exists, so add to its items list
			var push ListPushAction

			err := json.NewDecoder(r.Body).Decode(&push)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			list.Items = append(list.Items, push.Items...)

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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, exists := allUsers[payload.Username] // fetch user from user store
	if exists && user.Password == payload.Password {
		token, err := GenerateSessionToken()
		if err != nil {
			http.Error(w, fmt.Errorf("error generating token: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		sessions[token] = &Session{
			Expires:  time.Now().Add(1 * 24 * time.Hour),
			Username: user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// user doesn't exist, or password is wrong
	w.WriteHeader(http.StatusUnauthorized)
}
