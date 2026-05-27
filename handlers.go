package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lmittmann/tint"
)

// handleCreateList creates a new shopping list
// @Summary Create a new shopping list
// @Description Create a new shopping list with the provided data
// @Tags lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param list body ShopingList true "Shopping list data"
// @Success 201 {object} ShoppingList
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /lists [post]
func handleCreateList(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Creating new shopping list",
		slog.String("ip", r.RemoteAddr),
		slog.String("user", r.Header.Get("X-User")),
		slog.String("request_id", r.Header.Get("X-Request-ID")),
	)

	var list ShoppingList
	err := json.NewDecoder(r.Body).Decode(&list)
	if err != nil {
		logger.Error("Invalid request body", tint.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repository.CreateNewShoppingList(&list)
	if err != nil {
		logger.Error("Failed to create new shopping list", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleFetchAllLists get all the shopping lists
// @Summary List all shopping lists
// @Description Get all shopping lists for the authenticated user
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ShoppingList
// @Failure 401 {string} string "Unauthorzed"
// @Router /lists [get]
func handleFetchAllLists(w http.ResponseWriter, r *http.Request) {
	shoppingLists, err := repository.GetAllShoppingLists()
	logger.Debug("Fetching all shopping lists")
	if err != nil {
		logger.Error("Failed to fetch all shopping lists", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(shoppingLists)
	if err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		logger.Error("Failed to write data to client", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleDeleteList deletes a shopping list by ID
// @Summary Delete a shopping list
// @Description Delete a shopping list by its ID
// @Tags lists
// @Security BearerAuth
// @Param id path string true "Shopping list ID"
// @Success 204 "No Content"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Shopping list not found"
// @Failure 500 {string} string
// @Router /lists/{id} [delete]
func handleDeleteList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error("Invalid id", slog.String("id", r.PathValue("id")), tint.Err(err))
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	err = repository.DeleteShoppingList(id)
	logger.Debug("Deleting shopping list", slog.Int("id", id))
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		logger.Error("Failed to delete shopping list", tint.Err(err))
		return
	}

	// deleted the list, so invalidate the cache
	listsCache.Remove(fmt.Sprintf("%d", id))
	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateList updates a shopping list completely
// @Summary Update a shopping list
// @Description Update a shopping list with new data (full replacement)
// @Tags lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Shopping list ID"
// @Param list body ShoppingList true "Updated shopping list data"
// @Success 200 {object} ShoppingList
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Shopping list not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /lists/{id} [put]
func handleUpdateList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error("Invalid id", slog.String("id", r.PathValue("id")), tint.Err(err))
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	l, err := repository.GetListByID(id)
	logger.Debug("Fetching shopping list", slog.Int("id", id))
	if err != nil || l == nil {
		logger.Error("Failed to get shopping list", tint.Err(err))
		http.Error(w, "Shopping list not found", http.StatusNotFound)
		return
	}

	// it exists, so now decode the body
	var updatedList ShoppingList // TODO: Change to ShoppingListUpdateRequest type and change repository method to take ID as well
	err = json.NewDecoder(r.Body).Decode(&updatedList)
	if err != nil {
		logger.Error("Invalid request body", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = repository.UpdateShoppingList(&updatedList)
	logger.Debug("Replacing shopping list", slog.Any("replacement", updatedList))
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		logger.Error("Error updating shopping list", tint.Err(err))
		return
	}

	// modified the list by updating, so invalidate the cache
	listsCache.Remove(fmt.Sprintf("%d", id))

	if err := json.NewEncoder(w).Encode(updatedList); err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handlePartialUpdateList partially updates a shopping list
// @Summary Partially update a shopping list
// @Description Update specific fields of a shopping list (partial update)
// @Tags lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Shopping list ID"
// @Param patch body ShoppingListPatchRequest true "Partial shopping list data"
// @Success 200 {object} ShoppingList
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Shopping list not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /lists/{id} [patch]
func handlePartialUpdateList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error("Invalid id", slog.String("id", r.PathValue("id")), tint.Err(err))
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	// the list exists: decode payload
	var patch ShoppingListPatchRequest
	err = json.NewDecoder(r.Body).Decode(&patch)
	if err != nil {
		logger.Error("Invalid request body", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = repository.PatchShoppingList(id, &patch)
	logger.Debug("Patching shopping list", slog.Any("patch", patch))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// modified the list by patching, so invalidate the cache
	listsCache.Remove(fmt.Sprintf("%d", id))

	err = json.NewEncoder(w).Encode(ShoppingList{
		ID:    id,
		Name:  *patch.Name,
		Items: patch.Items,
	})
	if err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleFetchListById retrieves a specific shopping list by ID
// @Summary Get a shopping list by ID
// @Description Retrieve a shopping list by its ID with caching support
// @Tags lists
// @Produce json
// @Security BearerAuth
// @Param id path string true "Shopping list ID"
// @Success 200 {object} ShoppingList
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Shopping list not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /lists/{id} [get]
func handleFetchListById(w http.ResponseWriter, r *http.Request) {
	// Use browser caching here to reduce and ETags for reusing cached responses.

	var err error
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error("Invalid id", slog.String("id", r.PathValue("id")), tint.Err(err))
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	// ok = cache hit
	list, ok := listsCache.Get(fmt.Sprintf("%d", id))
	if !ok {
		// NOT OK = cache miss - so fetch from db
		list, err = repository.GetListByID(id)
		logger.Debug("Fetching shopping list", slog.Int("id", id))
		if err != nil {
			switch {
			case errors.Is(err, ErrRecordNotFound):
				http.Error(w, "Shopping list not found", http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			logger.Error("Failed to get shopping list", tint.Err(err))
			return
		}
		// add to cache for next time
		listsCache.Add(fmt.Sprintf("%d", id), list)
	}

	data, err := json.Marshal(list)
	if err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
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

// handleListPush adds an item to a shopping list
// @Summary Add an item to a shopping list
// @Description Add a new item to an existing shopping list
// @Tags lists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Shopping list ID"
// @Param item body ShoppingListPushRequest true "Items to add to the list"
// @Success 200 {object} ShoppingList
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Shopping list not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /lists/{id}/push [post]
func handleListPush(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error("Invalid id", slog.String("id", r.PathValue("id")), tint.Err(err))
		http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
		return
	}

	list, err := repository.GetListByID(id)
	logger.Debug("Fetching shopping list", slog.Int("id", id))
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, "Shopping list not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		logger.Error("Failed to get shopping list", tint.Err(err))
		return
	}

	var push ShoppingListPushRequest
	err = json.NewDecoder(r.Body).Decode(&push)
	if err != nil {
		logger.Error("Invalid request body", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list.Items = append(list.Items, push.Items...)

	err = repository.PatchShoppingList(id, &ShoppingListPatchRequest{
		Name:  nil,
		Items: list.Items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// modified the list by patching, so invalidate the cache
	listsCache.Remove(fmt.Sprintf("%d", id))

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		logger.Error("Failed to marshal json", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleLogin authenticates a user and returns a session token
// @Summary User login
// @Description Authenticate user credentials and return a session token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} map[string]string "token"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /login [post]
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		logger.Error("Invalid request body", tint.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fetch user from user store
	user, exists := allUsers[payload.Username] // fetch user from user store
	if exists && user.Password == payload.Password {
		session, err := repository.AddSession(user.Username)
		logger.Debug("Adding session", slog.String("username", user.Username))
		if err != nil {
			logger.Error("Failed to add session", tint.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": session.Token})
		if err != nil {
			logger.Error("Failed to marshal json", tint.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// user doesn't exist, or password is wrong
	w.WriteHeader(http.StatusUnauthorized)
}
