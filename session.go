package main

// Session store
var sessions map[string]*Session

// User store with default users
var allUsers = map[string]*User{
	"admin": {"admin", "admin", "password"},
	"user":  {"user", "user", "password"},
}
