package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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

	err = repository.UpdateShoppingList(&updatedList)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if err := json.NewEncoder(w).Encode(updatedList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handlePartialUpdateList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	// the list exists: decode payload
	var patch ShoppingListPatch
	err = json.NewDecoder(r.Body).Decode(&patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = repository.PatchShoppingList(id, &patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(ShoppingList{
		ID:    id,
		Name:  *patch.Name,
		Items: patch.Items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Use browser caching here to reduce and ETags for reusing cached responses.
// handleFetchListById fetches a single shopping list by its ID
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

	data, err := json.Marshal(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// In this case, we want to pass the Cache-Control: no-cache configuration
	// to ensure the client always revalidates the cache with the server using
	// the recently added ETag header.
	w.Header().Set("Cache-Control", "no-cache")

	etag := fmt.Sprintf(`"%x`, sha256.Sum256(data))
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Write(data)
}

func handleListPush(w http.ResponseWriter, r *http.Request) {
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

	var push ShoppingListPushRequest
	err = json.NewDecoder(r.Body).Decode(&push)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list.Items = append(list.Items, push.Items...)

	err = repository.PatchShoppingList(id, &ShoppingListPatch{
		Name:  nil,
		Items: list.Items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fetch user from user store
	user, exists := allUsers[payload.Username] // fetch user from user store
	if exists && user.Password == payload.Password {
		// token, err := GenerateSessionToken()
		// if err != nil {
		// 	http.Error(w, fmt.Errorf("error generating token: %w", err).Error(), http.StatusInternalServerError)
		// 	return
		// }

		// sessions[token] = &Session{
		// 	Expires:  time.Now().Add(1 * 24 * time.Hour),
		// 	Username: user.Username,
		// }

		session, err := repository.AddSession(user.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": session.Token})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// user doesn't exist, or password is wrong
	w.WriteHeader(http.StatusUnauthorized)
}
