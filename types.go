package main

import "time"

// ShoppingList represents a shopping list with items
// @Description Shopping list with items
type ShoppingList struct {
	ID    int      `json:"id" example:"1001"`
	Name  string   `json:"name" example:"Grocery List"`
	Items []string `json:"items" example:"milk,eggs,bread"`
}

// TODO: Potentially create CreateShoppingListRequest and CreateShoppingListResponse

// ShoppingListUpdateRequest represents complete updates to a shopping list
// @Description Complete shopping list update structure
type ShoppingListUpdateRequest struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

// ShoppingListPatch represents partial updates to a shopping list
// @Description Partial shopping list update structure
type ShoppingListPatchRequest struct {
	Name  *string  `json:"name"`
	Items []string `json:"items"`
}

// ShoppingListPushRequest represents items to append to a shopping list
// @Description Request body for pushing items onto a list
type ShoppingListPushRequest struct {
	Items []string `json:"items"`
}

// Authentication types
type User struct {
	Role     string
	Username string
	Password string
}

type Session struct {
	Token    string // Add token to session object
	Expires  time.Time
	Username string
}

// LoginRequest represents user login credentials
// @Description User login request structure
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
