package main

import "time"

// Handler types
type ShoppingList struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type ShoppingListUpdateRequest struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type ShoppingListPatch struct {
	Name  *string  `json:"name"`
	Items []string `json:"items"`
}

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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
